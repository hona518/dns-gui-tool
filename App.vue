<template>
  <div class="app-container">
    <!-- 顶部菜单栏与厂商标签栏 -->
    <div class="header-area">
      <div class="provider-tabs">
        <button v-for="tab in providers" :key="tab.id"
                :class="{ active: currentProvider === tab.id }"
                @click="selectProvider(tab.id)">
          {{ tab.name }}
        </button>
      </div>
      <button class="settings-btn" @click="openSettings">⚙️ API 配置</button>
    </div>

    <!-- 主内容区 -->
    <div class="main-content">
      <div class="domain-list">
        <ul>
          <li v-for="domain in domains" :key="domain"
              :class="{ active: currentDomain === domain, warning: domain.includes('配置') }"
              @click="loadRecords(domain)">
            {{ domain }}
          </li>
        </ul>
      </div>

      <div class="record-manager">
        <div v-if="currentDomain && !currentDomain.includes('配置')">
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
      </div>
    </div>

    <!-- API 配置弹窗 (Modal) -->
    <div class="modal-overlay" v-if="showSettings">
      <div class="modal-content">
        <h3>API 密钥配置</h3>
        
        <div class="form-group">
          <label>华为云 Access Key (AK):</label>
          <input type="text" v-model="config.huaweiAK" placeholder="输入 AK..." />
        </div>
        <div class="form-group">
          <label>华为云 Secret Key (SK):</label>
          <input type="password" v-model="config.huaweiSK" placeholder="输入 SK..." />
        </div>
        
        <div class="modal-actions">
          <button @click="showSettings = false">取消</button>
          <button class="primary" @click="saveSettings">保存并生效</button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue';

const providers = [
  { id: 'huawei', name: '华为云' },
  { id: 'cloudflare', name: 'Cloudflare' },
  { id: 'aliyun', name: '阿里云' },
  { id: 'tencent', name: '腾讯云' }
];

const currentProvider = ref('');
const currentDomain = ref('');
const domains = ref([]);
const records = ref([]);

// 设置弹窗状态与表单数据
const showSettings = ref(false);
const config = ref({ huaweiAK: '', huaweiSK: '' });

const selectProvider = async (pid) => {
  currentProvider.value = pid;
  currentDomain.value = '';
  records.value = [];
  if (window.go && window.go.main) {
    domains.value = await window.go.main.DNSService.GetDomains(pid);
  }
};

const loadRecords = async (domain) => {
  currentDomain.value = domain;
  if (window.go && window.go.main && !domain.includes('配置')) {
    records.value = await window.go.main.DNSService.GetRecords(currentProvider.value, domain);
  }
};

// 打开设置时，读取已保存的配置
const openSettings = async () => {
  if (window.go && window.go.main) {
    config.value = await window.go.main.DNSService.LoadConfig();
  }
  showSettings.value = true;
};

// 保存设置
const saveSettings = async () => {
  if (window.go && window.go.main) {
    await window.go.main.DNSService.SaveConfig(config.value);
    showSettings.value = false;
    selectProvider(currentProvider.value); // 重新刷新左侧列表
  }
};

onMounted(() => {
  selectProvider('huawei');
});
</script>

<style scoped>
.app-container { display: flex; flex-direction: column; height: 100vh; background-color: #f9f9f9; font-family: 'Microsoft YaHei', sans-serif; }
.header-area { display: flex; justify-content: space-between; align-items: center; background: #e0e0e0; border-bottom: 1px solid #ccc; padding-right: 15px; }
.provider-tabs { display: flex; flex: 1; }
.provider-tabs button { padding: 12px 20px; border: none; background: transparent; cursor: pointer; color: #333; outline: none; }
.provider-tabs button:hover { background: #d5d5d5; }
.provider-tabs button.active { background: #fff; border-top: 3px solid #E61520; font-weight: bold; } /* 华为红 */
.settings-btn { padding: 6px 12px; cursor: pointer; border: 1px solid #ccc; background: #fff; border-radius: 4px; }

.main-content { display: flex; flex: 1; overflow: hidden; }
.domain-list { width: 220px; background: #fff; border-right: 1px solid #ddd; overflow-y: auto; }
.domain-list ul { list-style: none; margin: 0; padding: 0; }
.domain-list li { padding: 15px 20px; cursor: pointer; border-bottom: 1px solid #f0f0f0; }
.domain-list li:hover { background: #f5f5f5; }
.domain-list li.active { background: #e5f1fb; color: #0078D7; border-left: 4px solid #0078D7; }
.domain-list li.warning { color: #d9534f; font-size: 13px; }

.record-manager { flex: 1; padding: 30px; background: #fafafa; overflow-y: auto; }
.record-table { width: 100%; border-collapse: collapse; margin-top: 20px; background: #fff; }
.record-table th, .record-table td { border: 1px solid #eaeaea; padding: 12px 15px; text-align: left; }
.record-table th { background: #f4f4f4; }

/* 弹窗样式 */
.modal-overlay { position: fixed; top: 0; left: 0; width: 100%; height: 100%; background: rgba(0,0,0,0.4); display: flex; justify-content: center; align-items: center; }
.modal-content { background: #fff; padding: 25px; width: 400px; border-radius: 6px; box-shadow: 0 4px 12px rgba(0,0,0,0.15); }
.form-group { margin-bottom: 15px; }
.form-group label { display: block; margin-bottom: 5px; font-size: 14px; }
.form-group input { width: 100%; padding: 8px; border: 1px solid #ccc; border-radius: 4px; box-sizing: border-box; }
.modal-actions { display: flex; justify-content: flex-end; gap: 10px; margin-top: 20px; }
.modal-actions button { padding: 8px 15px; cursor: pointer; border: 1px solid #ccc; background: #fff; border-radius: 4px; }
.modal-actions button.primary { background: #0078D7; color: #fff; border: none; }
</style>
