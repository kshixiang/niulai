<script setup>
import { ref, computed, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { showToast } from 'vant'
import { getOrderDetail } from '@/api'

const route = useRoute()
const router = useRouter()

const loading = ref(true)
const order = ref(null)

// 详情已单列「充值账号」，这里只展示下单时填写的其余项，避免重复
const extraFields = computed(() =>
  (order.value?.fields || []).filter((f) => f.value !== order.value?.account),
)

/* 订单状态：与后端 model.OrderStatus* 保持一致（待支付/已支付/充值中/成功/失败/已退款） */
const statusMap = {
  0: { text: '待支付', tone: 'warn'  },
  1: { text: '已支付', tone: 'brand' },
  2: { text: '充值中', tone: 'brand' },
  3: { text: '成功',   tone: 'ok'    },
  4: { text: '失败',   tone: 'fail'  },
  5: { text: '已退款', tone: 'muted' },
}
const statusOf = (s) => statusMap[s] || { text: '未知', tone: 'muted' }

const fmtMoney = (v) => (v == null ? '0.00' : Number(v).toFixed(2))
const fmtTime = (v) => {
  if (!v) return '—'
  const d = new Date(v)
  if (isNaN(d.getTime())) return String(v)
  const p = (n) => String(n).padStart(2, '0')
  return `${d.getFullYear()}/${p(d.getMonth() + 1)}/${p(d.getDate())} ${p(d.getHours())}:${p(d.getMinutes())}:${p(d.getSeconds())}`
}

const copyNo = async (no) => {
  if (!no) return
  try {
    await navigator.clipboard.writeText(no)
    showToast('订单号已复制')
  } catch {
    showToast('复制失败，请手动复制')
  }
}

const goBack = () => {
  if (window.history.state?.back) router.back()
  else router.replace('/orders')
}

onMounted(async () => {
  try {
    const data = await getOrderDetail(route.params.id)
    order.value = data || null
  } catch {
    order.value = null
  } finally {
    loading.value = false
  }
})
</script>

<template>
  <div class="detail">
    <van-nav-bar title="订单详情" left-arrow class="detail__nav" @click-left="goBack" />

    <div class="detail__body">
      <div v-if="loading" class="state">
        <van-loading size="20" /> 加载中…
      </div>

      <div v-else-if="order" class="detail__content">
        <!-- 商品 + 状态 -->
        <section class="hero">
          <h2 class="hero__name">{{ order.product_name || `商品 #${order.id}` }}</h2>
          <span class="hero__status" :class="`tone--${statusOf(order.status).tone}`">
            {{ statusOf(order.status).text }}
          </span>
        </section>

        <!-- 订单号 -->
        <section class="block">
          <div class="row">
            <span class="row__label">订单编号</span>
            <span class="row__value row__value--mono">
              {{ order.order_no || '—' }}
              <button type="button" class="copy" @click="copyNo(order.order_no)">
                <van-icon name="orders-o" />
              </button>
            </span>
          </div>
          <div class="row">
            <span class="row__label">充值账号</span>
            <span class="row__value">{{ order.account || '—' }}</span>
          </div>
          <div v-for="(f, i) in extraFields" :key="i" class="row">
            <span class="row__label">{{ f.label }}</span>
            <span class="row__value">{{ f.value }}</span>
          </div>
          <div class="row">
            <span class="row__label">下单时间</span>
            <span class="row__value">{{ fmtTime(order.created_at) }}</span>
          </div>
        </section>

        <!-- 金额明细 -->
        <section class="block">
          <div class="row">
            <span class="row__label">订单金额</span>
            <span class="row__value row__value--strong">¥{{ fmtMoney(order.amount) }}</span>
          </div>
          <div class="row">
            <span class="row__label">到账金额</span>
            <span class="row__value row__value--strong">¥{{ fmtMoney(order.received_amount) }}</span>
          </div>
          <div class="row">
            <span class="row__label">剩余金额</span>
            <span class="row__value row__value--strong">¥{{ fmtMoney(order.remaining_amount) }}</span>
          </div>
        </section>
      </div>

      <div v-else class="state">订单不存在或已删除</div>
    </div>
  </div>
</template>

<style scoped>
.detail {
  height: 100%;
  display: flex;
  flex-direction: column;
  background: var(--bg-page);
}
.detail__nav {
  flex: 0 0 auto;
  background: var(--bg-card);
}
.detail :deep(.van-nav-bar__title) { color: var(--text-primary); font-weight: 700; }
.detail :deep(.van-nav-bar__arrow) { color: var(--text-primary); }

.detail__body { flex: 1; overflow-y: auto; padding: 16px; }
.state {
  display: flex; align-items: center; justify-content: center; gap: 8px;
  padding: 48px 0; font-size: 12px; color: var(--text-muted);
}

.hero {
  display: flex; align-items: center; justify-content: space-between; gap: 12px;
  margin-bottom: 16px;
}
.hero__name { font-size: 20px; font-weight: 800; color: var(--text-primary); }
.hero__status { font-size: 13px; font-weight: 700; flex: 0 0 auto; }
.tone--warn  { color: var(--c-accent); }
.tone--brand { color: var(--c-brand-2); }
.tone--ok    { color: var(--c-success); }
.tone--fail  { color: var(--c-danger); }
.tone--muted { color: var(--text-muted); }

.block {
  margin-bottom: 12px; padding: 4px 14px;
  border: 1px solid var(--border-card); border-radius: var(--radius-card);
  background: var(--bg-card);
}
.row {
  display: flex; align-items: center; justify-content: space-between; gap: 12px;
  padding: 12px 0; font-size: 13px;
}
.row + .row { border-top: 1px solid var(--border-card); }
.row__label { color: var(--text-secondary); flex: 0 0 auto; }
.row__value {
  color: var(--text-primary); text-align: right;
  overflow: hidden; text-overflow: ellipsis; white-space: nowrap;
}
.row__value--mono { font-variant-numeric: tabular-nums; letter-spacing: .3px; }
.row__value--strong { font-weight: 800; }
.copy {
  display: inline-flex; vertical-align: middle; margin-left: 6px;
  background: none; border: none; color: var(--text-secondary);
}
</style>
