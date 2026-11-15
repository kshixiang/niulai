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

// 话费商品必须填写的项：field_key 与服务端 defaultPhoneFields 一致，后台可自行调整其它内容
const PHONE_DEFAULT_FIELDS = [
  { field_key: 'phone', label: '手机号', type: 'text', required: true, placeholder: '请输入充值手机号', options: [] },
  { field_key: 'username', label: '用户名', type: 'text', required: true, placeholder: '请输入用户名', options: [] },
  { field_key: 'carrier', label: '运营商', type: 'select', required: true, placeholder: '请选择运营商', options: ['移动', '电信', '联通', '广电'] },
]
const FIELD_TYPE_OPTIONS = [
  { label: '文本', value: 'text' },
  { label: '数字', value: 'number' },
  { label: '下拉选择', value: 'select' },
]
const COVER_TYPES = ['image/jpeg', 'image/png', 'image/gif', 'image/webp']
const MAX_COVER_SIZE = 2 * 1024 * 1024

// 价格类别行 / 填写项行用稳定 key，删除中间行时不会串值
let denomUid = 0
const newDenomination = () => ({ _uid: ++denomUid, name: '', face_value: null })
let fieldUid = 0
const newField = () => ({
  _uid: ++fieldUid,
  field_key: '',
  label: '',
  type: 'text',
  required: false,
  placeholder: '',
  options: [],
})

const emptyForm = () => ({
  name: '',
  type: '',
  cover_image: '',
  channel_discount: 10, // 渠道折扣(几折)，10=不打折
  discount: 10, // 出售折扣(几折)
  denominations: [newDenomination()],
  fields: [], // 下单填写项，可选
  channel_id: null,
  description: '',
  status: 0, // 仅新增时提交：1=立即上架 0=暂不上架
})

const form = reactive(emptyForm())
const formRef = ref(null)
const submitting = ref(false)
const uploading = ref(false)

const isNumber = (v) => typeof v === 'number' && !Number.isNaN(v)

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

// 字段标识：小写字母开头，仅小写字母/数字/下划线，且同一商品内唯一
const fieldKeyRules = (index) => [
  {
    validator: (rule, value, callback) => {
      const key = (value ?? '').trim()
      if (!key) return callback(new Error('请输入英文标识'))
      if (!/^[a-z][a-z0-9_]*$/.test(key)) return callback(new Error('小写字母开头，仅小写字母/数字/下划线'))
      const dup = form.fields.some((f, i) => i !== index && (f.field_key ?? '').trim() === key)
      if (dup) return callback(new Error('标识重复'))
      callback()
    },
    trigger: ['blur', 'change'],
  },
]
const labelRules = [
  { validator: (rule, value, callback) => ((value ?? '').trim() ? callback() : callback(new Error('请输入中文名'))), trigger: ['blur', 'change'] },
]
const optionsRules = (index) => [
  {
    validator: (rule, value, callback) => {
      const f = form.fields[index]
      if (f && f.type === 'select' && (!Array.isArray(value) || value.length === 0)) {
        return callback(new Error('请配置下拉选项'))
      }
      callback()
    },
    trigger: ['change'],
  },
]

const rules = {
  name: [{ required: true, message: '请输入商品名称', trigger: 'blur' }],
  type: [{ required: true, message: '请选择商品类型', trigger: 'change' }],
  cover_image: [{ required: true, message: '请上传封面图片', trigger: 'change' }],
  channel_id: [{ required: true, message: '请选择绑定渠道', trigger: 'change' }],
  channel_discount: [{ validator: checkDiscount, trigger: ['blur', 'change'] }],
  discount: [{ validator: checkDiscount, trigger: ['blur', 'change'] }],
}

