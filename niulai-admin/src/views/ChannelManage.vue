<script setup>
import { ref, reactive, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { Plus, Edit } from '@element-plus/icons-vue'
import { fetchChannels, updateChannel } from '@/api/channel'
import ChannelForm from './components/ChannelForm.vue'

const loading = ref(false)
const rows = ref([])
const total = ref(0)
const query = reactive({ page: 1, pageSize: 10 })

const dialogVisible = ref(false)
const editingChannel = ref(null)

async function loadData() {
  loading.value = true
  try {
    // 拦截器已解包 data，后端返回 { list, total }
    const { list = [], total: t = 0 } =
      (await fetchChannels({ page: query.page, page_size: query.pageSize })) || {}
    rows.value = list.map((r) => ({ ...r, statusLoading: false }))
    total.value = t
  } finally {
    loading.value = false
  }
}

function openCreate() {
  editingChannel.value = null
  dialogVisible.value = true
}
function openEdit(row) {
  editingChannel.value = { ...row }
  dialogVisible.value = true
}
function onPageSizeChange() {
  query.page = 1
  loadData()
}

async function handleStatusChange(row) {
  // @change 时 v-model 已是新值，先记旧值用于失败回滚
  const origin = row.status === 1 ? 0 : 1
  row.statusLoading = true
  try {
    // 后端 Update 的 api_url / api_key 可省略，仅提交名称与状态
    await updateChannel(row.id, { name: row.name, status: row.status })
    ElMessage.success(row.status === 1 ? '已设为可用' : '已设为熔断')
  } catch {
    row.status = origin
  } finally {
    row.statusLoading = false
  }
}

const formatTime = (v) => (v ? String(v).replace('T', ' ').slice(0, 19) : '—')

onMounted(loadData)
</script>

<template>
  <div class="page">
    <div class="toolbar">
      <el-button type="primary" :icon="Plus" @click="openCreate">新增渠道</el-button>
      <span class="tip">熔断状态的渠道不能用于商品上架</span>
    </div>

    <el-table v-loading="loading" :data="rows" border stripe row-key="id" class="table">
      <el-table-column prop="id" label="ID" width="90" align="center" />
      <el-table-column prop="name" label="渠道名称" min-width="140" show-overflow-tooltip />
      <el-table-column label="接口地址" min-width="220" show-overflow-tooltip>
        <template #default="{ row }">{{ row.api_url || '—' }}</template>
      </el-table-column>
      <el-table-column label="健康状态" width="130" align="center">
        <template #default="{ row }">
          <el-switch
            v-model="row.status"
            :active-value="1"
            :inactive-value="0"
            :loading="row.statusLoading"
            inline-prompt
            active-text="可用"
            inactive-text="熔断"
            @change="handleStatusChange(row)"
          />
        </template>
      </el-table-column>
      <el-table-column label="创建时间" width="180" align="center">
        <template #default="{ row }">{{ formatTime(row.created_at) }}</template>
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
        @size-change="onPageSizeChange"
      />
    </div>

    <ChannelForm v-model="dialogVisible" :channel="editingChannel" @success="loadData" />
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
  gap: 16px;
  margin-bottom: 18px;
}
.tip {
  color: #9ca3af;
  font-size: 12px;
}
.pagination {
  display: flex;
  justify-content: flex-end;
  margin-top: 18px;
}
</style>
