-- 007 商品双折扣与下单填写项：新增渠道折扣列、商品下单表单字段配置子表。
-- 仅对已有数据卷手动执行；首次初始化时 docker 会自动按文件名顺序执行本文件。
-- 与 internal/model/product.go、internal/model/product_order_field.go 中的字段定义保持一致。
SET NAMES utf8mb4;
USE niulai;

-- 渠道折扣(我的成本)：与 discount(出售折扣) 同几折语义，渠道价=面值×渠道折扣÷10。
-- 取值范围 0 < channel_discount <= 10，默认 10 表示不打折；仅后台可见，禁止出现在公开接口。
ALTER TABLE `products`
  ADD COLUMN `channel_discount` DECIMAL(4,2) NOT NULL DEFAULT 10.00
  COMMENT '渠道折扣(几折) 我的成本 渠道价=面值×渠道折扣÷10' AFTER `discount`;

-- 商品下单填写项：驱动 H5 下单页动态输入字段，按商品独立配置。
-- options 存 JSON 数组字符串(如 ["移动","联通","电信"])，非 select 存 '[]'。
CREATE TABLE `product_order_fields` (
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
