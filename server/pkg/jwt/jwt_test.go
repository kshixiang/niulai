package jwt

import (
	"encoding/base64"
	"errors"
	"strings"
	"testing"
	"time"
)

const testSecret = "test-secret-at-least-32-bytes-long"

func TestSignAndVerify(t *testing.T) {
	token, err := Sign("admin", testSecret, 24)
	if err != nil {
		t.Fatalf("签发失败: %v", err)
	}
	claims, err := Verify(token, testSecret)
	if err != nil {
		t.Fatalf("校验失败: %v", err)
	}
	if claims.Sub != "admin" {
		t.Fatalf("令牌主体不符: %s", claims.Sub)
	}
	if claims.Exp <= time.Now().Unix() {
		t.Fatal("过期时间无效")
	}
}

func TestVerifyRejectsWrongSecret(t *testing.T) {
	token, _ := Sign("admin", testSecret, 24)
	if _, err := Verify(token, "another-secret"); !errors.Is(err, ErrInvalidToken) {
		t.Fatalf("应拒绝错误密钥, got %v", err)
	}
}

func TestVerifyRejectsTamperedPayload(t *testing.T) {
	token, _ := Sign("admin", testSecret, 24)
	parts := strings.Split(token, ".")
	raw, _ := base64.RawURLEncoding.DecodeString(parts[1])
	tampered := base64.RawURLEncoding.EncodeToString(
		[]byte(strings.Replace(string(raw), "admin", "root", 1)),
	)
	if _, err := Verify(parts[0]+"."+tampered+"."+parts[2], testSecret); !errors.Is(err, ErrInvalidToken) {
		t.Fatalf("应拒绝篡改载荷, got %v", err)
	}
}

func TestVerifyRejectsAlgNone(t *testing.T) {
	h := base64.RawURLEncoding.EncodeToString([]byte(`{"alg":"none","typ":"JWT"}`))
	p := base64.RawURLEncoding.EncodeToString([]byte(`{"sub":"admin","iat":1,"exp":9999999999}`))
	if _, err := Verify(h+"."+p+".", testSecret); !errors.Is(err, ErrInvalidToken) {
		t.Fatalf("应拒绝 alg=none, got %v", err)
	}
}

func TestVerifyRejectsExpired(t *testing.T) {
	p := base64.RawURLEncoding.EncodeToString([]byte(`{"sub":"admin","iat":1,"exp":1}`))
	unsigned := header + "." + p
	token := unsigned + "." + sign(unsigned, testSecret)
	if _, err := Verify(token, testSecret); !errors.Is(err, ErrExpiredToken) {
		t.Fatalf("应拒绝过期令牌, got %v", err)
	}
}

func TestVerifyRejectsMalformed(t *testing.T) {
	for _, token := range []string{"", "a.b", "a.b.c.d", "!!!.@@@.###"} {
		if _, err := Verify(token, testSecret); !errors.Is(err, ErrInvalidToken) {
			t.Fatalf("应拒绝畸形令牌 %q, got %v", token, err)
		}
	}
}

func TestSignRejectsInvalidArgs(t *testing.T) {
	if _, err := Sign("admin", "", 24); err == nil {
		t.Fatal("空密钥应报错")
	}
	if _, err := Sign("", testSecret, 24); err == nil {
		t.Fatal("空主体应报错")
	}
	if _, err := Sign("admin", testSecret, 0); err == nil {
		t.Fatal("非正有效期应报错")
	}
}
