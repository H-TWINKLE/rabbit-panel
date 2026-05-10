<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref, shallowRef, watch } from 'vue'
import { Box, Connection, Monitor, Picture } from '@element-plus/icons-vue'
import * as echarts from 'echarts'
import { useSystemStore } from '@/stores/system'
import { useContainerStore } from '@/stores/containers'
import { useI18n } from '@/composables/useI18n'
import UpdateBanner from '@/components/common/UpdateBanner.vue'
import UpdateDialog from '@/components/common/UpdateDialog.vue'

const { t } = useI18n()
const systemStore = useSystemStore()
const containerStore = useContainerStore()

const lineChartRef = ref<HTMLElement | null>(null)
const pieChartRef = ref<HTMLElement | null>(null)
const gaugeChartRef = ref<HTMLElement | null>(null)
const lineChart = shallowRef<echarts.ECharts | null>(null)
const pieChart = shallowRef<echarts.ECharts | null>(null)
const gaugeChart = shallowRef<echarts.ECharts | null>(null)
const cpuHistory = ref<number[]>([])
const memoryHistory = ref<number[]>([])
const historyLabels = ref<string[]>([])
const showUpdateDialog = ref(false)
const maxHistoryLength = 20

function formatSize(kb: number): string {
  if (kb === 0) return '0 B'
  const gb = kb / 1024 / 1024
  if (gb >= 1) return `${gb.toFixed(1)} GB`
  const mb = kb / 1024
  if (mb >= 1) return `${mb.toFixed(1)} MB`
  return `${kb.toFixed(0)} KB`
}

function updateHistory() {
  const now = new Date()
  const timeLabel = `${now.getHours().toString().padStart(2, '0')}:${now.getMinutes().toString().padStart(2, '0')}:${now.getSeconds().toString().padStart(2, '0')}`
  cpuHistory.value.push(systemStore.stats.cpu)
  memoryHistory.value.push(systemStore.stats.memory)
  historyLabels.value.push(timeLabel)

  if (cpuHistory.value.length > maxHistoryLength) {
    cpuHistory.value.shift()
    memoryHistory.value.shift()
    historyLabels.value.shift()
  }

  updateLineChart()
  updateGaugeChart()
}

function initLineChart() {
  if (!lineChartRef.value) return
  lineChart.value = echarts.init(lineChartRef.value)
  updateLineChart()
}

function updateLineChart() {
  if (!lineChart.value) return
  const option: echarts.EChartsOption = {
    tooltip: {
      trigger: 'axis',
      formatter: (params: any) => {
        if (!Array.isArray(params)) return ''
        let result = params[0]?.axisValue || ''
        params.forEach((item: any) => {
          result += `<br/>${item.marker}${item.seriesName}: ${item.value.toFixed(1)}%`
        })
        return result
      },
    },
    legend: { data: ['CPU', t('system.memory')], bottom: 0 },
    grid: { left: '3%', right: '4%', bottom: '15%', top: '10%', containLabel: true },
    xAxis: { type: 'category', boundaryGap: false, data: historyLabels.value, axisLabel: { fontSize: 10, rotate: 30 } },
    yAxis: { type: 'value', min: 0, max: 100, axisLabel: { formatter: '{value}%' } },
    series: [
      {
        name: 'CPU',
        type: 'line',
        smooth: true,
        data: cpuHistory.value,
        itemStyle: { color: '#409EFF' },
        areaStyle: {
          color: new echarts.graphic.LinearGradient(0, 0, 0, 1, [
            { offset: 0, color: 'rgba(64,158,255,0.3)' },
            { offset: 1, color: 'rgba(64,158,255,0.05)' },
          ]),
        },
      },
      {
        name: t('system.memory'),
        type: 'line',
        smooth: true,
        data: memoryHistory.value,
        itemStyle: { color: '#67C23A' },
        areaStyle: {
          color: new echarts.graphic.LinearGradient(0, 0, 0, 1, [
            { offset: 0, color: 'rgba(103,194,58,0.3)' },
            { offset: 1, color: 'rgba(103,194,58,0.05)' },
          ]),
        },
      },
    ],
  }
  lineChart.value.setOption(option)
}

function initPieChart() {
  if (!pieChartRef.value) return
  pieChart.value = echarts.init(pieChartRef.value)
  updatePieChart()
}

function updatePieChart() {
  if (!pieChart.value) return
  const stats = containerStats.value
  const pieData = [
    { value: stats.running, name: t('container.running'), itemStyle: { color: '#67C23A' } },
    { value: stats.stopped, name: t('container.stopped'), itemStyle: { color: '#F56C6C' } },
    { value: stats.paused, name: t('container.paused'), itemStyle: { color: '#E6A23C' } },
  ]
  const displayData = pieData.filter(item => item.value > 0)
  if (displayData.length === 0) {
    displayData.push({ value: 1, name: 'No Containers', itemStyle: { color: '#dcdfe6' } })
  }

  const option: echarts.EChartsOption = {
    tooltip: { trigger: 'item', formatter: '{b}: {c} ({d}%)' },
    legend: { bottom: 0, itemWidth: 10, itemHeight: 10, data: pieData.map(item => item.name) },
    series: [
      {
        type: 'pie',
        radius: ['40%', '70%'],
        center: ['50%', '45%'],
        itemStyle: { borderRadius: 4, borderColor: '#fff', borderWidth: 2 },
        label: {
          show: true,
          position: 'center',
          formatter: () => `${stats.total}\nContainers`,
          fontSize: 16,
          fontWeight: 'bold',
        },
        data: displayData,
      },
    ],
  }
  pieChart.value.setOption(option)
}

