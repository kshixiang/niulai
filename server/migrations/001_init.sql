-- 数字权益充值分销系统 建表 SQL (MySQL 8.0)
-- 字符集 utf8mb4，存储引擎 InnoDB
--
-- 由 docker-compose 在 mysql 数据卷首次初始化时执行（/docker-entrypoint-initdb.d）。
-- 也可手工执行：mysql -uroot -p niulai < server/migrations/001_init.sql
-- 表结构变更请新增 002_*.sql，不要直接修改本文件。
SET NAMES utf8mb4;
USE niulai;
-- ============================================================
-- 1. 用户表
-- ============================================================
CREATE TABLE `users` (
  `id`         BIGINT       NOT NULL AUTO_INCREMENT COMMENT '用户ID',
  `phone`      VARCHAR(20)  NOT NULL DEFAULT '' COMMENT '手机号',
  `password`   VARCHAR(255) NOT NULL DEFAULT '' COMMENT '加密密码',
  `nickname`   VARCHAR(64)  NOT NULL DEFAULT '' COMMENT '昵称',
  `status`     TINYINT      NOT NULL DEFAULT 1 COMMENT '状态：1正常 0封禁',
  `created_at` DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  `deleted_at` DATETIME              DEFAULT NULL COMMENT '软删除时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_phone` (`phone`),
  KEY `idx_status` (`status`),
  KEY `idx_deleted_at` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户表';
-- ============================================================
-- 2. 代理关系表
-- ============================================================
CREATE TABLE `agents` (
  `id`          BIGINT      NOT NULL AUTO_INCREMENT COMMENT '主键',
  `user_id`     BIGINT      NOT NULL COMMENT '用户ID(关联users.id)',
  `parent_id`   BIGINT      NOT NULL DEFAULT 0 COMMENT '上级代理ID，顶级为0',
  `path`        VARCHAR(255) NOT NULL DEFAULT '0' COMMENT '层级路径，如 0,1,5',
  `level`       TINYINT     NOT NULL DEFAULT 1 COMMENT '代理等级',
  `invite_code` VARCHAR(32) NOT NULL DEFAULT '' COMMENT '邀请码',
  `created_at`  DATETIME    NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at`  DATETIME    NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  `deleted_at`  DATETIME             DEFAULT NULL COMMENT '软删除时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_user_id` (`user_id`),
  UNIQUE KEY `uk_invite_code` (`invite_code`),
  KEY `idx_parent_id` (`parent_id`),
  KEY `idx_path` (`path`),
  KEY `idx_level` (`level`),
  KEY `idx_deleted_at` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='代理关系表';
-- ============================================================
-- 3. 资金账户表
-- ============================================================
CREATE TABLE `wallets` (
  `id`             BIGINT       NOT NULL AUTO_INCREMENT COMMENT '主键',
  `user_id`        BIGINT       NOT NULL COMMENT '用户ID(关联users.id)',
  `balance`        DECIMAL(10,2) NOT NULL DEFAULT 0.00 COMMENT '可用余额',
  `frozen_balance` DECIMAL(10,2) NOT NULL DEFAULT 0.00 COMMENT '冻结余额',
  `version`        INT          NOT NULL DEFAULT 1 COMMENT '乐观锁版本号',
  `created_at`     DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at`     DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_user_id` (`user_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='资金账户表';
-- ============================================================
-- 4. 商品表
-- ============================================================
CREATE TABLE `products` (
  `id`           BIGINT       NOT NULL AUTO_INCREMENT COMMENT '商品ID',
  `name`         VARCHAR(128) NOT NULL DEFAULT '' COMMENT '商品名称，如“话费1”',
  `type`         VARCHAR(32)  NOT NULL DEFAULT '' COMMENT '类型：话费、卡券、会员',
  `cover_image`  VARCHAR(512) NOT NULL DEFAULT '' COMMENT '封面图',
  `sale_price`   DECIMAL(10,2) NOT NULL DEFAULT 0.00 COMMENT '销售价',
  `cost_price`   DECIMAL(10,2) NOT NULL DEFAULT 0.00 COMMENT '成本价',
  `status`       TINYINT      NOT NULL DEFAULT 1 COMMENT '上下架状态：1上架 0下架',
  `created_at`   DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at`   DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  `deleted_at`   DATETIME              DEFAULT NULL COMMENT '软删除时间',
  PRIMARY KEY (`id`),
  KEY `idx_type_status` (`type`, `status`),
  KEY `idx_deleted_at` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='商品表';
-- ============================================================
-- 5. 上游渠道表
-- ============================================================
CREATE TABLE `channels` (
  `id`         BIGINT       NOT NULL AUTO_INCREMENT COMMENT '渠道ID',
  `name`       VARCHAR(128) NOT NULL DEFAULT '' COMMENT '渠道名称，如“渠道A”',
  `api_url`    VARCHAR(512) NOT NULL DEFAULT '' COMMENT '接口地址',
  `api_key`    VARCHAR(255) NOT NULL DEFAULT '' COMMENT '接口密钥',
  `status`     TINYINT      NOT NULL DEFAULT 1 COMMENT '健康状态：1可用 0熔断',
  `created_at` DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  `deleted_at` DATETIME              DEFAULT NULL COMMENT '软删除时间',
  PRIMARY KEY (`id`),
  KEY `idx_status` (`status`),
  KEY `idx_deleted_at` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='上游渠道表';
-- ============================================================
-- 6. 订单表
-- ============================================================
CREATE TABLE `orders` (
  `id`           BIGINT       NOT NULL AUTO_INCREMENT COMMENT '订单ID',
  `order_no`     VARCHAR(64)  NOT NULL DEFAULT '' COMMENT '唯一订单号',
  `user_id`      BIGINT       NOT NULL COMMENT '用户ID(关联users.id)',
  `product_id`   BIGINT       NOT NULL COMMENT '商品ID(关联products.id)',
  `channel_id`   BIGINT       NOT NULL DEFAULT 0 COMMENT '渠道ID(关联channels.id)',
  `amount`       DECIMAL(10,2) NOT NULL DEFAULT 0.00 COMMENT '实付金额',
  `cost_amount`  DECIMAL(10,2) NOT NULL DEFAULT 0.00 COMMENT '渠道成本',
  `profit`       DECIMAL(10,2) NOT NULL DEFAULT 0.00 COMMENT '利润',
  `status`       TINYINT      NOT NULL DEFAULT 0 COMMENT '0待支付 1已支付 2充值中 3成功 4失败 5已退款',
  `account`      VARCHAR(128) NOT NULL DEFAULT '' COMMENT '充值账号',
  `callback_at`  DATETIME              DEFAULT NULL COMMENT '回调时间',
  `created_at`   DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at`   DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  `deleted_at`   DATETIME              DEFAULT NULL COMMENT '软删除时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_order_no` (`order_no`),
  KEY `idx_user_status` (`user_id`, `status`),
  KEY `idx_product_id` (`product_id`),
  KEY `idx_channel_id` (`channel_id`),
  KEY `idx_status` (`status`),
  KEY `idx_account` (`account`),
  KEY `idx_deleted_at` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='订单表';
-- ============================================================
-- 7. 分润流水表
-- ============================================================
CREATE TABLE `commission_records` (
  `id`         BIGINT       NOT NULL AUTO_INCREMENT COMMENT '主键',
  `order_id`   BIGINT       NOT NULL COMMENT '订单ID(关联orders.id)',
  `user_id`    BIGINT       NOT NULL COMMENT '获得分润的代理(关联users.id)',
  `amount`     DECIMAL(10,2) NOT NULL DEFAULT 0.00 COMMENT '分润金额',
  `level`      TINYINT      NOT NULL DEFAULT 1 COMMENT '分润层级',
  `status`     TINYINT      NOT NULL DEFAULT 0 COMMENT '0待结算 1已结算',
  `created_at` DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  KEY `idx_order_id` (`order_id`),
  KEY `idx_user_status` (`user_id`, `status`),
  KEY `idx_level` (`level`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='分润流水表';
