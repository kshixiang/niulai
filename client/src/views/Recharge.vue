<script setup>
import { ref, computed, onMounted, onBeforeUnmount } from 'vue'
import { useRouter } from 'vue-router'
import { Dialog as VanDialog, showToast } from 'vant'
import QRCode from 'qrcode'
import { getRechargeConfig, createRecharge } from '@/api'
import { copyText } from '@/utils/clipboard'
import { normalizeRechargeSubmission } from '@/utils/recharge'

const router = useRouter()
const loading = ref(true)
const failed = ref(false)
const config = ref(null)
const qr = ref('')
const qrFailed = ref(false)
const showForm = ref(false)
const amount = ref('')
const txHash = ref('')
const formError = ref('')
const submitting = ref(false)
let alive = true
const available = computed(() => config.value?.enabled && !!config.value.address)
const preview = computed(() => {
  try { return normalizeRechargeSubmission(config.value?.address, amount.value, txHash.value).creditedAmount }
  catch { return '' }
})

function goBack() {
  if (window.history.state?.back) router.back()
  else router.replace('/mine')
}

async function loadConfig() {
  if (submitting.value) return
  loading.value = true
  failed.value = false
  config.value = null
  qr.value = ''
  qrFailed.value = false
  try {
    const data = await getRechargeConfig()
    if (!alive) return
    config.value = data
    if (data?.enabled && data.address) {
      try {
        qr.value = await QRCode.toDataURL(data.address, { width: 420, margin: 2, errorCorrectionLevel: 'H' })
      } catch { qrFailed.value = true }
    }
  } catch { failed.value = true }
  finally { loading.value = false }
}

async function copyAddress() {
  if (!available.value) return
  try { await copyText(config.value.address); showToast('充值地址已复制') }
  catch { showToast('复制失败，请长按地址手动复制') }
}

function openForm() {
  if (!available.value || loading.value) return
  formError.value = ''
  showForm.value = true
}

async function submit() {
  if (submitting.value) return
  let submission
  try { submission = normalizeRechargeSubmission(config.value?.address, amount.value, txHash.value) }
  catch (error) { formError.value = error.message; return }
  formError.value = ''
  submitting.value = true
  try {
    await createRecharge(submission.payload)
    showForm.value = false
    showToast('申报已提交，等待核实到账')
    router.push('/recharge-records')
  } catch (error) { formError.value = error.message || '提交失败，请重试' }
  finally { submitting.value = false }
}

onMounted(loadConfig)
onBeforeUnmount(() => { alive = false })
</script>

