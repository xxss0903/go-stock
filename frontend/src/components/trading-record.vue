<script setup>
import {computed, h, onBeforeMount, onBeforeUnmount, onMounted, reactive, ref} from "vue";
import {
  CreateTradingRecord,
  DeleteTradingRecord,
  GetLimitUpDownSectors,
  GetTradingRecord,
  GetTradingRecordByDate,
  GetTradingRecordList,
  UpdateTradingRecord,
  SummaryStockNews,
  GetAIResponseResult,
  SaveAIResponseResult,
  SaveAsMarkdown,
  ShareAnalysis,
  GetAiConfigs,
  GetPromptTemplates,
  GetLimitListFromTushare
} from "../../wailsjs/go/main/App";
import {EventsOn, EventsOff} from "../../wailsjs/runtime";
import {
  NButton,
  NCard,
  NDataTable,
  NDatePicker,
  NForm,
  NFormItem,
  NGrid,
  NGridItem,
  NInput,
  NInputNumber,
  NModal,
  NScrollbar,
  NSelect,
  NSwitch,
  NTag,
  NText,
  NFlex,
  NSpin,
  useMessage
} from "naive-ui";
import {CalendarOutline, DocumentTextOutline, TrashOutline, PulseOutline} from "@vicons/ionicons5";
import {MdPreview} from "md-editor-v3";

const message = useMessage();
const loading = ref(false);
const showModal = ref(false);
const editingRecord = ref(null);
const tradeDate = ref(new Date());
const limitUpSectors = ref([]);
const limitDownSectors = ref([]);
const records = ref([]);
const total = ref(0);
const page = ref(1);
const pageSize = ref(10);

// 涨跌停股票列表
const limitUpStocks = ref([]);
const limitDownStocks = ref([]);
const limitUpStocksLoading = ref(false);
const limitDownStocksLoading = ref(false);
const showLimitStocksModal = ref(false);
const limitStocksType = ref("U"); // U涨停 D跌停 Z炸板

// AI总结相关状态
const aiSummaryModal = ref(false);
const aiSummary = ref("");
const aiSummaryTime = ref("");
const aiSummaryModelName = ref("");
const aiSummaryChatId = ref("");
const aiSummaryQuestion = ref("总结和分析今日股票市场的表现，包括涨停跌停板块情况、市场热点、资金流向等，并给出投资建议");
const aiSummaryLoading = ref(false);
const aiConfigId = ref(null);
const sysPromptId = ref(null);
const enableTools = ref(true);
const thinkingMode = ref(true);
const aiConfigs = ref([]);
const sysPromptOptions = ref([]);
const userPromptOptions = ref([]);
const promptTemplates = ref([]);
const theme = ref("light");

// 计算属性：只显示涨停数大于3的板块，并按涨停数降序排序
const filteredLimitUpSectors = computed(() => {
  return limitUpSectors.value
    .filter(sector => sector.stockCount > 3)
    .sort((a, b) => b.stockCount - a.stockCount);
});

// 计算属性：只显示跌停数大于3的板块，并按跌停数降序排序
const filteredLimitDownSectors = computed(() => {
  return limitDownSectors.value
    .filter(sector => sector.stockCount > 3)
    .sort((a, b) => b.stockCount - a.stockCount);
});

const formData = reactive({
  tradeDate: "",
  summary: "",
  review: "",
  limitUpSectors: "",
  limitDownSectors: ""
});

const columns = [
  {
    title: "日期",
    key: "tradeDate",
    width: 120
  },
  {
    title: "总结",
    key: "summary",
    ellipsis: {
      tooltip: true
    }
  },
  {
    title: "操作",
    key: "actions",
    width: 150,
    render(row) {
      return h("div", {style: "display: flex; gap: 8px;"}, [
        h(NButton, {
          size: "small",
          onClick: () => editRecord(row)
        }, {default: () => "编辑"}),
        h(NButton, {
          size: "small",
          type: "error",
          onClick: () => deleteRecord(row.id)
        }, {default: () => "删除"})
      ]);
    }
  }
];

