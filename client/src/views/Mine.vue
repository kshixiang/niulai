<script setup>
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { showToast } from 'vant'
import { useUserStore } from '@/stores/user'
import { getWallet, updateProfile } from '@/api'

const router = useRouter()
const user = useUserStore()

/* ---------- 资金账户 ---------- */
const wallet = ref({ balance: '0.00', frozen_balance: '0.00' })
const walletLoading = ref(false)

const fmtMoney = (v) => {
  const n = Number(v)
  return Number.isFinite(n) ? n.toFixed(2) : '0.00'
}
// 我的余额 = 可用余额 + 冻结余额
const totalText = computed(() =>
  (Number(wallet.value.balance) + Number(wallet.value.frozen_balance)).toFixed(2),
)

const loadWallet = async () => {
  if (!user.token) return
  walletLoading.value = true
  try {
    const data = await getWallet()
    wallet.value = {
      balance: fmtMoney(data?.balance),
      frozen_balance: fmtMoney(data?.frozen_balance),
    }
  } catch {
    // 拦截器已提示；请求失败时保留上次展示值，不把余额误清零
  } finally {
    walletLoading.value = false
  }
}

/* ---------- 用户资料 ---------- */
// 凭 token 恢复资料，同时校验会话是否仍有效（全站唯一的会话校验点）
const loadProfile = async () => {
  if (!user.token) return
  try {
    await user.fetchProfile()
  } catch {
    // 拦截器统一提示并清理过期 token
  }
}

const nickname = computed(() => user.userInfo?.nickname || '未登录')
// 由用户ID派生稳定的展示 UID，无需后端额外字段
const uidText = computed(() => {
  const id = Number(user.userInfo?.id)
  return id > 0 ? `U${id.toString(36).toUpperCase().padStart(6, '0')}` : '—'
})
const statusText = computed(() => (user.userInfo?.status === 0 ? '已禁用' : '正常'))
const fmtDate = (v) => (v ? String(v).replace('T', ' ').slice(0, 10) : '—')

onMounted(() => { loadWallet(); loadProfile() })

/* ---------- 复制 UID ---------- */
const copyUid = async () => {
  if (uidText.value === '—') return
  try {
    await navigator.clipboard.writeText(uidText.value)
    showToast('UID 已复制')
  } catch {
    showToast('复制失败，请手动复制')
  }
}

/* ---------- 尚未开通的功能统一占位 ---------- */
const comingSoon = (name) => showToast(`${name}功能开发中，敬请期待`)

/* ---------- 弹层 ---------- */
const dialog = ref('') // '' | 'nickname' | 'settings' | 'logout'
const nicknameInput = ref('')
const saving = ref(false)

const openNickname = () => {
  nicknameInput.value = user.userInfo?.nickname || ''
  dialog.value = 'nickname'
}
const saveNickname = async () => {
  const v = nicknameInput.value.trim()
  if (!v) { showToast('昵称不能为空'); return }
  if (saving.value) return
  saving.value = true
  try {
    const data = await updateProfile({ nickname: v })
    user.setUserInfo(data)
    dialog.value = ''
    showToast('昵称已更新')
  } catch {
    // 拦截器已提示
  } finally {
    saving.value = false
  }
}

/* ---------- 退出登录 ---------- */
const doLogout = () => {
  user.logout()
  dialog.value = ''
  router.replace('/login')
}

/* ---------- 菜单配置 ---------- */
const tradeMenus = [
  { label: '收支方式', icon: 'exchange' },
  { label: '黑名单', icon: 'friends-o' },
]
const otherMenus = [
  { label: '查询余额', icon: 'search' },
  { label: '异常充值号码反馈', icon: 'warning-o' },
]
</script>

