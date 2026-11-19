package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"unicode/utf8"

	mysqlDriver "github.com/go-sql-driver/mysql"
	"gorm.io/gorm"

	"niulai/internal/model"
)

var (
	ErrInvalidCategoryName = errors.New("分类名称须为1至32个字符")
	ErrCategoryNameExists  = errors.New("分类名称已存在")
	ErrCategoryNotFound    = errors.New("分类不存在")
	ErrInvalidCategory     = errors.New("请选择有效的商品分类")
)

type CategoryService struct{ db *gorm.DB }

func NewCategoryService(db *gorm.DB) *CategoryService { return &CategoryService{db: db} }

func normalizeCategoryName(name string) (string, error) {
	name = strings.TrimSpace(name)
	if name == "" || !utf8.ValidString(name) || utf8.RuneCountInString(name) > 32 {
		return "", ErrInvalidCategoryName
	}
	return name, nil
}

func categoryWriteError(err error) error {
	var mysqlErr *mysqlDriver.MySQLError
	if errors.As(err, &mysqlErr) && mysqlErr.Number == 1062 {
		return ErrCategoryNameExists
	}
	return fmt.Errorf("保存分类失败: %w", err)
}

// List 不分页：分类选择器必须获得完整选项。
func (s *CategoryService) List(ctx context.Context) ([]model.ProductCategory, error) {
	list := make([]model.ProductCategory, 0)
	if err := s.db.WithContext(ctx).Order("id ASC").Find(&list).Error; err != nil {
		return nil, fmt.Errorf("查询分类失败: %w", err)
	}
	return list, nil
}

func (s *CategoryService) Create(ctx context.Context, name string) (*model.ProductCategory, error) {
	name, err := normalizeCategoryName(name)
	if err != nil {
		return nil, err
	}
	var token [12]byte
	if _, err := rand.Read(token[:]); err != nil {
		return nil, fmt.Errorf("生成分类标识失败: %w", err)
	}
	c := &model.ProductCategory{Code: "cat_" + hex.EncodeToString(token[:]), Name: name}
	if err := s.db.WithContext(ctx).Create(c).Error; err != nil {
		return nil, categoryWriteError(err)
	}
	return c, nil
}

func (s *CategoryService) Update(ctx context.Context, id uint64, name string) (*model.ProductCategory, error) {
	name, err := normalizeCategoryName(name)
	if err != nil {
		return nil, err
	}
	var c model.ProductCategory
	if err := s.db.WithContext(ctx).First(&c, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrCategoryNotFound
		}
		return nil, fmt.Errorf("查询分类失败: %w", err)
	}
	if err := s.db.WithContext(ctx).Model(&c).Update("name", name).Error; err != nil {
		return nil, categoryWriteError(err)
	}
	return &c, nil
}

func ensureCategoryExists(ctx context.Context, db *gorm.DB, code string) error {
	if strings.TrimSpace(code) == "" {
		return ErrInvalidCategory
	}
	var c model.ProductCategory
	if err := db.WithContext(ctx).Where("code = ?", code).Take(&c).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrInvalidCategory
		}
		return fmt.Errorf("查询分类失败: %w", err)
	}
	// MySQL 的默认排序规则忽略大小写及尾随空格；固定标识必须精确匹配，
	// 避免保存“话费 ”后绕过基于“话费”标识的下单规则。
	if c.Code != code {
		return ErrInvalidCategory
	}
	return nil
}

// categoryNames 一次批量读取名称，未找到的旧分类回退显示原值。
func categoryNames(ctx context.Context, db *gorm.DB, codes []string) (map[string]string, error) {
	names := make(map[string]string, len(codes))
	unique := make([]string, 0, len(codes))
	for _, code := range codes {
		if code == "" {
			continue
		}
		if _, exists := names[code]; !exists {
			names[code] = code
			unique = append(unique, code)
		}
	}
	if len(unique) == 0 {
		return names, nil
	}
	var categories []model.ProductCategory
	if err := db.WithContext(ctx).Select("code, name").Where("code IN ?", unique).Find(&categories).Error; err != nil {
		return nil, fmt.Errorf("查询分类名称失败: %w", err)
	}
	for _, c := range categories {
		names[c.Code] = c.Name
	}
	return names, nil
}