const limitStocksColumns = [
  {
    title: "股票代码",
    key: "tsCode",
    width: 120
  },
  {
    title: "股票名称",
    key: "name",
    width: 100
  },
  {
    title: "所属行业",
    key: "industry",
    width: 120
  },
  {
    title: "收盘价",
    key: "close",
    width: 100,
    render(row) {
      return h("span", row.close.toFixed(2));
    }
  },
  {
    title: "涨跌幅",
    key: "pctChg",
    width: 100,
    render(row) {
      const color = row.pctChg >= 0 ? "error" : "info";
      return h(NTag, {type: color, size: "small"}, {default: () => `${row.pctChg.toFixed(2)}%`});
    }
  },
  {
    title: "成交额(万)",
    key: "amount",
    width: 120,
    render(row) {
      return h("span", (row.amount / 10000).toFixed(2));
    }
  },
  {
    title: "换手率",
    key: "turnoverRatio",
    width: 100,
    render(row) {
      return h("span", `${row.turnoverRatio.toFixed(2)}%`);
    }
  },
  {
    title: "封单金额(万)",
    key: "fdAmount",
    width: 120,
    render(row) {
      return h("span", (row.fdAmount / 10000).toFixed(2));
    }
  },
  {
    title: "连板数",
    key: "limitTimes",
    width: 80
  },
  {
    title: "涨停统计",
    key: "upStat",
    width: 100
  }
];

onBeforeMount(() => {
  loadRecords();
  loadLimitUpDownSectors();
  initAIConfig();
  // 监听AI总结事件
  EventsOn("summaryStockNews", handleAISummaryEvent);
});

onBeforeUnmount(() => {
  EventsOff("summaryStockNews");
});

function initAIConfig() {
  GetAiConfigs().then(result => {
    if (result && result.length > 0) {
      aiConfigs.value = result;
      aiConfigId.value = result[0].ID;
    }
  });
  
  GetPromptTemplates().then(result => {
    if (result) {
      promptTemplates.value = result;
      sysPromptOptions.value = result.filter(t => t.type === "system");
      userPromptOptions.value = result.filter(t => t.type === "user");
      if (sysPromptOptions.value.length > 0) {
        sysPromptId.value = sysPromptOptions.value[0].ID;
      }
    }
  });
}

function handleAISummaryEvent(msg) {
  aiSummaryLoading.value = false;
  if (msg === "DONE") {
    SaveAIResponseResult("炒股复盘", "炒股复盘", aiSummary.value, aiSummaryChatId.value, aiSummaryQuestion.value, aiConfigId.value).then(() => {
      message.info("AI分析完成！");
      message.destroyAll();
    });
  } else {
    if (msg.code === 0 && msg.content) {
      message.error("AI分析出错，请查看下方错误信息");
      aiSummary.value = aiSummary.value + "\n\n" + msg.content;
      return;
    }
    if (msg.chatId) {
      aiSummaryChatId.value = msg.chatId;
    }
    if (msg.question) {
      aiSummaryQuestion.value = msg.question;
    }
    if (msg.content) {
      aiSummary.value = aiSummary.value + msg.content;
    }
    if (msg.extraContent) {
      aiSummary.value = aiSummary.value + msg.extraContent;
    }
    if (msg.model) {
      aiSummaryModelName.value = msg.model;
    }
    if (msg.time) {
      aiSummaryTime.value = msg.time;
    }
  }
}

function getAISummary() {
  aiSummaryModal.value = true;
  aiSummaryLoading.value = true;
  const dateStr = tradeDate.value instanceof Date ? tradeDate.value.toISOString().split('T')[0] : tradeDate.value;
  GetAIResponseResult("炒股复盘-" + dateStr).then(result => {
    aiSummaryLoading.value = false;
    if (result.content) {
      aiSummary.value = result.content;
      aiSummaryQuestion.value = result.question;
      const date = new Date(result.CreatedAt);
      const year = date.getFullYear();
      const month = String(date.getMonth() + 1).padStart(2, '0');
      const day = String(date.getDate()).padStart(2, '0');
      const hours = String(date.getHours()).padStart(2, '0');
      const minutes = String(date.getMinutes()).padStart(2, '0');
      const seconds = String(date.getSeconds()).padStart(2, '0');
      aiSummaryTime.value = `${year}-${month}-${day} ${hours}:${minutes}:${seconds}`;
      aiSummaryModelName.value = result.modelName;
    } else {
      aiSummaryTime.value = "";
      aiSummary.value = "";
      aiSummaryModelName.value = "";
    }
  });
}

