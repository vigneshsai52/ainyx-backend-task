-- +migrate Up
CREATE TABLE IF NOT EXISTS users (
    id   SERIAL PRIMARY KEY,
    name TEXT   NOT NULL,
    dob  DATE   NOT NULL
);

-- +migrate Down
DROP TABLE IF EXISTS users;
