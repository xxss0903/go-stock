<script setup>
import {computed, h, nextTick, onBeforeMount, onBeforeUnmount, onMounted, reactive, ref, watch} from 'vue'
import * as echarts from 'echarts';
import {
  AddGroup,
  AddPrompt,
  AddStockGroup,
  Follow,
  GetAiConfigs,
  GetAIResponseResult,
  GetConfig,
  GetFollowList,
  GetGroupList,
  GetPromptTemplates,
  GetStockKLine,
  GetStockList,
  GetStockMinutePriceLineData,
  GetVersionInfo,
  Greet,
  InitializeGroupSort,
  NewChatStream,
  OpenURL,
  RemoveGroup,
  RemoveStockGroup,
  SaveAIResponseResult,
  SaveAsMarkdown,
  SaveImage,
  SaveWordFile,
  SearchStock,
  SendDingDingMessageByType,
  SetAlarmChangePercent,
  SetCostPriceAndVolume,
  SetStockAICron,
  SetStockSort,
  ShareAnalysis,
  SummaryStockNews,
  UnFollow,
  UpdateGroupSort,
  UpdateGroup
} from '../../wailsjs/go/main/App'
import {
  NAlert,
  NAvatar,
  NButton,
  NButtonGroup,
  NDataTable,
  NDatePicker,
  NDropdown,
  NEllipsis,
  NEmpty,
  NFlex,
  NForm,
  NFormItem,
  NGradientText,
  NIcon,
  NInput,
  NInputGroup,
  NInputNumber,
  NList,
  NListItem,
  NModal,
  NScrollbar,
  NSelect,
  NSpace,
  NStatistic,
  NSpin,
  NSwitch,
  NTag,
  NText,
  useDialog,
  useMessage,
  useNotification
} from 'naive-ui'
import {
  Environment,
  EventsEmit,
  EventsOff,
  EventsOn,
  WindowFullscreen,
  WindowReload,
  WindowUnfullscreen
} from '../../wailsjs/runtime'
import {Add, ChatboxOutline, PulseOutline, GridOutline, ListOutline, SchoolOutline, TrashOutline} from '@vicons/ionicons5'
import {MdEditor, MdPreview} from 'md-editor-v3';
// preview.css相比style.css少了编辑器那部分样式
//import 'md-editor-v3/lib/preview.css';
import 'md-editor-v3/lib/style.css';

import {ExportPDF} from '@vavt/v3-extension';
import '@vavt/v3-extension/lib/asset/ExportPDF.css';
import html2canvas from "html2canvas";
import {asBlob} from 'html-docx-js-typescript';

import vueDanmaku from 'vue3-danmaku'
import {keys, padStart} from "lodash";
import {useRoute, useRouter} from 'vue-router'
import MoneyTrend from "./moneyTrend.vue";
import StockSparkLine from "./stockSparkLine.vue";

const route = useRoute()
const router = useRouter()

const danmus = ref([])
const ws = ref(null)
const dialog = useDialog()
const toolbars = [0];

const upColor = '#ec0000';
const upBorderColor = '';
const downColor = '#00da3c';
const downBorderColor = '';
const kLineChartRef = ref(null);
const kLineChartRef2 = ref(null);


const handleProgress = (progress) => {
  //console.log(`Export progress: ${progress.ratio * 100}%`);
};
const enableEditor = ref(false)
const mdPreviewRef = ref(null)
const mdEditorRef = ref(null)
const tipsRef = ref(null)
const message = useMessage()
const notify = useNotification()
const stocks = ref([])
const results = ref({})
const stockList = ref([])
const followList = ref([])
const groupList = ref([])
const options = ref([])
const modalShow = ref(false)
const modalShow2 = ref(false)
const modalShow3 = ref(false)
const modalShow4 = ref(false)
const modalShow5 = ref(false)
const summaryModal = ref(false) // AI总结对话框
const aiStockSelectModal = ref(false) // AI选股对话框
const addBTN = ref(true)
const editingGroupId = ref(null) // 正在编辑的分组ID
const editingGroupName = ref('') // 正在编辑的分组名称
const isSavingGroupName = ref(false) // 是否正在保存分组名称
const enableTools = ref(true) // 默认开启AI调用函数功能
const thinkingMode = ref(false)
const formModel = ref({
  name: "",
  code: "",
  costPrice: 0.000,
  volume: 0,
  buyDate: null, // 买入日期
  alarm: 0,
  alarmPrice: 0,
  sort: 999,
  cron: "",
})

const promptTemplates = ref([])
const aiConfigs = ref([])
const sysPromptOptions = ref([])
const userPromptOptions = ref([])
// AI总结相关状态
const aiSummary = ref('')
const aiSummaryTime = ref('')
const aiSummaryModelName = ref('')
const aiSummaryChatId = ref('')
const aiSummaryQuestion = ref('')
const aiSummaryConfigId = ref(null)
const aiSummarySysPromptId = ref(null)
const aiSummaryLoading = ref(false)

// AI选股相关状态
const aiStockSelectCondition = ref('') // 选股条件
const aiStockSelectResult = ref('') // AI选股结果
const aiStockSelectLoading = ref(false) // 加载状态
const aiStockSelectConfigId = ref(null) // AI配置ID
const aiStockSelectSysPromptId = ref(null) // 系统提示词ID
const aiStockSelectChatId = ref('') // 对话ID
const aiStockSelectModelName = ref('') // 模型名称
const aiStockSelectTime = ref('') // 时间
const isAiStockSelectMode = ref(false) // 标志：是否是AI选股模式

// 指标选股相关状态
const indicatorStockSelectModal = ref(false) // 指标选股对话框
const indicatorStockSelectCondition = ref('') // 选股条件
const indicatorStockSelectLoading = ref(false) // 加载状态
const indicatorStockSelectColumns = ref([]) // 表格列
const indicatorStockSelectDataList = ref([]) // 股票数据列表
const indicatorStockSelectTraceInfo = ref('') // 选股条件解析信息
const indicatorStockSelectTableScrollX = ref(2800) // 表格滚动宽度

// 显示模式：'card' 卡片式，'list' 列表式
const displayMode = ref('card') // 默认卡片式

// 补仓计算器相关状态（已移至learning.vue，保留以兼容旧代码）
const costCalculator = reactive({
  // 初始持仓
  initialPrice: 10, // 初始买入价格
  initialVolume: 1000, // 初始买入数量
  // 补仓记录
  additions: [],
  // 当前价格
  currentPrice: 9,
  // 下一个补仓记录的ID
  nextId: 1
})

const data = reactive({
  modelName: "",
  chatId: "",
  question: "",
  sysPromptId: null,
  aiConfigId: null,
  name: "",
  code: "",
  fenshiURL: "",
  kURL: "",
  resultText: "Please enter your name below 👇",
  fullscreen: false,
  airesult: "",
  openAiEnable: false,
  loading: true,
  enableDanmu: false,
  darkTheme: false,
  changePercent: 0
})
const feishiInterval = ref(null)


const currentGroupId = ref(0)
// "全部"的sort值，从localStorage读取，默认为0
const allGroupSort = ref(Number(localStorage.getItem('allGroupSort') || '0'))


const theme = computed(() => {
  return data.darkTheme ? 'dark' : 'light'
})

const danmakuColor = computed(() => {
  return data.darkTheme ? 'color:#fff' : 'color:#000'
})

const icon = ref('https://raw.githubusercontent.com/ArvinLovegood/go-stock/master/build/appicon.png');

const sortedResults = computed(() => {
  const sortedKeys = keys(results.value).sort();
  const sortedObject = {};
  sortedKeys.forEach(key => {
    sortedObject[key] = results.value[key];
  });
  return sortedObject
});

const groupResults = computed(() => {
  const group = {}
  if (currentGroupId.value === 0) {
    return sortedResults.value
  } else {
    for (const key in sortedResults.value) {
      if (stocks.value.includes(sortedResults.value[key]['股票代码'])) {
        group[key] = sortedResults.value[key]
      }
    }
    return group
  }
})

// 排序后的分组列表（包括"全部"）
const sortedGroupList = computed(() => {
  const allGroup = { ID: 0, name: '全部', sort: allGroupSort.value }
  const groups = [allGroup, ...groupList.value]
  return groups.sort((a, b) => a.sort - b.sort)
})
const showPopover = ref(false)
// 拖拽相关变量
const dragSourceIndex = ref(null)
const dragTargetIndex = ref(null)

// 拖拽处理函数
function handleTabDragStart(event, name) {
  // 允许"全部"标签也参与拖拽
  dragSourceIndex.value = name;
  event.dataTransfer.effectAllowed = 'move';
  event.target.classList.add('tab-dragging');
}


function handleTabDragOver(event) {
  event.preventDefault()
  event.dataTransfer.dropEffect = 'move'
}

function handleTabDragEnter(event, name) {
  event.preventDefault();
  // 允许"全部"标签也作为拖拽目标
  dragTargetIndex.value = name;
  if (event.target.classList) {
    // 查找最近的标签元素并添加高亮样式
    let tabElement = event.target.closest('.n-tabs-tab');
    if (tabElement) {
      tabElement.classList.add('tab-drag-over');
    }
  }
}

function handleTabDragLeave(event) {
  // 查找最近的标签元素并移除高亮样式
  let tabElement = event.target.closest('.n-tabs-tab')
  if (tabElement && tabElement.classList) {
    tabElement.classList.remove('tab-drag-over')
  }
  // 不要重置 dragTargetIndex，因为可能会在元素间快速移动
}

function handleTabDrop(event) {
  event.preventDefault();

  // 移除所有高亮样式
  const tabs = document.querySelectorAll('.n-tabs-tab');
  tabs.forEach(tab => {
    tab.classList.remove('tab-drag-over');
  });

  if (dragSourceIndex.value !== null && dragTargetIndex.value !== null &&
    dragSourceIndex.value !== dragTargetIndex.value) {

    const sourceId = dragSourceIndex.value;
    const targetId = dragTargetIndex.value;
    
    // 获取目标分组的sort值
    let targetSort = 0;
    if (targetId === 0) {
      // 目标是"全部"
      targetSort = allGroupSort.value;
    } else {
      // 目标是自定义分组
      const targetGroup = groupList.value.find(g => g.ID === targetId);
      if (targetGroup) {
        targetSort = targetGroup.sort;
      } else {
        // 重置状态
        dragSourceIndex.value = null;
        dragTargetIndex.value = null;
        return;
      }
    }

    if (sourceId === 0) {
      // 拖拽的是"全部"
      // 更新"全部"的sort值
      allGroupSort.value = targetSort;
      localStorage.setItem('allGroupSort', String(targetSort));
      // 通知App.vue更新菜单
      EventsEmit("updateAllGroupSort", { sort: targetSort });
      message.success('分组排序更新成功');
      // 重新获取分组列表以更新界面
      GetGroupList().then(result => {
        groupList.value = result;
      });
    } else if (targetId === 0) {
      // 拖拽自定义分组到"全部"位置
      const sourceGroup = groupList.value.find(g => g.ID === sourceId);
      if (sourceGroup) {
        // 更新自定义分组的sort值为"全部"的sort值
        UpdateGroupSort(sourceGroup.ID, allGroupSort.value).then(result => {
          if (result) {
            message.success('分组排序更新成功');
            // 重新获取分组列表以更新界面
            GetGroupList().then(result => {
              groupList.value = result;
            });
          } else {
            message.error('分组排序更新失败');
          }
        }).catch(error => {
          message.error('分组排序更新失败: ' + error.message);
        });
      }
    } else {
      // 拖拽的是自定义分组
      const sourceGroup = groupList.value.find(g => g.ID === sourceId);
      if (sourceGroup) {
        // 调用后端API更新组排序
        UpdateGroupSort(sourceGroup.ID, targetSort).then(result => {
          if (result) {
            message.success('分组排序更新成功');
            // 重新获取分组列表以更新界面
            GetGroupList().then(result => {
              groupList.value = result;
            });
          } else {
            message.error('分组排序更新失败');
          }
        }).catch(error => {
          message.error('分组排序更新失败: ' + error.message);
        });
      }
    }
  }

  // 重置状态
  dragSourceIndex.value = null;
  dragTargetIndex.value = null;
}

function handleTabDragEnd(event) {
  // 移除所有高亮样式
  const tabs = document.querySelectorAll('.n-tabs-tab')
  tabs.forEach(tab => {
    tab.classList.remove('tab-drag-over', 'tab-dragging')
  })

  dragSourceIndex.value = null
  dragTargetIndex.value = null
}

onBeforeMount(() => {
  // 从localStorage读取"全部"的sort值
  allGroupSort.value = Number(localStorage.getItem('allGroupSort') || '0')
  
  GetGroupList().then(result => {
    groupList.value = result
    // 检查是否存在相同的序号
    const sorts = result.map(item => item.sort);
    const uniqueSorts = new Set(sorts);
    // 如果存在重复的序号，则重新初始化序号
    if (sorts.length !== uniqueSorts.size) {
      // 调用InitializeGroupSort重新初始化序号
      // 然后重新获取分组列表
      fetchGroupList();
    } else {
      // 没有重复序号，继续正常流程
      if (route.query.groupId !== undefined) {
        // 如果路由上已经带了groupId，就按照路由参数来
        const gid = Number(route.query.groupId)
        currentGroupId.value = gid
        if (route.query.groupName) {
          message.success("切换分组:" + route.query.groupName)
        }
      } else if (groupList.value.length > 0) {
        // 否则使用分组中排在最前面的那个作为默认分组（不包含“全部”）
        const firstGroup = groupList.value[0]
        currentGroupId.value = firstGroup.ID
        // 更新路由上的查询参数，保持状态一致
        router.replace({
          name: 'stock',
          query: {
            groupName: firstGroup.name,
            groupId: firstGroup.ID,
          },
        })
        // 触发一次切换分组逻辑，加载该分组的股票
        updateTab(String(firstGroup.ID))
      }
    }
  })
  GetStockList("").then(result => {
    stockList.value = result
    options.value = result.map(item => {
      return {
        label: item.name + " - " + item.ts_code,
        value: item.ts_code
      }
    })
  })
  GetConfig().then(result => {
    if (result.openAiEnable) {
      data.openAiEnable = true
    }
    if (result.enableDanmu) {
      data.enableDanmu = true
    }
    if (result.darkTheme) {
      data.darkTheme = true
    }
  })
  GetPromptTemplates("", "").then(res => {
    promptTemplates.value = res

    sysPromptOptions.value = promptTemplates.value.filter(item => item.type === '模型系统Prompt')
    userPromptOptions.value = promptTemplates.value.filter(item => item.type === '模型用户Prompt')

  })

  GetAiConfigs().then(res => {
    aiConfigs.value = res
    data.aiConfigId = res[0].ID
    // 初始化AI总结和AI选股的配置
    if (res.length > 0) {
      aiSummaryConfigId.value = res[0].ID
      aiStockSelectConfigId.value = res[0].ID
    }
  })

  EventsOn("loadingDone", (data) => {
    message.loading("刷新股票基础数据...")
    GetStockList("").then(result => {
      stockList.value = result
      options.value = result.map(item => {
        return {
          label: item.name + " - " + item.ts_code,
          value: item.ts_code
        }
      })
    })
  })

  EventsOn("refresh", (data) => {
    message.success(data)
  })

  EventsOn("showSearch", (data) => {
    addBTN.value = data === 1;
  })

  EventsOn("stock_price", (data) => {
    updateData(data)
  })

  EventsOn("refreshFollowList", (data) => {

    WindowReload()
  })

  EventsOn("newChatStream", async (msg) => {
    data.loading = false
    if (msg === "DONE") {
      SaveAIResponseResult(data.code, data.name, data.airesult, data.chatId, data.question, data.aiConfigId)
      message.info("AI分析完成！")
      message.destroyAll()
    } else {
      // 检查是否是错误消息 (code === 0 表示错误)
      if (msg.code === 0 && msg.content) {
        // 显示错误提示
        message.error("AI分析出错，请查看下方错误信息")
        // 确保错误内容被添加到结果中
        data.airesult = data.airesult + "\n\n" + msg.content
        return
      }
      if (msg.chatId) {
        data.chatId = msg.chatId
      }
      if (msg.question) {
        data.question = msg.question
      }
      if (msg.content) {
        data.airesult = data.airesult + msg.content
      }
      if (msg.extraContent) {
        data.airesult = data.airesult + msg.extraContent
      }

    }
  })

  // 监听市场行情的AI总结事件（股票自选AI对话和AI选股）
  EventsOn("summaryStockNews", async (msg) => {
    // 使用标志变量判断是AI选股模式还是AI总结模式
    const isStockSelectMode = isAiStockSelectMode.value
    
    console.log('summaryStockNews事件:', {
      isStockSelectMode,
      flag: isAiStockSelectMode.value,
      modalOpen: aiStockSelectModal.value,
      loading: aiStockSelectLoading.value,
      msg: msg === "DONE" ? "DONE" : typeof msg,
      hasResult: aiStockSelectResult.value?.length > 0
    })
    
    if (isStockSelectMode) {
      // AI选股模式
      if (msg === "DONE") {
        aiStockSelectLoading.value = false
        await SaveAIResponseResult("股票自选-选股", "股票自选-选股", aiStockSelectResult.value, aiStockSelectChatId.value, aiStockSelectCondition.value, aiStockSelectConfigId.value)
        
        // 解析股票数据并添加到自选
        console.log('AI选股结果:', aiStockSelectResult.value)
        const stocks = parseStocksFromAiResult(aiStockSelectResult.value)
        console.log('解析出的股票:', stocks)
        if (stocks.length > 0) {
          console.log('开始添加股票到自选，数量:', stocks.length)
          await addParsedStocksToFollow(stocks)
        } else {
          console.warn('未找到股票数据，AI返回内容:', aiStockSelectResult.value)
          message.warning('未在AI返回结果中找到股票数据，请检查AI返回的JSON格式。AI返回内容已显示在对话框中。')
        }
        
        // 重置标志
        isAiStockSelectMode.value = false
        
        // 显示完成通知
        message.info("AI选股完成！")
        message.destroyAll()
        
        notify.success({
          avatar: () =>
            h(NAvatar, {
              size: 'small',
              round: false,
              src: icon.value
            }),
          title: 'AI选股完成',
          content: `选股条件：${aiStockSelectCondition.value}\n模型：${aiStockSelectModelName.value || '未知'}\n已解析并添加股票到自选。`,
          duration: 5000,
        })
      } else {
        // 检查是否是错误消息
        if (msg.code === 0 && msg.content) {
          aiStockSelectLoading.value = false
          isAiStockSelectMode.value = false // 重置标志
          message.error("AI选股出错，请查看下方错误信息")
          aiStockSelectResult.value = aiStockSelectResult.value + "\n\n" + msg.content
          notify.error({
            avatar: () =>
              h(NAvatar, {
                size: 'small',
                round: false,
                src: icon.value
              }),
            title: 'AI选股出错',
            content: msg.content || '选股过程中发生错误',
            duration: 5000,
          })
          return
        }
        if (msg.chatId) {
          aiStockSelectChatId.value = msg.chatId
        }
        if (msg.content) {
          aiStockSelectResult.value = aiStockSelectResult.value + msg.content
        }
        if (msg.extraContent) {
          aiStockSelectResult.value = aiStockSelectResult.value + msg.extraContent
        }
        if (msg.model) {
          aiStockSelectModelName.value = msg.model
        }
        if (msg.time) {
          aiStockSelectTime.value = msg.time
        }
      }
    } else {
      // AI总结模式
      aiSummaryLoading.value = false
      if (msg === "DONE") {
        await SaveAIResponseResult("股票自选", "股票自选", aiSummary.value, aiSummaryChatId.value, aiSummaryQuestion.value, aiSummaryConfigId.value)
        
        // 显示完成通知
        message.info("AI分析完成！")
        message.destroyAll()
        
        // 显示更详细的通知
        notify.success({
          avatar: () =>
            h(NAvatar, {
              size: 'small',
              round: false,
              src: icon.value
            }),
          title: 'AI分析完成',
          content: `股票自选AI分析已完成！\n问题：${aiSummaryQuestion.value || '未设置'}\n模型：${aiSummaryModelName.value || '未知'}\n结果已保存，可在对话框中查看。`,
          duration: 5000,
          meta: () => h('div', {
            style: {
              'font-size': '12px',
              'color': 'var(--n-text-color-2)',
              'margin-top': '8px'
            }
          }, {default: () => `分析时间：${aiSummaryTime.value || new Date().toLocaleString()}`})
        })
      } else {
        // 检查是否是错误消息 (code === 0 表示错误)
        if (msg.code === 0 && msg.content) {
          message.error("AI分析出错，请查看下方错误信息")
          aiSummary.value = aiSummary.value + "\n\n" + msg.content
          
          // 显示错误通知
          notify.error({
            avatar: () =>
              h(NAvatar, {
                size: 'small',
                round: false,
                src: icon.value
              }),
            title: 'AI分析出错',
            content: msg.content || '分析过程中发生错误，请查看详细信息',
            duration: 5000,
          })
          return
        }
        if (msg.chatId) {
          aiSummaryChatId.value = msg.chatId
        }
        if (msg.question) {
          aiSummaryQuestion.value = msg.question
        }
        if (msg.content) {
          aiSummary.value = aiSummary.value + msg.content
        }
        if (msg.extraContent) {
          aiSummary.value = aiSummary.value + msg.extraContent
        }
        if (msg.model) {
          aiSummaryModelName.value = msg.model
        }
        if (msg.time) {
          aiSummaryTime.value = msg.time
        }
      }
    }
  })

  EventsOn("changeTab", async (msg) => {
    currentGroupId.value = Number(msg.ID)
    nextTick(() => {
      updateTab(currentGroupId.value);
    });
  })


  EventsOn("updateVersion", async (msg) => {
    const githubTimeStr = msg.published_at;
    // 创建一个 Date 对象
    const utcDate = new Date(githubTimeStr);
// 获取本地时间
    const date = new Date(utcDate.getTime());
    const year = date.getFullYear();
// getMonth 返回值是 0 - 11，所以要加 1
    const month = String(date.getMonth() + 1).padStart(2, '0');
    const day = String(date.getDate()).padStart(2, '0');
    const hours = String(date.getHours()).padStart(2, '0');
    const minutes = String(date.getMinutes()).padStart(2, '0');
    const seconds = String(date.getSeconds()).padStart(2, '0');

    const formattedDate = `${year}-${month}-${day} ${hours}:${minutes}:${seconds}`;

    notify.info({
      avatar: () =>
        h(NAvatar, {
          size: 'small',
          round: false,
          src: icon.value
        }),
      title: '发现新版本: ' + msg.tag_name,
      content: () => {
        //return h(MdPreview, {theme:'dark',modelValue:msg.commit?.message}, null)
        return h('div', {
          style: {
            'text-align': 'left',
            'font-size': '14px',
          }
        }, {default: () => msg.commit?.message})
      },
      duration: 5000,
      meta: "发布时间:" + formattedDate,
      action: () => {
        return h(NButton, {
          type: 'primary',
          size: 'small',
          onClick: () => {
            Environment().then(env => {
              switch (env.platform) {
                case 'windows':
                  window.open(msg.html_url)
                  break
                default :
                  OpenURL(msg.html_url)
              }
            })
          }
        }, {default: () => '查看'})
      }
    })
  })

  EventsOn("warnMsg", async (msg) => {
    notify.error({
      avatar: () =>
        h(NAvatar, {
          size: 'small',
          round: false,
          src: icon.value
        }),
      title: '警告',
      duration: 5000,
      content: () => {
        return h('div', {
          style: {
            'text-align': 'left',
            'font-size': '14px',
          }
        }, {default: () => msg})
      },
    })
  })
})

