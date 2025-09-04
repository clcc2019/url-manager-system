-- ephemeral_urls表
CREATE TABLE IF NOT EXISTS ephemeral_urls (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    path TEXT NOT NULL,
    image TEXT NOT NULL,
    env JSONB,
    replicas INTEGER NOT NULL DEFAULT 1,
    resources JSONB,
    status TEXT NOT NULL DEFAULT 'creating',
    k8s_deployment_name TEXT,
    k8s_service_name TEXT,
    k8s_secret_name TEXT,
    error_message TEXT,
    expire_at TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE (project_id, path)
);

-- 为ephemeral_urls表创建索引
CREATE INDEX IF NOT EXISTS idx_ephemeral_urls_project_id ON ephemeral_urls(project_id);
CREATE INDEX IF NOT EXISTS idx_ephemeral_urls_path ON ephemeral_urls(path);
CREATE INDEX IF NOT EXISTS idx_ephemeral_urls_status ON ephemeral_urls(status);
CREATE INDEX IF NOT EXISTS idx_ephemeral_urls_expire_at ON ephemeral_urls(expire_at);
CREATE INDEX IF NOT EXISTS idx_ephemeral_urls_created_at ON ephemeral_urls(created_at);

-- 为状态值添加检查约束
ALTER TABLE ephemeral_urls ADD CONSTRAINT chk_status 
    CHECK (status IN ('creating', 'active', 'deleting', 'deleted', 'failed'));