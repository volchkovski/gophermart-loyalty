package pg

import (
	"embed"
	"fmt"
	"sync"
)

var (
	once sync.Once
	q    queries
)

type queries struct {
	UserByLogin        string
	UserInsert         string
	Orders             string
	OrderUserID        string
	OrderInsert        string
	LoyaltyGot         string
	LoyaltyWithdrawn   string
	LoyaltyRegisterTx  string
	LoyaltyWithdrawals string
	UpdaterOrders      string
	UpdaterDeleteOrder string
	UpdaterOrderUpdate string
}

//go:embed queries/*.sql
var queryFS embed.FS

func loadQueries() error {
	var initErr error
	once.Do(func() {
		userByLoginQ, err := loadQuery("user_by_login")
		if err != nil {
			initErr = err
			return
		}
		userInsertQ, err := loadQuery("user_insert")
		if err != nil {
			initErr = err
			return
		}
		ordersQ, err := loadQuery("orders")
		if err != nil {
			initErr = err
			return
		}
		orderUserIDQ, err := loadQuery("order_user_id")
		if err != nil {
			initErr = err
			return
		}
		orderInsertQ, err := loadQuery("order_insert")
		if err != nil {
			initErr = err
			return
		}
		loyaltyGotQ, err := loadQuery("loyalty_got")
		if err != nil {
			initErr = err
			return
		}
		loyaltyWithdrawnQ, err := loadQuery("loyalty_withdrawn")
		if err != nil {
			initErr = err
			return
		}
		loyaltyRegisterTxQ, err := loadQuery("loyalty_register_tx")
		if err != nil {
			initErr = err
			return
		}
		loyaltyWithdrawalsQ, err := loadQuery("loyalty_withdrawals")
		if err != nil {
			initErr = err
			return
		}
		updaterOrdersQ, err := loadQuery("updater_orders")
		if err != nil {
			initErr = err
			return
		}
		updaterDeleteOrderQ, err := loadQuery("updater_delete_order")
		if err != nil {
			initErr = err
			return
		}
		updaterOrderUpdateQ, err := loadQuery("updater_order_update")
		if err != nil {
			initErr = err
			return
		}
		q = queries{
			UserByLogin:        userByLoginQ,
			UserInsert:         userInsertQ,
			Orders:             ordersQ,
			OrderUserID:        orderUserIDQ,
			OrderInsert:        orderInsertQ,
			LoyaltyGot:         loyaltyGotQ,
			LoyaltyWithdrawn:   loyaltyWithdrawnQ,
			LoyaltyRegisterTx:  loyaltyRegisterTxQ,
			LoyaltyWithdrawals: loyaltyWithdrawalsQ,
			UpdaterOrders:      updaterOrdersQ,
			UpdaterDeleteOrder: updaterDeleteOrderQ,
			UpdaterOrderUpdate: updaterOrderUpdateQ,
		}
	})
	if initErr != nil {
		return fmt.Errorf("failed to load queries: %w", initErr)
	}
	return nil
}

func loadQuery(filename string) (string, error) {
	fp := fmt.Sprintf("queries/%s.sql", filename)
	query, err := queryFS.ReadFile(fp)
	if err != nil {
		return "", fmt.Errorf("failed to read %s query: %w", fp, err)
	}
	return string(query), nil
}
