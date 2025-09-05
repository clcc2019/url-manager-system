-- 删除container_config字段和索引
DROP INDEX IF EXISTS idx_ephemeral_urls_container_config;
ALTER TABLE ephemeral_urls DROP COLUMN IF EXISTS container_config;
