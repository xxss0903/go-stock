<script setup>
import { computed, nextTick, reactive, ref } from 'vue'
import {
  NAlert,
  NButton,
  NCard,
  NEmpty,
  NFormItem,
  NGrid,
  NGi,
  NIcon,
  NInputNumber,
  NList,
  NListItem,
  NScrollbar,
  NSpace,
  NStatistic,
  NText,
  useMessage,
} from 'naive-ui'
import { BookOutline, CalculatorOutline } from '@vicons/ionicons5'

const message = useMessage()

// 当前选中的学习主题
const activeTopic = ref('补仓')

// 学习主题列表
const topics = [
  { key: '补仓', label: '补仓策略', icon: CalculatorOutline },
  { key: 'K线', label: 'K线基础', icon: BookOutline },
  { key: '技术指标', label: '技术指标', icon: BookOutline },
  { key: '基本面', label: '基本面分析', icon: BookOutline },
  { key: '风险管理', label: '风险管理', icon: BookOutline },
]

// 补仓计算器相关状态
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
              <!-- 补仓成本计算器 -->
              <n-gi>
                <n-card title="补仓成本计算器" :bordered="true">
                  <n-space vertical size="large">
                    <!-- 初始持仓 -->
                    <n-card title="初始持仓" size="small" :bordered="true">
                      <n-grid :cols="3" :x-gap="12">
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
                        <n-card v-for="(add, index) in costCalculator.additions" :key="add.id" size="small" :bordered="true" style="margin-bottom: 8px;">
                          <n-grid :cols="4" :x-gap="12">
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
                      <n-form-item label="当前股价（元）">
                        <n-input-number v-model:value="costCalculator.currentPrice" :min="0.01" :step="0.01" :precision="2" style="width: 200px"/>
                      </n-form-item>
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
