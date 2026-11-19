import request from '@/utils/request'

export const getRecharges = (params) => request.get('/v1/admin/recharges', { params })
export const reviewRecharge = (id, data) => request.put(`/v1/admin/recharges/${id}/review`, data)
