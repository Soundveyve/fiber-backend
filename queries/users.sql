-- name: CreateUser :one
-- Создание нового пользователя
-- :one означает что запрос вернет одну строку
-- RETURNING * возвращает все поля созданной записи
INSERT INTO users (
    email,
    username,
    password_hash,
    first_name,
    last_name
) VALUES (
    $1, $2, $3, $4, $5
) RETURNING *;

-- name: GetUserByID :one
-- Получение пользователя по ID
-- sqlc автоматически создаст функцию с параметром типа int
SELECT * FROM users
WHERE id = $1 LIMIT 1;

-- name: GetUserByEmail :one
-- Получение пользователя по email
-- Используется для аутентификации
SELECT * FROM users
WHERE email = $1 LIMIT 1;

-- name: GetUserByUsername :one
-- Получение пользователя по username
SELECT * FROM users
WHERE username = $1 LIMIT 1;

-- name: ListUsers :many
-- Получение списка пользователей с пагинацией
-- :many означает что запрос вернет массив записей
-- $1 - limit (количество записей)
-- $2 - offset (смещение для пагинации)
SELECT * FROM users
ORDER BY created_at DESC
LIMIT $1 OFFSET $2;

-- name: UpdateUser :one
-- Обновление данных пользователя
-- COALESCE используется для обновления только переданных полей
-- Если значение NULL, оставляем старое значение
UPDATE users
SET
    email = COALESCE($2, email),
    username = COALESCE($3, username),
    first_name = COALESCE($4, first_name),
    last_name = COALESCE($5, last_name),
    is_active = COALESCE($6, is_active),
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1
RETURNING *;

-- name: UpdateUserPassword :exec
-- Обновление пароля пользователя
-- :exec означает что запрос не возвращает данных
UPDATE users
SET
    password_hash = $2,
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1;

-- name: DeleteUser :exec
-- Удаление пользователя (физическое удаление)
-- В production часто используют soft delete (is_active = false)
DELETE FROM users
WHERE id = $1;

-- name: DeactivateUser :exec
-- Деактивация пользователя (soft delete)
-- Предпочтительный способ "удаления" в production
UPDATE users
SET
    is_active = false,
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1;

-- name: CountUsers :one
-- Подсчет общего количества пользователей
-- Полезно для пагинации
SELECT COUNT(*) FROM users;

-- name: CountActiveUsers :one
-- Подсчет активных пользователей
SELECT COUNT(*) FROM users
WHERE is_active = true;