<template>
  <div class="mine page--scroll">
    <!-- 顶部紫色装饰背景 -->
    <div class="mine__bg" aria-hidden="true"></div>

    <!-- 用户信息栏 -->
    <header class="mine__header">
      <div class="mine__identity">
        <div class="mine__logo">牛</div>
        <div class="mine__meta">
          <div class="mine__name-row">
            <span class="mine__name">{{ nickname }}</span>
            <van-icon name="edit" class="mine__edit" @click="openNickname" />
          </div>
          <div class="mine__uid-row">
            <span class="mine__uid">UID:{{ uidText }}</span>
            <van-icon name="orders-o" class="mine__copy" @click="copyUid" />
          </div>
        </div>
      </div>
      <div class="mine__actions">
        <button type="button" class="mine__icon-btn" aria-label="设置" @click="dialog = 'settings'">
          <van-icon name="setting-o" size="18" />
        </button>
        <button type="button" class="mine__icon-btn" aria-label="客服" @click="comingSoon('在线客服')">
          <van-icon name="service-o" size="18" />
        </button>
      </div>
    </header>

    <!-- 余额卡片 -->
    <section class="wallet">
      <div class="wallet__top">
        <div class="wallet__label">
          我的余额
          <van-icon
            name="replay"
            class="wallet__refresh"
            :class="{ 'is-spinning': walletLoading }"
            @click="loadWallet"
          />
        </div>
        <button type="button" class="wallet__detail" @click="comingSoon('账户明细')">
          账户明细 <van-icon name="arrow" />
        </button>
      </div>

      <div class="wallet__amount">
        <span class="wallet__num">{{ totalText }}</span>
        <span class="wallet__unit">USDT</span>
      </div>

      <div class="wallet__stats">
        <div class="wallet__stat">
          <b>{{ wallet.balance }}</b>
          <span>可用余额</span>
        </div>
        <div class="wallet__stat wallet__stat--frozen">
          <b>{{ wallet.frozen_balance }}</b>
          <span>冻结</span>
        </div>
      </div>

      <div class="wallet__watermark" aria-hidden="true"><van-icon name="gold-coin-o" size="86" /></div>

      <div class="wallet__btns">
        <button type="button" class="wallet__btn wallet__btn--blue" @click="router.push('/recharge')">
          <van-icon name="arrow-down" /> 充值
        </button>
        <button type="button" class="wallet__btn wallet__btn--blue" @click="comingSoon('提现')">
          <van-icon name="arrow-up" /> 提现
        </button>
        <button type="button" class="wallet__btn wallet__btn--purple" @click="comingSoon('转账')">
          <van-icon name="exchange" /> 转账
        </button>
      </div>
    </section>

    <!-- C2C 快捷交易 -->
    <section class="c2c" @click="comingSoon('C2C 交易')">
      <div class="c2c__art" aria-hidden="true">
        <span class="c2c__glow"></span>
        <span class="c2c__coin c2c__coin--1">T</span>
        <span class="c2c__coin c2c__coin--2">T</span>
        <span class="c2c__shield"><van-icon name="fire-o" size="26" /></span>
      </div>
      <div class="c2c__content">
        <h3 class="c2c__title">C2C快捷交易</h3>
        <p class="c2c__desc">自由交易，快速到账，安全可靠</p>
        <span class="c2c__go">前往交易 <van-icon name="arrow" /></span>
      </div>
    </section>

    <!-- 交易配置 -->
    <h3 class="mine__section-title">交易配置</h3>
    <section class="menu">
      <button
        v-for="(m, i) in tradeMenus"
        :key="m.label"
        type="button"
        class="menu__row"
        :class="{ 'menu__row--divider': i < tradeMenus.length - 1 }"
        @click="comingSoon(m.label)"
      >
        <span class="menu__icon"><van-icon :name="m.icon" size="18" /></span>
        <span class="menu__label">{{ m.label }}</span>
        <van-icon name="arrow" class="menu__arrow" />
      </button>
    </section>

    <!-- 其他 -->
    <h3 class="mine__section-title">其他</h3>
    <section class="menu">
      <button
        v-for="(m, i) in otherMenus"
        :key="m.label"
        type="button"
        class="menu__row"
        :class="{ 'menu__row--divider': i < otherMenus.length - 1 }"
        @click="comingSoon(m.label)"
      >
        <span class="menu__icon"><van-icon :name="m.icon" size="18" /></span>
        <span class="menu__label">{{ m.label }}</span>
        <van-icon name="arrow" class="menu__arrow" />
      </button>
    </section>

    <button type="button" class="mine__logout" @click="dialog = 'logout'">退出登录</button>

    <!-- 弹层统一挂到 body，避免被滚动容器裁剪 -->
    <Teleport to="body">
      <!-- 修改昵称 -->
      <div v-if="dialog === 'nickname'" class="mask" @click.self="dialog = ''">
        <div class="dialog">
          <h3 class="dialog__title">修改昵称</h3>
          <input
            v-model="nicknameInput"
            class="dialog__input"
            type="text"
            maxlength="64"
            placeholder="请输入新昵称"
            @keyup.enter="saveNickname"
          />
          <div class="dialog__actions">
            <button type="button" class="dialog__btn" @click="dialog = ''">取消</button>
            <button
              type="button"
              class="dialog__btn dialog__btn--primary"
              :disabled="saving"
              @click="saveNickname"
            >{{ saving ? '保存中…' : '保存' }}</button>
          </div>
        </div>
      </div>

      <!-- 账号信息 -->
      <div v-if="dialog === 'settings'" class="mask" @click.self="dialog = ''">
        <div class="dialog">
          <h3 class="dialog__title">账号信息</h3>
          <div class="dialog__rows">
            <div class="dialog__row"><span>UID</span><b>{{ uidText }}</b></div>
            <div class="dialog__row"><span>登录账号</span><b>{{ user.userInfo?.phone || '—' }}</b></div>
            <div class="dialog__row"><span>账号状态</span><b>{{ statusText }}</b></div>
            <div class="dialog__row"><span>注册时间</span><b>{{ fmtDate(user.userInfo?.created_at) }}</b></div>
          </div>
          <button type="button" class="dialog__btn dialog__btn--primary dialog__btn--full" @click="dialog = ''">关闭</button>
        </div>
      </div>

      <!-- 退出登录确认 -->
      <div v-if="dialog === 'logout'" class="mask" @click.self="dialog = ''">
        <div class="dialog">
          <h3 class="dialog__title">退出登录</h3>
          <p class="dialog__msg">确定要退出当前账号吗？</p>
          <div class="dialog__actions">
            <button type="button" class="dialog__btn" @click="dialog = ''">取消</button>
            <button type="button" class="dialog__btn dialog__btn--danger" @click="doLogout">退出</button>
          </div>
        </div>
      </div>
    </Teleport>
  </div>
