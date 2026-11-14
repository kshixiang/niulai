<script setup>
import { ref, reactive, computed, watch, nextTick } from 'vue'
import { ElMessage } from 'element-plus'
import { Plus, Delete } from '@element-plus/icons-vue'
import { createProduct, updateProduct } from '@/api/product'
import { uploadCover } from '@/api/upload'

const props = defineProps({
  modelValue: { type: Boolean, default: false },
  product: { type: Object, default: null }, // 传行数据=编辑，传 null=新增
  channels: { type: Array, default: () => [] },
})
const emit = defineEmits(['update:modelValue', 'success'])

const visible = computed({
  get: () => props.modelValue,
  set: (v) => emit('update:modelValue', v),
})
const isEdit = computed(() => !!props.product?.id)
const title = computed(() => (isEdit.value ? '编辑商品' : '新增商品'))

// 与后端 CreateProductRequest 的 binding:"oneof=话费 卡券" 保持一致
const TYPE_OPTIONS = ['话费', '卡券']
const COVER_TYPES = ['image/jpeg', 'image/png', 'image/gif', 'image/webp']
const MAX_COVER_SIZE = 2 * 1024 * 1024

const emptyForm = () => ({
  name: '',
  type: '',
  cover_image: '',
  cost_price: null,
  sale_price: null,
  channel_id: null,
  discount_label: '',
  description: '',
  status: 0, // 仅新增时提交：1=立即上架 0=暂不上架
})

const form = reactive(emptyForm())
const formRef = ref(null)
const submitting = ref(false)
const uploading = ref(false)

const isNumber = (v) => typeof v === 'number' && !Number.isNaN(v)

const checkPrice = (label) => (rule, value, callback) => {
  if (value === null || value === undefined || value === '') return callback(new Error(`请输入${label}`))
  if (!isNumber(value)) return callback(new Error(`${label}必须是数字`))
  if (value < 0) return callback(new Error(`${label}不能为负`))
  callback()
}

const validateSalePrice = (rule, value, callback) => {
  if (value === null || value === undefined || value === '') return callback(new Error('请输入销售价'))
  if (!isNumber(value)) return callback(new Error('销售价必须是数字'))
  if (value < 0) return callback(new Error('销售价不能为负'))
  if (isNumber(form.cost_price) && value < form.cost_price) return callback(new Error('销售价不能低于成本价'))
  callback()
}

const rules = {
  name: [{ required: true, message: '请输入商品名称', trigger: 'blur' }],
  type: [{ required: true, message: '请选择商品类型', trigger: 'change' }],
  cover_image: [{ required: true, message: '请上传封面图片', trigger: 'change' }],
  channel_id: [{ required: true, message: '请选择绑定渠道', trigger: 'change' }],
  cost_price: [{ validator: checkPrice('成本价'), trigger: ['blur', 'change'] }],
  sale_price: [{ validator: validateSalePrice, trigger: ['blur', 'change'] }],
}

function pick(p) {
  if (!p) return {}
  // 不带 status：编辑时不改状态，避免触发后端「上架需校验渠道」的限制
  return {
    name: p.name ?? '',
    type: p.type ?? '',
    cover_image: p.cover_image ?? '',
    cost_price: p.cost_price == null ? null : Number(p.cost_price),
    sale_price: p.sale_price == null ? null : Number(p.sale_price),
    channel_id: p.channel_id == null ? null : Number(p.channel_id),
    discount_label: p.discount_label ?? '',
    description: p.description ?? '',
  }
}

watch(
  () => props.modelValue,
  (open) => {
    if (!open) return
    Object.assign(form, emptyForm(), pick(props.product))
    nextTick(() => formRef.value?.clearValidate())
  },
)

// 成本价变动时联动重校销售价，保证「不低于成本价」实时生效
watch(
  () => form.cost_price,
  () => {
    if (isNumber(form.sale_price) && formRef.value) {
      formRef.value.validateField('sale_price').catch(() => {})
    }
  },
)

function beforeCoverUpload(file) {
  if (!COVER_TYPES.includes(file.type)) {
    ElMessage.error('仅支持 jpg / png / gif / webp 图片')
    return false
  }
  if (file.size > MAX_COVER_SIZE) {
    ElMessage.error('图片不能超过 2MB')
    return false
  }
  return true
}

async function handleCoverUpload({ file }) {
  uploading.value = true
  try {
    // 拦截器已解包 data → { url }
    const data = await uploadCover(file)
    form.cover_image = data.url
    // 清掉「请上传封面图片」的校验提示
    formRef.value?.validateField('cover_image').catch(() => {})
  } finally {
    uploading.value = false
  }
}

function removeCover() {
  form.cover_image = ''
}

