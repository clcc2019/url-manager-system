-- 移除logs字段从ephemeral_urls表
DROP INDEX IF EXISTS idx_ephemeral_urls_logs;
ALTER TABLE ephemeral_urls DROP COLUMN logs;
