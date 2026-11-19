package service

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	mysqlDriver "github.com/go-sql-driver/mysql"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func rechargeTestDB(t *testing.T) (*RechargeService, sqlmock.Sqlmock) {
	t.Helper()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })
	gdb, err := gorm.Open(mysql.New(mysql.Config{Conn: db, SkipInitializeWithVersion: true}), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Error(err)
		}
	})
	return NewRechargeService(gdb, NewSettingService(gdb)), mock
}

func rechargeRow(status int64) *sqlmock.Rows {
	return sqlmock.NewRows([]string{"id", "user_id", "status", "amount", "reported_amount", "address", "tx_hash"}).
		AddRow(9, 42, status, "10.12", "10.129999", "TJRabPrwbZy45sbavfcjinPJC18kjpRTv8", strings.Repeat("ab", 32))
}

func TestRechargeReviewRejectsRepeatBeforeTouchingWallet(t *testing.T) {
	for _, status := range []int64{1, 2} {
		t.Run(string(rune('0'+status)), func(t *testing.T) {
			svc, mock := rechargeTestDB(t)
			mock.ExpectBegin()
			mock.ExpectQuery("SELECT .*recharges.*FOR UPDATE").WithArgs(9, 1).WillReturnRows(rechargeRow(status))
			mock.ExpectRollback()
			_, err := svc.Review(context.Background(), 9, &ReviewRechargeRequest{Status: 1, Amount: "10.129999", Remark: "已核实付款归属"})
			if !errors.Is(err, ErrRechargeReviewed) {
				t.Fatalf("重复审核必须失败: %v", err)
			}
		})
	}
}

func expectRechargeWalletCredit(mock sqlmock.Sqlmock) {
	mock.ExpectBegin()
	mock.ExpectQuery("SELECT .*recharges.*FOR UPDATE").WithArgs(9, 1).WillReturnRows(rechargeRow(0))
	mock.ExpectExec("INSERT INTO `wallets`").WillReturnResult(sqlmock.NewResult(3, 1))
	// 即使创建钱包得到ID，也只能按当前记录的 user_id 读取并锁定钱包。
	mock.ExpectQuery("SELECT .*wallets.*FOR UPDATE").WithArgs(42, 1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "balance", "frozen_balance", "version"}).AddRow(3, 42, "5", "2", 1))
	mock.ExpectExec("UPDATE `wallets` SET").WithArgs("15.12", sqlmock.AnyArg(), 42).WillReturnResult(sqlmock.NewResult(0, 1))
}

func TestRechargeReviewCreditsTruncatedAmountAndLedgerAtomically(t *testing.T) {
	svc, mock := rechargeTestDB(t)
	expectRechargeWalletCredit(mock)
	mock.ExpectExec("INSERT INTO `wallet_transactions`").WithArgs(
		sqlmock.AnyArg(), sqlmock.AnyArg(), 42, 0, 5, "10.12", "5", "15.12", "2", "2", "RECHARGE-9", "USDT/TRC-20充值到账",
	).WillReturnResult(sqlmock.NewResult(10, 1))
	mock.ExpectExec("UPDATE `recharges` SET").WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()
	row, err := svc.Review(context.Background(), 9, &ReviewRechargeRequest{Status: 1, Amount: "10.129999", Remark: "已核实付款归属"})
	if err != nil {
		t.Fatal(err)
	}
	if row.Status != 1 || row.Amount.StringFixed(2) != "10.12" || row.ReviewedAt == nil {
		t.Fatalf("审核结果不正确: %+v", row)
	}
}

func TestRechargeReviewRollsBackWhenLedgerFails(t *testing.T) {
	svc, mock := rechargeTestDB(t)
	expectRechargeWalletCredit(mock)
	mock.ExpectExec("INSERT INTO `wallet_transactions`").WillReturnError(errors.New("账本不可写"))
	mock.ExpectRollback()
	row, err := svc.Review(context.Background(), 9, &ReviewRechargeRequest{Status: 1, Amount: "10.12", Remark: "已核实付款归属"})
	if err == nil || row != nil {
		t.Fatal("账本写入失败必须回滚且不返回成功")
	}
}

