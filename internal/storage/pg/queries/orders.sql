SELECT number, status, accrual, uploaded_at FROM orders WHERE user_id = $1;
