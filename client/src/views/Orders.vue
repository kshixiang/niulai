<script setup>
import { ref, watch, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { showToast } from 'vant'
import { getOrders, getOrderCounts } from '@/api'

const router = useRouter()

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

/* 顶部大 tab：订单中心 / 确认中心 */
const topTab = ref('center') // center | confirm

/* 状态子 tab：null 表示「全部」 */
const statusTabs = [
  { key: 'all',       label: '全部',   status: null },
  { key: 'pending',   label: '待支付', status: 0 },
  { key: 'paid',      label: '已支付', status: 1 },
  { key: 'recharging', label: '充值中', status: 2 },
  { key: 'success',   label: '成功',   status: 3 },
  { key: 'failed',    label: '失败',   status: 4 },
  { key: 'refunded',  label: '已退款', status: 5 },
]
const activeStatus = ref(null)

/* 搜索：下拉选择搜索维度 + 关键字 */
const searchBy = ref('all') // all | order_no | account
const keyword = ref('')
const searchByOptions = [
  { value: 'all',      label: '全部' },
  { value: 'order_no', label: '订单号' },
  { value: 'account',  label: '主充值号码' },
]

/* 列表 + 计数 */
const list = ref([])
const counts = ref({})
const total = ref(0)
const loading = ref(false)

const maskNo = (no) => {
  if (!no) return '—'
  if (no.length <= 12) return no
  return no.slice(0, 6) + '...' + no.slice(-6)
}
const fmtMoney = (v) => (v == null ? '0.00' : Number(v).toFixed(2))
const fmtTime = (v) => {
  if (!v) return ''
  const d = new Date(v)
  if (isNaN(d.getTime())) return String(v)
  const p = (n) => String(n).padStart(2, '0')
  return `${d.getFullYear()}/${p(d.getMonth() + 1)}/${p(d.getDate())} ${p(d.getHours())}:${p(d.getMinutes())}:${p(d.getSeconds())}`
}
const badgeOf = (status) => (status == null ? total.value : (counts.value[status] || 0))

const copyNo = async (no) => {
  if (!no) return
  try {
    await navigator.clipboard.writeText(no)
    showToast('订单号已复制')
  } catch {
    showToast('复制失败，请手动复制')
  }
}
const goDetail = (item) => router.push(`/order-detail/${item.id}`)

const loadCounts = async () => {
  try {
    const data = await getOrderCounts()
    counts.value = data?.counts || {}
    total.value = data?.total ?? 0
  } catch { /* 拦截器已提示 */ }
}

const loadList = async () => {
  loading.value = true
  try {
    const data = await getOrders({
      status: activeStatus.value,
      keyword: keyword.value.trim() || undefined,
      search_by: searchBy.value,
      page: 1,
      page_size: 50,
    })
    list.value = Array.isArray(data?.list) ? data.list : []
    total.value = data?.total ?? 0
  } catch {
    list.value = []
  } finally {
    loading.value = false
  }
}

/* 状态 / 搜索维度变化立即刷新；关键字做轻量防抖 */
watch(activeStatus, loadList)
watch(searchBy, loadList)

let timer = null
const onSearch = () => {
  clearTimeout(timer)
  timer = setTimeout(loadList, 300)
}

onMounted(async () => {
  await Promise.all([loadCounts(), loadList()])
})
</script>

<template>
  <div class="orders page--scroll">
    <!-- 顶部大 tab -->
    <nav class="top-tabs">
      <button type="button" :class="{ active: topTab === 'center' }" @click="topTab = 'center'">订单中心</button>
      <button type="button" :class="{ active: topTab === 'confirm' }" @click="topTab = 'confirm'">确认中心</button>
    </nav>

    <!-- 确认中心：本期占位，后续接入代理接单确认流程 -->
    <div v-if="topTab === 'confirm'" class="placeholder">
      <van-icon name="completed-o" size="44" />
      <p>确认中心即将开放</p>
    </div>

    <!-- 订单中心 -->
    <template v-else>
      <!-- 搜索条：维度下拉 + 关键字 -->
      <div class="search">
        <div class="search__select">
          <select v-model="searchBy" aria-label="搜索维度">
            <option v-for="o in searchByOptions" :key="o.value" :value="o.value">{{ o.label }}</option>
          </select>
          <van-icon name="arrow-down" />
        </div>
        <div class="search__input">
          <van-icon name="search" />
          <input
            v-model="keyword"
            type="search"
            placeholder="订单号/主充值号码"
            @input="onSearch"
            @keyup.enter="onSearch"
          />
        </div>
      </div>

      <!-- 状态子 tab（带角标计数） -->
      <div class="status-tabs">
        <button
          v-for="t in statusTabs"
          :key="t.key"
          type="button"
          class="status-tabs__item"
          :class="{ active: activeStatus === t.status }"
          @click="activeStatus = t.status"
        >
          {{ t.label }}
          <span v-if="badgeOf(t.status)" class="status-tabs__badge">{{ badgeOf(t.status) }}</span>
        </button>
      </div>

      <!-- 列表 -->
      <div v-if="loading" class="tip">
        <van-loading size="20" /> 加载中…
      </div>
      <ul v-else-if="list.length" class="list">
        <li v-for="item in list" :key="item.id" class="card" @click="goDetail(item)">
          <div class="card__top">
            <span class="card__name">{{ item.product_name || `商品 #${item.id}` }}</span>
            <span class="card__status" :class="`tone--${statusOf(item.status).tone}`">
              {{ statusOf(item.status).text }}
              <van-icon name="arrow" />
            </span>
          </div>

          <div class="card__no">
            <span class="card__no-text">订单编号：{{ maskNo(item.order_no) }}</span>
            <button type="button" class="card__copy" @click.stop="copyNo(item.order_no)">
              <van-icon name="orders-o" />
            </button>
          </div>

          <div class="card__amounts">
            <div class="amt"><span>订单金额</span><b>¥{{ fmtMoney(item.amount) }}</b></div>
            <div class="amt"><span>到账金额</span><b>¥{{ fmtMoney(item.received_amount) }}</b></div>
            <div class="amt"><span>剩余金额</span><b>¥{{ fmtMoney(item.remaining_amount) }}</b></div>
          </div>

          <div class="card__time">下单时间：{{ fmtTime(item.created_at) }}</div>
        </li>
      </ul>
      <p v-else class="tip">暂无订单</p>
    </template>
  </div>
</template>

<style scoped>
.orders { background: var(--bg-page); }

/* ---------- 顶部大 tab ---------- */
.top-tabs {
  display: flex; gap: 24px; margin-bottom: 18px;
  border-bottom: 1px solid var(--border-card);
}
.top-tabs button {
  position: relative; padding: 6px 0 12px;
  background: none; border: none;
  font-size: 16px; font-weight: 700; color: var(--text-muted);
}
.top-tabs button.active { color: var(--text-primary); }
.top-tabs button.active::after {
  content: ''; position: absolute; left: 0; right: 0; bottom: -1px;
  height: 2px; border-radius: 2px; background: var(--gradient-brand);
}

/* ---------- 搜索条 ---------- */
.search { display: flex; gap: 10px; margin-bottom: 16px; }
.search__select {
  position: relative; flex: 0 0 auto; display: flex; align-items: center;
}
.search__select select {
  appearance: none; -webkit-appearance: none;
  padding: 0 26px 0 12px; height: 38px;
  border: 1px solid var(--border-card); border-radius: 12px;
  background: var(--bg-card); color: var(--text-primary); font-size: 13px;
  color-scheme: dark;
}
.search__select select option { background: var(--c-surface-2); color: var(--text-primary); }
.search__select :deep(.van-icon) {
  position: absolute; right: 8px; pointer-events: none; color: var(--text-muted);
}
.search__input {
  flex: 1; display: flex; align-items: center; gap: 6px;
  padding: 0 12px; height: 38px;
  border: 1px solid var(--border-card); border-radius: 12px; background: var(--bg-card);
}
.search__input :deep(.van-icon) { color: var(--text-muted); }
.search__input input {
  flex: 1; border: none; outline: none; background: transparent;
  color: var(--text-primary); font-size: 13px;
}
.search__input input::placeholder { color: var(--text-muted); }

/* ---------- 状态子 tab ---------- */
.status-tabs {
  display: flex; gap: 8px 18px; margin-bottom: 16px;
  overflow-x: auto; scrollbar-width: none; white-space: nowrap;
}
.status-tabs::-webkit-scrollbar { display: none; }
.status-tabs__item {
  position: relative; flex: 0 0 auto; padding: 6px 2px;
  background: none; border: none;
  font-size: 13px; color: var(--text-secondary);
}
.status-tabs__item.active { color: var(--text-primary); font-weight: 700; }
.status-tabs__item.active::after {
  content: ''; position: absolute; left: 0; right: 0; bottom: -4px;
  height: 2px; border-radius: 2px; background: var(--gradient-brand);
}
.status-tabs__badge {
  margin-left: 4px; min-width: 16px; height: 16px; padding: 0 4px;
  display: inline-grid; place-items: center;
  border-radius: var(--radius-pill); font-size: 10px; font-weight: 700;
  color: #fff; background: var(--c-brand);
  vertical-align: middle;
}

/* ---------- 列表 ---------- */
.list { display: flex; flex-direction: column; gap: 12px; }
.card {
  padding: 14px;
  border: 1px solid var(--border-card); border-radius: var(--radius-card);
  background: var(--bg-card); transition: border-color .2s ease;
}
.card:active { border-color: var(--c-brand); }

.card__top { display: flex; align-items: center; justify-content: space-between; gap: 8px; }
.card__name {
  font-size: 15px; font-weight: 700; color: var(--text-primary);
  overflow: hidden; text-overflow: ellipsis; white-space: nowrap;
}
.card__status {
  flex: 0 0 auto; display: inline-flex; align-items: center; gap: 2px;
  font-size: 12px; font-weight: 700;
}
.tone--warn  { color: var(--c-accent); }
.tone--brand { color: var(--c-brand-2); }
.tone--ok    { color: var(--c-success); }
.tone--fail  { color: var(--c-danger); }
.tone--muted { color: var(--text-muted); }

.card__no {
  display: flex; align-items: center; justify-content: space-between;
  margin-top: 10px; font-size: 12px; color: var(--text-muted);
}
.card__no-text { overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.card__copy {
  flex: 0 0 auto; margin-left: 8px; padding: 4px;
  background: none; border: none; color: var(--text-secondary);
}

.card__amounts {
  display: flex; gap: 8px; margin-top: 12px;
  padding-top: 12px; border-top: 1px solid var(--border-card);
}
.amt {
  flex: 1; display: flex; flex-direction: column; gap: 4px;
  text-align: center;
}
.amt span { font-size: 11px; color: var(--text-muted); }
.amt b { font-size: 15px; font-weight: 800; color: var(--text-primary); }

.card__time {
  margin-top: 12px; padding-top: 10px; border-top: 1px solid var(--border-card);
  font-size: 11px; color: var(--text-muted);
}

/* ---------- 占位 / 空态 ---------- */
.tip, .placeholder {
  display: flex; align-items: center; justify-content: center; gap: 8px;
  padding: 48px 0; font-size: 12px; color: var(--text-muted);
}
.placeholder { flex-direction: column; gap: 12px; }
</style>
