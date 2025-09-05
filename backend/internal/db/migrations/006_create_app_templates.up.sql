-- app_templates表：存储应用模版配置
CREATE TABLE IF NOT EXISTS app_templates (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT NOT NULL UNIQUE,
    description TEXT,
    yaml_spec TEXT NOT NULL, -- 存储 Deployment/Pod/Service 等 YAML 定义
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- 为app_templates表创建索引
CREATE INDEX IF NOT EXISTS idx_app_templates_name ON app_templates(name);
CREATE INDEX IF NOT EXISTS idx_app_templates_created_at ON app_templates(created_at);

-- 添加模版名称长度检查约束
ALTER TABLE app_templates ADD CONSTRAINT chk_template_name_length 
    CHECK (length(name) >= 1 AND length(name) <= 100);

-- 添加YAML规范非空检查约束
ALTER TABLE app_templates ADD CONSTRAINT chk_yaml_spec_not_empty 
    CHECK (length(trim(yaml_spec)) > 0);