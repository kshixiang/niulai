import request from '@/utils/request'

// baseURL=/api，因此下面路径实际请求 /api/v1/...
// 商品/订单/代理为公开查询接口；注册/登录免鉴权，资料查询需 Bearer token
const CLIENT = '/v1'

// 在售商品 GET /api/v1/products?type=&page=&page_size=
export const getProducts = (params) => request.get(`${CLIENT}/products`, { params })

// 商品详情（供下单页）GET /api/v1/products/:id
export const getProductDetail = (id) => request.get(`${CLIENT}/products/${id}`)

// 订单列表 GET /api/v1/orders?status=&keyword=&search_by=&page=&page_size=
export const getOrders = (params) => request.get(`${CLIENT}/orders`, { params })

// 订单各状态计数 GET /api/v1/orders/counts → { counts:{0:n,...}, total:n }
export const getOrderCounts = () => request.get(`${CLIENT}/orders/counts`)

// 订单详情 GET /api/v1/orders/:id
export const getOrderDetail = (id) => request.get(`${CLIENT}/orders/${id}`)

// 创建订单 POST /api/v1/orders（需 Bearer token）
export const createOrder = (data) => request.post(`${CLIENT}/orders`, data)

// 代理 GET /api/v1/agents?page=&page_size=
export const getAgents = (params) => request.get(`${CLIENT}/agents`, { params })

// 注册配置 GET /api/v1/register-config → { register_code_required }
export const getRegisterConfig = () => request.get(`${CLIENT}/register-config`)

// 公开汇率 GET /api/v1/exchange-rate → { usdt_cny_rate }（1 USDT 可兑换的 CNY 数量）
export const getExchangeRate = () => request.get(`${CLIENT}/exchange-rate`)

// 用户注册 POST /api/v1/auth/register
export const register = (data) => request.post(`${CLIENT}/auth/register`, data)

// 用户登录 POST /api/v1/auth/login
export const login = (data) => request.post(`${CLIENT}/auth/login`, data)

// 当前用户资料 GET /api/v1/auth/profile（需 Bearer token）
export const getProfile = () => request.get(`${CLIENT}/auth/profile`)

// 修改当前用户昵称 PUT /api/v1/auth/profile（需 Bearer token）→ 最新用户资料
export const updateProfile = (data) => request.put(`${CLIENT}/auth/profile`, data)

// 当前用户资金账户 GET /api/v1/wallet（需 Bearer token）→ { balance, frozen_balance }
export const getWallet = () => request.get(`${CLIENT}/wallet`)
