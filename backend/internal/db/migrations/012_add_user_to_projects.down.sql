-- 删除项目表的用户关联
ALTER TABLE projects DROP COLUMN IF EXISTS user_id;