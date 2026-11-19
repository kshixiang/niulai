# 商品分类管理 Implementation Plan

> **For agentic workers:** Use superpowers:executing-plans in this session. User approved direct implementation and verification; no additional planning approval is required.

**Goal:** 后台新增和编辑统一商品分类，商品、渠道和客户端显示最新分类名，保留话费业务规则。

**Architecture:** product_categories 保存唯一固定 code 与可编辑 name；现有 products.type/channels.category 保留为 code。提供管理员分类接口，查询响应增加显示名称；表单复用动态分类选择器。

**Tech Stack:** Go/Gin/GORM/MySQL, Vue/Element Plus, Vite.

**Spec:** docs/superpowers/specs/2026-11-19-product-categories-design.md

## Task 1: 分类服务与数据库迁移

Interfaces: CategoryService.List(ctx), Create(ctx, name), Update(ctx, id, name), categoryNames(ctx, db, codes), ensureCategoryExists(ctx, db, code). 数据库表 product_categories(code,name)，code/name 唯一。分类名称去首尾空格、非空、最多32个字符。

- [x] 先写测试：名称校验、创建生成固定标识、重名失败、改名保留标识、分类不存在。
- [x] 运行 go test ./internal/service，观察新测试失败，再实现 model/category.go、service/category_service.go、011_product_categories.sql 和 AutoMigrate 注册。
- [x] 迁移初始化话费、卡券及存量分类；重复运行不覆盖已改名的名称。
- [x] 运行相关测试通过。

## Task 2: API 接入与已有商品渠道兼容

Interfaces: GET/POST /api/v1/admin/categories、PUT /api/v1/admin/categories/:id。列表返回全部分类数组；POST/PUT 接收 {name}；商品响应 type_name、渠道响应 category_name，原有 type/category 不变。

- [x] 先写测试：分类接口拒绝匿名和用户令牌、动态分类被请求校验接受、商品和渠道拒绝不存在的分类、商品查询解析最新名称。
- [x] 观察测试失败，再实现 controller、路由、业务错误映射、商品渠道验证和查询名称解析。
- [x] 运行 go test ./...；确认话费默认字段及上架规则没有回归。

## Task 3: 管理端和客户端界面

Interfaces: 分类 API、CategorySelect.vue(modelValue string, update:modelValue)，CategoryManage.vue；表单每次打开刷新分类列表，管理分类入口可打开分类管理。

- [x] 增加分类管理导航、列表与新增/编辑弹窗；统一复用分类选择器。
- [x] 商品、渠道表单替换固定枚举；表格和客户端显示分类名，筛选仍用固定 code。
- [x] 构建管理端、客户端；浏览器验证新增、重命名、刷新持久化、商品/渠道选项及话费默认字段。

## Task 4: 整体验证与交付

- [x] 在隔离测试数据库执行全部 SQL 迁移及重复执行 011，验证旧分类导入和改名保留。
- [x] 运行完整 Go 测试与两端构建，检查 git diff --check。
- [x] 独立审查最终变更，处理重要问题，记录最终验证结果和部署条件。

## Review Focus

- 重名与并发冲突必须由数据库唯一约束保证，不能仅依赖先查再写。
- 话费分类改名保持默认字段和手机号校验；未知旧分类名称能兼容显示。
- 查询分类名称应批量获取，避免逐商品查询；分类选项不受100条分页截断。
- 分类接口管理员鉴权，公开商品只新增分类显示名称，不泄露渠道和成本数据。
- 分类名称32字符边界、空白与空字符串、显式清空商品类型的校验。

## Progress / Rulings

- Ruling: 按用户“直接做完并验证”在当前 feat/recharge 工作分支实现；不另建 worktree、不替用户提交或推送；现有工作区仅有本任务设计文档。
- Pre-flight: Task 1 code/name 与 Task 2 查询校验、Task 3 选择器提交值一致；旧“话费”标识保留，业务判断无需依赖新名称。

- Task 1–4: complete; evidence recorded in docs/superpowers/validation/2026-11-19-product-categories.md.
- Final review: independent read-only reviewer found one Important code-collation mismatch; fixed via RED→GREEN regression test and full green Go suite; actual HTTP rejects variants on all four product/channel write paths. No deferred minor findings.
- Ruling: local SQL/browser QA used temporary MariaDB 11.4.5 because Docker engine was unavailable; production MySQL 8.0 remains unchanged. MySQL-8-specific container execution was not verified.
