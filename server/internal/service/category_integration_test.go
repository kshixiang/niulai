package service

import (
	"context"
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gin-gonic/gin/binding"
)

func TestDynamicCategoryPassesRequestValidation(t *testing.T) {
	typeCode := "cat_membership"
	for _, req := range []any{
		&CreateChannelRequest{Name: "上游", Category: typeCode, Status: 1},
		&UpdateChannelRequest{Name: "上游", Category: typeCode, Status: 1},
		&CreateProductRequest{Name: "会员商品", Type: typeCode, CoverImage: "/cover.png", ChannelDiscount: "8", Discount: "9", Denominations: []DenominationInput{{FaceValue: "100"}}, ChannelID: 1},
		&UpdateProductRequest{Type: &typeCode},
	} {
		if err := binding.Validator.ValidateStruct(req); err != nil {
			t.Errorf("动态分类不应被固定枚举拦截(%T): %v", req, err)
		}
	}
}

func TestChannelRejectsMissingCategoryBeforeWriting(t *testing.T) {
	db, mock := categoryTestDB(t)
	mock.ExpectQuery("SELECT .*product_categories").WithArgs("missing", 1).WillReturnRows(sqlmock.NewRows([]string{"id", "code", "name"}))
	_, err := NewChannelService(db).Create(context.Background(), &CreateChannelRequest{Name: "渠道", Category: "missing", Status: 1})
	if !errors.Is(err, ErrInvalidCategory) {
		t.Fatalf("不存在的分类不能写入渠道: %v", err)
	}
}

func TestProductRejectsMissingCategoryBeforeWriting(t *testing.T) {
	db, mock := categoryTestDB(t)
	mock.ExpectQuery("SELECT .*product_categories").WithArgs("missing", 1).WillReturnRows(sqlmock.NewRows([]string{"id", "code", "name"}))
	_, err := NewProductService(db).Create(context.Background(), &CreateProductRequest{Type: "missing", ChannelDiscount: "8", Discount: "9"})
	if !errors.Is(err, ErrInvalidCategory) {
		t.Fatalf("不存在的分类不能写入商品: %v", err)
	}
}

func TestProductUpdateRejectsExplicitEmptyCategory(t *testing.T) {
	db, mock := categoryTestDB(t)
	mock.ExpectQuery("SELECT .*products").WithArgs(9, 1).WillReturnRows(sqlmock.NewRows([]string{"id", "type", "status"}).AddRow(9, "话费", 0))
	empty := ""
	_, err := NewProductService(db).Update(context.Background(), 9, &UpdateProductRequest{Type: &empty})
	if !errors.Is(err, ErrInvalidCategory) {
		t.Fatalf("显式清空分类应失败: %v", err)
	}
}

func TestChannelListResolvesRenamedCategory(t *testing.T) {
	db, mock := categoryTestDB(t)
	mock.ExpectQuery("SELECT count\\(\\*\\) FROM `channels`").WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	mock.ExpectQuery("SELECT .*channels.*ORDER BY id DESC").WillReturnRows(sqlmock.NewRows([]string{"id", "name", "category", "status"}).AddRow(3, "上游", "话费", 1))
	mock.ExpectQuery("SELECT .*product_categories.*code IN").WithArgs("话费").WillReturnRows(sqlmock.NewRows([]string{"code", "name"}).AddRow("话费", "手机充值"))
	list, total, err := NewChannelService(db).List(context.Background(), 1, 10)
	if err != nil {
		t.Fatal(err)
	}
	if total != 1 || len(list) != 1 || list[0].Category != "话费" || list[0].CategoryName != "手机充值" {
		t.Fatalf("应保留标识并显示新名称: %+v", list)
	}
}

func TestClientProductListResolvesNamesWithoutExposingChannel(t *testing.T) {
	db, mock := categoryTestDB(t)
	mock.ExpectQuery("SELECT count\\(\\*\\) FROM `products`").WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	mock.ExpectQuery("SELECT id, name, type, cover_image, discount, description FROM `products`").WillReturnRows(sqlmock.NewRows([]string{"id", "name", "type", "discount"}).AddRow(1, "快充", "话费", "9"))
	mock.ExpectQuery("SELECT .*product_categories.*code IN").WithArgs("话费").WillReturnRows(sqlmock.NewRows([]string{"code", "name"}).AddRow("话费", "手机充值"))
	list, _, err := NewClientService(db).ListProducts(context.Background(), &ListClientProductRequest{Page: 1, PageSize: 10})
	if err != nil {
		t.Fatal(err)
	}
	if len(list) != 1 || list[0].Type != "话费" || list[0].TypeName != "手机充值" {
		t.Fatalf("客户端应返回新显示名与固定标识: %+v", list)
	}
}
