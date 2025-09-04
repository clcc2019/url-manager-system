-- 删除projects表的索引
DROP INDEX IF EXISTS idx_projects_created_at;
DROP INDEX IF EXISTS idx_projects_name;

-- 删除projects表
DROP TABLE IF EXISTS projects;