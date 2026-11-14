import request from '@/utils/request'

const ADMIN = '/v1/admin'

// 渠道分页列表：GET /api/v1/admin/channels → { list, total }
export const fetchChannels = (params) => request.get(`${ADMIN}/channels`, { params })

// 商品表单下拉用：一次取回（上限 100），解包成数组
export const getChannelOptions = async () => {
  const data = await request.get(`${ADMIN}/channels`, { params: { page: 1, page_size: 100 } })
  return Array.isArray(data?.list) ? data.list : []
}

export const createChannel = (data) => request.post(`${ADMIN}/channels`, data)
export const updateChannel = (id, data) => request.put(`${ADMIN}/channels/${id}`, data)
