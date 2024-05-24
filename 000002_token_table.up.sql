CREATE TABLE IF NOT EXISTS tokens (
    hash VARCHAR(255) PRIMARY KEY,
    user_id bigint NOT NULL REFERENCES users ON DELETE CASCADE,
    expiry timestamp(0) with time zone NOT NULL
);