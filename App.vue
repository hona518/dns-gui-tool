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
              :class="{ active: currentDomain === domain, warning: domain.includes('暂无') || domain.includes('报错') }"
              @click="loadRecords(domain)">
            {{ domain }}
          </li>
        </ul>
      </div>

      <div class="record-manager">
        <div v-if="currentDomain && !currentDomain.includes('暂无') && !currentDomain.includes('报错')">
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
                <td class="content-cell">{{ record.content }}</td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>
    </div>

    <!-- 可拖拽高度的日志面板 -->
    <div class="log-panel-container" v-show="showLogs" :style="{ height: logPanelHeight + 'px' }">
      <!-- 拖拽调整条 -->
      <div class="resizer" @mousedown="startDrag">
        <div class="resizer-line"></div>
      </div>
      <div class="log-panel">
        <div class="log-header">
          <span>实时运行日志</span>
          <button class="hide-log-btn" @click="showLogs = false">▼ 隐藏</button>
        </div>
        <div class="log-content">
          <div v-for="(log, index) in logs" :key="index" :class="['log-item', log.level]">
            <span class="log-time">[{{ log.time }}]</span> {{ log.msg }}
          </div>
        </div>
      </div>
    </div>

    <!-- 悬浮显示日志按钮 -->
    <button class="floating-log-btn" v-show="!showLogs" @click="showLogs = true">
      ▲ 显示日志
    </button>

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

          <div class="form-group" v-if="formType !== 'cloudflare'">
            <label><span class="required">*</span> {{ formType === 'tencent' ? 'Secret ID' : 'Access key' }}</label>
            <input type="text" v-model="form.accessKey" />
          </div>
          <div class="form-group">
            <label><span class="required">*</span> {{ formType === 'cloudflare' ? 'API Token' : 'Secret key' }}</label>
            <input type="password" v-model="form.secretKey" />
            <span class="hint" v-if="formType === 'cloudflare'">请勿使用 Global API Key</span>
          </div>

          <div class="form-group" v-if="formType === 'cloudflare'">
            <label>EMAIL</label>
            <input type="text" v-model="form.email" />
          </div>

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
import { ref, onMounted, onUnmounted } from 'vue';

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

// 日志面板拖拽逻辑
const showLogs = ref(true);
const logPanelHeight = ref(200);
let isDragging = false;
let startY = 0;
let startHeight = 0;

const startDrag = (e) => {
  isDragging = true;
  startY = e.clientY;
  startHeight = logPanelHeight.value;
  document.addEventListener('mousemove', onDrag);
  document.addEventListener('mouseup', stopDrag);
  document.body.style.userSelect = 'none'; // 拖拽时防止选中文本
};

const onDrag = (e) => {
  if (!isDragging) return;
  const dy = startY - e.clientY;
  let newHeight = startHeight + dy;
  if (newHeight < 100) newHeight = 100;
  if (newHeight > window.innerHeight - 150) newHeight = window.innerHeight - 150;
  logPanelHeight.value = newHeight;
};

const stopDrag = () => {
  isDragging = false;
  document.removeEventListener('mousemove', onDrag);
  document.removeEventListener('mouseup', stopDrag);
  document.body.style.userSelect = '';
};

onUnmounted(() => {
  document.removeEventListener('mousemove', onDrag);
  document.removeEventListener('mouseup', stopDrag);
});

const addLog = (msg, level = 'info') => {
  const time = new Date().toLocaleTimeString('zh-CN', { hour12: false });
  logs.value.unshift({ time, msg, level });
};

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
    if (domains.value.length > 0 && domains.value[0].includes('报错')) {
      addLog(`拉取域名失败: ${domains.value[0]}`, 'error');
    } else {
      addLog(`拉取域名列表: 找到 ${domains.value.length} 条记录`, 'success');
    }
  }
};

const loadRecords = async (domain) => {
  currentDomain.value = domain;
  addLog(`准备加载域名 ${domain} 的解析记录...`, 'info');
  if (window.go && window.go.main && !domain.includes('暂无') && !domain.includes('报错')) {
    records.value = await window.go.main.DNSService.GetRecords(currentProvider.value, domain);
    if (records.value.length > 0 && records.value[0].Type === 'ERROR') {
      addLog(`加载记录失败: ${records.value[0].Content}`, 'error');
    } else {
      addLog(`记录加载成功，共 ${records.value.length} 条`, 'success');
    }
  }
};

