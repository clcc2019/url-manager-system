-- 为项目表添加用户关联
ALTER TABLE projects ADD COLUMN user_id UUID REFERENCES users(id);

-- 创建索引
CREATE INDEX IF NOT EXISTS idx_projects_user_id ON projects(user_id);

-- 为现有项目设置默认用户（可选，后面会创建默认管理员用户）
-- 这里先不设置，等创建用户后再处理