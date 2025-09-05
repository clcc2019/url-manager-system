-- 添加ingress_host字段到ephemeral_urls表
ALTER TABLE ephemeral_urls ADD COLUMN ingress_host TEXT;
