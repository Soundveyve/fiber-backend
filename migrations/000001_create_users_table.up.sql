-- Создание таблицы пользователей
-- Миграция использует IF NOT EXISTS для идемпотентности
-- Это значит что миграцию можно запускать многократно без ошибок

CREATE TABLE IF NOT EXISTS users (
    -- id - первичный ключ с автоинкрементом
    -- SERIAL в PostgreSQL это shorthand для INTEGER с автоинкрементом
    id SERIAL PRIMARY KEY,
    
    -- email должен быть уникальным для каждого пользователя
    -- NOT NULL означает что поле обязательно для заполнения
    email VARCHAR(255) NOT NULL UNIQUE,
    
    -- имя пользователя
    username VARCHAR(100) NOT NULL UNIQUE,
    
    -- хешированный пароль
    -- Никогда не храните пароли в открытом виде!
    password_hash VARCHAR(255) NOT NULL,
    
    -- имя и фамилия (опциональные поля)
    first_name VARCHAR(100),
    last_name VARCHAR(100),
    
    -- флаг активности пользователя
    -- DEFAULT TRUE означает что новые пользователи активны по умолчанию
    is_active BOOLEAN DEFAULT TRUE,
    
    -- временные метки создания и обновления
    -- CURRENT_TIMESTAMP автоматически устанавливает текущее время
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Создаем индексы для ускорения поиска
-- Индексы ускоряют SELECT запросы, но замедляют INSERT/UPDATE
-- Индекс на email т.к. часто будем искать пользователей по email
CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);

-- Индекс на username для быстрого поиска по имени пользователя
CREATE INDEX IF NOT EXISTS idx_users_username ON users(username);

-- Индекс на created_at для сортировки по дате
CREATE INDEX IF NOT EXISTS idx_users_created_at ON users(created_at DESC);

-- Комментарии к таблице и полям для документации
COMMENT ON TABLE users IS 'Таблица пользователей системы';
COMMENT ON COLUMN users.id IS 'Уникальный идентификатор пользователя';
COMMENT ON COLUMN users.email IS 'Email адрес пользователя (уникальный)';
COMMENT ON COLUMN users.username IS 'Имя пользователя (уникальное)';
COMMENT ON COLUMN users.password_hash IS 'Хеш пароля (bcrypt)';
COMMENT ON COLUMN users.is_active IS 'Флаг активности пользователя';
COMMENT ON COLUMN users.created_at IS 'Дата и время создания записи';
COMMENT ON COLUMN users.updated_at IS 'Дата и время последнего обновления';
