-- 添加container_config字段到ephemeral_urls表
ALTER TABLE ephemeral_urls ADD COLUMN container_config JSONB DEFAULT '{}'::jsonb;

-- 为container_config字段创建索引（如果需要查询）
CREATE INDEX IF NOT EXISTS idx_ephemeral_urls_container_config ON ephemeral_urls USING gin(container_config);