onMounted(() => {
  nextTick(() => {
    initDraggableTabs();
  });

  // 监听分组列表变化，重新初始化拖拽
  const unwatch = watch(groupList, () => {
    nextTick(() => {
      initDraggableTabs();
    });
  });

  // 在组件卸载时清理监听器
  onBeforeUnmount(() => {
    unwatch();
  });
  message.loading("Loading...")
  GetFollowList(currentGroupId.value).then(result => {

    followList.value = result
    for (const followedStock of result) {
      if (followedStock.StockCode.startsWith("us")) {
        followedStock.StockCode = "gb_" + followedStock.StockCode.replace("us", "").toLowerCase()
      }
      if (!stocks.value.includes(followedStock.StockCode)) {
        stocks.value.push(followedStock.StockCode)
      }
      Greet(followedStock.StockCode).then(result => {
        updateData(result)
      })
    }
    //monitor()
    message.destroyAll()
  })

  GetVersionInfo().then((res) => {
    icon.value = res.icon;
  });
  // 创建 WebSocket 连接
  ws.value = new WebSocket('ws://8.134.249.145:16688/ws'); // 替换为你的 WebSocket 服务器地址
  //ws.value = new WebSocket('ws://localhost:16688/ws'); // 替换为你的 WebSocket 服务器地址

  ws.value.onopen = () => {
    //console.log('WebSocket 连接已打开');
  };

  ws.value.onmessage = (event) => {
    if (data.enableDanmu) {
      danmus.value.push(event.data);
    }
  };

  ws.value.onerror = (error) => {
    console.error('WebSocket 错误:', error);
  };

  ws.value.onclose = () => {
    //console.log('WebSocket 连接已关闭');
  };
})
// 清理拖拽事件监听器
// 清理拖拽事件监听器
function cleanupDraggableTabs() {
  const tabs = document.querySelectorAll('.n-tabs-tab');
  tabs.forEach((tab) => {
    // 移除所有可能的拖拽事件监听器
    tab.removeEventListener('dragstart', handleTabDragStart);
    tab.removeEventListener('dragover', handleTabDragOver);
    tab.removeEventListener('dragenter', handleTabDragEnter);
    tab.removeEventListener('dragleave', handleTabDragLeave);
    tab.removeEventListener('drop', handleTabDrop);
    tab.removeEventListener('dragend', handleTabDragEnd);
    // 移除draggable属性
    tab.removeAttribute('draggable');
  });
}

// 初始化可拖拽选项卡
function initDraggableTabs() {
  // 移除之前可能添加的事件监听器
  cleanupDraggableTabs();

  // 添加拖拽事件监听器到选项卡元素
  setTimeout(() => {
    const tabs = document.querySelectorAll('.n-tabs-tab');
    tabs.forEach((tab, index) => {
      // 从tab的data-name属性获取ID
      let name = null;
      const dataIndex = tab.getAttribute('data-name');
      if (dataIndex) {
        name = parseInt(dataIndex);
      } else {
        // 尝试从tab的文本内容匹配
        const tabText = tab.textContent?.trim();
        if (tabText === '全部') {
          name = 0;
        } else {
          // 从sortedGroupList中查找匹配的分组
          const matchedGroup = sortedGroupList.value.find(g => g.name === tabText);
          if (matchedGroup) {
            name = matchedGroup.ID;
          }
        }
      }

      // 为所有标签（包括"全部"）添加拖拽功能
      if (!isNaN(name) && name !== null) {
        tab.setAttribute('draggable', 'true');
        tab.setAttribute('data-name', String(name));
        tab.addEventListener('dragstart', (e) => handleTabDragStart(e, name));
        tab.addEventListener('dragover', handleTabDragOver);
        tab.addEventListener('dragenter', (e) => handleTabDragEnter(e, name));
        tab.addEventListener('dragleave', handleTabDragLeave);
        tab.addEventListener('drop', handleTabDrop);
        tab.addEventListener('dragend', handleTabDragEnd);
      }
    });
  }, 100);
}

onBeforeUnmount(() => {
  // //console.log(`the component is now unmounted.`)
  //clearInterval(ticker.value)
  ws.value.close()
  message.destroyAll()
  notify.destroyAll()
  clearInterval(feishiInterval.value)

  EventsOff("refresh")
  EventsOff("showSearch")
  EventsOff("stock_price")
  EventsOff("refreshFollowList")
  EventsOff("newChatStream")
  EventsOff("summaryStockNews")
  EventsOff("changeTab")
  EventsOff("updateVersion")
  EventsOff("warnMsg")
  EventsOff("loadingDone")

  cleanupDraggableTabs()

})

//判断是否是A股交易时间
function isTradingTime() {
  const now = new Date();
  const day = now.getDay(); // 获取星期几，0表示周日，1-6表示周一至周六
  if (day >= 1 && day <= 5) { // 周一至周五
    const hours = now.getHours();
    const minutes = now.getMinutes();
    const totalMinutes = hours * 60 + minutes;
    const startMorning = 9 * 60 + 15; // 上午9点15分换算成分钟数
    const endMorning = 11 * 60 + 30; // 上午11点30分换算成分钟数
    const startAfternoon = 13 * 60; // 下午13点换算成分钟数
    const endAfternoon = 15 * 60; // 下午15点换算成分钟数
    if ((totalMinutes >= startMorning && totalMinutes < endMorning) ||
      (totalMinutes >= startAfternoon && totalMinutes < endAfternoon)) {
      return true;
    }
  }
  return false;
}

// 添加一个获取分组列表的函数，用于处理初始化逻辑
function fetchGroupList() {
  InitializeGroupSort().then(initResult => {
    if (initResult) {
      GetGroupList().then(result => {
        groupList.value = result
        if (route.query.groupId) {
          message.success("切换分组:" + route.query.groupName)
          currentGroupId.value = Number(route.query.groupId)
        }
      })
    } else {
      message.error("初始化分组序号失败")
    }
  })
}

function AddStock() {
  if (!data?.code) {
    message.error("请输入有效股票代码");
    return;
  }
  if (!stocks.value.includes(data.code)) {
    Follow(data.code).then(result => {
      if (result === "关注成功") {
        if (data.code.startsWith("us")) {
          data.code = "gb_" + data.code.replace("us", "").toLowerCase()
        }
        stocks.value.push(data.code)
        message.success(result)
        GetFollowList(currentGroupId.value).then(result => {
          followList.value = result
        })
        monitor();
      } else {
        message.error(result)
      }
    })
  } else {
    message.error("已经关注了")
  }
}


function removeMonitor(code, name, key) {
  //console.log("removeMonitor",name,code,key)
  stocks.value.splice(stocks.value.indexOf(code), 1)
  //console.log("removeMonitor-key",key)
  //console.log("removeMonitor-v",results.value[key])

  delete results.value[key]
  //console.log("removeMonitor-v",results.value[key])

  UnFollow(code).then(result => {
    message.success(result)
    monitor()
  })
}

// 一键清理所有自选股票
function clearAllStocks() {
  const stockCount = followList.value.length
  if (stockCount === 0) {
    message.warning('当前没有关注的股票')
    return
  }
  
  dialog.warning({
    title: '确认清理',
    content: `确定要取消关注所有 ${stockCount} 只股票吗？此操作不可恢复！`,
    positiveText: '确定清理',
    negativeText: '取消',
    onPositiveClick: async () => {
      message.loading('正在清理...', { duration: 0 })
      
      try {
        // 先获取原始列表（未转换的代码）
        const originalList = await GetFollowList(currentGroupId.value)
        
        // 获取所有股票代码，处理代码格式转换
        const allStockCodes = originalList.map(stock => {
          let code = stock.StockCode
          // 如果代码是 gb_ 格式（数据库中的美股格式），需要转换为 us 格式
          if (code.startsWith("gb_")) {
            code = "us" + code.replace("gb_", "").toUpperCase()
          }
          return code
        })
        
        // 批量取消关注
        const promises = allStockCodes.map(code => UnFollow(code))
        await Promise.all(promises)
        
        // 清空本地数据
        stocks.value = []
        results.value = {}
        followList.value = []
        
        // 重新获取列表（虽然应该是空的）
        await GetFollowList(currentGroupId.value).then(result => {
          followList.value = result
        })
        
        message.destroyAll()
        message.success(`已成功清理 ${stockCount} 只股票`)
        
        // 停止监控
        monitor()
      } catch (error) {
        message.destroyAll()
        message.error('清理过程中出现错误：' + (error.message || error))
      }
    }
  })
}

function SendDanmu() {
  //danmus.value.push(data.name)
  //console.log("SendDanmu",data.name)
  //console.log("SendDanmu-readyState", ws.value.readyState)
  ws.value.send(data.name)
}

function getStockList(value) {


  // //console.log("getStockList",value)
  let result;
  result = stockList.value.filter(item => item.name.includes(value) || item.ts_code.includes(value))
  options.value = result.map(item => {
    return {
      label: item.name + " - " + item.ts_code,
      value: item.ts_code
    }
  })
  if (value && value.indexOf("-") <= 0) {
    data.code = value
  }

  //console.log("getStockList-options",data.code)

  if (data.code) {
    let findId = data.code
    if (findId.startsWith("us")) {
      findId = "gb_" + findId.replace("us", "").toLowerCase()
    }
    blinkBorder(findId)
  }


}

function blinkBorder(findId) {
  // 获取要滚动到的元素
  let element = document.getElementById(findId);
  //console.log("blinkBorder",findId,element)
  if (element) {
    // 滚动到该元素
    element.scrollIntoView({behavior: 'smooth'});
    const pelement = document.getElementById(findId + '_gi');
    if (pelement) {
      // 添加闪烁效果
      pelement.classList.add('blink-border');
      // 3秒后移除闪烁效果
      setTimeout(() => {
        pelement.classList.remove('blink-border');
      }, 1000 * 5);
    } else {
      console.error(`Element with ID ${findId}_gi not found`);
    }
  }
}

async function updateData(result) {
  ////console.log("stock_price",result['日期'],result['时间'],result['股票代码'],result['股票名称'],result['当前价格'],result['盘前盘后'])

  if (result["当前价格"] <= 0) {
    result["当前价格"] = result["卖一报价"]
  }

  if (result.changePercent > 0) {
    result.type = "error"
    result.color = "#E88080"
  } else if (result.changePercent < 0) {
    result.type = "success"
    result.color = "#63E2B7"
  } else {
    result.type = "default"
    result.color = "#FFFFFF"
  }

  if (result.profitAmount > 0) {
    result.profitType = "error"
  } else if (result.profitAmount < 0) {
    result.profitType = "success"
  }
  if (result["当前价格"]) {
    if (result.alarmChangePercent > 0 && Math.abs(result.changePercent) >= result.alarmChangePercent) {
      SendMessage(result, 1)
    }

    if (result.alarmPrice > 0 && result["当前价格"] >= result.alarmPrice) {
      SendMessage(result, 2)
    }

    if (result.costPrice > 0 && result["当前价格"] >= result.costPrice) {
      SendMessage(result, 3)
    }
  }

  // result.key=result.sort
  results.value = Object.fromEntries(
    Object.entries(results.value).filter(
      ([key]) => !key.includes(result["股票代码"])
    ));

  result.key = GetSortKey(result.sort, result["股票代码"])
  results.value[result.key] = result
  if (!stocks.value.includes(result["股票代码"])) {
    delete results.value[result.key]
  }
}


async function monitor() {
  if (stocks.value && stocks.value.length === 0) {
    showPopover.value = true
  }
  for (let code of stocks.value) {
    Greet(code).then(result => {
      updateData(result)
    })
  }
}


function GetSortKey(sort, code) {
  return padStart(sort, 8, '0') + "_" + code
}

function onSelect(item) {
  ////console.log("onSelect",item)

  if (item.indexOf("-") > 0) {
    item = item.split("-")[1].toLowerCase()
  }
  if (item.indexOf(".") > 0) {
    data.code = item.split(".")[1].toLowerCase() + item.split(".")[0]
  }

}

function openCenteredWindow(url, width, height) {
  const left = (window.screen.width - width) / 2;
  const top = (window.screen.height - height) / 2;
  Environment().then(env => {
    switch (env.platform) {
      case 'windows':
        window.open(
          url,
          'centeredWindow',
          `width=${width},height=${height},left=${left},top=${top},location=no,menubar=no,toolbar=no,display=standalone`
        )
        break
      default :
        OpenURL(url)
        break
    }
  })


  //
  // return window.open(
  //     url,
  //     'centeredWindow',
  //     `width=${width},height=${height},left=${left},top=${top}`
  // );
}

function search(code, name) {
  setTimeout(() => {
    //window.open("https://xueqiu.com/S/"+code)
    //window.open("https://www.cls.cn/stock?code="+code)
    //window.open("https://quote.eastmoney.com/"+code+".html")
    //window.open("https://finance.sina.com.cn/realstock/company/"+code+"/nc.shtml")
    //window.open("https://www.iwencai.com/unifiedwap/result?w=" + name)
    //window.open("https://www.iwencai.com/chat/?question="+code)

    openCenteredWindow("https://www.iwencai.com/unifiedwap/result?w=" + name, 1000, 800)

  }, 500)
}

