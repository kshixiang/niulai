import { createRouter, createWebHashHistory } from 'vue-router'
import TabbarLayout from '@/layouts/TabbarLayout.vue'

const routes = [
  {
    path: '/',
    component: TabbarLayout,
    redirect: '/home',
    children: [
      { path: 'home',   name: 'home',   component: () => import('@/views/Home.vue'),   meta: { title: '首页' } },
      { path: 'orders', name: 'orders', component: () => import('@/views/Orders.vue'), meta: { title: '订单' } },
      { path: 'agent',  name: 'agent',  component: () => import('@/views/Agent.vue'),  meta: { title: '代理' } },
      { path: 'mine',   name: 'mine',   component: () => import('@/views/Mine.vue'),   meta: { title: '我的' } },
    ],
  },
  { path: '/login', name: 'login', component: () => import('@/views/Login.vue'), meta: { title: '登录' } },
  { path: '/:pathMatch(.*)*', redirect: '/home' },
]

const router = createRouter({
  history: createWebHashHistory(), // 静态托管 / 内嵌 WebView 不需要服务端 rewrite
  routes,
  scrollBehavior: () => ({ top: 0 }),
})

router.beforeEach((to) => {
  document.title = to.meta.title ? `${to.meta.title} · 牛来充值` : '牛来充值'
  // 需要鉴权时给路由加 meta.requiresAuth，再引入 getToken：
  // if (to.meta.requiresAuth && !getToken()) return { name: 'login', query: { redirect: to.fullPath } }
  return true
})

export default router
