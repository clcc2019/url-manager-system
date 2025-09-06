-- 删除parsed_spec字段的索引
DROP INDEX IF EXISTS idx_app_templates_parsed_spec;

-- 删除parsed_spec字段
ALTER TABLE app_templates DROP COLUMN IF EXISTS parsed_spec;
