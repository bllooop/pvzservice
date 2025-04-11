-- +goose Up
-- +goose StatementBegin
CREATE TYPE role_enum AS ENUM ('employee', 'moderator');
CREATE TABLE userlist
(
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email varchar(255) NOT NULL unique,
    password varchar(255) NOT NULL,
    role role_enum NOT NULL
); 
CREATE EXTENSION IF NOT EXISTS pgcrypto;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP userlist;
-- +goose StatementEnd
