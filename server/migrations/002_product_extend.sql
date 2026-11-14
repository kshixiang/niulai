-- 002 商品表扩展：新增折扣标签、描述、上游渠道绑定，并调整默认状态为下架。
-- 仅对已有数据卷手动执行；首次初始化时 docker 会自动按文件名顺序执行本文件。
-- 与 internal/model/product.go 中的字段定义保持一致。
SET NAMES utf8mb4;
USE niulai;

ALTER TABLE `products`
  ADD COLUMN `discount_label` VARCHAR(32)  NOT NULL DEFAULT '' COMMENT '折扣标签 如7.8折' AFTER `cost_price`,
  ADD COLUMN `description`    VARCHAR(512) NOT NULL DEFAULT '' COMMENT '提示语/描述'        AFTER `discount_label`,
  ADD COLUMN `channel_id`     BIGINT       NOT NULL DEFAULT 0  COMMENT '绑定的上游渠道ID'   AFTER `description`,
  MODIFY COLUMN `status`      TINYINT      NOT NULL DEFAULT 0  COMMENT '上下架状态：1上架 0下架';
