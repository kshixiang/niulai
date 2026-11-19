package service

import (
	"strings"
	"testing"
)

func TestRechargeAmountTruncatesWithoutRounding(t *testing.T) {
	for _, tc := range []struct{ input, want string }{
		{"1", "1.00"}, {"1.009999", "1.00"}, {"12.999999", "12.99"},
		{" 100.129 ", "100.12"}, {"99999999.99", "99999999.99"},
	} {
		got, err := parseRechargeAmount(tc.input)
		if err != nil || got.StringFixed(2) != tc.want {
			t.Errorf("%q: got %s, %v; want %s", tc.input, got, err, tc.want)
		}
	}
	for _, raw := range []string{"", "0", "0.999999", "-1", "1e2", "1.1234567", "100000000", "NaN", "+1", "1.2.3"} {
		if _, err := parseRechargeAmount(raw); err == nil {
			t.Errorf("应拒绝无效充值金额 %q", raw)
		}
	}
}

func TestRechargeAddressChecksBase58Checksum(t *testing.T) {
	valid := "TJRabPrwbZy45sbavfcjinPJC18kjpRTv8"
	if err := validateRechargeAddress(valid); err != nil {
		t.Fatalf("有效 TRON 地址被拒绝: %v", err)
	}
	for _, raw := range []string{"", "0x123", strings.Repeat("T", 34), valid[:33] + "9", "T0RabPrwbZy45sbavfcjinPJC18kjpRTv8"} {
		if err := validateRechargeAddress(raw); err == nil {
			t.Errorf("应拒绝无效地址 %q", raw)
		}
	}
}

func TestRechargeHashNormalizesAndRejectsMalformed(t *testing.T) {
	want := strings.Repeat("ab", 32)
	got, err := normalizeRechargeHash("  " + strings.ToUpper(want) + "  ")
	if err != nil || got != want {
		t.Fatalf("哈希未归一化: %q, %v", got, err)
	}
	for _, raw := range []string{"", strings.Repeat("g", 64), strings.Repeat("a", 63), "0x" + want} {
		if _, err := normalizeRechargeHash(raw); err == nil {
			t.Errorf("应拒绝无效哈希 %q", raw)
		}
	}
}
