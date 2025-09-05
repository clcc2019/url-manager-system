-- 添加logs字段到ephemeral_urls表
ALTER TABLE ephemeral_urls ADD COLUMN logs JSONB DEFAULT '[]'::jsonb;

-- 为logs字段创建索引（可选，用于查询包含特定日志条目的记录）
CREATE INDEX IF NOT EXISTS idx_ephemeral_urls_logs ON ephemeral_urls USING gin(logs);
