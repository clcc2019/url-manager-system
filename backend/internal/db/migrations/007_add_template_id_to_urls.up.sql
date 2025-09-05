-- 为ephemeral_urls表添加template_id字段
ALTER TABLE ephemeral_urls 
ADD COLUMN template_id UUID REFERENCES app_templates(id) ON DELETE SET NULL;

-- 创建template_id字段索引
CREATE INDEX IF NOT EXISTS idx_ephemeral_urls_template_id ON ephemeral_urls(template_id);