import request from '@/utils/request'

// baseURL=/api + /v1/admin/... = 实际请求 /api/v1/admin/...
const ADMIN = '/v1/admin'

// 商品列表：GET /api/v1/admin/products?name=&page=&pageSize=
// 约定返回 data = { list, total }
export const getProducts = (params) => request.get(`${ADMIN}/products`, { params })
export const getProduct = (id) => request.get(`${ADMIN}/products/${id}`)
export const createProduct = (data) => request.post(`${ADMIN}/products`, data)
export const updateProduct = (id, data) => request.put(`${ADMIN}/products/${id}`, data)

// 快捷上下架：PUT /api/v1/admin/products/:id/status，body: { status: 1 | 0 }
export const updateProductStatus = (id, status) =>
  request.put(`${ADMIN}/products/${id}/status`, { status })
