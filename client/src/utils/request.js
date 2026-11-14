import axios from 'axios'
import { showToast } from 'vant'
import { getToken, removeToken } from './auth'
import router from '@/router'

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
  showToast({ message: message || '登录已过期，请重新登录', position: 'bottom' })
  if (redirecting) return
  redirecting = true
  const fullPath = router.currentRoute.value.fullPath
  const redirect = fullPath && !fullPath.startsWith('/login')
    ? `?redirect=${encodeURIComponent(fullPath)}`
    : ''
  router.replace(`/login${redirect}`).finally(() => { redirecting = false })
}

/* 请求拦截器：自动携带 Token */
service.interceptors.request.use(
  (config) => {
    const token = getToken()
    if (token) config.headers.Authorization = `Bearer ${token}`
    // 需要签名 / 设备号 / 时间戳也在这里统一加
    return config
  },
  (error) => Promise.reject(error),
)

/* 响应拦截器：统一处理后端 code */
service.interceptors.response.use(
  (response) => {
    // 下载 / 导出不做业务解包
    if (['blob', 'arraybuffer'].includes(response.config.responseType)) return response

    const res = response.data
    if (!res || typeof res.code === 'undefined') return res   // 兼容裸数据

    // 后端统一响应体的消息字段是 msg（历史代码曾读 message，这里做兼容）
    const msg = res.msg || res.message

    if (res.code === SUCCESS_CODE) return res.data             // 直接把 data 交给调用方
    if (res.code === UNAUTHORIZED) {
      toLogin(msg)
      return Promise.reject(new Error(msg || 'Unauthorized'))
    }
    showToast({ message: msg || `请求失败（${res.code}）`, position: 'bottom' })
    return Promise.reject(new Error(msg || `code ${res.code}`))
  },
  (error) => {
    const status = error.response?.status
    if (status === UNAUTHORIZED) {
      toLogin()
      return Promise.reject(error)
    }
    const map = { 400: '请求参数错误', 403: '没有权限', 404: '接口不存在', 500: '服务开小差了', 503: '服务暂不可用' }
    showToast({ message: map[status] || '网络异常，请稍后重试', position: 'bottom' })
    return Promise.reject(error)
  },
)

export default service
