WITH locked_row AS (SELECT id
                    FROM orders
                    WHERE number = $3
                        FOR UPDATE)
UPDATE orders
SET status  = $1,
    accrual = $2
WHERE number = $3;
