# 项目规则

## 项目概述

牛来充值系统。仓库下有三个平级目录，各自独立，不是 monorepo，也没有根级包管理：

- `client/`：移动端 H5 前端
- `server/`：HTTP 后端服务
- `niulai-admin/`：PC 端运营后台（独立 Vite 项目）

改动前先确认属于哪一侧，不要跨目录混放文件。

## 目录结构

```
niulai/
├── client/                 # 前端
│   ├── src/
│   │   ├── views/          # 页面（Home / Orders / Agent / Mine / Login）
│   │   ├── components/     # 通用组件
│   │   ├── layouts/        # 布局（TabbarLayout）
│   │   ├── router/         # 路由
│   │   ├── stores/         # Pinia 状态
│   │   ├── utils/          # request.js(axios 封装) / auth.js(token)
│   │   └── assets/styles/  # variables.css / global.css
│   ├── .env.development    # VITE_API_BASE_URL=/api
│   └── vite.config.js      # @ 别名、/api 代理到 :8080
├── niulai-admin/           # PC 运营后台（Vue3 + Vite + Element Plus + axios）
│   ├── src/
│   │   ├── api/            # 接口层（product.js）
│   │   ├── utils/          # request.js(axios 封装)
│   │   ├── router/         # 路由
│   │   ├── layout/         # AdminLayout（侧栏 + Header + 主内容）
│   │   ├── views/          # ProductManage / Login / components/ProductForm
│   │   └── styles/         # 全局样式
│   ├── .env.development    # VITE_API_BASE_URL=/api
│   └── vite.config.js      # @ 别名、/api 代理到 :8080（端口 5174）
└── server/                 # 后端
    ├── configs/config.yaml # 配置
    ├── migrations/         # 建表 SQL（首次初始化执行）
    ├── internal/
    │   ├── config/         # viper 配置结构与加载
    │   └── model/          # GORM 数据模型
    ├── pkg/
    │   ├── database/       # MySQL 初始化与迁移
    │   └── response/       # 统一响应封装
    └── main.go             # 入口、路由、优雅关闭
```

## 技术栈

**client**：Vue 3（`<script setup>`）、Vite 6、vue-router 4（hash 模式）、Pinia、Vant 4、axios。ESM（`type: module`），别名 `@` 指向 `src`。

**niulai-admin**：Vue 3（`<script setup>`）、Vite 6、vue-router 4（hash 模式）、Element Plus、axios。ESM（`type: module`），别名 `@` 指向 `src`，开发端口 5174。

**server**：Go 1.22、Gin、GORM + MySQL、Viper、shopspring/decimal。module 名为 `niulai`，内部包用 `niulai/internal/...`、`niulai/pkg/...` 引用。

## 常用命令

```bash
# 前端（在 client/ 下）
npm run dev       # 开发服务器 :5173
npm run build     # 产物输出到 client/dist
npm run preview

# 运营后台（在 niulai-admin/ 下）
npm run dev       # 开发服务器 :5174
npm run build     # 产物输出到 niulai-admin/dist
npm run preview

# 后端（在 server/ 下）
go run .          # 启动服务 :8080（需先启动 MySQL）
go build ./...
```

## 部署

在仓库根目录：

```bash
cp .env.example .env     # 按需修改密码与密钥
docker compose up -d --build
```

- `web`（nginx，端口由 `WEB_PORT` 控制，默认 80）托管前端静态产物，并把 `/api` 反代到 `server:8080`；保留 `/api` 前缀，与 vite 代理行为一致，后端路由需注册在 `/api` 下。
- `server` 不对宿主机暴露端口；MySQL 数据保存在 `mysql-data` 卷。
- 敏感值用 `APP_` 前缀环境变量覆盖（如 `APP_MYSQL_HOST`、`APP_JWT_SECRET`），`internal/config` 已设 `SetEnvKeyReplacer`，不要改回。
- 表结构由 `server/migrations/001_init.sql` 显式管理，仅在 mysql 数据卷首次初始化时执行；生产用 `APP_SERVER_MODE=release` 关闭 GORM 自动迁移。已有数据卷不会重跑初始化脚本，后续改表结构需新增迁移文件。
- `.env` 已被忽略，不要提交真实密码与密钥。

## 接口与数据约定

- 前端所有请求走 `@/utils/request.js`，`baseURL` 取 `VITE_API_BASE_URL`（默认 `/api`）；开发期由 vite 代理转发到 `http://localhost:8080`，联调不要写死绝对地址。
- 后端统一响应体为 `{ "code": 200, "msg": "success", "data": ... }`，用 `pkg/response` 的 `Success / Fail / Page / Error` 返回，禁止在 handler 里直接拼 `c.JSON` 裸结构。
- 业务结果与 HTTP 状态解耦：成功失败 HTTP 均返回 200，由 `code` 区分；`CodeSuccess=200`、`CodeError=500`、`CodeInvalidParam=400`、`CodeUnauthorized=401`、`CodeForbidden=403`、`CodeNotFound=404`。
- 前端响应拦截器以 `code === 200` 为成功并把 `data` 直接交给调用方；`401` 清 token 并跳登录。新增接口沿用该约定，不要另造字段。消息字段以后端 `msg` 为准（`request.js` 里读的是 `res.message`，与后端不一致，涉及错误提示时按 `msg` 对齐）。
- 鉴权用 JWT：请求拦截器自动加 `Authorization: Bearer <token>`，token 存取统一走 `@/utils/auth.js`。
- 业务表内嵌 `BaseModel`（自增 `uint64` 主键、`created_at`/`updated_at`、`gorm.DeletedAt` 软删除）；账本类表 `wallets`、`commission_records` 无软删除列，单独声明基础字段。金额类字段用 `decimal.Decimal`，不要用浮点，列精度按迁移脚本为 `decimal(10,2)`。
- `internal/model` 的列名与类型必须与 `migrations/*.sql` 完全一致；新增或改字段时两处同步。

## 编码约定

- 注释、日志、错误提示用中文；标识符用英文。
- 前端页面按 `views/`、可复用 UI 放 `components/`；路由用懒加载 `() => import(...)`，`meta.title` 用于文档标题。
- 后端配置集中在 `configs/config.yaml`，结构体定义在 `internal/config`，生产敏感值通过 `APP_` 前缀环境变量覆盖，不要把密码/密钥写死进代码。
- 配置默认值在 `internal/config.setDefaults`，必填校验在 `validate`，新增配置项三处同步（yaml、结构体、默认值）。
- 仅非 `release` 模式自动建表迁移，生产环境需显式管理表结构。
- 改动保持最小范围，不引入与当前功能无关的抽象、依赖或配置。
