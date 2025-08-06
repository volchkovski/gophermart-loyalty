package orders

import (
	"context"
	"fmt"
	"github.com/volchkovski/gophermart-loyalty/internal/models"
	"net/http"
)

type ordersDB interface {
	Order(ctx context.Context, orderNumber int) (*models.Order, error)
	Orders(ctx context.Context, userID int) ([]*models.Order, error)
	SaveOrder(context.Context, *models.Order) error
}

type Manager struct {
	db ordersDB
}

func (m *Manager) Orders(ctx context.Context, userID int) ([]*models.Order, error) {
	orders, err := m.db.Orders(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("error getting orders: %w", err)
	}
	return orders, nil
}

func (m *Manager) SaveOrder(ctx context.Context, order *models.Order) (*models.SaveOrderResult, error) {
	o, err := m.db.Order(ctx, order.Number)
	if err != nil {
		return nil, fmt.Errorf("error getting order by number: %w", err)
	}
	if o == nil {
		if err = m.db.SaveOrder(ctx, order); err != nil {
			return nil, fmt.Errorf("error saving order: %w", err)
		}
		return &models.SaveOrderResult{
			StatusCode: http.StatusAccepted,
			Fail:       nil,
		}, nil
	}
	if o.UserID == order.UserID {
		return &models.SaveOrderResult{
			StatusCode: http.StatusOK,
			Fail:       nil,
		}, nil
	}
	return &models.SaveOrderResult{
		StatusCode: 0,
		Fail: &models.Fail{
			StatusCode: http.StatusConflict,
			Msg:        "order exists already",
		},
	}, nil
}
