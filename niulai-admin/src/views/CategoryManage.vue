<script setup>
import { ref, reactive, onMounted, nextTick } from 'vue'
import { ElMessage } from 'element-plus'
import { Plus, Edit, Refresh } from '@element-plus/icons-vue'
import { fetchCategories, createCategory, updateCategory } from '@/api/category'

const rows = ref([])
const loading = ref(false)
const failed = ref(false)
const dialogVisible = ref(false)
const editingId = ref(null)
const form = reactive({ name: '' })
const formRef = ref(null)
const submitting = ref(false)
const rules = {
  name: [{
    validator: (_rule, value, callback) => {
      const name = String(value || '').trim()
      if (!name || Array.from(name).length > 32) return callback(new Error('请输入1至32个字符的分类名称'))
      callback()
    },
    trigger: 'blur',
  }],
}

async function loadData() {
  loading.value = true
  failed.value = false
  try {
    const list = await fetchCategories()
    rows.value = Array.isArray(list) ? list : []
  } catch {
    failed.value = true
  } finally {
    loading.value = false
  }
}

function openForm(row = null) {
  editingId.value = row?.id ?? null
  form.name = row?.name ?? ''
  dialogVisible.value = true
  nextTick(() => formRef.value?.clearValidate())
}

async function submit() {
  if (submitting.value || !(await formRef.value.validate().catch(() => false))) return
  submitting.value = true
  try {
    const name = form.name.trim()
    if (editingId.value) await updateCategory(editingId.value, name)
    else await createCategory(name)
    ElMessage.success(editingId.value ? '分类已更新' : '分类已新增')
    dialogVisible.value = false
    await loadData()
  } catch {
    // 业务错误由统一请求拦截器展示，保留弹窗和输入方便修改。
  } finally {
    submitting.value = false
  }
}

const formatTime = (v) => v ? String(v).replace('T', ' ').slice(0, 19) : '—'
onMounted(loadData)
</script>

<template>
  <div class="page">
    <div class="toolbar">
      <el-button type="primary" :icon="Plus" @click="openForm()">新增分类</el-button>
      <el-button :icon="Refresh" :loading="loading" @click="loadData">刷新</el-button>
      <span class="tip">商品与渠道共用分类，修改名称后关联内容同步显示。</span>
    </div>
    <el-alert v-if="failed" title="分类加载失败，请点击刷新重试" type="error" :closable="false" show-icon class="error" />
    <el-table v-loading="loading" :data="rows" border stripe row-key="id" empty-text="暂无分类，点击新增分类开始添加">
      <el-table-column prop="id" label="ID" width="90" align="center" />
      <el-table-column prop="name" label="分类名称" min-width="180" show-overflow-tooltip />
      <el-table-column label="创建时间" width="180" align="center">
        <template #default="{ row }">{{ formatTime(row.created_at) }}</template>
      </el-table-column>
      <el-table-column label="操作" width="110" align="center">
        <template #default="{ row }"><el-button type="primary" link :icon="Edit" @click="openForm(row)">编辑</el-button></template>
      </el-table-column>
    </el-table>
    <div class="total">共 {{ rows.length }} 个分类</div>

    <el-dialog v-model="dialogVisible" :title="editingId ? '编辑分类' : '新增分类'" width="440px" :close-on-click-modal="false" :close-on-press-escape="!submitting" :show-close="!submitting" destroy-on-close>
      <el-form ref="formRef" :model="form" :rules="rules" label-width="90px" @submit.prevent="submit">
        <el-form-item label="分类名称" prop="name">
          <el-input v-model="form.name" placeholder="如：会员、油卡" :disabled="submitting" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button :disabled="submitting" @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="submitting" @click="submit">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<style scoped>
.page { background: #fff; border-radius: 10px; padding: 20px; box-shadow: 0 1px 3px rgba(0,0,0,.04); }
.toolbar { display: flex; flex-wrap: wrap; align-items: center; gap: 12px; margin-bottom: 18px; }
.toolbar :deep(.el-button + .el-button) { margin-left: 0; }
.tip, .total { color: #6b7280; font-size: 13px; }
.total { margin-top: 18px; text-align: right; }
.error { margin-bottom: 16px; }
</style>
