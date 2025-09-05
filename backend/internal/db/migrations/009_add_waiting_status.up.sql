-- 先删除现有的状态约束
ALTER TABLE ephemeral_urls DROP CONSTRAINT IF EXISTS chk_status;

-- 添加新的状态约束，包含waiting状态
ALTER TABLE ephemeral_urls ADD CONSTRAINT chk_status 
    CHECK (status IN ('creating', 'waiting', 'active', 'deleting', 'deleted', 'failed'));