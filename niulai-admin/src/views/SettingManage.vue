<script setup>
import { ref, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { getSettings, updateSettings } from '@/api/setting'

const loading = ref(false)
const saving = ref(false)
const required = ref(false)

async function loadSettings() {
  loading.value = true
  try {
    const data = await getSettings()
    required.value = !!data.register_code_required
  } finally {
    loading.value = false
  }
}

async function onChange(value) {
  saving.value = true
  try {
    await updateSettings({ register_code_required: value })
    ElMessage.success(value ? '已开启：注册必须填写注册码' : '已关闭：注册码可不填')
  } catch {
    required.value = !value
  } finally {
    saving.value = false
  }
}

onMounted(loadSettings)
</script>

<template>
  <div class="page">
    <el-card v-loading="loading" shadow="never" class="card">
      <template #header><span class="card-title">注册设置</span></template>

      <div class="row">
        <div class="info">
          <div class="title">注册时强制填写注册码</div>
          <div class="desc">
            开启后，用户注册必须填写注册码；关闭后可留空。
            无论是否强制，只要填写了注册码，都会校验对应邀请人是否存在。
          </div>
        </div>
        <el-switch
          v-model="required"
          :loading="saving"
          active-text="强制"
          inactive-text="非强制"
          @change="onChange"
        />
      </div>

      <el-alert
        class="hint"
        type="info"
        :closable="false"
        show-icon
        title="注册码即代理邀请码（agents.invite_code）"
        description="设置保存后立即生效，无需重启服务。"
      />
    </el-card>
  </div>
</template>

<style scoped>
.page {
  background: #fff;
  border-radius: 10px;
  padding: 20px;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.04);
}
.card { max-width: 760px; }
.card-title { font-weight: 600; }
.row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 24px;
  padding: 8px 0 20px;
}
.info .title { font-size: 15px; color: #111827; margin-bottom: 6px; }
.info .desc { font-size: 13px; color: #6b7280; line-height: 1.6; }
.hint { margin-top: 8px; }
</style>
