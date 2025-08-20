package pg

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/volchkovski/gophermart-loyalty/internal/models"
	"github.com/volchkovski/gophermart-loyalty/internal/storage"
	"github.com/volchkovski/gophermart-loyalty/internal/storage/pg/migrator"
	"time"
)

const (
	MaxOpenConns = 5
	MaxIdleConns
	MaxLifetime = 5 * time.Minute
	MaxIdleTime = 10 * time.Minute
)

type Pg struct {
	db *sql.DB
}

func configureConnPool(db *sql.DB) {
	db.SetMaxOpenConns(MaxOpenConns)
	db.SetMaxIdleConns(MaxIdleConns)
	db.SetConnMaxLifetime(MaxLifetime)
	db.SetConnMaxIdleTime(MaxIdleTime)
}

func New(dsn string) (*Pg, error) {
	if err := migrator.Run(dsn); err != nil {
		return nil, fmt.Errorf("migrations failed: %w", err)
	}
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, err
	}
	configureConnPool(db)
	if err = loadQueries(); err != nil {
		return nil, fmt.Errorf("failed to load queries: %w", err)
	}
	return &Pg{db}, nil
}

func (pg *Pg) Close() error {
	return pg.db.Close()
}

func (pg *Pg) User(ctx context.Context, login string) (*models.User, error) {
	var u models.User
	if err := pg.db.QueryRowContext(ctx, q.UserByLogin, login).Scan(&u.ID, &u.Login, &u.HashedPassword); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &u, nil
}

func (pg *Pg) NewUser(ctx context.Context, login string, passwordHash string) (*models.User, error) {
	var userID int64
	if err := pg.db.QueryRowContext(ctx, q.UserInsert, login, passwordHash).Scan(&userID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, storage.ErrUserLoginAlreadyExists
		}
		return nil, fmt.Errorf("failed to creating new user: %w", err)
	}
	return &models.User{
		ID:             userID,
		Login:          login,
		HashedPassword: passwordHash,
	}, nil
}

func (pg *Pg) Orders(ctx context.Context, userID int64) (userOrders []*models.Order, err error) {
	userOrders = make([]*models.Order, 0, 20)
	rows, err := pg.db.QueryContext(ctx, q.Orders, userID)
	if err != nil {
		return
	}
	defer func() {
		if errClose := rows.Close(); errClose != nil {
			err = errors.Join(err, errClose)
		}
	}()
	for rows.Next() {
		var o models.Order
		if err = rows.Scan(&o.Number, &o.Status, &o.Accrual, &o.UploadedAt); err != nil {
			return
		}
		userOrders = append(userOrders, &o)
	}
	err = rows.Err()
	return
}

func (pg *Pg) SaveOrder(ctx context.Context, userID int64, orderNumber string) (err error) {
	tx, err := pg.db.BeginTx(ctx, nil)
	if err != nil {
		return
	}
	defer func() {
		if err == nil {
			return
		}
		if errRB := tx.Rollback(); errRB != nil {
			err = errors.Join(err, errRB)
		}
	}()
	var dbUserID int64
	if err = tx.QueryRowContext(ctx, q.OrderUserID, userID).Scan(&dbUserID); err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			return
		}
		if _, err = tx.ExecContext(ctx, q.OrderInsert, userID, orderNumber); err != nil {
			return
		}
		err = tx.Commit()
		return
	}
	if dbUserID == userID {
		return storage.ErrOrderAlreadyExists
	}
	return storage.ErrOrderAnotherUser
}

func (pg *Pg) Balance(ctx context.Context, userID int64) (b *models.Balance, err error) {
	tx, err := pg.db.BeginTx(ctx, nil)
	if err != nil {
		return
	}
	defer func() {
		if err == nil {
			return
		}
		if errRB := tx.Rollback(); errRB != nil {
			err = errors.Join(err, errRB)
		}
	}()
	var (
		got       int64
		withdrawn int64
	)
	if err = tx.QueryRowContext(ctx, q.LoyaltyGot, userID).Scan(&got); err != nil {
		return
	}
	if err = tx.QueryRowContext(ctx, q.LoyaltyWithdrawn, userID).Scan(&withdrawn); err != nil {
		return
	}
	if err = tx.Commit(); err != nil {
		return
	}
	b = &models.Balance{
		Current:   got - withdrawn,
		Withdrawn: withdrawn,
	}
	return
}

func (pg *Pg) RegisterTx(ctx context.Context, userID, amount int64, orderNumber string) (err error) {
	tx, err := pg.db.BeginTx(ctx, nil)
	if err != nil {
		return
	}
	defer func() {
		if err == nil {
			return
		}
		if errRB := tx.Rollback(); errRB != nil {
			err = errors.Join(err, errRB)
		}
	}()
	var id int64
	if err = tx.QueryRowContext(ctx, q.LoyaltyRegisterTx, userID, orderNumber, amount).Scan(&id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return storage.ErrLoyaltyNotEnough
		}
		return
	}
	err = tx.Commit()
	return
}

func (pg *Pg) Withdrawals(ctx context.Context, userID int64) (ws []*models.Withdrawal, err error) {
	rows, err := pg.db.QueryContext(ctx, q.LoyaltyWithdrawals, userID)
	if err != nil {
		return
	}
	defer func() {
		if errClose := rows.Close(); errClose != nil {
			err = errors.Join(err, errClose)
		}
	}()
	ws = make([]*models.Withdrawal, 0, 20)
	for rows.Next() {
		var w models.Withdrawal
		if err = rows.Scan(&w.Order, &w.Sum, &w.ProcessedAt); err != nil {
			return
		}
		ws = append(ws, &w)
	}
	err = rows.Err()
	return
}

func (pg *Pg) UnprocessedOrders(ctx context.Context) (ordrs []*models.UnprocessedOrder, err error) {
	tx, err := pg.db.BeginTx(ctx, nil)
	if err != nil {
		return
	}
	func() {
		if err == nil {
			return
		}
		if errRB := tx.Rollback(); errRB != nil {
			err = errors.Join(err, errRB)
		}
	}()
	rows, err := tx.QueryContext(ctx, q.UpdaterOrders)
	if err != nil {
		return
	}
	ordrs = make([]*models.UnprocessedOrder, 0, 20)
	for rows.Next() {
		var ordr models.UnprocessedOrder
		if err = rows.Scan(&ordr.Number, &ordr.UserID); err != nil {
			return
		}
		ordrs = append(ordrs, &ordr)
	}
	if err = rows.Err(); err != nil {
		return
	}
	err = tx.Commit()
	return
}

func (pg *Pg) UpdateOrder(ctx context.Context, number, status string, accrual int64) (err error) {
	tx, err := pg.db.BeginTx(ctx, nil)
	if err != nil {
		return
	}
	func() {
		if err == nil {
			return
		}
		if errRB := tx.Rollback(); errRB != nil {
			err = errors.Join(err, errRB)
		}
	}()
	if _, err = tx.ExecContext(ctx, q.UpdaterOrderUpdate, status, accrual, number); err != nil {
		return
	}
	err = tx.Commit()
	return
}

func (pg *Pg) DeleteProcessedOrder(ctx context.Context, orderNumber string) (err error) {
	tx, err := pg.db.BeginTx(ctx, nil)
	if err != nil {
		return
	}
	func() {
		if err == nil {
			return
		}
		if errRB := tx.Rollback(); errRB != nil {
			err = errors.Join(err, errRB)
		}
	}()
	if _, err = tx.ExecContext(ctx, q.UpdaterDeleteOrder, orderNumber); err != nil {
		return
	}
	err = tx.Commit()
	return
}