function initGaugeChart() {
  if (!gaugeChartRef.value) return
  gaugeChart.value = echarts.init(gaugeChartRef.value)
  updateGaugeChart()
}

function updateGaugeChart() {
  if (!gaugeChart.value) return
  const diskValue = systemStore.stats.disk
  const option: echarts.EChartsOption = {
    series: [
      {
        type: 'gauge',
        startAngle: 200,
        endAngle: -20,
        min: 0,
        max: 100,
        itemStyle: { color: diskValue >= 90 ? '#F56C6C' : diskValue >= 70 ? '#E6A23C' : '#67C23A' },
        progress: { show: true, width: 20 },
        pointer: { show: false },
        axisLine: { lineStyle: { width: 20, color: [[1, '#e0e0e0']] } },
        axisTick: { show: false },
        splitLine: { show: false },
        axisLabel: { show: false },
        title: { show: true, offsetCenter: [0, '30%'], fontSize: 14, color: '#999' },
        detail: { valueAnimation: true, offsetCenter: [0, '-10%'], fontSize: 28, fontWeight: 'bold', formatter: '{value}%', color: 'inherit' },
        data: [{ value: diskValue, name: `${t('system.disk')} ${t('system.usage')}` }],
      },
    ],
  }
  gaugeChart.value.setOption(option)
}

function handleResize() {
  lineChart.value?.resize()
  pieChart.value?.resize()
  gaugeChart.value?.resize()
}

function getProgressColor(percent: number): string {
  if (percent >= 90) return '#F56C6C'
  if (percent >= 70) return '#E6A23C'
  return '#67C23A'
}

const containerStats = computed(() => {
  const containers = containerStore.containers
  const running = containers.filter(c => c.state === 'running').length
  const paused = containers.filter(c => c.state === 'paused').length
  const stopped = containers.filter(c => c.state !== 'running' && c.state !== 'paused').length
  return { total: containers.length, running, stopped, paused }
})

let historyTimer: ReturnType<typeof setInterval> | null = null

onMounted(() => {
  systemStore.startPolling()
  containerStore.fetchContainers()
  for (let i = 0; i < 5; i++) {
    cpuHistory.value.push(0)
    memoryHistory.value.push(0)
    historyLabels.value.push('--:--:--')
  }
  setTimeout(() => {
    initLineChart()
    initPieChart()
    initGaugeChart()
  }, 100)
  historyTimer = setInterval(updateHistory, 5000)
  setTimeout(updateHistory, 1000)
  window.addEventListener('resize', handleResize)
})

onUnmounted(() => {
  systemStore.stopPolling()
  if (historyTimer) clearInterval(historyTimer)
  window.removeEventListener('resize', handleResize)
  lineChart.value?.dispose()
  pieChart.value?.dispose()
  gaugeChart.value?.dispose()
})

watch(() => containerStore.containers, () => updatePieChart(), { deep: true })
</script>

