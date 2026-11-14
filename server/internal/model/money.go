package model

import "github.com/shopspring/decimal"

// Money 货币金额类型别名。
//
// 为什么不使用 float64？
//   - 浮点数采用 IEEE-754 二进制表示，无法精确表达多数十进制小数
//     （如 0.1 + 0.2 != 0.3），在金额累加、对账、佣金结算场景下
//     会产生肉眼不可见但致命的精度误差。
//
// 为什么使用 decimal.Decimal 而不是 int64(分)？
//   - int64 存“分”可规避浮点误差，但需在全链路（展示/计算/入库）手动做
//     单位换算，极易因漏换算导致金额被放大/缩小 100 倍。
//   - decimal.Decimal 原生支持四则运算与四舍五入，且已实现
//     database/sql 的 Scanner/Valuer 接口，配合 GORM 可直接映射为
//     MySQL 的 DECIMAL 列，无需手写类型转换。
//   - 分佣比例、费率等本身就不是整数“分”概念，用 decimal 统一表达更自然。
//
// 注意：DECIMAL 列必须显式指定精度（如 decimal(18,2)），
// 否则 GORM 不会推断为定长精度列，存在默认精度不足的风险。
type Money = decimal.Decimal

// NewMoney 便捷构造：以字符串创建金额，避免字面量精度丢失。
func NewMoney(s string) Money {
	return decimal.RequireFromString(s)
}
