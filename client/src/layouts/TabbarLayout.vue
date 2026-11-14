<script setup>
import AppTabbar from '@/components/AppTabbar.vue'
</script>

<template>
  <div class="layout">
    <router-view v-slot="{ Component }">
      <keep-alive :max="4"><component :is="Component" /></keep-alive>
    </router-view>
    <AppTabbar />
  </div>
</template>

<style scoped>
.layout { height: 100%; position: relative; }
</style>
