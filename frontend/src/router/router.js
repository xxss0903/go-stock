import {createMemoryHistory, createRouter, createWebHashHistory, createWebHistory} from 'vue-router'

import stockView from '../components/stock.vue'
import settingsView from '../components/settings.vue'
import aboutView from "../components/about.vue";
import marketView from "../components/market.vue";
import agentChat from "../components/agent-chat.vue"
import tradingRecordView from "../components/trading-record.vue"
import learningView from "../components/learning.vue"
import leaderStockView from "../components/leader-stock.vue"

const routes = [
    { path: '/', component: stockView,name: 'stock'},
    { path: '/settings', component: settingsView,name: 'settings' },
    { path: '/about', component: aboutView,name: 'about' },
    { path: '/market', component: marketView,name: 'market' },
    { path: '/agent', component: agentChat,name: 'agent' },
    { path: '/trading-record', component: tradingRecordView,name: 'trading-record' },
    { path: '/learning', component: learningView,name: 'learning' },
    { path: '/leader-stock', component: leaderStockView,name: 'leader-stock' },
]

const router = createRouter({
    //history: createWebHistory(),
    history: createWebHashHistory(),
    routes,
})

export default router