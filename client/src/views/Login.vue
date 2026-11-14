<script setup>
import { useRoute, useRouter } from 'vue-router'
import { showToast } from 'vant'
import { useUserStore } from '@/stores/user'

const route = useRoute()
const router = useRouter()
const user = useUserStore()

// 当前 C 端查询接口均为公开接口，无需鉴权；此处保留演示登录入口，
// 后续接入用户体系时再替换为真实登录请求。
const login = () => {
  user.saveToken('mock-token-123456')
  showToast('登录成功')
  router.replace(String(route.query.redirect || '/home'))
}
</script>

<template>
  <div class="login">
    <div class="login__logo">牛</div>
    <h1>牛来充值</h1>
    <p>登录后查看订单与代理权益</p>
    <van-button block round class="login__btn" @click="login">一键模拟登录</van-button>
  </div>
</template>

<style scoped>
.login { min-height: 100%; padding: 120px 28px 0; text-align: center; background: var(--bg-page); }
.login__logo {
  width: 68px; height: 68px; margin: 0 auto 20px;
  display: grid; place-items: center; border-radius: 20px;
  background: var(--gradient-brand); color: #fff; font-size: 30px; font-weight: 800;
  box-shadow: 0 10px 26px rgba(123, 92, 255, .4);
}
h1 { font-size: 22px; color: var(--text-primary); }
p  { font-size: 12px; color: var(--text-muted); margin: 8px 0 40px; }
.login__btn { background: var(--gradient-brand); border: none; color: #fff; font-weight: 700; }
.login__btn :deep(.van-button__text) { color: #fff; }
</style>
