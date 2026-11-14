import { createApp } from 'vue'
import { createPinia } from 'pinia'
import { Tabbar, TabbarItem, Icon, Button, Loading, Field, CellGroup, Tab, Tabs } from 'vant'

// 顺序关键：先加载 Vant 默认 :root 变量，再用 variables.css 覆盖（同优先级后者胜）
import 'vant/lib/index.css'
import './assets/styles/variables.css'
import './assets/styles/global.css'

import App from './App.vue'
import router from './router'

const app = createApp(App)
app.use(createPinia())
app.use(router)

// 全局注册（若装了 VantResolver 可删掉这段）；showToast 是函数，无需注册
;[Tabbar, TabbarItem, Icon, Button, Loading, Field, CellGroup, Tab, Tabs].forEach((c) => app.use(c))

app.mount('#app')
