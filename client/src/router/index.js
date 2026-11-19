import { createRouter, createWebHashHistory } from 'vue-router'
import TabbarLayout from '@/layouts/TabbarLayout.vue'
import { getToken } from '@/utils/auth'

const routes = [
  {
    path: '/',
    component: TabbarLayout,
    redirect: '/home',
    children: [
      { path: 'home',   name: 'home',   component: () => import('@/views/Home.vue'),   meta: { title: '首页' } },
      { path: 'orders', name: 'orders', component: () => import('@/views/Orders.vue'), meta: { title: '订单', requiresAuth: true } },
      { path: 'agent',  name: 'agent',  component: () => import('@/views/Agent.vue'),  meta: { title: '代理', requiresAuth: true } },
      { path: 'mine',   name: 'mine',   component: () => import('@/views/Mine.vue'),   meta: { title: '我的', requiresAuth: true } },
    ],
  },
  { path: '/login', name: 'login', component: () => import('@/views/Login.vue'), meta: { title: '登录' } },
  { path: '/recharge', name: 'recharge', component: () => import('@/views/Recharge.vue'), meta: { title: '极速通道', requiresAuth: true } },
  { path: '/recharge-records', name: 'recharge-records', component: () => import('@/views/RechargeRecords.vue'), meta: { title: '充值记录', requiresAuth: true } },
  // 下单页为独立全屏页（无 tabbar），与 /login 并列
  { path: '/order/:id', name: 'order', component: () => import('@/views/Order.vue'), meta: { title: '下单', requiresAuth: true } },
  // 订单详情页（独立全屏，无 tabbar）
  { path: '/order-detail/:id', name: 'order-detail', component: () => import('@/views/OrderDetail.vue'), meta: { title: '订单详情', requiresAuth: true } },
  { path: '/:pathMatch(.*)*', redirect: '/home' },
]

const router = createRouter({
  history: createWebHashHistory(), // 静态托管 / 内嵌 WebView 不需要服务端 rewrite
  routes,
  scrollBehavior: () => ({ top: 0 }),
})

router.beforeEach((to) => {
  document.title = to.meta.title ? `${to.meta.title} · 牛来充值` : '牛来充值'
  // 用 getToken 而非 useUserStore，避免 router → store → api → request → router 循环依赖
  if (to.meta.requiresAuth && !getToken()) {
    return { name: 'login', query: { redirect: to.fullPath } }
  }
  if (to.name === 'login' && getToken()) return { name: 'home' }
  return true
})

export default router
