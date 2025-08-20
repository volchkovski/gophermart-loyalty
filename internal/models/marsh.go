package models

import (
	"encoding/json"
	"math"
)

func (b Balance) MarshalJSON() ([]byte, error) {
	type Alias struct {
		Current   float64 `json:"current"`
		Withdrawn float64 `json:"withdrawn"`
	}
	return json.Marshal(Alias{
		Current:   float64(b.Current) / 100,
		Withdrawn: float64(b.Withdrawn) / 100,
	})
}

func (w *Withdrawal) UnmarshalJSON(data []byte) error {
	type Alias Withdrawal
	var temp struct {
		Alias
		Sum float64 `json:"sum"`
	}

	if err := json.Unmarshal(data, &temp); err != nil {
		return err
	}

	*w = Withdrawal(temp.Alias)
	w.Sum = int64(math.Round(temp.Sum * 100))

	return nil
}

func (w *Withdrawal) MarshalJSON() ([]byte, error) {
	type Alias Withdrawal
	return json.Marshal(&struct {
		Sum float64 `json:"sum"`
		*Alias
	}{
		Sum:   float64(w.Sum) / 100,
		Alias: (*Alias)(w),
	})
}

func (o *Order) MarshalJSON() ([]byte, error) {
	type Alias Order
	return json.Marshal(&struct {
		Accrual float64 `json:"accrual,omitempty"`
		*Alias
	}{
		Accrual: float64(o.Accrual) / 100,
		Alias:   (*Alias)(o),
	})
}
