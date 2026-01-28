-- Откат миграции - удаление таблицы пользователей
-- Эта миграция используется когда нужно откатить изменения

-- Сначала удаляем индексы
-- IF EXISTS предотвращает ошибку если индекс уже удален
DROP INDEX IF EXISTS idx_users_created_at;
DROP INDEX IF EXISTS idx_users_username;
DROP INDEX IF EXISTS idx_users_email;

-- Затем удаляем саму таблицу
-- CASCADE автоматически удалит все зависимые объекты
DROP TABLE IF EXISTS users CASCADE;
