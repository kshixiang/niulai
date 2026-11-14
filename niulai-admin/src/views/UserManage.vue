<script setup>
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Search, Refresh } from '@element-plus/icons-vue'
import { getUsers, updateUserStatus } from '@/api/user'

const loading = ref(false)
const rows = ref([])
const total = ref(0)
const query = reactive({ keyword: '', status: '', page: 1, pageSize: 10 })

const statusOptions = [
  { label: '正常', value: 1 },
  { label: '封禁', value: 0 },
]

async function loadData() {
  loading.value = true
  try {
    const params = { page: query.page, page_size: query.pageSize }
    if (query.keyword) params.keyword = query.keyword
    // el-select clearable 清空后值为 null/undefined，需一并排除
    if (query.status !== '' && query.status != null) params.status = query.status
    // 拦截器已解包 data，后端返回 { list, total }
    const { list = [], total: t = 0 } = (await getUsers(params)) || {}
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
  query.keyword = ''
  query.status = ''
  query.page = 1
  loadData()
}

async function handleStatusChange(row) {
  // @change 时 v-model 已是新值，先记旧值用于取消/失败时回滚
  const origin = row.status === 1 ? 0 : 1
  // 封禁影响用户登录，先确认再提交
  if (row.status === 0) {
    try {
      await ElMessageBox.confirm(
        `确认封禁账号「${row.phone}」？封禁后该用户将无法登录。`,
        '封禁确认',
        { type: 'warning', confirmButtonText: '确认封禁', cancelButtonText: '取消' },
      )
    } catch {
      row.status = origin
      return
    }
  }
  row.statusLoading = true
  try {
    await updateUserStatus(row.id, row.status)
    ElMessage.success(row.status === 1 ? '已恢复正常' : '已封禁')
    // 状态过滤生效时，切换后该行可能不再命中过滤条件，刷新保持一致
    if (query.status !== '' && query.status != null) loadData()
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
      <div class="toolbar-title">已注册用户</div>
      <div class="toolbar-right">
        <el-input
          v-model="query.keyword"
          placeholder="账号 / 昵称"
          clearable
          style="width: 220px"
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
      <el-table-column prop="id" label="ID" width="90" align="center" />
      <el-table-column prop="phone" label="登录账号" min-width="160" show-overflow-tooltip />
      <el-table-column prop="nickname" label="昵称" min-width="140" show-overflow-tooltip />
      <el-table-column label="状态" width="130" align="center">
        <template #default="{ row }">
          <el-switch
            v-model="row.status"
            :active-value="1"
            :inactive-value="0"
            :loading="row.statusLoading"
            inline-prompt
            active-text="正常"
            inactive-text="封禁"
            @change="handleStatusChange(row)"
          />
        </template>
      </el-table-column>
      <el-table-column label="注册时间" width="180" align="center">
        <template #default="{ row }">{{ formatTime(row.created_at) }}</template>
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
</style>
