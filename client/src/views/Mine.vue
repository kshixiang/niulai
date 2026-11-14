<script setup>
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useUserStore } from '@/stores/user'
import { getProducts } from '@/api'

const router = useRouter()
const user = useUserStore()
const productTotal = ref(null)

// 账户页展示一项来自服务端的真实数据：当前在售商品数
const load = async () => {
  try {
    const data = await getProducts({ page: 1, page_size: 1 })
    productTotal.value = data?.total ?? 0
  } catch {
    productTotal.value = null
  }
}

// 凭 token 恢复用户资料，同时校验会话是否仍有效（全站唯一的会话校验点）
const loadProfile = async () => {
  if (!user.token) return
  try {
    await user.fetchProfile()
  } catch {
    // 拦截器会统一提示并清理过期 token
  }
}

onMounted(() => { load(); loadProfile() })

const logout = () => { user.logout(); router.replace('/login') }
const goLogin = () => router.replace('/login')
</script>

<template>
  <div class="mine page--scroll">
    <div class="mine__avatar"><van-icon name="user-o" size="34" /></div>
    <h2>{{ user.userInfo?.nickname || '未登录' }}</h2>
    <p class="mine__sub">
      {{ user.userInfo?.phone || '账户与设置' }}
      <span v-if="user.userInfo?.status === 0" class="mine__tag">已禁用</span>
    </p>

    <div class="mine__stat">
      <span class="mine__stat-label">当前在售商品</span>
      <span class="mine__stat-value">{{ productTotal === null ? '—' : `${productTotal} 件` }}</span>
    </div>

    <van-button v-if="user.token" block round class="mine__btn" @click="logout">退出登录</van-button>
    <van-button v-else block round class="mine__btn mine__btn--primary" @click="goLogin">去登录</van-button>
  </div>
</template>

<style scoped>
.mine { padding-top: 40px; text-align: center; }
.mine__avatar {
  width: 72px; height: 72px; margin: 0 auto 16px;
  display: grid; place-items: center; border-radius: 50%;
  background: var(--gradient-brand); color: #fff;
}
h2 { font-size: 20px; color: var(--text-primary); }
.mine__sub { font-size: 12px; color: var(--text-muted); margin-top: 4px; }
.mine__tag {
  margin-left: 8px; padding: 1px 6px; border-radius: 4px;
  font-size: 10px; color: var(--c-danger); background: rgba(255, 92, 108, .16);
}
.mine__stat {
  display: flex; align-items: center; justify-content: space-between;
  margin: 28px 0 24px; padding: 16px;
  border: 1px solid var(--border-card); border-radius: var(--radius-card);
  background: var(--bg-card); text-align: left;
}
.mine__stat-label { font-size: 13px; color: var(--text-secondary); }
.mine__stat-value { font-size: 15px; font-weight: 800; color: var(--c-accent); }
.mine__btn { background: var(--c-surface-2); border: 1px solid var(--border-card); color: var(--text-secondary); }
.mine__btn :deep(.van-button__text) { color: var(--text-secondary); }
.mine__btn--primary { background: var(--gradient-brand); border: none; color: #fff; font-weight: 700; }
.mine__btn--primary :deep(.van-button__text) { color: #fff; }
</style>
