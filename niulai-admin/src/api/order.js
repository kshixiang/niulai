import request from '@/utils/request'

// baseURL=/api + /v1/admin/... = 实际请求 /api/v1/admin/...
const ADMIN = '/v1/admin'

// 订单分页列表：GET /api/v1/admin/orders?status=&keyword=&page=&page_size= → { list, total }
export const getOrders = (params) => request.get(`${ADMIN}/orders`, { params })

// 订单详情：GET /api/v1/admin/orders/:id（含填写项快照与成本/利润）
export const getOrderDetail = (id) => request.get(`${ADMIN}/orders/${id}`)
