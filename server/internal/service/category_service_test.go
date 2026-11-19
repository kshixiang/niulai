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

func categoryTestDB(t *testing.T) (*gorm.DB, sqlmock.Sqlmock) {
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
	return gdb, mock
}

func TestCategoryNamesAreNormalizedAndValidated(t *testing.T) {
	for _, name := range []string{"", " \t\n", strings.Repeat("类", 33)} {
		t.Run(name, func(t *testing.T) {
			db, _ := categoryTestDB(t)
			if _, err := NewCategoryService(db).Create(context.Background(), name); !errors.Is(err, ErrInvalidCategoryName) {
				t.Fatalf("非法分类名应在写库前拒绝: %v", err)
			}
		})
	}
	db, mock := categoryTestDB(t)
	mock.ExpectBegin()
	mock.ExpectExec("INSERT INTO `product_categories`").WithArgs(sqlmock.AnyArg(), "会员", sqlmock.AnyArg(), sqlmock.AnyArg()).WillReturnResult(sqlmock.NewResult(3, 1))
	mock.ExpectCommit()
	got, err := NewCategoryService(db).Create(context.Background(), " 会员 ")
	if err != nil {
		t.Fatal(err)
	}
	if got.Name != "会员" || got.Code == "" || got.Code == "会员" || len(got.Code) > 32 {
		t.Fatalf("新增分类应有独立固定标识: %+v", got)
	}
}

func TestCategoryCreateMapsUniqueConflict(t *testing.T) {
	db, mock := categoryTestDB(t)
	mock.ExpectBegin()
	mock.ExpectExec("INSERT INTO `product_categories`").WillReturnError(&mysqlDriver.MySQLError{Number: 1062, Message: "Duplicate entry"})
	mock.ExpectRollback()
	if _, err := NewCategoryService(db).Create(context.Background(), "会员"); !errors.Is(err, ErrCategoryNameExists) {
		t.Fatalf("重名必须返回可识别业务错误: %v", err)
	}
}

func TestCategoryAccepts32UnicodeCharacters(t *testing.T) {
	db, mock := categoryTestDB(t)
	name := strings.Repeat("类", 32)
	mock.ExpectBegin()
	mock.ExpectExec("INSERT INTO `product_categories`").WithArgs(sqlmock.AnyArg(), name, sqlmock.AnyArg(), sqlmock.AnyArg()).WillReturnResult(sqlmock.NewResult(4, 1))
	mock.ExpectCommit()
	got, err := NewCategoryService(db).Create(context.Background(), name)
	if err != nil || got.Name != name {
		t.Fatalf("32个中文字符应被接受: %+v %v", got, err)
	}
}

func TestCategoryRenameRejectsDuplicatesAndMissingIDs(t *testing.T) {
	t.Run("duplicate", func(t *testing.T) {
		db, mock := categoryTestDB(t)
		mock.ExpectQuery("SELECT .*product_categories").WithArgs(1, 1).WillReturnRows(sqlmock.NewRows([]string{"id", "code", "name"}).AddRow(1, "话费", "话费"))
		mock.ExpectBegin()
		mock.ExpectExec("UPDATE `product_categories`").WillReturnError(&mysqlDriver.MySQLError{Number: 1062})
		mock.ExpectRollback()
		if _, err := NewCategoryService(db).Update(context.Background(), 1, "卡券"); !errors.Is(err, ErrCategoryNameExists) {
			t.Fatalf("分类重名应拒绝: %v", err)
		}
	})
	t.Run("missing", func(t *testing.T) {
		db, mock := categoryTestDB(t)
		mock.ExpectQuery("SELECT .*product_categories").WithArgs(999, 1).WillReturnRows(sqlmock.NewRows([]string{"id", "code", "name"}))
		if _, err := NewCategoryService(db).Update(context.Background(), 999, "会员"); !errors.Is(err, ErrCategoryNotFound) {
			t.Fatalf("不存在的分类应返回404业务错误: %v", err)
		}
	})
}

