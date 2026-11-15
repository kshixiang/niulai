package router

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"niulai/internal/config"
	"niulai/pkg/jwt"
)

const routerTestSecret = "test-secret-for-router-role-isolation"

// newTestRouter 装配与生产一致的真实路由。db 传 nil：以下用例都在到达 service
// 之前被参数校验或鉴权中间件拦截，不会触发任何数据库操作。
func newTestRouter(t *testing.T) *gin.Engine {
	t.Helper()
	gin.SetMode(gin.TestMode)
	config.GlobalConfig = &config.Config{
		JWT:    config.JWTConfig{Secret: routerTestSecret, Expire: 24},
		Admin:  config.AdminConfig{Username: "admin", Password: "secret"},
		Upload: config.UploadConfig{Dir: t.TempDir()},
	}
	r := gin.New()
	Register(r, nil)
	return r
}

func call(t *testing.T, r *gin.Engine, method, path, token string) int {
	t.Helper()
	req := httptest.NewRequest(method, path, nil)
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	var body struct {
		Code int `json:"code"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("解析响应失败: %v, body=%s", err, w.Body.String())
	}
	if body.Code == 0 {
		t.Fatalf("响应缺少业务码, body=%s", w.Body.String())
	}
	return body.Code
}

func TestAuthRoutesRegistered(t *testing.T) {
	r := newTestRouter(t)
	// 无请求体：命中路由后由 bindJSON 返回 400；若路由未注册会得到 404
	if code := call(t, r, http.MethodPost, "/api/v1/auth/register", ""); code != 400 {
		t.Fatalf("注册路由未正确响应参数校验, code=%d", code)
	}
	if code := call(t, r, http.MethodPost, "/api/v1/auth/login", ""); code != 400 {
		t.Fatalf("登录路由未正确响应参数校验, code=%d", code)
	}
	if code := call(t, r, http.MethodGet, "/api/v1/auth/profile", ""); code != 401 {
		t.Fatalf("资料接口未受 UserRequired 保护, code=%d", code)
	}
}

func TestRoleIsolationOnRealRoutes(t *testing.T) {
	r := newTestRouter(t)
	userToken, err := jwt.SignUser(42, routerTestSecret, 24)
	if err != nil {
		t.Fatalf("签发用户令牌失败: %v", err)
	}
	adminToken, err := jwt.Sign("admin", routerTestSecret, 24)
	if err != nil {
		t.Fatalf("签发管理员令牌失败: %v", err)
	}

	if code := call(t, r, http.MethodGet, "/api/v1/admin/products", userToken); code != 401 {
		t.Fatalf("用户令牌不应访问后台接口, code=%d", code)
	}
	if code := call(t, r, http.MethodGet, "/api/v1/admin/users", userToken); code != 401 {
		t.Fatalf("用户令牌不应访问用户管理接口, code=%d", code)
	}
	if code := call(t, r, http.MethodGet, "/api/v1/auth/profile", adminToken); code != 401 {
		t.Fatalf("管理员令牌不应访问用户资料接口, code=%d", code)
	}
}

func TestAdminUserRoutesProtected(t *testing.T) {
	r := newTestRouter(t)
	// 未携带令牌：命中路由后被 AdminRequired 拦截返回 401；若路由未注册会得到 404
	if code := call(t, r, http.MethodGet, "/api/v1/admin/users", ""); code != 401 {
		t.Fatalf("用户列表接口未受 AdminRequired 保护, code=%d", code)
	}
	if code := call(t, r, http.MethodPut, "/api/v1/admin/users/1/status", ""); code != 401 {
		t.Fatalf("用户状态接口未受 AdminRequired 保护, code=%d", code)
	}
}

func TestAdminOrderRoutesProtected(t *testing.T) {
	r := newTestRouter(t)
	// 未携带令牌：命中路由后被 AdminRequired 拦截返回 401；若路由未注册会得到 404
	if code := call(t, r, http.MethodGet, "/api/v1/admin/orders", ""); code != 401 {
		t.Fatalf("订单列表接口未受 AdminRequired 保护, code=%d", code)
	}
	if code := call(t, r, http.MethodGet, "/api/v1/admin/orders/1", ""); code != 401 {
		t.Fatalf("订单详情接口未受 AdminRequired 保护, code=%d", code)
	}
}

func TestCreateOrderRouteProtected(t *testing.T) {
	r := newTestRouter(t)
	// 未携带令牌：命中路由后被 UserRequired 拦截返回 401；若路由未注册会得到 404
	if code := call(t, r, http.MethodPost, "/api/v1/orders", ""); code != 401 {
		t.Fatalf("下单接口未受 UserRequired 保护, code=%d", code)
	}
	// 已登录但无请求体：由 bindJSON 返回 400，不会触发数据库操作
	userToken, err := jwt.SignUser(42, routerTestSecret, 24)
	if err != nil {
		t.Fatalf("签发用户令牌失败: %v", err)
	}
	if code := call(t, r, http.MethodPost, "/api/v1/orders", userToken); code != 400 {
		t.Fatalf("下单接口未正确响应参数校验, code=%d", code)
	}
}
