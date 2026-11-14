<script setup>
import { useRouter } from 'vue-router'
import { useUserStore } from '@/stores/user'
const router = useRouter()
const user = useUserStore()
const logout = () => { user.logout(); router.replace('/login') }
</script>

<template>
  <div class="mine page--scroll">
    <div class="mine__avatar"><van-icon name="user-o" size="34" /></div>
    <h2>我的</h2>
    <p>账户与设置占位页</p>
    <van-button block round class="mine__btn" @click="logout">退出登录</van-button>
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
p  { font-size: 12px; color: var(--text-muted); margin-bottom: 24px; }
.mine__btn { background: var(--c-surface-2); border: 1px solid var(--border-card); color: var(--text-secondary); }
.mine__btn :deep(.van-button__text) { color: var(--text-secondary); }
</style>
