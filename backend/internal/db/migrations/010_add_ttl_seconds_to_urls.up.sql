-- 为ephemeral_urls表添加ttl_seconds字段，用于存储TTL值
ALTER TABLE ephemeral_urls 
ADD COLUMN ttl_seconds INTEGER NOT NULL DEFAULT 3600;

-- 创建ttl_seconds字段索引
CREATE INDEX IF NOT EXISTS idx_ephemeral_urls_ttl_seconds ON ephemeral_urls(ttl_seconds);