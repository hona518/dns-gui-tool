<template>
  <div class="app-container">
    <!-- 顶部菜单栏 -->
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
              :class="{ active: currentDomain === domain, warning: domain.includes('暂无') }"
              @click="loadRecords(domain)">
            {{ domain }}
          </li>
        </ul>
      </div>

      <div class="record-manager">
        <div v-if="currentDomain && !currentDomain.includes('暂无')">
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

    <!-- 底部运行日志面板 -->
    <div class="log-panel">
      <div class="log-header">实时运行日志</div>
      <div class="log-content">
        <div v-for="(log, index) in logs" :key="index" :class="['log-item', log.level]">
          <span class="log-time">[{{ log.time }}]</span> {{ log.msg }}
        </div>
      </div>
    </div>

    <!-- 1Panel 风格 API 配置弹窗 -->
    <div class="modal-overlay" v-if="showSettings">
      <div class="modal-content">
        <div class="modal-header">
          <h3>创建</h3>
          <button class="close-btn" @click="showSettings = false">×</button>
        </div>
        
        <div class="modal-body">
          <div class="form-group">
            <label><span class="required">*</span> 名称</label>
            <input type="text" v-model="form.name" />
          </div>

          <div class="form-group">
            <label><span class="required">*</span> 类型</label>
            <select v-model="formType" @change="syncFormToType">
              <option value="huawei">华为云</option>
              <option value="aliyun">阿里云</option>
              <option value="cloudflare">Cloudflare</option>
              <option value="tencent">腾讯云</option>
            </select>
          </div>

          <!-- 华为云 / 阿里云 / 腾讯云 通用字段 -->
          <div class="form-group" v-if="formType !== 'cloudflare'">
            <label><span class="required">*</span> {{ formType === 'tencent' ? 'Secret ID' : 'Access key' }}</label>
            <input type="text" v-model="form.accessKey" />
          </div>
          <div class="form-group">
            <label><span class="required">*</span> {{ formType === 'cloudflare' ? 'API Token' : 'Secret key' }}</label>
            <input type="password" v-model="form.secretKey" />
            <span class="hint" v-if="formType === 'cloudflare'">请勿使用 Global API Key</span>
          </div>

          <!-- Cloudflare 特有字段 -->
          <div class="form-group" v-if="formType === 'cloudflare'">
            <label>EMAIL</label>
            <input type="text" v-model="form.email" />
          </div>

          <!-- 华为云 特有字段 -->
          <div class="form-group" v-if="formType === 'huawei'">
            <label>Region</label>
            <input type="text" v-model="form.region" placeholder="cn-north-1" />
          </div>
        </div>
        
        <div class="modal-footer">
          <button class="btn-cancel" @click="showSettings = false">取消</button>
          <button class="btn-confirm" @click="saveSettings">确认</button>
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
const logs = ref([]);

const addLog = (msg, level = 'info') => {
  const time = new Date().toLocaleTimeString('zh-CN', { hour12: false });
  logs.value.unshift({ time, msg, level });
};

// 全局配置与弹窗表单状态
const showSettings = ref(false);
const globalConfig = ref({});
const formType = ref('huawei');
const form = ref({ name: '', accessKey: '', secretKey: '', region: 'cn-north-1', email: '' });

const selectProvider = async (pid) => {
  currentProvider.value = pid;
  currentDomain.value = '';
  records.value = [];
  addLog(`切换至 ${providers.find(p => p.id === pid).name} 界面`, 'info');
  
  if (window.go && window.go.main) {
    domains.value = await window.go.main.DNSService.GetDomains(pid);
    addLog(`拉取域名列表: 找到 ${domains.value.length} 条记录`, 'success');
  }
};

const loadRecords = async (domain) => {
  currentDomain.value = domain;
  addLog(`准备加载域名 ${domain} 的解析记录...`, 'info');
  if (window.go && window.go.main && !domain.includes('暂无')) {
    records.value = await window.go.main.DNSService.GetRecords(currentProvider.value, domain);
    addLog(`记录加载成功，共 ${records.value.length} 条`, 'success');
  }
};

const openSettings = async () => {
  if (window.go && window.go.main) {
    globalConfig.value = await window.go.main.DNSService.LoadConfig();
  }
  formType.value = currentProvider.value || 'huawei';
  syncFormToType();
  showSettings.value = true;
  addLog('打开 1Panel 风格 API 配置面板', 'info');
};

const syncFormToType = () => {
  const data = globalConfig.value[formType.value] || {};
  form.value = {
    name: data.name || '',
    accessKey: data.accessKey || '',
    secretKey: data.secretKey || '',
    region: data.region || (formType.value === 'huawei' ? 'cn-north-1' : ''),
    email: data.email || ''
  };
};

