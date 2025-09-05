-- 删除ephemeral_urls表的用户关联
ALTER TABLE ephemeral_urls DROP COLUMN IF EXISTS user_id;