function setStock(code, name) {
  let res = followList.value.filter(item => item.StockCode === code)
  ////console.log("res:",res)
  formModel.value.name = name
  formModel.value.code = code
  formModel.value.volume = res[0].Volume ? res[0].Volume : 0
  formModel.value.costPrice = res[0].CostPrice
  formModel.value.buyDate = res[0].BuyDate ? new Date(res[0].BuyDate) : null
  formModel.value.alarm = res[0].AlarmChangePercent
  formModel.value.alarmPrice = res[0].AlarmPrice
  formModel.value.sort = res[0].Sort
  formModel.value.cron = res[0].Cron
  modalShow.value = true
}

function clearFeishi() {
  //console.log("clearFeishi")
  clearInterval(feishiInterval.value)
}

function showFsChart(code, name) {
  data.name = name
  data.code = code
  const chart = echarts.init(kLineChartRef2.value);
  GetStockMinutePriceLineData(code, name).then(result => {
    // console.log("GetStockMinutePriceLineData", result)
    const priceData = result.priceData
    let category = []
    let price = []
    let openprice = 0
    let closeprice = 0
    let volume = []
    let volumeRate = []
    let min = 0
    let max = 0
    openprice = priceData[0].price
    closeprice = priceData[priceData.length - 1].price
    for (let i = 0; i < priceData.length; i++) {
      category.push(priceData[i].time)
      price.push(priceData[i].price)
      if (min === 0 || min > priceData[i].price) {
        min = priceData[i].price
      }
      if (max < priceData[i].price) {
        max = priceData[i].price
      }
      if (i > 0) {
        let b = priceData[i].volume - priceData[i - 1].volume
        volumeRate.push(((b - volume[i - 1]) / volume[i - 1] * 100).toFixed(2))
        volume.push(b)
      } else {
        volume.push(priceData[i].volume)
        volumeRate.push(0)
      }
    }

    let option = {
      title: {
        subtext: "[" + result.date + "] 开盘:" + openprice + " 最新:" + closeprice + " 最高:" + max + " 最低:" + min,
        left: 'center',
        top: '10',
        textStyle: {
          color: data.darkTheme ? '#ccc' : '#456'
        }
      },
      legend: {
        data: ['股价', '成交量'],
        //orient: 'vertical',
        textStyle: {
          color: data.darkTheme ? '#ccc' : '#456'
        },
        right: 50,
      },
      darkMode: data.darkTheme,
      tooltip: {
        trigger: 'axis',
        axisPointer: {
          type: 'cross',
          animation: false,
          label: {
            backgroundColor: '#505765'
          }
        }
      },
      axisPointer: {
        link: [
          {
            xAxisIndex: 'all'
          }
        ],
        label: {
          backgroundColor: '#888'
        }
      },
      xAxis: [
        {
          type: 'category',
          data: category,
          axisLabel: {
            show: false
          }
        },
        {
          gridIndex: 1,
          type: 'category',
          data: category,
        },
      ],
      grid: [
        {
          left: '8%',
          right: '8%',
          height: '50%',
        },
        {
          left: '8%',
          right: '8%',
          top: '70%',
          height: '15%'
        },
      ],
      yAxis: [
        {
          axisLine: {
            show: true
          },
          splitLine: {
            show: false
          },
          name: "股价",
          min: (min - min * 0.01).toFixed(2),
          max: (max + max * 0.01).toFixed(2),
          minInterval: 0.01,
          type: 'value'
        },
        {
          gridIndex: 1,
          axisLine: {
            show: true
          },
          splitLine: {
            show: false
          },
          name: "成交量",
          type: 'value',
        },
      ],
      visualMap: {
        type: 'piecewise',
        seriesIndex: 0,
        top: 0,
        left: 10,
        orient: 'horizontal',
        textStyle: {
          color: data.darkTheme ? '#fff' : '#456'
        },
        pieces: [
          {
            text: '低于开盘价',
            gt: 0,
            lte: openprice,
            color: '#31F113',
            textStyle: {
              color: data.darkTheme ? '#fff' : '#456'
            },
          },
          {
            text: '大于开盘价小于收盘价',
            gt: openprice,
            lte: closeprice,
            color: '#1651EF',
            textStyle: {
              color: data.darkTheme ? '#fff' : '#456'
            },
          },
          {
            text: '大于收盘价',
            gt: closeprice,
            color: '#AC3B2A',
            textStyle: {
              color: data.darkTheme ? '#fff' : '#456'
            },
          }
        ],
      },
      series: [
        {
          name: "股价",
          data: price,
          type: 'line',
          smooth: false,
          showSymbol: false,
          lineStyle: {
            width: 3
          },
          markPoint: {
            symbol: 'arrow',
            symbolRotate: 90,
            symbolSize: [10, 20],
            symbolOffset: [10, 0],
            itemStyle: {
              color: '#FC290D'
            },
            label: {
              position: 'right',
            },
            data: [
              {type: 'max', name: 'Max'},
              {type: 'min', name: 'Min'}
            ]
          },
          markLine: {
            symbol: 'none',
            data: [
              {type: 'average', name: 'Average'},
              {
                lineStyle: {
                  color: '#FFCB00',
                  width: 0.5
                },
                yAxis: openprice,
                name: '开盘价'
              },
              {
                yAxis: closeprice,
                symbol: 'none',
                lineStyle: {
                  color: 'red',
                  width: 0.5
                },
              }
            ]
          },
        },
        {
          xAxisIndex: 1,
          yAxisIndex: 1,
          name: "成交量",
          data: volume,
          type: 'bar',
        },

      ]
    };
    chart.setOption(option);
  })
}

function showFenshi(code, name, changePercent) {
  data.code = code
  data.name = name
  data.changePercent = changePercent
  data.fenshiURL = 'http://image.sinajs.cn/newchart/min/n/' + data.code + '.gif' + "?t=" + Date.now()

  if (code.startsWith('hk')) {
    data.fenshiURL = 'http://image.sinajs.cn/newchart/hk_stock/min/' + data.code.replace("hk", "") + '.gif' + "?t=" + Date.now()
  }
  if (code.startsWith('gb_')) {
    data.fenshiURL = 'http://image.sinajs.cn/newchart/usstock/min/' + data.code.replace("gb_", "") + '.gif' + "?t=" + Date.now()
  }

  modalShow2.value = true
}

function handleFeishi() {
  showFsChart(data.code, data.name);
  feishiInterval.value = setInterval(() => {
    showFsChart(data.code, data.name);
  }, 1000 * 10)
}

function calculateMA(dayCount, values) {
  var result = [];
  for (var i = 0, len = values.length; i < len; i++) {
    if (i < dayCount) {
      result.push('-');
      continue;
    }
    var sum = 0;
    for (var j = 0; j < dayCount; j++) {
      sum += +values[i - j][1];
    }
    result.push((sum / dayCount).toFixed(2));
  }
  return result;
}

function handleKLine() {
  GetStockKLine(data.code, data.name, 365).then(result => {
    //console.log("GetStockKLine",result)
    const chart = echarts.init(kLineChartRef.value);
    const categoryData = [];
    const values = [];
    const volumns = [];
    for (let i = 0; i < result.length; i++) {
      let resultElement = result[i]
      //console.log("resultElement:{}",resultElement)
      categoryData.push(resultElement.day)
      let flag = resultElement.close > resultElement.open ? 1 : -1
      values.push([
        resultElement.open,
        resultElement.close,
        resultElement.low,
        resultElement.high
      ])
      volumns.push([i, resultElement.volume / 10000, flag])
    }
    ////console.log("categoryData",categoryData)
    ////console.log("values",values)
    let option = {
      darkMode: data.darkTheme,
      //backgroundColor: '#1c1c1c',
      // color:['#5470c6', '#91cc75', '#fac858', '#ee6666', '#73c0de', '#3ba272', '#fc8452', '#9a60b4', '#ea7ccc'],
      animation: false,
      legend: {
        bottom: 10,
        left: 'center',
        data: ['日K', 'MA5', 'MA10', 'MA20', 'MA30'],
        textStyle: {
          color: data.darkTheme ? '#ccc' : '#456'
        },
      },
      tooltip: {
        trigger: 'axis',
        axisPointer: {
          type: 'cross',
          lineStyle: {
            color: '#376df4',
            width: 1,
            opacity: 1
          }
        },
        borderWidth: 2,
        borderColor: data.darkTheme ? '#456' : '#ccc',
        backgroundColor: data.darkTheme ? '#456' : '#fff',
        padding: 10,
        textStyle: {
          color: data.darkTheme ? '#ccc' : '#456'
        },
        formatter: function (params) {//修改鼠标划过显示为中文
          //console.log("params",params)
          let volum = params[5].data;//ma5的值
          let ma5 = params[1].data;//ma5的值
          let ma10 = params[2].data;//ma10的值
          let ma20 = params[3].data;//ma20的值
          let ma30 = params[4].data;//ma30的值
          params = params[0];//开盘收盘最低最高数据汇总
          let currentItemData = params.data;

          return params.name + '<br>' +
            '开盘:' + currentItemData[1] + '<br>' +
            '收盘:' + currentItemData[2] + '<br>' +
            '最低:' + currentItemData[3] + '<br>' +
            '最高:' + currentItemData[4] + '<br>' +
            '成交量(万手):' + volum[1] + '<br>' +
            'MA5日均线:' + ma5 + '<br>' +
            'MA10日均线:' + ma10 + '<br>' +
            'MA20日均线:' + ma20 + '<br>' +
            'MA30日均线:' + ma30
        }
        // position: function (pos, params, el, elRect, size) {
        //   const obj = {
        //     top: 10
        //   };
        //   obj[['left', 'right'][+(pos[0] < size.viewSize[0] / 2)]] = 30;
        //   return obj;
        // }
        // extraCssText: 'width: 170px'
      },
      axisPointer: {
        link: [
          {
            xAxisIndex: 'all'
          }
        ],
        label: {
          backgroundColor: '#888'
        }
      },
      visualMap: {
        show: false,
        seriesIndex: 5,
        dimension: 2,
        pieces: [
          {
            value: -1,
            color: downColor
          },
          {
            value: 1,
            color: upColor
          }
        ]
      },
      grid: [
        {
          left: '10%',
          right: '8%',
          height: '50%',
        },
        {
          left: '10%',
          right: '8%',
          top: '63%',
          height: '16%'
        }
      ],
      xAxis: [
        {
          type: 'category',
          data: categoryData,
          boundaryGap: false,
          axisLine: {onZero: false},
          splitLine: {show: false},
          min: 'dataMin',
          max: 'dataMax',
          axisPointer: {
            z: 100
          }
        },
        {
          type: 'category',
          gridIndex: 1,
          data: categoryData,
          boundaryGap: false,
          axisLine: {onZero: false},
          axisTick: {show: false},
          splitLine: {show: false},
          axisLabel: {show: false},
          min: 'dataMin',
          max: 'dataMax'
        }
      ],
      yAxis: [
        {
          scale: true,
          splitArea: {
            show: true
          }
        },
        {
          scale: true,
          gridIndex: 1,
          splitNumber: 2,
          axisLabel: {show: false},
          axisLine: {show: false},
          axisTick: {show: false},
          splitLine: {show: false}
        }
      ],
      dataZoom: [
        {
          type: 'inside',
          xAxisIndex: [0, 1],
          start: 86,
          end: 100
        },
        {
          show: true,
          xAxisIndex: [0, 1],
          type: 'slider',
          top: '85%',
          start: 86,
          end: 100
        }
      ],

      series: [
        {
          name: '日K',
          type: 'candlestick',
          data: values,
          itemStyle: {
            color: upColor,
            color0: downColor,
            // borderColor: upBorderColor,
            // borderColor0: downBorderColor
          },
          markPoint: {
            label: {
              formatter: function (param) {
                return param != null ? param.value + '' : '';
              }
            },
            data: [
              {
                name: '最高',
                type: 'max',
                valueDim: 'highest'
              },
              {
                name: '最低',
                type: 'min',
                valueDim: 'lowest'
              },
              {
                name: '平均收盘价',
                type: 'average',
                valueDim: 'close'
              }
            ],
            tooltip: {
              formatter: function (param) {
                return param.name + '<br>' + (param.data.coord || '');
              }
            }
          },
          markLine: {
            symbol: ['none', 'none'],
            data: [
              [
                {
                  name: 'from lowest to highest',
                  type: 'min',
                  valueDim: 'lowest',
                  symbol: 'circle',
                  symbolSize: 10,
                  label: {
                    show: false
                  },
                  emphasis: {
                    label: {
                      show: false
                    }
                  }
                },
                {
                  type: 'max',
                  valueDim: 'highest',
                  symbol: 'circle',
                  symbolSize: 10,
                  label: {
                    show: false
                  },
                  emphasis: {
                    label: {
                      show: false
                    }
                  }
                }
              ],
              {
                name: 'min line on close',
                type: 'min',
                valueDim: 'close'
              },
              {
                name: 'max line on close',
                type: 'max',
                valueDim: 'close'
              }
            ]
          }
        },
        {
          name: 'MA5',
          type: 'line',
          data: calculateMA(5, values),
          smooth: true,
          showSymbol: false,
          lineStyle: {
            opacity: 0.6
          }
        },
        {
          name: 'MA10',
          type: 'line',
          data: calculateMA(10, values),
          smooth: true,
          showSymbol: false,
          lineStyle: {
            opacity: 0.6
          }
        },
        {
          name: 'MA20',
          type: 'line',
          data: calculateMA(20, values),
          smooth: true,
          showSymbol: false,
          lineStyle: {
            opacity: 0.6
          }
        },
        {
          name: 'MA30',
          type: 'line',
          data: calculateMA(30, values),
          smooth: true,
          showSymbol: false,
          lineStyle: {
            opacity: 0.6
          }
        },
        {
          name: '成交量(手)',
          type: 'bar',
          xAxisIndex: 1,
          yAxisIndex: 1,
          itemStyle: {
            color: '#7fbe9e'
          },
          data: volumns
        }
      ]
    };
    chart.setOption(option);
    chart.on('click', {seriesName: '日K'}, function (params) {
      //console.log("click:",params);
    });
  })
}

function showMoney(code, name) {
  data.code = code
  data.name = name
  modalShow5.value = true
}

function showK(code, name) {
  data.code = code
  data.name = name
  data.kURL = 'http://image.sinajs.cn/newchart/daily/n/' + data.code + '.gif' + "?t=" + Date.now()
  if (code.startsWith('hk')) {
    data.kURL = 'http://image.sinajs.cn/newchart/hk_stock/daily/' + data.code.replace("hk", "") + '.gif' + "?t=" + Date.now()
  }
  if (code.startsWith('gb_')) {
    data.kURL = 'http://image.sinajs.cn/newchart/usstock/daily/' + data.code.replace("gb_", "") + '.gif' + "?t=" + Date.now()
  }
  modalShow3.value = true
  //https://image.sinajs.cn/newchart/usstock/daily/dji.gif
  //https://image.sinajs.cn/newchart/hk_stock/daily/06030.gif?1740729404273
}

function showWeekK(code, name) {
  data.code = code
  data.name = name
  data.kURL = 'http://image.sinajs.cn/newchart/weekly/n/' + data.code + '.gif' + "?t=" + Date.now()
  if (code.startsWith('hk')) {
    data.kURL = 'http://image.sinajs.cn/newchart/hk_stock/weekly/' + data.code.replace("hk", "") + '.gif' + "?t=" + Date.now()
  }
  if (code.startsWith('gb_')) {
    data.kURL = 'http://image.sinajs.cn/newchart/usstock/weekly/' + data.code.replace("gb_", "") + '.gif' + "?t=" + Date.now()
  }
  modalShow3.value = true
}

function showMonthK(code, name) {
  data.code = code
  data.name = name
  data.kURL = 'http://image.sinajs.cn/newchart/monthly/n/' + data.code + '.gif' + "?t=" + Date.now()
  if (code.startsWith('hk')) {
    data.kURL = 'http://image.sinajs.cn/newchart/hk_stock/monthly/' + data.code.replace("hk", "") + '.gif' + "?t=" + Date.now()
  }
  if (code.startsWith('gb_')) {
    data.kURL = 'http://image.sinajs.cn/newchart/usstock/monthly/' + data.code.replace("gb_", "") + '.gif' + "?t=" + Date.now()
  }
  modalShow3.value = true
}


function updateCostPriceAndVolumeNew(code, price, volume, alarm, formModel) {
  if (formModel.sort) {
    SetStockSort(formModel.sort, code).then(result => {
      //message.success(result)
    })
  }
  if (formModel.cron) {
    SetStockAICron(formModel.cron, code).then(result => {
      //message.success(result)
    })
  }

  if (alarm || formModel.alarmPrice) {
    SetAlarmChangePercent(alarm, formModel.alarmPrice, code).then(result => {
      //message.success(result)
    })
  }
  // 转换买入日期为时间戳或null
  let buyDateValue = null
  if (formModel.buyDate) {
    buyDateValue = new Date(formModel.buyDate)
  }
  SetCostPriceAndVolume(code, price, volume, buyDateValue).then(result => {
    modalShow.value = false
    message.success(result)
    GetFollowList(currentGroupId.value).then(result => {
      followList.value = result
      stocks.value = []
      for (const followedStock of result) {
        if (!stocks.value.includes(followedStock.StockCode)) {
          stocks.value.push(followedStock.StockCode)
        }
      }
      monitor()
      message.destroyAll()
    })
  })
}

function fullscreen() {
  if (data.fullscreen) {
    WindowUnfullscreen()
  } else {
    WindowFullscreen()
  }
  data.fullscreen = !data.fullscreen
}


