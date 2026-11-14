import request from '@/utils/request'

const ADMIN = '/v1/admin'

// 已注册用户分页列表：GET /api/v1/admin/users → { list, total }
export const getUsers = (params) => request.get(`${ADMIN}/users`, { params })

// 启用/封禁：PUT /api/v1/admin/users/:id/status，body: { status: 1 | 0 }
export const updateUserStatus = (id, status) =>
  request.put(`${ADMIN}/users/${id}/status`, { status })
