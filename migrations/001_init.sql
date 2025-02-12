-- migrations/001_init.sql
-- Таблица пользователей
CREATE TABLE users (
    username VARCHAR(255) PRIMARY KEY,
    password_hash TEXT NOT NULL CHECK (LENGTH(password_hash) >= 60),
    coins INT NOT NULL DEFAULT 1000 CHECK (coins >= 0)
);
-- Таблица товаров
CREATE TABLE merch_items (
    name VARCHAR(50) PRIMARY KEY,
    price INT NOT NULL CHECK (price > 0)
);
-- Очистка таблицы товаров
TRUNCATE merch_items;
-- Добавление мерча в таблицу
INSERT INTO merch_items (name, price)
VALUES ('t-shirt', 80),
    ('cup', 20),
    ('book', 50),
    ('pen', 10),
    ('powerbank', 200),
    ('hoody', 300),
    ('umbrella', 200),
    ('socks', 10),
    ('wallet', 50),
    ('pink-hoody', 500);
-- Таблица инвентаря (с использованием username вместо user_id)
CREATE TABLE inventory (
    user_name VARCHAR(255) REFERENCES users(username) ON DELETE CASCADE,
    item_name VARCHAR(50) REFERENCES merch_items(name) ON DELETE CASCADE,
    quantity INT NOT NULL DEFAULT 0 CHECK (quantity >= 0),
    PRIMARY KEY (user_name, item_name)
);
-- История переводов монет (с использованием username вместо user_id)
CREATE TABLE transfer_history (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    from_user_name VARCHAR(255) REFERENCES users(username) ON DELETE CASCADE,
    to_user_name VARCHAR(255) REFERENCES users(username) ON DELETE CASCADE,
    amount INT NOT NULL CHECK (amount > 0),
    CHECK (from_user_name <> to_user_name)
);
-- Индексы для ускорения поиска
CREATE INDEX idx_users_username ON users(username);
CREATE INDEX idx_merch_items_name ON merch_items(name);