//type 报警类型: 1 涨跌报警;2 股价报警 3 成本价报警
function SendMessage(result, type) {
  let typeName = getTypeName(type)
  let img = 'http://image.sinajs.cn/newchart/min/n/' + result["股票代码"] + '.gif' + "?t=" + Date.now()
  let markdown = "### go-stock [" + typeName + "]\n\n" +
    "### " + result["股票名称"] + "(" + result["股票代码"] + ")\n" +
    "- 当前价格: " + result["当前价格"] + "  " + result.changePercent + "%\n" +
    "- 最高价: " + result["今日最高价"] + "  " + result.highRate + "\n" +
    "- 最低价: " + result["今日最低价"] + "  " + result.lowRate + "\n" +
    "- 昨收价: " + result["昨日收盘价"] + "\n" +
    "- 今开价: " + result["今日开盘价"] + "\n" +
    "- 成本价: " + result.costPrice + "  " + result.profit + "%  " + result.profitAmount + " ¥\n" +
    "- 成本数量: " + result.costVolume + "股\n" +
    "- 日期: " + result["日期"] + "  " + result["时间"] + "\n\n" +
    "![image](" + img + ")\n"
  let title = result["股票名称"] + "(" + result["股票代码"] + ") " + result["当前价格"] + " " + result.changePercent

  let msg = '{' +
    '     "msgtype": "markdown",' +
    '     "markdown": {' +
    '         "title":"[' + typeName + "]" + title + '",' +
    '         "text": "' + markdown + '"' +
    '     },' +
    '      "at": {' +
    '          "isAtAll": true' +
    '      }' +
    ' }'
  // SendDingDingMessage(msg,result["股票代码"])
  SendDingDingMessageByType(msg, result["股票代码"], type)
}

function aiReCheckStock(stock, stockCode) {
  data.modelName = ""
  data.airesult = ""
  data.time = ""
  data.name = stock
  data.code = stockCode
  data.loading = true
  modalShow4.value = true
  message.loading("ai检测中...", {
    duration: 0,
  })
  //

  //message.info("sysPromptId:"+data.sysPromptId)
  NewChatStream(stock, stockCode, data.question, data.aiConfigId, data.sysPromptId, enableTools.value,thinkingMode.value)
}

function aiCheckStock(stock, stockCode) {
  GetAIResponseResult(stockCode).then(result => {
    if (result.content) {
      data.modelName = result.modelName
      data.chatId = result.chatId
      data.question = result.question
      data.name = stock
      data.code = stockCode
      data.loading = false
      modalShow4.value = true
      data.airesult = result.content
      const date = new Date(result.CreatedAt);
      const year = date.getFullYear();
      const month = String(date.getMonth() + 1).padStart(2, '0');
      const day = String(date.getDate()).padStart(2, '0');
      const hours = String(date.getHours()).padStart(2, '0');
      const minutes = String(date.getMinutes()).padStart(2, '0');
      const seconds = String(date.getSeconds()).padStart(2, '0');
      data.time = `${year}-${month}-${day} ${hours}:${minutes}:${seconds}`
    } else {
      data.modelName = ""
      data.question = ""
      data.airesult = ""
      data.time = ""
      data.name = stock
      data.code = stockCode
      data.loading = false
      modalShow4.value = true
      // message.loading("ai检测中...", {
      //   duration: 0,
      // })
      // NewChatStream(stock, stockCode, "", data.sysPromptId)
    }
  })
}

function getTypeName(type) {
  switch (type) {
    case 1:
      return "涨跌报警"
    case 2:
      return "股价报警"
    case 3:
      return "成本价报警"
    default:
      return ""
  }
}

//获取高度
function getHeight() {
  return document.documentElement.clientHeight
}

window.onerror = function (msg, source, lineno, colno, error) {
  // 将错误信息发送给后端
  EventsEmit("frontendError", {
    page: "stock.vue",
    message: msg,
    source: source,
    lineno: lineno,
    colno: colno,
    error: error ? error.stack : null,
    data: data,
    results: results,
    followList: followList,
    stockList: stockList,
    stocks: stocks,
    formModel: formModel,
  });
  message.error("发生错误:" + msg)
  return true;
};

function saveAsImage(name, code) {
  Environment().then(env => {
    switch (env.platform) {
      case 'windows':
        const element = document.querySelector('.md-editor-preview');
        if (element) {
          html2canvas(element, {
            useCORS: true, // 解决跨域图片问题
            scale: 2, // 提高截图质量
            allowTaint: true, // 允许跨域图片
          }).then(canvas => {
            const link = document.createElement('a');
            link.href = canvas.toDataURL('image/png');
            link.download = name + "[" + code + ']-ai-analysis-result.png';
            link.click();
          });
        } else {
          message.error('无法找到分析结果元素');
        }
        break
      default :
        saveCanvasImage(name)
    }
  })
}

async function saveCanvasImage(name) {
  const element = document.querySelector('.md-editor-preview'); // 要截图的 DOM 节点
  const canvas = await html2canvas(element)

  const dataUrl = canvas.toDataURL('image/png') // base64 格式
  const base64 = dataUrl.replace(/^data:image\/png;base64,/, '')

  // 调用 Go 后端保存文件（Wails 绑定方法）
  await SaveImage(name, base64).then(result => {
    message.success(result)
  })
}

async function copyToClipboard() {
  try {
    await navigator.clipboard.writeText(data.airesult);
    message.success('分析结果已复制到剪切板');
  } catch (err) {
    message.error('复制失败: ' + err);
  }
}

function saveAsMarkdown() {
  SaveAsMarkdown(data.code, data.name).then(result => {
    message.success(result)
  })
}

function saveAsMarkdown_old() {
  const blob = new Blob([data.airesult], {type: 'text/markdown;charset=utf-8'});
  const link = document.createElement('a');
  link.href = URL.createObjectURL(blob);
  link.download = `${data.name}[${data.code}]-${data.time}ai-analysis-result.md`;
  link.click();
  URL.revokeObjectURL(link.href);
  link.remove()
}

function getHtml(ref) {
  if (ref.value) {
    // 获取 MdPreview 组件的根元素
    const rootElement = ref.value.$el;
    // 获取 HTML 内容
    return rootElement.innerHTML;
  } else {
    console.error('mdPreviewRef is not yet available');
    return "";
  }
}

// 导出文档
async function saveAsWord() {
  // 将富文本内容拼接为一个完整的html
  const html = getHtml(mdPreviewRef)
  const tipsHtml = getHtml(tipsRef)
  const value = `
         ${html}
         <hr>
         <div style="font-size: 12px;color: red">
         ${tipsHtml}
          </div>
<br>
本报告由go-stock项目生成：
<p>
<a href="https://github.com/ArvinLovegood/go-stock">
AI赋能股票分析：自选股行情获取，成本盈亏展示，涨跌报警推送，市场整体/个股情绪分析，K线技术指标分析等。数据全部保留在本地。支持DeepSeek，OpenAI， Ollama，LMStudio，AnythingLLM，硅基流动，火山方舟，阿里云百炼等平台或模型。
</a></p>
`
  // landscape就是横着的，portrait是竖着的，默认是竖屏portrait。
  const blob = await asBlob(value, {orientation: 'portrait'})
  const {platform} = await Environment()
  switch (platform) {
    case 'windows':
      const a = document.createElement('a')
      a.href = URL.createObjectURL(blob)
      a.download = `${data.name}[${data.code}]-ai-analysis-result.docx`;
      a.click()
      // 下载后将标签移除
      URL.revokeObjectURL(a.href);
      a.remove()
      break
    default:
      const arrayBuffer = await blob.arrayBuffer()
      const uint8Array = new Uint8Array(arrayBuffer)
      const binary = uint8Array.reduce((data, byte) => data + String.fromCharCode(byte), '')
      const base64 = btoa(binary)
      await SaveWordFile(`${data.name}[${data.code}]-ai-analysis-result.docx`, base64).then(result => {
        message.success(result)
      })
  }
}

function share(code, name) {
  ShareAnalysis(code, name).then(msg => {
    //message.info(msg)
    notify.info({
      avatar: () =>
        h(NAvatar, {
          size: 'small',
          round: false,
          src: icon.value
        }),
      title: '分享到社区',
      duration: 1000 * 30,
      content: () => {
        return h('div', {
          style: {
            'text-align': 'left',
            'font-size': '14px',
          }
        }, {default: () => msg})
      },
    })
  })
}

const addTabModel = ref({
  name: '',
  sort: 1,
})
const addTabPane = ref(false)

// AI总结相关函数
function reAiSummary() {
  aiSummary.value = ""
  summaryModal.value = true
  aiSummaryLoading.value = true
  
  // 保存prompt到历史记录
  if (aiSummaryQuestion.value && aiSummaryQuestion.value.trim()) {
    savePromptToHistory(aiSummaryQuestion.value, 'AI总结')
  }
  
  SummaryStockNews(aiSummaryQuestion.value, aiSummaryConfigId.value, aiSummarySysPromptId.value, enableTools.value, thinkingMode.value)
}

function getAiSummary() {
  summaryModal.value = true
  aiSummaryLoading.value = true
  GetAIResponseResult("股票自选").then(result => {
    aiSummaryLoading.value = false
    if (result.content) {
      aiSummary.value = result.content
      aiSummaryQuestion.value = result.question
      aiSummaryLoading.value = false

      const date = new Date(result.CreatedAt);
      const year = date.getFullYear();
      const month = String(date.getMonth() + 1).padStart(2, '0');
      const day = String(date.getDate()).padStart(2, '0');
      const hours = String(date.getHours()).padStart(2, '0');
      const minutes = String(date.getMinutes()).padStart(2, '0');
      const seconds = String(date.getSeconds()).padStart(2, '0');
      aiSummaryTime.value = `${year}-${month}-${day} ${hours}:${minutes}:${seconds}`
      aiSummaryModelName.value = result.modelName
    } else {
      aiSummaryTime.value = ""
      aiSummary.value = ""
      aiSummaryModelName.value = ""
    }
  })
}

async function copyAiSummaryToClipboard() {
  try {
    await navigator.clipboard.writeText(aiSummary.value);
    message.success('分析结果已复制到剪切板');
  } catch (err) {
    message.error('复制失败: ' + err);
  }
}

function saveAiSummaryAsMarkdown() {
  SaveAsMarkdown('股票自选', '股票自选').then(result => {
    message.success(result)
  })
}

function shareAiSummary() {
  ShareAnalysis('股票自选', '股票自选').then(msg => {
    notify.info({
      avatar: () =>
        h(NAvatar, {
          size: 'small',
          round: false,
          src: icon.value
        }),
      title: '分享到社区',
      duration: 1000 * 30,
      content: () => {
        return h('div', {
          style: {
            'text-align': 'left',
            'font-size': '14px',
          }
        }, {default: () => msg})
      },
    })
  })
}

// AI选股相关函数
function openAiStockSelect() {
  aiStockSelectModal.value = true
  aiStockSelectCondition.value = ''
  aiStockSelectResult.value = ''
  aiStockSelectLoading.value = false
  aiStockSelectChatId.value = ''
  aiStockSelectModelName.value = ''
  aiStockSelectTime.value = ''
  isAiStockSelectMode.value = false // 重置标志
  // 初始化AI配置
  if (aiConfigs.value.length > 0) {
    aiStockSelectConfigId.value = data.aiConfigId || aiConfigs.value[0].ID
  }
  aiStockSelectSysPromptId.value = data.sysPromptId
}

// 执行AI选股
function executeAiStockSelect() {
  if (!aiStockSelectCondition.value.trim()) {
    message.warning('请输入选股条件')
    return
  }
  
  if (!aiStockSelectConfigId.value) {
    message.warning('请选择AI模型配置')
    return
  }
  
  // 确保对话框打开
  aiStockSelectModal.value = true
  aiStockSelectLoading.value = true
  aiStockSelectResult.value = ''
  aiStockSelectChatId.value = ''
  aiStockSelectModelName.value = ''
  aiStockSelectTime.value = ''
  isAiStockSelectMode.value = true // 设置标志为AI选股模式
  
  // 构建选股prompt
  const stockSelectPrompt = `请根据以下选股条件筛选股票：${aiStockSelectCondition.value}

**重要要求：**
1. 请使用可用的工具函数（如SearchStockByIndicators、GetStockKLine等）来获取股票数据
2. 根据选股条件进行筛选和分析
3. 在回复的最后，必须按照以下JSON格式返回筛选结果

**返回格式（必须严格遵守）：**

\`\`\`json
{
  "stocks": [
    {
      "code": "sh000001",
      "name": "平安银行",
      "market": "A股"
    },
    {
      "code": "sz000002",
      "name": "万科A",
      "market": "A股"
    }
  ]
}
\`\`\`

**格式要求：**
1. 股票代码格式：
   - A股：sh000001（上海）或 sz000001（深圳），代码必须是6位数字
   - 港股：hk00700，代码必须是5位数字
   - 美股：usAAPL 或 gb_aapl（小写）
2. 股票名称必须是完整的中文名称（A股、港股）或英文名称（美股）
3. market字段：A股、港股、美股
4. 如果未找到符合条件的股票，返回：{"stocks": []}
5. JSON格式必须正确，可以被解析
6. JSON代码块必须放在回复的最后

**工作流程：**
1. 使用工具函数获取股票数据
2. 根据选股条件筛选股票（如：市值最大的3个、市盈率小于20等）
3. 在回复的最后提供JSON格式的筛选结果

请开始筛选。`

  // 保存prompt到历史记录
  if (aiStockSelectCondition.value && aiStockSelectCondition.value.trim()) {
    savePromptToHistory(aiStockSelectCondition.value, 'AI选股')
  }
  
  // 调用AI分析
  SummaryStockNews(stockSelectPrompt, aiStockSelectConfigId.value, aiStockSelectSysPromptId.value, enableTools.value, thinkingMode.value)
}

// 从AI结果中解析股票JSON数据
function parseStocksFromAiResult(content) {
  const stocks = []
  try {
    // 方法1: 尝试从JSON代码块中解析
    const jsonBlockRegex = /```json\s*([\s\S]*?)\s*```/i
    const jsonMatch = content.match(jsonBlockRegex)
    if (jsonMatch) {
      const jsonStr = jsonMatch[1].trim()
      const data = JSON.parse(jsonStr)
      if (data.stocks && Array.isArray(data.stocks)) {
        return data.stocks
      }
    }
    
    // 方法2: 尝试直接查找JSON对象（不在代码块中）
    const jsonObjRegex = /\{\s*"stocks"\s*:\s*\[([\s\S]*?)\]\s*\}/i
    const jsonObjMatch = content.match(jsonObjRegex)
    if (jsonObjMatch) {
      // 尝试提取完整的JSON
      const jsonStart = content.lastIndexOf('{')
      const jsonEnd = content.lastIndexOf('}') + 1
      if (jsonStart >= 0 && jsonEnd > jsonStart) {
        const jsonStr = content.substring(jsonStart, jsonEnd)
        const data = JSON.parse(jsonStr)
        if (data.stocks && Array.isArray(data.stocks)) {
          return data.stocks
        }
      }
    }
  } catch (error) {
    console.error('解析股票数据失败:', error)
    message.error('解析股票数据失败，请检查AI返回的格式')
  }
  
  return stocks
}

// 使用AI生成分组名称（从选股结果中提取或根据条件生成）
function generateGroupNameByAI(stocks, condition, aiResult) {
  // 方法1: 尝试从AI返回结果中提取分组名称建议
  if (aiResult) {
    // 查找类似"分组名称"、"建议名称"等关键词
    const namePatterns = [
      /分组名称[：:：]\s*([^\n]{2,10})/,
      /建议名称[：:：]\s*([^\n]{2,10})/,
      /推荐名称[：:：]\s*([^\n]{2,10})/,
      /名称[：:：]\s*([^\n]{2,10})/,
    ]
    
    for (const pattern of namePatterns) {
      const match = aiResult.match(pattern)
      if (match && match[1]) {
        let name = match[1].trim().replace(/[：:：\s]+/g, '').substring(0, 10)
        if (name.length >= 2) {
          return name
        }
      }
    }
  }
  
  // 方法2: 根据选股条件智能生成名称
  let groupName = ''
  const conditionLower = condition.toLowerCase()
  
  if (conditionLower.includes('市值最大') || conditionLower.includes('大市值')) {
    const numMatch = condition.match(/(\d+)/)
    groupName = numMatch ? `市值Top${numMatch[1]}` : '高市值股票'
  } else if (conditionLower.includes('市值最小') || conditionLower.includes('小市值')) {
    const numMatch = condition.match(/(\d+)/)
    groupName = numMatch ? `小市值${numMatch[1]}只` : '小市值股票'
  } else if (conditionLower.includes('市盈率') && (conditionLower.includes('小于') || conditionLower.includes('低于'))) {
    groupName = '低估值股票'
  } else if (conditionLower.includes('市盈率') && (conditionLower.includes('大于') || conditionLower.includes('高于'))) {
    groupName = '高估值股票'
  } else if (conditionLower.includes('涨幅') && (conditionLower.includes('最大') || conditionLower.includes('最高'))) {
    const numMatch = condition.match(/(\d+)/)
    groupName = numMatch ? `涨幅Top${numMatch[1]}` : '强势股票'
  } else if (conditionLower.includes('跌幅') || conditionLower.includes('下跌')) {
    groupName = '弱势股票'
  } else if (conditionLower.includes('行业')) {
    groupName = '行业精选'
  } else if (conditionLower.includes('概念')) {
    groupName = '概念板块'
  } else if (conditionLower.includes('roe') || conditionLower.includes('净资产收益率')) {
    groupName = '高ROE股票'
  } else if (conditionLower.includes('营收') || conditionLower.includes('收入')) {
    groupName = '高增长股票'
  } else {
    // 默认名称：根据股票数量
    const stockCount = stocks.length
    groupName = `AI选股${stockCount}只`
  }
  
  return groupName || `AI选股-${new Date().toLocaleDateString()}`
}

