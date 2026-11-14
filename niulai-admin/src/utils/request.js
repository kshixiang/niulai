import axios from 'axios'
import { ElMessage } from 'element-plus'
import router from '@/router'
import { getToken, removeToken } from './auth'

const SUCCESS_CODE = 200
const UNAUTHORIZED = 401

const service = axios.create({
  baseURL: import.meta.env.VITE_API_BASE_URL || '/api',
  timeout: 15000,
  headers: { 'Content-Type': 'application/json;charset=utf-8' },
})

let redirecting = false
function toLogin(message) {
  removeToken()
  ElMessage.error(message || '登录已过期，请重新登录')
  // 已在登录页时不再重复导航（登录失败走的就是这条路径）
  if (redirecting || router.currentRoute.value.name === 'Login') return
  redirecting = true
  router.replace({ name: 'Login' }).finally(() => { redirecting = false })
}

// 请求拦截器：自动携带 Token
service.interceptors.request.use(
  (config) => {
    const token = getToken()
    if (token) config.headers.Authorization = `Bearer ${token}`
    return config
  },
  (error) => Promise.reject(error),
)

// 响应拦截器：统一处理后端 { code, msg, data } 约定
service.interceptors.response.use(
  (response) => {
    if (['blob', 'arraybuffer'].includes(response.config.responseType)) return response

    const res = response.data
    if (!res || typeof res.code === 'undefined') return res // 兼容裸数据
    if (res.code === SUCCESS_CODE) return res.data // data 直接交给调用方
    if (res.code === UNAUTHORIZED) {
      toLogin(res.msg ?? res.message)
      return Promise.reject(new Error(res.msg ?? res.message))
    }
    ElMessage.error(res.msg ?? res.message ?? `请求失败（${res.code}）`)
    return Promise.reject(new Error(res.msg ?? res.message ?? `code ${res.code}`))
  },
  (error) => {
    const status = error.response?.status
    if (status === UNAUTHORIZED) {
      toLogin()
      return Promise.reject(error)
    }
    const map = { 400: '请求参数错误', 403: '没有权限', 404: '接口不存在', 500: '服务开小差了', 503: '服务暂不可用' }
    ElMessage.error(map[status] || '网络异常，请稍后重试')
    return Promise.reject(error)
  },
)

export default service
