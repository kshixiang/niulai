-- 004 系统设置：注册码是否强制填写。
-- 改为幂等：表已存在则跳过建表；配置已存在则忽略插入。
SET NAMES utf8mb4;
USE niulai;

CREATE TABLE IF NOT EXISTS `settings` (
  `id`           BIGINT       NOT NULL AUTO_INCREMENT COMMENT '主键',
  `config_key`   VARCHAR(64)  NOT NULL COMMENT '配置键',
  `config_value` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '配置值',
  `created_at`   DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at`   DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_config_key` (`config_key`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='系统设置表';

INSERT IGNORE INTO `settings` (`config_key`, `config_value`) VALUES ('register_code_required', '0');
