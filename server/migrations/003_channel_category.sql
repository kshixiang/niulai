-- 003 上游渠道扩展：新增商品分类。
-- 与 internal/model/channel.go 中的字段定义保持一致。
-- 改为幂等：列已存在则跳过，避免已有数据卷重复执行报错。
SET NAMES utf8mb4;
USE niulai;

SET @db = DATABASE();
SET @tbl = 'channels';
SET @col = 'category';
SET @sql = (SELECT IF(
  (SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_SCHEMA=@db AND TABLE_NAME=@tbl AND COLUMN_NAME=@col)=0,
  'ALTER TABLE `channels` ADD COLUMN `category` VARCHAR(32) NOT NULL DEFAULT '''' COMMENT ''商品分类 话费/卡券'' AFTER `name`',
  'SELECT 1'));
PREPARE s FROM @sql; EXECUTE s; DEALLOCATE PREPARE s;
