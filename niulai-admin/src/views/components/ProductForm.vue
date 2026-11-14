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

// 价格类别行用稳定 key，删除中间行时不会串值
let denomUid = 0
const newDenomination = () => ({ _uid: ++denomUid, name: '', face_value: null })

const emptyForm = () => ({
  name: '',
  type: '',
  cover_image: '',
  cost_price: null,
  discount: 10, // 几折，10=不打折
  denominations: [newDenomination()],
  channel_id: null,
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

const checkDiscount = (rule, value, callback) => {
  if (value === null || value === undefined || value === '') return callback(new Error('请输入折扣'))
  if (!isNumber(value)) return callback(new Error('折扣必须是数字'))
  if (value <= 0 || value > 10) return callback(new Error('折扣必须在 0~10 之间（如 6 表示 6 折）'))
  callback()
}

const checkFaceValue = (rule, value, callback) => {
  if (value === null || value === undefined || value === '') return callback(new Error('请输入面值'))
  if (!isNumber(value)) return callback(new Error('面值必须是数字'))
  if (value <= 0) return callback(new Error('面值必须大于 0'))
  callback()
}

const faceValueRules = [{ validator: checkFaceValue, trigger: ['blur', 'change'] }]

const rules = {
  name: [{ required: true, message: '请输入商品名称', trigger: 'blur' }],
  type: [{ required: true, message: '请选择商品类型', trigger: 'change' }],
  cover_image: [{ required: true, message: '请上传封面图片', trigger: 'change' }],
  channel_id: [{ required: true, message: '请选择绑定渠道', trigger: 'change' }],
  cost_price: [{ validator: checkPrice('成本价'), trigger: ['blur', 'change'] }],
  discount: [{ validator: checkDiscount, trigger: ['blur', 'change'] }],
}

// 折后价预览：面值 × 折扣 ÷ 10，仅用于展示，计算以后端为准
const previewPrice = (d) => {
  const fv = Number(d.face_value)
  const disc = Number(form.discount)
  if (!isNumber(fv) || fv <= 0 || !isNumber(disc) || disc <= 0) return '—'
  return (fv * disc / 10).toFixed(2)
}
const examplePrice = computed(() => {
  const disc = Number(form.discount)
  if (!isNumber(disc) || disc <= 0) return '—'
  return (200 * disc / 10).toFixed(2)
})

function addDenomination() {
  form.denominations.push(newDenomination())
}
function removeDenomination(index) {
  form.denominations.splice(index, 1)
}

function pick(p) {
  if (!p) return {}
  // 不带 status：编辑时不改状态，避免触发后端「上架需校验渠道」的限制
  return {
    name: p.name ?? '',
    type: p.type ?? '',
    cover_image: p.cover_image ?? '',
    cost_price: p.cost_price == null ? null : Number(p.cost_price),
    discount: p.discount == null ? 10 : Number(p.discount),
    denominations:
      Array.isArray(p.denominations) && p.denominations.length
        ? [...p.denominations]
            .sort((a, b) => Number(a.sort) - Number(b.sort))
            .map((d) => ({
              _uid: ++denomUid,
              name: d.name ?? '',
              face_value: d.face_value == null ? null : Number(d.face_value),
            }))
        : [newDenomination()],
    channel_id: p.channel_id == null ? null : Number(p.channel_id),
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
  if (!form.denominations.length) {
    ElMessage.warning('请至少添加一个价格类别')
    return
  }
  if (form.denominations.some((d) => !(Number(d.face_value) > 0))) {
    ElMessage.warning('价格类别的面值必须大于 0')
    return
  }
  if (uploading.value) {
    ElMessage.warning('封面还在上传中，请稍候')
    return
  }
  submitting.value = true
  try {
    // 后端金额 / 折扣 / 面值字段是 string（decimal 用字符串解析），不能传 number
    const payload = {
      name: form.name,
      type: form.type,
      cover_image: form.cover_image,
      cost_price: String(form.cost_price),
      discount: String(form.discount),
      denominations: form.denominations.map((d) => ({
        name: d.name || '',
        face_value: String(d.face_value),
      })),
      channel_id: form.channel_id,
      description: form.description,
    }
    if (!isEdit.value) payload.status = form.status
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
    width="660px"
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
          <el-form-item label="折扣" prop="discount">
            <el-input-number
              v-model="form.discount"
              :min="0"
              :max="10"
              :precision="2"
              :step="0.1"
              :controls="false"
              placeholder="如 6"
              style="width: 100%"
            />
          </el-form-item>
        </el-col>
      </el-row>

      <el-form-item label="价格类别">
        <div class="denom-field">
          <div class="denom-hint">
            折后价 = 面值 × 折扣 ÷ 10；当前折扣 <b>{{ form.discount }} 折</b>，
            例：面值 200 → <b>¥{{ examplePrice }}</b>。可添加多个价格类别，共用同一折扣。
          </div>

          <div v-for="(d, index) in form.denominations" :key="d._uid" class="denom-row">
            <span class="denom-index">{{ index + 1 }}</span>
            <el-input v-model="d.name" class="denom-name" placeholder="类别名称（选填）" maxlength="64" clearable />
            <el-form-item
              class="denom-subitem denom-face"
              label-width="0px"
              :prop="`denominations.${index}.face_value`"
              :rules="faceValueRules"
            >
              <el-input-number
                v-model="d.face_value"
                :min="0"
                :precision="2"
                :step="1"
                :controls="false"
                placeholder="面值"
                style="width: 100%"
              />
            </el-form-item>
            <span class="denom-final">折后 ¥{{ previewPrice(d) }}</span>
            <el-button
              link
              type="danger"
              :icon="Delete"
              :disabled="form.denominations.length <= 1"
              @click="removeDenomination(index)"
            />
          </div>

          <el-button :icon="Plus" size="small" @click="addDenomination">添加价格类别</el-button>
        </div>
      </el-form-item>

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
.denom-field {
  width: 100%;
}
.denom-hint {
  color: #9ca3af;
  font-size: 12px;
  line-height: 1.6;
  margin-bottom: 10px;
}
.denom-hint b {
  color: #409eff;
}
.denom-row {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 8px;
}
.denom-index {
  flex: 0 0 18px;
  text-align: center;
  color: #9ca3af;
  font-size: 12px;
}
.denom-name {
  flex: 1 1 150px;
  min-width: 0;
}
.denom-face {
  flex: 1 1 130px;
  min-width: 0;
}
.denom-subitem {
  margin-bottom: 0 !important;
}
.denom-final {
  flex: 0 0 100px;
  text-align: right;
  white-space: nowrap;
  font-size: 12px;
  color: #e6a23c;
}
</style>
