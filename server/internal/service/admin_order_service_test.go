package service

import (
	"errors"
	"testing"

	"github.com/shopspring/decimal"
)

func TestParseOrderAmount(t *testing.T) {
	cases := []struct {
		in   string
		want string
		ok   bool
	}{
		{"0", "0", true},
		{"60", "60", true},
		{"60.00", "60.00", true},
		{"100.5", "100.5", true},
		{"10000000", "10000000", true}, // 上限内
		{"", "", false},
		{"-1", "", false},
		{"1.234", "", false}, // 超过 2 位小数
		{"abc", "", false},
		{"10000001", "", false}, // 超过上限
	}
	for _, c := range cases {
		got, err := parseOrderAmount(c.in)
		if c.ok {
			if err != nil {
				t.Errorf("%q 期望成功，实际失败: %v", c.in, err)
				continue
			}
			if !got.Equal(decimal.RequireFromString(c.want)) {
				t.Errorf("%q 期望 %s，实际 %s", c.in, c.want, got.String())
			}
			continue
		}
		if err == nil {
			t.Errorf("%q 期望失败，实际返回 %s", c.in, got.String())
		} else if !errors.Is(err, ErrInvalidOrderParam) {
			t.Errorf("%q 错误类型不符: %v", c.in, err)
		}
	}
}
