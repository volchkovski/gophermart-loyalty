package loyalty

import (
	"context"
	"errors"
	"fmt"
	"github.com/volchkovski/gophermart-loyalty/internal/models"
	"github.com/volchkovski/gophermart-loyalty/internal/storage"
)

type db interface {
	Balance(ctx context.Context, userID int64) (*models.Balance, error)
	RegisterTx(ctx context.Context, userID, amount int64, orderNumber string) error
	Withdrawals(ctx context.Context, userID int64) ([]*models.Withdrawal, error)
}

type Manager struct {
	db db
}

func NewManager(db db) *Manager {
	return &Manager{db: db}
}

func (m *Manager) Balance(ctx context.Context, userID int64) (*models.Balance, error) {
	b, err := m.db.Balance(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed getting loyalty: %w", err)
	}
	return b, nil
}

func (m *Manager) Withdraw(ctx context.Context, userID, sum int64, orderNumber string) error {
	if err := m.db.RegisterTx(ctx, userID, -1*sum, orderNumber); err != nil {
		if errors.Is(err, storage.ErrLoyaltyNotEnough) {
			return ErrNotEnough
		}
		return fmt.Errorf("failed to register order transaction: %w", err)
	}
	return nil
}

func (m *Manager) Withdrawals(ctx context.Context, userID int64) ([]*models.Withdrawal, error) {
	return m.db.Withdrawals(ctx, userID)
}
