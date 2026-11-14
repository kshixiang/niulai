-- 007 商品双折扣与下单填写项：新增渠道折扣列、商品下单表单字段配置子表。
-- 与 internal/model/product.go、internal/model/product_order_field.go 中的字段定义保持一致。
-- 改为幂等：列已存在则跳过；表已存在则跳过建表。
SET NAMES utf8mb4;
USE niulai;

SET @db = DATABASE();
SET @tbl = 'products';
SET @col = 'channel_discount';
SET @sql = (SELECT IF(
  (SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_SCHEMA=@db AND TABLE_NAME=@tbl AND COLUMN_NAME=@col)=0,
  'ALTER TABLE `products` ADD COLUMN `channel_discount` DECIMAL(4,2) NOT NULL DEFAULT 10.00 COMMENT ''渠道折扣(几折) 我的成本 渠道价=面值×渠道折扣÷10'' AFTER `discount`',
  'SELECT 1'));
PREPARE s FROM @sql; EXECUTE s; DEALLOCATE PREPARE s;

CREATE TABLE IF NOT EXISTS `product_order_fields` (
  `id`          BIGINT       NOT NULL AUTO_INCREMENT COMMENT '主键',
  `product_id`  BIGINT       NOT NULL DEFAULT 0  COMMENT '商品ID(关联products.id)',
  `field_key`   VARCHAR(64)  NOT NULL DEFAULT '' COMMENT '字段英文标识 同一商品内唯一 如 phone',
  `label`       VARCHAR(128) NOT NULL DEFAULT '' COMMENT '中文名 如 手机号',
  `type`        VARCHAR(16)  NOT NULL DEFAULT 'text' COMMENT 'text/number/select',
  `required`    TINYINT      NOT NULL DEFAULT 0  COMMENT '1必填 0选填',
  `placeholder` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '输入提示语',
  `options`     TEXT                  COMMENT 'select 选项 JSON数组字符串',
  `sort`        INT          NOT NULL DEFAULT 0  COMMENT '展示排序(升序)',
  `created_at`  DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at`  DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  `deleted_at`  DATETIME              DEFAULT NULL COMMENT '软删除时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_product_field` (`product_id`, `field_key`),
  KEY `idx_deleted_at` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='商品下单填写项配置表';
