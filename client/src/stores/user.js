import { defineStore } from 'pinia'
import { ref } from 'vue'
import { getToken, setToken, removeToken } from '@/utils/auth'

export const useUserStore = defineStore('user', () => {
  const token = ref(getToken())
  const saveToken = (v) => { token.value = v; setToken(v) }
  const logout = () => { token.value = ''; removeToken() }
  return { token, saveToken, logout }
})
