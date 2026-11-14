import request from '@/utils/request'

// 系统设置：GET /api/v1/admin/settings → { register_code_required }
export const getSettings = () => request.get('/v1/admin/settings')
export const updateSettings = (data) => request.put('/v1/admin/settings', data)
