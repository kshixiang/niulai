package service

import (
	"testing"

	"niulai/internal/model"
)

func TestAppendPhoneFields(t *testing.T) {
	configured := func(keys ...string) []model.ProductOrderField {
		out := make([]model.ProductOrderField, 0, len(keys))
		for _, k := range keys {
			out = append(out, model.ProductOrderField{FieldKey: k, Label: "自定义"})
		}
		return out
	}
	keysOf := func(rows []model.ProductOrderField) []string {
		out := make([]string, 0, len(rows))
		for _, r := range rows {
			out = append(out, r.FieldKey)
		}
		return out
	}

	t.Run("非话费商品不补字段", func(t *testing.T) {
		got := appendPhoneFields(configured("remark"), model.ProductTypeCard)
		if len(got) != 1 || got[0].FieldKey != "remark" {
			t.Fatalf("卡券商品不应补字段, got=%v", keysOf(got))
		}
	})

	t.Run("话费未配置时补齐三项且必填", func(t *testing.T) {
		got := appendPhoneFields(nil, model.ProductTypePhone)
		if len(got) != 3 {
			t.Fatalf("期望 3 项, got=%v", keysOf(got))
		}
		want := []string{"phone", "username", "carrier"}
		for i, k := range want {
			if got[i].FieldKey != k {
				t.Fatalf("顺序应为 phone, username, carrier, got=%v", keysOf(got))
			}
		}
		for _, f := range got {
			if f.Required != 1 {
				t.Errorf("%s 应为必填", f.FieldKey)
			}
		}
		// 运营商必须是下拉，且四个选项完整可被前端解析
		carrier := got[2]
		if carrier.Type != OrderFieldTypeSelect {
			t.Fatalf("运营商应为 select 类型, got=%s", carrier.Type)
		}
		opts := parseOptions(carrier.Options)
		wantOpts := []string{"移动", "电信", "联通", "广电"}
		if len(opts) != len(wantOpts) {
			t.Fatalf("运营商选项数量不符, got=%v", opts)
		}
		for i, o := range wantOpts {
			if opts[i] != o {
				t.Fatalf("运营商选项顺序不符, got=%v", opts)
			}
		}
	})

	t.Run("后台已配置的以配置为准", func(t *testing.T) {
		got := appendPhoneFields(configured("phone", "remark"), model.ProductTypePhone)
		if len(got) != 4 {
			t.Fatalf("期望 4 项, got=%v", keysOf(got))
		}
		if got[0].FieldKey != "username" || got[1].FieldKey != "carrier" {
			t.Fatalf("缺的 username/carrier 应排最前, got=%v", keysOf(got))
		}
		for _, f := range got {
			if f.FieldKey == "phone" && f.Label != "自定义" {
				t.Errorf("后台配置的 phone 不应被覆盖, label=%s", f.Label)
			}
		}
	})

	t.Run("三项都已配置则不重复补", func(t *testing.T) {
		got := appendPhoneFields(configured("phone", "username", "carrier"), model.ProductTypePhone)
		if len(got) != 3 {
			t.Fatalf("不应重复补字段, got=%v", keysOf(got))
		}
	})
}
