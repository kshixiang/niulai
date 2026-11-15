import request from '@/utils/request'

// 系统设置：GET /api/v1/admin/settings → { register_code_required, usdt_cny_rate }
// PUT 可传 register_code_required 和/或 usdt_cny_rate，至少一项，未传的字段保持不变
export const getSettings = () => request.get('/v1/admin/settings')
export const updateSettings = (data) => request.put('/v1/admin/settings', data)
