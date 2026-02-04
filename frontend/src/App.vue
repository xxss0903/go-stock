<script setup>
import {
  EventsEmit,
  EventsOff,
  EventsOn,
  Quit,
  WindowFullscreen,
  WindowHide,
  WindowUnfullscreen,
  WindowSetTitle
} from '../wailsjs/runtime'
import {h, onBeforeMount, onBeforeUnmount, onMounted, ref} from "vue";
import {RouterLink, useRouter} from 'vue-router'
import {createDiscreteApi,darkTheme,lightTheme , NIcon, NText,NButton,dateZhCN,zhCN} from 'naive-ui'
import {
  AlarmOutline,
  AnalyticsOutline,
  BarChartSharp, Bonfire, BonfireOutline, EaselSharp,
  ExpandOutline, Flag,
  Flame, FlameSharp, InformationOutline,
  LogoGithub,
  NewspaperOutline,
  NewspaperSharp, Notifications,
  PowerOutline, Pulse,
  ReorderTwoOutline,
  SettingsOutline, Skull, SkullOutline, SkullSharp,
  SparklesOutline,
  StarOutline,
  Wallet, WarningOutline,
  DocumentTextOutline,
  BookOutline,
} from '@vicons/ionicons5'
import {AnalyzeSentiment, GetConfig, GetGroupList,GetVersionInfo} from "../wailsjs/go/main/App";
import {Dragon, Fire, FirefoxBrowser, Gripfire, Robot} from "@vicons/fa";
import {ReportSearch} from "@vicons/tabler";
import {LocalFireDepartmentRound} from "@vicons/material";
import {BoxSearch20Regular, CommentNote20Filled} from "@vicons/fluent";
import {FireFilled, FireOutlined, NotificationFilled, StockOutlined} from "@vicons/antd";




