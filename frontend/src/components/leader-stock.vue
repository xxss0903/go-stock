<script setup>
import {computed, onMounted, reactive, ref} from "vue";
import {
  NButton,
  NCard,
  NDatePicker,
  NFlex,
  NInputNumber,
  NScrollbar,
  NTag,
  NText,
  useMessage
} from "naive-ui";
import {GetLimitUpDownSectors} from "../../wailsjs/go/main/App";

const message = useMessage();
const leaderLoading = ref(false);
const leaderLimitUpSectors = ref([]);
const leaderSnapshots = ref([]);
const leaderMarks = ref({});

const leaderConfig = reactive({
  date: new Date(),
  sectorThreshold: 3,
  streakDays: 2
});

function formatDateForRecord(value) {
  if (!value) return "";
  const date = value instanceof Date ? value : new Date(value);
  if (isNaN(date.getTime())) return "";
  const year = date.getFullYear();
  const month = String(date.getMonth() + 1).padStart(2, "0");
  const day = String(date.getDate()).padStart(2, "0");
  return `${year}-${month}-${day}`;
}

function parseLocalDateFromYmd(value) {
  if (!value) return null;
  const parts = String(value).split("-");
  if (parts.length !== 3) return null;
  const year = Number(parts[0]);
  const month = Number(parts[1]);
  const day = Number(parts[2]);
  if (!year || !month || !day) return null;
  return new Date(year, month - 1, day);
}

function loadLeaderSnapshots() {
  try {
    const raw = localStorage.getItem("leaderSectorSnapshots");
    leaderSnapshots.value = raw ? JSON.parse(raw) : [];
  } catch (e) {
    leaderSnapshots.value = [];
  }
}

function saveLeaderSnapshots() {
  localStorage.setItem("leaderSectorSnapshots", JSON.stringify(leaderSnapshots.value));
}

function loadLeaderMarks() {
  try {
    const raw = localStorage.getItem("leaderMarks");
    leaderMarks.value = raw ? JSON.parse(raw) : {};
  } catch (e) {
    leaderMarks.value = {};
  }
}

function saveLeaderMarks() {
  localStorage.setItem("leaderMarks", JSON.stringify(leaderMarks.value));
}

function getLeaderDateStr() {
  return formatDateForRecord(leaderConfig.date);
}

function buildStockSet(sectors) {
  const set = new Set();
  sectors.forEach(sector => {
    (sector.stocks || []).forEach(stock => set.add(stock));
  });
  return set;
}

const leaderStreakMap = computed(() => {
  const snapshots = [...leaderSnapshots.value].filter(item => item && item.sectors);
  snapshots.sort((a, b) => (b.date || "").localeCompare(a.date || ""));
  if (snapshots.length === 0) return {};
  const latestSet = buildStockSet(snapshots[0].sectors);
  const streak = {};
  latestSet.forEach(stock => {
    streak[stock] = 1;
  });
  let active = new Set(latestSet);
  for (let i = 1; i < snapshots.length; i++) {
    if (active.size === 0) break;
    const currentSet = buildStockSet(snapshots[i].sectors);
    for (const stock of Array.from(active)) {
      if (currentSet.has(stock)) {
        streak[stock] += 1;
      } else {
        active.delete(stock);
      }
    }
  }
  return streak;
});

const leaderFilteredSectors = computed(() => {
  const threshold = leaderConfig.sectorThreshold || 0;
  const minStreak = leaderConfig.streakDays || 1;
  const streakMap = leaderStreakMap.value;
  const sectors = leaderLimitUpSectors.value || [];
  return sectors
    .filter(sector => (sector.stockCount || 0) >= threshold)
    .map(sector => {
      const stocks = (sector.stocks || []).map(name => ({
        name,
        streak: streakMap[name] || 0,
        marked: Boolean(leaderMarks.value[name])
      }));
      const displayStocks = stocks.filter(item => item.streak >= minStreak);
      return {
        ...sector,
        displayStocks
      };
    })
    .filter(sector => sector.displayStocks.length > 0);
});

function toggleLeaderMark(stockName) {
  if (!stockName) return;
  const next = {...leaderMarks.value};
  if (next[stockName]) {
    delete next[stockName];
  } else {
    next[stockName] = {markedAt: new Date().toISOString()};
  }
  leaderMarks.value = next;
  saveLeaderMarks();
}

function loadLeaderSectors() {
  const dateStr = getLeaderDateStr();
  if (!dateStr) {
    message.warning("请选择日期");
    return;
  }
  leaderLoading.value = true;
  GetLimitUpDownSectors(dateStr).then(result => {
    leaderLoading.value = false;
    if (result.error) {
      message.warning("获取涨停板块失败: " + result.error);
      return;
    }
    leaderLimitUpSectors.value = result.limitUpSectors || [];
    const nextSnapshots = [...leaderSnapshots.value];
    const index = nextSnapshots.findIndex(item => item.date === dateStr);
    const snapshot = {date: dateStr, sectors: leaderLimitUpSectors.value};
    if (index >= 0) {
      nextSnapshots[index] = snapshot;
    } else {
      nextSnapshots.push(snapshot);
    }
    leaderSnapshots.value = nextSnapshots;
    saveLeaderSnapshots();
    message.success("已采集涨停板块");
  });
}

onMounted(() => {
  loadLeaderSnapshots();
  loadLeaderMarks();
  if (leaderSnapshots.value.length > 0) {
    const sortedSnapshots = [...leaderSnapshots.value].sort((a, b) => (b.date || "").localeCompare(a.date || ""));
    leaderLimitUpSectors.value = sortedSnapshots[0].sectors || [];
    leaderConfig.date = parseLocalDateFromYmd(sortedSnapshots[0].date) || new Date();
  }
});
</script>

<template>
  <n-scrollbar style="height: calc(100vh - 120px);">
    <n-card title="龙头板块" style="margin: 16px;">
      <template #header-extra>
        <n-button size="small" type="primary" :loading="leaderLoading" @click="loadLeaderSectors">采集</n-button>
      </template>
      <n-flex wrap :style="{ gap: '8px', marginBottom: '12px' }">
        <n-date-picker v-model:value="leaderConfig.date" type="date" size="small" />
        <n-input-number v-model:value="leaderConfig.sectorThreshold" size="small" min="1" placeholder="板块阈值">
          <template #suffix>只</template>
        </n-input-number>
        <n-input-number v-model:value="leaderConfig.streakDays" size="small" min="1" placeholder="连板天数">
          <template #suffix>天</template>
        </n-input-number>
      </n-flex>

      <n-scrollbar style="max-height: calc(100vh - 260px);">
        <n-text v-if="leaderFilteredSectors.length === 0" depth="3">暂无符合条件的板块/股票</n-text>
        <div v-for="sector in leaderFilteredSectors" :key="sector.sectorName" style="margin-bottom: 12px;">
          <n-flex align="center">
            <n-tag size="small" type="error">{{ sector.sectorName }}</n-tag>
            <n-text depth="3" style="margin-left: 6px;">{{ sector.stockCount }}涨停</n-text>
          </n-flex>
          <n-flex wrap style="margin-top: 8px; gap: 6px;">
            <n-tag
              v-for="stock in sector.displayStocks"
              :key="stock.name"
              size="small"
              :type="stock.marked ? 'warning' : 'info'"
              :bordered="false"
              style="cursor: pointer;"
              @click="toggleLeaderMark(stock.name)"
            >
              {{ stock.name }} {{ stock.streak }}连
            </n-tag>
          </n-flex>
        </div>
      </n-scrollbar>
    </n-card>
  </n-scrollbar>
</template>

