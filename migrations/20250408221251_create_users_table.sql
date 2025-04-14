-- +goose Up
-- +goose StatementBegin
DO $$ 
BEGIN 
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'role_enum') THEN 
        CREATE TYPE role_enum AS ENUM ('employee', 'moderator'); 
    END IF;

    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'reception_status_enum') THEN 
        CREATE TYPE reception_status_enum AS ENUM ('in_progress', 'close'); 
    END IF;

    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'product_type_enum') THEN 
        CREATE TYPE product_type_enum AS ENUM ('электроника', 'одежда', 'обувь'); 
    END IF;
END; $$;
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
DROP TYPE IF EXISTS role_enum;
DROP TABLE userlist;
-- +goose StatementEnd
