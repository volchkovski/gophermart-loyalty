INSERT INTO orders_updating (number, user_id, deadline)
SELECT o.number, user_id, (NOW() AT TIME ZONE 'UTC' + INTERVAL '1 hour')
FROM orders o
         LEFT JOIN orders_updating ou ON o.number = ou.number
WHERE o.status NOT IN ('INVALID', 'PROCESSED')
  AND (ou.number IS NULL OR ou.deadline < (NOW() AT TIME ZONE 'UTC'))
ON CONFLICT (number)
    DO UPDATE SET deadline = EXCLUDED.deadline
RETURNING number, user_id;
