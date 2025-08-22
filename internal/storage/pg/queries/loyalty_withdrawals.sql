SELECT order_number, ABS(amount) as amount, processed_at
FROM transactions
WHERE user_id = $1
  AND amount < 0
ORDER BY processed_at DESC;