function reAISummary() {
  aiSummary.value = "";
  aiSummaryModal.value = true;
  aiSummaryLoading.value = true;
  const dateStr = tradeDate.value instanceof Date ? tradeDate.value.toISOString().split('T')[0] : tradeDate.value;
  const prompt = `日期：${dateStr}\n${aiSummaryQuestion.value}\n\n请结合当日的涨停跌停板块情况进行分析。`;
  SummaryStockNews(prompt, aiConfigId.value, sysPromptId.value, enableTools.value, thinkingMode.value);
}

async function copyAISummaryToClipboard() {
  try {
    await navigator.clipboard.writeText(aiSummary.value);
    message.success('分析结果已复制到剪切板');
  } catch (err) {
    message.error('复制失败: ' + err);
  }
}

function saveAISummaryAsMarkdown() {
  const dateStr = tradeDate.value instanceof Date ? tradeDate.value.toISOString().split('T')[0] : tradeDate.value;
  SaveAsMarkdown('炒股复盘-' + dateStr, '炒股复盘').then(result => {
    message.success(result);
  });
}

function shareAISummary() {
  const dateStr = tradeDate.value instanceof Date ? tradeDate.value.toISOString().split('T')[0] : tradeDate.value;
  ShareAnalysis('炒股复盘-' + dateStr, '炒股复盘', aiSummary.value).then(result => {
    message.success(result);
  });
}

function loadRecords() {
  loading.value = true;
  GetTradingRecordList(page.value, pageSize.value).then(result => {
    if (result.error) {
      message.error("加载记录失败: " + result.error);
    } else {
      records.value = result.records || [];
      total.value = result.total || 0;
    }
    loading.value = false;
  });
}

function loadLimitUpDownSectors() {
  if (!tradeDate.value) return;
  const date = tradeDate.value instanceof Date ? tradeDate.value : new Date(tradeDate.value);
  const dateStr = date.toISOString().split('T')[0];
  GetLimitUpDownSectors(dateStr).then(result => {
    if (result.error) {
      message.warning("获取涨停跌停板块信息失败: " + result.error);
    } else {
      limitUpSectors.value = result.limitUpSectors || [];
      limitDownSectors.value = result.limitDownSectors || [];
    }
  });
  
  // 同时加载Tushare的涨跌停股票列表
  loadLimitStocksFromTushare(dateStr);
}

function loadLimitStocksFromTushare(dateStr) {
  if (!dateStr) {
    if (!tradeDate.value) return;
    const date = tradeDate.value instanceof Date ? tradeDate.value : new Date(tradeDate.value);
    dateStr = date.toISOString().split('T')[0];
  }
  
  // 加载涨停股票
  limitUpStocksLoading.value = true;
  GetLimitListFromTushare(dateStr, "U").then(result => {
    limitUpStocksLoading.value = false;
    if (result.error) {
      message.warning("获取涨停股票列表失败: " + result.error);
    } else {
      limitUpStocks.value = result.stocks || [];
    }
    
    // 延迟1.5秒后加载跌停股票，避免API限流
    setTimeout(() => {
      limitDownStocksLoading.value = true;
      GetLimitListFromTushare(dateStr, "D").then(result => {
        limitDownStocksLoading.value = false;
        if (result.error) {
          message.warning("获取跌停股票列表失败: " + result.error);
        } else {
          limitDownStocks.value = result.stocks || [];
        }
      });
    }, 1500);
  });
}

function showLimitStocks(type) {
  limitStocksType.value = type;
  showLimitStocksModal.value = true;
  const date = tradeDate.value instanceof Date ? tradeDate.value : new Date(tradeDate.value);
  const dateStr = date.toISOString().split('T')[0];
  loadLimitStocksFromTushare(dateStr);
}

function editRecord(record) {
  editingRecord.value = record;
  formData.tradeDate = record.tradeDate;
  formData.summary = record.summary || "";
  formData.review = record.review || "";
  formData.limitUpSectors = record.limitUpSectors || "";
  formData.limitDownSectors = record.limitDownSectors || "";
  showModal.value = true;
}

