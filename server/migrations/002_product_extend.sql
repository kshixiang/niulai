-- 002 商品表扩展：新增折扣标签、描述、上游渠道绑定，并调整默认状态为下架。
-- 与 internal/model/product.go 中的字段定义保持一致。
-- 改为幂等：列已存在则跳过，避免已有数据卷重复执行报错。
SET NAMES utf8mb4;
USE niulai;

SET @db = DATABASE();
SET @tbl = 'products';

SET @col = 'discount_label';
SET @sql = (SELECT IF(
  (SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_SCHEMA=@db AND TABLE_NAME=@tbl AND COLUMN_NAME=@col)=0,
  'ALTER TABLE `products` ADD COLUMN `discount_label` VARCHAR(32) NOT NULL DEFAULT '''' COMMENT ''折扣标签 如7.8折'' AFTER `cost_price`',
  'SELECT 1'));
PREPARE s FROM @sql; EXECUTE s; DEALLOCATE PREPARE s;

SET @col = 'description';
SET @sql = (SELECT IF(
  (SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_SCHEMA=@db AND TABLE_NAME=@tbl AND COLUMN_NAME=@col)=0,
  'ALTER TABLE `products` ADD COLUMN `description` VARCHAR(512) NOT NULL DEFAULT '''' COMMENT ''提示语/描述'' AFTER `discount_label`',
  'SELECT 1'));
PREPARE s FROM @sql; EXECUTE s; DEALLOCATE PREPARE s;

SET @col = 'channel_id';
SET @sql = (SELECT IF(
  (SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_SCHEMA=@db AND TABLE_NAME=@tbl AND COLUMN_NAME=@col)=0,
  'ALTER TABLE `products` ADD COLUMN `channel_id` BIGINT NOT NULL DEFAULT 0 COMMENT ''绑定的上游渠道ID'' AFTER `description`',
  'SELECT 1'));
PREPARE s FROM @sql; EXECUTE s; DEALLOCATE PREPARE s;

-- status 列重复 MODIFY 幂等，不报错
ALTER TABLE `products` MODIFY COLUMN `status` TINYINT NOT NULL DEFAULT 0 COMMENT '上下架状态：1上架 0下架';
