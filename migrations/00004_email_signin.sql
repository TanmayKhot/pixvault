-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';
CREATE TABLE email_signin (
  id SERIAL PRIMARY KEY,
  user_id INT UNIQUE REFERENCES users (id) ON DELETE CASCADE,
  token_hash TEXT UNIQUE NOT NULL,
  expires_at TIMESTAMPTZ NOT NULL
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABlE email_signin;
-- +goose StatementEnd
