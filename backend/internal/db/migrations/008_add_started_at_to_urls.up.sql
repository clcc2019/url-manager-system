-- 为ephemeral_urls表添加started_at字段，用于记录Pod Ready的时间
ALTER TABLE ephemeral_urls 
ADD COLUMN started_at TIMESTAMP WITH TIME ZONE;

-- 创建started_at字段索引
CREATE INDEX IF NOT EXISTS idx_ephemeral_urls_started_at ON ephemeral_urls(started_at);

-- 为现有记录设置started_at = created_at（保持向后兼容）
UPDATE ephemeral_urls 
SET started_at = created_at 
WHERE status = 'active' AND started_at IS NULL;