const openSettings = async () => {
  if (window.go && window.go.main) {
    globalConfig.value = await window.go.main.DNSService.LoadConfig();
  }
  formType.value = currentProvider.value || 'huawei';
  syncFormToType();
  showSettings.value = true;
  addLog('打开 API 配置面板', 'info');
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
/* 强制全局文字颜色为深灰色，解决深色模式下白底白字问题 */
.app-container { display: flex; flex-direction: column; height: 100vh; background-color: #f0f2f5; font-family: 'Microsoft YaHei', sans-serif; color: #333 !important; }
.header-area { display: flex; justify-content: space-between; align-items: center; background: #fff; border-bottom: 1px solid #dcdfe6; padding-right: 20px; }
.provider-tabs { display: flex; }
.provider-tabs button { padding: 15px 25px; border: none; background: transparent; cursor: pointer; color: #606266; font-size: 15px; }
.provider-tabs button:hover { color: #1890ff; }
.provider-tabs button.active { background: #fff; color: #1890ff; border-bottom: 2px solid #1890ff; font-weight: bold; }
.settings-btn { padding: 8px 15px; cursor: pointer; border: 1px solid #dcdfe6; background: #fff; border-radius: 4px; color: #333; }

.main-content { display: flex; flex: 1; overflow: hidden; position: relative; }
.domain-list { width: 220px; background: #fff; border-right: 1px solid #e4e7ed; overflow-y: auto; color: #333; }
.domain-list ul { list-style: none; margin: 0; padding: 0; }
.domain-list li { padding: 15px 20px; cursor: pointer; border-bottom: 1px solid #ebeef5; font-size: 14px; color: #333; }
.domain-list li.active { background: #ecf5ff; color: #409eff; }
.domain-list li.warning { color: #f56c6c; font-size: 13px; }

.record-manager { flex: 1; padding: 25px; overflow-y: auto; color: #333; }
.record-manager h3 { color: #333; }
.record-table { width: 100%; border-collapse: collapse; margin-top: 15px; background: #fff; color: #333; }
.record-table th, .record-table td { border: 1px solid #ebeef5; padding: 12px; text-align: left; color: #333; }
.record-table th { background: #fafafa; font-weight: bold; color: #333; }
.content-cell { word-break: break-all; }

/* 拖拽式日志面板 */
.log-panel-container { display: flex; flex-direction: column; width: 100%; z-index: 10; background: #1e1e1e; }
.resizer { height: 10px; background: #dcdfe6; cursor: ns-resize; display: flex; justify-content: center; align-items: center; transition: background 0.2s; }
.resizer:hover { background: #c0c4cc; }
.resizer-line { width: 40px; height: 3px; background: #909399; border-radius: 2px; }
.log-panel { flex: 1; display: flex; flex-direction: column; overflow: hidden; }
.log-header { padding: 8px 15px; background: #2d2d2d; color: #ccc; font-size: 13px; font-weight: bold; border-bottom: 1px solid #1e1e1e; display: flex; justify-content: space-between; align-items: center; }
.hide-log-btn { background: none; border: 1px solid #555; color: #ccc; cursor: pointer; border-radius: 3px; padding: 2px 8px; font-size: 12px; }
.hide-log-btn:hover { background: #444; }
.log-content { flex: 1; padding: 10px 15px; overflow-y: auto; font-family: 'Consolas', monospace; font-size: 13px; }
.log-item { margin-bottom: 5px; line-height: 1.4; }
.log-time { color: #888; margin-right: 8px; }
.log-item.info { color: #d4d4d4; }
.log-item.success { color: #4CAF50; }
.log-item.error { color: #F44336; }

/* 悬浮打开日志按钮 */
.floating-log-btn { position: absolute; bottom: 15px; right: 20px; z-index: 5; padding: 8px 15px; border: none; background: #333; color: #fff; border-radius: 20px; cursor: pointer; box-shadow: 0 2px 10px rgba(0,0,0,0.2); font-size: 13px; }
.floating-log-btn:hover { background: #555; }

/* 弹窗及表单颜色修正 */
.modal-overlay { position: fixed; top: 0; left: 0; width: 100%; height: 100%; background: rgba(0,0,0,0.5); display: flex; justify-content: center; align-items: center; z-index: 100; }
.modal-content { background: #fff; width: 480px; border-radius: 4px; box-shadow: 0 4px 16px rgba(0,0,0,0.2); display: flex; flex-direction: column; }
.modal-header { padding: 15px 20px; border-bottom: 1px solid #ebeef5; display: flex; justify-content: space-between; align-items: center; }
.modal-header h3 { margin: 0; font-size: 16px; color: #333; font-weight: normal; }
.close-btn { background: none; border: none; font-size: 20px; color: #909399; cursor: pointer; }
.modal-body { padding: 20px 30px; }
.form-group { margin-bottom: 20px; }
.form-group label { display: block; margin-bottom: 8px; font-size: 14px; color: #333; }
.required { color: #f56c6c; margin-right: 4px; }
.form-group input, .form-group select { width: 100%; padding: 10px; border: 1px solid #dcdfe6; border-radius: 4px; box-sizing: border-box; outline: none; font-size: 14px; color: #333; background: #fff;}
.form-group input:focus, .form-group select:focus { border-color: #409eff; }
.hint { display: block; font-size: 12px; color: #909399; margin-top: 5px; }
.modal-footer { padding: 15px 20px; border-top: 1px solid #ebeef5; display: flex; justify-content: flex-end; gap: 10px; }
.btn-cancel { padding: 9px 20px; cursor: pointer; border: 1px solid #dcdfe6; background: #fff; border-radius: 4px; color: #333; }
.btn-confirm { padding: 9px 20px; cursor: pointer; border: none; background: #165dff; border-radius: 4px; color: #fff; }
</style>
