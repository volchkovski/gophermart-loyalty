package orders

import (
	"context"
	"errors"
	"fmt"
	"github.com/volchkovski/gophermart-loyalty/internal/models"
	"github.com/volchkovski/gophermart-loyalty/internal/storage"
	"time"
)

type db interface {
	Orders(ctx context.Context, userID int64) ([]*models.Order, error)
	SaveOrder(ctx context.Context, userID int64, orderNumber string) error
}

type Manager struct {
	db db
}

func NewManager(db db) *Manager {
	return &Manager{db: db}
}

type Order struct {
	Number     string    `json:"number"`
	Status     string    `json:"status"`
	Accrual    int64     `json:"accrual,omitempty"`
	UploadedAt time.Time `json:"uploaded_at"`
}

func (m *Manager) Orders(ctx context.Context, userID int64) ([]*models.Order, error) {
	orders, err := m.db.Orders(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("error getting orders: %w", err)
	}
	return orders, nil
}

func (m *Manager) SaveOrder(ctx context.Context, userID int64, orderNumber string) error {
	if err := m.db.SaveOrder(ctx, userID, orderNumber); err != nil {
		if errors.Is(err, storage.ErrOrderAnotherUser) {
			return ErrAnotherUser
		}
		if errors.Is(err, storage.ErrOrderAlreadyExists) {
			return ErrAlreadyExists
		}
		return fmt.Errorf("failed saving order: %w", err)
	}
	return nil
}
