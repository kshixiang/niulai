<script setup>
import { ref, onMounted } from 'vue'
import { getOrders } from '@/api'

/* 订单状态与后端 model.OrderStatus* 保持一致 */
const statusMap = {
  0: { text: '待支付', tone: 'warn'  },
  1: { text: '已支付', tone: 'brand' },
  2: { text: '充值中', tone: 'brand' },
  3: { text: '成功',   tone: 'ok'    },
  4: { text: '失败',   tone: 'fail'  },
  5: { text: '已退款', tone: 'muted' },
}
const statusOf = (s) => statusMap[s] || { text: '未知', tone: 'muted' }
const fmtTime = (v) => (v ? String(v).replace('T', ' ').slice(0, 16) : '')

const list = ref([])
const total = ref(0)
const loading = ref(false)

const load = async () => {
  loading.value = true
  try {
    const data = await getOrders({ page: 1, page_size: 50 })
    list.value = Array.isArray(data?.list) ? data.list : []
    total.value = data?.total ?? 0
  } catch {
    list.value = []
  } finally {
    loading.value = false
  }
}
onMounted(load)
</script>

<template>
  <div class="orders page--scroll">
    <header class="orders__head">
      <h2>订单</h2>
      <span class="orders__count">共 {{ total }} 笔</span>
    </header>

    <div v-if="loading" class="orders__tip">
      <van-loading size="20" /> 加载中…
    </div>

    <ul v-else class="orders__list">
      <li v-for="item in list" :key="item.id" class="order">
        <div class="order__top">
          <span class="order__name">{{ item.product_name || `商品 #${item.id}` }}</span>
          <span class="order__status" :class="`order__status--${statusOf(item.status).tone}`">
            {{ statusOf(item.status).text }}
          </span>
        </div>
        <div class="order__row">
          <span class="order__no">{{ item.order_no }}</span>
          <span class="order__amount">¥{{ item.amount }}</span>
        </div>
        <div class="order__foot">
          <span>充值账号：{{ item.account || '—' }}</span>
          <span>{{ fmtTime(item.created_at) }}</span>
        </div>
      </li>
    </ul>

    <p v-if="!loading && !list.length" class="orders__tip">暂无订单</p>
  </div>
</template>

<style scoped>
.orders__head { display: flex; align-items: baseline; justify-content: space-between; margin-bottom: 16px; }
h2 { font-size: 20px; color: var(--text-primary); }
.orders__count { font-size: 12px; color: var(--text-muted); }

.orders__tip {
  display: flex; align-items: center; justify-content: center; gap: 8px;
  padding: 48px 0; font-size: 12px; color: var(--text-muted);
}
.orders__list { display: flex; flex-direction: column; gap: 12px; }

.order {
  padding: 14px;
  border: 1px solid var(--border-card); border-radius: var(--radius-card);
  background: var(--bg-card);
}
.order__top { display: flex; align-items: center; justify-content: space-between; gap: 8px; }
.order__name { font-size: 14px; font-weight: 700; color: var(--text-primary); overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.order__status { flex: 0 0 auto; font-size: 11px; font-weight: 700; }
.order__status--warn  { color: var(--c-accent); }
.order__status--brand { color: var(--c-brand-2); }
.order__status--ok    { color: var(--c-success); }
.order__status--fail  { color: var(--c-danger); }
.order__status--muted { color: var(--text-muted); }

.order__row { display: flex; align-items: center; justify-content: space-between; margin-top: 10px; }
.order__no { font-size: 11px; color: var(--text-muted); }
.order__amount { font-size: 16px; font-weight: 800; color: var(--c-accent); }

.order__foot {
  display: flex; align-items: center; justify-content: space-between;
  margin-top: 10px; padding-top: 10px; border-top: 1px solid var(--border-card);
  font-size: 11px; color: var(--text-muted);
}
</style>
