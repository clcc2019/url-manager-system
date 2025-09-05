-- 删除app_templates表的用户关联
ALTER TABLE app_templates DROP COLUMN IF EXISTS user_id;