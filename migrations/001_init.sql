-- migrations/001_init.sql
-- Таблица пользователей
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    username VARCHAR(255) UNIQUE NOT NULL,
    password_hash TEXT NOT NULL CHECK (LENGTH(password_hash) >= 60),
    coins INT NOT NULL DEFAULT 1000 CHECK (coins >= 0)
);
-- Таблица товаров
CREATE TABLE merch_items (
    name VARCHAR(50) PRIMARY KEY,
    price INT NOT NULL CHECK (price > 0)
);
-- Таблица инвентаря пользователя
CREATE TABLE inventory (
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    item_name VARCHAR(50) REFERENCES merch_items(name) ON DELETE CASCADE,
    quantity INT NOT NULL DEFAULT 0 CHECK (quantity >= 0),
    PRIMARY KEY (user_id, item_name)
);
-- История переводов монет
CREATE TABLE transfer_history (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    from_user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    to_user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    amount INT NOT NULL CHECK (amount > 0),
    timestamp TIMESTAMP DEFAULT NOW(),
    CHECK (from_user_id <> to_user_id)
);
-- Индексы для ускорения поиска
CREATE INDEX idx_users_username ON users(username);
CREATE INDEX idx_merch_items_name ON merch_items(name);