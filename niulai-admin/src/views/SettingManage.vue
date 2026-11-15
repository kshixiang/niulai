<script setup>
import { ref, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { getSettings, updateSettings } from '@/api/setting'

const loading = ref(false)
const saving = ref(false)
const required = ref(false)

const rate = ref(6.69)
const rateSaving = ref(false)

async function loadSettings() {
  loading.value = true
  try {
    const data = await getSettings()
    required.value = !!data.register_code_required
    const n = Number(data.usdt_cny_rate)
    if (Number.isFinite(n) && n > 0) rate.value = n
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

async function saveRate() {
  if (rateSaving.value) return
  const n = Number(rate.value)
  if (!Number.isFinite(n) || n <= 0) {
    ElMessage.warning('请输入大于 0 的汇率')
    return
  }
  rateSaving.value = true
  try {
    const data = await updateSettings({ usdt_cny_rate: n.toFixed(2) })
    const saved = Number(data.usdt_cny_rate)
    if (Number.isFinite(saved) && saved > 0) rate.value = saved
    ElMessage.success('汇率已保存')
  } catch {
    // 拦截器已提示
  } finally {
    rateSaving.value = false
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

    <el-card v-loading="loading" shadow="never" class="card">
      <template #header><span class="card-title">汇率设置</span></template>

      <div class="row">
        <div class="info">
          <div class="title">USDT → CNY 汇率</div>
          <div class="desc">
            1 USDT 可兑换的 CNY(RMB) 数量。订单金额按 RMB 结算、充值余额按 USDT 结算，
            下单时按此汇率把应付金额换算为 USDT 后校验余额是否充足。保存后立即生效，无需重启服务。
          </div>
        </div>
        <div class="rate-edit">
          <el-input-number
            v-model="rate"
            :min="0.01"
            :max="1000"
            :precision="2"
            :step="0.01"
            :disabled="rateSaving"
          />
          <el-button type="primary" :loading="rateSaving" @click="saveRate">保存</el-button>
        </div>
      </div>

      <el-alert
        class="hint"
        type="info"
        :closable="false"
        show-icon
        title="示例：汇率 6.69 时，应付 ¥139.30 折合 20.82 USDT"
        description="H5 下单页会展示汇率与折合 USDT 金额，前后端共用同一份配置。"
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
.card { max-width: 760px; margin-bottom: 16px; }
.card:last-child { margin-bottom: 0; }
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
.rate-edit { display: flex; align-items: center; gap: 12px; flex: 0 0 auto; }
.hint { margin-top: 8px; }
</style>
