INSERT INTO users (login, password_hash)
VALUES ($1, $2)
ON CONFLICT (login) DO NOTHING
RETURNING id;
