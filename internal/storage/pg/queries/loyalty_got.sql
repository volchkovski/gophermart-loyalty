SELECT COALESCE(SUM(amount), 0) FROM transactions WHERE user_id = $1 AND amount > 0;
