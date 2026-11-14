package service

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"

	"niulai/internal/model"
)

// 渠道相关业务错误
var (
	ErrChannelNotFound = errors.New("渠道不存在")
)

// ChannelService 渠道管理业务逻辑
type ChannelService struct {
	db *gorm.DB
}

func NewChannelService(db *gorm.DB) *ChannelService {
	return &ChannelService{db: db}
}

// CreateChannelRequest 新增渠道请求
type CreateChannelRequest struct {
	Name   string `json:"name" binding:"required,max=128"`
	APIURL string `json:"api_url" binding:"required,max=512"`
	APIKey string `json:"api_key" binding:"required,max=255"`
	Status int8   `json:"status" binding:"oneof=0 1"`
}

// UpdateChannelRequest 更新渠道请求。
// api_url / api_key 为空表示不修改（列表不回显密钥，编辑时无法原样带回）。
type UpdateChannelRequest struct {
	Name   string `json:"name" binding:"required,max=128"`
	APIURL string `json:"api_url" binding:"omitempty,max=512"`
	APIKey string `json:"api_key" binding:"omitempty,max=255"`
	Status int8   `json:"status" binding:"oneof=0 1"`
}

// ListChannelRequest 渠道分页查询
type ListChannelRequest struct {
	Page     int `form:"page"`
	PageSize int `form:"page_size"`
}

// Create 新增渠道
func (s *ChannelService) Create(ctx context.Context, req *CreateChannelRequest) (*model.Channel, error) {
	ch := &model.Channel{
		Name:   req.Name,
		APIURL: req.APIURL,
		APIKey: req.APIKey,
		Status: req.Status,
	}
	if err := s.db.WithContext(ctx).Create(ch).Error; err != nil {
		return nil, fmt.Errorf("创建渠道失败: %w", err)
	}
	return ch, nil
}

// List 分页查询渠道列表
func (s *ChannelService) List(ctx context.Context, page, pageSize int) ([]model.Channel, int64, error) {
	var (
		list  []model.Channel
		total int64
	)
	if err := s.db.WithContext(ctx).Model(&model.Channel{}).Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("统计渠道数量失败: %w", err)
	}
	if err := s.db.WithContext(ctx).
		Order("id DESC").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find(&list).Error; err != nil {
		return nil, 0, fmt.Errorf("查询渠道列表失败: %w", err)
	}
	return list, total, nil
}

// Update 更新渠道信息。api_url / api_key 仅在非空时覆盖，避免编辑时清空密钥。
func (s *ChannelService) Update(ctx context.Context, id uint64, req *UpdateChannelRequest) (*model.Channel, error) {
	var ch model.Channel
	if err := s.db.WithContext(ctx).First(&ch, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrChannelNotFound
		}
		return nil, fmt.Errorf("查询渠道失败: %w", err)
	}
	ch.Name = req.Name
	if req.APIURL != "" {
		ch.APIURL = req.APIURL
	}
	if req.APIKey != "" {
		ch.APIKey = req.APIKey
	}
	ch.Status = req.Status
	if err := s.db.WithContext(ctx).Save(&ch).Error; err != nil {
		return nil, fmt.Errorf("更新渠道失败: %w", err)
	}
	return &ch, nil
}
