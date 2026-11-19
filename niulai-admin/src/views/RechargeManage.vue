<script setup>
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { getRecharges, reviewRecharge } from '@/api/recharge'

const rows = ref([])
const total = ref(0)
const loading = ref(false)
const loadFailed = ref(false)
const query = reactive({ page: 1, pageSize: 10, status: '' })
const statuses = [{ value: 0, label: '待核实', type: 'warning' }, { value: 1, label: '已到账', type: 'success' }, { value: 2, label: '已驳回', type: 'danger' }]
const statusOf = (value) => statuses.find((s) => s.value === value) || { label: '未知', type: 'info' }
const formatTime = (value) => value ? new Date(value).toLocaleString('zh-CN', { hour12: false }) : '—'
const showReview = ref(false)
const selected = ref(null)
const saving = ref(false)
const form = reactive({ status: 1, amount: '', remark: '', verified: false })

async function load() {
  loading.value = true
  loadFailed.value = false
  try {
    const params = { page: query.page, page_size: query.pageSize }
    if (query.status !== '' && query.status != null) params.status = query.status
    const data = await getRecharges(params)
    rows.value = data.list || []
    total.value = data.total || 0
  } catch { loadFailed.value = true }
  finally { loading.value = false }
}
function search() { query.page = 1; load() }
function openReview(row, status) {
  selected.value = row
  Object.assign(form, { status, amount: row.reported_amount, remark: '', verified: false })
  showReview.value = true
}
async function submit() {
  if (saving.value || !selected.value) return
  if (form.status === 1) {
    if (!/^\d{1,8}(\.\d{1,6})?$/.test(form.amount.trim()) || Number(form.amount) < 1 || Number(form.amount) > 99999999.99) {
      ElMessage.warning('请输入1至99999999.99之间的实际到账金额，最多6位小数')
      return
    }
    if (!form.verified) { ElMessage.warning('请先核实链上到账信息及付款归属并勾选确认'); return }
    if (!form.remark.trim()) { ElMessage.warning('请填写付款归属核验说明，公开交易哈希不能作为归属证明'); return }
  } else if (!form.remark.trim()) { ElMessage.warning('请填写驳回原因'); return }
  saving.value = true
  try {
    if (form.status === 1) {
      const [whole, fraction = ''] = form.amount.trim().split('.')
      const amount = `${Number(whole)}.${fraction.padEnd(2, '0').slice(0, 2)}`
      await ElMessageBox.confirm(`确认向用户 #${selected.value.user_id} 入账 ${amount} USDT？确认后不可重复审核。`, '确认充值入账', { type: 'warning', confirmButtonText: '确认入账', cancelButtonText: '取消' })
    }
    await reviewRecharge(selected.value.id, { status: form.status, amount: form.status === 1 ? form.amount.trim() : '', remark: form.remark.trim() })
    showReview.value = false
    ElMessage.success(form.status === 1 ? '充值已入账，资金流水已生成' : '充值已驳回')
    await load()
  } catch { /* 取消和接口错误均保留表单，拦截器处理接口提示 */ }
  finally { saving.value = false }
}
onMounted(load)
</script>

