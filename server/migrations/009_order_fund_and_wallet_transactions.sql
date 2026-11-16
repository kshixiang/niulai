-- 009 下单资金闭环：orders 增加资金快照列，并新建 wallet_transactions 资金流水表。
-- 与 internal/model/order.go、internal/model/wallet_transaction.go 保持一致。
-- 幂等：列/表已存在则跳过，避免已有数据卷重复执行报错。
SET NAMES utf8mb4;
USE niulai;

-- orders.frozen_usdt：下单冻结/实扣的 USDT 快照，资金动作的唯一金额依据
SET @db = DATABASE();
SET @tbl = 'orders';
SET @col = 'frozen_usdt';
SET @sql = (SELECT IF(
  (SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS
     WHERE TABLE_SCHEMA = @db AND TABLE_NAME = @tbl AND COLUMN_NAME = @col) = 0,
  'ALTER TABLE `orders` ADD COLUMN `frozen_usdt` DECIMAL(10,2) NOT NULL DEFAULT 0.00 COMMENT ''冻结/实扣USDT快照'' AFTER `profit`',
  'SELECT 1'
));
PREPARE s FROM @sql;
EXECUTE s;
DEALLOCATE PREPARE s;

-- orders.usdt_cny_rate：下单时的汇率快照，仅对账用
SET @col = 'usdt_cny_rate';
SET @sql = (SELECT IF(
  (SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS
     WHERE TABLE_SCHEMA = @db AND TABLE_NAME = @tbl AND COLUMN_NAME = @col) = 0,
  'ALTER TABLE `orders` ADD COLUMN `usdt_cny_rate` DECIMAL(10,2) NOT NULL DEFAULT 0.00 COMMENT ''下单汇率快照(1USDT=N CNY)'' AFTER `frozen_usdt`',
  'SELECT 1'
));
PREPARE s FROM @sql;
EXECUTE s;
DEALLOCATE PREPARE s;

-- orders.fund_status：资金状态幂等键 0冻结中 1已扣款 2已退还
SET @col = 'fund_status';
SET @sql = (SELECT IF(
  (SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS
     WHERE TABLE_SCHEMA = @db AND TABLE_NAME = @tbl AND COLUMN_NAME = @col) = 0,
  'ALTER TABLE `orders` ADD COLUMN `fund_status` TINYINT NOT NULL DEFAULT 0 COMMENT ''0冻结中 1已扣款 2已退还'' AFTER `usdt_cny_rate`',
  'SELECT 1'
));
PREPARE s FROM @sql;
EXECUTE s;
DEALLOCATE PREPARE s;

-- wallet_transactions：资金流水（账本类，无软删除）
CREATE TABLE IF NOT EXISTS `wallet_transactions` (
  `id`             BIGINT       NOT NULL AUTO_INCREMENT COMMENT '主键',
  `user_id`        BIGINT       NOT NULL COMMENT '用户ID(关联users.id)',
  `order_id`       BIGINT       NOT NULL DEFAULT 0 COMMENT '订单ID(关联orders.id)',
  `type`           TINYINT      NOT NULL COMMENT '类型 1下单冻结 2成功扣款 3失败解冻 4退款归还',
  `amount`         DECIMAL(10,2) NOT NULL DEFAULT 0.00 COMMENT '变动USDT金额(恒为正,方向由type表达)',
  `balance_before` DECIMAL(10,2) NOT NULL DEFAULT 0.00 COMMENT '变动前可用余额',
  `balance_after`  DECIMAL(10,2) NOT NULL DEFAULT 0.00 COMMENT '变动后可用余额',
  `frozen_before`  DECIMAL(10,2) NOT NULL DEFAULT 0.00 COMMENT '变动前冻结余额',
  `frozen_after`   DECIMAL(10,2) NOT NULL DEFAULT 0.00 COMMENT '变动后冻结余额',
  `ref`            VARCHAR(64)  NOT NULL DEFAULT '' COMMENT '业务号(订单号)',
  `remark`         VARCHAR(255) NOT NULL DEFAULT '' COMMENT '备注',
  `created_at`     DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at`     DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  KEY `idx_user_created` (`user_id`, `created_at`),
  KEY `idx_order_id` (`order_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='资金流水表';