function newRecord() {
  editingRecord.value = null;
  if (!tradeDate.value) {
    tradeDate.value = new Date();
  }
  const date = tradeDate.value instanceof Date ? tradeDate.value : new Date(tradeDate.value);
  const dateStr = date.toISOString().split('T')[0];
  formData.tradeDate = dateStr;
  formData.summary = "";
  formData.review = "";
  formData.limitUpSectors = "";
  formData.limitDownSectors = "";
  
  // 检查是否已有当天的记录
  GetTradingRecordByDate(dateStr).then(record => {
    if (record) {
      editRecord(record);
    } else {
      showModal.value = true;
      loadLimitUpDownSectors();
    }
  });
}

function saveRecord() {
  if (!formData.tradeDate) {
    message.error("请选择日期");
    return;
  }

  // 将涨停跌停板块信息转换为JSON
  const limitUpJson = JSON.stringify(limitUpSectors.value);
  const limitDownJson = JSON.stringify(limitDownSectors.value);

  const record = {
    id: editingRecord.value?.id || 0,
    tradeDate: formData.tradeDate,
    summary: formData.summary,
    review: formData.review,
    limitUpSectors: limitUpJson,
    limitDownSectors: limitDownJson
  };

  loading.value = true;
  const promise = editingRecord.value
    ? UpdateTradingRecord(record)
    : CreateTradingRecord(record);

  promise.then(result => {
    if (result.includes("失败")) {
      message.error(result);
    } else {
      message.success(result);
      showModal.value = false;
      loadRecords();
    }
    loading.value = false;
  });
}

function deleteRecord(id) {
  if (!confirm("确定要删除这条记录吗？")) {
    return;
  }
  loading.value = true;
  DeleteTradingRecord(id).then(result => {
    if (result.includes("失败")) {
      message.error(result);
    } else {
      message.success(result);
      loadRecords();
    }
    loading.value = false;
  });
}

function handleDateChange(value) {
  tradeDate.value = value;
  if (!value) return;
  const date = value instanceof Date ? value : new Date(value);
  const dateStr = date.toISOString().split('T')[0];
  loadLimitUpDownSectors();
  GetTradingRecordByDate(dateStr).then(record => {
    if (record) {
      formData.tradeDate = record.tradeDate;
      formData.summary = record.summary || "";
      formData.review = record.review || "";
      formData.limitUpSectors = record.limitUpSectors || "";
      formData.limitDownSectors = record.limitDownSectors || "";
      
      // 解析涨停跌停板块信息
      if (record.limitUpSectors) {
        try {
          limitUpSectors.value = JSON.parse(record.limitUpSectors);
        } catch (e) {
          limitUpSectors.value = [];
        }
      } else {
        limitUpSectors.value = [];
      }
      if (record.limitDownSectors) {
        try {
          limitDownSectors.value = JSON.parse(record.limitDownSectors);
        } catch (e) {
          limitDownSectors.value = [];
        }
      } else {
        limitDownSectors.value = [];
      }
    } else {
      formData.tradeDate = dateStr;
      formData.summary = "";
      formData.review = "";
      limitUpSectors.value = [];
      limitDownSectors.value = [];
    }
  });
}
</script>