<template>
  <div class="recharge-manage">
    <div class="toolbar"><h2>充值管理</h2><div class="filters"><el-select v-model="query.status" placeholder="全部状态" clearable :disabled="loading" style="width: 140px" @change="search"><el-option v-for="s in statuses" :key="s.value" :label="s.label" :value="s.value" /></el-select><el-button :loading="loading" @click="load">刷新</el-button></div></div>
    <el-alert type="warning" :closable="false" show-icon title="核实到账及付款归属后再确认入账" description="核对币种、网络、收款地址、交易成功及实际金额，并通过独立凭证核实付款属于该用户。交易哈希是公开信息，不能单独证明付款归属。确认后会增加余额，金额直接截断两位。" />
    <el-alert v-if="loadFailed" class="load-error" type="error" :closable="false" title="充值记录加载失败，请点击刷新重试" />
    <el-table v-loading="loading" :data="rows" stripe class="recharge-table" empty-text="暂无充值记录">
      <el-table-column prop="id" label="编号" width="80" />
      <el-table-column prop="user_id" label="用户ID" width="95" />
      <el-table-column prop="reported_amount" label="申报金额 USDT" min-width="150" />
      <el-table-column label="核实 / 待核实金额" min-width="155"><template #default="{ row }">{{ row.amount }} USDT</template></el-table-column>
      <el-table-column label="状态" width="105"><template #default="{ row }"><el-tag :type="statusOf(row.status).type">{{ statusOf(row.status).label }}</el-tag></template></el-table-column>
      <el-table-column prop="address" label="收款地址" min-width="220" show-overflow-tooltip />
      <el-table-column label="交易哈希" min-width="230"><template #default="{ row }"><a :href="`https://tronscan.org/#/transaction/${row.tx_hash}`" target="_blank" rel="noopener noreferrer" class="hash-link">{{ row.tx_hash }}</a></template></el-table-column>
      <el-table-column label="申报时间" width="185"><template #default="{ row }">{{ formatTime(row.created_at) }}</template></el-table-column>
      <el-table-column label="审核时间" width="185"><template #default="{ row }">{{ formatTime(row.reviewed_at) }}</template></el-table-column>
      <el-table-column prop="remark" label="审核备注" min-width="180" show-overflow-tooltip />
      <el-table-column label="操作" width="180" fixed="right"><template #default="{ row }"><template v-if="row.status === 0"><el-button type="primary" link @click="openReview(row, 1)">核实入账</el-button><el-button type="danger" link @click="openReview(row, 2)">驳回</el-button></template><span v-else class="muted">已审核</span></template></el-table-column>
    </el-table>
    <el-pagination v-model:current-page="query.page" v-model:page-size="query.pageSize" :total="total" :page-sizes="[10, 20, 50]" :disabled="loading" layout="total, sizes, prev, pager, next" class="pagination" @current-change="load" @size-change="search" />

    <el-dialog v-model="showReview" :title="form.status === 1 ? '核实充值入账' : '驳回充值申报'" width="560px" :close-on-click-modal="!saving" :close-on-press-escape="!saving" :show-close="!saving">
      <template v-if="selected">
        <el-descriptions :column="1" border><el-descriptions-item label="用户">#{{ selected.user_id }}</el-descriptions-item><el-descriptions-item label="收款地址"><span class="selectable">{{ selected.address }}</span></el-descriptions-item><el-descriptions-item label="交易哈希"><a :href="`https://tronscan.org/#/transaction/${selected.tx_hash}`" target="_blank" rel="noopener noreferrer" class="hash-link">{{ selected.tx_hash }}</a></el-descriptions-item><el-descriptions-item label="申报金额">{{ selected.reported_amount }} USDT</el-descriptions-item></el-descriptions>
        <el-form label-position="top" class="review-form" @submit.prevent="submit">
          <el-form-item v-if="form.status === 1" label="实际到账金额（USDT，直接截断两位）" required><el-input v-model="form.amount" placeholder="请输入链上核实的实际到账金额" maxlength="32" :disabled="saving" /></el-form-item>
          <el-form-item :label="form.status === 2 ? '驳回原因（用户可见）' : '付款归属核验说明（用户可见，请勿填写敏感资料）'" required><el-input v-model="form.remark" type="textarea" :rows="3" maxlength="255" show-word-limit :disabled="saving" /></el-form-item>
          <el-checkbox v-if="form.status === 1" v-model="form.verified" :disabled="saving">已核实网络、币种、地址、交易成功、实际金额及付款归属</el-checkbox>
        </el-form>
      </template>
      <template #footer><el-button :disabled="saving" @click="showReview = false">取消</el-button><el-button :type="form.status === 1 ? 'primary' : 'danger'" :loading="saving" @click="submit">{{ form.status === 1 ? '确认入账' : '确认驳回' }}</el-button></template>
    </el-dialog>
  </div>
</template>

<style scoped>
.recharge-manage { padding: 20px; border-radius: 10px; background: #fff; }
.toolbar { display: flex; align-items: center; justify-content: space-between; margin-bottom: 18px; gap: 20px; }.toolbar h2 { margin: 0; font-size: 18px; }.filters { display: flex; gap: 12px; }
.recharge-table { margin-top: 20px; }.hash-link { color: #409eff; overflow-wrap: anywhere; user-select: all; text-decoration: none; font-family: monospace; }.hash-link:hover { text-decoration: underline; }.selectable { overflow-wrap: anywhere; user-select: all; }
.pagination { margin-top: 20px; justify-content: flex-end; }.review-form { margin-top: 20px; }.muted { color: #909399; }.load-error { margin-top: 12px; }
</style>
