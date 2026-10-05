CREATE TABLE IF NOT EXISTS sessions (
     token_hash TEXT PRIMARY KEY,
     user_id BIGINT NOT NULL,
     created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
     expires_at TIMESTAMPTZ NOT NULL CHECK (expires_at > created_at),

     CONSTRAINT fk_sessions
     FOREIGN KEY (user_id)
     REFERENCES users(id) ON DELETE CASCADE
);
