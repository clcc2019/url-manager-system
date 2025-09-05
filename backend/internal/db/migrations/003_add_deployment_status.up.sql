-- 添加新的状态值：draft（草稿状态，未部署）
ALTER TABLE ephemeral_urls DROP CONSTRAINT IF EXISTS chk_status;
ALTER TABLE ephemeral_urls ADD CONSTRAINT chk_status 
    CHECK (status IN ('draft', 'creating', 'active', 'deleting', 'deleted', 'failed'));

-- 添加部署状态字段
ALTER TABLE ephemeral_urls ADD COLUMN IF NOT EXISTS deployed BOOLEAN DEFAULT FALSE;
ALTER TABLE ephemeral_urls ADD COLUMN IF NOT EXISTS deployment_requested_at TIMESTAMP WITH TIME ZONE;

-- 创建索引
CREATE INDEX IF NOT EXISTS idx_ephemeral_urls_deployed ON ephemeral_urls(deployed);
