package models

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"
)

func (o Order) MarshalJSON() ([]byte, error) {
	type Alias Order
	return json.Marshal(&struct {
		UploadedAt string `json:"uploaded_at"`
		*Alias
	}{
		UploadedAt: time.Unix(int64(o.UploadedAt), 0).Format(time.RFC3339),
		Alias:      (*Alias)(&o),
	})
}

func (b BalanceResult) MarshalJSON() ([]byte, error) {
	type Alias BalanceResult
	return json.Marshal(&struct {
		Current float64 `json:"current"`
		*Alias
	}{
		Current: float64(b.Current) / 100.0,
		Alias:   (*Alias)(&b),
	})
}

func (w *Withdrawal) UnmarshalJSON(data []byte) error {
	var temp struct {
		Order string `json:"order"`
		*Withdrawal
	}
	if err := json.Unmarshal(data, &temp); err != nil {
		return err
	}
	order, err := strconv.Atoi(temp.Order)
	if err != nil {
		return fmt.Errorf("failed to convert string order to integer: %w", err)
	}

	w.Order = order
	w.Sum = temp.Sum
	return nil
}

func (w *Withdrawal) MarshalJSON() ([]byte, error) {
	type Alias Withdrawal
	return json.Marshal(&struct {
		Order       string `json:"order"`
		ProcessedAt string `json:"processed_at"`
		*Alias
	}{
		Order:       strconv.Itoa(w.Order),
		ProcessedAt: time.Unix(int64(w.ProcessedAt), 0).Format(time.RFC3339),
		Alias:       (*Alias)(w),
	})
}
