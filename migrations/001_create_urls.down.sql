-- 001_create_urls.down.sql
-- Reverte tudo que o .up.sql criou, na ordem inversa.

DROP INDEX IF EXISTS idx_urls_expired;
DROP INDEX IF EXISTS idx_urls_code;
DROP TABLE IF EXISTS urls;