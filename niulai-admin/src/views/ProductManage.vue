<script setup>
import { ref, reactive, computed, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { Plus, Search, Refresh, Edit } from '@element-plus/icons-vue'
import { getProducts, updateProductStatus } from '@/api/product'
import { getChannelOptions } from '@/api/channel'
import ProductForm from './components/ProductForm.vue'

const loading = ref(false)
const rows = ref([])
const total = ref(0)
const channels = ref([])
const query = reactive({ name: '', page: 1, pageSize: 10 })

const dialogVisible = ref(false)
const editingProduct = ref(null)

// 渠道 id → 名称，供表格回显（与 model/channels 的 id/name 对应）
const channelMap = computed(() =>
  Object.fromEntries(channels.value.map((c) => [String(c.id), c.name])),
)

async function loadChannels() {
  channels.value = (await getChannelOptions()) || []
}

async function loadData() {
  loading.value = true
  try {
    // 拦截器已解包 data，后端返回 { list, total }；参数名对齐后端 form tag
    const { list = [], total: t = 0 } =
      (await getProducts({ name: query.name, page: query.page, page_size: query.pageSize })) || {}
    rows.value = list.map((r) => ({ ...r, statusLoading: false }))
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
  query.name = ''
  query.page = 1
  loadData()
}
function openCreate() {
  editingProduct.value = null
  dialogVisible.value = true
}
function openEdit(row) {
  editingProduct.value = { ...row }
  dialogVisible.value = true
}

async function handleStatusChange(row) {
  // @change 时 v-model 已是新值，先记旧值用于失败回滚
  const origin = row.status === 1 ? 0 : 1
  row.statusLoading = true
  try {
    await updateProductStatus(row.id, row.status)
    ElMessage.success(row.status === 1 ? '已上架' : '已下架')
  } catch {
    row.status = origin
  } finally {
    row.statusLoading = false
  }
}

const formatPrice = (v) => (Number.isFinite(Number(v)) ? `¥${Number(v).toFixed(2)}` : '—')
const formatDiscount = (v) => (v == null || v === '' ? '—' : String(Number(v)))
const formatFace = (v) => {
  const n = Number(v)
  return Number.isFinite(n) ? String(n) : '—'
}

onMounted(() => {
  loadChannels()
  loadData()
})
</script>

<template>
  <div class="page">
    <!-- 顶部操作区 -->
    <div class="toolbar">
      <el-button type="primary" :icon="Plus" @click="openCreate">新增商品</el-button>
      <div class="toolbar-right">
        <el-input
          v-model="query.name"
          placeholder="按商品名搜索"
          clearable
          style="width: 240px"
          :prefix-icon="Search"
          @keyup.enter="onSearch"
        />
        <el-button :icon="Search" @click="onSearch">搜索</el-button>
        <el-button :icon="Refresh" @click="onReset">重置</el-button>
      </div>
    </div>

    <!-- 数据表格 -->
    <el-table v-loading="loading" :data="rows" border stripe row-key="id" class="table">
      <el-table-column prop="id" label="ID" width="90" align="center" />
      <el-table-column label="封面" width="88" align="center">
        <template #default="{ row }">
          <el-image
            v-if="row.cover_image"
            class="cover"
            :src="row.cover_image"
            fit="cover"
            :preview-src-list="[row.cover_image]"
            preview-teleported
          />
          <span v-else class="muted">无</span>
        </template>
      </el-table-column>
      <el-table-column prop="name" label="名称" min-width="140" show-overflow-tooltip />
      <el-table-column label="成本价" width="110" align="right">
        <template #default="{ row }">{{ formatPrice(row.cost_price) }}</template>
      </el-table-column>
      <el-table-column label="折扣" width="90" align="center">
        <template #default="{ row }">
          <el-tag v-if="row.discount != null" type="danger" effect="light" size="small">
            {{ formatDiscount(row.discount) }}折
          </el-tag>
          <span v-else class="muted">—</span>
        </template>
      </el-table-column>
      <el-table-column label="价格类别" min-width="180">
        <template #default="{ row }">
          <template v-if="row.denominations && row.denominations.length">
            <el-tag
              v-for="d in row.denominations"
              :key="d.id"
              size="small"
              effect="plain"
              class="denom-tag"
            >
              {{ d.name ? `${d.name} ${formatFace(d.face_value)}` : formatFace(d.face_value) }}
            </el-tag>
          </template>
          <span v-else class="muted">—</span>
        </template>
      </el-table-column>
      <el-table-column label="绑定渠道" min-width="120" show-overflow-tooltip>
        <template #default="{ row }">
          {{ channelMap[String(row.channel_id)] || row.channel_name || '—' }}
        </template>
      </el-table-column>
      <el-table-column label="状态" width="120" align="center">
        <template #default="{ row }">
          <el-switch
            v-model="row.status"
            :active-value="1"
            :inactive-value="0"
            :loading="row.statusLoading"
            inline-prompt
            active-text="上架"
            inactive-text="下架"
            @change="handleStatusChange(row)"
          />
        </template>
      </el-table-column>
      <el-table-column label="操作" width="100" align="center" fixed="right">
        <template #default="{ row }">
          <el-button type="primary" link :icon="Edit" @click="openEdit(row)">编辑</el-button>
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

    <ProductForm
      v-model="dialogVisible"
      :product="editingProduct"
      :channels="channels"
      @success="loadData"
    />
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
.toolbar-right {
  display: flex;
  align-items: center;
  gap: 8px;
}
.cover {
  width: 44px;
  height: 44px;
  border-radius: 6px;
}
.muted {
  color: #c0c4cc;
}
.denom-tag {
  margin: 2px 4px 2px 0;
}
.pagination {
  display: flex;
  justify-content: flex-end;
  margin-top: 18px;
}
</style>