// 添加解析出的股票到自选
async function addParsedStocksToFollow(parsedStocks) {
  console.log('addParsedStocksToFollow 被调用，股票数量:', parsedStocks?.length)
  if (!parsedStocks || parsedStocks.length === 0) {
    message.warning('未找到符合条件的股票')
    return
  }
  
  // 使用AI生成分组名称（从选股结果中提取或根据条件生成）
  const groupName = generateGroupNameByAI(parsedStocks, aiStockSelectCondition.value, aiStockSelectResult.value)
  console.log('生成的分组名称:', groupName)
  
  // 创建新分组
  const maxSort = groupList.value.length > 0 
    ? Math.max(...groupList.value.map(g => g.sort)) + 1 
    : 1
  
  message.loading('正在创建分组...')
  const createGroupResult = await AddGroup({
    name: groupName,
    sort: maxSort
  })
  message.destroyAll()
  
  if (!createGroupResult) {
    message.error('创建分组失败')
    return
  }
  
  // 刷新分组列表并获取新创建的分组ID
  const updatedGroups = await GetGroupList()
  groupList.value = updatedGroups
  const newGroup = updatedGroups.find(g => g.name === groupName && g.sort === maxSort)
  if (!newGroup) {
    message.error('无法找到新创建的分组')
    return
  }
  
  const newGroupId = newGroup.ID
  console.log('新分组ID:', newGroupId, '分组名称:', groupName)
  
  let successCount = 0
  let failCount = 0
  const failMessages = []
  const addedStocks = [] // 记录成功添加的股票代码，用于后续刷新
  
  // 添加股票到自选并分组
  for (const stock of parsedStocks) {
    try {
      let code = stock.code
      if (!code) {
        console.warn('股票数据缺少code字段:', stock)
        failCount++
        continue
      }
      
      // 转换股票代码格式
      code = code.trim().toLowerCase()
      console.log('处理股票代码:', code, '原始数据:', stock)
      
      // 如果是6位数字的A股代码，需要根据代码判断市场
      if (/^\d{6}$/.test(code)) {
        const firstDigit = code[0]
        // 6开头是上海，0/3开头是深圳
        if (firstDigit === '6') {
          code = 'sh' + code
        } else if (firstDigit === '0' || firstDigit === '3') {
          code = 'sz' + code
        }
      } else if (code.startsWith('gb_')) {
        // 美股gb_格式转换为us格式
        code = 'us' + code.replace('gb_', '').toUpperCase()
      }
      
      console.log('准备添加股票，代码:', code)
      
      // 先添加到自选
      const followResult = await Follow(code)
      console.log('Follow结果:', followResult, '代码:', code)
      
      if (followResult === "关注成功") {
        // 添加到分组
        let groupCode = code
        if (code.startsWith("gb_")) {
          groupCode = "us" + code.replace("gb_", "").toLowerCase()
        }
        const groupResult = await AddStockGroup(newGroupId, groupCode)
        console.log('AddStockGroup结果:', groupResult, '代码:', groupCode)
        
        if (groupResult === '添加成功' || groupResult === '已经关注了') {
          successCount++
          addedStocks.push(code)
          // 更新stocks数组（这是组件中的ref）
          if (!stocks.value.includes(code)) {
            stocks.value.push(code)
          }
        } else {
          failCount++
          failMessages.push(`${stock.name || code}: ${groupResult}`)
        }
      } else if (followResult === "已经关注了") {
        // 如果已经关注，也尝试添加到分组
        let groupCode = code
        if (code.startsWith("gb_")) {
          groupCode = "us" + code.replace("gb_", "").toLowerCase()
        }
        const groupResult = await AddStockGroup(newGroupId, groupCode)
        if (groupResult === '添加成功' || groupResult === '已经关注了') {
          successCount++
          addedStocks.push(code)
        } else {
          failCount++
          failMessages.push(`${stock.name || code}: ${groupResult}`)
        }
      } else {
        failCount++
        failMessages.push(`${stock.name || code}: ${followResult}`)
        console.warn('添加股票失败:', code, followResult)
      }
    } catch (error) {
      failCount++
      failMessages.push(`${stock.name || stock.code}: ${error.message}`)
      console.error('添加股票异常:', stock, error)
    }
  }
  
  // 刷新分组列表
  await GetGroupList().then(result => {
    groupList.value = result
  })
  
  // 刷新自选列表并切换到新分组
  await GetFollowList(newGroupId).then(result => {
    followList.value = result
    // 更新stocks数组
    stocks.value = []
    for (const followedStock of result) {
      let code = followedStock.StockCode
      if (code.startsWith("us")) {
        code = "gb_" + code.replace("us", "").toLowerCase()
      }
      stocks.value.push(code)
      // 获取股票数据
      Greet(code).then(result => {
        updateData(result)
      })
    }
    // 切换到新分组
    currentGroupId.value = newGroupId
    message.destroyAll()
  })
  
  // 显示结果
  if (successCount > 0) {
    message.success(`成功添加 ${successCount} 只股票到分组"${groupName}"${failCount > 0 ? `，失败 ${failCount} 只` : ''}`)
    if (failCount > 0 && failMessages.length > 0) {
      console.warn('添加失败的股票:', failMessages)
    }
  } else {
    message.warning(`添加失败，共 ${failCount} 只股票`)
    if (failMessages.length > 0) {
      message.error(failMessages.join('; '))
    }
  }
}

function addTab() {
  addTabPane.value = true
}

function saveTabPane() {
  AddGroup(addTabModel.value).then(result => {
    message.info(result)
    addTabPane.value = false
    GetGroupList().then(result => {
      groupList.value = result
    })
  })
}

function AddStockGroupInfo(groupId, code, name) {
  if (code.startsWith("gb_")) {
    code = "us" + code.replace("gb_", "").toLowerCase()
  }
  AddStockGroup(groupId, code).then(result => {
    message.info(result)
    GetGroupList().then(result => {
      groupList.value = result
    })
  })

}

// 开始编辑分组名称
function startEditGroup(groupId, currentName) {
  editingGroupId.value = groupId
  editingGroupName.value = currentName
}

// 保存分组名称
function saveGroupName(groupId) {
  // 如果正在保存，直接返回，避免重复触发
  if (isSavingGroupName.value) {
    return
  }
  
  // 如果已经不在编辑状态，直接返回
  if (editingGroupId.value !== groupId) {
    return
  }
  
  // 先保存当前值到局部变量，避免后续被清空
  const currentName = editingGroupName.value ? editingGroupName.value.trim() : ''
  
  if (!currentName) {
    message.warning('分组名称不能为空')
    cancelEditGroup()
    return
  }
  
  // 设置保存标志
  isSavingGroupName.value = true
  
  // 先取消编辑状态，避免重复触发
  cancelEditGroup()
  
  UpdateGroup(groupId, currentName).then(result => {
    isSavingGroupName.value = false
    if (result === '更新成功') {
      message.success('分组名称已更新')
      // 更新本地分组列表
      const group = groupList.value.find(g => g.ID === groupId)
      if (group) {
        group.name = currentName
      }
      // 刷新分组列表
      GetGroupList().then(result => {
        groupList.value = result
      })
    } else {
      message.error('更新失败：' + result)
    }
  }).catch(error => {
    isSavingGroupName.value = false
    message.error('更新失败：' + error.message)
  })
}

// 取消编辑分组名称
function cancelEditGroup() {
  editingGroupId.value = null
  editingGroupName.value = ''
}

// 指标选股相关函数
function openIndicatorStockSelect() {
  indicatorStockSelectModal.value = true
  indicatorStockSelectCondition.value = ''
  indicatorStockSelectColumns.value = []
  indicatorStockSelectDataList.value = []
  indicatorStockSelectTraceInfo.value = ''
  indicatorStockSelectLoading.value = false
}

// 计算表格总宽度
function calculateIndicatorTableWidth(cols) {
  let totalWidth = 0;
  
  cols.forEach(col => {
    if (col.children && col.children.length > 0) {
      // 有子列的情况
      let childrenWidth = 0;
      col.children.forEach(child => {
        childrenWidth += child.width || child.minWidth || 100;
      });
      // 取标题列宽度和子列总宽度的较大值
      totalWidth += Math.max(col.width || col.minWidth || 200, childrenWidth);
    } else {
      // 没有子列的情况
      totalWidth += col.width || col.minWidth || 120;
    }
  });
  
  // 加上操作列的宽度
  totalWidth += 100;
  
  return Math.max(totalWidth, 1200); // 最小宽度1200
}

// 判断是否为数字
function isNumeric(value) {
  return !isNaN(parseFloat(value)) && isFinite(value);
}

// 执行指标选股搜索
function executeIndicatorStockSelect() {
  if (!indicatorStockSelectCondition.value.trim()) {
    message.warning('请输入选股指标或者要求')
    return
  }

  indicatorStockSelectLoading.value = true
  const loading = message.loading("正在获取选股数据...", {duration: 0});
  
  SearchStock(indicatorStockSelectCondition.value).then(res => {
    loading.destroy()
    indicatorStockSelectLoading.value = false
    
    if (res.code == 100) {
      indicatorStockSelectTraceInfo.value = res.data.traceInfo.showText
      
      // 处理表格列
      indicatorStockSelectColumns.value = res.data.result.columns
        .filter(item => !item.hiddenNeed && (item.title != "市场码" && item.title != "市场简称"))
        .map(item => {
          if (item.children) {
            return {
              title: item.title + (item.unit ? '[' + item.unit + ']' : ''),
              key: item.key,
              resizable: true,
              minWidth: 200,
              ellipsis: {
                tooltip: true
              },
              children: item.children.filter(item => !item.hiddenNeed).map(item => {
                return {
                  title: item.dateMsg,
                  key: item.key,
                  minWidth: 100,
                  resizable: true,
                  ellipsis: {
                    tooltip: true
                  },
                  sorter: (row1, row2) => {
                    if (isNumeric(row1[item.key]) && isNumeric(row2[item.key])) {
                      return row1[item.key] - row2[item.key];
                    } else {
                      return 'default'
                    }
                  },
                }
              })
            }
          } else {
            return {
              title: item.title + (item.unit ? '[' + item.unit + ']' : ''),
              key: item.key,
              resizable: true,
              minWidth: 120,
              ellipsis: {
                tooltip: true
              },
              sorter: (row1, row2) => {
                if (isNumeric(row1[item.key]) && isNumeric(row2[item.key])) {
                  return row1[item.key] - row2[item.key];
                } else {
                  return 'default'
                }
              },
            }
          }
        })
      
      // 添加操作列
      indicatorStockSelectColumns.value.push({
        title: '操作',
        key: 'actions',
        width: 80,
        fixed: 'right',
        render: (row) => {
          return h(
            NButton,
            {
              strong: true,
              tertiary: true,
              size: 'small',
              type: 'warning',
              style: 'font-size: 14px; padding: 0 10px;',
              onClick: () => handleIndicatorStockFollow(row)
            },
            { default: () => '关注' }
          )
        }
      });
      
      indicatorStockSelectDataList.value = res.data.result.dataList
      
      // 计算并设置表格宽度
      indicatorStockSelectTableScrollX.value = calculateIndicatorTableWidth(indicatorStockSelectColumns.value);
      
      message.success(`找到 ${indicatorStockSelectDataList.value.length} 只股票`)
    } else {
      if (res.msg) {
        message.error(res.msg)
      }
      if (res.message) {
        message.error(res.message)
      }
    }
  }).catch(err => {
    loading.destroy()
    indicatorStockSelectLoading.value = false
    message.error(err)
  })
}

// 关注单个股票（指标选股）
function handleIndicatorStockFollow(row) {
  let code = row.MARKET_SHORT_NAME.toLowerCase() + row.SECURITY_CODE
  Follow(code).then(result => {
    if (result === "关注成功") {
      message.success(result)
    } else {
      message.error(result)
    }
  });
}

// 一键关注全部（指标选股）
async function followAllIndicatorStocks() {
  if (!indicatorStockSelectDataList.value || indicatorStockSelectDataList.value.length === 0) {
    message.warning('没有可关注的股票')
    return
  }
  
  const stocks = indicatorStockSelectDataList.value
  const stockCount = stocks.length
  
  // 询问是否创建新分组
  dialog.warning({
    title: '一键关注全部',
    content: `将关注 ${stockCount} 只股票。是否创建新分组来存放这些股票？`,
    positiveText: '创建新分组',
    negativeText: '添加到当前分组',
    onPositiveClick: async () => {
      await followAllStocksWithNewGroup(stocks)
    },
    onNegativeClick: async () => {
      await followAllStocksToCurrentGroup(stocks)
    }
  })
}

// 创建新分组并添加所有股票
async function followAllStocksWithNewGroup(stocks) {
  // 生成分组名称
  const conditionLower = indicatorStockSelectCondition.value.toLowerCase()
  let groupName = '指标选股'
  
  if (conditionLower.includes('换手率')) {
    groupName = '高换手率股票'
  } else if (conditionLower.includes('量比')) {
    groupName = '高量比股票'
  } else if (conditionLower.includes('涨幅')) {
    groupName = '强势股票'
  } else if (conditionLower.includes('市盈率')) {
    groupName = '低估值股票'
  } else if (conditionLower.includes('roe') || conditionLower.includes('净资产收益率')) {
    groupName = '高ROE股票'
  } else {
    groupName = `指标选股${stocks.length}只`
  }
  
  // 创建新分组
  const maxSort = groupList.value.length > 0 
    ? Math.max(...groupList.value.map(g => g.sort)) + 1 
    : 1
  
  message.loading('正在创建分组...')
  const createGroupResult = await AddGroup({
    name: groupName,
    sort: maxSort
  })
  message.destroyAll()
  
  if (!createGroupResult) {
    message.error('创建分组失败')
    return
  }
  
  // 刷新分组列表并获取新创建的分组ID
  const updatedGroups = await GetGroupList()
  groupList.value = updatedGroups
  const newGroup = updatedGroups.find(g => g.name === groupName && g.sort === maxSort)
  if (!newGroup) {
    message.error('无法找到新创建的分组')
    return
  }
  
  const newGroupId = newGroup.ID
  
  // 批量添加股票
  await batchAddStocksToGroup(stocks, newGroupId, groupName)
}

// 添加到当前分组
async function followAllStocksToCurrentGroup(stocks) {
  const groupId = currentGroupId.value
  const groupName = groupId === 0 ? '全部' : groupList.value.find(g => g.ID === groupId)?.name || '当前分组'
  
  await batchAddStocksToGroup(stocks, groupId, groupName)
}

// 批量添加股票到分组
async function batchAddStocksToGroup(stocks, groupId, groupName) {
  let successCount = 0
  let failCount = 0
  const failMessages = []
  
  message.loading(`正在添加股票到分组"${groupName}"...`, { duration: 0 })
  
  for (const stock of stocks) {
    try {
      let code = stock.MARKET_SHORT_NAME.toLowerCase() + stock.SECURITY_CODE
      
      // 先关注股票
      const followResult = await Follow(code)
      
      if (followResult === "关注成功" || followResult === "已经关注了") {
        // 如果groupId不为0，添加到分组
        if (groupId !== 0) {
          const groupResult = await AddStockGroup(groupId, code)
          if (groupResult === '添加成功' || groupResult === '已经关注了') {
            successCount++
          } else {
            failCount++
            failMessages.push(`${stock.SECURITY_SHORT_NAME || code}: ${groupResult}`)
          }
        } else {
          successCount++
        }
      } else {
        failCount++
        failMessages.push(`${stock.SECURITY_SHORT_NAME || code}: ${followResult}`)
      }
    } catch (error) {
      failCount++
      failMessages.push(`${stock.SECURITY_SHORT_NAME || stock.SECURITY_CODE}: ${error.message}`)
      console.error('添加股票异常:', stock, error)
    }
  }
  
  message.destroyAll()
  
  // 刷新分组列表和自选列表
  await GetGroupList().then(result => {
    groupList.value = result
  })
  
  if (groupId !== 0) {
    await GetFollowList(groupId).then(result => {
      followList.value = result
      stocks.value = []
      for (const followedStock of result) {
        let code = followedStock.StockCode
        if (code.startsWith("us")) {
          code = "gb_" + code.replace("us", "").toLowerCase()
        }
        stocks.value.push(code)
        Greet(code).then(result => {
          updateData(result)
        })
      }
    })
  } else {
    // 刷新全部列表
    GetFollowList(0).then(result => {
      followList.value = result
      stocks.value = []
      for (const followedStock of result) {
        let code = followedStock.StockCode
        if (code.startsWith("us")) {
          code = "gb_" + code.replace("us", "").toLowerCase()
        }
        stocks.value.push(code)
        Greet(code).then(result => {
          updateData(result)
        })
      }
    })
  }
  
  // 显示结果
  if (successCount > 0) {
    message.success(`成功添加 ${successCount} 只股票到分组"${groupName}"${failCount > 0 ? `，失败 ${failCount} 只` : ''}`)
    if (failCount > 0 && failMessages.length > 0) {
      console.warn('添加失败的股票:', failMessages)
    }
  } else {
    message.warning(`添加失败，共 ${failCount} 只股票`)
    if (failMessages.length > 0) {
      message.error(failMessages.slice(0, 5).join('; ') + (failMessages.length > 5 ? '...' : ''))
    }
  }
}

// 保存prompt到历史记录
function savePromptToHistory(promptContent, promptType) {
  if (!promptContent || !promptContent.trim()) {
    return
  }
  
  const trimmedContent = promptContent.trim()
  
  // 生成prompt名称（截取前30个字符）
  let promptName = trimmedContent
  if (promptName.length > 30) {
    promptName = promptName.substring(0, 30) + '...'
  }
  promptName = `${promptType}-${promptName}`
  
  // 检查是否已存在相同的prompt（避免重复保存）
  const exists = promptTemplates.value && promptTemplates.value.some(t => 
    t.type === '模型用户Prompt' && t.content === trimmedContent
  )
  
  if (!exists) {
    // 保存为"模型用户Prompt"类型
    AddPrompt({
      ID: 0,
      Name: promptName,
      Content: trimmedContent,
      Type: '模型用户Prompt'
    }).then(result => {
      // 静默保存，不显示提示，避免打扰用户
      console.log('Prompt已保存到历史记录:', promptName)
      // 刷新prompt模板列表
      GetPromptTemplates("", "").then(res => {
        if (res) {
          promptTemplates.value = res
          sysPromptOptions.value = promptTemplates.value.filter(item => item.type === '模型系统Prompt')
          userPromptOptions.value = promptTemplates.value.filter(item => item.type === '模型用户Prompt')
        }
      })
    }).catch(error => {
      console.error('保存prompt失败:', error)
    })
  }
}

function updateTab(name) {
  stocks.value = []
  const tabId= Number(name)
  currentGroupId.value = tabId;
  GetFollowList(tabId).then(result => {
    followList.value = result

    for (const followedStock of result) {
      if (followedStock.StockCode.startsWith("us")) {
        followedStock.StockCode = "gb_" + followedStock.StockCode.replace("us", "").toLowerCase()
      }
      stocks.value.push(followedStock.StockCode)
      Greet(followedStock.StockCode).then(result => {
        updateData(result)
      })
    }
    //monitor()
    message.destroyAll()
  })
}