<template>
  <n-scrollbar style="height: calc(100vh - 120px);">
    <n-card title="炒股复盘记录" style="margin: 16px;">
      <template #header-extra>
        <n-flex gap="8">
          <n-button type="primary" @click="newRecord">
            <template #icon>
              <n-icon><DocumentTextOutline /></n-icon>
            </template>
            新建复盘
          </n-button>
          <n-button type="info" @click="getAISummary">
            <template #icon>
              <n-icon><PulseOutline /></n-icon>
            </template>
            AI复盘
          </n-button>
        </n-flex>
      </template>

      <n-form :model="formData" label-placement="left" label-width="80px" style="margin-bottom: 16px;">
        <n-form-item label="选择日期">
          <n-date-picker
            v-model:value="tradeDate"
            type="date"
            clearable
            @update:value="handleDateChange"
            style="width: 200px;"
          />
        </n-form-item>
      </n-form>

      <!-- 涨停跌停板块信息 - 左右分栏布局 -->
      <n-grid :cols="2" :x-gap="16" style="margin-bottom: 16px;">
        <!-- 左侧：涨停板块 -->
        <n-grid-item>
          <n-card v-if="filteredLimitUpSectors.length > 0" title="涨停板块（涨停数>3）" size="small">
            <template #header-extra>
              <n-button size="small" type="error" @click="showLimitStocks('U')">查看涨停股票</n-button>
            </template>
            <n-grid :cols="1" :y-gap="8">
              <n-grid-item v-for="sector in filteredLimitUpSectors" :key="sector.sectorName">
                <n-card size="small" hoverable style="cursor: pointer;">
                  <template #header>
                    <n-tag type="error" size="small">{{ sector.sectorName }}</n-tag>
                  </template>
                  <n-text strong type="error" style="font-size: 16px;">{{ sector.stockCount }}只涨停</n-text>
                  <div style="margin-top: 8px;">
                    <n-tag
                      v-for="(stock, index) in sector.stocks.slice(0, 5)"
                      :key="stock"
                      size="small"
                      style="margin-right: 4px; margin-bottom: 4px; display: inline-block;"
                    >
                      {{ stock }}
                    </n-tag>
                    <n-text v-if="sector.stocks.length > 5" depth="3" style="font-size: 12px;">
                      等{{ sector.stocks.length }}只
                    </n-text>
                  </div>
                </n-card>
              </n-grid-item>
            </n-grid>
          </n-card>
          <n-card v-else title="涨停板块（涨停数>3）" size="small">
            <n-text depth="3">暂无涨停板块数据</n-text>
          </n-card>
        </n-grid-item>

        <!-- 右侧：跌停板块 -->
        <n-grid-item>
          <n-card v-if="filteredLimitDownSectors.length > 0" title="跌停板块（跌停数>3）" size="small">
            <template #header-extra>
              <n-button size="small" type="info" @click="showLimitStocks('D')">查看跌停股票</n-button>
            </template>
            <n-grid :cols="1" :y-gap="8">
              <n-grid-item v-for="sector in filteredLimitDownSectors" :key="sector.sectorName">
                <n-card size="small" hoverable style="cursor: pointer;">
                  <template #header>
                    <n-tag type="info" size="small">{{ sector.sectorName }}</n-tag>
                  </template>
                  <n-text strong type="info" style="font-size: 16px;">{{ sector.stockCount }}只跌停</n-text>
                  <div style="margin-top: 8px;">
                    <n-tag
                      v-for="(stock, index) in sector.stocks.slice(0, 5)"
                      :key="stock"
                      size="small"
                      type="info"
                      style="margin-right: 4px; margin-bottom: 4px; display: inline-block;"
                    >
                      {{ stock }}
                    </n-tag>
                    <n-text v-if="sector.stocks.length > 5" depth="3" style="font-size: 12px;">
                      等{{ sector.stocks.length }}只
                    </n-text>
                  </div>
                </n-card>
              </n-grid-item>
            </n-grid>
          </n-card>
          <n-card v-else title="跌停板块（跌停数>3）" size="small">
            <n-text depth="3">暂无跌停板块数据</n-text>
          </n-card>
        </n-grid-item>
      </n-grid>

      <!-- 记录列表 -->
      <n-data-table
        :columns="columns"
        :data="records"
        :loading="loading"
        :pagination="{
          page: page,
          pageSize: pageSize,
          pageCount: Math.ceil(total / pageSize),
          showSizePicker: true,
          pageSizes: [10, 20, 50],
          onChange: (p) => { page = p; loadRecords(); },
          onUpdatePageSize: (s) => { pageSize = s; page = 1; loadRecords(); }
        }"
      />
    </n-card>

    <!-- 编辑/新建模态框 -->
    <n-modal
      v-model:show="showModal"
      preset="dialog"
      title="复盘记录"
      positive-text="保存"
      negative-text="取消"
      @positive-click="saveRecord"
      style="width: 800px;"
    >
      <n-form :model="formData" label-placement="left" label-width="80px">
        <n-form-item label="日期">
          <n-date-picker
            v-model:value="tradeDate"
            type="date"
            @update:value="(v) => { 
              if (v) {
                const date = v instanceof Date ? v : new Date(v);
                formData.tradeDate = date.toISOString().split('T')[0];
              }
            }"
            style="width: 100%;"
          />
        </n-form-item>
        <n-form-item label="当日总结">
          <n-input
            v-model:value="formData.summary"
            type="textarea"
            :rows="3"
            placeholder="简要总结当日的交易情况和市场表现"
          />
        </n-form-item>
        <n-form-item label="复盘内容">
          <n-input
            v-model:value="formData.review"
            type="textarea"
            :rows="10"
            placeholder="详细复盘内容，包括：\n1. 市场整体表现\n2. 关注的股票表现\n3. 操作回顾\n4. 经验教训\n5. 明日计划等"
          />
        </n-form-item>
        <n-form-item label="涨停板块">
          <n-scrollbar style="max-height: 200px;">
            <n-grid v-if="filteredLimitUpSectors.length > 0" :cols="2" :x-gap="8" :y-gap="8">
              <n-grid-item v-for="sector in filteredLimitUpSectors" :key="sector.sectorName">
                <n-card size="small" style="padding: 8px;">
                  <n-tag type="error" size="small">{{ sector.sectorName }}</n-tag>
                  <n-text strong type="error" style="margin-left: 8px;">{{ sector.stockCount }}只</n-text>
                </n-card>
              </n-grid-item>
            </n-grid>
            <n-text v-else depth="3">暂无涨停板块数据（涨停数>3）</n-text>
          </n-scrollbar>
        </n-form-item>
        <n-form-item label="跌停板块">
          <n-scrollbar style="max-height: 200px;">
            <n-grid v-if="filteredLimitDownSectors.length > 0" :cols="2" :x-gap="8" :y-gap="8">
              <n-grid-item v-for="sector in filteredLimitDownSectors" :key="sector.sectorName">
                <n-card size="small" style="padding: 8px;">
                  <n-tag type="info" size="small">{{ sector.sectorName }}</n-tag>
                  <n-text strong type="info" style="margin-left: 8px;">{{ sector.stockCount }}只</n-text>
                </n-card>
              </n-grid-item>
            </n-grid>
            <n-text v-else depth="3">暂无跌停板块数据（跌停数>3）</n-text>
          </n-scrollbar>
        </n-form-item>
      </n-form>
    </n-modal>

    <!-- 涨跌停股票列表模态框 -->
    <n-modal v-model:show="showLimitStocksModal" preset="dialog" :title="limitStocksType === 'U' ? '涨停股票列表' : '跌停股票列表'" style="width: 1000px;">
      <n-spin :show="limitStocksType === 'U' ? limitUpStocksLoading : limitDownStocksLoading">
        <n-data-table
          :columns="limitStocksColumns"
          :data="limitStocksType === 'U' ? limitUpStocks : limitDownStocks"
          :loading="limitStocksType === 'U' ? limitUpStocksLoading : limitDownStocksLoading"
          max-height="500px"
          scrollbar-props="{ trigger: 'none' }"
        />
      </n-spin>
    </n-modal>

    <!-- AI复盘模态框 -->
    <n-modal transform-origin="center" v-model:show="aiSummaryModal" preset="card" style="width: 800px;"
             :title="'AI复盘'">
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
          <n-text type="error" style="margin-left: 10px">*AI函数工具调用可以增强AI获取数据的能力,但会消耗更多tokens。</n-text>
        </n-flex>
        <n-flex justify="space-between" style="margin-bottom: 10px">
          <n-select style="width: 32%" v-model:value="aiConfigId" label-field="name" value-field="ID"
                    :options="aiConfigs" placeholder="请选择AI模型服务配置"/>
          <n-select style="width: 32%" v-model:value="sysPromptId" label-field="name" value-field="ID"
                    :options="sysPromptOptions" placeholder="请选择系统提示词"/>
          <n-select style="width: 32%" v-model:value="aiSummaryQuestion" label-field="name" value-field="content"
                    :options="userPromptOptions" placeholder="请选择用户提示词"/>
        </n-flex>
        <n-flex justify="right">
          <n-input v-model:value="aiSummaryQuestion" style="text-align: left" clearable
                   type="textarea"
                   :show-count="true"
                   placeholder="请输入您的问题:例如 总结和分析今日股票市场的表现，包括涨停跌停板块情况、市场热点、资金流向等，并给出投资建议"
                   :autosize="{
              minRows: 2,
              maxRows: 5
            }"
          />
          <n-button size="tiny" type="warning" @click="reAISummary">再次总结</n-button>
          <n-button size="tiny" type="success" @click="copyAISummaryToClipboard">复制到剪切板</n-button>
          <n-button size="tiny" type="primary" @click="saveAISummaryAsMarkdown">保存为Markdown文件</n-button>
          <n-button size="tiny" type="error" @click="shareAISummary">分享到项目社区</n-button>
        </n-flex>
      </template>
    </n-modal>

  </n-scrollbar>
</template>

<style scoped>
</style>

