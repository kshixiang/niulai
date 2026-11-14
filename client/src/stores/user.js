import { defineStore } from 'pinia'
import { ref } from 'vue'
import { getToken, setToken, removeToken } from '@/utils/auth'
import { getProfile } from '@/api'

export const useUserStore = defineStore('user', () => {
  const token = ref(getToken())
  const userInfo = ref(null)

  const saveToken = (v) => {
    token.value = v
    setToken(v)
    userInfo.value = null // 换号时清掉旧用户资料
  }
  const setUserInfo = (v) => { userInfo.value = v }
  const logout = () => {
    token.value = ''
    removeToken()
    userInfo.value = null
  }
  // 刷新后凭 token 重新拉取资料，同时校验会话是否仍有效
  const fetchProfile = async () => {
    const data = await getProfile()
    userInfo.value = data
    return data
  }

  return { token, userInfo, saveToken, setUserInfo, logout, fetchProfile }
})
