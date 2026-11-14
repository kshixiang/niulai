<script setup>
import { ref, reactive, onMounted, h } from 'vue'
import { ElMessageBox } from 'element-plus'
import { Plus } from '@element-plus/icons-vue'
import {
  getRegisterCodes,
  generateRegisterCodes,
  updateRegisterCodeStatus,
} from '@/api/registerCode'

const loading = ref(false)
const rows = ref([])
const total = ref(0)
const query = reactive({ page: 1, pageSize: 10 })

const dialogVisible = ref(false)
const generating = ref(false)
const generateForm = reactive({ count: 5, inviterId: null })

async function loadData() {
  loading.value = true
  try {
    const { list = [], total: t = 0 } =
      (await getRegisterCodes({ page: query.page, page_size: query.pageSize })) || {}
    rows.value = list.map((r) => ({ ...r, statusLoading: false }))
    total.value = t
  } finally {
    loading.value = false
  }
}

function openGenerate() {
  generateForm.count = 5
  generateForm.inviterId = null
  dialogVisible.value = true
}

async function submitGenerate() {
  generating.value = true
  try {
    const payload = { count: Number(generateForm.count) }
    if (generateForm.inviterId) payload.inviter_id = Number(generateForm.inviterId)
    const list = await generateRegisterCodes(payload)
    dialogVisible.value = false
    ElMessageBox.alert(
      h(
        'pre',
        { style: 'margin:0; white-space:pre-wrap; font-size:13px; line-height:1.9' },
        list.map((c) => c.code).join('\n'),
      ),
      `已生成 ${list.length} 个注册码`,
      { confirmButtonText: '复制好了' },
    ).catch(() => {})
    query.page = 1
    loadData()
  } finally {
    generating.value = false
  }
}

async function handleStatusChange(row) {
  const origin = row.status === 1 ? 0 : 1
  row.statusLoading = true
  try {
    await updateRegisterCodeStatus(row.id, row.status)
  } catch {
    row.status = origin
  } finally {
    row.statusLoading = false
  }
}

function onPageSizeChange() {
  query.page = 1
  loadData()
}

const formatTime = (v) => (v ? String(v).replace('T', ' ').slice(0, 19) : '—')

onMounted(loadData)
</script>

<template>
  <div class="page">
    <div class="toolbar">
      <el-button type="primary" :icon="Plus" @click="openGenerate">生成注册码</el-button>
      <span class="tip">注册码由后台生成；绑定邀请人后，注册时还会校验该用户存在</span>
    </div>

    <el-table v-loading="loading" :data="rows" border stripe row-key="id" class="table">
      <el-table-column prop="id" label="ID" width="90" align="center" />
      <el-table-column label="注册码" min-width="220">
        <template #default="{ row }"><code class="code">{{ row.code }}</code></template>
      </el-table-column>
      <el-table-column label="邀请人ID" width="120" align="center">
        <template #default="{ row }">{{ row.inviter_id || '—' }}</template>
      </el-table-column>
      <el-table-column label="状态" width="120" align="center">
        <template #default="{ row }">
          <el-switch
            v-model="row.status"
            :active-value="1"
            :inactive-value="0"
            :loading="row.statusLoading"
            inline-prompt
            active-text="可用"
            inactive-text="禁用"
            @change="handleStatusChange(row)"
          />
        </template>
      </el-table-column>
      <el-table-column label="创建时间" width="180" align="center">
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
        @size-change="onPageSizeChange"
      />
    </div>

    <el-dialog v-model="dialogVisible" title="生成注册码" width="440px" :close-on-click-modal="false">
      <el-form label-width="110px">
        <el-form-item label="生成数量">
          <el-input-number v-model="generateForm.count" :min="1" :max="100" />
          <span class="hint">一次最多 100 个</span>
        </el-form-item>
        <el-form-item label="邀请人ID">
          <el-input-number
            v-model="generateForm.inviterId"
            :min="1"
            :controls="false"
            placeholder="选填"
            style="width: 100%"
          />
          <span class="hint">选填。填写后，注册时会校验该用户存在且状态正常</span>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="generating" @click="submitGenerate">生成</el-button>
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
  gap: 16px;
  margin-bottom: 18px;
}
.tip,
.hint {
  color: #9ca3af;
  font-size: 12px;
}
.hint {
  display: block;
  margin-top: 4px;
  line-height: 1.5;
}
.code {
  font-family: ui-monospace, Consolas, monospace;
  color: #1f2937;
  letter-spacing: 0.5px;
}
.pagination {
  display: flex;
  justify-content: flex-end;
  margin-top: 18px;
}
</style>