// 金额预览：面值 × 折扣 ÷ 10，仅用于未保存时的即时展示，实际金额以后端 decimal 计算为准
const priceOf = (faceValue, discount) => {
  const fv = Number(faceValue)
  const disc = Number(discount)
  if (!isNumber(fv) || fv <= 0 || !isNumber(disc) || disc <= 0) return null
  return fv * disc / 10
}
const previewChannelPrice = (d) => {
  const v = priceOf(d.face_value, form.channel_discount)
  return v == null ? '—' : v.toFixed(2)
}
const previewSellPrice = (d) => {
  const v = priceOf(d.face_value, form.discount)
  return v == null ? '—' : v.toFixed(2)
}
const previewProfit = (d) => {
  const channel = priceOf(d.face_value, form.channel_discount)
  const sell = priceOf(d.face_value, form.discount)
  return channel == null || sell == null ? '—' : (sell - channel).toFixed(2)
}
const exampleLine = computed(() => {
  const channel = priceOf(200, form.channel_discount)
  const sell = priceOf(200, form.discount)
  if (channel == null || sell == null) return '—'
  return `渠道价 ¥${channel.toFixed(2)} / 售价 ¥${sell.toFixed(2)} / 盈利 ¥${(sell - channel).toFixed(2)}`
})

function addDenomination() {
  form.denominations.push(newDenomination())
}
function removeDenomination(index) {
  form.denominations.splice(index, 1)
  nextTick(() => formRef.value?.clearValidate())
}
// 话费商品补齐后台未配置的手机号/用户名必填项，规则与服务端 appendPhoneFields 一致
function ensurePhoneFields() {
  if (form.type !== '话费') return
  const existing = new Set(form.fields.map((f) => (f.field_key || '').trim()))
  const missing = PHONE_DEFAULT_FIELDS.filter((f) => !existing.has(f.field_key))
  if (missing.length) {
    form.fields = [...missing.map((f) => ({ ...f, _uid: ++fieldUid })), ...form.fields]
  }
}

function addField() {
  form.fields.push(newField())
  nextTick(() => formRef.value?.clearValidate())
}
function removeField(index) {
  form.fields.splice(index, 1)
  nextTick(() => formRef.value?.clearValidate())
}

function pick(p) {
  if (!p) return {}
  // 不带 status：编辑时不改状态，避免触发后端「上架需校验渠道」的限制
  return {
    name: p.name ?? '',
    type: p.type ?? '',
    cover_image: p.cover_image ?? '',
    channel_discount: p.channel_discount == null ? 10 : Number(p.channel_discount),
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
    fields: Array.isArray(p.fields)
      ? [...p.fields]
          .sort((a, b) => Number(a.sort) - Number(b.sort))
          .map((f) => ({
            _uid: ++fieldUid,
            field_key: f.field_key ?? '',
            label: f.label ?? '',
            type: f.type ?? 'text',
            required: !!f.required,
            placeholder: f.placeholder ?? '',
            options: Array.isArray(f.options) ? [...f.options] : [],
          }))
      : [],
    channel_id: p.channel_id == null ? null : Number(p.channel_id),
    description: p.description ?? '',
  }
}

watch(
  () => props.modelValue,
  (open) => {
    if (!open) return
    Object.assign(form, emptyForm(), pick(props.product))
    ensurePhoneFields() // 编辑存量话费商品时也补齐必填项
    nextTick(() => formRef.value?.clearValidate())
  },
)

