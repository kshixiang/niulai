-- 006 商品定价改造：新增数字折扣(几折)与商品价格类别(面值)子表。
-- 仅对已有数据卷手动执行；首次初始化时 docker 会自动按文件名顺序执行本文件。
-- 与 internal/model/product.go、internal/model/product_denomination.go 中的字段定义保持一致。
SET NAMES utf8mb4;
USE niulai;

-- 折扣按“几折”录入：6 表示 6 折(付 60%，即 ×0.6)，折后价 = 面值 × 折扣 ÷ 10。
-- 取值范围 0 < discount <= 10，默认 10 表示不打折。
ALTER TABLE `products`
  ADD COLUMN `discount` DECIMAL(4,2) NOT NULL DEFAULT 10.00 COMMENT '折扣(几折) 如6表示6折 折后价=面值×折扣÷10' AFTER `discount_label`;

-- 商品价格类别(面值)：一个商品可配置多个面值，共用商品级折扣。
CREATE TABLE `product_denominations` (
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
