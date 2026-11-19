-- 分类名称可编辑，code 固定关联商品和渠道，兼容原有话费业务规则。
SET NAMES utf8mb4;
USE niulai;

CREATE TABLE IF NOT EXISTS `product_categories` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `code` VARCHAR(32) NOT NULL COMMENT '固定关联标识，改名时不变',
  `name` VARCHAR(32) NOT NULL COMMENT '分类显示名称',
  `created_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  `updated_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_product_category_code` (`code`),
  UNIQUE KEY `uk_product_category_name` (`name`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='商品分类';

-- 重复执行只补缺失分类，绝不覆盖已编辑的显示名。
INSERT IGNORE INTO `product_categories` (`code`, `name`) VALUES ('话费', '话费'), ('卡券', '卡券');
INSERT IGNORE INTO `product_categories` (`code`, `name`)
  SELECT DISTINCT `type`, `type` FROM `products` WHERE TRIM(`type`) <> '';
INSERT IGNORE INTO `product_categories` (`code`, `name`)
  SELECT DISTINCT `category`, `category` FROM `channels` WHERE TRIM(`category`) <> '';
