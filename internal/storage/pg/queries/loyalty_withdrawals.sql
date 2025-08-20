SELECT order_number, sum, processed_at
FROM transactions
WHERE user_id = $1
  AND sum < 0;
