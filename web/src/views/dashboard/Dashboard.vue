<template>
  <div class="dashboard">
    <el-row :gutter="20" class="stat-cards">
      <el-col :span="6">
        <el-card shadow="hover" class="stat-card">
          <div class="stat-content">
            <div class="stat-info">
              <div class="stat-label">总用户数</div>
              <div class="stat-value">{{ stats.total_users }}</div>
            </div>
            <el-icon class="stat-icon" style="color: #409eff"><User /></el-icon>
          </div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card shadow="hover" class="stat-card">
          <div class="stat-content">
            <div class="stat-info">
              <div class="stat-label">活跃用户</div>
              <div class="stat-value">{{ stats.active_users }}</div>
            </div>
            <el-icon class="stat-icon" style="color: #67c23a"><UserFilled /></el-icon>
          </div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card shadow="hover" class="stat-card">
          <div class="stat-content">
            <div class="stat-info">
              <div class="stat-label">在线节点</div>
              <div class="stat-value">{{ stats.online_nodes }} / {{ stats.total_nodes }}</div>
            </div>
            <el-icon class="stat-icon" style="color: #e6a23c"><Connection /></el-icon>
          </div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card shadow="hover" class="stat-card">
          <div class="stat-content">
            <div class="stat-info">
              <div class="stat-label">今日收入</div>
              <div class="stat-value">&yen;{{ stats.today_income.toFixed(2) }}</div>
            </div>
            <el-icon class="stat-icon" style="color: #f56c6c"><Wallet /></el-icon>
          </div>
        </el-card>
      </el-col>
    </el-row>

    <el-row :gutter="20" style="margin-top: 20px">
      <el-col :span="16">
        <el-card shadow="hover">
          <template #header>
            <div class="card-header">
              <span>流量趋势（近7天）</span>
            </div>
          </template>
          <div class="traffic-placeholder">
            <el-empty v-if="trafficData.length === 0" description="暂无流量数据" />
            <div v-else class="traffic-list">
              <div v-for="item in trafficData" :key="item.date" class="traffic-item">
                <span class="traffic-date">{{ item.date }}</span>
                <span class="traffic-value">{{ formatTraffic(item.upload + item.download) }}</span>
              </div>
            </div>
          </div>
        </el-card>
      </el-col>
      <el-col :span="8">
        <el-card shadow="hover">
          <template #header>
            <div class="card-header">
              <span>本月概览</span>
            </div>
          </template>
          <div class="overview-list">
            <div class="overview-item">
              <span>本月收入</span>
              <span class="overview-value">&yen;{{ stats.month_income.toFixed(2) }}</span>
            </div>
            <div class="overview-item">
              <span>总流量使用</span>
              <span class="overview-value">{{ formatTraffic(stats.total_traffic) }}</span>
            </div>
            <div class="overview-item">
              <span>节点在线率</span>
              <span class="overview-value">{{ nodeOnlineRate }}%</span>
            </div>
          </div>
        </el-card>
      </el-col>
    </el-row>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted } from 'vue'
import { getOverview, getTraffic } from '@/api/stats'
import type { TrafficStats } from '@/types'

const stats = reactive({
  total_users: 0,
  active_users: 0,
  total_traffic: 0,
  total_nodes: 0,
  online_nodes: 0,
  today_income: 0,
  month_income: 0
})

const trafficData = ref<TrafficStats[]>([])

const nodeOnlineRate = computed(() => {
  if (stats.total_nodes === 0) return 0
  return ((stats.online_nodes / stats.total_nodes) * 100).toFixed(1)
})

function formatTraffic(bytes: number): string {
  if (bytes === 0) return '0 B'
  const units = ['B', 'KB', 'MB', 'GB', 'TB']
  const i = Math.floor(Math.log(bytes) / Math.log(1024))
  return (bytes / Math.pow(1024, i)).toFixed(2) + ' ' + units[i]
}

onMounted(async () => {
  try {
    const res = await getOverview()
    Object.assign(stats, res.data)
  } catch {
    // ignore
  }
  try {
    const res = await getTraffic(7)
    trafficData.value = res.data
  } catch {
    // ignore
  }
})
</script>

<style scoped>
.stat-card .stat-content {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.stat-label {
  font-size: 14px;
  color: #909399;
  margin-bottom: 8px;
}

.stat-value {
  font-size: 24px;
  font-weight: 600;
  color: #303133;
}

.stat-icon {
  font-size: 48px;
  opacity: 0.3;
}

.card-header {
  font-weight: 600;
  color: #303133;
}

.traffic-placeholder {
  min-height: 300px;
  display: flex;
  align-items: center;
  justify-content: center;
}

.traffic-list {
  width: 100%;
}

.traffic-item {
  display: flex;
  justify-content: space-between;
  padding: 10px 0;
  border-bottom: 1px solid #f0f0f0;
}

.traffic-date {
  color: #606266;
}

.traffic-value {
  font-weight: 600;
  color: #303133;
}

.overview-list {
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.overview-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 10px 0;
  border-bottom: 1px solid #f0f0f0;
}

.overview-value {
  font-weight: 600;
  color: #409eff;
}
</style>
