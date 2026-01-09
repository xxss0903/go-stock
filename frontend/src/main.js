import {createApp} from 'vue'
import naive from 'naive-ui'
import App from './App.vue'
import router from './router/router'
// 引入组件库的少量全局样式变量
import 'tdesign-vue-next/es/style/index.css';

// 禁用所有console日志输出
if (process.env.NODE_ENV === 'production' || true) { // 始终禁用日志
  console.log = () => {};
  console.error = () => {};
  console.warn = () => {};
  console.debug = () => {};
  console.info = () => {};
}

const app = createApp(App)
app.use(router)
app.use(naive)
app.mount('#app')