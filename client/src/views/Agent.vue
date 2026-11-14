<script setup>
import { ref, onMounted } from 'vue'
import { getAgents } from '@/api'

const list = ref([])
const total = ref(0)
const loading = ref(false)

const fmtTime = (v) => (v ? String(v).replace('T', ' ').slice(0, 10) : '')
const levelText = (level) => (level ? `Lv.${level}` : 'Lv.1')

const load = async () => {
  loading.value = true
  try {
    const data = await getAgents({ page: 1, page_size: 50 })
    list.value = Array.isArray(data?.list) ? data.list : []
    total.value = data?.total ?? 0
  } catch {
    list.value = []
  } finally {
    loading.value = false
  }
}
onMounted(load)
</script>

<template>
  <div class="agent page--scroll">
    <header class="agent__head">
      <h2>代理</h2>
      <span class="agent__count">共 {{ total }} 位</span>
    </header>

    <section class="agent__banner">
      <p class="agent__eyebrow">分销代理概览</p>
      <p class="agent__desc">数据来自服务端代理关系表，实时同步上/下级变动</p>
    </section>

    <div v-if="loading" class="agent__tip">
      <van-loading size="20" /> 加载中…
    </div>

    <ul v-else class="agent__list">
      <li v-for="item in list" :key="item.id" class="agent__item">
        <div class="agent__level">{{ levelText(item.level) }}</div>
        <div class="agent__info">
          <p class="agent__code">{{ item.invite_code || '无邀请码' }}</p>
          <p class="agent__meta">用户 #{{ item.user_id }} · {{ fmtTime(item.created_at) }}</p>
        </div>
      </li>
    </ul>

    <p v-if="!loading && !list.length" class="agent__tip">暂无代理数据</p>
  </div>
</template>

<style scoped>
.agent__head { display: flex; align-items: baseline; justify-content: space-between; margin-bottom: 16px; }
h2 { font-size: 20px; color: var(--text-primary); }
.agent__count { font-size: 12px; color: var(--text-muted); }

.agent__banner {
  padding: 16px; margin-bottom: 16px; border-radius: var(--radius-card);
  background: var(--gradient-banner);
  box-shadow: 0 10px 24px rgba(90, 80, 255, .24);
}
.agent__eyebrow { font-size: 14px; font-weight: 700; color: #fff; }
.agent__desc { margin-top: 6px; font-size: 11px; color: rgba(255, 255, 255, .8); }

.agent__tip {
  display: flex; align-items: center; justify-content: center; gap: 8px;
  padding: 40px 0; font-size: 12px; color: var(--text-muted);
}
.agent__list { display: flex; flex-direction: column; gap: 10px; }
.agent__item {
  display: flex; align-items: center; gap: 12px;
  padding: 14px; border: 1px solid var(--border-card);
  border-radius: var(--radius-card); background: var(--bg-card);
}
.agent__level {
  flex: 0 0 auto; min-width: 44px; padding: 6px 8px; text-align: center;
  border-radius: 10px; background: var(--c-fill);
  color: var(--c-brand); font-size: 12px; font-weight: 800;
}
.agent__info { min-width: 0; }
.agent__code { font-size: 13px; font-weight: 700; color: var(--text-primary); }
.agent__meta { margin-top: 3px; font-size: 11px; color: var(--text-muted); }
</style>
