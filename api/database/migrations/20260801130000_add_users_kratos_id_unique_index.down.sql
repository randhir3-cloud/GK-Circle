-- +migrate Down
DROP INDEX IF EXISTS users_kratos_id_unique;
