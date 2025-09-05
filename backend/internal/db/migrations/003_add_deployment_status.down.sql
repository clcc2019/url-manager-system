-- 回滚部署状态相关字段
DROP INDEX IF EXISTS idx_ephemeral_urls_deployed;
ALTER TABLE ephemeral_urls DROP COLUMN IF EXISTS deployment_requested_at;
ALTER TABLE ephemeral_urls DROP COLUMN IF EXISTS deployed;

-- 恢复原始状态约束
ALTER TABLE ephemeral_urls DROP CONSTRAINT IF EXISTS chk_status;
ALTER TABLE ephemeral_urls ADD CONSTRAINT chk_status 
    CHECK (status IN ('creating', 'active', 'deleting', 'deleted', 'failed'));
