package orderupdater

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/volchkovski/gophermart-loyalty/internal/logger"
	"github.com/volchkovski/gophermart-loyalty/internal/models"
)

const (
	StatusProcessed  = "PROCESSED"
	StatusProcessing = "PROCESSING"
	StatusInvalid    = "INVALID"
	StatusRegistered = "REGISTERED"
)

type AccrualResponse struct {
	Order   string  `json:"order"`
	Status  string  `json:"status"`
	Accrual float64 `json:"accrual,omitempty"`
}

type db interface {
	UnprocessedOrders(ctx context.Context) ([]*models.UnprocessedOrder, error)
	UpdateOrder(ctx context.Context, number, status string, accrual int64) error
	RegisterTx(ctx context.Context, userID, amount int64, orderNumber string) error
	DeleteProcessedOrder(ctx context.Context, orderNumber string) error
}

type OrderUpdater struct {
	client      *resty.Client
	db          db
	accrualAddr string
	notify      chan error
}

func New(db db, accrualAddr string) *OrderUpdater {
	return &OrderUpdater{
		client:      newRestyClient(),
		db:          db,
		accrualAddr: accrualAddr,
		notify:      make(chan error, 1),
	}
}

func (ou *OrderUpdater) Notify() chan error {
	return ou.notify
}

func (ou *OrderUpdater) Start(ctx context.Context) {
	logger.Log.Info("OrderUpdater starting...")
	go func() {
		defer close(ou.notify)
		orderCh := make(chan *models.UnprocessedOrder)
		defer close(orderCh)
		errsCh := make(chan error)
		defer close(errsCh)

		ou.startWorkers(ctx, orderCh, errsCh)

		logger.Log.Info("OrderUpdater: Initial order processing...")
		if err := ou.processOrders(ctx, orderCh); err != nil {
			logger.Log.Errorf("OrderUpdater: Initial processing error: %s", err)
			ou.notify <- err
			return
		}

		ticker := time.NewTicker(2 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				logger.Log.Info("OrderUpdater: Context cancelled, stopping...")
				return
			case err := <-errsCh:
				logger.Log.Errorf("OrderUpdater: Worker error: %s", err)
				ou.notify <- err
				return
			case <-ticker.C:
				logger.Log.Debug("OrderUpdater: Processing orders on tick...")
				if err := ou.processOrders(ctx, orderCh); err != nil {
					logger.Log.Errorf("OrderUpdater: Processing error: %s", err)
					ou.notify <- err
					return
				}
			}
		}
	}()
}

func (ou *OrderUpdater) processOrders(ctx context.Context, orderCh chan<- *models.UnprocessedOrder) error {
	ordrs, err := ou.db.UnprocessedOrders(ctx)
	if err != nil {
		return err
	}
	logger.Log.Infof("OrderUpdater: Found %d unprocessed orders", len(ordrs))
	for _, o := range ordrs {
		logger.Log.Debugf("OrderUpdater: Sending order %s (user %d) to worker", o.Number, o.UserID)
		orderCh <- o
	}
	return nil
}

func (ou *OrderUpdater) startWorkers(ctx context.Context, ordrCh <-chan *models.UnprocessedOrder, errs chan<- error) {
	for i := 0; i < 5; i++ {
		go ou.orderWorker(ctx, ordrCh, errs)
	}

}

func (ou *OrderUpdater) orderWorker(ctx context.Context, ordrCh <-chan *models.UnprocessedOrder, errs chan<- error) {
	for ordr := range ordrCh {
		if err := ou.processOrder(ctx, ordr); err != nil {
			errs <- err
			return
		}
	}
}