const router = useRouter()
const loading = ref(true)
const loadingMsg = ref("加载数据中...")
const enableNews = ref(false)
const contentStyle = ref("")
const enableAgent = ref(false)
const enableDarkTheme = ref(null)
const content = ref('未经授权,禁止商业目的!\n\n数据来源于网络,仅供参考;投资有风险,入市需谨慎')
const isFullscreen = ref(false)
const activeKey = ref('stock')
const containerRef = ref({})
const realtimeProfit = ref(0)
const telegraph = ref([])
const groupList = ref([])
const officialStatement= ref("")
// "全部"的sort值，从localStorage读取，默认为0
const allGroupSort = ref(Number(localStorage.getItem('allGroupSort') || '0'))
const menuOptions = ref([
  {
    label: () =>
        h(
            RouterLink,
            {
              to: {
                name: 'stock',
                query: {},
                params: {},
              },
              onClick: () => {
                activeKey.value = 'stock'
                // 获取排序后的第一个分组（优先自定义分组，如果没有则使用"全部"）
                if (groupList.value && groupList.value.length > 0) {
                  // 获取排序后的第一个自定义分组（排除"全部"）
                  const sortedGroups = [...groupList.value].sort((a, b) => a.sort - b.sort)
                  
                  if (sortedGroups.length > 0) {
                    // 使用第一个自定义分组
                    const firstGroup = sortedGroups[0]
                    router.push({
                      name: 'stock',
                      query: {
                        groupName: firstGroup.name,
                        groupId: firstGroup.ID,
                      },
                    })
                    setTimeout(() => {
                      EventsEmit("changeTab", firstGroup)
                    }, 100)
                  } else {
                    // 没有自定义分组，使用"全部"
                    router.push({
                      name: 'stock',
                      query: {
                        groupName: '全部',
                        groupId: 0,
                      },
                    })
                    EventsEmit("changeTab", {ID: 0, name: '全部', sort: allGroupSort.value})
                  }
                } else {
                  // 如果还没有加载分组列表，先跳转到默认路由，让stock.vue处理
                  router.push({
                    name: 'stock',
                    query: {},
                  })
                }
              },
            },
            {default: () => '股票自选',}
        ),
    key: 'stock',
    icon: renderIcon(StarOutline),
    children: [
      {
        label: () =>
            h(
                'a',
                {
                  href: '#',
                  type: 'info',
                  onClick: () => {
                    activeKey.value = 'stock'
                    //console.log("push",item)
                    router.push({
                      name: 'stock',
                      query: {
                        groupName: '全部',
                        groupId: 0,
                      },
                    })
                    EventsEmit("changeTab", {ID: 0, name: '全部', sort: allGroupSort.value})
                  },
                  to: {
                    name: 'stock',
                    query: {
                      groupName: '全部',
                      groupId: 0,
                    },
                  }
                },
                {default: () => '全部',}
            ),
        key: 0,
        sort: allGroupSort.value,
      }
    ],
  },
  {
    label: () =>
        h(
            RouterLink,
            {
              href: '#',
              to: {
                name: 'market',
                params: {}
              },
              onClick: () => {
                activeKey.value = 'market'
                EventsEmit("changeMarketTab", {ID: 0, name: '市场快讯'})
              },
            },
            {default: () => '市场行情'}
        ),
    key: 'market',
    icon: renderIcon(NewspaperOutline),
    children: [
      {
        label: () =>
            h(
                RouterLink,
                {
                  href: '#',
                  to: {
                    name: 'market',
                    query: {
                      name: "市场快讯",
                    }
                  },
                  onClick: () => {
                    activeKey.value = 'market'
                    EventsEmit("changeMarketTab", {ID: 0, name: '市场快讯'})
                  },
                },
                {default: () => '市场快讯',}
            ),
        key: 'market1',
        icon: renderIcon(NewspaperSharp),
      },
      {
        label: () =>
            h(
                RouterLink,
                {
                  href: '#',
                  to: {
                    name: 'market',
                    query: {
                      name: "全球股指",
                    },
                  },
                  onClick: () => {
                    activeKey.value = 'market'
                    EventsEmit("changeMarketTab", {ID: 0, name: '全球股指'})
                  },
                },
                {default: () => '全球股指',}
            ),
        key: 'market2',
        icon: renderIcon(BarChartSharp),
      },
      {
        label: () =>
            h(
                RouterLink,
                {
                  href: '#',
                  to: {
                    name: 'market',
                    query: {
                      name: "重大指数",
                    }
                  },
                  onClick: () => {
                    activeKey.value = 'market'
                    EventsEmit("changeMarketTab", {ID: 0, name: '重大指数'})
                  },
                },
                {default: () => '重大指数',}
            ),
        key: 'market3',
        icon: renderIcon(AnalyticsOutline),
      },
      {
        label: () =>
            h(
                RouterLink,
                {
                  href: '#',
                  to: {
                    name: 'market',
                    query: {
                      name: "行业排名",
                    }
                  },
                  onClick: () => {
                    activeKey.value = 'market'
                    EventsEmit("changeMarketTab", {ID: 0, name: '行业排名'})
                  },
                },
                {default: () => '行业排名',}
            ),
        key: 'market4',
        icon: renderIcon(Flag),
      },
      {
        label: () =>
            h(
                RouterLink,
                {
                  href: '#',
                  to: {
                    name: 'market',
                    query: {
                      name: "个股资金流向",
                    }
                  },
                  onClick: () => {
                    activeKey.value = 'market'
                    EventsEmit("changeMarketTab", {ID: 0, name: '个股资金流向'})
                  },
                },
                {default: () => '个股资金流向',}
            ),
        key: 'market5',
        icon: renderIcon(Pulse),
      },
      {
        label: () =>
            h(
                RouterLink,
                {
                  href: '#',
                  to: {
                    name: 'market',
                    query: {
                      name: "龙虎榜",
                    }
                  },
                  onClick: () => {
                    activeKey.value = 'market'
                    EventsEmit("changeMarketTab", {ID: 0, name: '龙虎榜'})
                  },
                },
                {default: () => '龙虎榜',}
            ),
        key: 'market6',
        icon: renderIcon(Dragon),
      },
      {
        label: () =>
            h(
                RouterLink,
                {
                  href: '#',
                  to: {
                    name: 'market',
                    query: {
                      name: "个股研报",
                    }
                  },
                  onClick: () => {
                    activeKey.value = 'market'
                    EventsEmit("changeMarketTab", {ID: 0, name: '个股研报'})
                  },
                },
                {default: () => '个股研报',}
            ),
        key: 'market7',
        icon: renderIcon(StockOutlined),
      },
      {
        label: () =>
            h(
                RouterLink,
                {
                  href: '#',
                  to: {
                    name: 'market',
                    query: {
                      name: "公司公告",
                    }
                  },
                  onClick: () => {
                    activeKey.value = 'market'
                    EventsEmit("changeMarketTab", {ID: 0, name: '公司公告'})
                  },
                },
                {default: () => '公司公告',}
            ),
        key: 'market8',
        icon: renderIcon(NotificationFilled),
      },
      {
        label: () =>
            h(
                RouterLink,
                {
                  href: '#',
                  to: {
                    name: 'market',
                    query: {
                      name: "行业研究",
                    }
                  },
                  onClick: () => {
                    activeKey.value = 'market'
                    EventsEmit("changeMarketTab", {ID: 0, name: '行业研究'})
                  },
                },
                {default: () => '行业研究',}
            ),
        key: 'market9',
        icon: renderIcon(ReportSearch),
      },
      {
        label: () =>
            h(
                RouterLink,
                {
                  href: '#',
                  to: {
                    name: 'market',
                    query: {
                      name: "当前热门",
                    }
                  },
                  onClick: () => {
                    activeKey.value = 'market'
                    EventsEmit("changeMarketTab", {ID: 0, name: '当前热门'})
                  },
                },
                {default: () => '当前热门',}
            ),
        key: 'market10',
        icon: renderIcon(Gripfire),
      },
      {
        label: () =>
            h(
                RouterLink,
                {
                  href: '#',
                  to: {
                    name: 'market',
                    query: {
                      name: "指标选股",
                    }
                  },
                  onClick: () => {
                    activeKey.value = 'market'
                    EventsEmit("changeMarketTab", {ID: 0, name: '指标选股'})
                  },
                },
                {default: () => '指标选股',}
            ),
        key: 'market11',
        icon: renderIcon(BoxSearch20Regular),
      },
      {
        label: () =>
            h(
                RouterLink,
                {
                  href: '#',
                  to: {
                    name: 'market',
                    query: {
                      name: "名站优选",
                    }
                  },
                  onClick: () => {
                    activeKey.value = 'market'
                    EventsEmit("changeMarketTab", {ID: 0, name: '名站优选'})
                  },
                },
                {default: () => '名站优选',}
            ),
        key: 'market12',
        icon: renderIcon(FirefoxBrowser),
      },
    ]
  },
  {
    label: () =>
        h(
            RouterLink,
            {
              to: {
                name: 'trading-record',
                query: {
                  name:"炒股复盘",
                },
                onClick: () => {
                  activeKey.value = 'trading-record'
                },
              }
            },
            {default: () => '炒股复盘'}
        ),
    key: 'trading-record',
    icon: renderIcon(DocumentTextOutline),
  },
  {
    label: () =>
        h(
            RouterLink,
            {
              to: {
                name: 'leader-stock',
                query: {
                  name:"龙头板块",
                },
                onClick: () => {
                  activeKey.value = 'leader-stock'
                },
              }
            },
            {default: () => '龙头板块'}
        ),
    key: 'leader-stock',
    icon: renderIcon(Flame),
  },
  {
    label: () =>
        h(
            RouterLink,
            {
              to: {
                name: 'learning',
                query: {
                  name:"学习",
                },
                onClick: () => {
                  activeKey.value = 'learning'
                },
              }
            },
            {default: () => '学习'}
        ),
    key: 'learning',
    icon: renderIcon(BookOutline),
  },
  {
    label: () =>
        h(
            RouterLink,
            {
              to: {
                name: 'agent',
                query: {
                  name:"Ai智能体",
                },
                onClick: () => {
                  activeKey.value = 'agent'
                },
              }
            },
            {default: () => 'Ai智能体'}
        ),
    key: 'agent',
    show:enableAgent.value,
    icon: renderIcon(Robot),
  },
  {
    label: () =>
        h(
            RouterLink,
            {
              to: {
                name: 'settings',
                query: {
                  name:"设置",
                },
                onClick: () => {
                  activeKey.value = 'settings'
                },
              }
            },
            {default: () => '设置'}
        ),
    key: 'settings',
    icon: renderIcon(SettingsOutline),
  },
  {
    label: () =>
        h(
            RouterLink,
            {
              to: {
                name: 'about',
                query: {
                  name:"关于",
                }
              },
              onClick: () => {
                activeKey.value = 'about'
              },
            },
            {default: () => '关于'}
        ),
    key: 'about',
    icon: renderIcon(LogoGithub),
  },
  {
    show:false,
    label: () => h("a", {
      href: '#',
      onClick: toggleFullscreen,
      title: '全屏 Ctrl+F 退出全屏 Esc',
    }, {default: () => isFullscreen.value ? '取消全屏' : '全屏'}),
    key: 'full',
    icon: renderIcon(ExpandOutline),
  },
  {
    label: () => h("a", {
      href: '#',
      onClick: WindowHide,
      title: '隐藏到托盘区 Ctrl+Z',
    }, {default: () => '隐藏到托盘区'}),
    key: 'hide',
    icon: renderIcon(ReorderTwoOutline),
  },
  // {
  //   label: ()=> h("a", {
  //     href: 'javascript:void(0)',
  //     style: 'cursor: move;',
  //     onClick: toggleStartMoveWindow,
  //   }, { default: () => '移动' }),
  //   key: 'move',
  //   icon: renderIcon(MoveOutline),
  // },
  {
    label: () => h("a", {
      href: '#',
      onClick: Quit,
    }, {default: () => '退出程序'}),
    key: 'exit',
    icon: renderIcon(PowerOutline),
  },
])

