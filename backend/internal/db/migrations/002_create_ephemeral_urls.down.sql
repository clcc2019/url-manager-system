-- 删除ephemeral_urls表的约束
ALTER TABLE ephemeral_urls DROP CONSTRAINT IF EXISTS chk_status;

-- 删除ephemeral_urls表的索引
DROP INDEX IF EXISTS idx_ephemeral_urls_created_at;
DROP INDEX IF EXISTS idx_ephemeral_urls_expire_at;
DROP INDEX IF EXISTS idx_ephemeral_urls_status;
DROP INDEX IF EXISTS idx_ephemeral_urls_path;
DROP INDEX IF EXISTS idx_ephemeral_urls_project_id;

-- 删除ephemeral_urls表
DROP TABLE IF EXISTS ephemeral_urls;