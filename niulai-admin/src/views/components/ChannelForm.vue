<script setup>
import { ref, reactive, computed, watch, nextTick } from 'vue'
import { ElMessage } from 'element-plus'
import { createChannel, updateChannel } from '@/api/channel'

const props = defineProps({
  modelValue: { type: Boolean, default: false },
  channel: { type: Object, default: null }, // 传行数据=编辑，传 null=新增
})
const emit = defineEmits(['update:modelValue', 'success'])

const visible = computed({
  get: () => props.modelValue,
  set: (v) => emit('update:modelValue', v),
})
const isEdit = computed(() => !!props.channel?.id)
const title = computed(() => (isEdit.value ? '编辑渠道' : '新增渠道'))

const emptyForm = () => ({
  name: '',
  api_url: '',
  api_key: '',
  status: 1, // 1=可用 0=熔断
})

const form = reactive(emptyForm())
const formRef = ref(null)
const submitting = ref(false)

const rules = {
  name: [{ required: true, message: '请输入渠道名称', trigger: 'blur' }],
  api_url: [{ required: true, message: '请输入接口地址', trigger: 'blur' }],
  api_key: [
    {
      // 后端不回显密钥，编辑时留空表示不修改
      validator: (rule, value, callback) => {
        if (!isEdit.value && !value) return callback(new Error('请输入接口密钥'))
        callback()
      },
      trigger: 'blur',
    },
  ],
}

function pick(c) {
  if (!c) return {}
  return {
    name: c.name ?? '',
    api_url: c.api_url ?? '',
    api_key: '',
    status: c.status ?? 1,
  }
}

watch(
  () => props.modelValue,
  (open) => {
    if (!open) return
    Object.assign(form, emptyForm(), pick(props.channel))
    nextTick(() => formRef.value?.clearValidate())
  },
)

async function submit() {
  const ok = await formRef.value.validate().catch(() => false)
  if (!ok) return
  submitting.value = true
  try {
    // api_key 留空时不提交，后端保持原密钥不变
    const payload = { name: form.name, api_url: form.api_url, status: form.status }
    if (form.api_key) payload.api_key = form.api_key
    if (isEdit.value) await updateChannel(props.channel.id, payload)
    else await createChannel(payload)
    ElMessage.success(isEdit.value ? '编辑成功' : '新增成功')
    emit('success')
    visible.value = false
  } finally {
    submitting.value = false
  }
}
</script>

<template>
  <el-dialog
    v-model="visible"
    :title="title"
    width="560px"
    :close-on-click-modal="false"
    destroy-on-close
    class="channel-dialog"
  >
    <el-form ref="formRef" :model="form" :rules="rules" label-width="100px">
      <el-form-item label="渠道名称" prop="name">
        <el-input v-model="form.name" placeholder="如：渠道A" maxlength="128" show-word-limit clearable />
      </el-form-item>

      <el-form-item label="接口地址" prop="api_url">
        <el-input v-model="form.api_url" placeholder="上游充值接口地址" maxlength="512" show-word-limit clearable />
      </el-form-item>

      <el-form-item label="接口密钥" prop="api_key">
        <el-input
          v-model="form.api_key"
          type="password"
          show-password
          clearable
          maxlength="255"
          :placeholder="isEdit ? '留空则不修改原密钥' : '请输入接口密钥'"
        />
      </el-form-item>

      <el-form-item label="健康状态">
        <div class="status-field">
          <el-switch
            v-model="form.status"
            :active-value="1"
            :inactive-value="0"
            active-text="可用"
            inactive-text="熔断"
          />
          <span class="hint">熔断状态的渠道不能用于商品上架</span>
        </div>
      </el-form-item>
    </el-form>

    <template #footer>
      <el-button @click="visible = false">取消</el-button>
      <el-button type="primary" :loading="submitting" @click="submit">确定</el-button>
    </template>
  </el-dialog>
</template>

<style scoped>
.status-field {
  display: flex;
  align-items: center;
  gap: 12px;
}
.hint {
  color: #9ca3af;
  font-size: 12px;
}
</style>