<template>
  <div class="recharge">
    <header class="recharge__header">
      <button type="button" class="recharge__back" aria-label="返回" @click="goBack"><van-icon name="arrow-left" size="24" /></button>
      <h1>极速通道</h1>
      <button type="button" class="recharge__records" @click="router.push('/recharge-records')">充值记录</button>
    </header>

    <main class="recharge__body">
      <section class="recharge__qr-card" aria-label="USDT充值二维码" :aria-busy="loading">
        <div v-if="loading" class="recharge__state" role="status"><van-loading color="var(--c-brand)" /><p>正在获取充值地址…</p></div>
        <div v-else-if="failed" class="recharge__state" role="alert"><van-icon name="warning-o" size="36" /><p>充值信息加载失败</p><button type="button" @click="loadConfig">重新加载</button></div>
        <div v-else-if="!available" class="recharge__state"><van-icon name="clock-o" size="36" /><p>充值通道暂未开放</p><span>请稍后重试或联系客服</span><button type="button" @click="loadConfig">刷新状态</button></div>
        <template v-else>
          <div v-if="qr" class="recharge__qr">
            <img :src="qr" alt="USDT TRC-20充值地址二维码" width="210" height="210" />
            <span class="recharge__coin" aria-hidden="true"><svg viewBox="0 0 40 40"><path d="M11 9h18v5h-7v4c7 .2 11 1.5 11 3s-4 2.8-11 3v8h-4v-8c-7-.2-11-1.5-11-3s4-2.8 11-3v-4h-7z" fill="currentColor" /><ellipse cx="20" cy="21" rx="13" ry="2.5" fill="none" stroke="var(--text-primary)" stroke-width="1" /></svg></span>
          </div>
          <div v-else class="recharge__state"><p>{{ qrFailed ? '二维码生成失败，请复制下方地址' : '正在生成二维码…' }}</p><button v-if="qrFailed" type="button" @click="loadConfig">重试</button></div>
          <p class="recharge__verify">请仔细核对地址</p>
        </template>
      </section>

      <section class="recharge__address-card">
        <h2>充值地址(仅支持trc-20)</h2>
        <div class="recharge__address-box">
          <span class="recharge__address">{{ config?.address || (loading ? '正在加载充值地址…' : '暂无可用充值地址') }}</span>
          <button type="button" aria-label="复制充值地址" :disabled="!available || loading" @click="copyAddress"><svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" aria-hidden="true"><rect x="8" y="8" width="12" height="13" rx="2" /><path d="M15 4H6a2 2 0 0 0-2 2v10" /></svg></button>
        </div>
      </section>

      <aside class="recharge__warning" aria-label="充值须知">
        <ul>
          <li>请仔细核查转账地址，非正确币种接收地址，会导致资产无法找回</li>
          <li>最小充币额 1 USDT</li>
          <li>每笔到账金额截取至小数点后两位，不四舍五入；不足 0.01 USDT 的尾数直接舍弃，不累计。</li>
          <li>您只能向此地址充入对应网络的币种，充入其他资产将无法找回</li>
        </ul>
      </aside>
    </main>

    <footer class="recharge__footer">
      <van-button block round :disabled="loading || !available" @click="openForm">我已完成转账</van-button>
    </footer>

    <VanDialog v-model:show="showForm" title="提交转账信息" :show-confirm-button="false" :close-on-click-overlay="!submitting">
      <form class="recharge-form" @submit.prevent="submit">
        <p class="recharge-form__hint">请在完成 USDT/TRC-20 转账后提交，核实到账后计入余额。</p>
        <label for="recharge-amount">转账金额（USDT）</label>
        <input id="recharge-amount" v-model="amount" type="text" inputmode="decimal" autocomplete="off" placeholder="至少 1 USDT，最多6位小数" maxlength="32" :disabled="submitting" />
        <label for="recharge-hash">交易哈希（TxID）</label>
        <textarea id="recharge-hash" v-model="txHash" rows="3" autocomplete="off" autocapitalize="off" spellcheck="false" placeholder="从转出钱包复制64位交易哈希" maxlength="66" :disabled="submitting"></textarea>
        <p v-if="preview" class="recharge-form__hint">按申报金额预计入账 {{ preview }} USDT，以核实到账金额为准。</p>
        <p v-if="formError" class="recharge-form__error" role="alert">{{ formError }}</p>
        <div class="recharge-form__actions"><van-button :disabled="submitting" @click="showForm = false">取消</van-button><van-button type="primary" native-type="submit" :loading="submitting">提交核实</van-button></div>
      </form>
    </VanDialog>
  </div>
</template>

