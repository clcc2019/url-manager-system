-- 移除ingress_host字段从ephemeral_urls表
ALTER TABLE ephemeral_urls DROP COLUMN ingress_host;
