-- 003 上游渠道扩展：新增商品分类。
-- 仅对已有数据卷手动执行；首次初始化时 docker 会自动按文件名顺序执行本文件。
-- 与 internal/model/channel.go 中的字段定义保持一致。
SET NAMES utf8mb4;
USE niulai;

ALTER TABLE `channels`
  ADD COLUMN `category` VARCHAR(32) NOT NULL DEFAULT '' COMMENT '商品分类 话费/卡券' AFTER `name`;
