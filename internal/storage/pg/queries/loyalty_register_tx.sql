WITH locked_rows AS (SELECT id
                     FROM transactions
                     WHERE user_id = $1 FOR UPDATE),
     current_balance AS (SELECT COALESCE(SUM(amount), 0) AS balance
                         FROM transactions
                         WHERE user_id = $1),
     new_transaction AS (
         INSERT INTO transactions (user_id, order_number, amount)
             SELECT $1, $2, $3
             FROM current_balance
             WHERE (current_balance.balance + $3) >= 0
             RETURNING id)
SELECT *
FROM new_transaction;
