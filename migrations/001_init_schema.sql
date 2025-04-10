-- +goose Up
-- +goose StatementBegin
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email TEXT NOT NULL UNIQUE,
    password TEXT NOT NULL,
    role TEXT NOT NULL CHECK (role IN ('employee', 'moderator'))
);

CREATE TABLE pvz (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    registration_date TIMESTAMP NOT NULL DEFAULT NOW(),
    city TEXT NOT NULL CHECK (city IN ('Москва', 'Санкт-Петербург', 'Казань'))
);

CREATE TABLE receptions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    date_time TIMESTAMP NOT NULL DEFAULT NOW(),
    pvz_id UUID REFERENCES pvz(id),
    status TEXT NOT NULL CHECK (status IN ('in_progress', 'close'))
);

CREATE TABLE products (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    date_time TIMESTAMP NOT NULL DEFAULT NOW(),
    type TEXT NOT NULL CHECK (type IN ('электроника', 'одежда', 'обувь')),
    reception_id UUID REFERENCES receptions(id)
);
-- +goose StatementEnd