<style scoped>
.recharge { --bg-page: #0e0e0e; --bg-card: #1b1927; --c-surface-2: #252134; --c-danger: #ff434d; --c-brand: #6840f5; display: flex; flex-direction: column; height: 100%; background: var(--bg-page); }
.recharge__header { position: relative; display: flex; align-items: center; justify-content: space-between; flex-shrink: 0; height: calc(84px + var(--safe-top)); padding: var(--safe-top) 16px 0; }
.recharge__header h1 { position: absolute; left: 50%; transform: translateX(-50%); margin: 0; color: var(--text-primary); font-size: 18px; font-weight: 700; white-space: nowrap; }
.recharge__header button { background: none; border: 0; padding: 8px 0; }
.recharge__back { display: grid; place-items: center; color: var(--text-primary); min-width: 28px; min-height: 44px; }
.recharge__records { color: var(--c-brand); text-decoration: underline; text-underline-offset: 4px; font-size: 16px; font-weight: 700; }
.recharge__body { flex: 1; min-height: 0; overflow-y: auto; padding: 16px 15px 24px; }
.recharge__qr-card, .recharge__address-card { border-radius: 16px; background: var(--bg-card); }
.recharge__qr-card { display: flex; flex-direction: column; align-items: center; justify-content: center; min-height: 291px; padding: 21px 16px 18px; }
.recharge__qr { width: 210px; height: 210px; position: relative; overflow: hidden; border-radius: 15px; background: var(--text-primary); }
.recharge__qr img { width: 100%; height: 100%; }
.recharge__coin { position: absolute; left: 50%; top: 50%; transform: translate(-50%, -50%); display: grid; place-items: center; width: 50px; height: 50px; border: 5px solid var(--text-primary); border-radius: 50%; background: #279e7d; color: var(--text-primary); }
.recharge__coin svg { width: 34px; height: 34px; }
.recharge__verify { margin-top: 19px; color: var(--text-primary); font-size: 14px; font-weight: 700; }
.recharge__state { display: flex; flex-direction: column; align-items: center; justify-content: center; gap: 12px; min-height: 210px; color: var(--text-secondary); text-align: center; }
.recharge__state span { color: var(--text-muted); font-size: 12px; }
.recharge__state button { padding: 5px 16px; border: 1px solid var(--c-brand); border-radius: 20px; background: none; color: var(--c-brand); }
.recharge__address-card { margin-top: 20px; padding: 20px 15px; }
.recharge__address-card h2 { font-size: 16px; color: var(--text-primary); font-weight: 700; margin-bottom: 19px; }
.recharge__address-box { display: flex; align-items: center; gap: 10px; min-height: 73px; padding: 12px 15px; border-radius: 16px; background: var(--c-surface-2); }
.recharge__address { flex: 1; min-width: 0; overflow-wrap: anywhere; user-select: all; color: var(--text-secondary); font-size: 14px; line-height: 1.6; }
.recharge__address-box button { display: grid; place-items: center; flex: 0 0 28px; min-height: 36px; padding: 0; border: none; background: none; color: var(--text-muted); }
.recharge__address-box button:disabled { opacity: .4; cursor: default; }
.recharge__warning { margin-top: 20px; padding: 16px 18px 16px 35px; border-radius: 16px; background: #3b1018; color: var(--c-danger); font-size: 14px; line-height: 1.45; }
.recharge__warning ul { list-style: disc; }
.recharge__warning li { padding-left: 2px; }
.recharge__footer { flex-shrink: 0; padding: 10px 15px calc(10px + var(--safe-bottom)); }
.recharge__footer .van-button { height: 55px; border: none; background: var(--c-brand); color: var(--text-primary); font-size: 17px; font-weight: 700; }
.recharge-form { padding: 18px 22px 22px; }
.recharge-form__hint { margin-bottom: 16px; font-size: 12px; line-height: 1.7; color: var(--text-secondary); }
.recharge-form label { display: block; margin: 12px 0 8px; font-size: 13px; }
.recharge-form input, .recharge-form textarea { width: 100%; padding: 11px 12px; border: 1px solid var(--border-card); border-radius: 9px; background: var(--c-surface-2); color: var(--text-primary); font: inherit; font-size: 14px; resize: none; }
.recharge-form input:focus, .recharge-form textarea:focus { outline: 1px solid var(--c-brand); }
.recharge-form__error { margin: 10px 0; color: var(--c-danger); font-size: 12px; }
.recharge-form__actions { display: grid; grid-template-columns: 1fr 1fr; gap: 12px; margin-top: 20px; }
.recharge-form__actions .van-button { border-radius: 22px; }
@media (max-width: 360px) { .recharge__records { font-size: 14px; } .recharge__address { font-size: 12px; } }
</style>
