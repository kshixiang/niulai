// Package jwt 提供后台管理员与 C 端用户登录所需的 HS256 令牌签发与校验。
// 仅依赖标准库，避免为简单的签发/校验场景引入外部依赖。
package jwt

import (
	"crypto/hmac"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
)

var (
	// ErrInvalidToken 令牌格式、算法或签名不合法。
	ErrInvalidToken = errors.New("无效的令牌")
	// ErrExpiredToken 令牌已过期。
	ErrExpiredToken = errors.New("登录已过期")
)

// 令牌角色：管理员与 C 端用户共用同一密钥，靠 role 区分，
// 由各自的鉴权中间件校验，避免用户令牌访问后台接口或反之。
const (
	RoleAdmin = "admin"
	RoleUser  = "user"
)

// header 固定为 HS256；Verify 会重新解析并只接受该算法，
// 以拒绝 alg=none 与算法混淆攻击。
var header = base64.RawURLEncoding.EncodeToString([]byte(`{"alg":"HS256","typ":"JWT"}`))

// Claims 令牌载荷，只包含本系统需要的最小字段。
// Role 用 omitempty 兼容无角色的历史令牌，校验交由中间件处理。
type Claims struct {
	Sub  string `json:"sub"`            // 主体：管理员用户名或用户ID
	Iat  int64  `json:"iat"`            // 签发时间（Unix 秒）
	Exp  int64  `json:"exp"`            // 过期时间（Unix 秒）
	Role string `json:"role,omitempty"` // 角色
}

// Sign 签发管理员令牌（后台登录使用），expireHours 为有效期（小时）。
func Sign(subject, secret string, expireHours int) (string, error) {
	return signToken(subject, RoleAdmin, secret, expireHours)
}

// SignUser 签发 C 端用户令牌，主体为用户ID，角色为 RoleUser。
func SignUser(userID uint64, secret string, expireHours int) (string, error) {
	if userID == 0 {
		return "", errors.New("用户ID不能为空")
	}
	return signToken(strconv.FormatUint(userID, 10), RoleUser, secret, expireHours)
}

func signToken(subject, role, secret string, expireHours int) (string, error) {
	if secret == "" {
		return "", errors.New("JWT 密钥不能为空")
	}
	if subject == "" {
		return "", errors.New("令牌主体不能为空")
	}
	if expireHours <= 0 {
		return "", errors.New("令牌有效期必须大于 0")
	}

	now := time.Now().Unix()
	payload, err := json.Marshal(Claims{
		Sub:  subject,
		Iat:  now,
		Exp:  now + int64(expireHours)*3600,
		Role: role,
	})
	if err != nil {
		return "", fmt.Errorf("序列化令牌载荷失败: %w", err)
	}

	unsigned := header + "." + base64.RawURLEncoding.EncodeToString(payload)
	return unsigned + "." + sign(unsigned, secret), nil
}

// Verify 校验格式、算法、签名与有效期，成功后返回载荷。
func Verify(token, secret string) (*Claims, error) {
	if secret == "" {
		return nil, errors.New("JWT 密钥不能为空")
	}

	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return nil, ErrInvalidToken
	}

	// 必须重新校验头部算法，不能只比对签名
	rawHeader, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return nil, ErrInvalidToken
	}
	var h struct {
		Alg string `json:"alg"`
		Typ string `json:"typ"`
	}
	if err := json.Unmarshal(rawHeader, &h); err != nil {
		return nil, ErrInvalidToken
	}
	if h.Alg != "HS256" || h.Typ != "JWT" {
		return nil, ErrInvalidToken
	}

	unsigned := parts[0] + "." + parts[1]
	expected := sign(unsigned, secret)
	if subtle.ConstantTimeCompare([]byte(parts[2]), []byte(expected)) != 1 {
		return nil, ErrInvalidToken
	}

	rawPayload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, ErrInvalidToken
	}
	var claims Claims
	if err := json.Unmarshal(rawPayload, &claims); err != nil {
		return nil, ErrInvalidToken
	}
	if claims.Sub == "" {
		return nil, ErrInvalidToken
	}
	if claims.Exp <= time.Now().Unix() {
		return nil, ErrExpiredToken
	}
	return &claims, nil
}

func sign(unsigned, secret string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(unsigned))
	return base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}