func TestRechargeRejectionNeverCreditsWallet(t *testing.T) {
	svc, mock := rechargeTestDB(t)
	mock.ExpectBegin()
	mock.ExpectQuery("SELECT .*recharges.*FOR UPDATE").WithArgs(9, 1).WillReturnRows(rechargeRow(0))
	mock.ExpectExec("UPDATE `recharges` SET").WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()
	row, err := svc.Review(context.Background(), 9, &ReviewRechargeRequest{Status: 2, Remark: "未查询到到账交易"})
	if err != nil || row.Status != 2 {
		t.Fatalf("驳回失败: %+v, %v", row, err)
	}
}

func TestRechargeListAlwaysScopesToUserIncludingStatusZero(t *testing.T) {
	svc, mock := rechargeTestDB(t)
	status := int8(0)
	mock.ExpectQuery("SELECT count.*recharges.*user_id = . AND status = .").WithArgs(42, 0).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	mock.ExpectQuery("SELECT .*recharges.*user_id = . AND status = . ORDER BY id DESC").WithArgs(42, 0, 10).
		WillReturnRows(rechargeRow(0))
	rows, total, err := svc.List(context.Background(), 42, &ListRechargeRequest{Page: 1, PageSize: 10, Status: &status})
	if err != nil || total != 1 || len(rows) != 1 || rows[0].UserID != 42 {
		t.Fatalf("本人记录查询错误: %+v, %d, %v", rows, total, err)
	}
}

func expectRechargeConfig(mock sqlmock.Sqlmock, address string) {
	mock.ExpectQuery("SELECT .*settings.*config_key =").WithArgs(SettingRechargeAddress, 1).
		WillReturnRows(sqlmock.NewRows([]string{"config_key", "config_value"}).AddRow(SettingRechargeAddress, address))
}

func TestRechargeCreateRejectsStaleAddressBeforeSaving(t *testing.T) {
	svc, mock := rechargeTestDB(t)
	expectRechargeConfig(mock, "TJRabPrwbZy45sbavfcjinPJC18kjpRTv8")
	_, err := svc.Create(context.Background(), 42, &CreateRechargeRequest{Address: "old-address", Amount: "1", TxHash: strings.Repeat("a", 64)})
	if !errors.Is(err, ErrRechargeAddressChanged) {
		t.Fatalf("旧地址不得申报: %v", err)
	}
}

func TestRechargeCreateDuplicateHashDoesNotCreateSecondRecord(t *testing.T) {
	svc, mock := rechargeTestDB(t)
	address := "TJRabPrwbZy45sbavfcjinPJC18kjpRTv8"
	expectRechargeConfig(mock, address)
	mock.ExpectQuery("SELECT .*users.*").WithArgs(42, 1).WillReturnRows(sqlmock.NewRows([]string{"id", "status"}).AddRow(42, 1))
	mock.ExpectBegin()
	mock.ExpectExec("INSERT INTO `recharges`").WillReturnError(&mysqlDriver.MySQLError{Number: 1062, Message: "duplicate"})
	mock.ExpectRollback()
	_, err := svc.Create(context.Background(), 42, &CreateRechargeRequest{Address: address, Amount: "1", TxHash: strings.Repeat("A", 64)})
	if !errors.Is(err, ErrRechargeDuplicate) {
		t.Fatalf("重复哈希须明确拒绝: %v", err)
	}
}

