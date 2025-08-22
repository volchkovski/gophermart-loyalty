CREATE TABLE IF NOT EXISTS users
(
    id            SERIAL PRIMARY KEY,
    login         VARCHAR(50) NOT NULL UNIQUE,
    password_hash TEXT        NOT NULL
);

CREATE TABLE IF NOT EXISTS transactions
(
    id           SERIAL PRIMARY KEY,
    user_id      INTEGER                  NOT NULL REFERENCES users (id),
    order_number TEXT                     NOT NULL,
    amount       INTEGER                  NOT NULL,
    processed_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT (NOW() AT TIME ZONE 'UTC')
);

CREATE TYPE order_status AS ENUM ('NEW', 'PROCESSING', 'INVALID', 'PROCESSED');

CREATE TABLE IF NOT EXISTS orders
(
    id          SERIAL PRIMARY KEY,
    user_id     INTEGER                  NOT NULL REFERENCES users (id),
    number      TEXT                     NOT NULL UNIQUE,
    status      order_status             NOT NULL DEFAULT 'NEW',
    accrual     INTEGER                  NOT NULL DEFAULT 0,
    uploaded_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT (NOW() AT TIME ZONE 'UTC')
);


CREATE TABLE IF NOT EXISTS orders_updating
(
    number TEXT REFERENCES orders (number) PRIMARY KEY,
    user_id INTEGER NOT NULL,
    deadline TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT (NOW() AT TIME ZONE 'UTC' + INTERVAL '1 hour')
);
