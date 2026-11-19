import request from '@/utils/request'

const CATEGORIES = '/v1/admin/categories'

export const fetchCategories = () => request.get(CATEGORIES)
export const createCategory = (name) => request.post(CATEGORIES, { name })
export const updateCategory = (id, name) => request.put(`${CATEGORIES}/${id}`, { name })