function renderIcon(icon) {
  return () => h(NIcon, null, {default: () => h(icon)})
}

function toggleFullscreen(e) {
  activeKey.value = 'full'
  //console.log(e)
  if (isFullscreen.value) {
    WindowUnfullscreen()
    //e.target.innerHTML = '全屏'
  } else {
    WindowFullscreen()
    // e.target.innerHTML = '取消全屏'
  }
  isFullscreen.value = !isFullscreen.value
}

// const drag = ref(false)
// const lastPos= ref({x:0,y:0})
// function toggleStartMoveWindow(e) {
//   drag.value=!drag.value
//   lastPos.value={x:e.clientX,y:e.clientY}
// }
// function dragstart(e) {
//   if (drag.value) {
//     let x=e.clientX-lastPos.value.x
//     let y=e.clientY-lastPos.value.y
//     WindowGetPosition().then((pos) => {
//       WindowSetPosition(pos.x+x,pos.y+y)
//     })
//   }
// }
// window.addEventListener('mousemove', dragstart)

EventsOn("realtime_profit", (data) => {
  realtimeProfit.value = data
})
EventsOn("telegraph", (data) => {
  telegraph.value = data
})

EventsOn("loadingMsg", (data) => {
  if(data==="done"){
    loadingMsg.value = "加载完成..."
    EventsEmit("loadingDone", "app")
    loading.value  = false
  }else{
    loading.value  = true
    loadingMsg.value = data
  }
})