function delTab(groupId) {
  let infos = groupList.value = groupList.value.filter(item => item.ID === Number(groupId))
  dialog.create({
    title: '删除分组',
    type: 'warning',
    content: '确定要删除[' + infos[0].name + ']分组吗？分组数据将不能恢复哟！',
    positiveText: '确定',
    negativeText: '取消',
    onPositiveClick: () => {
      RemoveGroup(Number(groupId)).then(result => {
        message.info(result)
        GetGroupList().then(result => {
          groupList.value = result
        })
      })
    }
  })
}

function delStockGroup(code, name, groupId) {
  RemoveStockGroup(code, name, groupId).then(result => {
    updateTab(groupId)
    message.info(result)
  })
}

function searchNotice(stockCode) {
  router.push({
    name: 'market',
    query: {
      name: '公司公告',
      stockCode: stockCode,
    },
  })
}

function searchStockReport(stockCode) {
  router.push({
    name: 'market',
    query: {
      name: '个股研报',
      stockCode: stockCode,
    },
  })
}

// 获取更多操作的下拉菜单选项
function getMoreOptions(result) {
  const options = []

  // 成本
  options.push({
    label: '成本',
    key: 'cost'
  })

  // 资金（如果有买一报价）
  if (result['买一报价'] > 0) {
    options.push({
      label: '资金',
      key: 'money'
    })
  }

  // 详情
  options.push({
    label: '详情',
    key: 'detail'
  })

  // 公告（如果有买一报价）
  if (result['买一报价'] > 0) {
    options.push({
      label: '公告',
      key: 'notice'
    })
  }

  // 研报（如果有买一报价）
  if (result['买一报价'] > 0) {
    options.push({
      label: '研报',
      key: 'report'
    })
  }

  // 设置分组 - 使用子菜单
  const groupOptions = groupList.value.map(group => ({
    label: group.name,
    key: `group_${group.ID}`
  }))

  if (groupOptions.length > 0) {
    options.push({
      label: '设置分组',
      key: 'group',
      children: groupOptions
    })
  }

  return options
}

// 处理更多操作的下拉菜单选择
function handleMoreAction(key, result) {
  if (key === 'cost') {
    setStock(result['股票代码'], result['股票名称'])
  } else if (key === 'money') {
    showMoney(result['股票代码'], result['股票名称'])
  } else if (key === 'detail') {
    search(result['股票代码'], result['股票名称'])
  } else if (key === 'notice') {
    searchNotice(result['股票代码'])
  } else if (key === 'report') {
    searchStockReport(result['股票代码'])
  } else if (key.startsWith('group_')) {
    const groupId = parseInt(key.replace('group_', ''))
    AddStockGroupInfo(groupId, result['股票代码'], result['股票名称'])
  }
}

// 补仓计算器相关函数
// 计算平均成本
const averageCost = computed(() => {
  let totalCost = costCalculator.initialPrice * costCalculator.initialVolume
  let totalVolume = costCalculator.initialVolume
  
  costCalculator.additions.forEach(add => {
    totalCost += add.price * add.volume
    totalVolume += add.volume
  })
  
  return totalVolume > 0 ? (totalCost / totalVolume).toFixed(2) : costCalculator.initialPrice.toFixed(2)
})

// 计算总持仓数量
const totalVolume = computed(() => {
  return costCalculator.initialVolume + costCalculator.additions.reduce((sum, add) => sum + add.volume, 0)
})

// 计算总投入成本
const totalCost = computed(() => {
  const initialCost = costCalculator.initialPrice * costCalculator.initialVolume
  const additionsCost = costCalculator.additions.reduce((sum, add) => sum + add.price * add.volume, 0)
  return (initialCost + additionsCost).toFixed(2)
})

// 计算当前市值
const currentMarketValue = computed(() => {
  return (totalVolume.value * costCalculator.currentPrice).toFixed(2)
})

// 计算盈亏
const profitLoss = computed(() => {
  return (parseFloat(currentMarketValue.value) - parseFloat(totalCost.value)).toFixed(2)
})

// 计算盈亏比例
const profitLossPercent = computed(() => {
  if (parseFloat(totalCost.value) === 0) return '0.00'
  return ((parseFloat(profitLoss.value) / parseFloat(totalCost.value)) * 100).toFixed(2)
})

// 添加补仓记录
function addCostAddition() {
  if (costCalculator.additions.length === 0 || 
      (costCalculator.additions[costCalculator.additions.length - 1].price > 0 && 
       costCalculator.additions[costCalculator.additions.length - 1].volume > 0)) {
    costCalculator.additions.push({
      price: 0,
      volume: 0,
      id: costCalculator.nextId++
    })
  }
  // 自动聚焦到最后一个输入框
  nextTick(() => {
    const inputs = document.querySelectorAll('.cost-addition-input')
    if (inputs.length > 0) {
      const lastInput = inputs[inputs.length - 1]
      if (lastInput && lastInput.focus) {
        lastInput.focus()
      }
    }
  })
}

// 删除补仓记录
function removeCostAddition(id) {
  const index = costCalculator.additions.findIndex(add => add.id === id)
  if (index > -1) {
    costCalculator.additions.splice(index, 1)
  }
}

// 清空所有补仓记录
function clearAllAdditions() {
  costCalculator.additions = []
}

// 计算需要补仓多少才能达到目标成本
function calculateTargetCost(targetCost) {
  const currentTotalCost = parseFloat(totalCost.value)
  const currentTotalVolume = totalVolume.value
  
  if (targetCost >= costCalculator.currentPrice) {
    return {
      volume: 0,
      message: '目标成本不能高于当前价格'
    }
  }
  
  // 公式：目标成本 = (当前总成本 + 补仓价格 * 补仓数量) / (当前总数量 + 补仓数量)
  // 解方程得到：补仓数量 = (当前总成本 - 目标成本 * 当前总数量) / (目标成本 - 补仓价格)
  const volume = (currentTotalCost - targetCost * currentTotalVolume) / (targetCost - costCalculator.currentPrice)
  
  if (volume <= 0) {
    return {
      volume: 0,
      message: '当前成本已低于目标成本，无需补仓'
    }
  }
  
  return {
    volume: Math.ceil(volume),
    cost: (volume * costCalculator.currentPrice).toFixed(2),
    message: `需要在当前价格 ${costCalculator.currentPrice} 元补仓 ${Math.ceil(volume)} 股，投入 ${(volume * costCalculator.currentPrice).toFixed(2)} 元`
  }
}

// 计算补仓后成本
const targetCostResult = ref(null)
const targetCostInput = ref(0)

function calculateTarget() {
  if (targetCostInput.value <= 0) {
    message.warning('请输入有效的目标成本')
    return
  }
  targetCostResult.value = calculateTargetCost(targetCostInput.value)
}
</script>

