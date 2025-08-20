SELECT order_number, amount, processed_at
FROM transactions
WHERE user_id = $1
  AND amount < 0
ORDER BY processed_at DESC;
