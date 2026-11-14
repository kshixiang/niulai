package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"niulai/internal/model"
)

// 用户认证相关业务错误。
var (
	ErrPhoneExists       = errors.New("该账号已注册")
	ErrInvalidCredential = errors.New("账号或密码错误")
	ErrPasswordTooLong   = errors.New("密码过长，请缩短后重试")
	ErrUserDisabled      = errors.New("账号已被禁用，请联系客服")
	ErrUserNotFound      = errors.New("用户不存在")
)

// bcrypt 对密码有 72 字节上限；validator 的 max 按字符数计算，
// 含非 ASCII 字符时仍需按字节兜底，避免生成哈希时报 500。
const maxPasswordBytes = 72

// ---------- 请求参数 ----------

// UserRegisterRequest 用户注册请求。登录账号必填，长度 3~20，不限手机号格式。
type UserRegisterRequest struct {
	Phone    string `json:"phone"    binding:"required,min=3,max=20"`
	Password string `json:"password" binding:"required,min=6,max=72"`
	Nickname string `json:"nickname" binding:"omitempty,max=64"`
}

// UserLoginRequest 用户登录请求。
type UserLoginRequest struct {
	Phone    string `json:"phone"    binding:"required,min=3,max=20"`
	Password string `json:"password" binding:"required,max=72"`
}

// ---------- 对外视图 ----------

// UserView 用户信息对外视图，绝不包含密码。
type UserView struct {
	ID        uint64    `json:"id"`
	Phone     string    `json:"phone"`
	Nickname  string    `json:"nickname"`
	Status    int8      `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}

func toUserView(u *model.User) *UserView {
	return &UserView{
		ID:        u.ID,
		Phone:     u.Phone,
		Nickname:  u.Nickname,
		Status:    u.Status,
		CreatedAt: u.CreatedAt,
	}
}

// ---------- 服务 ----------

// UserService C 端用户认证服务。
type UserService struct {
	db *gorm.DB
}

func NewUserService(db *gorm.DB) *UserService {
	return &UserService{db: db}
}

// Register 注册用户并返回对外视图，controller 据此签发令牌实现注册即登录。
func (s *UserService) Register(ctx context.Context, req *UserRegisterRequest) (*UserView, error) {
	if len([]byte(req.Password)) > maxPasswordBytes {
		return nil, ErrPasswordTooLong
	}
	if s.accountExists(ctx, req.Phone) {
		return nil, ErrPhoneExists
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("密码加密失败: %w", err)
	}

	nickname := req.Nickname
	if nickname == "" {
		if len(req.Phone) >= 4 {
			nickname = "用户" + req.Phone[len(req.Phone)-4:]
		} else {
			nickname = "用户" + req.Phone
		}
	}

	u := &model.User{
		Phone:    req.Phone,
		Password: string(hash),
		Nickname: nickname,
		Status:   model.UserStatusNormal, // 零值是「封禁」，必须显式赋值
	}
	if err := s.db.WithContext(ctx).Create(u).Error; err != nil {
		// 并发下唯一索引兜底：账号已被占用时统一返回已注册提示，避免 500
		if s.accountExists(ctx, req.Phone) {
			return nil, ErrPhoneExists
		}
		return nil, fmt.Errorf("创建用户失败: %w", err)
	}
	return toUserView(u), nil
}

// accountExists 判重需 Unscoped：软删除过的账号仍占用唯一索引，否则重注会触发数据库报错。
func (s *UserService) accountExists(ctx context.Context, account string) bool {
	var count int64
	return s.db.WithContext(ctx).Unscoped().Model(&model.User{}).
		Where("phone = ?", account).Count(&count).Error == nil && count > 0
}

// Login 校验账号与密码，未命中与密码错误统一返回同一错误，避免账号枚举。
func (s *UserService) Login(ctx context.Context, req *UserLoginRequest) (*UserView, error) {
	var u model.User
	if err := s.db.WithContext(ctx).Where("phone = ?", req.Phone).First(&u).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrInvalidCredential
		}
		return nil, fmt.Errorf("查询用户失败: %w", err)
	}
	if bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(req.Password)) != nil {
		return nil, ErrInvalidCredential
	}
	if u.Status != model.UserStatusNormal {
		return nil, ErrUserDisabled
	}
	return toUserView(&u), nil
}

// GetProfile 按用户ID查询资料，供登录后恢复会话使用。
func (s *UserService) GetProfile(ctx context.Context, userID uint64) (*UserView, error) {
	var u model.User
	if err := s.db.WithContext(ctx).First(&u, userID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("查询用户失败: %w", err)
	}
	return toUserView(&u), nil
}
