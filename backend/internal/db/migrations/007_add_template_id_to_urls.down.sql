-- 删除ephemeral_urls表的template_id字段
DROP INDEX IF EXISTS idx_ephemeral_urls_template_id;
ALTER TABLE ephemeral_urls DROP COLUMN IF EXISTS template_id;