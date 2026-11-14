import { createRouter, createWebHashHistory } from 'vue-router'
import AdminLayout from '@/layout/AdminLayout.vue'
import { getToken } from '@/utils/auth'

const routes = [
  {
    path: '/',
    component: AdminLayout,
    redirect: '/products',
    children: [
      {
        path: 'products',
        name: 'Products',
        component: () => import('@/views/ProductManage.vue'),
        meta: { title: '商品管理' },
      },
      {
        path: 'channels',
        name: 'Channels',
        component: () => import('@/views/ChannelManage.vue'),
        meta: { title: '渠道管理' },
      },
      {
        path: 'users',
        name: 'Users',
        component: () => import('@/views/UserManage.vue'),
        meta: { title: '用户管理' },
      },
    ],
  },
  {
    path: '/login',
    name: 'Login',
    component: () => import('@/views/Login.vue'),
    meta: { title: '登录' },
  },
]

const router = createRouter({
  history: createWebHashHistory(),
  routes,
})

// 未登录统一跳登录页；已登录访问登录页直接回列表
router.beforeEach((to) => {
  if (to.meta.title) document.title = `${to.meta.title} · 牛来运营后台`

  const loggedIn = !!getToken()
  if (to.name === 'Login') {
    return loggedIn ? { name: 'Products' } : true
  }
  if (!loggedIn) {
    return {
      name: 'Login',
      query: to.fullPath && to.fullPath !== '/' ? { redirect: to.fullPath } : {},
    }
  }
  return true
})

export default router
