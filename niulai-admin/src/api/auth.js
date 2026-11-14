import request from '@/utils/request'

// 管理员登录：POST /api/v1/admin/login → { token, username }
export const login = (data) => request.post('/v1/admin/login', data)