func (ou *OrderUpdater) processOrder(ctx context.Context, ordr *models.UnprocessedOrder) (err error) {
	logger.Log.Infof("OrderUpdater: Processing order %s (user %d)", ordr.Number, ordr.UserID)
	defer func() {
		if errDel := ou.db.DeleteProcessedOrder(ctx, ordr.Number); errDel != nil {
			logger.Log.Errorf("OrderUpdater: Failed to delete processed order %s: %s", ordr.Number, errDel)
			err = errors.Join(err, errDel)
		}
	}()

	var accrualURL string
	if strings.HasPrefix(ou.accrualAddr, "http://") || strings.HasPrefix(ou.accrualAddr, "https://") {
		accrualURL = ou.accrualAddr + "/api/orders/" + ordr.Number
	} else {
		accrualURL = "http://" + ou.accrualAddr + "/api/orders/" + ordr.Number
	}
	logger.Log.Debugf("OrderUpdater: Requesting accrual for order %s from %s", ordr.Number, accrualURL)

	resp, err := ou.client.R().
		SetContext(ctx).
		Get(accrualURL)
	if err != nil {
		logger.Log.Errorf("OrderUpdater: HTTP request error for order %s: %s", ordr.Number, err)
		return err
	}

	logger.Log.Debugf("OrderUpdater: Received response for order %s: status=%d, body=%s",
		ordr.Number, resp.StatusCode(), string(resp.Body()))

	switch resp.StatusCode() {
	case 204:
		logger.Log.Infof("OrderUpdater: Order %s not yet registered in accrual system (204)", ordr.Number)
		return nil
	case 429:
		logger.Log.Warnf("OrderUpdater: Rate limit exceeded for order %s (429), retry later", ordr.Number)
		return nil
	case 500:
		logger.Log.Errorf("OrderUpdater: Accrual system error for order %s (500)", ordr.Number)
		return fmt.Errorf("accrual system error (500) for order %s", ordr.Number)
	case 200:

	default:
		logger.Log.Errorf("OrderUpdater: Unexpected HTTP status %d for order %s", resp.StatusCode(), ordr.Number)
		return fmt.Errorf("unexpected HTTP status %d for order %s", resp.StatusCode(), ordr.Number)
	}

	var accrualResp AccrualResponse
	if err = json.Unmarshal(resp.Body(), &accrualResp); err != nil {
		logger.Log.Errorf("OrderUpdater: JSON unmarshal error for order %s: %s", ordr.Number, err)
		return err
	}

	logger.Log.Infof("OrderUpdater: Parsed accrual response for order %s: status=%s, accrual=%.2f",
		ordr.Number, accrualResp.Status, accrualResp.Accrual)

	o := &models.Order{
		Number:  accrualResp.Order,
		Status:  accrualResp.Status,
		Accrual: int64(math.Round(accrualResp.Accrual * 100)),
	}

	switch o.Status {
	case StatusProcessed:
		logger.Log.Infof("OrderUpdater: Order %s processed with accrual %d kopecks", o.Number, o.Accrual)
		if err = ou.updateOrder(ctx, o); err != nil {
			logger.Log.Errorf("OrderUpdater: Failed to update order %s: %s", o.Number, err)
			return
		}
		if err = ou.db.RegisterTx(ctx, ordr.UserID, o.Accrual, o.Number); err != nil {
			logger.Log.Errorf("OrderUpdater: Failed to register transaction for order %s: %s", o.Number, err)
			return
		}
		logger.Log.Infof("OrderUpdater: Successfully processed and registered order %s", o.Number)
	case StatusProcessing:
		logger.Log.Infof("OrderUpdater: Order %s still processing", o.Number)
		if err = ou.updateOrder(ctx, o); err != nil {
			logger.Log.Errorf("OrderUpdater: Failed to update processing order %s: %s", o.Number, err)
			return
		}
	case StatusInvalid:
		logger.Log.Infof("OrderUpdater: Order %s marked as invalid", o.Number)
		if err = ou.updateOrder(ctx, o); err != nil {
			logger.Log.Errorf("OrderUpdater: Failed to update invalid order %s: %s", o.Number, err)
			return
		}
	case StatusRegistered:
		logger.Log.Infof("OrderUpdater: Order %s registered, waiting for processing", o.Number)
	default:
		logger.Log.Errorf("OrderUpdater: Unexpected status %s for order %s", o.Status, o.Number)
		return fmt.Errorf("unexpected status: %s", o.Status)
	}
	return
}

func (ou *OrderUpdater) updateOrder(ctx context.Context, o *models.Order) error {
	if err := ou.db.UpdateOrder(ctx, o.Number, o.Status, o.Accrual); err != nil {
		return err
	}
	return nil
}
