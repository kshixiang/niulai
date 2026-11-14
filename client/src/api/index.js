import request from '@/utils/request'

// baseURL=/api，因此下面路径实际请求 /api/v1/...
// 商品/订单/代理为公开查询接口；注册/登录免鉴权，资料查询需 Bearer token
const CLIENT = '/v1'

// 在售商品 GET /api/v1/products?type=&page=&page_size=
export const getProducts = (params) => request.get(`${CLIENT}/products`, { params })

// 订单 GET /api/v1/orders?status=&page=&page_size=
export const getOrders = (params) => request.get(`${CLIENT}/orders`, { params })

// 代理 GET /api/v1/agents?page=&page_size=
export const getAgents = (params) => request.get(`${CLIENT}/agents`, { params })

// 用户注册 POST /api/v1/auth/register
export const register = (data) => request.post(`${CLIENT}/auth/register`, data)

// 用户登录 POST /api/v1/auth/login
export const login = (data) => request.post(`${CLIENT}/auth/login`, data)

// 当前用户资料 GET /api/v1/auth/profile（需 Bearer token）
export const getProfile = () => request.get(`${CLIENT}/auth/profile`)
