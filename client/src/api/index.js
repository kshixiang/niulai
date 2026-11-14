import request from '@/utils/request'

// baseURL=/api，因此下面路径实际请求 /api/v1/...
// C 端公开接口，无需鉴权；后端统一返回 data = { list, total }
const CLIENT = '/v1'

// 在售商品 GET /api/v1/products?type=&page=&page_size=
export const getProducts = (params) => request.get(`${CLIENT}/products`, { params })

// 订单 GET /api/v1/orders?status=&page=&page_size=
export const getOrders = (params) => request.get(`${CLIENT}/orders`, { params })

// 代理 GET /api/v1/agents?page=&page_size=
export const getAgents = (params) => request.get(`${CLIENT}/agents`, { params })