const saveSettings = async () => {
  if (!form.value.name || !form.value.secretKey) {
    addLog('保存失败: 请填写所有带红星的必填项', 'error');
    return;
  }
  
  globalConfig.value[formType.value] = { ...form.value };
  
  if (window.go && window.go.main) {
    await window.go.main.DNSService.SaveConfig(globalConfig.value);
    showSettings.value = false;
    addLog(`[${form.value.name}] API 凭证本地保存成功`, 'success');
    selectProvider(currentProvider.value); 
  }
};

onMounted(() => {
  addLog('终端界面初始化完毕，等待 API 连通测试', 'info');
  selectProvider('huawei');
});
</script>

<style scoped>
.app-container { display: flex; flex-direction: column; height: 100vh; background-color: #f0f2f5; font-family: 'Microsoft YaHei', sans-serif; }
.header-area { display: flex; justify-content: space-between; align-items: center; background: #fff; border-bottom: 1px solid #dcdfe6; padding-right: 20px; }
.provider-tabs { display: flex; }
.provider-tabs button { padding: 15px 25px; border: none; background: transparent; cursor: pointer; color: #606266; font-size: 15px; }
.provider-tabs button:hover { color: #1890ff; }
.provider-tabs button.active { background: #fff; color: #1890ff; border-bottom: 2px solid #1890ff; font-weight: bold; }
.settings-btn { padding: 8px 15px; cursor: pointer; border: 1px solid #dcdfe6; background: #fff; border-radius: 4px; color: #606266; }

.main-content { display: flex; flex: 1; overflow: hidden; }
.domain-list { width: 220px; background: #fff; border-right: 1px solid #e4e7ed; overflow-y: auto; }
.domain-list ul { list-style: none; margin: 0; padding: 0; }
.domain-list li { padding: 15px 20px; cursor: pointer; border-bottom: 1px solid #ebeef5; font-size: 14px; }
.domain-list li.active { background: #ecf5ff; color: #409eff; }
.domain-list li.warning { color: #f56c6c; font-size: 13px; }

.record-manager { flex: 1; padding: 25px; overflow-y: auto; }
.record-table { width: 100%; border-collapse: collapse; margin-top: 15px; background: #fff; }
.record-table th, .record-table td { border: 1px solid #ebeef5; padding: 12px; text-align: left; }
.record-table th { background: #fafafa; color: #909399; font-weight: normal; }

/* 日志面板 */
.log-panel { height: 160px; background: #1e1e1e; border-top: 1px solid #333; display: flex; flex-direction: column; }
.log-header { padding: 8px 15px; background: #2d2d2d; color: #ccc; font-size: 13px; font-weight: bold; border-bottom: 1px solid #1e1e1e; }
.log-content { flex: 1; padding: 10px 15px; overflow-y: auto; font-family: 'Consolas', monospace; font-size: 13px; }
.log-item { margin-bottom: 5px; line-height: 1.4; }
.log-time { color: #888; margin-right: 8px; }
.log-item.info { color: #d4d4d4; }
.log-item.success { color: #4CAF50; }
.log-item.error { color: #F44336; }

/* 1Panel 风格弹窗 */
.modal-overlay { position: fixed; top: 0; left: 0; width: 100%; height: 100%; background: rgba(0,0,0,0.5); display: flex; justify-content: center; align-items: center; }
.modal-content { background: #fff; width: 480px; border-radius: 4px; box-shadow: 0 4px 16px rgba(0,0,0,0.2); display: flex; flex-direction: column; }
.modal-header { padding: 15px 20px; border-bottom: 1px solid #ebeef5; display: flex; justify-content: space-between; align-items: center; }
.modal-header h3 { margin: 0; font-size: 16px; color: #303133; font-weight: normal; }
.close-btn { background: none; border: none; font-size: 20px; color: #909399; cursor: pointer; }
.modal-body { padding: 20px 30px; }
.form-group { margin-bottom: 20px; }
.form-group label { display: block; margin-bottom: 8px; font-size: 14px; color: #606266; }
.required { color: #f56c6c; margin-right: 4px; }
.form-group input, .form-group select { width: 100%; padding: 10px; border: 1px solid #dcdfe6; border-radius: 4px; box-sizing: border-box; outline: none; font-size: 14px; }
.form-group input:focus, .form-group select:focus { border-color: #409eff; }
.hint { display: block; font-size: 12px; color: #909399; margin-top: 5px; }
.modal-footer { padding: 15px 20px; border-top: 1px solid #ebeef5; display: flex; justify-content: flex-end; gap: 10px; }
.btn-cancel { padding: 9px 20px; cursor: pointer; border: 1px solid #dcdfe6; background: #fff; border-radius: 4px; color: #606266; }
.btn-confirm { padding: 9px 20px; cursor: pointer; border: none; background: #165dff; border-radius: 4px; color: #fff; }
</style>
