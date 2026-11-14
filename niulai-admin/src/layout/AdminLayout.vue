<script setup>
import { computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessageBox } from 'element-plus'
import { Goods, Shop, SwitchButton, ArrowDown } from '@element-plus/icons-vue'

const route = useRoute()
const router = useRouter()

const activeMenu = computed(() => route.path)
const title = computed(() => route.meta.title || '控制台')
const menus = [{ path: '/products', title: '商品管理', icon: Goods }]

function logout() {
  ElMessageBox.confirm('确认退出登录？', '提示', { type: 'warning' })
    .then(() => {
      localStorage.removeItem('admin_token')
      router.replace({ name: 'Login' })
    })
    .catch(() => {})
}

function onCommand(cmd) {
  if (cmd === 'logout') logout()
}
</script>

<template>
  <el-container class="admin">
    <el-aside class="sidebar" width="220px">
      <div class="logo">
        <el-icon :size="22"><Shop /></el-icon>
        <span>牛来 · 运营后台</span>
      </div>
      <el-menu
        class="side-menu"
        :default-active="activeMenu"
        router
        background-color="#0f172a"
        text-color="#94a3b8"
        active-text-color="#ffffff"
      >
        <el-menu-item v-for="m in menus" :key="m.path" :index="m.path">
          <el-icon><component :is="m.icon" /></el-icon>
          <span>{{ m.title }}</span>
        </el-menu-item>
      </el-menu>
    </el-aside>

    <el-container class="main-wrap">
      <el-header class="header">
        <div class="crumb">首页 <span class="sep">/</span> <strong>{{ title }}</strong></div>
        <el-dropdown @command="onCommand">
          <div class="user">
            <el-avatar :size="30" class="avatar">运</el-avatar>
            <span>运营管理员</span>
            <el-icon><ArrowDown /></el-icon>
          </div>
          <template #dropdown>
            <el-dropdown-menu>
              <el-dropdown-item command="logout" :icon="SwitchButton">退出登录</el-dropdown-item>
            </el-dropdown-menu>
          </template>
        </el-dropdown>
      </el-header>

      <el-main class="content">
        <router-view v-slot="{ Component }">
          <transition name="fade" mode="out-in"><component :is="Component" /></transition>
        </router-view>
      </el-main>
    </el-container>
  </el-container>
</template>

<style scoped>
.admin {
  height: 100vh;
}
.sidebar {
  display: flex;
  flex-direction: column;
  background: linear-gradient(180deg, #111c33, #0b1220);
}
.logo {
  height: 64px;
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 0 20px;
  color: #fff;
  font-size: 16px;
  font-weight: 600;
  border-bottom: 1px solid rgba(255, 255, 255, 0.06);
}
.logo .el-icon {
  color: #409eff;
}
.side-menu {
  flex: 1;
  border-right: none;
  padding: 8px;
}
.side-menu .el-menu-item {
  height: 46px;
  border-radius: 8px;
  margin-bottom: 4px;
}
.side-menu .el-menu-item.is-active {
  background: linear-gradient(90deg, #2563eb, #3b82f6) !important;
  color: #fff !important;
}
.main-wrap {
  background: var(--admin-bg);
}
.header {
  height: 64px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  background: #fff;
  box-shadow: 0 1px 4px rgba(0, 0, 0, 0.06);
  z-index: 1;
}
.crumb {
  color: #6b7280;
  font-size: 14px;
}
.crumb .sep {
  margin: 0 8px;
  color: #d1d5db;
}
.crumb strong {
  color: #111827;
}
.user {
  display: flex;
  align-items: center;
  gap: 8px;
  cursor: pointer;
  font-size: 14px;
  color: #374151;
}
.avatar {
  background: #409eff;
}
.content {
  padding: 20px;
  overflow: auto;
}
.fade-enter-active,
.fade-leave-active {
  transition: opacity 0.18s ease;
}
.fade-enter-from,
.fade-leave-to {
  opacity: 0;
}
</style>
