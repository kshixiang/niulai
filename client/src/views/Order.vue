<script setup>
import { ref, reactive, computed, watch, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { showToast } from 'vant'
import { getProductDetail, createOrder, getWallet } from '@/api'

const route = useRoute()
const router = useRouter()

const loading = ref(true)
const submitting = ref(false)
const product = ref(null)
const form = reactive({}) // 按后台配置的 field_key 动态添加键
const selectedId = ref(null) // 选中的价格类别 id
const customAmount = ref('') // 自定义面值
const customFieldRef = ref(null)
const showDesc = ref(false)
const showInsufficient = ref(false) // 余额不足弹窗

const fields = computed(() => product.value?.fields || [])
const denominations = computed(() => product.value?.denominations || [])

const discount = computed(() => {
  const n = Number(product.value?.discount)
  return Number.isFinite(n) ? n : null
})
const discountText = computed(() => (discount.value == null ? '—' : `${discount.value}折`))

const selectedDenom = computed(() =>
  denominations.value.find((d) => d.id === selectedId.value) || null,
)

// 订单金额：自定义金额优先，否则取选中档的面值
const amount = computed(() => {
  if (customAmount.value !== '') {
    const n = Number(customAmount.value)
    return Number.isFinite(n) && n > 0 ? n : 0
  }
  const d = selectedDenom.value
  return d ? Number(d.face_value) || 0 : 0
})

// 应付金额(售价)：预设档直接用后端算好的 final_price(decimal 权威值)；
// 仅自定义金额时用 面值 × 出售折扣 ÷ 10 即时试算（不落库）。
const sellPrice = computed(() => {
  if (!amount.value) return null
  if (selectedDenom.value && customAmount.value === '') {
    const fp = Number(selectedDenom.value.final_price)
    return Number.isFinite(fp) ? fp : null
  }
  if (discount.value == null) return null
  return Math.round((amount.value * discount.value / 10) * 100) / 100
})

const discountAmount = computed(() =>
  sellPrice.value == null ? null : Math.round((amount.value - sellPrice.value) * 100) / 100,
)

const fmt = (v) => (v == null || !Number.isFinite(Number(v)) ? '-' : Number(v).toFixed(2))

const chipLabel = (d) => {
  const face = Number(d.face_value)
  const faceText = Number.isFinite(face) ? String(face) : '—'
  return d.name ? `${d.name} ${faceText}` : faceText
}

function selectDenom(d) {
  selectedId.value = d.id
  customAmount.value = ''
}
watch(customAmount, (v) => {
  if (v !== '') selectedId.value = null
})
function useCustom() {
  selectedId.value = null
  customFieldRef.value?.focus?.()
}
function goBack() {
  if (window.history.state?.back) router.back()
  else router.replace('/home')
}
// 立即充值：跳到「我的」，顶部余额卡片即充值入口
function goRecharge() {
  showInsufficient.value = false
  router.push('/mine')
}

onMounted(async () => {
  try {
    const data = await getProductDetail(route.params.id)
    product.value = data || null
    // 按配置初始化动态表单键（Vue3 Proxy 支持动态加键）
    fields.value.forEach((f) => {
      form[f.field_key] = ''
    })
  } catch {
    // 错误提示由 request.js 拦截器统一弹出
    product.value = null
  } finally {
    loading.value = false
  }
})

// 与服务端 phonePattern 保持一致：中国大陆 11 位手机号
const PHONE_PATTERN = /^1[3-9]\d{9}$/

function validate() {
  if (!amount.value || amount.value <= 0) {
    showToast('请选择或输入订单金额')
    return false
  }
  for (const f of fields.value) {
    const v = String(form[f.field_key] ?? '').trim()
    if (f.required && !v) {
      showToast(`请填写${f.label}`)
      return false
    }
    if (f.field_key === 'phone' && v && !PHONE_PATTERN.test(v)) {
      showToast('请输入正确的11位手机号')
      return false
    }
  }
  return true
}

// 主充值账号取第一个有值的填写项（后台配置顺序中通常是手机号/充值账号）
function accountOfFields() {
  const hit = fields.value.find((f) => String(form[f.field_key] ?? '').trim())
  return hit ? String(form[hit.field_key]).trim() : ''
}

async function submit() {
  if (submitting.value || !validate()) return
  submitting.value = true
  try {
    // 余额预检：读取当前可用余额，与应付金额比较（价格均按 USDT 结算，直接比较）
    const wallet = await getWallet()
    if (Number(wallet.balance) < (sellPrice.value || 0)) {
      showInsufficient.value = true
      return
    }
    const payload = {
      product_id: Number(route.params.id),
      fields: { ...form },
      account: accountOfFields(),
    }
    // 预设档只传 id，面值由后端读库取权威值；自定义金额才传前端输入的面值
    if (selectedId.value != null) {
      payload.denomination_id = selectedId.value
    } else {
      payload.face_value = String(amount.value)
    }
    const data = await createOrder(payload)
    showToast('下单成功')
    router.replace(`/order-detail/${data.id}`)
  } catch {
    // 错误提示由 request.js 拦截器统一弹出
  } finally {
    submitting.value = false
  }
}
</script>

<template>
  <div class="order">
    <van-nav-bar title="下单" left-arrow class="order__nav" @click-left="goBack" />

    <div class="order__body">
      <div v-if="loading" class="order__loading">
        <van-loading size="20" /> 加载中…
      </div>

      <template v-else-if="product">
        <!-- 商品名 + 充值说明 -->
        <div class="order__head">
          <h2 class="order__name">{{ product.name }}</h2>
          <button type="button" class="order__desc-link" @click="showDesc = !showDesc">
            充值说明 <van-icon name="question-o" />
          </button>
        </div>
        <p v-if="showDesc" class="order__desc">{{ product.description || '暂无充值说明' }}</p>

        <!-- 动态填写项：由后台配置驱动 -->
        <section v-if="fields.length" class="order__card">
          <h3 class="order__section-title">填写信息</h3>
          <van-cell-group inset class="order__fields">
            <van-field
              v-for="f in fields"
              :key="f.field_key"
              :label="f.label"
              :required="!!f.required"
              :placeholder="f.placeholder"
            >
              <template #input>
                <select
                  v-if="f.type === 'select'"
                  v-model="form[f.field_key]"
                  class="field-select"
                >
                  <option value="" disabled>{{ f.placeholder || '请选择' }}</option>
                  <option v-for="o in f.options" :key="o" :value="o">{{ o }}</option>
                </select>
                <input
                  v-else
                  v-model="form[f.field_key]"
                  class="field-input"
                  :type="f.field_key === 'phone' ? 'tel' : (f.type === 'number' ? 'number' : 'text')"
                  :maxlength="f.field_key === 'phone' ? 11 : undefined"
                  :inputmode="f.field_key === 'phone' ? 'numeric' : undefined"
                  :placeholder="f.placeholder"
                />
              </template>
            </van-field>
          </van-cell-group>
        </section>

        <!-- 订单金额：后台配置的价格类别 + 自定义 -->
        <section class="order__card">
          <h3 class="order__section-title">订单金额</h3>
          <div class="chips">
            <button
              v-for="d in denominations"
              :key="d.id"
              type="button"
              class="chip"
              :class="{ 'is-active': selectedId === d.id }"
              @click="selectDenom(d)"
            >{{ chipLabel(d) }}</button>
            <button
              type="button"
              class="chip"
              :class="{ 'is-active': customAmount !== '' }"
              @click="useCustom"
            >自定义金额</button>
          </div>
          <van-cell-group inset class="order__fields order__fields--custom">
            <van-field
              ref="customFieldRef"
              v-model="customAmount"
              type="number"
              label="自定义金额"
              placeholder="输入其它面值"
            />
          </van-cell-group>
        </section>

        <!-- 汇总面板 -->
        <section class="order__card summary">
          <div class="summary__row">
            <span>订单金额</span><span>{{ amount ? fmt(amount) : '-' }} USDT</span>
          </div>
          <div class="summary__row">
            <span>折扣</span><span>{{ discountText }}</span>
          </div>
          <div class="summary__row">
            <span>优惠金额</span><span>{{ fmt(discountAmount) }} USDT</span>
          </div>
          <div class="summary__row summary__row--total">
            <span>应付金额</span><span class="summary__pay">{{ fmt(sellPrice) }} USDT</span>
          </div>
        </section>
      </template>

      <div v-else class="order__empty">商品不存在或已下架</div>
    </div>

    <div class="order__footer">
      <van-button
        block
        round
        class="order__submit"
        :loading="submitting"
        :disabled="submitting || loading || !product"
        @click="submit"
      >下单</van-button>
    </div>

    <!-- 余额不足提示：挂到 body，避免被滚动容器裁剪 -->
    <Teleport to="body">
      <div v-if="showInsufficient" class="mask" @click.self="showInsufficient = false">
        <div class="dialog">
          <button type="button" class="dialog__close" aria-label="关闭" @click="showInsufficient = false">
            <van-icon name="cross" />
          </button>
          <p class="dialog__msg">余额不足，请先充值</p>
          <div class="dialog__actions">
            <button type="button" class="dialog__btn dialog__btn--outline" @click="showInsufficient = false">取消</button>
            <button type="button" class="dialog__btn dialog__btn--primary" @click="goRecharge">立即充值</button>
          </div>
        </div>
      </div>
    </Teleport>
  </div>
</template>

<style scoped>
.order {
  height: 100%;
  display: flex;
  flex-direction: column;
  background: var(--bg-page);
  --van-field-label-color: var(--text-secondary);
}
.order__nav {
  flex: 0 0 auto;
  background: var(--bg-card);
}
.order :deep(.van-nav-bar__title) { color: var(--text-primary); font-weight: 700; }
.order :deep(.van-nav-bar__arrow) { color: var(--text-primary); }
.order :deep(.van-nav-bar__text) { color: var(--text-primary); }

.order__body {
  flex: 1;
  overflow-y: auto;
  padding: 16px 16px 24px;
}
.order__loading,
.order__empty {
  padding: 48px 0;
  text-align: center;
  font-size: 12px;
  color: var(--text-muted);
}
.order__loading { display: flex; align-items: center; justify-content: center; gap: 8px; }

.order__head { display: flex; align-items: center; justify-content: space-between; gap: 12px; margin-bottom: 4px; }
.order__name { font-size: 22px; font-weight: 800; color: var(--text-primary); }
.order__desc-link {
  display: inline-flex; align-items: center; gap: 2px;
  background: none; border: none; padding: 4px;
  color: var(--text-secondary); font-size: 12px;
}
.order__desc {
  margin: 6px 0 14px; padding: 10px 12px;
  border-radius: 10px; font-size: 12px; line-height: 1.7;
  background: var(--c-fill); color: var(--text-secondary);
}

.order__card {
  margin-top: 14px;
  padding: 14px;
  border: 1px solid var(--border-card);
  border-radius: var(--radius-card);
  background: var(--bg-card);
}
.order__section-title {
  margin: 0 0 12px; font-size: 14px; font-weight: 700; color: var(--text-primary);
}
.order__fields { margin: 0 !important; }
.order__fields--custom { margin-top: 12px !important; }
.order :deep(.van-cell-group--inset) { margin: 0; }
.order :deep(.van-field__control::placeholder) { color: var(--text-muted); }

.field-input,
.field-select {
  width: 100%; border: none; outline: none; background: transparent;
  color: var(--text-primary); font-size: 14px;
}
.field-select {
  padding: 2px 0;
  color-scheme: dark;
}
.field-select option { background: var(--c-surface-2); color: var(--text-primary); }

.chips { display: grid; grid-template-columns: repeat(3, minmax(0, 1fr)); gap: 10px; }
.chip {
  padding: 12px 6px; border: 1px solid var(--border-card); border-radius: 10px;
  background: var(--c-surface-2); color: var(--text-primary);
  font-size: 13px; font-weight: 700; text-align: center;
}
.chip.is-active {
  background: var(--gradient-brand); color: #fff; border-color: transparent;
}

.summary__row {
  display: flex; align-items: center; justify-content: space-between;
  padding: 8px 0; font-size: 13px; color: var(--text-secondary);
}
.summary__row + .summary__row { border-top: 1px solid var(--c-line); }
.summary__row--total { color: var(--text-primary); font-weight: 700; }
.summary__pay { font-size: 17px; font-weight: 800; color: var(--c-accent); }

.order__footer {
  flex: 0 0 auto;
  padding: 12px 16px calc(12px + var(--safe-bottom));
  background: var(--bg-page);
}
.order__submit { background: var(--gradient-brand); border: none; color: #fff; font-weight: 700; }
.order__submit :deep(.van-button__text) { color: #fff; }

/* ---------- 余额不足弹层 ---------- */
.mask {
  position: fixed; inset: 0; z-index: 2000;
  display: grid; place-items: center; padding: 24px;
  background: rgba(0, 0, 0, .6);
  -webkit-backdrop-filter: blur(2px); backdrop-filter: blur(2px);
}
.dialog {
  position: relative;
  width: 100%; max-width: 320px; padding: 26px 20px 20px;
  border: 1px solid var(--border-card); border-radius: 18px;
  background: var(--c-surface-2);
  box-shadow: 0 18px 50px rgba(0, 0, 0, .5);
}
.dialog__close {
  position: absolute; top: 10px; right: 10px;
  width: 26px; height: 26px; display: grid; place-items: center;
  border: none; background: none; color: var(--text-secondary); font-size: 18px;
}
.dialog__msg { margin: 0; font-size: 15px; font-weight: 700; color: var(--text-primary); text-align: center; }
.dialog__actions { display: flex; gap: 12px; margin-top: 24px; }
.dialog__btn {
  flex: 1; height: 42px; border-radius: var(--radius-pill);
  border: 1px solid var(--border-card); background: var(--bg-card);
  color: var(--text-secondary); font-size: 14px; font-weight: 700;
}
.dialog__btn--outline { background: transparent; border-color: var(--c-brand); color: var(--c-brand); }
.dialog__btn--primary { background: var(--gradient-brand); border: none; color: #fff; }
</style>