</template>

<style scoped>
.mine { position: relative; background: var(--bg-page); }

/* ---------- 顶部装饰 ---------- */
.mine__bg {
  position: absolute; top: 0; left: 0; right: 0; height: 240px;
  pointer-events: none;
  background:
    radial-gradient(58% 75% at 88% 4%, rgba(123, 92, 255, .42), transparent 70%),
    radial-gradient(70% 85% at 2% 0%, rgba(76, 141, 255, .26), transparent 70%),
    linear-gradient(180deg, #2b1f60 0%, var(--bg-page) 100%);
}

/* ---------- 用户信息栏 ---------- */
.mine__header {
  position: relative; z-index: 1;
  display: flex; align-items: center; justify-content: space-between;
  gap: 12px; margin-bottom: 20px;
}
.mine__identity { display: flex; align-items: center; gap: 12px; min-width: 0; }
.mine__logo {
  flex: 0 0 auto; width: 48px; height: 48px;
  display: grid; place-items: center; border-radius: 14px;
  background: linear-gradient(135deg, #5b8cff, #7b5cff);
  color: #fff; font-size: 22px; font-weight: 800;
  box-shadow: 0 6px 18px rgba(91, 108, 255, .45);
}
.mine__meta { min-width: 0; }
.mine__name-row { display: flex; align-items: center; gap: 8px; }
.mine__name {
  max-width: 150px; font-size: 17px; font-weight: 700; color: #fff;
  overflow: hidden; text-overflow: ellipsis; white-space: nowrap;
}
.mine__edit { color: rgba(255, 255, 255, .6); font-size: 14px; }
.mine__uid-row { display: flex; align-items: center; gap: 6px; margin-top: 5px; }
.mine__uid { font-size: 11px; color: rgba(255, 255, 255, .6); letter-spacing: .3px; }
.mine__copy { color: rgba(255, 255, 255, .6); font-size: 12px; }
.mine__actions { display: flex; gap: 10px; flex: 0 0 auto; }
.mine__icon-btn {
  width: 34px; height: 34px; display: grid; place-items: center;
  border: none; border-radius: 50%;
  background: rgba(255, 255, 255, .12); color: #fff;
}
.mine__icon-btn:active { background: rgba(255, 255, 255, .22); }

/* ---------- 余额卡片 ---------- */
.wallet {
  position: relative; z-index: 1; overflow: hidden;
  padding: 18px 18px 16px;
  border: 1px solid var(--border-card); border-radius: var(--radius-card);
  background:
    radial-gradient(120% 120% at 100% 0%, rgba(123, 92, 255, .18), transparent 60%),
    var(--bg-card);
  box-shadow: 0 12px 28px rgba(0, 0, 0, .22);
}
.wallet > *:not(.wallet__watermark) { position: relative; z-index: 1; }
.wallet__watermark {
  position: absolute; right: 12px; bottom: 66px; z-index: 0;
  color: rgba(255, 255, 255, .05); pointer-events: none;
}
.wallet__top { display: flex; align-items: center; justify-content: space-between; }
.wallet__label { display: flex; align-items: center; gap: 6px; font-size: 13px; color: var(--text-secondary); }
.wallet__refresh { color: var(--text-muted); font-size: 13px; }
.wallet__refresh.is-spinning { animation: spin .8s linear infinite; }
@keyframes spin { to { transform: rotate(360deg); } }
.wallet__detail {
  display: flex; align-items: center; gap: 2px; padding: 0;
  background: none; border: none; font-size: 12px; color: var(--text-secondary);
}
.wallet__amount { display: flex; align-items: baseline; gap: 6px; margin-top: 14px; }
.wallet__num { font-size: 34px; font-weight: 800; letter-spacing: .5px; line-height: 1; color: var(--text-primary); }
.wallet__unit { font-size: 13px; font-weight: 700; color: var(--c-brand); }
.wallet__stats { display: flex; align-items: center; gap: 24px; margin-top: 18px; }
.wallet__stat { display: flex; flex-direction: column; gap: 3px; }
.wallet__stat b { font-size: 16px; font-weight: 800; color: var(--text-primary); }
.wallet__stat span { font-size: 11px; color: var(--text-muted); }
.wallet__stat--frozen { padding-left: 24px; border-left: 1px solid var(--border-card); }
.wallet__btns { display: grid; grid-template-columns: repeat(3, 1fr); gap: 10px; margin-top: 20px; }
.wallet__btn {
  display: flex; align-items: center; justify-content: center; gap: 4px;
  height: 40px; border: none; border-radius: 10px;
  color: #fff; font-size: 13px; font-weight: 700;
}
.wallet__btn:active { transform: scale(.97); }
.wallet__btn--blue { background: linear-gradient(135deg, #4c8dff, #3a6cf0); box-shadow: 0 6px 14px rgba(76, 141, 255, .3); }
.wallet__btn--purple { background: linear-gradient(135deg, #8a5cff, #6a4bff); box-shadow: 0 6px 14px rgba(123, 92, 255, .35); }

/* ---------- C2C 快捷交易 ---------- */
.c2c {
  position: relative; z-index: 1; overflow: hidden;
  display: flex; align-items: center; gap: 12px;
  margin-top: 16px; padding: 18px 16px;
  border: 1px solid var(--border-card); border-radius: var(--radius-card);
  background: linear-gradient(120deg, #2a2150, #1d1b3a 60%, #241c4a);
}
.c2c:active { transform: scale(.99); }
.c2c__art { position: relative; flex: 0 0 92px; width: 92px; height: 84px; }
.c2c__glow {
  position: absolute; inset: 6px; border-radius: 50%;
  background: radial-gradient(circle, rgba(123, 92, 255, .4), transparent 70%);
}
.c2c__coin {
  position: absolute; width: 22px; height: 22px; display: grid; place-items: center;
  border-radius: 50%; font-size: 11px; font-weight: 800; color: #0d3b2e;
  background: var(--c-success); box-shadow: 0 3px 8px rgba(0, 0, 0, .3);
}
.c2c__coin--1 { top: 4px; left: 2px; }
.c2c__coin--2 { bottom: 6px; right: 0; }
.c2c__shield {
  position: absolute; top: 50%; left: 50%;
  transform: translate(-50%, -50%) rotate(-8deg);
  width: 52px; height: 58px; display: grid; place-items: center;
  border-radius: 16px 16px 20px 20px;
  background: linear-gradient(160deg, #8a5cff, #5a3fd6);
  color: #7bf0c0; border: 1px solid rgba(255, 255, 255, .18);
  box-shadow: 0 8px 18px rgba(90, 63, 214, .55);
}
.c2c__content { flex: 1; min-width: 0; }
.c2c__title { font-size: 17px; font-weight: 800; color: #fff; }
.c2c__desc { margin-top: 6px; font-size: 11px; color: rgba(255, 255, 255, .65); }
.c2c__go {
  display: inline-flex; align-items: center; gap: 4px; margin-top: 12px;
  padding: 6px 14px; border-radius: var(--radius-pill);
  background: rgba(255, 255, 255, .1); border: 1px solid rgba(255, 255, 255, .14);
  font-size: 11px; color: #fff;
}

/* ---------- 分组标题 ---------- */
.mine__section-title {
  position: relative; z-index: 1;
  margin: 22px 0 12px; padding-left: 12px;
  font-size: 15px; font-weight: 700; color: var(--text-primary);
}
.mine__section-title::before {
  content: ''; position: absolute; left: 0; top: 50%; transform: translateY(-50%);
  width: 3px; height: 15px; border-radius: 2px; background: var(--gradient-brand);
}
.mine__section-title::after {
  content: ''; position: absolute; left: 0; top: calc(50% - 9px);
  width: 3px; height: 4px; border-radius: 2px; background: var(--c-brand-2);
}

/* ---------- 菜单卡片 ---------- */
.menu {
  position: relative; z-index: 1; overflow: hidden;
  border: 1px solid var(--border-card); border-radius: var(--radius-card);
  background: var(--bg-card);
}
.menu__row {
  display: flex; align-items: center; gap: 12px; width: 100%;
  padding: 16px 14px; background: none; border: none; text-align: left;
}
.menu__row:active { background: var(--c-fill); }
.menu__row--divider { border-bottom: 1px solid var(--border-card); }
.menu__icon {
  flex: 0 0 auto; width: 30px; height: 30px; display: grid; place-items: center;
  border-radius: 9px; background: var(--c-fill); color: var(--c-brand-2);
}
.menu__label { flex: 1; font-size: 14px; font-weight: 600; color: var(--text-primary); }
.menu__arrow { color: var(--text-muted); font-size: 14px; }

.mine__logout {
  position: relative; z-index: 1;
  display: block; width: 100%; margin-top: 24px; padding: 14px;
  border: 1px solid var(--border-card); border-radius: var(--radius-card);
  background: var(--bg-card); color: var(--text-secondary); font-size: 14px;
}
.mine__logout:active { color: var(--c-danger); border-color: var(--c-danger); }

/* ---------- 弹层 ---------- */
.mask {
  position: fixed; inset: 0; z-index: 2000;
  display: grid; place-items: center; padding: 24px;
  background: rgba(0, 0, 0, .6);
  -webkit-backdrop-filter: blur(2px); backdrop-filter: blur(2px);
}
.dialog {
  width: 100%; max-width: 320px; padding: 20px;
  border: 1px solid var(--border-card); border-radius: 18px;
  background: var(--c-surface-2);
  box-shadow: 0 18px 50px rgba(0, 0, 0, .5);
}
.dialog__title { font-size: 16px; font-weight: 700; color: var(--text-primary); text-align: center; margin-bottom: 18px; }
.dialog__msg { font-size: 13px; color: var(--text-secondary); text-align: center; }
.dialog__input {
  width: 100%; height: 40px; padding: 0 12px;
  border: 1px solid var(--border-card); border-radius: 10px;
  background: var(--bg-card); color: var(--text-primary); font-size: 14px;
  outline: none;
}
.dialog__input:focus { border-color: var(--c-brand); }
.dialog__input::placeholder { color: var(--text-muted); }
.dialog__rows { display: flex; flex-direction: column; gap: 12px; }
.dialog__row { display: flex; align-items: center; justify-content: space-between; gap: 12px; font-size: 13px; }
.dialog__row span { color: var(--text-muted); }
.dialog__row b { color: var(--text-primary); font-weight: 600; }
.dialog__actions { display: flex; gap: 10px; margin-top: 18px; }
.dialog__btn {
  flex: 1; height: 38px; border-radius: 10px;
  border: 1px solid var(--border-card); background: var(--bg-card);
  color: var(--text-secondary); font-size: 14px;
}
.dialog__btn--primary { background: var(--gradient-brand); border: none; color: #fff; font-weight: 700; }
.dialog__btn--danger { background: rgba(255, 92, 108, .14); border: none; color: var(--c-danger); font-weight: 700; }
.dialog__btn--full { width: 100%; margin-top: 18px; }
</style>
