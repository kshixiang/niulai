-- 充值申报与人工核实入账。已有环境需在发布新版本前执行此迁移。
SET NAMES utf8mb4;
USE niulai;

CREATE TABLE IF NOT EXISTS `recharges` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `created_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  `updated_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  `user_id` BIGINT UNSIGNED NOT NULL,
  `address` VARCHAR(34) NOT NULL COMMENT '申报时收款地址快照',
  `tx_hash` CHAR(64) NOT NULL COMMENT '小写TRON交易哈希，全局唯一',
  `reported_amount` VARCHAR(32) NOT NULL COMMENT '用户申报原始金额，最多6位小数',
  `amount` DECIMAL(10,2) NOT NULL COMMENT '截断两位的待核实/实际入账USDT金额',
  `status` TINYINT NOT NULL DEFAULT 0 COMMENT '0待核实 1已到账 2已驳回',
  `remark` VARCHAR(255) NOT NULL DEFAULT '',
  `reviewed_at` DATETIME(3) DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_recharge_tx_hash` (`tx_hash`),
  KEY `idx_recharge_user_created` (`user_id`),
  KEY `idx_recharge_status` (`status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='USDT充值记录';

INSERT IGNORE INTO `settings` (`config_key`, `config_value`) VALUES ('recharge_trc20_address', '');
ALTER TABLE `wallet_transactions` MODIFY COLUMN `type` TINYINT NOT NULL COMMENT '1下单冻结 2成功扣款 3失败解冻 4退款归还 5充值到账';
