CREATE TABLE IF NOT EXISTS transactions (
     id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
     user_id BIGINT NOT NULL,
     from_acc_id BIGINT,
     to_acc_id BIGINT,
     type TEXT NOT NULL CHECK (type IN ('expense', 'income', 'transfer')),
     amount BIGINT NOT NULL CHECK (amount > 0),
     description TEXT NOT NULL DEFAULT '',
     occurred_at TIMESTAMPTZ NOT NULL,
     created_at TIMESTAMPTZ NOT NULL DEFAULT now(),

     CONSTRAINT fk_transactions_user
          FOREIGN KEY (user_id)
          REFERENCES users(id) ON DELETE CASCADE,
     CONSTRAINT fk_transactions_from_acc
          FOREIGN KEY (from_acc_id)
          REFERENCES accounts(id),
     CONSTRAINT fk_transactions_to_acc
          FOREIGN KEY (to_acc_id)
          REFERENCES accounts(id),

     CONSTRAINT chk_transaction_accounts CHECK (
          (type = 'expense' AND from_acc_id IS NOT NULL AND to_acc_id IS NULL) OR
          (type = 'income' AND from_acc_id IS NULL AND to_acc_id IS NOT NULL) OR
          (type = 'transfer' AND from_acc_id IS NOT NULL AND to_acc_id IS NOT NULL AND from_acc_id <> to_acc_id)
     )
);