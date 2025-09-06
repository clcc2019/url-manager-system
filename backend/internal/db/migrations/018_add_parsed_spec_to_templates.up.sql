-- 添加parsed_spec字段到app_templates表，用于存储解析后的结构化数据
ALTER TABLE app_templates ADD COLUMN parsed_spec JSONB;

-- 为parsed_spec字段创建索引（可选，用于查询优化）
CREATE INDEX IF NOT EXISTS idx_app_templates_parsed_spec ON app_templates USING gin(parsed_spec);

-- 为现有记录设置默认的空JSON对象
UPDATE app_templates SET parsed_spec = '{}'::jsonb WHERE parsed_spec IS NULL;
