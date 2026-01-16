<script setup>
import { computed, nextTick, reactive, ref } from 'vue'
import {
  NAutoComplete,
  NAlert,
  NButton,
  NCard,
  NDatePicker,
  NEmpty,
  NFormItem,
  NGrid,
  NGi,
  NIcon,
  NInput,
  NInputNumber,
  NList,
  NListItem,
  NScrollbar,
  NSpace,
  NStatistic,
  NText,
  useMessage,
} from 'naive-ui'
import { BookOutline, CalculatorOutline, SearchOutline, CashOutline } from '@vicons/ionicons5'
import { GetStockList, GetStockKLine } from '../../wailsjs/go/main/App'
import { format } from 'date-fns'
import KLineChart from './KLineChart.vue'

const message = useMessage()

// 当前选中的学习主题
const activeTopic = ref('补仓')

// 学习主题列表
const topics = [
  { key: '补仓', label: '补仓策略', icon: CalculatorOutline },
  { key: '分红', label: '分红策略', icon: CashOutline },
  { key: 'K线', label: 'K线基础', icon: BookOutline },
  { key: '技术指标', label: '技术指标', icon: BookOutline },
  { key: '基本面', label: '基本面分析', icon: BookOutline },
  { key: '风险管理', label: '风险管理', icon: BookOutline },
]

// 股票选择相关状态
const selectedStock = ref(null) // { code: '', name: '' }
const stockSearchValue = ref('')
const stockOptions = ref([])
const stockKLineData = ref([]) // K线数据缓存
const stockListCache = ref([]) // 股票列表缓存

// 补仓计算器相关状态
const costCalculator = reactive({
  // 初始持仓
  initialPrice: 10, // 初始买入价格
  initialVolume: 1000, // 初始买入数量
  initialDate: null, // 初始买入日期
  // 补仓记录
  additions: [],
  // 当前价格
  currentPrice: 9,
  // 下一个补仓记录的ID
  nextId: 1
})

// 按日期排序的补仓记录
const sortedAdditions = computed(() => {
  return [...costCalculator.additions].sort((a, b) => {
    if (!a.date && !b.date) return 0
    if (!a.date) return 1
    if (!b.date) return -1
    return new Date(a.date) - new Date(b.date)
  })
})

// 计算平均成本
const averageCost = computed(() => {
  let totalCost = costCalculator.initialPrice * costCalculator.initialVolume
  let totalVolume = costCalculator.initialVolume
  
  costCalculator.additions.forEach(add => {
    if (add.price > 0 && add.volume > 0) {
      totalCost += add.price * add.volume
      totalVolume += add.volume
    }
  })
  
  return totalVolume > 0 ? (totalCost / totalVolume).toFixed(2) : costCalculator.initialPrice.toFixed(2)
})

// 计算总持仓数量
const totalVolume = computed(() => {
  const validAdditions = costCalculator.additions.filter(add => add.price > 0 && add.volume > 0)
  return costCalculator.initialVolume + validAdditions.reduce((sum, add) => sum + add.volume, 0)
})

// 计算总投入成本
const totalCost = computed(() => {
  const initialCost = costCalculator.initialPrice * costCalculator.initialVolume
  const validAdditions = costCalculator.additions.filter(add => add.price > 0 && add.volume > 0)
  const additionsCost = validAdditions.reduce((sum, add) => sum + add.price * add.volume, 0)
  return (initialCost + additionsCost).toFixed(2)
})

// 获取最新价格（从K线数据或手动输入）
const latestPrice = computed(() => {
  if (selectedStock.value && stockKLineData.value && stockKLineData.value.length > 0) {
    // 获取最新的收盘价
    const sortedData = [...stockKLineData.value].sort((a, b) => {
      const dateA = a.day || a.date || a.trade_date || ''
      const dateB = b.day || b.date || b.trade_date || ''
      return dateB.localeCompare(dateA)
    })
    if (sortedData.length > 0) {
      return parseFloat(sortedData[0].close || sortedData[0].Close || costCalculator.currentPrice)
    }
  }
  return costCalculator.currentPrice
})

