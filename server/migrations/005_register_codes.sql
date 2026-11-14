-- 005 注册码：由管理后台批量生成，注册时校验。
-- 改为幂等：表已存在则跳过建表。
SET NAMES utf8mb4;
USE niulai;

CREATE TABLE IF NOT EXISTS `register_codes` (
  `id`         BIGINT      NOT NULL AUTO_INCREMENT COMMENT '主键',
  `code`       VARCHAR(32) NOT NULL COMMENT '注册码',
  `inviter_id` BIGINT      NOT NULL DEFAULT 0 COMMENT '邀请人用户ID，0表示不绑定',
  `status`     TINYINT     NOT NULL DEFAULT 1 COMMENT '1可用 0禁用',
  `created_at` DATETIME    NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` DATETIME    NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_code` (`code`),
  KEY `idx_inviter_id` (`inviter_id`),
  KEY `idx_status` (`status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='注册码表';
