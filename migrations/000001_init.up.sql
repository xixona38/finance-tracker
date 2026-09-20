CREATE TABLE IF NOT EXISTS accounts (
     id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
     name TEXT NOT NULL CHECK (char_length(name) BETWEEN 1 AND 100),
     type TEXT NOT NULL CHECK (type IN ('card', 'cash', 'credit', 'savings')),
     currency CHAR(3) NOT NULL,
     initial_balance BIGINT NOT NULL DEFAULT 0,
     created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);