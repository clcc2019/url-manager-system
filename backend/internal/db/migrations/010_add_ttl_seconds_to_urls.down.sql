-- 删除ephemeral_urls表的ttl_seconds字段
DROP INDEX IF EXISTS idx_ephemeral_urls_ttl_seconds;
ALTER TABLE ephemeral_urls DROP COLUMN IF EXISTS ttl_seconds;