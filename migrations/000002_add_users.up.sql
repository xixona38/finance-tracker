CREATE TABLE IF NOT EXISTS users (
     id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
     email TEXT NOT NULL UNIQUE,
     password_hash TEXT NOT NULL,
     created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

ALTER TABLE accounts
ADD COLUMN user_id BIGINT NOT NULL;

ALTER TABLE accounts
ADD CONSTRAINT accounts_user_id_fkey
FOREIGN KEY (user_id) REFERENCES users(id);

CREATE INDEX accounts_user_id_idx
ON accounts (user_id);