<template>
  <div class="dashboard">
    <UpdateBanner @open-dialog="showUpdateDialog = true" />

    <div class="stats-row">
      <el-card class="stat-card" shadow="hover">
        <div class="stat-content">
          <div class="stat-icon cpu"><el-icon :size="32"><Monitor /></el-icon></div>
          <div class="stat-info">
            <div class="stat-value">{{ systemStore.stats.cpu.toFixed(1) }}%</div>
            <div class="stat-label">CPU {{ t('system.usage') }}</div>
          </div>
        </div>
        <el-progress :percentage="systemStore.stats.cpu" :stroke-width="6" :color="getProgressColor(systemStore.stats.cpu)" :show-text="false" />
      </el-card>

      <el-card class="stat-card" shadow="hover">
        <div class="stat-content">
          <div class="stat-icon memory"><el-icon :size="32"><Box /></el-icon></div>
          <div class="stat-info">
            <div class="stat-value">{{ systemStore.stats.memory.toFixed(1) }}%</div>
            <div class="stat-label">{{ t('system.memory') }} {{ t('system.usage') }}</div>
            <div class="stat-detail">{{ formatSize(systemStore.stats.memoryUsed) }} / {{ formatSize(systemStore.stats.memoryTotal) }}</div>
          </div>
        </div>
        <el-progress :percentage="systemStore.stats.memory" :stroke-width="6" :color="getProgressColor(systemStore.stats.memory)" :show-text="false" />
      </el-card>

      <el-card class="stat-card" shadow="hover">
        <div class="stat-content">
          <div class="stat-icon disk"><el-icon :size="32"><Picture /></el-icon></div>
          <div class="stat-info">
            <div class="stat-value">{{ systemStore.stats.disk.toFixed(1) }}%</div>
            <div class="stat-label">{{ t('system.disk') }} {{ t('system.usage') }}</div>
            <div class="stat-detail">{{ formatSize(systemStore.stats.diskUsed) }} / {{ formatSize(systemStore.stats.diskTotal) }}</div>
          </div>
        </div>
        <el-progress :percentage="systemStore.stats.disk" :stroke-width="6" :color="getProgressColor(systemStore.stats.disk)" :show-text="false" />
      </el-card>

      <el-card class="stat-card" shadow="hover">
        <div class="stat-content">
          <div class="stat-icon containers"><el-icon :size="32"><Connection /></el-icon></div>
          <div class="stat-info">
            <div class="stat-value">{{ containerStats.running }} / {{ containerStats.total }}</div>
            <div class="stat-label">{{ t('nav.containers') }}</div>
          </div>
        </div>
        <div class="container-stats">
          <el-tag type="success" size="small">{{ t('container.running') }}: {{ containerStats.running }}</el-tag>
          <el-tag type="danger" size="small">{{ t('container.stopped') }}: {{ containerStats.stopped }}</el-tag>
        </div>
      </el-card>
    </div>

    <div class="charts-row">
      <el-card class="chart-card" shadow="hover">
        <template #header><span>CPU / {{ t('system.memory') }} {{ t('system.usage') }}</span></template>
        <div ref="lineChartRef" class="chart-container"></div>
      </el-card>

      <el-card class="chart-card" shadow="hover">
        <template #header><span>{{ t('nav.containers') }}</span></template>
        <div ref="pieChartRef" class="chart-container"></div>
      </el-card>
    </div>

    <div class="charts-row-2">
      <el-card class="chart-card" shadow="hover">
        <template #header><span>{{ t('system.disk') }} {{ t('system.usage') }}</span></template>
        <div ref="gaugeChartRef" class="chart-container"></div>
      </el-card>

      <el-card class="info-card" shadow="hover">
        <template #header><span>Server Info</span></template>
        <div class="server-info">
          <div class="info-item"><span class="info-label">{{ t('system.serverTime') }}</span><span class="info-value">{{ systemStore.stats.time || '--' }}</span></div>
          <div class="info-item"><span class="info-label">Containers</span><span class="info-value">{{ containerStats.total }}</span></div>
          <div class="info-item"><span class="info-label">Running</span><span class="info-value success">{{ containerStats.running }}</span></div>
          <div class="info-item"><span class="info-label">Stopped</span><span class="info-value danger">{{ containerStats.stopped }}</span></div>
        </div>
      </el-card>
    </div>
  </div>

  <UpdateDialog v-model="showUpdateDialog" />
</template>

<style scoped>
.dashboard { padding: 20px; }
.stats-row { display: grid; grid-template-columns: repeat(4, 1fr); gap: 20px; margin-bottom: 20px; }
.stat-card { padding: 10px; }
.stat-content { display: flex; align-items: center; gap: 15px; margin-bottom: 15px; }
.stat-icon { width: 60px; height: 60px; border-radius: 12px; display: flex; align-items: center; justify-content: center; color: white; }
.stat-icon.cpu { background: linear-gradient(135deg, #409EFF, #66b1ff); }
.stat-icon.memory { background: linear-gradient(135deg, #67C23A, #85ce61); }
.stat-icon.disk { background: linear-gradient(135deg, #E6A23C, #ebb563); }
.stat-icon.containers { background: linear-gradient(135deg, #909399, #a6a9ad); }
.stat-info { flex: 1; }
.stat-value { font-size: 28px; font-weight: 600; color: var(--el-text-color-primary); }
.stat-label { font-size: 14px; color: var(--el-text-color-secondary); margin-top: 4px; }
.stat-detail { font-size: 12px; color: var(--el-text-color-secondary); margin-top: 2px; }
.container-stats { display: flex; gap: 8px; margin-top: 10px; }
.charts-row { display: grid; grid-template-columns: 2fr 1fr; gap: 20px; margin-bottom: 20px; }
.charts-row-2 { display: grid; grid-template-columns: 1fr 1fr; gap: 20px; margin-bottom: 20px; }
.chart-card { min-height: 300px; }
.chart-container { width: 100%; height: 250px; }
.info-card { min-height: 300px; }
.server-info { display: grid; grid-template-columns: repeat(2, 1fr); gap: 30px; padding: 20px; }
.info-item { display: flex; flex-direction: column; gap: 8px; }
.info-label { font-size: 14px; color: var(--el-text-color-secondary); }
.info-value { font-size: 24px; font-weight: 600; color: var(--el-text-color-primary); }
.info-value.success { color: #67C23A; }
.info-value.danger { color: #F56C6C; }
@media (max-width: 1200px) { .stats-row { grid-template-columns: repeat(2, 1fr); } .charts-row, .charts-row-2 { grid-template-columns: 1fr; } }
@media (max-width: 768px) { .stats-row { grid-template-columns: 1fr; } .server-info { grid-template-columns: 1fr; gap: 20px; } }
</style>
