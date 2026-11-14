<script setup>
/**
 * 颜色不写死：Vant 的 --van-tabbar-* 已在 variables.css 里映射到主题变量，
 * 下面再补一层兜底，保证「背景黑 + 未选中灰 + 选中紫高亮」稳定生效。
 */
</script>

<template>
  <van-tabbar class="app-tabbar" route fixed :border="false" safe-area-inset-bottom>
    <van-tabbar-item replace to="/home"   icon="home-o">首页</van-tabbar-item>
    <van-tabbar-item replace to="/orders" icon="orders-o" badge="2">订单</van-tabbar-item>
    <van-tabbar-item replace to="/agent"  icon="friends-o">代理</van-tabbar-item>
    <van-tabbar-item replace to="/mine"   icon="user-o">我的</van-tabbar-item>
  </van-tabbar>
</template>

<style scoped>
.app-tabbar {
  left: 0;
  right: 0;
  max-width: 480px;          /* 桌面预览时跟随 #app 居中，不横向铺满 */
  margin: 0 auto;
  background: var(--bg-tabbar);
  border-top: 1px solid var(--border-card);
  box-shadow: 0 -8px 24px rgba(0, 0, 0, .35);
}

.app-tabbar :deep(.van-tabbar-item) {
  background: transparent;
  color: var(--text-muted);   /* 未选中：灰 */
  font-size: 11px;
}
.app-tabbar :deep(.van-tabbar-item__icon) { font-size: 22px; margin-bottom: 3px; }

.app-tabbar :deep(.van-tabbar-item--active) {
  color: var(--color-primary);  /* 选中：紫，高亮 */
}
.app-tabbar :deep(.van-tabbar-item--active .van-tabbar-item__icon) {
  filter: drop-shadow(0 0 8px rgba(123, 92, 255, .65));
}
.app-tabbar :deep(.van-badge) { background: var(--c-danger); border: none; }
</style>
