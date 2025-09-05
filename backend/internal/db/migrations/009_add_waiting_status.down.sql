-- 恢复到原来的状态约束（不包含waiting）
ALTER TABLE ephemeral_urls DROP CONSTRAINT IF EXISTS chk_status;

ALTER TABLE ephemeral_urls ADD CONSTRAINT chk_status 
    CHECK (status IN ('creating', 'active', 'deleting', 'deleted', 'failed'));