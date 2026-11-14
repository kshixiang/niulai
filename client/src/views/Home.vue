<script setup>
import { ref, computed } from 'vue'
import { showToast } from 'vant'

/* 商品分类 Tab */
const categories = [
  { key: 'all',     label: '全部' },
  { key: 'fast',    label: '快充' },
  { key: 'life',    label: '生活缴费' },
  { key: 'finance', label: '金融财税' },
  { key: 'game',    label: '游戏充值' },
  { key: 'vip',     label: '视频会员' },
]
const activeCategory = ref('all')

/* 示例数据；接后端时换成 request.get('/products') */
const products = [
  { id: 1, name: '话费充值', desc: '三网通用 · 秒到账',   discount: '9.8折', icon: 'phone-o',       cat: 'fast',    tone: 'purple' },
  { id: 2, name: '流量包',   desc: '全国流量 · 即充即用', discount: '8.8折', icon: 'fire-o',        cat: 'fast',    tone: 'blue'   },
  { id: 3, name: '电费缴纳', desc: '国网南网 · 在线缴费', discount: '9.5折', icon: 'bulb-o',        cat: 'life',    tone: 'cyan'   },
  { id: 4, name: '水费燃气', desc: '一站式生活缴费',       discount: '9.6折', icon: 'wap-home-o',    cat: 'life',    tone: 'green'  },
  { id: 5, name: '财税代缴', desc: '企业个税 · 安全合规',  discount: '9.2折', icon: 'balance-o',     cat: 'finance', tone: 'orange' },
  { id: 6, name: '游戏点券', desc: '主流游戏 · 官方直充',  discount: '8.5折', icon: 'gift-o',        cat: 'game',    tone: 'pink'   },
  { id: 7, name: '视频会员', desc: '全网会员 · 自动发货',  discount: '7.9折', icon: 'play-circle-o', cat: 'vip',     tone: 'purple' },
  { id: 8, name: '加油卡',   desc: '两桶油 · 在线直充',    discount: '9.7折', icon: 'logistics',     cat: 'all',     tone: 'blue'   },
]

const visibleProducts = computed(() =>
  activeCategory.value === 'all'
    ? products
    : products.filter((p) => p.cat === activeCategory.value),
)

const onProduct = (item) => showToast(`即将打开「${item.name}」`)
</script>

<template>
  <div class="home page--scroll">
    <!-- 顶部标题栏 -->
    <header class="home__header">
      <div class="home__brand">
        <span class="home__logo">牛</span>
        <div>
          <p class="home__hello">Hi，欢迎回来</p>
          <h1 class="home__title">牛来充值</h1>
        </div>
      </div>
      <van-icon name="bell" size="22" class="home__bell" />
    </header>

    <!-- 渐变紫蓝 Banner -->
    <section class="banner">
      <div class="banner__content">
        <p class="banner__eyebrow">一站式数字生活服务</p>
        <h2 class="banner__title">随时随地充值<br />生活更自在</h2>
        <p class="banner__desc">话费 · 流量 · 生活缴费 · 会员权益</p>
        <van-button round size="small" class="banner__btn" @click="onProduct({ name: '立即充值' })">
          立即充值
        </van-button>
      </div>

      <!-- 右侧插图占位：接设计稿时替换为 <img class="banner__img" src="..." alt="" /> -->
      <div class="banner__art" aria-hidden="true">
        <span class="banner__glow"></span>
        <span class="banner__card banner__card--1"></span>
        <span class="banner__card banner__card--2"></span>
        <van-icon name="gift-card-o" size="48" class="banner__icon" />
      </div>
    </section>

    <!-- 商品分类 Tab（横向可滚动） -->
    <nav class="cats" aria-label="商品分类">
      <button
        v-for="c in categories"
        :key="c.key"
        type="button"
        class="cats__item"
        :class="{ 'is-active': activeCategory === c.key }"
        @click="activeCategory = c.key"
      >{{ c.label }}</button>
    </nav>

    <!-- 商品双列网格 -->
    <section class="section">
      <div class="section__head">
        <h3 class="section__title">热门推荐</h3>
        <span class="section__more">更多 <van-icon name="arrow" /></span>
      </div>

      <div class="grid">
        <article
          v-for="item in visibleProducts"
          :key="item.id"
          class="card"
          :class="`card--${item.tone}`"
          @click="onProduct(item)"
        >
          <div class="card__thumb"><van-icon :name="item.icon" size="28" /></div>
          <h4 class="card__name">{{ item.name }}</h4>
          <p class="card__desc">{{ item.desc }}</p>
          <span class="card__badge">{{ item.discount }}</span>
        </article>
      </div>

      <p v-if="!visibleProducts.length" class="grid__empty">该分类暂无商品</p>
    </section>
  </div>
</template>

<style scoped>
.home {
  background:
    radial-gradient(120% 55% at 100% 0%, rgba(123, 92, 255, .16), transparent 60%),
    var(--bg-page);
}

/* ---------- Header ---------- */
.home__header { display: flex; align-items: center; justify-content: space-between; margin-bottom: 18px; }
.home__brand { display: flex; align-items: center; gap: 10px; }
.home__logo {
  width: 38px; height: 38px; display: grid; place-items: center;
  border-radius: 12px; background: var(--gradient-brand);
  color: #fff; font-size: 18px; font-weight: 800;
  box-shadow: 0 6px 16px rgba(123, 92, 255, .35);
}
.home__hello { font-size: 11px; color: var(--text-muted); }
.home__title { font-size: 17px; font-weight: 700; color: var(--text-primary); letter-spacing: .5px; }
.home__bell { position: relative; color: var(--text-secondary); }
.home__bell::after {
  content: ''; position: absolute; top: 0; right: 0;
  width: 7px; height: 7px; border-radius: 50%; background: var(--c-danger);
}