async function submit() {
  const ok = await formRef.value.validate().catch(() => false)
  if (!ok) return
  if (uploading.value) {
    ElMessage.warning('封面还在上传中，请稍候')
    return
  }
  submitting.value = true
  try {
    // 后端金额字段是 string（decimal 用字符串解析），不能传 number
    const payload = {
      ...form,
      cost_price: String(form.cost_price),
      sale_price: String(form.sale_price),
    }
    if (isEdit.value) delete payload.status
    if (isEdit.value) await updateProduct(props.product.id, payload)
    else await createProduct(payload)
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
    width="620px"
    :close-on-click-modal="false"
    destroy-on-close
    class="product-dialog"
  >
    <el-alert
      v-if="!channels.length"
      class="channel-alert"
      type="warning"
      :closable="false"
      show-icon
      title="暂无可用渠道，请先创建渠道后再新增商品"
    />

    <el-form ref="formRef" :model="form" :rules="rules" label-width="100px">
      <el-form-item label="商品名称" prop="name">
        <el-input v-model="form.name" placeholder="请输入商品名称" maxlength="128" show-word-limit clearable />
      </el-form-item>

      <el-form-item label="商品类型" prop="type">
        <el-select v-model="form.type" placeholder="请选择商品类型" style="width: 100%">
          <el-option v-for="t in TYPE_OPTIONS" :key="t" :label="t" :value="t" />
        </el-select>
      </el-form-item>

      <el-form-item label="封面图片" prop="cover_image">
        <div class="cover-field">
          <el-upload
            class="cover-upload"
            :show-file-list="false"
            :http-request="handleCoverUpload"
            :before-upload="beforeCoverUpload"
            :disabled="uploading"
            accept="image/jpeg,image/png,image/gif,image/webp"
          >
            <div v-if="form.cover_image" class="cover-preview-wrap">
              <el-image :src="form.cover_image" fit="cover" class="cover-preview" />
              <span class="cover-change">点击更换</span>
            </div>
            <el-button v-else :loading="uploading" :icon="Plus">上传封面</el-button>
          </el-upload>
          <el-button v-if="form.cover_image" link type="danger" :icon="Delete" @click="removeCover">
            移除
          </el-button>
          <span class="cover-hint">jpg / png / gif / webp，≤2MB</span>
        </div>
      </el-form-item>

      <el-row :gutter="16">
        <el-col :span="12">
          <el-form-item label="成本价" prop="cost_price">
            <el-input-number
              v-model="form.cost_price"
              :min="0"
              :precision="2"
              :step="0.01"
              :controls="false"
              placeholder="0.00"
              style="width: 100%"
            />
          </el-form-item>
        </el-col>
        <el-col :span="12">
          <el-form-item label="销售价" prop="sale_price">
            <el-input-number
              v-model="form.sale_price"
              :min="0"
              :precision="2"
              :step="0.01"
              :controls="false"
              placeholder="0.00"
              style="width: 100%"
            />
          </el-form-item>
        </el-col>
      </el-row>

      <el-form-item label="绑定渠道" prop="channel_id">
        <el-select v-model="form.channel_id" placeholder="请选择渠道" filterable style="width: 100%">
          <el-option v-for="c in channels" :key="c.id" :label="c.name" :value="Number(c.id)" />
        </el-select>
      </el-form-item>

      <el-form-item v-if="!isEdit" label="上架状态">
        <div class="status-field">
          <el-switch
            v-model="form.status"
            :active-value="1"
            :inactive-value="0"
            active-text="立即上架"
            inactive-text="暂不上架"
          />
          <span class="status-hint">上架需绑定可用的上游渠道</span>
        </div>
      </el-form-item>

      <el-form-item label="折扣标签" prop="discount_label">
        <el-input v-model="form.discount_label" placeholder="如：7.8折（留空则不展示）" clearable />
      </el-form-item>

      <el-form-item label="描述提示语" prop="description">
        <el-input
          v-model="form.description"
          type="textarea"
          :rows="3"
          maxlength="512"
          show-word-limit
          placeholder="用户可见的提示语 / 商品描述"
        />
      </el-form-item>
    </el-form>

    <template #footer>
      <el-button @click="visible = false">取消</el-button>
      <el-button type="primary" :loading="submitting || uploading" @click="submit">确定</el-button>
    </template>
  </el-dialog>
</template>

<style scoped>
.channel-alert {
  margin-bottom: 18px;
}
.cover-field {
  display: flex;
  align-items: center;
  gap: 12px;
  width: 100%;
}
.cover-preview-wrap {
  position: relative;
  width: 72px;
  height: 72px;
  cursor: pointer;
}
.cover-preview {
  width: 72px;
  height: 72px;
  border-radius: 8px;
  border: 1px solid #ebeef5;
}
.cover-change {
  position: absolute;
  left: 0;
  right: 0;
  bottom: 0;
  line-height: 20px;
  text-align: center;
  font-size: 12px;
  color: #fff;
  background: rgba(0, 0, 0, 0.55);
  border-radius: 0 0 8px 8px;
}
.cover-hint,
.status-hint {
  color: #9ca3af;
  font-size: 12px;
}
.status-field {
  display: flex;
  align-items: center;
  gap: 12px;
}
</style>
