import request from '@/utils/request'

// 上传商品封面 → { url }（相对路径，两端 /api 反代均可访问）
export const uploadCover = (file) => {
  const data = new FormData()
  data.append('file', file)
  return request.post('/v1/admin/upload/cover', data)
}
