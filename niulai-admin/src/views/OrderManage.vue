<script setup>
import { ref, reactive, onMounted } from 'vue'
import { Search, Refresh, View } from '@element-plus/icons-vue'
import { getOrders, getOrderDetail } from '@/api/order'

const loading = ref(false)
const rows = ref([])
const total = ref(0)
const query = reactive({ keyword: '', status: '', page: 1, pageSize: 10 })

/* 订单状态：与后端 model.OrderStatus* 保持一致 */
const statusOptions = [
  { label: '待支付', value: 0 },
  { label: '已支付', value: 1 },
  { label: '充值中', value: 2 },
  { label: '成功', value: 3 },
  { label: '失败', value: 4 },
  { label: '已退款', value: 5 },
]
const statusMap = {
  0: { text: '待支付', type: 'warning' },
  1: { text: '已支付', type: 'primary' },
  2: { text: '充值中', type: 'primary' },
  3: { text: '成功', type: 'success' },
  4: { text: '失败', type: 'danger' },
  5: { text: '已退款', type: 'info' },
}
const statusOf = (s) => statusMap[s] || { text: '未知', type: 'info' }

const detailVisible = ref(false)
const detailLoading = ref(false)
const detail = ref(null)

async function loadData() {
  loading.value = true
  try {
    const params = { page: query.page, page_size: query.pageSize }
    if (query.keyword) params.keyword = query.keyword
    // el-select clearable 清空后值为 null/undefined，需一并排除
    if (query.status !== '' && query.status != null) params.status = query.status
    // 拦截器已解包 data，后端返回 { list, total }
    const { list = [], total: t = 0 } = (await getOrders(params)) || {}
    rows.value = list
    total.value = t
  } finally {
    loading.value = false
  }
}

function onSearch() {
  query.page = 1
  loadData()
}
function onReset() {
  query.keyword = ''
  query.status = ''
  query.page = 1
  loadData()
}

async function openDetail(row) {
  detail.value = null
  detailVisible.value = true
  detailLoading.value = true
  try {
    detail.value = await getOrderDetail(row.id)
  } catch {
    // 错误提示由 request.js 拦截器统一弹出
    detailVisible.value = false
  } finally {
    detailLoading.value = false
  }
}

const fmtMoney = (v) => `¥${Number(v || 0).toFixed(2)}`
const fmtTime = (v) => (v ? String(v).replace('T', ' ').slice(0, 19) : '—')
const userName = (o) =>
  o.user_phone ? `${o.user_phone}${o.user_nickname ? `（${o.user_nickname}）` : ''}` : `用户 #${o.user_id}`

onMounted(loadData)
</script>

