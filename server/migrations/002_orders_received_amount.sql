-- 订单表新增「已到账金额」列：订单中心需展示 到账金额 / 剩余金额（剩余=订单金额-已到账）。
-- 订单状态语义在 001 之后调整为 待接单/已接单/已完成/已取消（tinyint 列类型不变，无需改列）。
--
-- 已有 mysql 数据卷不会重跑 001，需手动执行本文件（或在 release 迁移流程中执行）：
--   mysql -uroot -p niulai < server/migrations/002_orders_received_amount.sql
-- 开发环境（非 release）由 GORM 自动迁移补齐该列，无需手动执行。
ALTER TABLE `orders`
  ADD COLUMN `received_amount` DECIMAL(10,2) NOT NULL DEFAULT 0.00 COMMENT '已到账金额' AFTER `amount`;
