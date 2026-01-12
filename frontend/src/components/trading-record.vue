<script setup>
import {computed, h, onBeforeMount, onMounted, reactive, ref} from "vue";
import {
  CreateTradingRecord,
  DeleteTradingRecord,
  GetLimitUpDownSectors,
  GetTradingRecord,
  GetTradingRecordByDate,
  GetTradingRecordList,
  UpdateTradingRecord
} from "../../wailsjs/go/main/App";
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
  NTag,
  NText,
  useMessage
} from "naive-ui";
import {CalendarOutline, DocumentTextOutline, TrashOutline} from "@vicons/ionicons5";

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

onBeforeMount(() => {
  loadRecords();
  loadLimitUpDownSectors();
});

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
        <n-button type="primary" @click="newRecord">
          <template #icon>
            <n-icon><DocumentTextOutline /></n-icon>
          </template>
          新建复盘
        </n-button>
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
  </n-scrollbar>
</template>

<style scoped>
</style>

