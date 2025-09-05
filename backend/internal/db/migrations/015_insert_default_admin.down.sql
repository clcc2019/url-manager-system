-- 删除默认管理员用户（注意：这会级联删除相关的项目、URL和模版）
DELETE FROM users WHERE username = 'admin';