func TestCategoryRenamePreservesPhoneIdentity(t *testing.T) {
	db, mock := categoryTestDB(t)
	mock.ExpectQuery("SELECT .*product_categories").WithArgs(1, 1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "code", "name"}).AddRow(1, "话费", "话费"))
	mock.ExpectBegin()
	mock.ExpectExec("UPDATE `product_categories` SET `name`=\\?,`updated_at`=\\? WHERE `id` = \\?").
		WithArgs("手机充值", sqlmock.AnyArg(), 1).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()
	got, err := NewCategoryService(db).Update(context.Background(), 1, "手机充值")
	if err != nil {
		t.Fatal(err)
	}
	if got.Code != "话费" || got.Name != "手机充值" {
		t.Fatalf("改名不得改写固定标识: %+v", got)
	}
	if len(appendPhoneFields(nil, got.Code)) != 3 {
		t.Fatal("话费改名后必须保留默认下单字段")
	}
}

func TestCategoryLookupUsesLatestNamesAndLegacyFallback(t *testing.T) {
	db, mock := categoryTestDB(t)
	mock.ExpectQuery("SELECT .*product_categories.*code IN").WithArgs("话费", "卡券", "legacy").
		WillReturnRows(sqlmock.NewRows([]string{"code", "name"}).AddRow("话费", "手机充值").AddRow("卡券", "礼品卡"))
	names, err := categoryNames(context.Background(), db, []string{"话费", "卡券", "话费", "legacy", ""})
	if err != nil {
		t.Fatal(err)
	}
	if names["话费"] != "手机充值" || names["卡券"] != "礼品卡" || names["legacy"] != "legacy" {
		t.Fatalf("应批量解析当前名称并兼容旧数据: %v", names)
	}
}

func TestCategoryReferenceMustExist(t *testing.T) {
	db, mock := categoryTestDB(t)
	if err := ensureCategoryExists(context.Background(), db, ""); !errors.Is(err, ErrInvalidCategory) {
		t.Fatalf("空分类应拒绝: %v", err)
	}
	mock.ExpectQuery("SELECT .*product_categories").WithArgs("missing", 1).WillReturnRows(sqlmock.NewRows([]string{"id", "code", "name"}))
	if err := ensureCategoryExists(context.Background(), db, "missing"); !errors.Is(err, ErrInvalidCategory) {
		t.Fatalf("不存在的分类应拒绝: %v", err)
	}
}

func TestCategoryReferenceRejectsNonCanonicalCodes(t *testing.T) {
	for _, tc := range []struct{ input, stored string }{
		{"话费 ", "话费"},
		{"CAT_abcdef", "cat_abcdef"},
	} {
		t.Run(tc.input, func(t *testing.T) {
			db, mock := categoryTestDB(t)
			// MySQL utf8mb4_unicode_ci 会匹配大小写/尾随空格；业务标识必须精确一致。
			mock.ExpectQuery("SELECT .*product_categories").WithArgs(tc.input, 1).WillReturnRows(sqlmock.NewRows([]string{"id", "code", "name"}).AddRow(1, tc.stored, "分类"))
			if err := ensureCategoryExists(context.Background(), db, tc.input); !errors.Is(err, ErrInvalidCategory) {
				t.Fatalf("非规范标识不能绕过话费下单规则: %v", err)
			}
		})
	}
}

func TestCategoryListReturnsAllOptions(t *testing.T) {
	db, mock := categoryTestDB(t)
	rows := sqlmock.NewRows([]string{"id", "code", "name"})
	for i := 1; i <= 105; i++ {
		rows.AddRow(i, strings.Repeat("x", i), "分类")
	}
	mock.ExpectQuery("SELECT .*product_categories.*ORDER BY id ASC$").WillReturnRows(rows)
	list, err := NewCategoryService(db).List(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(list) != 105 {
		t.Fatalf("分类选项不应被100条分页截断: %d", len(list))
	}
}
