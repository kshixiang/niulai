package middleware

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"niulai/internal/config"
	"niulai/pkg/jwt"
)

const authTestSecret = "test-secret-for-middleware-role-isolation"

// runProtected 用指定中间件保护一条测试路由，返回响应记录器。
// onOK 在鉴权通过后执行，可用于断言写入上下文的身份。
func runProtected(mw gin.HandlerFunc, token string, onOK func(*gin.Context)) *httptest.ResponseRecorder {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/protected", mw, func(c *gin.Context) {
		if onOK != nil {
			onOK(c)
		}
		c.JSON(http.StatusOK, gin.H{"code": 200})
	})

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func bizCode(t *testing.T, w *httptest.ResponseRecorder) int {
	t.Helper()
	var body struct {
		Code int `json:"code"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("解析响应失败: %v, body=%s", err, w.Body.String())
	}
	return body.Code
}

func setupAuthConfig() {
	config.GlobalConfig = &config.Config{JWT: config.JWTConfig{Secret: authTestSecret}}
}

func TestAdminRequiredRejectsUserToken(t *testing.T) {
	setupAuthConfig()
	userToken, err := jwt.SignUser(42, authTestSecret, 24)
	if err != nil {
		t.Fatalf("签发用户令牌失败: %v", err)
	}
	w := runProtected(AdminRequired(), userToken, nil)
	if bizCode(t, w) != 401 {
		t.Fatalf("用户令牌不应通过后台鉴权, body=%s", w.Body.String())
	}
}

func TestAdminRequiredAcceptsAdminToken(t *testing.T) {
	setupAuthConfig()
	adminToken, err := jwt.Sign("admin", authTestSecret, 24)
	if err != nil {
		t.Fatalf("签发管理员令牌失败: %v", err)
	}
	var username string
	w := runProtected(AdminRequired(), adminToken, func(c *gin.Context) {
		username = c.GetString("admin_username")
	})
	if bizCode(t, w) != 200 {
		t.Fatalf("管理员令牌应通过后台鉴权, body=%s", w.Body.String())
	}
	if username != "admin" {
		t.Fatalf("管理员用户名不符: %s", username)
	}
}

func TestUserRequiredRejectsAdminToken(t *testing.T) {
	setupAuthConfig()
	adminToken, err := jwt.Sign("admin", authTestSecret, 24)
	if err != nil {
		t.Fatalf("签发管理员令牌失败: %v", err)
	}
	w := runProtected(UserRequired(), adminToken, nil)
	if bizCode(t, w) != 401 {
		t.Fatalf("管理员令牌不应通过用户鉴权, body=%s", w.Body.String())
	}
}

func TestUserRequiredSetsUserID(t *testing.T) {
	setupAuthConfig()
	userToken, err := jwt.SignUser(42, authTestSecret, 24)
	if err != nil {
		t.Fatalf("签发用户令牌失败: %v", err)
	}
	var userID uint64
	w := runProtected(UserRequired(), userToken, func(c *gin.Context) {
		userID = c.GetUint64("user_id")
	})
	if bizCode(t, w) != 200 {
		t.Fatalf("用户令牌应通过用户鉴权, body=%s", w.Body.String())
	}
	if userID != 42 {
		t.Fatalf("用户ID不符: %d", userID)
	}
}