// 新增商品时选择「话费」立即带出默认填写项
watch(() => form.type, ensurePhoneFields)

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
      channel_discount: String(form.channel_discount),
      discount: String(form.discount),
      denominations: form.denominations.map((d) => ({
        name: d.name || '',
        face_value: String(d.face_value),
      })),
      fields: form.fields.map((f) => ({
        field_key: (f.field_key || '').trim(),
        label: (f.label || '').trim(),
        type: f.type,
        required: !!f.required,
        placeholder: f.placeholder || '',
        options: f.type === 'select' ? (f.options || []) : [],
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
    width="720px"
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
          <el-form-item label="渠道折扣" prop="channel_discount">
            <el-input-number
              v-model="form.channel_discount"
              :min="0"
              :max="10"
              :precision="2"
              :step="0.1"
              :controls="false"
              placeholder="如 5"
              style="width: 100%"
            />
          </el-form-item>
        </el-col>
        <el-col :span="12">
          <el-form-item label="出售折扣" prop="discount">
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
            渠道价 = 面值 × 渠道折扣 ÷ 10，售价 = 面值 × 出售折扣 ÷ 10，盈利 = 售价 − 渠道价。<br />
            当前渠道 <b>{{ form.channel_discount }} 折</b> / 出售 <b>{{ form.discount }} 折</b>；
            例：面值 200 → <b>{{ exampleLine }}</b>。（预览，以后端为准）
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
            <div class="denom-prices">
              <span>渠道 ¥{{ previewChannelPrice(d) }}</span>
              <span>售 ¥{{ previewSellPrice(d) }}</span>
              <span class="denom-profit">赚 ¥{{ previewProfit(d) }}</span>
            </div>
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

      <el-form-item label="下单填写项">
        <div class="field-editor">
          <div class="field-hint">
            配置 H5 下单页需要用户填写的字段，前端按此动态渲染；可随时增删改。下拉类型需配置选项。
          </div>

          <div v-for="(f, index) in form.fields" :key="f._uid" class="field-row">
            <div class="field-row__head">
              <span class="field-index">{{ index + 1 }}</span>
              <el-form-item
                class="field-subitem field-key"
                label-width="0px"
                :prop="`fields.${index}.field_key`"
                :rules="fieldKeyRules(index)"
              >
                <el-input v-model="f.field_key" placeholder="英文标识 如 phone" maxlength="64" clearable />
              </el-form-item>
              <el-form-item
                class="field-subitem field-label"
                label-width="0px"
                :prop="`fields.${index}.label`"
                :rules="labelRules"
              >
                <el-input v-model="f.label" placeholder="中文名 如 手机号" maxlength="128" clearable />
              </el-form-item>
              <el-select v-model="f.type" class="field-type" placeholder="类型">
                <el-option v-for="t in FIELD_TYPE_OPTIONS" :key="t.value" :label="t.label" :value="t.value" />
              </el-select>
              <el-checkbox v-model="f.required" class="field-required">必填</el-checkbox>
              <el-button link type="danger" :icon="Delete" @click="removeField(index)" />
            </div>
            <div class="field-row__body">
              <el-input v-model="f.placeholder" class="field-placeholder" placeholder="输入提示语（选填）" maxlength="255" clearable />
              <el-form-item
                v-if="f.type === 'select'"
                class="field-subitem field-options"
                label-width="0px"
                :prop="`fields.${index}.options`"
                :rules="optionsRules(index)"
              >
                <el-select
                  v-model="f.options"
                  multiple
                  filterable
                  allow-create
                  default-first-option
                  :reserve-keyword="false"
                  placeholder="输入选项后回车添加，如 移动 / 联通 / 电信"
                  style="width: 100%"
                >
                  <el-option v-for="o in f.options" :key="o" :label="o" :value="o" />
                </el-select>
              </el-form-item>
            </div>
          </div>

          <el-button :icon="Plus" size="small" @click="addField">添加填写项</el-button>
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
.denom-field,
.field-editor {
  width: 100%;
}
.denom-hint,
.field-hint {
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
  flex: 1 1 130px;
  min-width: 0;
}
.denom-face {
  flex: 1 1 110px;
  min-width: 0;
}
.denom-subitem,
.field-subitem {
  margin-bottom: 0 !important;
}
.denom-prices {
  flex: 0 0 auto;
  display: flex;
  flex-direction: column;
  gap: 1px;
  font-size: 11px;
  line-height: 1.35;
  color: #9ca3af;
  white-space: nowrap;
}
.denom-profit {
  color: #e6a23c;
  font-weight: 700;
}
.field-row {
  border: 1px solid #ebeef5;
  border-radius: 8px;
  padding: 10px;
  margin-bottom: 10px;
  background: #fafafa;
}
.field-row__head,
.field-row__body {
  display: flex;
  align-items: center;
  gap: 8px;
}
.field-row__body {
  margin-top: 8px;
}
.field-index {
  flex: 0 0 18px;
  text-align: center;
  color: #9ca3af;
  font-size: 12px;
}
.field-key {
  flex: 1 1 150px;
  min-width: 0;
}
.field-label {
  flex: 1 1 150px;
  min-width: 0;
}
.field-type {
  flex: 0 0 120px;
}
.field-required {
  flex: 0 0 auto;
  margin-right: 0;
}
.field-placeholder {
  flex: 1 1 240px;
  min-width: 0;
}
.field-options {
  flex: 1 1 260px;
  min-width: 0;
}
</style>
