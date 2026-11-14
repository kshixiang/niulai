package service

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"

	"niulai/internal/model"
)

// AdminUserService 运营后台用户管理业务逻辑。
// 管理对象是已通过 C 端注册的用户，后台不负责创建，仅提供查询与启用/封禁。
type AdminUserService struct {
	db *gorm.DB
}

func NewAdminUserService(db *gorm.DB) *AdminUserService {
	return &AdminUserService{db: db}
}

// ListUserRequest 用户分页查询。keyword 同时匹配登录账号与昵称。
type ListUserRequest struct {
	Keyword  string `form:"keyword"`
	Status   *int8  `form:"status"`
	Page     int    `form:"page"`
	PageSize int    `form:"page_size"`
}

// UpdateUserStatusRequest 启用/封禁切换请求，状态仅允许 1正常 0封禁。
type UpdateUserStatusRequest struct {
	Status int8 `json:"status" binding:"oneof=0 1"`
}

// List 分页查询用户，返回不含密码的视图（复用 C 端 UserView）。
func (s *AdminUserService) List(ctx context.Context, req *ListUserRequest) ([]*UserView, int64, error) {
	var (
		users []model.User
		total int64
	)
	q := s.db.WithContext(ctx).Model(&model.User{})
	if req.Keyword != "" {
		like := "%" + req.Keyword + "%"
		q = q.Where("phone LIKE ? OR nickname LIKE ?", like, like)
	}
	if req.Status != nil {
		q = q.Where("status = ?", *req.Status)
	}
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("统计用户数量失败: %w", err)
	}
	if err := q.Order("id DESC").
		Offset((req.Page - 1) * req.PageSize).
		Limit(req.PageSize).
		Find(&users).Error; err != nil {
		return nil, 0, fmt.Errorf("查询用户列表失败: %w", err)
	}
	list := make([]*UserView, 0, len(users))
	for i := range users {
		list = append(list, toUserView(&users[i]))
	}
	return list, total, nil
}

// UpdateStatus 启用/封禁用户，返回更新后的视图。
func (s *AdminUserService) UpdateStatus(ctx context.Context, id uint64, status int8) (*UserView, error) {
	var u model.User
	if err := s.db.WithContext(ctx).First(&u, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("查询用户失败: %w", err)
	}
	if err := s.db.WithContext(ctx).Model(&u).Update("status", status).Error; err != nil {
		return nil, fmt.Errorf("更新用户状态失败: %w", err)
	}
	u.Status = status
	return toUserView(&u), nil
}
