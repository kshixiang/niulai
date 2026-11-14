-- 订单表新增「已到账金额」列：订单中心需展示 到账金额 / 剩余金额（剩余=订单金额-已到账）。
-- 订单状态语义在 001 之后调整为 6 态（tinyint 列类型不变，无需改列）。
--
-- 说明：挂载到 /docker-entrypoint-initdb.d 的脚本仅在 mysql 数据卷【首次初始化】时自动执行，
-- 已存在的数据卷不会重跑，因此本文件需在部署脚本里手动补齐（见仓库根 deploy.sh）。
-- 为保证可重复执行（已存在该列时不报错），用信息_schema 判断后动态 ALTER。
SET @db = DATABASE();
SET @tbl = 'orders';
SET @col = 'received_amount';
SET @sql = (
  SELECT IF(
    (SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS
       WHERE TABLE_SCHEMA = @db AND TABLE_NAME = @tbl AND COLUMN_NAME = @col) = 0,
    'ALTER TABLE `orders` ADD COLUMN `received_amount` DECIMAL(10,2) NOT NULL DEFAULT 0.00 COMMENT ''已到账金额'' AFTER `amount`',
    'SELECT 1'
  )
);
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;
