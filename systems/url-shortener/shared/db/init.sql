-- Schema for the URL shortener. Runs once when the postgres volume is created.
CREATE TABLE IF NOT EXISTS urls (
    code       VARCHAR(16) PRIMARY KEY,
    target     TEXT        NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
