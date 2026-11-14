<script setup>
import { ref, reactive, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { showToast } from 'vant'
import { login, register, getRegisterConfig } from '@/api'
import { useUserStore } from '@/stores/user'

const route = useRoute()
const router = useRouter()
const user = useUserStore()

const mode = ref('login') // login / register
const submitting = ref(false)
const form = reactive({ phone: '', password: '', nickname: '', confirm: '', registerCode: '' })
const registerCodeRequired = ref(false)

// 注册码是否必填由管理后台控制
onMounted(async () => {
  try {
    const data = await getRegisterConfig()
    registerCodeRequired.value = !!data.register_code_required
  } catch {
    // 拉取失败时按非强制处理，提交时由后端兜底校验
  }
})

const validate = () => {
  const account = form.phone.trim()
  if (!account) {
    showToast('请输入登录账号')
    return false
  }
  if (account.length < 3 || account.length > 20) {
    showToast('账号长度需为 3-20 位')
    return false
  }
  if (!form.password) {
    showToast('请输入密码')
    return false
  }
  if (mode.value === 'register') {
    if (form.password.length < 6) {
      showToast('密码至少6位')
      return false
    }
    if (form.password !== form.confirm) {
      showToast('两次输入的密码不一致')
      return false
    }
    if (registerCodeRequired.value && !form.registerCode.trim()) {
      showToast('请输入注册码')
      return false
    }
  }
  return true
}

const submit = async () => {
  if (submitting.value || !validate()) return
  submitting.value = true
  try {
    const account = form.phone.trim()
    const payload = mode.value === 'login'
      ? { phone: account, password: form.password }
      : {
          phone: account,
          password: form.password,
          nickname: form.nickname.trim(),
          register_code: form.registerCode.trim(),
        }
    // 拦截器已解包 data → { token, user }
    const data = mode.value === 'login' ? await login(payload) : await register(payload)
    user.saveToken(data.token)
    user.setUserInfo(data.user)
    showToast(mode.value === 'login' ? '登录成功' : '注册成功')
    router.replace(String(route.query.redirect || '/home'))
  } catch {
    // 错误提示由 request.js 拦截器统一弹出，此处不重复提示
  } finally {
    submitting.value = false
  }
}
</script>

<template>
  <div class="login">
    <div class="login__brand">
      <div class="login__logo">牛</div>
      <h1>牛来充值</h1>
      <p>登录后查看订单与代理权益</p>
    </div>

    <van-tabs v-model:active="mode" class="login__tabs">
      <van-tab name="login" title="登录" />
      <van-tab name="register" title="注册" />
    </van-tabs>

    <van-cell-group inset class="login__form">
      <van-field
        v-model="form.phone"
        type="text"
        label="账号"
        placeholder="请输入登录账号（3-20位，不限手机号）"
        maxlength="20"
        clearable
      />
      <van-field
        v-model="form.password"
        type="password"
        label="密码"
        placeholder="请输入密码"
        clearable
      />
      <template v-if="mode === 'register'">
        <van-field
          v-model="form.registerCode"
          label="注册码"
          :placeholder="registerCodeRequired ? '请输入邀请人的注册码' : '选填，邀请人的注册码'"
          :required="registerCodeRequired"
          maxlength="32"
          clearable
        />
        <van-field
          v-model="form.nickname"
          label="昵称"
          placeholder="选填，不填自动生成"
          maxlength="64"
          clearable
        />
        <van-field
          v-model="form.confirm"
          type="password"
          label="确认密码"
          placeholder="请再次输入密码"
          clearable
        />
      </template>
    </van-cell-group>

    <van-button
      block
      round
      class="login__btn"
      :loading="submitting"
      :disabled="submitting"
      @click="submit"
    >{{ mode === 'login' ? '登录' : '注册并登录' }}</van-button>
  </div>
</template>

<style scoped>
.login {
  height: 100%;
  overflow-y: auto;
  padding: 64px 20px 40px;
  background: var(--bg-page);
  --van-field-label-color: var(--text-secondary);
}
.login__brand { text-align: center; margin-bottom: 28px; }
.login__logo {
  width: 68px; height: 68px; margin: 0 auto 20px;
  display: grid; place-items: center; border-radius: 20px;
  background: var(--gradient-brand); color: #fff; font-size: 30px; font-weight: 800;
  box-shadow: 0 10px 26px rgba(123, 92, 255, .4);
}
h1 { font-size: 22px; color: var(--text-primary); }
p  { font-size: 12px; color: var(--text-muted); margin: 8px 0 0; }
.login__tabs { margin-bottom: 16px; }
.login__form { margin-bottom: 24px; }
.login :deep(.van-field__control::placeholder) { color: var(--text-muted); }
.login__btn { background: var(--gradient-brand); border: none; color: #fff; font-weight: 700; }
.login__btn :deep(.van-button__text) { color: #fff; }
</style>
