<script setup>
import { ref, watch } from 'vue'
import { useRouter } from 'vue-router'
import { fetchCategories } from '@/api/category'

const props = defineProps({
  modelValue: { type: String, default: '' },
  active: { type: Boolean, default: true },
})
const emit = defineEmits(['update:modelValue'])
const categories = ref([])
const loading = ref(false)
const failed = ref(false)
const manageURL = useRouter().resolve('/categories').href
let requestId = 0

async function refresh() {
  if (!props.active) return
  const id = ++requestId
  loading.value = true
  failed.value = false
  try {
    const list = await fetchCategories()
    if (id === requestId) categories.value = Array.isArray(list) ? list : []
  } catch {
    if (id === requestId) failed.value = true
  } finally {
    if (id === requestId) loading.value = false
  }
}

watch(() => props.active, (open) => { if (open) refresh() }, { immediate: true })
</script>

<template>
  <div class="category-field">
    <el-select
      :model-value="modelValue"
      :loading="loading"
      placeholder="请选择商品分类"
      filterable
      class="category-select"
      @update:model-value="emit('update:modelValue', $event)"
      @visible-change="(open) => { if (open) refresh() }"
    >
      <el-option v-for="c in categories" :key="c.code" :label="c.name" :value="c.code" />
      <template #empty>
        <div class="empty">{{ loading ? '加载中…' : (failed ? '加载失败，请重试' : '暂无分类，请先新增分类') }}</div>
      </template>
    </el-select>
    <el-link :href="manageURL" target="_blank" rel="noopener" type="primary" :underline="false">管理分类</el-link>
    <el-button v-if="failed" link type="primary" @click="refresh">重试</el-button>
  </div>
</template>

<style scoped>
.category-field { display: flex; align-items: center; gap: 12px; width: 100%; }
.category-select { flex: 1; min-width: 0; }
.category-field :deep(.el-link) { white-space: nowrap; }
.empty { padding: 14px; text-align: center; color: #909399; }
</style>
