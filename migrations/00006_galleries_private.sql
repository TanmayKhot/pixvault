-- +goose Up
-- +goose StatementBegin
ALTER TABLE galleries
ADD COLUMN isprivate BOOLEAN DEFAULT TRUE;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE galleries
DROP COLUMN isprivate;
-- +goose StatementEnd