/* ---------- Banner ---------- */
.banner {
  position: relative; overflow: hidden; min-height: 168px;
  padding: 20px; border-radius: 20px;
  display: flex; align-items: center;
  background: var(--gradient-banner);
  box-shadow: 0 12px 28px rgba(90, 80, 255, .28);
}
.banner::after {
  content: ''; position: absolute; inset: 0; pointer-events: none;
  background: radial-gradient(80% 120% at 100% 0%, rgba(255,255,255,.18), transparent 55%);
}
.banner__content { position: relative; z-index: 2; flex: 1; min-width: 0; }
.banner__eyebrow { font-size: 11px; color: rgba(255,255,255,.78); letter-spacing: 1px; margin-bottom: 8px; }
.banner__title { font-size: 22px; line-height: 1.35; font-weight: 800; color: #fff; }
.banner__desc { font-size: 11px; color: rgba(255,255,255,.8); margin-top: 8px; white-space: nowrap; }
.banner__btn {
  margin-top: 14px; height: 30px; padding: 0 16px; border: none;
  background: #fff; color: #5a4bd8; font-size: 12px; font-weight: 700;
}
.banner__btn :deep(.van-button__text) { color: #5a4bd8; }

.banner__art { position: relative; flex: 0 0 104px; width: 104px; height: 104px; display: grid; place-items: center; }
.banner__glow {
  position: absolute; inset: 6px; border-radius: 50%;
  background: rgba(255,255,255,.12); border: 1px solid rgba(255,255,255,.28);
}
.banner__card {
  position: absolute; width: 62px; height: 40px; border-radius: 10px;
  background: rgba(255,255,255,.22); border: 1px solid rgba(255,255,255,.35);
}
.banner__card--1 { transform: translate(-14px, -16px) rotate(-14deg); }
.banner__card--2 { transform: translate(12px, 16px) rotate(12deg); background: rgba(255,255,255,.10); }
.banner__icon { position: relative; color: #fff; filter: drop-shadow(0 4px 10px rgba(0,0,0,.2)); }

/* ---------- 分类 Tab ---------- */
.cats { display: flex; gap: 8px; margin: 22px 0 18px; overflow-x: auto; scrollbar-width: none; }
.cats::-webkit-scrollbar { display: none; }
.cats__item {
  flex: 0 0 auto; padding: 8px 15px; white-space: nowrap;
  border: 1px solid var(--border-card); border-radius: var(--radius-pill);
  background: var(--c-fill); color: var(--text-secondary);
  font-size: 12px; transition: all .2s ease;
}
.cats__item.is-active {
  background: var(--gradient-brand); color: #fff; border-color: transparent;
  font-weight: 700; box-shadow: 0 4px 14px rgba(123, 92, 255, .4);
}

/* ---------- 区块标题 ---------- */
.section__head { display: flex; align-items: center; justify-content: space-between; margin-bottom: 12px; }
.section__title {
  position: relative; padding-left: 12px;
  font-size: 16px; font-weight: 700; color: var(--text-primary);
}
.section__title::before {
  content: ''; position: absolute; left: 0; top: 50%; transform: translateY(-50%);
  width: 3px; height: 14px; border-radius: 2px; background: var(--gradient-brand);
}
.section__more { display: flex; align-items: center; gap: 2px; font-size: 11px; color: var(--text-muted); }

/* ---------- 双列网格 ---------- */
.grid { display: grid; grid-template-columns: repeat(2, minmax(0, 1fr)); gap: 12px; }
.grid__empty { padding: 32px 0; text-align: center; font-size: 12px; color: var(--text-muted); }

.card {
  position: relative; overflow: hidden;
  padding: 14px 12px 32px;
  border: 1px solid var(--border-card); border-radius: var(--radius-card);
  background: var(--bg-card);
  transition: transform .2s ease, border-color .2s ease;
}
.card:active { transform: scale(.97); border-color: var(--c-brand); }
.card::before {
  content: ''; position: absolute; top: -40px; right: -30px;
  width: 110px; height: 110px; border-radius: 50%;
  background: var(--card-tone, var(--c-brand)); opacity: .16; filter: blur(18px);
}
.card--purple { --card-tone: #8a5cff; }
.card--blue   { --card-tone: #3aa0ff; }
.card--cyan   { --card-tone: #35d0d8; }
.card--green  { --card-tone: #35d0a0; }
.card--orange { --card-tone: #ff9f45; }
.card--pink   { --card-tone: #ff6f9c; }

.card__thumb {
  width: 46px; height: 46px; margin-bottom: 12px;
  display: grid; place-items: center; border-radius: 14px; color: #fff;
  background: linear-gradient(135deg, var(--card-tone, var(--c-brand)), var(--c-brand-2));
  box-shadow: 0 6px 14px rgba(0, 0, 0, .25);
}
.card__name { font-size: 14px; font-weight: 700; color: var(--text-primary); }
.card__desc {
  margin-top: 4px; font-size: 11px; color: var(--text-muted);
  overflow: hidden; text-overflow: ellipsis; white-space: nowrap;
}

/* 左下角黄色折扣角标 */
.card__badge {
  position: absolute; left: 0; bottom: 0;
  padding: 4px 12px;
  border-radius: 0 12px 0 0;
  background: var(--gradient-gold);
  color: var(--c-accent-ink);
  font-size: 11px; font-weight: 800;
  box-shadow: 2px -2px 8px rgba(255, 181, 46, .18);
}
</style>
