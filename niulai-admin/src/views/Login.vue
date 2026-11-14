<script setup>
import { ref, reactive } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { login } from '@/api/auth'
import { setToken, setUsername } from '@/utils/auth'

const route = useRoute()
const router = useRouter()

const formRef = ref(null)
const submitting = ref(false)
const form = reactive({ username: '', password: '' })

const rules = {
  username: [{ required: true, message: '请输入管理员账号', trigger: 'blur' }],
  password: [{ required: true, message: '请输入密码', trigger: 'blur' }],
}

async function submit() {
  const ok = await formRef.value.validate().catch(() => false)
  if (!ok) return
  submitting.value = true
  try {
    // 拦截器已解包 data → { token, username }
    const data = await login({ username: form.username, password: form.password })
    setToken(data.token)
    setUsername(data.username)
    router.replace(route.query.redirect || '/products')
  } catch {
    // 错误提示由 request.js 拦截器统一弹出，此处不重复提示
  } finally {
    submitting.value = false
  }
}
</script>

<template>
  <div class="login-page">
    <el-card class="card">
      <div class="brand">
        <h2>牛来 · 运营后台</h2>
        <p>请使用管理员账号登录</p>
      </div>
      <el-form ref="formRef" :model="form" :rules="rules" label-position="top" @keyup.enter="submit">
        <el-form-item label="管理员账号" prop="username">
          <el-input v-model="form.username" placeholder="请输入账号" clearable />
        </el-form-item>
        <el-form-item label="密码" prop="password">
          <el-input v-model="form.password" type="password" placeholder="请输入密码" show-password clearable />
        </el-form-item>
        <el-button type="primary" class="submit" :loading="submitting" @click="submit">登录</el-button>
      </el-form>
    </el-card>
  </div>
</template>

<style scoped>
.login-page {
  display: flex;
  align-items: center;
  justify-content: center;
  height: 100vh;
  background: linear-gradient(135deg, #eef2f9, #dbe4f3);
}
.card {
  width: 400px;
  padding: 12px 12px 4px;
  border-radius: 12px;
}
.brand {
  text-align: center;
  margin-bottom: 24px;
}
.brand h2 {
  margin: 0 0 8px;
  font-size: 20px;
  color: #111827;
}
.brand p {
  margin: 0;
  color: #6b7280;
  font-size: 14px;
}
.submit {
  width: 100%;
  margin-top: 4px;
}
</style>