onBeforeUnmount(() => {
  EventsOff("realtime_profit")
  EventsOff("loadingMsg")
  EventsOff("telegraph")
  EventsOff("newsPush")
})

window.onerror = function (msg, source, lineno, colno, error) {
  // 将错误信息发送给后端
  EventsEmit("frontendError", {
    page: "App.vue",
    message: msg,
    source: source,
    lineno: lineno,
    colno: colno,
    error: error ? error.stack : null,
  });
  return true;
};

onBeforeMount(() => {
  GetVersionInfo().then(result => {
    if(result.officialStatement){
      content.value = result.officialStatement+"\n\n"+content.value
      officialStatement.value = result.officialStatement
    }
  })

  GetGroupList().then(result => {
    groupList.value = result
    menuOptions.value.map((item) => {
      //console.log(item)
      if (item.key === 'stock') {
        // 添加自定义分组
        item.children.push(...groupList.value.map(item => {
          return {
            label: () =>
                h(
                    'a',
                    {
                      href: '#',
                      type: 'info',
                      onClick: () => {
                        //console.log("push",item)
                        router.push({
                          name: 'stock',
                          query: {
                            groupName: item.name,
                            groupId: item.ID,
                          },
                        })
                        setTimeout(() => {
                          EventsEmit("changeTab", item)
                        }, 100)
                      },
                      to: {
                        name: 'stock',
                        query: {
                          groupName: item.name,
                          groupId: item.ID,
                        },
                      }
                    },
                    {default: () => item.name,}
                ),
            key: item.ID,
            sort: item.sort,
          }
        }))
        // 根据sort值排序children（包括"全部"）
        item.children.sort((a, b) => {
          const sortA = a.sort !== undefined ? a.sort : allGroupSort.value
          const sortB = b.sort !== undefined ? b.sort : allGroupSort.value
          return sortA - sortB
        })
      }
    })
  })
  
  // 监听"全部"sort值的变化，更新菜单
  EventsOn("updateAllGroupSort", (data) => {
    allGroupSort.value = data.sort
    localStorage.setItem('allGroupSort', String(data.sort))
    // 重新排序菜单
    menuOptions.value.map((item) => {
      if (item.key === 'stock') {
        item.children.sort((a, b) => {
          const sortA = a.sort !== undefined ? a.sort : allGroupSort.value
          const sortB = b.sort !== undefined ? b.sort : allGroupSort.value
          return sortA - sortB
        })
      }
    })
  })


  GetConfig().then((res) => {
    //console.log(res)
    enableAgent.value = res.enableAgent

    menuOptions.value.filter((item) => {
      if (item.key === 'agent') {
        item.show = res.enableAgent
      }
    })

    if (res.darkTheme) {
      enableDarkTheme.value = darkTheme
    } else {
      enableDarkTheme.value = null
    }
  })
})

