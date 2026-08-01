-- +migrate Up
CREATE UNIQUE INDEX IF NOT EXISTS users_kratos_id_unique
    ON users (kratos_id)
    WHERE kratos_id IS NOT NULL;
