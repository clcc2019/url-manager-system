-- 回滚：重新创建系统配置表（虽然我们不再使用它）
CREATE TABLE IF NOT EXISTS system_config (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    k8s_config TEXT,
    api_base_url TEXT,
    database_url TEXT,
    environment TEXT NOT NULL DEFAULT 'development',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);