onMounted(() => {
  WindowSetTitle("go-stock：AI赋能股票分析✨ "+officialStatement.value+"  未经授权,禁止商业目的！ [数据来源于网络,仅供参考;投资有风险,入市需谨慎]")
  contentStyle.value = "max-height: calc(92vh);overflow: hidden"
  GetConfig().then((res) => {
    if (res.enableNews) {
      enableNews.value = true
    }
    enableAgent.value = res.enableAgent
    const {notification } =createDiscreteApi(["notification"], {
      configProviderProps: {
        theme: enableDarkTheme.value ? darkTheme : lightTheme ,
        max: 3,
      },
    })
    EventsOn("newsPush", (data) => {
      //console.log(data)
      if(data.isRed){
        notification.create({
          //type:"error",
         // avatar: () => h(NIcon,{component:Notifications,color:"red"}),
          title: data.time,
          content: () => h('div',{type:"error",style:{
              "text-align":"left",
              "font-size":"14px",
              "color":"#f67979"
            }}, { default: () => data.content }),
          meta: () => h(NText,{type:"warning"}, { default: () => data.source}),
          duration:1000*40,
        })
      }else{
         notification.create({
          //type:"info",
          //avatar: () => h(NIcon,{component:Notifications}),
          title: data.time,
          content: () => h('div',{type:"info",style:{
            "text-align":"left",
              "font-size":"14px",
              "color": data.source==="go-stock"?"#F98C24":"#549EC8"
            }}, { default: () => data.content }),
          meta: () => h(NText,{type:"warning"}, { default: () => data.source}),
          duration:1000*30 ,
        })
      }
    })
  })
})
</script>
<template>
  <n-config-provider ref="containerRef" :theme="enableDarkTheme" :locale="zhCN" :date-locale="dateZhCN">
    <n-message-provider>
      <n-notification-provider>
        <n-modal-provider>
          <n-dialog-provider>
            <n-watermark
                :content="''"
                cross
                selectable
                :font-size="16"
                :line-height="16"
                :width="500"
                :height="400"
                :x-offset="50"
                :y-offset="150"
                :rotate="-15"
            >
              <n-flex>
                <n-grid x-gap="12" :cols="1">
                  <n-gi>
                    <n-spin :show="loading">
                      <template #description>
                        {{ loadingMsg }}
                      </template>
                      <n-marquee :speed="100" style="position: relative;top:0;z-index: 19;width: 100%"
                                 v-if="(telegraph.length>0)&&(enableNews)">
                        <n-tag type="warning" v-for="item in telegraph" style="margin-right: 10px">
                          {{ item }}
                        </n-tag>
                      </n-marquee>
                      <n-scrollbar :style="contentStyle">
                        <n-skeleton v-if="loading" height="calc(100vh)" />
                        <RouterView/>
                      </n-scrollbar>
                    </n-spin>
                  </n-gi>
                  <n-gi style="position: fixed;bottom:0;z-index: 9;width: 100%;">
                    <n-card size="small" style="--wails-draggable:no-drag">
                      <n-menu style="font-size: 18px;"
                              v-model:value="activeKey"
                              mode="horizontal"
                              :options="menuOptions"
                              responsive
                      />
                    </n-card>
                  </n-gi>
                </n-grid>
              </n-flex>
            </n-watermark>
          </n-dialog-provider>
        </n-modal-provider>
      </n-notification-provider>
    </n-message-provider>
  </n-config-provider>
</template>
<style>

</style>
