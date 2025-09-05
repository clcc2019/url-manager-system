-- 删除ephemeral_urls表的started_at字段
DROP INDEX IF EXISTS idx_ephemeral_urls_started_at;
ALTER TABLE ephemeral_urls DROP COLUMN IF EXISTS started_at;