-- 006 商品定价改造：新增数字折扣(几折)与商品价格类别(面值)子表。
-- 与 internal/model/product.go、internal/model/product_denomination.go 中的字段定义保持一致。
-- 改为幂等：列已存在则跳过；表已存在则跳过建表。
SET NAMES utf8mb4;
USE niulai;

SET @db = DATABASE();
SET @tbl = 'products';
SET @col = 'discount';
SET @sql = (SELECT IF(
  (SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_SCHEMA=@db AND TABLE_NAME=@tbl AND COLUMN_NAME=@col)=0,
  'ALTER TABLE `products` ADD COLUMN `discount` DECIMAL(4,2) NOT NULL DEFAULT 10.00 COMMENT ''折扣(几折) 如6表示6折 折后价=面值×折扣÷10'' AFTER `discount_label`',
  'SELECT 1'));
PREPARE s FROM @sql; EXECUTE s; DEALLOCATE PREPARE s;

CREATE TABLE IF NOT EXISTS `product_denominations` (
  `id`          BIGINT        NOT NULL AUTO_INCREMENT COMMENT '主键',
  `product_id`  BIGINT        NOT NULL DEFAULT 0  COMMENT '商品ID(关联products.id)',
  `name`        VARCHAR(64)   NOT NULL DEFAULT '' COMMENT '价格类别名称(可选) 如200元档',
  `face_value`  DECIMAL(10,2) NOT NULL DEFAULT 0.00 COMMENT '面值/原价',
  `sort`        INT           NOT NULL DEFAULT 0  COMMENT '展示排序(升序)',
  `created_at`  DATETIME      NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at`  DATETIME      NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  `deleted_at`  DATETIME               DEFAULT NULL COMMENT '软删除时间',
  PRIMARY KEY (`id`),
  KEY `idx_product_id` (`product_id`),
  KEY `idx_deleted_at` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='商品价格类别(面值)表';
