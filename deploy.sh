#!/usr/bin/env bash
# 牛来充值 一键部署脚本
#
# 用法：在仓库根目录执行
#     ./deploy.sh
#
# 做了什么：
#   1. 确保 .env 存在（首次从 .env.example 复制，需自行修改密码/密钥）
#   2. docker compose 构建并重启 server / web(H5) / admin-web(运营后台)
#   3. 等 MySQL 就绪后，补齐「数据卷已存在时不会自动重跑」的表结构迁移（跳过 001）
#
# 说明：前端与 server 都在各自的 Dockerfile 里构建（npm ci / go build），
#       无需在本机预先 build 前端产物；迁移脚本保证可重复执行（已存在即跳过）。
set -euo pipefail

# 切到仓库根目录（脚本所在目录）
cd "$(dirname "$0")"

# ---------- 1. .env ----------
if [ ! -f .env ]; then
  echo "未找到 .env，已从 .env.example 复制一份。" >&2
  echo "请先编辑 .env 填入真实的密码与密钥（MYSQL_ROOT_PASSWORD / JWT_SECRET / ADMIN_*），再重新运行本脚本。" >&2
  cp .env.example .env
  exit 1
fi

# 从 .env 安全读取单个 KEY 的值：只做文本解析，不把 .env 当作脚本执行，
# 因此文件里有空白行、注释或偶发的异常行都不会让脚本崩溃。
load_env_value() {
  local key="$1" val
  val="$(grep -E "^[[:space:]]*${key}[[:space:]]*=" .env 2>/dev/null | tail -n1 | sed -E "s/^[[:space:]]*${key}[[:space:]]*=[[:space:]]*//")"
  val="${val%\"}"; val="${val#\"}"
  val="${val%\'}"; val="${val#\'}"
  printf '%s' "$val"
}
# docker compose 会自动加载同目录的 .env，迁移步骤只需用到 MySQL 密码
MYSQL_ROOT_PASSWORD="$(load_env_value MYSQL_ROOT_PASSWORD)"
WEB_PORT="$(load_env_value WEB_PORT)"
ADMIN_WEB_PORT="$(load_env_value ADMIN_WEB_PORT)"

# ---------- 2. 选择 docker compose 命令 ----------
if docker compose version >/dev/null 2>&1; then
  DC="docker compose"
elif command -v docker-compose >/dev/null 2>&1; then
  DC="docker-compose"
else
  echo "错误：未检测到 docker compose / docker-compose，请先安装 Docker。" >&2
  exit 1
fi

# ---------- 3. 构建并启动 ----------
echo "==> 构建并启动服务（server / web / admin-web）..."
$DC up -d --build

# ---------- 4. 等 MySQL 就绪，补齐迁移 ----------
echo "==> 等待 MySQL 就绪..."
ready=0
for _ in $(seq 1 60); do
  if $DC exec -T mysql mysqladmin ping -h 127.0.0.1 -uroot "-p${MYSQL_ROOT_PASSWORD}" >/dev/null 2>&1; then
    ready=1
    break
  fi
  sleep 2
done
if [ "$ready" -ne 1 ]; then
  echo "警告：等待 MySQL 超时，跳过迁移步骤，请稍后手动执行 deploy 中的迁移部分。" >&2
else
  echo "==> 应用表结构迁移（跳过 001 初始化脚本，已存在的表结构不重复执行）..."
  for f in server/migrations/*.sql; do
    base="$(basename "$f")"
    [ "$base" = "001_init.sql" ] && continue
    echo "    -> $base"
    if $DC exec -T mysql mysql -uroot "-p${MYSQL_ROOT_PASSWORD}" niulai < "$f"; then
      :
    else
      echo "    警告：$base 执行失败，请检查后手动处理。" >&2
    fi
  done
fi

echo "==> 部署完成。"
echo "    H5 前端：   http://localhost:${WEB_PORT:-80}"
echo "    运营后台： http://localhost:${ADMIN_WEB_PORT:-1009}"
