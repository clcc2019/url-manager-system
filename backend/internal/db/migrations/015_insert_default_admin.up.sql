-- 插入默认管理员用户
-- 默认密码是 'admin123'，实际生产环境应该使用更强的密码
-- bcrypt hash for 'admin123' with cost 12
INSERT INTO users (username, password_hash, role, email) VALUES 
('admin', '$2a$12$LQv3c1yqBWVHxkd0LHAkCOYz6TtxMQJqhN8/lewgMt19.CqKjj6SO', 'admin', 'admin@example.com')
ON CONFLICT (username) DO NOTHING;

-- 更新现有项目的用户关联为默认管理员
UPDATE projects SET user_id = (SELECT id FROM users WHERE username = 'admin') WHERE user_id IS NULL;

-- 更新现有URL的用户关联为默认管理员
UPDATE ephemeral_urls SET user_id = (SELECT id FROM users WHERE username = 'admin') WHERE user_id IS NULL;

-- 更新现有模版的用户关联为默认管理员
UPDATE app_templates SET user_id = (SELECT id FROM users WHERE username = 'admin') WHERE user_id IS NULL;