func TestRechargeCreateOnlySavesPendingRecordWithoutCreditingWallet(t *testing.T) {
	svc, mock := rechargeTestDB(t)
	address := "TJRabPrwbZy45sbavfcjinPJC18kjpRTv8"
	expectRechargeConfig(mock, address)
	mock.ExpectQuery("SELECT .*users.*").WithArgs(42, 1).WillReturnRows(sqlmock.NewRows([]string{"id", "status"}).AddRow(42, 1))
	mock.ExpectBegin()
	mock.ExpectExec("INSERT INTO `recharges`").WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), 42, address, strings.Repeat("ab", 32), "12.999999", "12.99", 0, "", nil).WillReturnResult(sqlmock.NewResult(9, 1))
	mock.ExpectCommit()
	row, err := svc.Create(context.Background(), 42, &CreateRechargeRequest{Address: address, Amount: " 12.999999 ", TxHash: strings.Repeat("AB", 32)})
	if err != nil || row.Status != 0 || row.Amount.StringFixed(2) != "12.99" || row.UserID != 42 {
		t.Fatalf("申报只能生成本人的待核实记录: %+v, %v", row, err)
	}
}

func TestRechargeConfigNeverExposesInvalidAddress(t *testing.T) {
	for _, address := range []string{"", "T111111111111111111111111111111111"} {
		svc, mock := rechargeTestDB(t)
		expectRechargeConfig(mock, address)
		cfg, err := svc.Config(context.Background())
		if err != nil || cfg.Enabled || cfg.Address != "" {
			t.Fatalf("未配置或错误地址不得开放充值: %+v, %v", cfg, err)
		}
	}
}

func TestRechargeReviewRequiresPayerOwnershipEvidence(t *testing.T) {
	for _, req := range []ReviewRechargeRequest{
		{Status: 1, Amount: "10", Remark: ""},
		{Status: 1, Amount: "10", Remark: "   "},
	} {
		svc, _ := rechargeTestDB(t)
		_, err := svc.Review(context.Background(), 9, &req)
		if !errors.Is(err, ErrInvalidRecharge) {
			t.Fatalf("未记录付款归属核验依据不得入账: %v", err)
		}
	}
}

func TestRechargeReviewFinalSaveFailureRollsBackCreditAndLedger(t *testing.T) {
	svc, mock := rechargeTestDB(t)
	expectRechargeWalletCredit(mock)
	mock.ExpectExec("INSERT INTO `wallet_transactions`").WillReturnResult(sqlmock.NewResult(10, 1))
	mock.ExpectExec("UPDATE `recharges` SET").WillReturnError(errors.New("审核状态不可写"))
	mock.ExpectRollback()
	row, err := svc.Review(context.Background(), 9, &ReviewRechargeRequest{Status: 1, Amount: "10.12", Remark: "已核实付款归属"})
	if err == nil || row != nil {
		t.Fatal("最终状态失败必须回滚钱包与流水")
	}
}

func TestRechargeReviewUsesVerifiedAmountRatherThanUserReport(t *testing.T) {
	svc, mock := rechargeTestDB(t)
	mock.ExpectBegin()
	mock.ExpectQuery("SELECT .*recharges.*FOR UPDATE").WithArgs(9, 1).WillReturnRows(rechargeRow(0))
	mock.ExpectExec("INSERT INTO `wallets`").WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectQuery("SELECT .*wallets.*FOR UPDATE").WithArgs(42, 1).WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "balance", "frozen_balance", "version"}).AddRow(3, 42, "5", "2", 1))
	mock.ExpectExec("UPDATE `wallets` SET").WithArgs("14.87", sqlmock.AnyArg(), 42).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec("INSERT INTO `wallet_transactions`").WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), 42, 0, 5, "9.87", "5", "14.87", "2", "2", "RECHARGE-9", "USDT/TRC-20充值到账").WillReturnResult(sqlmock.NewResult(10, 1))
	mock.ExpectExec("UPDATE `recharges` SET").WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()
	row, err := svc.Review(context.Background(), 9, &ReviewRechargeRequest{Status: 1, Amount: "9.879999", Remark: "已核实付款归属"})
	if err != nil || row.Amount.StringFixed(2) != "9.87" || row.ReportedAmount != "10.129999" {
		t.Fatalf("应保留用户申报且按实际核实金额入账: %+v, %v", row, err)
	}
}
