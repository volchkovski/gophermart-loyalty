package orderupdater

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/go-resty/resty/v2"
	"github.com/volchkovski/gophermart-loyalty/internal/models"
	"time"
)

const (
	StatusProcessed  = "PROCESSED"
	StatusProcessing = "PROCESSING"
	StatusInvalid    = "INVALID"
	StatusRegistered = "REGISTERED"
)

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
	go func() {
		defer close(ou.notify)
		orderCh := make(chan *models.UnprocessedOrder)
		defer close(orderCh)
		errsCh := make(chan error)
		defer close(errsCh)

		ou.startWorkers(ctx, orderCh, errsCh)

		ticker := time.NewTicker(3 * time.Minute)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case err := <-errsCh:
				ou.notify <- err
				return
			case <-ticker.C:
				if err := ou.processOrders(ctx, orderCh); err != nil {
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
	for _, o := range ordrs {
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
	defer func() {
		if errDel := ou.db.DeleteProcessedOrder(ctx, ordr.Number); errDel != nil {
			err = errors.Join(err, errDel)
		}
	}()
	resp, err := ou.client.R().
		SetContext(ctx).
		Get("http://" + ou.accrualAddr + "/api/orders/" + ordr.Number)
	if err != nil {
		return err
	}
	var o models.Order
	if err = json.Unmarshal(resp.Body(), &o); err != nil {
		return err
	}
	switch o.Status {
	case StatusProcessed:
		if err = ou.updateOrder(ctx, &o); err != nil {
			return
		}
		if err = ou.db.RegisterTx(ctx, ordr.UserID, o.Accrual, o.Number); err != nil {
			return
		}
	case StatusProcessing, StatusInvalid:
		if err = ou.updateOrder(ctx, &o); err != nil {
			return
		}
	case StatusRegistered:
	default:
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
