<template>
  <div class="app-container">
    <!-- 顶部厂商标签栏 -->
    <div class="provider-tabs">
      <button v-for="tab in providers" :key="tab.id"
              :class="{ active: currentProvider === tab.id }"
              @click="selectProvider(tab.id)">
        {{ tab.name }}
      </button>
    </div>

    <!-- 主内容区 -->
    <div class="main-content">
      <!-- 左侧域名列表 -->
      <div class="domain-list">
        <ul>
          <li v-for="domain in domains" :key="domain"
              :class="{ active: currentDomain === domain }"
              @click="loadRecords(domain)">
            {{ domain }}
          </li>
          <li v-if="domains.length === 0" class="empty-text">暂无域名</li>
        </ul>
      </div>

      <!-- 右侧解析记录管理 -->
      <div class="record-manager">
        <div v-if="currentDomain">
          <h3>{{ currentDomain }} 解析记录</h3>
          <table class="record-table">
            <thead>
              <tr>
                <th>类型</th>
                <th>主机记录</th>
                <th>记录值</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="record in records" :key="record.id">
                <td>{{ record.type }}</td>
                <td>{{ record.name }}</td>
                <td>{{ record.content }}</td>
              </tr>
            </tbody>
          </table>
        </div>
        <div v-else class="empty-text" style="margin-top: 50px;">
          👈 请在左侧选择一个域名查看解析记录
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue';

const providers = [
  { id: 'cloudflare', name: 'Cloudflare' },
  { id: 'huawei', name: '华为云' },
  { id: 'aliyun', name: '阿里云' },
  { id: 'tencent', name: '腾讯云' }
];

const currentProvider = ref('');
const currentDomain = ref('');
const domains = ref([]);
const records = ref([]);

// 动态调用 Go 后端方法，彻底避免 import 路径报错问题
const selectProvider = async (pid) => {
  currentProvider.value = pid;
  currentDomain.value = '';
  records.value = [];
  
  if (window.go && window.go.main && window.go.main.DNSService) {
    domains.value = await window.go.main.DNSService.GetDomains(pid);
  }
};

const loadRecords = async (domain) => {
  currentDomain.value = domain;
  if (window.go && window.go.main && window.go.main.DNSService) {
    records.value = await window.go.main.DNSService.GetRecords(currentProvider.value, domain);
  }
};

// 关键修复：当软件打开（组件挂载）时，自动触发点击 Cloudflare 标签
onMounted(() => {
  selectProvider('cloudflare');
});
</script>

<style scoped>
/* 针对 Windows 原生质感进行样式优化 */
.app-container { display: flex; flex-direction: column; height: 100vh; background-color: #f9f9f9; font-family: 'Microsoft YaHei', 'Segoe UI', sans-serif; margin: 0; padding: 0; }
.provider-tabs { display: flex; background: #e0e0e0; border-bottom: 1px solid #ccc; }
.provider-tabs button { flex: 1; padding: 12px 0; border: none; background: transparent; cursor: pointer; font-size: 14px; color: #333; outline: none; transition: background 0.2s; }
.provider-tabs button:hover { background: #d5d5d5; }
.provider-tabs button.active { background: #fff; border-top: 3px solid #0078D7; font-weight: bold; }

.main-content { display: flex; flex: 1; overflow: hidden; }
.domain-list { width: 200px; background: #fff; border-right: 1px solid #ddd; overflow-y: auto; }
.domain-list ul { list-style: none; margin: 0; padding: 0; }
.domain-list li { padding: 15px 20px; cursor: pointer; border-bottom: 1px solid #f0f0f0; transition: background 0.2s; font-size: 14px; }
.domain-list li:hover { background: #f5f5f5; }
.domain-list li.active { background: #e5f1fb; color: #0078D7; font-weight: bold; border-left: 4px solid #0078D7; padding-left: 16px; }

.record-manager { flex: 1; padding: 30px; background: #fafafa; overflow-y: auto; }
.record-table { width: 100%; border-collapse: collapse; margin-top: 20px; background: #fff; box-shadow: 0 1px 3px rgba(0,0,0,0.1); }
.record-table th, .record-table td { border: 1px solid #eaeaea; padding: 12px 15px; text-align: left; font-size: 14px; }
.record-table th { background: #f4f4f4; color: #555; font-weight: 500; }
h3 { margin-top: 0; color: #333; font-weight: 500; }
.empty-text { color: #999; text-align: center; padding: 20px; font-size: 14px; }
</style>
