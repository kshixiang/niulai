<script setup>
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { showToast } from 'vant'
import { getRecharges } from '@/api'
import { copyText } from '@/utils/clipboard'

const router = useRouter()
const rows = ref([])
const total = ref(0)
const page = ref(1)
const loading = ref(false)
const failed = ref(false)
const status = ref('')
const retryReset = ref(true)
const tabs = [{ value: '', label: '全部' }, { value: 0, label: '待核实' }, { value: 1, label: '已到账' }, { value: 2, label: '已驳回' }]
const statusText = (value) => tabs.find((t) => t.value === value)?.label || '未知状态'
const formatTime = (value) => value ? new Date(value).toLocaleString('zh-CN', { hour12: false }) : '—'

async function load(reset = false) {
  if (loading.value) return
  loading.value = true
  failed.value = false
  retryReset.value = reset
  const nextPage = reset ? 1 : page.value + 1
  try {
    const data = await getRecharges({ page: nextPage, page_size: 10, status: status.value === '' ? undefined : status.value })
    rows.value = reset ? (data.list || []) : [...rows.value, ...(data.list || [])]
    total.value = data.total || 0
    page.value = nextPage
  } catch { failed.value = true }
  finally { loading.value = false }
}

function changeStatus(value) {
  if (loading.value || status.value === value) return
  status.value = value
  rows.value = []
  total.value = 0
  load(true)
}
function goBack() {
  if (window.history.state?.back) router.back()
  else router.replace('/recharge')
}
async function copyHash(hash) {
  try { await copyText(hash); showToast('交易哈希已复制') }
  catch { showToast('复制失败，请长按手动复制') }
}
onMounted(() => load(true))
</script>

<template>
  <div class="records">
    <van-nav-bar title="充值记录" left-arrow :border="false" @click-left="goBack"><template #right><button class="records__refresh" type="button" :disabled="loading" @click="load(true)">刷新</button></template></van-nav-bar>
    <nav class="records__tabs" aria-label="充值状态"><button v-for="tab in tabs" :key="tab.value" type="button" :class="{ active: status === tab.value }" :disabled="loading" :aria-pressed="status === tab.value" @click="changeStatus(tab.value)">{{ tab.label }}</button></nav>
    <main class="records__body" :aria-busy="loading">
      <article v-for="row in rows" :key="row.id" class="records__card">
        <div class="records__top"><strong>{{ row.amount }} <small>USDT</small></strong><span class="records__badge" :class="`status-${row.status}`">{{ statusText(row.status) }}</span></div>
        <p class="records__time">{{ formatTime(row.created_at) }}</p>
        <dl><div><dt>充值编号</dt><dd>#{{ row.id }}</dd></div><div><dt>充值网络</dt><dd>USDT / TRC-20</dd></div><div><dt>申报金额</dt><dd>{{ row.reported_amount }} USDT</dd></div></dl>
        <div class="records__hash"><span>交易哈希</span><button type="button" aria-label="复制交易哈希" @click="copyHash(row.tx_hash)"><van-icon name="description-o" /> 复制</button><p>{{ row.tx_hash }}</p></div>
        <p v-if="row.status === 0" class="records__note">已提交转账信息，正在等待核实到账。</p>
        <p v-if="row.remark" class="records__note" :class="{ 'records__note--error': row.status === 2 }">{{ row.status === 2 ? '驳回原因：' : '备注：' }}{{ row.remark }}</p>
        <p v-if="row.reviewed_at" class="records__time">审核时间：{{ formatTime(row.reviewed_at) }}</p>
      </article>
      <div v-if="loading" class="records__empty" role="status"><van-loading color="var(--c-brand)" size="24" />加载中…</div>
      <div v-else-if="failed" class="records__empty" role="alert"><p>记录加载失败，请重试</p><button type="button" @click="load(retryReset)">重新加载</button></div>
      <div v-else-if="!rows.length" class="records__empty"><van-icon name="records-o" size="46" /><p>暂无{{ status === '' ? '充值' : statusText(status) }}记录</p><button type="button" @click="router.push('/recharge')">去充值</button></div>
      <button v-else-if="rows.length < total" type="button" class="records__more" @click="load()">加载更多</button>
      <p v-else class="records__end">已展示全部 {{ total }} 条记录</p>
    </main>
  </div>
</template>

<style scoped>
.records { display: flex; flex-direction: column; height: 100%; background: var(--bg-page); padding-top: var(--safe-top); }
.records :deep(.van-nav-bar__title), .records :deep(.van-nav-bar__arrow) { color: var(--text-primary); }
.records__refresh { color: var(--c-brand); font-size: 14px; background: none; border: 0; padding: 8px; }
.records__tabs { display: grid; grid-template-columns: repeat(4, 1fr); padding: 12px 16px; gap: 8px; }
.records__tabs button { padding: 9px 0; background: var(--bg-card); border: 1px solid var(--border-card); color: var(--text-secondary); border-radius: 22px; font-size: 13px; }
.records__tabs button.active { color: var(--text-primary); background: var(--c-brand); border-color: var(--c-brand); }
.records__body { flex: 1; min-height: 0; overflow-y: auto; padding: 4px 16px calc(24px + var(--safe-bottom)); }
.records__card { padding: 18px; margin-bottom: 14px; border-radius: 16px; background: var(--bg-card); border: 1px solid var(--border-card); }
.records__top { display: flex; justify-content: space-between; align-items: center; gap: 12px; }
.records__top strong { font-size: 23px; font-variant-numeric: tabular-nums; overflow-wrap: anywhere; }
.records__top small { font-size: 12px; color: var(--text-secondary); }
.records__badge { padding: 4px 9px; border-radius: 6px; white-space: nowrap; font-size: 12px; background: var(--c-fill); }
.status-0 { color: var(--c-accent); }.status-1 { color: var(--c-success); }.status-2 { color: var(--c-danger); }
.records__time { margin-top: 8px; font-size: 12px; color: var(--text-muted); }
.records__card dl { margin: 16px 0 0; font-size: 12px; }
.records__card dl div { display: flex; justify-content: space-between; gap: 12px; margin: 8px 0; }
.records__card dt { color: var(--text-secondary); }.records__card dd { margin: 0; overflow-wrap: anywhere; }
.records__hash { display: flex; flex-wrap: wrap; justify-content: space-between; margin-top: 12px; padding-top: 12px; border-top: 1px solid var(--border-card); color: var(--text-secondary); font-size: 12px; }
.records__hash p { width: 100%; margin-top: 8px; overflow-wrap: anywhere; user-select: all; font-family: monospace; line-height: 1.6; }
.records__hash button { border: 0; padding: 0; background: none; color: var(--c-brand); }
.records__note { margin-top: 12px; padding: 10px; background: var(--c-fill); border-radius: 8px; color: var(--text-secondary); font-size: 12px; overflow-wrap: anywhere; }.records__note--error { color: var(--c-danger); }
.records__empty { display: flex; align-items: center; justify-content: center; flex-direction: column; gap: 14px; padding: 70px 0; color: var(--text-muted); }
.records__empty button, .records__more { padding: 9px 22px; border: 1px solid var(--c-brand); border-radius: 22px; background: none; color: var(--c-brand); }
.records__more { display: block; margin: 20px auto; }.records__end { padding: 20px 0; text-align: center; color: var(--text-muted); font-size: 12px; }
</style>
