-- 008 订单表新增「下单填写项快照」列：保存 H5 下单页按后台配置填写的全部字段。
-- 与 internal/model/order.go 中的 FormData 字段保持一致。
-- 幂等：列已存在则跳过，避免已有数据卷重复执行报错。
SET NAMES utf8mb4;
USE niulai;

SET @db = DATABASE();
SET @tbl = 'orders';
SET @col = 'form_data';
SET @sql = (SELECT IF(
  (SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS
     WHERE TABLE_SCHEMA = @db AND TABLE_NAME = @tbl AND COLUMN_NAME = @col) = 0,
  'ALTER TABLE `orders` ADD COLUMN `form_data` TEXT NULL COMMENT ''下单填写项快照(JSON数组)'' AFTER `account`',
  'SELECT 1'
));
PREPARE s FROM @sql;
EXECUTE s;
DEALLOCATE PREPARE s;