<template>
  <div class="page">
    <div class="toolbar">
      <div class="toolbar-title">订单列表</div>
      <div class="toolbar-right">
        <el-input
          v-model="query.keyword"
          placeholder="订单号 / 充值账号 / 用户账号 / 商品名"
          clearable
          style="width: 280px"
          :prefix-icon="Search"
          @keyup.enter="onSearch"
          @clear="onSearch"
        />
        <el-select
          v-model="query.status"
          placeholder="全部状态"
          clearable
          style="width: 130px"
          @change="onSearch"
        >
          <el-option v-for="o in statusOptions" :key="o.value" :label="o.label" :value="o.value" />
        </el-select>
        <el-button :icon="Search" @click="onSearch">搜索</el-button>
        <el-button :icon="Refresh" @click="onReset">重置</el-button>
      </div>
    </div>

    <el-table v-loading="loading" :data="rows" border stripe row-key="id" class="table">
      <el-table-column prop="id" label="ID" width="80" align="center" />
      <el-table-column prop="order_no" label="订单号" min-width="180" show-overflow-tooltip />
      <el-table-column label="商品" min-width="140" show-overflow-tooltip>
        <template #default="{ row }">{{ row.product_name || `商品 #${row.product_id}` }}</template>
      </el-table-column>
      <el-table-column label="用户" min-width="160" show-overflow-tooltip>
        <template #default="{ row }">{{ userName(row) }}</template>
      </el-table-column>
      <el-table-column prop="account" label="充值账号" min-width="140" show-overflow-tooltip />
      <el-table-column label="实付" width="100" align="right">
        <template #default="{ row }">{{ fmtMoney(row.amount) }}</template>
      </el-table-column>
      <el-table-column label="已到账" width="100" align="right">
        <template #default="{ row }">{{ fmtMoney(row.received_amount) }}</template>
      </el-table-column>
      <el-table-column label="渠道成本" width="100" align="right">
        <template #default="{ row }">{{ fmtMoney(row.cost_amount) }}</template>
      </el-table-column>
      <el-table-column label="利润" width="100" align="right">
        <template #default="{ row }"><span class="profit">{{ fmtMoney(row.profit) }}</span></template>
      </el-table-column>
      <el-table-column label="状态" width="100" align="center">
        <template #default="{ row }">
          <el-tag :type="statusOf(row.status).type" effect="light" size="small">
            {{ statusOf(row.status).text }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column label="下单时间" width="170" align="center">
        <template #default="{ row }">{{ fmtTime(row.created_at) }}</template>
      </el-table-column>
      <el-table-column label="操作" width="90" align="center" fixed="right">
        <template #default="{ row }">
          <el-button type="primary" link :icon="View" @click="openDetail(row)">详情</el-button>
        </template>
      </el-table-column>
    </el-table>

    <div class="pagination">
      <el-pagination
        background
        layout="total, sizes, prev, pager, next, jumper"
        :total="total"
        :page-sizes="[10, 20, 50]"
        v-model:current-page="query.page"
        v-model:page-size="query.pageSize"
        @current-change="loadData"
        @size-change="onSearch"
      />
    </div>

    <el-dialog v-model="detailVisible" title="订单详情" width="680px" destroy-on-close>
      <div v-loading="detailLoading" class="detail">
        <template v-if="detail">
          <el-descriptions :column="2" border>
            <el-descriptions-item label="订单号">{{ detail.order_no || '—' }}</el-descriptions-item>
            <el-descriptions-item label="状态">
              <el-tag :type="statusOf(detail.status).type" effect="light" size="small">
                {{ statusOf(detail.status).text }}
              </el-tag>
            </el-descriptions-item>
            <el-descriptions-item label="商品">
              {{ detail.product_name || `商品 #${detail.product_id}` }}
            </el-descriptions-item>
            <el-descriptions-item label="渠道">
              {{ detail.channel_name || `渠道 #${detail.channel_id}` }}
            </el-descriptions-item>
            <el-descriptions-item label="用户">{{ userName(detail) }}</el-descriptions-item>
            <el-descriptions-item label="充值账号">{{ detail.account || '—' }}</el-descriptions-item>
            <el-descriptions-item label="实付金额">{{ fmtMoney(detail.amount) }}</el-descriptions-item>
            <el-descriptions-item label="已到账">{{ fmtMoney(detail.received_amount) }}</el-descriptions-item>
            <el-descriptions-item label="渠道成本">{{ fmtMoney(detail.cost_amount) }}</el-descriptions-item>
            <el-descriptions-item label="利润">
              <span class="profit">{{ fmtMoney(detail.profit) }}</span>
            </el-descriptions-item>
            <el-descriptions-item label="下单时间">{{ fmtTime(detail.created_at) }}</el-descriptions-item>
            <el-descriptions-item label="回调时间">{{ fmtTime(detail.callback_at) }}</el-descriptions-item>
          </el-descriptions>

          <div v-if="detail.fields && detail.fields.length" class="fields">
            <h4 class="fields__title">下单填写项</h4>
            <div v-for="(f, i) in detail.fields" :key="i" class="fields__row">
              <span class="fields__label">{{ f.label }}</span>
              <span class="fields__value">{{ f.value }}</span>
            </div>
          </div>
        </template>
      </div>
      <template #footer>
        <el-button @click="detailVisible = false">关闭</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<style scoped>
.page {
  background: #fff;
  border-radius: 10px;
  padding: 20px;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.04);
}
.toolbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  margin-bottom: 18px;
}
.toolbar-title {
  font-size: 16px;
  font-weight: 600;
  color: #1f2937;
}
.toolbar-right {
  display: flex;
  align-items: center;
  gap: 8px;
}
.pagination {
  display: flex;
  justify-content: flex-end;
  margin-top: 18px;
}
.profit {
  color: #e6a23c;
  font-weight: 700;
}
.detail {
  min-height: 120px;
}
.fields {
  margin-top: 18px;
}
.fields__title {
  margin: 0 0 10px;
  font-size: 14px;
  font-weight: 600;
  color: #1f2937;
}
.fields__row {
  display: flex;
  gap: 16px;
  padding: 8px 0;
  font-size: 13px;
  border-top: 1px solid #ebeef5;
}
.fields__label {
  flex: 0 0 140px;
  color: #6b7280;
}
.fields__value {
  color: #1f2937;
  word-break: break-all;
}
</style>
