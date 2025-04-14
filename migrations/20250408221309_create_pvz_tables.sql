-- +goose Up
-- +goose StatementBegin
DO $$ 
BEGIN 
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'city_enum') THEN 
        CREATE TYPE city_enum AS ENUM ('Москва', 'Санкт-Петербург', 'Казань'); 
    END IF;

    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'reception_status_enum') THEN 
        CREATE TYPE reception_status_enum AS ENUM ('in_progress', 'close'); 
    END IF;

    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'product_type_enum') THEN 
        CREATE TYPE product_type_enum AS ENUM ('электроника', 'одежда', 'обувь'); 
    END IF;
END; $$;

CREATE TABLE pvz (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    registrationdate TIMESTAMPTZ,
    city city_enum NOT NULL
);
CREATE TABLE product_reception (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    date_received TIMESTAMPTZ,
    pvz_id UUID REFERENCES pvz(id) ON DELETE CASCADE,
    status_reception reception_status_enum NOT NULL
);
CREATE TABLE product (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    date_received TIMESTAMPTZ,
    type_product product_type_enum NOT NULL,
    reception_id UUID REFERENCES product_reception(id) ON DELETE CASCADE,
    pvz_id UUID REFERENCES pvz(id) ON DELETE CASCADE
);
CREATE EXTENSION IF NOT EXISTS pgcrypto;
CREATE INDEX idx_product_reception_pvz_date ON product_reception(pvz_id, date_received DESC);
CREATE INDEX idx_product_reception_pvz_reception_date ON product(pvz_id, reception_id, date_received DESC);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TYPE IF EXISTS city_enum;
DROP TYPE IF EXISTS reception_status_enum;
DROP TYPE IF EXISTS product_type_enum;
DROP TABLE product;
DROP TABLE product_reception;
DROP TABLE pvz;
-- +goose StatementEnd
