import request from '@/utils/request'

const BASE = '/v1/admin/register-codes'

export const getRegisterCodes = (params) => request.get(BASE, { params })
export const generateRegisterCodes = (data) => request.post(BASE, data)
export const updateRegisterCodeStatus = (id, status) =>
  request.put(`${BASE}/${id}/status`, { status })