// 计算当前市值
const currentMarketValue = computed(() => {
  return (totalVolume.value * latestPrice.value).toFixed(2)
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

// 初始化股票列表缓存
async function initStockList() {
  if (stockListCache.value.length === 0) {
    try {
      const result = await GetStockList('')
      if (result && Array.isArray(result)) {
        // 过滤掉指数，只保留个股
        stockListCache.value = result.filter(item => !isIndex(item))
      }
    } catch (error) {
      console.error('加载股票列表失败:', error)
    }
  }
}

// 判断是否为指数（通过代码格式和名称）
function isIndex(item) {
  if (!item) return false
  
  // 如果名称包含"指数"，肯定是指数
  if (item.name && (item.name.includes('指数') || item.name.includes('Index'))) {
    return true
  }
  
  const tsCode = item.ts_code || ''
  if (!tsCode) return false
  
  const code = tsCode.toLowerCase()
  
  // 指数代码模式：
  // sh000xxx - 上证指数系列
  // sz399xxx - 深证指数系列
  // 个股代码模式：
  // sh600xxx, sh601xxx, sh603xxx, sh605xxx - 上海主板
  // sh688xxx - 科创板
  // sz000xxx (但000001-000999可能是指数，需要特殊处理)
  // sz002xxx - 中小板
  // sz300xxx - 创业板
  
  // 明确是指数的模式
  if (code.startsWith('sh000') || code.startsWith('sz399')) {
    return true
  }
  
  // 对于 sz000xxx，需要判断：
  // 000001-000999 通常是指数（如000001是深证成指）
  // 但有些个股也可能在这个范围内，不过大部分是指数
  if (code.startsWith('sz000')) {
    const numPart = parseInt(code.replace('sz000', ''))
    // 000001-000999 通常是指数
    if (numPart >= 1 && numPart <= 999) {
      return true
    }
  }
  
  return false
}

// 股票搜索
async function searchStock(keyword) {
  if (!keyword || keyword.length < 1) {
    stockOptions.value = []
    return
  }
  
  try {
    // 先初始化股票列表
    await initStockList()
    
    // 从缓存中搜索，过滤掉指数
    const filtered = stockListCache.value.filter(item => {
      // 过滤掉指数
      if (isIndex(item)) {
        return false
      }
      // 匹配名称或代码
      return (item.name && item.name.includes(keyword)) || 
             (item.ts_code && item.ts_code.includes(keyword.toUpperCase()))
    })
    
    // 限制结果数量
    const limitedResults = filtered.slice(0, 50)
    
    stockOptions.value = limitedResults.map(item => ({
      label: `${item.name} - ${item.ts_code}`,
      value: item.ts_code,
      name: item.name,
      code: item.ts_code
    }))
    
    // 如果缓存中没有结果，尝试从API获取
    if (stockOptions.value.length === 0 && keyword.length >= 2) {
      const result = await GetStockList(keyword)
      if (result && Array.isArray(result) && result.length > 0) {
        // 过滤掉指数
        const stockOnly = result.filter(item => !isIndex(item))
        
        stockOptions.value = stockOnly.slice(0, 50).map(item => ({
          label: `${item.name} - ${item.ts_code}`,
          value: item.ts_code,
          name: item.name,
          code: item.ts_code
        }))
        // 更新缓存（只缓存个股）
        const newStocks = stockOnly.filter(item => 
          !stockListCache.value.some(cached => cached.ts_code === item.ts_code)
        )
        stockListCache.value = [...stockListCache.value, ...newStocks]
      }
    }
  } catch (error) {
    console.error('搜索股票失败:', error)
    message.error('搜索股票失败：' + (error.message || error))
    stockOptions.value = []
  }
}

// 选择股票
async function onStockSelect(option) {
  if (option && option.value) {
    selectedStock.value = {
      code: option.value,
      name: option.name || option.label.split(' - ')[0]
    }
    stockSearchValue.value = `${option.name || option.label.split(' - ')[0]} - ${option.value}`
    // 加载K线数据
    await loadStockKLineData(option.value, option.name || option.label.split(' - ')[0])
    message.success(`已选择股票：${option.name || option.label.split(' - ')[0]}`)
  }
}

// 加载股票K线数据
async function loadStockKLineData(code, name) {
  try {
    message.loading('正在加载历史K线数据...', { duration: 0 })
    const result = await GetStockKLine(code, name, 365)
    // GetStockKLine返回的可能是数组或对象
    if (Array.isArray(result)) {
      stockKLineData.value = result
      message.destroyAll()
      message.success(`历史数据加载成功，共 ${result.length} 条记录`)
    } else if (result && result.data && Array.isArray(result.data)) {
      stockKLineData.value = result.data
      message.destroyAll()
      message.success(`历史数据加载成功，共 ${result.data.length} 条记录`)
    } else {
      message.destroyAll()
      message.warning('未获取到历史数据，请检查股票代码是否正确')
      stockKLineData.value = []
    }
  } catch (error) {
    message.destroyAll()
    message.error('加载历史数据失败：' + (error.message || error))
    stockKLineData.value = []
  }
}

// 根据日期获取股票价格
function getPriceByDate(date) {
  if (!date || !stockKLineData.value || stockKLineData.value.length === 0) {
    return null
  }
  
  // 处理日期格式
  let dateStr = ''
  if (date instanceof Date) {
    dateStr = format(date, 'yyyy-MM-dd')
  } else if (typeof date === 'number') {
    dateStr = format(new Date(date), 'yyyy-MM-dd')
  } else if (typeof date === 'string') {
    dateStr = date.substring(0, 10) // 取前10个字符 yyyy-MM-dd
  }
  
  // K线数据格式可能是不同的，需要根据实际数据结构调整
  const kline = stockKLineData.value.find(item => {
    const itemDate = item.day || item.date || item.trade_date || ''
    // 处理不同的日期格式
    const normalizedItemDate = itemDate.toString().substring(0, 10)
    return normalizedItemDate === dateStr
  })
  
  if (kline) {
    // 返回收盘价，尝试不同的字段名
    const closePrice = kline.close || kline.Close || kline[2] || 0
    const price = parseFloat(closePrice)
    return price > 0 ? price : null
  }
  
  // 如果找不到精确日期，找最近的交易日（小于等于目标日期）
  const sortedData = [...stockKLineData.value].sort((a, b) => {
    const dateA = (a.day || a.date || a.trade_date || '').toString().substring(0, 10)
    const dateB = (b.day || b.date || b.trade_date || '').toString().substring(0, 10)
    return dateB.localeCompare(dateA)
  })
  
  for (const item of sortedData) {
    const itemDate = (item.day || item.date || item.trade_date || '').toString().substring(0, 10)
    if (itemDate && itemDate <= dateStr) {
      const closePrice = item.close || item.Close || item[2] || 0
      const price = parseFloat(closePrice)
      if (price > 0) {
        return price
      }
    }
  }
  
  return null
}

// 更新初始买入价格（根据日期）
async function updateInitialPriceByDate() {
  if (costCalculator.initialDate && selectedStock.value) {
    const price = getPriceByDate(costCalculator.initialDate)
    if (price && price > 0) {
      costCalculator.initialPrice = price
      message.success(`已自动填充初始买入价格：${price.toFixed(2)} 元`)
    } else {
      message.warning('未找到该日期的价格数据，请手动输入价格')
    }
  }
}

// 添加补仓记录
function addCostAddition() {
  if (costCalculator.additions.length === 0 || 
      (costCalculator.additions[costCalculator.additions.length - 1].price > 0 && 
       costCalculator.additions[costCalculator.additions.length - 1].volume > 0)) {
    costCalculator.additions.push({
      price: 0,
      volume: 0,
      date: null,
      id: costCalculator.nextId++
    })
  }
}

// 更新补仓价格（根据日期）
async function updateAdditionPriceByDate(addition) {
  if (addition.date && selectedStock.value) {
    const price = getPriceByDate(addition.date)
    if (price && price > 0) {
      addition.price = price
      message.success(`已自动填充补仓价格：${price.toFixed(2)} 元`)
    } else {
      message.warning('未找到该日期的价格数据，请手动输入价格')
    }
  }
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
  <n-grid :cols="24" :x-gap="16" style="height: calc(100vh - 100px); padding: 16px;">
    <!-- 左侧知识列表 -->
    <n-gi :span="6">
      <n-card title="股票知识" :bordered="true" style="height: 100%;">
        <n-scrollbar style="max-height: calc(100vh - 180px);">
          <n-list>
            <n-list-item
              v-for="topic in topics"
              :key="topic.key"
              :class="{ 'active-topic': activeTopic === topic.key }"
              style="cursor: pointer; padding: 12px; border-radius: 4px; margin-bottom: 4px;"
              :style="activeTopic === topic.key ? { backgroundColor: 'var(--n-color-hover)' } : {}"
              @click="activeTopic = topic.key"
            >
              <template #prefix>
                <n-icon :component="topic.icon" style="margin-right: 8px;" />
              </template>
              {{ topic.label }}
            </n-list-item>
          </n-list>
        </n-scrollbar>
      </n-card>
    </n-gi>

    <!-- 右侧学习内容 -->
    <n-gi :span="18">
      <n-card :bordered="true" style="height: 100%;">
        <n-scrollbar style="max-height: calc(100vh - 180px);">
          <!-- 补仓内容 -->
          <template v-if="activeTopic === '补仓'">
            <n-grid :cols="1" :x-gap="16" :y-gap="16">
              <!-- 股票选择 -->
              <n-gi>
                <n-card title="选择股票" :bordered="true">
                  <n-space vertical>
                    <n-auto-complete
                      v-model:value="stockSearchValue"
                      :options="stockOptions"
                      placeholder="输入股票名称或代码搜索"
                      clearable
                      @update-value="searchStock"
                      @select="onStockSelect"
                      style="width: 100%"
                    >
                      <template #prefix>
                        <n-icon :component="SearchOutline" />
                      </template>
                    </n-auto-complete>
                    <n-text v-if="selectedStock" type="success">
                      已选择：{{ selectedStock.name }} ({{ selectedStock.code }})
                    </n-text>
                    <n-text v-else type="info" depth="3">
                      请搜索并选择股票，将自动加载历史K线数据
                    </n-text>
                  </n-space>
                </n-card>
              </n-gi>
              
              <!-- K线图表 -->
              <n-gi v-if="selectedStock">
                <n-card title="日K线图" :bordered="true">
                  <k-line-chart 
                    :code="selectedStock.code" 
                    :name="selectedStock.name" 
                    :k-days="365"
                    :chart-height="400"
                    :dark-theme="false"
                  />
                </n-card>
              </n-gi>
              
              <!-- 补仓成本计算器 -->
              <n-gi>
                <n-card title="补仓成本计算器" :bordered="true">
                  <n-space vertical size="large">
                    <!-- 初始持仓 -->
                    <n-card title="初始持仓" size="small" :bordered="true">
                      <n-grid :cols="4" :x-gap="12">
                        <n-gi>
                          <n-form-item label="买入日期">
                            <n-date-picker
                              v-model:value="costCalculator.initialDate"
                              type="date"
                              placeholder="选择买入日期"
                              style="width: 100%"
                              @update:value="updateInitialPriceByDate"
                            />
                          </n-form-item>
                        </n-gi>
                        <n-gi>
                          <n-form-item label="买入价格（元）">
                            <n-input-number v-model:value="costCalculator.initialPrice" :min="0.01" :step="0.01" :precision="2" style="width: 100%"/>
                          </n-form-item>
                        </n-gi>
                        <n-gi>
                          <n-form-item label="买入数量（股）">
                            <n-input-number v-model:value="costCalculator.initialVolume" :min="1" :step="100" style="width: 100%"/>
                          </n-form-item>
                        </n-gi>
                        <n-gi>
                          <n-form-item label="初始成本">
                            <n-text strong style="font-size: 16px;">
                              {{ (costCalculator.initialPrice * costCalculator.initialVolume).toFixed(2) }} 元
                            </n-text>
                          </n-form-item>
                        </n-gi>
                      </n-grid>
                    </n-card>
                    
                    <!-- 补仓记录 -->
                    <n-card title="补仓记录" size="small" :bordered="true">
                      <template #header-extra>
                        <n-button size="small" type="primary" @click="addCostAddition">添加补仓</n-button>
                        <n-button size="small" type="error" @click="clearAllAdditions" :disabled="costCalculator.additions.length === 0" style="margin-left: 8px;">清空</n-button>
                      </template>
                      <n-space vertical v-if="costCalculator.additions.length > 0">
                        <n-card v-for="(add, index) in sortedAdditions" :key="add.id" size="small" :bordered="true" style="margin-bottom: 8px;">
                          <n-grid :cols="5" :x-gap="12">
                            <n-gi>
                              <n-form-item label="补仓日期">
                                <n-date-picker
                                  v-model:value="add.date"
                                  type="date"
                                  placeholder="选择补仓日期"
                                  style="width: 100%"
                                  @update:value="() => updateAdditionPriceByDate(add)"
                                />
                              </n-form-item>
                            </n-gi>
                            <n-gi>
                              <n-form-item label="补仓价格（元）">
                                <n-input-number 
                                  v-model:value="add.price" 
                                  :min="0.01" 
                                  :step="0.01" 
                                  :precision="2" 
                                  style="width: 100%"
                                  class="cost-addition-input"
                                />
                              </n-form-item>
                            </n-gi>
                            <n-gi>
                              <n-form-item label="补仓数量（股）">
                                <n-input-number 
                                  v-model:value="add.volume" 
                                  :min="1" 
                                  :step="100" 
                                  style="width: 100%"
                                />
                              </n-form-item>
                            </n-gi>
                            <n-gi>
                              <n-form-item label="补仓成本">
                                <n-text>{{ (add.price * add.volume).toFixed(2) }} 元</n-text>
                              </n-form-item>
                            </n-gi>
                            <n-gi>
                              <n-button size="small" type="error" @click="removeCostAddition(add.id)">删除</n-button>
                            </n-gi>
                          </n-grid>
                        </n-card>
                      </n-space>
                      <n-empty v-else description='暂无补仓记录，点击"添加补仓"开始' size="small"/>
                    </n-card>
                    
                    <!-- 当前价格 -->
                    <n-card title="当前价格" size="small" :bordered="true">
                      <n-space vertical>
                        <n-form-item label="当前股价（元）">
                          <n-input-number v-model:value="costCalculator.currentPrice" :min="0.01" :step="0.01" :precision="2" style="width: 200px"/>
                        </n-form-item>
                        <n-text v-if="selectedStock && stockKLineData.length > 0" type="info" depth="3">
                          最新收盘价：{{ latestPrice.toFixed(2) }} 元
                        </n-text>
                      </n-space>
                    </n-card>
                    
                    <!-- 计算结果 -->
                    <n-card title="计算结果" size="small" :bordered="true">
                      <n-grid :cols="2" :x-gap="12" :y-gap="12">
                        <n-gi>
                          <n-statistic label="平均成本" :value="averageCost">
                            <template #suffix>元/股</template>
                          </n-statistic>
                        </n-gi>
                        <n-gi>
                          <n-statistic label="总持仓数量" :value="totalVolume">
                            <template #suffix>股</template>
                          </n-statistic>
                        </n-gi>
                        <n-gi>
                          <n-statistic label="总投入成本" :value="totalCost">
                            <template #suffix>元</template>
                          </n-statistic>
                        </n-gi>
                        <n-gi>
                          <n-statistic label="当前市值" :value="currentMarketValue">
                            <template #suffix>元</template>
                          </n-statistic>
                        </n-gi>
                        <n-gi>
                          <n-statistic 
                            :label="parseFloat(profitLoss) >= 0 ? '盈利' : '亏损'" 
                            :value="Math.abs(parseFloat(profitLoss))"
                            :value-style="{ color: parseFloat(profitLoss) >= 0 ? '#18a058' : '#d03050' }"
                          >
                            <template #suffix>元</template>
                          </n-statistic>
                        </n-gi>
                        <n-gi>
                          <n-statistic 
                            :label="parseFloat(profitLossPercent) >= 0 ? '盈利率' : '亏损率'" 
                            :value="Math.abs(parseFloat(profitLossPercent))"
                            :value-style="{ color: parseFloat(profitLossPercent) >= 0 ? '#18a058' : '#d03050' }"
                          >
                            <template #suffix>%</template>
                          </n-statistic>
                        </n-gi>
                      </n-grid>
                    </n-card>
                    
                    <!-- 目标成本计算 -->
                    <n-card title="目标成本计算" size="small" :bordered="true">
                      <n-space vertical>
                        <n-form-item label="目标成本（元/股）">
                          <n-input-number v-model:value="targetCostInput" :min="0.01" :step="0.01" :precision="2" style="width: 200px"/>
                          <n-button type="primary" @click="calculateTarget" style="margin-left: 12px;">计算</n-button>
                        </n-form-item>
                        <n-alert v-if="targetCostResult" :type="targetCostResult.volume > 0 ? 'info' : 'warning'">
                          {{ targetCostResult.message }}
                          <template v-if="targetCostResult.volume > 0">
                            <br/>补仓后平均成本：{{ ((parseFloat(totalCost) + parseFloat(targetCostResult.cost)) / (totalVolume + targetCostResult.volume)).toFixed(2) }} 元/股
                          </template>
                        </n-alert>
                      </n-space>
                    </n-card>
                  </n-space>
                </n-card>
              </n-gi>
              
              <!-- 补仓策略说明 -->
              <n-gi>
                <n-card title="补仓策略说明" :bordered="true">
                  <n-space vertical size="large">
                    <n-card title="什么是补仓？" size="small" :bordered="true">
                      <n-text>
                        补仓是指在持有股票后，当股价下跌时继续买入该股票，以降低平均持仓成本的行为。
                      </n-text>
                    </n-card>
                    
                    <n-card title="补仓的目的" size="small" :bordered="true">
                      <ul>
                        <li>降低平均持仓成本：通过低价补仓，可以拉低整体持仓成本</li>
                        <li>增加持仓数量：在看好股票长期走势时，通过补仓增加持仓</li>
                        <li>摊薄成本：当股价下跌时，补仓可以摊薄之前的买入成本</li>
                      </ul>
                    </n-card>
                    
                    <n-card title="补仓策略" size="small" :bordered="true">
                      <n-space vertical>
                        <n-card title="1. 等额补仓法" size="small" :bordered="true">
                          <n-text>
                            每次补仓使用相同的金额。例如：初始买入1000股，每次补仓都投入相同的金额（如10000元）。
                            <br/><strong>优点：</strong>操作简单，风险可控
                            <br/><strong>缺点：</strong>在持续下跌时，补仓效果有限
                          </n-text>
                        </n-card>
                        
                        <n-card title="2. 等量补仓法" size="small" :bordered="true">
                          <n-text>
                            每次补仓买入相同的数量。例如：初始买入1000股，每次补仓都买入1000股。
                            <br/><strong>优点：</strong>操作简单
                            <br/><strong>缺点：</strong>在股价持续下跌时，需要更多资金
                          </n-text>
                        </n-card>
                        
                        <n-card title="3. 金字塔补仓法" size="small" :bordered="true">
                          <n-text>
                            随着股价下跌，逐步增加补仓数量。例如：第一次补仓1000股，第二次补仓2000股，第三次补仓3000股。
                            <br/><strong>优点：</strong>在低位买入更多，成本降低效果明显
                            <br/><strong>缺点：</strong>需要较多资金，风险较大
                          </n-text>
                        </n-card>
                        
                        <n-card title="4. 倒金字塔补仓法" size="small" :bordered="true">
                          <n-text>
                            随着股价下跌，逐步减少补仓数量。例如：第一次补仓3000股，第二次补仓2000股，第三次补仓1000股。
                            <br/><strong>优点：</strong>在相对高位少买，低位多买
                            <br/><strong>缺点：</strong>需要判断股价走势
                          </n-text>
                        </n-card>
                      </n-space>
                    </n-card>
                    
                    <n-card title="补仓注意事项" size="small" :bordered="true">
                      <n-alert type="warning" style="margin-bottom: 12px;">
                        <ul>
                          <li><strong>不要盲目补仓：</strong>补仓前要分析股票的基本面和技术面，确认是否值得继续持有</li>
                          <li><strong>控制仓位：</strong>不要将所有资金都用于补仓，要保留一定的资金应对风险</li>
                          <li><strong>设置止损：</strong>如果股票基本面恶化，要及时止损，不要盲目补仓</li>
                          <li><strong>分批补仓：</strong>不要一次性补仓太多，可以分批进行，降低风险</li>
                          <li><strong>关注市场环境：</strong>在整体市场下跌时，要谨慎补仓</li>
                        </ul>
                      </n-alert>
                    </n-card>
                    
                    <n-card title="补仓计算公式" size="small" :bordered="true">
                      <n-text>
                        <strong>平均成本 = (初始成本 + 所有补仓成本之和) / (初始数量 + 所有补仓数量之和)</strong>
                        <br/><br/>
                        <strong>示例：</strong>
                        <br/>初始：10元买入1000股，成本10000元
                        <br/>第一次补仓：9元买入1000股，成本9000元
                        <br/>第二次补仓：8元买入1000股，成本8000元
                        <br/>平均成本 = (10000 + 9000 + 8000) / (1000 + 1000 + 1000) = 27000 / 3000 = 9元/股
                      </n-text>
                    </n-card>
                    
                    <n-card title="实际补仓盈利案例" size="small" :bordered="true">
                      <n-space vertical size="medium">
                        <n-alert type="info" title="案例背景">
                          假设您看好某只成长股，但买入后股价出现回调，通过合理的补仓策略最终实现盈利。
                        </n-alert>
                        
                        <n-card title="📊 操作时间线" size="small" :bordered="true">
                          <n-text>
                            <strong>第1次买入（初始建仓）：</strong>
                            <br/>时间：2024年1月15日
                            <br/>价格：20.00元/股
                            <br/>数量：1000股
                            <br/>投入资金：20,000元
                            <br/>持仓成本：20.00元/股
                            <br/>
                            <br/><strong>第2次补仓（股价回调）：</strong>
                            <br/>时间：2024年2月10日（股价跌至18元）
                            <br/>价格：18.00元/股
                            <br/>数量：1000股
                            <br/>投入资金：18,000元
                            <br/>累计持仓：2000股
                            <br/>平均成本：(20,000 + 18,000) / 2000 = <strong>19.00元/股</strong>
                            <br/>
                            <br/><strong>第3次补仓（继续下跌）：</strong>
                            <br/>时间：2024年3月5日（股价跌至16元）
                            <br/>价格：16.00元/股
                            <br/>数量：1500股（采用金字塔补仓法，低位多买）
                            <br/>投入资金：24,000元
                            <br/>累计持仓：3500股
                            <br/>平均成本：(20,000 + 18,000 + 24,000) / 3500 = <strong>17.71元/股</strong>
                            <br/>
                            <br/><strong>第4次补仓（底部区域）：</strong>
                            <br/>时间：2024年3月25日（股价跌至14元）
                            <br/>价格：14.00元/股
                            <br/>数量：2000股（继续加仓）
                            <br/>投入资金：28,000元
                            <br/>累计持仓：5500股
                            <br/>平均成本：(20,000 + 18,000 + 24,000 + 28,000) / 5500 = <strong>16.36元/股</strong>
                          </n-text>
                        </n-card>
                        
                        <n-card title="💰 盈利情况分析" size="small" :bordered="true">
                          <n-text>
                            <strong>总投入成本：</strong>20,000 + 18,000 + 24,000 + 28,000 = <strong>90,000元</strong>
                            <br/><strong>总持仓数量：</strong>5,500股
                            <br/><strong>平均持仓成本：</strong>16.36元/股
                            <br/>
                            <br/><strong>假设股价回升至18.50元（接近初始买入价）：</strong>
                            <br/>当前市值：5,500股 × 18.50元 = 101,750元
                            <br/>盈利金额：101,750 - 90,000 = <strong style="color: #18a058;">+11,750元</strong>
                            <br/>盈利率：(11,750 / 90,000) × 100% = <strong style="color: #18a058;">+13.06%</strong>
                            <br/>
                            <br/><strong>关键点分析：</strong>
                            <br/>✓ 如果不补仓，在20元买入1000股，当股价回到18.50元时，仍亏损7.5%
                            <br/>✓ 通过补仓，平均成本降至16.36元，股价回升至18.50元时盈利13.06%
                            <br/>✓ 补仓策略成功将亏损转为盈利
                          </n-text>
                        </n-card>
                        
                        <n-card title="📈 不同价格点的盈亏情况" size="small" :bordered="true">
                          <n-text>
                            <table style="width: 100%; border-collapse: collapse; margin-top: 8px;">
                              <thead>
                                <tr style="background-color: var(--n-color-hover);">
                                  <th style="padding: 8px; border: 1px solid var(--n-border-color); text-align: left;">股价</th>
                                  <th style="padding: 8px; border: 1px solid var(--n-border-color); text-align: right;">市值</th>
                                  <th style="padding: 8px; border: 1px solid var(--n-border-color); text-align: right;">盈亏</th>
                                  <th style="padding: 8px; border: 1px solid var(--n-border-color); text-align: right;">盈利率</th>
                                </tr>
                              </thead>
                              <tbody>
                                <tr>
                                  <td style="padding: 8px; border: 1px solid var(--n-border-color);">14.00元（最低点）</td>
                                  <td style="padding: 8px; border: 1px solid var(--n-border-color); text-align: right;">77,000元</td>
                                  <td style="padding: 8px; border: 1px solid var(--n-border-color); text-align: right; color: #d03050;">-13,000元</td>
                                  <td style="padding: 8px; border: 1px solid var(--n-border-color); text-align: right; color: #d03050;">-14.44%</td>
                                </tr>
                                <tr>
                                  <td style="padding: 8px; border: 1px solid var(--n-border-color);">16.36元（成本价）</td>
                                  <td style="padding: 8px; border: 1px solid var(--n-border-color); text-align: right;">90,000元</td>
                                  <td style="padding: 8px; border: 1px solid var(--n-border-color); text-align: right;">0元</td>
                                  <td style="padding: 8px; border: 1px solid var(--n-border-color); text-align: right;">0%</td>
                                </tr>
                                <tr>
                                  <td style="padding: 8px; border: 1px solid var(--n-border-color);">18.00元</td>
                                  <td style="padding: 8px; border: 1px solid var(--n-border-color); text-align: right;">99,000元</td>
                                  <td style="padding: 8px; border: 1px solid var(--n-border-color); text-align: right; color: #18a058;">+9,000元</td>
                                  <td style="padding: 8px; border: 1px solid var(--n-border-color); text-align: right; color: #18a058;">+10.00%</td>
                                </tr>
                                <tr>
                                  <td style="padding: 8px; border: 1px solid var(--n-border-color);">18.50元</td>
                                  <td style="padding: 8px; border: 1px solid var(--n-border-color); text-align: right;">101,750元</td>
                                  <td style="padding: 8px; border: 1px solid var(--n-border-color); text-align: right; color: #18a058;">+11,750元</td>
                                  <td style="padding: 8px; border: 1px solid var(--n-border-color); text-align: right; color: #18a058;">+13.06%</td>
                                </tr>
                                <tr>
                                  <td style="padding: 8px; border: 1px solid var(--n-border-color);">20.00元（初始买入价）</td>
                                  <td style="padding: 8px; border: 1px solid var(--n-border-color); text-align: right;">110,000元</td>
                                  <td style="padding: 8px; border: 1px solid var(--n-border-color); text-align: right; color: #18a058;">+20,000元</td>
                                  <td style="padding: 8px; border: 1px solid var(--n-border-color); text-align: right; color: #18a058;">+22.22%</td>
                                </tr>
                              </tbody>
                            </table>
                          </n-text>
                        </n-card>
                        
                        <n-card title="💡 案例启示" size="small" :bordered="true">
                          <n-alert type="success" style="margin-bottom: 12px;">
                            <ul>
                              <li><strong>补仓时机很重要：</strong>在股价下跌过程中，分批补仓可以降低平均成本</li>
                              <li><strong>金字塔补仓法有效：</strong>在低位买入更多，可以更有效地降低平均成本</li>
                              <li><strong>需要足够的资金：</strong>补仓需要预留足够的资金，不能一次性用完</li>
                              <li><strong>基本面要良好：</strong>补仓的前提是股票基本面没有恶化，只是短期回调</li>
                              <li><strong>要有耐心：</strong>补仓后需要等待股价回升，可能需要较长时间</li>
                              <li><strong>设置止损：</strong>如果股票基本面恶化，要及时止损，不要盲目补仓</li>
                            </ul>
                          </n-alert>
                        </n-card>
                      </n-space>
                    </n-card>
                  </n-space>
                </n-card>
              </n-gi>
            </n-grid>
          </template>

          <!-- 分红策略内容 -->
          <template v-else-if="activeTopic === '分红'">
            <n-grid :cols="1" :x-gap="16" :y-gap="16">
              <n-gi>
                <n-card title="分红策略" :bordered="true">
                  <n-space vertical size="large">
                    <n-card title="什么是分红策略？" size="small" :bordered="true">
                      <n-text>
                        分红策略是指通过购买高分红股票（如银行股、公用事业股等），长期持有并不断买入，通过每年获得的分红收益来实现长期盈利的投资策略。这是一种相对稳健的长期投资方式。
                      </n-text>
                    </n-card>
                    
                    <n-card title="分红股票的特点" size="small" :bordered="true">
                      <n-space vertical>
                        <n-card title="1. 银行股" size="small" :bordered="true">
                          <n-text>
                            <strong>特点：</strong>
                            <br/>• 分红率通常较高（3%-6%）
                            <br/>• 盈利稳定，现金流充足
                            <br/>• 股价波动相对较小
                            <br/>• 适合长期持有
                            <br/>
                            <br/><strong>代表股票：</strong>工商银行、建设银行、农业银行、中国银行等
                          </n-text>
                        </n-card>
                        
                        <n-card title="2. 公用事业股" size="small" :bordered="true">
                          <n-text>
                            <strong>特点：</strong>
                            <br/>• 业务稳定，需求刚性
                            <br/>• 分红率较高且稳定
                            <br/>• 受经济周期影响较小
                            <br/>
                            <br/><strong>代表股票：</strong>电力、水务、燃气等公用事业公司
                          </n-text>
                        </n-card>
                        
                        <n-card title="3. 高分红蓝筹股" size="small" :bordered="true">
                          <n-text>
                            <strong>特点：</strong>
                            <br/>• 行业龙头，盈利能力强
                            <br/>• 分红政策稳定
                            <br/>• 适合价值投资者
                            <br/>
                            <br/><strong>代表股票：</strong>大型央企、国企等
                          </n-text>
                        </n-card>
                      </n-space>
                    </n-card>
                    
                    <n-card title="分红策略的盈利方式" size="small" :bordered="true">
                      <n-space vertical>
                        <n-card title="1. 分红收益" size="small" :bordered="true">
                          <n-text>
                            <strong>直接收益：</strong>
                            <br/>• 每年获得现金分红
                            <br/>• 分红率 = 每股分红 / 股价 × 100%
                            <br/>• 例如：股价10元，每股分红0.5元，分红率 = 5%
                            <br/>
                            <br/><strong>复利效应：</strong>
                            <br/>• 将分红再投资买入更多股票
                            <br/>• 下一年获得更多分红
                            <br/>• 长期复利增长效果显著
                          </n-text>
                        </n-card>
                        
                        <n-card title="2. 股价上涨收益" size="small" :bordered="true">
                          <n-text>
                            <strong>长期价值回归：</strong>
                            <br/>• 高分红股票通常估值较低
                            <br/>• 随着业绩增长，股价可能上涨
                            <br/>• 获得"分红+价差"双重收益
                          </n-text>
                        </n-card>
                        
                        <n-card title="3. 成本摊薄" size="small" :bordered="true">
                          <n-text>
                            <strong>持续买入策略：</strong>
                            <br/>• 定期买入，降低平均成本
                            <br/>• 分红再投资，增加持股数量
                            <br/>• 长期持有，享受复利增长
                          </n-text>
                        </n-card>
                      </n-space>
                    </n-card>
                    
                    <n-card title="实际分红盈利案例（银行股）" size="small" :bordered="true">
                      <n-space vertical size="medium">
                        <n-alert type="info" title="案例背景">
                          假设您选择某大型银行股进行长期分红投资，通过持续买入和分红再投资，实现长期稳定盈利。
                        </n-alert>
                        
                        <n-card title="📊 投资时间线（5年计划）" size="small" :bordered="true">
                          <n-text>
                            <strong>初始投资：</strong>
                            <br/>时间：2024年1月
                            <br/>股票：某大型银行股
                            <br/>买入价格：5.00元/股
                            <br/>买入数量：10,000股
                            <br/>投入资金：50,000元
                            <br/>
                            <br/><strong>第1年（2024年）：</strong>
                            <br/>• 年中分红：每股0.25元，共2,500元
                            <br/>• 分红再投资：2,500元 ÷ 5.20元 = 480股（假设股价5.20元）
                            <br/>• 累计持股：10,480股
                            <br/>• 累计投入：50,000元
                            <br/>
                            <br/><strong>第2年（2025年）：</strong>
                            <br/>• 年中分红：每股0.26元，共2,725元（10,480股 × 0.26元）
                            <br/>• 额外买入：每月定投2,000元，全年24,000元
                            <br/>• 分红再投资：2,725元 ÷ 5.30元 = 514股
                            <br/>• 定投买入：24,000元 ÷ 5.30元 = 4,528股
                            <br/>• 累计持股：15,522股（10,480 + 514 + 4,528）
                            <br/>• 累计投入：74,000元（50,000 + 24,000）
                            <br/>
                            <br/><strong>第3年（2026年）：</strong>
                            <br/>• 年中分红：每股0.27元，共4,191元（15,522股 × 0.27元）
                            <br/>• 额外买入：每月定投2,000元，全年24,000元
                            <br/>• 分红再投资：4,191元 ÷ 5.50元 = 761股
                            <br/>• 定投买入：24,000元 ÷ 5.50元 = 4,364股
                            <br/>• 累计持股：20,647股
                            <br/>• 累计投入：98,000元
                            <br/>
                            <br/><strong>第4年（2027年）：</strong>
                            <br/>• 年中分红：每股0.28元，共5,781元
                            <br/>• 额外买入：每月定投2,000元，全年24,000元
                            <br/>• 分红再投资：5,781元 ÷ 5.70元 = 1,014股
                            <br/>• 定投买入：24,000元 ÷ 5.70元 = 4,211股
                            <br/>• 累计持股：25,872股
                            <br/>• 累计投入：122,000元
                            <br/>
                            <br/><strong>第5年（2028年）：</strong>
                            <br/>• 年中分红：每股0.30元，共7,762元
                            <br/>• 额外买入：每月定投2,000元，全年24,000元
                            <br/>• 分红再投资：7,762元 ÷ 6.00元 = 1,294股
                            <br/>• 定投买入：24,000元 ÷ 6.00元 = 4,000股
                            <br/>• 累计持股：31,166股
                            <br/>• 累计投入：146,000元
                          </n-text>
                        </n-card>
                        
                        <n-card title="💰 5年后的盈利分析" size="small" :bordered="true">
                          <n-text>
                            <strong>假设第5年末股价为6.50元：</strong>
                            <br/>
                            <br/><strong>持仓情况：</strong>
                            <br/>• 总持股数量：31,166股
                            <br/>• 总投入资金：146,000元
                            <br/>• 平均成本：146,000 ÷ 31,166 = 4.68元/股
                            <br/>
                            <br/><strong>资产价值：</strong>
                            <br/>• 当前市值：31,166股 × 6.50元 = 202,579元
                            <br/>• 累计分红收入：2,500 + 2,725 + 4,191 + 5,781 + 7,762 = 22,959元
                            <br/>
                            <br/><strong>总收益：</strong>
                            <br/>• 价差收益：202,579 - 146,000 = 56,579元
                            <br/>• 分红收益：22,959元
                            <br/>• 总收益：56,579 + 22,959 = 79,538元
                            <br/>• 总收益率：(79,538 / 146,000) × 100% = <strong style="color: #18a058;">54.48%</strong>
                            <br/>
                            <br/><strong>年化收益率：</strong>
                            <br/>• 5年总收益率：54.48%
                            <br/>• 年化收益率：约 <strong style="color: #18a058;">9.1%</strong>（复利计算）
                            <br/>
                            <br/><strong>关键优势：</strong>
                            <br/>✓ 分红收益稳定，每年都有现金流入
                            <br/>✓ 分红再投资，享受复利增长
                            <br/>✓ 持续定投，降低平均成本
                            <br/>✓ 长期持有，获得价差收益
                            <br/>✓ 风险相对较低，适合稳健投资者
                          </n-text>
                        </n-card>
                        
                        <n-card title="📈 不同分红率下的收益对比" size="small" :bordered="true">
                          <n-text>
                            <table style="width: 100%; border-collapse: collapse; margin-top: 8px;">
                              <thead>
                                <tr style="background-color: var(--n-color-hover);">
                                  <th style="padding: 8px; border: 1px solid var(--n-border-color); text-align: left;">分红率</th>
                                  <th style="padding: 8px; border: 1px solid var(--n-border-color); text-align: right;">年分红收益（10,000股）</th>
                                  <th style="padding: 8px; border: 1px solid var(--n-border-color); text-align: right;">5年累计分红</th>
                                  <th style="padding: 8px; border: 1px solid var(--n-border-color); text-align: right;">复利后总收益</th>
                                </tr>
                              </thead>
                              <tbody>
                                <tr>
                                  <td style="padding: 8px; border: 1px solid var(--n-border-color);">3%</td>
                                  <td style="padding: 8px; border: 1px solid var(--n-border-color); text-align: right;">1,500元</td>
                                  <td style="padding: 8px; border: 1px solid var(--n-border-color); text-align: right;">7,500元</td>
                                  <td style="padding: 8px; border: 1px solid var(--n-border-color); text-align: right; color: #18a058;">约8,200元</td>
                                </tr>
                                <tr>
                                  <td style="padding: 8px; border: 1px solid var(--n-border-color);">4%</td>
                                  <td style="padding: 8px; border: 1px solid var(--n-border-color); text-align: right;">2,000元</td>
                                  <td style="padding: 8px; border: 1px solid var(--n-border-color); text-align: right;">10,000元</td>
                                  <td style="padding: 8px; border: 1px solid var(--n-border-color); text-align: right; color: #18a058;">约11,200元</td>
                                </tr>
                                <tr>
                                  <td style="padding: 8px; border: 1px solid var(--n-border-color);">5%</td>
                                  <td style="padding: 8px; border: 1px solid var(--n-border-color); text-align: right;">2,500元</td>
                                  <td style="padding: 8px; border: 1px solid var(--n-border-color); text-align: right;">12,500元</td>
                                  <td style="padding: 8px; border: 1px solid var(--n-border-color); text-align: right; color: #18a058;">约14,500元</td>
                                </tr>
                                <tr>
                                  <td style="padding: 8px; border: 1px solid var(--n-border-color);">6%</td>
                                  <td style="padding: 8px; border: 1px solid var(--n-border-color); text-align: right;">3,000元</td>
                                  <td style="padding: 8px; border: 1px solid var(--n-border-color); text-align: right;">15,000元</td>
                                  <td style="padding: 8px; border: 1px solid var(--n-border-color); text-align: right; color: #18a058;">约17,800元</td>
                                </tr>
                              </tbody>
                            </table>
                            <br/>
                            <n-text depth="3">
                              * 假设初始投入50,000元买入10,000股（5元/股），分红再投资，不考虑价差收益
                            </n-text>
                          </n-text>
                        </n-card>
                        
                        <n-card title="💡 分红策略的关键要点" size="small" :bordered="true">
                          <n-space vertical>
                            <n-card title="1. 选择优质分红股票" size="small" :bordered="true">
                              <n-text>
                                <strong>筛选标准：</strong>
                                <br/>• 分红率稳定在3%以上
                                <br/>• 连续多年稳定分红
                                <br/>• 公司盈利稳定，现金流充足
                                <br/>• 行业地位稳固，竞争优势明显
                                <br/>• 估值合理，PE、PB处于合理区间
                              </n-text>
                            </n-card>
                            
                            <n-card title="2. 持续买入策略" size="small" :bordered="true">
                              <n-text>
                                <strong>定投方式：</strong>
                                <br/>• 每月固定金额买入
                                <br/>• 股价低时多买，高时少买
                                <br/>• 长期坚持，不要中断
                                <br/>• 分红后立即再投资
                              </n-text>
                            </n-card>
                            
                            <n-card title="3. 长期持有" size="small" :bordered="true">
                              <n-text>
                                <strong>持有周期：</strong>
                                <br/>• 分红策略适合长期持有（5年以上）
                                <br/>• 不要频繁买卖，减少交易成本
                                <br/>• 享受复利增长效应
                                <br/>• 忽略短期股价波动
                              </n-text>
                            </n-card>
                            
                            <n-card title="4. 风险控制" size="small" :bordered="true">
                              <n-text>
                                <strong>注意事项：</strong>
                                <br/>• 分散投资，不要只买一只股票
                                <br/>• 关注公司基本面变化
                                <br/>• 如果公司停止分红或基本面恶化，及时调整
                                <br/>• 不要因为短期股价下跌而恐慌卖出
                              </n-text>
                            </n-card>
                          </n-space>
                        </n-card>
                        
                        <n-card title="⚠️ 分红策略的局限性" size="small" :bordered="true">
                          <n-alert type="warning" style="margin-bottom: 12px;">
                            <ul>
                              <li><strong>收益相对较慢：</strong>分红策略是稳健型投资，收益增长相对缓慢，不适合追求快速收益的投资者</li>
                              <li><strong>需要长期资金：</strong>需要能够长期持有的资金，不适合短期资金</li>
                              <li><strong>股价可能长期不涨：</strong>高分红股票可能长期横盘，主要依靠分红收益</li>
                              <li><strong>分红可能减少：</strong>公司经营困难时可能减少或停止分红</li>
                              <li><strong>需要缴税：</strong>分红收益需要缴纳个人所得税（持股1年以上免税）</li>
                            </ul>
                          </n-alert>
                        </n-card>
                      </n-space>
                    </n-card>
                  </n-space>
                </n-card>
              </n-gi>
            </n-grid>
          </template>

          <!-- 其他主题内容（占位） -->
          <template v-else>
            <n-card :title="activeTopic" :bordered="true">
              <n-text>{{ activeTopic }}知识内容待完善...</n-text>
            </n-card>
          </template>
        </n-scrollbar>
      </n-card>
    </n-gi>
  </n-grid>
</template>

<style scoped>
.active-topic {
  font-weight: bold;
}
</style>