<template>
  <vue-danmaku v-model:danmus="danmus" useSlot
               style="height:100px; width:100%;z-index: 9;position:absolute; top: 400px; pointer-events: none;">
    <template v-slot:dm="{ index, danmu }">
      <n-gradient-text type="info">
        <n-icon :component="ChatboxOutline"/>
        {{ danmu }}
      </n-gradient-text>
    </template>
  </vue-danmaku>
  <!-- 显示模式切换按钮和清理按钮 -->
  <div style="position: fixed; top: 10px; right: 10px; z-index: 10; --wails-draggable:no-drag">
    <n-space>
      <n-button-group>
        <n-button 
          :type="displayMode === 'card' ? 'primary' : 'default'" 
          @click="displayMode = 'card'"
          size="small"
        >
          <template #icon>
            <n-icon :component="GridOutline"/>
          </template>
          卡片
        </n-button>
        <n-button 
          :type="displayMode === 'list' ? 'primary' : 'default'" 
          @click="displayMode = 'list'"
          size="small"
        >
          <template #icon>
            <n-icon :component="ListOutline"/>
          </template>
          列表
        </n-button>
      </n-button-group>
      <n-button 
        type="error" 
        size="small"
        @click="clearAllStocks"
        :disabled="followList.length === 0"
      >
        <template #icon>
          <n-icon :component="TrashOutline"/>
        </template>
        一键清理
      </n-button>
    </n-space>
  </div>

  <n-tabs type="card" style="--wails-draggable:no-drag" animated addable :data-currentGroupId="currentGroupId"
          :value="String(currentGroupId)" @add="addTab" @update:value="updateTab" placement="top" @close="(key)=>{delTab(key)}">

    <n-tab-pane v-for="group in sortedGroupList" :key="group.ID" :closable="group.ID !== 0" :name="String(group.ID)" :data-name="String(group.ID)">
      <template #tab v-if="group.ID === 0">
        <span :data-name="'0'">全部</span>
      </template>
      <template #tab v-else>
        <span v-if="editingGroupId !== group.ID" @dblclick.stop="startEditGroup(group.ID, group.name)" style="cursor: pointer; user-select: none;" :data-name="String(group.ID)">
          {{ group.name }}
        </span>
        <n-input
          v-else
          v-model:value="editingGroupName"
          size="small"
          style="width: 120px;"
          @blur="saveGroupName(group.ID)"
          @keyup.enter="saveGroupName(group.ID)"
          @keyup.esc="cancelEditGroup"
          autofocus
        />
      </template>
      <!-- "全部"分组的内容 -->
      <template v-if="group.ID === 0">
        <!-- 卡片式显示 -->
        <n-grid v-if="displayMode === 'card'" :x-gap="8" :cols="3" :y-gap="8">
          <n-gi :id="result['股票代码']+'_gi'" v-for="result in sortedResults" style="margin-left: 2px;">
          <n-card :data-sort="result.sort" :id="result['股票代码']" :data-code="result['股票代码']" :bordered="true"
                  :title="result['股票名称']" :closable="false"
                  @close="removeMonitor(result['股票代码'],result['股票名称'],result.key)">
            <n-grid :cols="1" :y-gap="6">
              <n-gi>
                <n-text :type="result.type">
                  <n-number-animation :duration="1000" :precision="2" :from="result['上次当前价格']"
                                      :to="Number(result['当前价格'])"/>
                  <n-tag size="small" :type="result.type" :bordered="false" v-if="result['盘前盘后']>0">
                    ({{ result['盘前盘后'] }} {{ result['盘前盘后涨跌幅'] }}%)
                  </n-tag>
                </n-text>
                <n-text style="padding-left: 10px;" :type="result.type">
                  <n-number-animation :duration="1000" :precision="3" :from="0" :to="result.changePercent"/>
                  %
                </n-text>&nbsp;
                <n-text size="small" v-if="result.costVolume>0" :type="result.type">
                  <n-number-animation :duration="1000" :precision="2" :from="0" :to="result.profitAmountToday"/>
                </n-text>
              </n-gi>
            </n-grid>
            <n-grid :cols="2" :y-gap="4" :x-gap="4">
              <n-gi>
                <n-text :type="'info'">{{ "最高 " + result["今日最高价"] + " " + result.highRate }}%</n-text>
              </n-gi>
              <n-gi>
                <n-text :type="'info'">{{ "最低 " + result["今日最低价"] + " " + result.lowRate }}%</n-text>
              </n-gi>
              <n-gi>
                <n-text :type="'info'">{{ "昨收 " + result["昨日收盘价"] }}</n-text>
              </n-gi>
              <n-gi>
                <n-text :type="'info'">{{ "今开 " + result["今日开盘价"] }}</n-text>
              </n-gi>
            </n-grid>
            <n-collapse accordion v-if="result['买一报价']>0">
              <n-collapse-item title="盘口" name="1" v-if="result['买一报价']>0">
                <template #header-extra>
                  <n-flex justify="space-between">
                    <n-text :type="'info'">{{ "买一 " + result["买一报价"] + '(' + result["买一申报"] + ")" }}</n-text>
                    <n-text :type="'info'">{{ "卖一 " + result["卖一报价"] + '(' + result["卖一申报"] + ")" }}</n-text>
                  </n-flex>
                </template>
                <n-grid :cols="2" :y-gap="4" :x-gap="4">
                  <n-gi v-if="result['买一报价']>0">
                    <n-text :type="'info'">{{ "买一 " + result["买一报价"] + '(' + result["买一申报"] + ")" }}</n-text>
                  </n-gi>
                  <n-gi v-if="result['卖一报价']>0">
                    <n-text :type="'info'">{{ "卖一 " + result["卖一报价"] + '(' + result["卖一申报"] + ")" }}</n-text>
                  </n-gi>

                  <n-gi v-if="result['买二报价']>0">
                    <n-text :type="'info'">{{ "买二 " + result["买二报价"] + '(' + result["买二申报"] + ")" }}</n-text>
                  </n-gi>
                  <n-gi v-if="result['卖二报价']>0">
                    <n-text :type="'info'">{{ "卖二 " + result["卖二报价"] + '(' + result["卖二申报"] + ")" }}</n-text>
                  </n-gi>

                  <n-gi v-if="result['买三报价']>0">
                    <n-text :type="'info'">{{ "买三 " + result["买三报价"] + '(' + result["买三申报"] + ")" }}</n-text>
                  </n-gi>
                  <n-gi v-if="result['卖三报价']>0">
                    <n-text :type="'info'">{{ "买三 " + result["卖三报价"] + '(' + result["卖三申报"] + ")" }}</n-text>
                  </n-gi>

                  <n-gi v-if="result['买四报价']>0">
                    <n-text :type="'info'">{{ "买四 " + result["买四报价"] + '(' + result["买四申报"] + ")" }}</n-text>
                  </n-gi>
                  <n-gi v-if="result['卖四报价']>0">
                    <n-text :type="'info'">{{ "卖四 " + result["卖四报价"] + '(' + result["卖四申报"] + ")" }}</n-text>
                  </n-gi>

                  <n-gi v-if="result['买五报价']>0">
                    <n-text :type="'info'">{{ "买五 " + result["买五报价"] + '(' + result["买五申报"] + ")" }}</n-text>
                  </n-gi>
                  <n-gi v-if="result['卖五报价']>0">
                    <n-text :type="'info'">{{ "卖五 " + result["卖五报价"] + '(' + result["卖五申报"] + ")" }}</n-text>
                  </n-gi>
                </n-grid>
              </n-collapse-item>
            </n-collapse>
            <template #header-extra>

              <n-tag size="small" :bordered="false">{{ result['股票代码'] }}</n-tag>&nbsp;
              <n-button size="tiny" secondary type="primary"
                        @click="removeMonitor(result['股票代码'],result['股票名称'],result.key)">
                取消关注
              </n-button>&nbsp;

              <n-button size="tiny" v-if="data.openAiEnable" secondary type="warning"
                        @click="aiCheckStock(result['股票名称'],result['股票代码'])">
                AI分析
              </n-button>
            </template>
            <template #footer>
              <n-flex justify="center" vertical>
                <n-flex justify="center">
                  <n-text :type="'info'">{{ result["日期"] + " " + result["时间"] }}</n-text>
                  <n-tag size="small" v-if="result.volume>0" :type="result.profitType">{{ result.volume + "股" }}</n-tag>
                  <n-tag size="small" v-if="result.costPrice>0" :type="result.profitType">
                    {{
                      "成本:" + result.costPrice + "*" + result.costVolume + " " + result.profit + "%" + " ( " + result.profitAmount + " ¥ )"
                    }}
                  </n-tag>
                </n-flex>
                <n-flex justify="center" v-if="result.buyDate">
                  <n-text size="small" type="info">买入日期: {{ new Date(result.buyDate).toLocaleDateString('zh-CN') }}</n-text>
                  <n-text size="small" type="info" v-if="result.holdingDays >= 0">持有天数: {{ result.holdingDays }}天</n-text>
                </n-flex>
              </n-flex>
            </template>
            <template #action>
              <n-flex justify="left">
                <!-- 常用功能：直接显示 -->
                <n-button size="tiny" type="error"
                          @click="showFenshi(result['股票代码'],result['股票名称'],result.changePercent)"> 分时
                </n-button>
                <n-button size="tiny" type="error" @click="showK(result['股票代码'],result['股票名称'])"> 日K</n-button>
                <n-button size="tiny" type="error" @click="showWeekK(result['股票代码'],result['股票名称'])"> 周K</n-button>
                <n-button size="tiny" type="error" @click="showMonthK(result['股票代码'],result['股票名称'])"> 月K</n-button>
                <n-button size="tiny" v-if="data.openAiEnable" type="warning" secondary
                          @click="aiCheckStock(result['股票名称'],result['股票代码'])"> AI分析
                </n-button>
                <!-- 其他功能：下拉菜单 -->
                <n-dropdown trigger="click" :options="getMoreOptions(result)" @select="(key) => handleMoreAction(key, result)">
                  <n-button size="tiny" type="info" secondary>更多</n-button>
                </n-dropdown>
              </n-flex>
            </template>
          </n-card>
        </n-gi>
      </n-grid>
      
      <!-- 列表式显示 -->
      <n-list v-else bordered hoverable>
        <n-list-item v-for="(result, index) in Object.values(sortedResults)" :key="result['股票代码']" 
                     :id="result['股票代码']+'_gi'" 
                     :data-sort="result.sort" 
                     :data-code="result['股票代码']">
          <n-card :bordered="true" style="width: 100%; margin-bottom: 8px;">
            <n-grid :cols="24" :x-gap="12">
              <!-- 序号 -->
              <n-gi :span="1">
                <n-text type="info" strong style="font-size: 14px;">{{ index + 1 }}</n-text>
              </n-gi>
              
              <!-- 股票名称和代码 -->
              <n-gi :span="3">
                <n-flex vertical>
                  <n-text strong>{{ result['股票名称'] }}</n-text>
                  <n-tag size="small" :bordered="false">{{ result['股票代码'] }}</n-tag>
                </n-flex>
              </n-gi>
              
              <!-- 价格和涨跌幅 -->
              <n-gi :span="4">
                <n-flex vertical>
                  <n-text :type="result.type" strong style="font-size: 18px;">
                    <n-number-animation :duration="1000" :precision="2" :from="result['上次当前价格']"
                                        :to="Number(result['当前价格'])"/>
                    <n-tag size="small" :type="result.type" :bordered="false" v-if="result['盘前盘后']>0" style="margin-left: 5px;">
                      ({{ result['盘前盘后'] }} {{ result['盘前盘后涨跌幅'] }}%)
                    </n-tag>
                  </n-text>
                  <n-text :type="result.type" style="font-size: 14px;">
                    <n-number-animation :duration="1000" :precision="3" :from="0" :to="result.changePercent"/>
                    %
                  </n-text>
                </n-flex>
              </n-gi>
              
              <!-- 最高最低 -->
              <n-gi :span="3">
                <n-flex vertical>
                  <n-text :type="'info'" size="small">最高: {{ result["今日最高价"] }} ({{ result.highRate }}%)</n-text>
                  <n-text :type="'info'" size="small">最低: {{ result["今日最低价"] }} ({{ result.lowRate }}%)</n-text>
                </n-flex>
              </n-gi>
              
              <!-- 昨收今开 -->
              <n-gi :span="3">
                <n-flex vertical>
                  <n-text :type="'info'" size="small">昨收: {{ result["昨日收盘价"] }}</n-text>
                  <n-text :type="'info'" size="small">今开: {{ result["今日开盘价"] }}</n-text>
                </n-flex>
              </n-gi>
              
              <!-- 成本信息 -->
              <n-gi :span="4" v-if="result.costPrice>0 || result.volume>0">
                <n-flex vertical>
                  <n-tag v-if="result.volume>0" size="small" :type="result.profitType">{{ result.volume + "股" }}</n-tag>
                  <n-tag v-if="result.costPrice>0" size="small" :type="result.profitType">
                    {{ "成本:" + result.costPrice + "*" + result.costVolume + " " + result.profit + "%" }}
                  </n-tag>
                  <n-text v-if="result.costVolume>0" size="small" :type="result.type">
                    盈亏: <n-number-animation :duration="1000" :precision="2" :from="0" :to="result.profitAmountToday"/>
                  </n-text>
                  <n-text v-if="result.buyDate" size="small" type="info">
                    买入: {{ new Date(result.buyDate).toLocaleDateString('zh-CN') }}
                    <span v-if="result.holdingDays >= 0"> ({{ result.holdingDays }}天)</span>
                  </n-text>
                </n-flex>
              </n-gi>
              
              <!-- 时间 -->
              <n-gi :span="3">
                <n-text :type="'info'" size="small">{{ result["日期"] + " " + result["时间"] }}</n-text>
              </n-gi>
              
              <!-- 操作按钮 -->
              <n-gi :span="4">
                <n-flex justify="end" wrap>
                  <n-button size="tiny" type="error" @click="showFenshi(result['股票代码'],result['股票名称'],result.changePercent)">分时</n-button>
                  <n-button size="tiny" type="error" @click="showK(result['股票代码'],result['股票名称'])">日K</n-button>
                  <n-button size="tiny" secondary type="primary" @click="removeMonitor(result['股票代码'],result['股票名称'],result.key)">取消关注</n-button>
                  <n-button size="tiny" v-if="data.openAiEnable" secondary type="warning" @click="aiCheckStock(result['股票名称'],result['股票代码'])">AI分析</n-button>
                  <n-dropdown trigger="click" :options="getMoreOptions(result)" @select="(key) => handleMoreAction(key, result)">
                    <n-button size="tiny" type="info" secondary>更多</n-button>
                  </n-dropdown>
                </n-flex>
              </n-gi>
            </n-grid>
          </n-card>
        </n-list-item>
      </n-list>
      </template>
      <!-- 其他分组的内容 -->
      <template v-else>
        <!-- 卡片式显示 -->
        <n-grid v-if="displayMode === 'card'" :x-gap="8" :cols="3" :y-gap="8">
          <n-gi :id="result['股票代码']+'_gi'" v-for="result in groupResults" style="margin-left: 2px;">
          <n-card :data-sort="result.sort" :id="result['股票代码']" :data-code="result['股票代码']" :bordered="true"
                  :title="result['股票名称']" :closable="false"
                  @close="removeMonitor(result['股票代码'],result['股票名称'],result.key)">
            <n-grid :cols="12" :y-gap="6">
              <n-gi :span="12">
                <n-text :type="result.type">
                  <n-number-animation :duration="1000" :precision="2" :from="result['上次当前价格']"
                                      :to="Number(result['当前价格'])"/>
                  <n-tag size="small" :type="result.type" :bordered="false" v-if="result['盘前盘后']>0">
                    ({{ result['盘前盘后'] }} {{ result['盘前盘后涨跌幅'] }}%)
                  </n-tag>
                </n-text>
                <n-text style="padding-left: 10px;" :type="result.type">
                  <n-number-animation :duration="1000" :precision="3" :from="0" :to="result.changePercent"/>
                  %
                </n-text>&nbsp;
                <n-text size="small" v-if="result.costVolume>0" :type="result.type">
                  <n-number-animation :duration="1000" :precision="2" :from="0" :to="result.profitAmountToday"/>
                </n-text>
              </n-gi>
            </n-grid>
            <n-grid :cols="2" :y-gap="4" :x-gap="4">
              <n-gi>
                <n-text :type="'info'">{{ "最高 " + result["今日最高价"] + " " + result.highRate }}%</n-text>
              </n-gi>
              <n-gi>
                <n-text :type="'info'">{{ "最低 " + result["今日最低价"] + " " + result.lowRate }}%</n-text>
              </n-gi>
              <n-gi>
                <n-text :type="'info'">{{ "昨收 " + result["昨日收盘价"] }}</n-text>
              </n-gi>
              <n-gi>
                <n-text :type="'info'">{{ "今开 " + result["今日开盘价"] }}</n-text>
              </n-gi>
            </n-grid>
            <n-collapse accordion v-if="result['买一报价']>0">
              <n-collapse-item title="盘口" name="1" v-if="result['买一报价']>0">
                <template #header-extra>
                  <n-flex justify="space-between">
                    <n-text :type="'info'">{{ "买一 " + result["买一报价"] + '(' + result["买一申报"] + ")" }}</n-text>
                    <n-text :type="'info'">{{ "卖一 " + result["卖一报价"] + '(' + result["卖一申报"] + ")" }}</n-text>
                  </n-flex>
                </template>
                <n-grid :cols="2" :y-gap="4" :x-gap="4">
                  <n-gi v-if="result['买一报价']>0">
                    <n-text :type="'info'">{{ "买一 " + result["买一报价"] + '(' + result["买一申报"] + ")" }}</n-text>
                  </n-gi>
                  <n-gi v-if="result['卖一报价']>0">
                    <n-text :type="'info'">{{ "卖一 " + result["卖一报价"] + '(' + result["卖一申报"] + ")" }}</n-text>
                  </n-gi>

                  <n-gi v-if="result['买二报价']>0">
                    <n-text :type="'info'">{{ "买二 " + result["买二报价"] + '(' + result["买二申报"] + ")" }}</n-text>
                  </n-gi>
                  <n-gi v-if="result['卖二报价']>0">
                    <n-text :type="'info'">{{ "卖二 " + result["卖二报价"] + '(' + result["卖二申报"] + ")" }}</n-text>
                  </n-gi>

                  <n-gi v-if="result['买三报价']>0">
                    <n-text :type="'info'">{{ "买三 " + result["买三报价"] + '(' + result["买三申报"] + ")" }}</n-text>
                  </n-gi>
                  <n-gi v-if="result['卖三报价']>0">
                    <n-text :type="'info'">{{ "买三 " + result["卖三报价"] + '(' + result["卖三申报"] + ")" }}</n-text>
                  </n-gi>

                  <n-gi v-if="result['买四报价']>0">
                    <n-text :type="'info'">{{ "买四 " + result["买四报价"] + '(' + result["买四申报"] + ")" }}</n-text>
                  </n-gi>
                  <n-gi v-if="result['卖四报价']>0">
                    <n-text :type="'info'">{{ "卖四 " + result["卖四报价"] + '(' + result["卖四申报"] + ")" }}</n-text>
                  </n-gi>

                  <n-gi v-if="result['买五报价']>0">
                    <n-text :type="'info'">{{ "买五 " + result["买五报价"] + '(' + result["买五申报"] + ")" }}</n-text>
                  </n-gi>
                  <n-gi v-if="result['卖五报价']>0">
                    <n-text :type="'info'">{{ "卖五 " + result["卖五报价"] + '(' + result["卖五申报"] + ")" }}</n-text>
                  </n-gi>
                </n-grid>
              </n-collapse-item>
            </n-collapse>
            <template #header-extra>

              <n-tag size="small" :bordered="false">{{ result['股票代码'] }}</n-tag>&nbsp;
              <n-button size="tiny" secondary type="primary"
                        @click="removeMonitor(result['股票代码'],result['股票名称'],result.key)">
                取消关注
              </n-button>&nbsp;

              <n-button size="tiny" v-if="data.openAiEnable" secondary type="warning"
                        @click="aiCheckStock(result['股票名称'],result['股票代码'])">
                AI分析
              </n-button>
              <n-button secondary type="error" size="tiny"
                        @click="delStockGroup(result['股票代码'],result['股票名称'],group.ID)">移出分组
              </n-button>
            </template>
            <template #footer>
              <n-flex justify="center" vertical>
                <n-flex justify="center">
                  <n-text :type="'info'">{{ result["日期"] + " " + result["时间"] }}</n-text>
                  <n-tag size="small" v-if="result.volume>0" :type="result.profitType">{{ result.volume + "股" }}</n-tag>
                  <n-tag size="small" v-if="result.costPrice>0" :type="result.profitType">
                    {{
                      "成本:" + result.costPrice + "*" + result.costVolume + " " + result.profit + "%" + " ( " + result.profitAmount + " ¥ )"
                    }}
                  </n-tag>
                </n-flex>
                <n-flex justify="center" v-if="result.buyDate">
                  <n-text size="small" type="info">买入日期: {{ new Date(result.buyDate).toLocaleDateString('zh-CN') }}</n-text>
                  <n-text size="small" type="info" v-if="result.holdingDays >= 0">持有天数: {{ result.holdingDays }}天</n-text>
                </n-flex>
              </n-flex>
            </template>
            <template #action>
              <n-flex justify="left">
                <!-- 常用功能：直接显示 -->
                <n-button size="tiny" type="error"
                          @click="showFenshi(result['股票代码'],result['股票名称'],result.changePercent)"> 分时
                </n-button>
                <n-button size="tiny" type="error" @click="showK(result['股票代码'],result['股票名称'])"> 日K</n-button>
                <n-button size="tiny" type="error" @click="showWeekK(result['股票代码'],result['股票名称'])"> 周K</n-button>
                <n-button size="tiny" type="error" @click="showMonthK(result['股票代码'],result['股票名称'])"> 月K</n-button>
                <n-button size="tiny" v-if="data.openAiEnable" type="warning" secondary
                          @click="aiCheckStock(result['股票名称'],result['股票代码'])"> AI分析
                </n-button>
                <!-- 其他功能：下拉菜单 -->
                <n-dropdown trigger="click" :options="getMoreOptions(result)" @select="(key) => handleMoreAction(key, result)">
                  <n-button size="tiny" type="info" secondary>更多</n-button>
                </n-dropdown>
              </n-flex>
            </template>
          </n-card>
        </n-gi>
      </n-grid>
      
      <!-- 列表式显示 -->
      <n-list v-else bordered hoverable>
        <n-list-item v-for="(result, index) in Object.values(groupResults)" :key="result['股票代码']" 
                     :id="result['股票代码']+'_gi'" 
                     :data-sort="result.sort" 
                     :data-code="result['股票代码']">
          <n-card :bordered="true" style="width: 100%; margin-bottom: 8px;">
            <n-grid :cols="24" :x-gap="12">
              <!-- 序号 -->
              <n-gi :span="1">
                <n-text type="info" strong style="font-size: 14px;">{{ index + 1 }}</n-text>
              </n-gi>
              
              <!-- 股票名称和代码 -->
              <n-gi :span="3">
                <n-flex vertical>
                  <n-text strong>{{ result['股票名称'] }}</n-text>
                  <n-tag size="small" :bordered="false">{{ result['股票代码'] }}</n-tag>
                </n-flex>
              </n-gi>
              
              <!-- 价格和涨跌幅 -->
              <n-gi :span="4">
                <n-flex vertical>
                  <n-text :type="result.type" strong style="font-size: 18px;">
                    <n-number-animation :duration="1000" :precision="2" :from="result['上次当前价格']"
                                        :to="Number(result['当前价格'])"/>
                    <n-tag size="small" :type="result.type" :bordered="false" v-if="result['盘前盘后']>0" style="margin-left: 5px;">
                      ({{ result['盘前盘后'] }} {{ result['盘前盘后涨跌幅'] }}%)
                    </n-tag>
                  </n-text>
                  <n-text :type="result.type" style="font-size: 14px;">
                    <n-number-animation :duration="1000" :precision="3" :from="0" :to="result.changePercent"/>
                    %
                  </n-text>
                </n-flex>
              </n-gi>
              
              <!-- 最高最低 -->
              <n-gi :span="3">
                <n-flex vertical>
                  <n-text :type="'info'" size="small">最高: {{ result["今日最高价"] }} ({{ result.highRate }}%)</n-text>
                  <n-text :type="'info'" size="small">最低: {{ result["今日最低价"] }} ({{ result.lowRate }}%)</n-text>
                </n-flex>
              </n-gi>
              
              <!-- 昨收今开 -->
              <n-gi :span="3">
                <n-flex vertical>
                  <n-text :type="'info'" size="small">昨收: {{ result["昨日收盘价"] }}</n-text>
                  <n-text :type="'info'" size="small">今开: {{ result["今日开盘价"] }}</n-text>
                </n-flex>
              </n-gi>
              
              <!-- 成本信息 -->
              <n-gi :span="4" v-if="result.costPrice>0 || result.volume>0">
                <n-flex vertical>
                  <n-tag v-if="result.volume>0" size="small" :type="result.profitType">{{ result.volume + "股" }}</n-tag>
                  <n-tag v-if="result.costPrice>0" size="small" :type="result.profitType">
                    {{ "成本:" + result.costPrice + "*" + result.costVolume + " " + result.profit + "%" }}
                  </n-tag>
                  <n-text v-if="result.costVolume>0" size="small" :type="result.type">
                    盈亏: <n-number-animation :duration="1000" :precision="2" :from="0" :to="result.profitAmountToday"/>
                  </n-text>
                </n-flex>
              </n-gi>
              
              <!-- 时间 -->
              <n-gi :span="3">
                <n-text :type="'info'" size="small">{{ result["日期"] + " " + result["时间"] }}</n-text>
              </n-gi>
              
              <!-- 操作按钮 -->
              <n-gi :span="4">
                <n-flex justify="end" wrap>
                  <n-button size="tiny" type="error" @click="showFenshi(result['股票代码'],result['股票名称'],result.changePercent)">分时</n-button>
                  <n-button size="tiny" type="error" @click="showK(result['股票代码'],result['股票名称'])">日K</n-button>
                  <n-button size="tiny" secondary type="primary" @click="removeMonitor(result['股票代码'],result['股票名称'],result.key)">取消关注</n-button>
                  <n-button size="tiny" v-if="data.openAiEnable" secondary type="warning" @click="aiCheckStock(result['股票名称'],result['股票代码'])">AI分析</n-button>
                  <n-button secondary type="error" size="tiny" @click="delStockGroup(result['股票代码'],result['股票名称'],group.ID)">移出分组</n-button>
                  <n-dropdown trigger="click" :options="getMoreOptions(result)" @select="(key) => handleMoreAction(key, result)">
                    <n-button size="tiny" type="info" secondary>更多</n-button>
                  </n-dropdown>
                </n-flex>
              </n-gi>
            </n-grid>
          </n-card>
        </n-list-item>
      </n-list>
      </template>
    </n-tab-pane>
    
  </n-tabs>

  <div style="position: fixed;bottom: 18px;right:5px;z-index: 10;width: 400px">
    <!--    <n-card :bordered="false">-->
    <n-input-group>
      <!--        <n-button  type="error" @click="addBTN=!addBTN" > <n-icon :component="Search"/>&nbsp;<n-text  v-if="addBTN">隐藏</n-text></n-button>-->

      <n-auto-complete v-model:value="data.name" v-if="addBTN"
                       :input-props="{
                                autocomplete: 'disabled',
                              }"
                       :options="options"
                       placeholder="股票指数名称/代码/弹幕"
                       clearable @update-value="getStockList" :on-select="onSelect"/>

      <n-popover trigger="manual" :show="showPopover">
        <template #trigger>
          <n-button type="primary" @click="AddStock" v-if="addBTN">
            <n-icon :component="Add"/> &nbsp;关注
          </n-button>
        </template>
        <span>输入股票名称/代码关键词开始吧~~~</span>
      </n-popover>

      <n-button type="info" @click="SendDanmu" v-if="data.enableDanmu">
        <n-icon :component="ChatboxOutline"/> &nbsp;发送弹幕
      </n-button>
    </n-input-group>
    <!--    </n-card>-->
  </div>
  <n-modal transform-origin="center" size="small" v-model:show="modalShow" :title="formModel.name" style="width: 400px"
           :preset="'card'">
    <n-form :model="formModel" :rules="{
              costPrice: { required: true, message: '请输入成本'},
              volume: { required: true, message: '请输入数量'},
              alarm:{required: true, message: '涨跌报警值'} ,
              alarmPrice: { required: true, message: '请输入报警价格'},
              sort: { required: true, message: '请输入排序值'},
            }" label-placement="left" label-width="80px">
      <n-form-item label="股票成本" path="costPrice">
        <n-input-number v-model:value="formModel.costPrice" min="0" placeholder="请输入股票成本">
          <template #suffix>
            {{ formModel.code.indexOf("hk") >= 0 ? "HK$" : "¥" }}
          </template>
        </n-input-number>
      </n-form-item>
      <n-form-item label="股票数量" path="volume">
        <n-input-number v-model:value="formModel.volume" min="0" step="100" placeholder="请输入股票数量">
          <template #suffix>
            股
          </template>
        </n-input-number>
      </n-form-item>
      <n-form-item label="买入日期" path="buyDate">
        <n-date-picker v-model:value="formModel.buyDate" type="date" placeholder="请选择买入日期" clearable />
      </n-form-item>
      <n-form-item label="涨跌提醒" path="alarm">
        <n-input-number v-model:value="formModel.alarm" min="0" placeholder="请输入涨跌报警值(%)">
          <template #suffix>
            %
          </template>
        </n-input-number>
      </n-form-item>
      <n-form-item label="股价提醒" path="alarmPrice">
        <n-input-number v-model:value="formModel.alarmPrice" min="0" placeholder="请输入股价报警值(¥)">
          <template #suffix>
            {{ formModel.code.indexOf("hk") >= 0 ? "HK$" : "¥" }}
          </template>
        </n-input-number>
      </n-form-item>
      <n-form-item label="股票排序" path="sort">
        <n-input-number v-model:value="formModel.sort" min="0" placeholder="请输入股价排序值">
        </n-input-number>
      </n-form-item>
      <n-form-item label="AI cron" path="cron">
        <n-input v-model:value="formModel.cron" placeholder="请输入cron表达式"/>
      </n-form-item>
    </n-form>
    <template #footer>
      <n-button type="primary"
                @click="updateCostPriceAndVolumeNew(formModel.code,formModel.costPrice,formModel.volume,formModel.alarm,formModel)">
        保存
      </n-button>
    </template>
  </n-modal>

  <n-modal v-model:show="addTabPane" title="添加分组" style="width: 400px;text-align: left" :preset="'card'">
    <n-form
      :model="addTabModel"
      size="medium"
      label-placement="left"
    >
      <n-grid :cols="2">
        <n-form-item-gi label="分组名称:" path="name" :span="5">
          <n-input v-model:value="addTabModel.name" style="width: 100%" placeholder="请输入分组名称"/>
        </n-form-item-gi>
        <n-form-item-gi label="分组排序:" path="sort" :span="5">
          <n-input-number v-model:value="addTabModel.sort" style="width: 100%" min="0"
                          placeholder="请输入分组排序值"></n-input-number>
        </n-form-item-gi>
      </n-grid>
    </n-form>
    <template #footer>
      <n-flex justify="end">
        <n-button type="primary" @click="saveTabPane">
          保存
        </n-button>
        <n-button type="warning" @click="addTabPane=false">
          取消
        </n-button>
      </n-flex>
    </template>
  </n-modal>
  <n-modal v-model:show="modalShow2" :title="data.name+' '+ data.changePercent+'%'" style="width: 1000px"
           :preset="'card'" @after-enter="handleFeishi" @after-leave="clearFeishi">
    <!--    <n-image :src="data.fenshiURL" />-->
    <div ref="kLineChartRef2" style="width: 1000px; height: 500px;"></div>
  </n-modal>
  <n-modal v-model:show="modalShow3" :title="data.name" style="width: 1000px" :preset="'card'"
           @after-enter="handleKLine">
    <!--    <n-image :src="data.kURL" />-->
    <div ref="kLineChartRef" style="width: 1000px; height: 500px;"></div>
  </n-modal>

  <n-modal transform-origin="center" v-model:show="modalShow4" preset="card" style="width: 800px;"
           :title="'['+data.name+']AI分析'">
    <n-spin size="small" :show="data.loading">
      <MdEditor v-if="enableEditor" :toolbars="toolbars" ref="mdEditorRef" style="height: 440px;text-align: left"
                :modelValue="data.airesult" :theme="theme">
        <template #defToolbars>
          <ExportPDF :file-name="data.name+'['+data.code+']AI分析报告'" style="text-align: left"
                     :modelValue="data.airesult" @onProgress="handleProgress"/>
        </template>
      </MdEditor>
      <MdPreview v-if="!enableEditor" ref="mdPreviewRef" style="height: 440px;text-align: left"
                 :modelValue="data.airesult" :theme="theme"/>
    </n-spin>
    <template #footer>
      <n-flex justify="space-between" ref="tipsRef">
        <n-text type="info" v-if="data.time">
          <n-tag v-if="data.modelName" type="warning" round :title="data.chatId" :bordered="false">
            {{ data.modelName }}
          </n-tag>
          {{ data.time }}
        </n-text>
        <n-text type="error">*AI分析结果仅供参考，请以实际行情为准。投资需谨慎，风险自担。</n-text>
      </n-flex>
    </template>
    <template #action>
      <n-flex justify="left" style="margin-bottom: 10px">
        <n-switch v-model:value="enableTools" :round="false">
          <template #checked>
            启用AI函数工具调用
          </template>
          <template #unchecked>
            不启用AI函数工具调用
          </template>
        </n-switch>
        <n-switch v-model:value="thinkingMode" :round="false">
          <template #checked>
            启用思考模式
          </template>
          <template #unchecked>
            不启用思考模式
          </template>
        </n-switch>
        <n-gradient-text type="error" style="margin-left: 10px">
          *AI函数工具调用可以增强AI获取数据的能力,但会消耗更多tokens。
        </n-gradient-text>
      </n-flex>
      <n-flex justify="space-between" style="margin-bottom: 10px">
        <n-select style="width: 31%" v-model:value="data.aiConfigId" label-field="name" value-field="ID"
                  :options="aiConfigs" placeholder="请选择AI模型服务配置"/>
        <n-select style="width: 31%" v-model:value="data.sysPromptId" label-field="name" value-field="ID"
                  :options="sysPromptOptions" placeholder="请选择系统提示词"/>
        <n-select style="width: 31%" v-model:value="data.question" label-field="name" value-field="content"
                  :options="userPromptOptions" placeholder="请选择用户提示词"/>
      </n-flex>
      <n-flex justify="right">
        <n-input v-model:value="data.question" style="text-align: left" clearable
                 type="textarea"
                 :show-count="true"
                 placeholder="请输入您的问题:例如{{stockName}}[{{stockCode}}]分析和总结"
                 :autosize="{
              minRows: 2,
              maxRows: 5
            }"
        />
        <!--        <n-button size="tiny" type="error" @click="enableEditor=!enableEditor">编辑/预览</n-button>-->
        <n-button size="tiny" type="warning" @click="aiReCheckStock(data.name,data.code)">开始AI分析</n-button>
        <n-button size="tiny" type="info" @click="saveAsImage(data.name,data.code)">保存为图片</n-button>
        <n-button size="tiny" type="success" @click="copyToClipboard">复制到剪切板</n-button>
        <n-button size="tiny" type="primary" @click="saveAsMarkdown">保存为Markdown文件</n-button>
        <n-button size="tiny" type="primary" @click="saveAsWord">保存为Word文件</n-button>
        <n-button size="tiny" type="error" @click="share(data.code,data.name)">分享到项目社区</n-button>
      </n-flex>
    </template>
  </n-modal>
  <n-modal v-model:show="modalShow5" :title="data.name+'资金趋势'" style="width: 1000px" :preset="'card'">
    <money-trend :code="data.code" :name="data.name" :days="360" :dark-theme="data.darkTheme"
                 :chart-height="500"></money-trend>
  </n-modal>

  <!-- AI总结对话框 -->
  <n-modal transform-origin="center" v-model:show="summaryModal" preset="card" style="width: 800px;"
           :title="'AI股票自选总结'">
    <n-spin size="small" :show="aiSummaryLoading">
      <MdPreview style="height: 440px;text-align: left" :modelValue="aiSummary" :theme="theme"/>
    </n-spin>
    <template #footer>
      <n-flex justify="space-between" ref="tipsRef">
        <n-text type="info" v-if="aiSummaryTime">
          <n-tag v-if="aiSummaryModelName" type="warning" round :title="aiSummaryChatId" :bordered="false">{{ aiSummaryModelName }}</n-tag>
          {{ aiSummaryTime }}
        </n-text>
        <n-text type="error">*AI分析结果仅供参考，请以实际行情为准。投资需谨慎，风险自担。</n-text>
      </n-flex>
    </template>
    <template #action>
      <n-flex justify="left" style="margin-bottom: 10px">
        <n-switch v-model:value="enableTools" :round="false">
          <template #checked>
            启用AI函数工具调用
          </template>
          <template #unchecked>
            不启用AI函数工具调用
          </template>
        </n-switch>
        <n-switch v-model:value="thinkingMode" :round="false">
          <template #checked>
            启用思考模式
          </template>
          <template #unchecked>
            不启用思考模式
          </template>
        </n-switch>
        <n-gradient-text type="error" style="margin-left: 10px">*AI函数工具调用可以增强AI获取数据的能力,但会消耗更多tokens。</n-gradient-text>
      </n-flex>
      <n-flex justify="space-between" style="margin-bottom: 10px">
        <n-select style="width: 32%" v-model:value="aiSummaryConfigId" label-field="name" value-field="ID"
                  :options="aiConfigs" placeholder="请选择AI模型服务配置"/>
        <n-select style="width: 32%" v-model:value="aiSummarySysPromptId" label-field="name" value-field="ID"
                  :options="sysPromptOptions" placeholder="请选择系统提示词"/>
        <n-select style="width: 32%" v-model:value="aiSummaryQuestion" label-field="name" value-field="content"
                  :options="userPromptOptions" placeholder="请选择用户提示词"/>
      </n-flex>
      <n-flex justify="right">
        <n-input v-model:value="aiSummaryQuestion" style="text-align: left" clearable
                 type="textarea"
                 :show-count="true"
                 placeholder="请输入您的问题:例如 总结和分析股票市场新闻中的投资机会"
                 :autosize="{
              minRows: 2,
              maxRows: 5
            }"
        />
        <n-button size="tiny" type="warning" @click="reAiSummary">再次总结</n-button>
        <n-button size="tiny" type="success" @click="copyAiSummaryToClipboard">复制到剪切板</n-button>
        <n-button size="tiny" type="primary" @click="saveAiSummaryAsMarkdown">保存为Markdown文件</n-button>
        <n-button size="tiny" type="error" @click="shareAiSummary">分享到项目社区</n-button>
      </n-flex>
    </template>
  </n-modal>

  <!-- AI总结按钮 -->
  <div style="position: fixed;bottom: 70px;right:5px;z-index: 10;" v-if="data.openAiEnable">
    <n-input-group>
      <n-button type="primary" @click="getAiSummary">
        <n-icon :component="PulseOutline"/> &nbsp;AI总结
      </n-button>
    </n-input-group>
  </div>

  <!-- AI选股按钮 -->
  <div style="position: fixed;bottom: 122px;right:5px;z-index: 10;" v-if="data.openAiEnable">
    <n-input-group>
      <n-button type="warning" @click="openAiStockSelect">
        <n-icon :component="ChatboxOutline"/> &nbsp;AI选股
      </n-button>
    </n-input-group>
  </div>

  <!-- 指标选股按钮 -->
  <div style="position: fixed;bottom: 174px;right:5px;z-index: 10;">
    <n-input-group>
      <n-button type="info" @click="openIndicatorStockSelect">
        <n-icon :component="PulseOutline"/> &nbsp;指标选股
      </n-button>
    </n-input-group>
  </div>

  <!-- AI选股对话框 -->
  <n-modal transform-origin="center" v-model:show="aiStockSelectModal" preset="card" style="width: 800px;"
           :title="'AI智能选股'"
           @after-leave="() => { isAiStockSelectMode.value = false; aiStockSelectLoading.value = false }">
    <n-spin size="small" :show="aiStockSelectLoading">
      <MdPreview style="height: 440px;text-align: left" :modelValue="aiStockSelectResult || (aiStockSelectLoading ? '正在分析中，请稍候...' : '')" :theme="theme"/>
    </n-spin>
    <template #footer>
      <n-flex justify="space-between">
        <n-text type="info" v-if="aiStockSelectTime">
          <n-tag v-if="aiStockSelectModelName" type="warning" round :title="aiStockSelectChatId" :bordered="false">{{ aiStockSelectModelName }}</n-tag>
          {{ aiStockSelectTime }}
        </n-text>
        <n-text type="error">*AI选股结果仅供参考，请以实际行情为准。投资需谨慎，风险自担。</n-text>
      </n-flex>
    </template>
    <template #action>
      <n-flex justify="left" style="margin-bottom: 10px">
        <n-switch v-model:value="enableTools" :round="false">
          <template #checked>
            启用AI函数工具调用
          </template>
          <template #unchecked>
            不启用AI函数工具调用
          </template>
        </n-switch>
        <n-switch v-model:value="thinkingMode" :round="false">
          <template #checked>
            启用思考模式
          </template>
          <template #unchecked>
            不启用思考模式
          </template>
        </n-switch>
        <n-gradient-text type="error" style="margin-left: 10px">*AI函数工具调用可以增强AI获取数据的能力,但会消耗更多tokens。</n-gradient-text>
      </n-flex>
      <n-flex justify="space-between" style="margin-bottom: 10px">
        <n-select style="width: 32%" v-model:value="aiStockSelectConfigId" label-field="name" value-field="ID"
                  :options="aiConfigs" placeholder="请选择AI模型服务配置"/>
        <n-select style="width: 32%" v-model:value="aiStockSelectSysPromptId" label-field="name" value-field="ID"
                  :options="sysPromptOptions" placeholder="请选择系统提示词"/>
        <n-text style="width: 32%" type="info">选股完成后会自动解析并添加到自选</n-text>
      </n-flex>
      <n-flex justify="right">
        <n-input v-model:value="aiStockSelectCondition" style="text-align: left" clearable
                 type="textarea"
                 :show-count="true"
                 placeholder="请输入选股条件，例如：市值最大的3个股票、市盈率小于20且市值大于100亿的股票"
                 :autosize="{
              minRows: 2,
              maxRows: 5
            }"
        />
        <n-button size="tiny" type="warning" @click="executeAiStockSelect" :loading="aiStockSelectLoading" :disabled="!aiStockSelectCondition.trim() || !aiStockSelectConfigId">开始选股</n-button>
      </n-flex>
    </template>
  </n-modal>

  <!-- 指标选股对话框 -->
  <n-modal transform-origin="center" v-model:show="indicatorStockSelectModal" preset="card" style="width: 90%; max-width: 1400px;"
           :title="'指标选股'">
    <n-spin size="small" :show="indicatorStockSelectLoading">
      <n-flex vertical style="height: 600px;">
        <n-input-group style="margin-bottom: 10px; --wails-draggable:no-drag">
          <n-input 
            v-model:value="indicatorStockSelectCondition" 
            placeholder="请输入选股指标或者要求，例如：换手率大于3%，量比大于2，涨幅大于2%小于7%"
            clearable
            @keyup.enter="executeIndicatorStockSelect"
          />
          <n-button type="primary" @click="executeIndicatorStockSelect" :loading="indicatorStockSelectLoading">搜索A股</n-button>
        </n-input-group>
        
        <n-flex justify="start" v-if="indicatorStockSelectTraceInfo" style="margin-bottom: 10px; --wails-draggable:no-drag">
          <n-ellipsis line-clamp="1" :tooltip="true">
            <n-text type="info" :bordered="false">选股条件：</n-text>
            <n-text type="warning" :bordered="true">{{ indicatorStockSelectTraceInfo }}</n-text>
            <template #tooltip>
              <div style="text-align: center;max-width: 580px">
                <n-text type="warning">{{ indicatorStockSelectTraceInfo }}</n-text>
              </div>
            </template>
          </n-ellipsis>
        </n-flex>
        
        <n-flex justify="space-between" style="margin-bottom: 10px; --wails-draggable:no-drag" v-if="indicatorStockSelectDataList.length > 0">
          <n-text type="info">
            共找到 <n-tag type="info" :bordered="false">{{ indicatorStockSelectDataList.length }}</n-tag> 只股票
          </n-text>
          <n-button type="success" @click="followAllIndicatorStocks" :disabled="indicatorStockSelectDataList.length === 0">
            一键关注全部
          </n-button>
        </n-flex>
        
        <n-data-table
          v-if="indicatorStockSelectDataList.length > 0"
          :striped="true"
          :max-height="'500px'"
          size="medium"
          :columns="indicatorStockSelectColumns"
          :data="indicatorStockSelectDataList"
          :pagination="{pageSize: 20}"
          :scroll-x="indicatorStockSelectTableScrollX"
          :render-cell="(value, rowData, column) => {
            if(column.key=='SECURITY_CODE'||column.key=='SERIAL'){
              return h(NText, { type: 'info',border: false }, { default: () => `${value}` })
            }
            if (isNumeric(value)) {
              let type='info';
              if (Number(value)<0){
                type='success';
              }
              if(Number(value)>=0&&Number(value)<=5){
                type='warning';
              }
              if (Number(value)>5){
                type='error';
              }
              return h(NText, { type: type }, { default: () => `${value}` })
            }else{
              if(column.key=='SECURITY_SHORT_NAME'){
                return h(NButton, { type: 'info',bordered: false ,size:'small',onClick:()=>{
                  OpenURL(`https://quote.eastmoney.com/${rowData.MARKET_SHORT_NAME}${rowData.SECURITY_CODE}.html#fullScreenChart`)
                }}, { default: () => `${value}` })
              }else{
                return h(NText, { type: 'info' }, { default: () => `${value}` })
              }
            }
          }"
        />
        
        <div v-else-if="!indicatorStockSelectLoading" style="text-align: center; padding: 50px;">
          <n-text type="info">请输入选股条件并点击"搜索A股"按钮开始选股</n-text>
        </div>
      </n-flex>
    </n-spin>
    <template #footer>
      <n-text type="error">*选股结果仅供参考，请以实际行情为准。投资需谨慎，风险自担。</n-text>
    </template>
  </n-modal>
</template>

<style scoped>
.md-editor-preview h3 {
  text-align: center !important;
}

.md-editor-preview p {
  text-align: left !important;
}

/* 添加闪烁效果的CSS类 */
.blink-border {
  animation: blink-border 1s linear infinite;
  border: 4px solid transparent;
}

@keyframes blink-border {
  0% {
    border-color: red;
  }
  50% {
    border-color: transparent;
  }
  100% {
    border-color: red;
  }
}

/* 所有标签的通用样式 */
:deep(.n-tabs-nav .n-tabs-tab) {
  position: relative;
  cursor: pointer;
}

/* 可拖拽标签的样式 */
:deep(.n-tabs-nav .n-tabs-tab[draggable="true"]) {
  user-select: none;
  cursor: move;
}

.tab-drag-over {
  background-color: #e6f7ff !important;
  border: 2px dashed #1890ff !important;
  transform: scale(1.02);
  transition: all 0.2s ease;
  z-index: 10;
}

.tab-drag-over::after {
  content: "";
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  z-index: -1;
}

.tab-dragging {
  opacity: 0.5;
}
</style>
