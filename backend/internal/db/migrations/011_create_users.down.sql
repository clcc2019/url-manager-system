-- 删除用户表的触发器
DROP TRIGGER IF EXISTS update_users_updated_at ON users;

-- 删除更新时间函数（如果没有其他表使用）
DROP FUNCTION IF EXISTS update_updated_at_column();

-- 删除用户表
DROP TABLE IF EXISTS users;