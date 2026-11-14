<script setup>
import { useRouter } from 'vue-router'

const router = useRouter()
const goHome = () => router.replace('/products')
</script>

<template>
  <div class="login-page">
    <el-card class="card">
      <h2>牛来 · 运营后台</h2>
      <p>登录状态已失效，请重新登录。</p>
      <el-button type="primary" @click="goHome">重新进入</el-button>
    </el-card>
  </div>
</template>

<style scoped>
.login-page {
  display: flex;
  align-items: center;
  justify-content: center;
  height: 100vh;
  background: #f5f7fa;
}
.card {
  width: 360px;
  text-align: center;
}
.card p {
  color: #6b7280;
  margin-bottom: 20px;
}
</style>
