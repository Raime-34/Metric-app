-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';
-- +goose StatementEnd
CREATE TABLE metrics (
    id TEXT PRIMARY KEY,
    mtype TEXT NOT NULL,
    delta BIGINT,
    value DOUBLE PRECISION,
    hash TEXT,
    CHECK (
        (mtype = 'counter' AND delta IS NOT NULL)
     OR (mtype = 'gauge'   AND value IS NOT NULL)
    )
);


-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
-- +goose StatementEnd
