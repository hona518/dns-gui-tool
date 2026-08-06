<template>
  <div class="app-container" @mousemove="onDrag" @mouseup="stopDrag" @mouseleave="stopDrag">
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
          <div class="manager-header">
            <h3>{{ currentDomain }} 解析记录</h3>
            <button class="btn-primary" @click="openRecordModal()">＋ 添加记录</button>
          </div>
          
          <table class="record-table">
            <thead>
              <tr>
                <th width="80">类型</th>
                <th width="120">主机记录</th>
                <th width="120">线路 (分流)</th>
                <th>记录值</th>
                <th width="80">TTL</th>
                <th width="120">操作</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="record in records" :key="record.id">
                <td>{{ record.type }}</td>
                <td>{{ record.name }}</td>
                <td>
                  <!-- 高亮非默认线路，视觉更直观 -->
                  <span :class="['badge', record.line === 'default' ? '' : 'highlight']">
                    {{ record.line === 'default' ? '默认' : record.line }}
                  </span>
                </td>
                <td class="content-cell">{{ record.content }}</td>
                <td>{{ record.ttl }}</td>
                <td>
                  <button class="action-btn" @click="openRecordModal(record)">修改</button>
                  <button class="action-btn danger" @click="deleteRecord(record)">删除</button>
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>
    </div>

    <!-- 日志面板 (自动滚动到底部) -->
    <div class="log-panel-container" v-show="showLogs" :style="{ height: logPanelHeight + 'px' }">
      <div class="resizer" @mousedown="startDrag">
        <div class="resizer-line"></div>
      </div>
      <div class="log-panel">
        <div class="log-header">
          <span>实时运行日志</span>
          <button class="hide-log-btn" @click="showLogs = false">▼ 隐藏</button>
        </div>
        <!-- 绑定 logContentRef 以实现自动滚动 -->
        <div class="log-content" ref="logContentRef">
          <div v-for="(log, index) in logs" :key="index" :class="['log-item', log.level]">
            <span class="log-time">[{{ log.time }}]</span> {{ log.msg }}
          </div>
        </div>
      </div>
    </div>

    <button class="floating-log-btn" v-show="!showLogs" @click="showLogs = true">
      ▲ 显示日志
    </button>

    <!-- 记录增改弹窗 -->
    <div class="modal-overlay" v-if="showRecordModal">
      <div class="modal-content">
        <div class="modal-header">
          <h3>{{ editingRecord.id ? '修改解析记录' : '添加解析记录' }}</h3>
          <button class="close-btn" @click="showRecordModal = false">×</button>
        </div>
        
        <div class="modal-body">
          <div class="form-row">
            <div class="form-group half">
              <label><span class="required">*</span> 记录类型</label>
              <select v-model="editingRecord.type">
                <option value="A">A</option>
                <option value="CNAME">CNAME</option>
                <option value="AAAA">AAAA</option>
                <option value="TXT">TXT</option>
              </select>
            </div>
            <div class="form-group half">
              <label><span class="required">*</span> 主机记录 (如 @, www)</label>
              <input type="text" v-model="editingRecord.name" />
            </div>
          </div>

          <div class="form-group">
            <label><span class="required">*</span> 记录值</label>
            <input type="text" v-model="editingRecord.content" placeholder="输入IP或域名" />
          </div>

          <div class="form-row">
            <div class="form-group half">
  <label>线路分流 (Line)</label>
  <select v-model="editingRecord.line">
    <option value="default">全网默认 (default)</option>
    <option value="telecom">电信 (telecom)</option>
    <option value="unicom">联通 (unicom)</option>
    <option value="mobile">移动 (mobile)</option>
    <option value="tietong">铁通 (tietong)</option>
    <option value="NA">北美洲 (NA)</option>
  </select>
</div>
            <div class="form-group half">
              <label>TTL (秒)</label>
              <input type="number" v-model="editingRecord.ttl" />
            </div>
          </div>
        </div>
        
        <div class="modal-footer">
          <button class="btn-cancel" @click="showRecordModal = false">取消</button>
          <button class="btn-primary" @click="saveRecord">确认提交</button>
        </div>
      </div>
    </div>

    <!-- API 密钥配置弹窗 -->
    <div class="modal-overlay" v-if="showSettings">
      <div class="modal-content">
        <div class="modal-header">
          <h3>API 密钥配置</h3>
          <button class="close-btn" @click="showSettings = false">×</button>
        </div>
        <div class="modal-body">
          <div class="form-group">
            <label><span class="required">*</span> Access Key</label>
            <input type="text" v-model="form.accessKey" />
          </div>
          <div class="form-group">
            <label><span class="required">*</span> Secret Key</label>
            <input type="password" v-model="form.secretKey" />
          </div>
          <div class="form-group">
            <label>Region</label>
            <input type="text" v-model="form.region" placeholder="cn-north-1" />
          </div>
        </div>
        <div class="modal-footer">
          <button class="btn-cancel" @click="showSettings = false">取消</button>
          <button class="btn-primary" @click="saveSettings">保存</button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted, nextTick } from 'vue';

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

// 日志拖拽与自动滚动逻辑
const showLogs = ref(true);
const logPanelHeight = ref(200);
const logContentRef = ref(null);
let isDragging = false;
let startY = 0;
let startHeight = 0;

const startDrag = (e) => {
  isDragging = true;
  startY = e.clientY;
  startHeight = logPanelHeight.value;
  document.body.style.userSelect = 'none'; 
};
const onDrag = (e) => {
  if (!isDragging) return;
  const dy = startY - e.clientY;
  let newHeight = startHeight + dy;
  if (newHeight < 100) newHeight = 100;
  if (newHeight > window.innerHeight - 200) newHeight = window.innerHeight - 200;
  logPanelHeight.value = newHeight;
};
const stopDrag = () => {
  if (isDragging) {
    isDragging = false;
    document.body.style.userSelect = '';
  }
};

const addLog = (msg, level = 'info') => {
  const time = new Date().toLocaleTimeString('zh-CN', { hour12: false });
  // 改为 push：新日志追加到底部
  logs.value.push({ time, msg, level });
  // 渲染完成后自动滚动到底部
  nextTick(() => {
    if (logContentRef.value) {
      logContentRef.value.scrollTop = logContentRef.value.scrollHeight;
    }
  });
};

// CRUD 操作逻辑
const showRecordModal = ref(false);
const editingRecord = ref({});

const openRecordModal = (record = null) => {
  if (record) {
    editingRecord.value = { ...record };
  } else {
    editingRecord.value = { type: 'A', name: '', content: '', line: 'default', ttl: 300, zoneId: '' };
  }
  showRecordModal.value = true;
};

const saveRecord = async () => {
  if (!editingRecord.value.name || !editingRecord.value.content) {
    addLog('请填写主机记录与记录值', 'error');
    return;
  }
  
  showRecordModal.value = false;
  
  if (editingRecord.value.id) {
    addLog(`正在提交修改记录 [${editingRecord.value.name}]...`, 'info');
    try {
      await window.go.main.DNSService.UpdateRecord(currentProvider.value, currentDomain.value, editingRecord.value);
      addLog(`记录 [${editingRecord.value.name}] 修改成功`, 'success');
    } catch (err) {
      addLog(`修改失败: ${err}`, 'error');
    }
  } else {
    addLog(`正在提交添加记录 [${editingRecord.value.name}]...`, 'info');
    try {
      await window.go.main.DNSService.AddRecord(currentProvider.value, currentDomain.value, editingRecord.value);
      addLog(`记录 [${editingRecord.value.name}] 添加成功`, 'success');
    } catch (err) {
      addLog(`添加失败: ${err}`, 'error');
    }
  }
  loadRecords(currentDomain.value);
};

const deleteRecord = async (record) => {
  if (confirm(`确定要彻底删除 [${record.name}] 的解析记录吗？此操作不可逆！`)) {
    addLog(`正在提交删除请求...`, 'info');
    try {
      await window.go.main.DNSService.DeleteRecord(currentProvider.value, record.zoneId, record.id);
      addLog(`记录 [${record.name}] 删除成功`, 'success');
      loadRecords(currentDomain.value);
    } catch (err) {
      addLog(`删除失败: ${err}`, 'error');
    }
  }
};

// API 配置存储
const showSettings = ref(false);
const globalConfig = ref({});
const form = ref({ name: '', accessKey: '', secretKey: '', region: 'cn-north-1', email: '' });

const openSettings = async () => {
  if (window.go && window.go.main) {
    globalConfig.value = await window.go.main.DNSService.LoadConfig();
    const data = globalConfig.value['huawei'] || {};
    form.value = { accessKey: data.accessKey || '', secretKey: data.secretKey || '', region: data.region || 'cn-north-1' };
  }
  showSettings.value = true;
};

const saveSettings = async () => {
  if (!globalConfig.value['huawei']) globalConfig.value['huawei'] = {};
  globalConfig.value['huawei'] = { ...form.value };
  
  if (window.go && window.go.main) {
    await window.go.main.DNSService.SaveConfig(globalConfig.value);
    showSettings.value = false;
    addLog(`API 凭证本地保存成功`, 'success');
    selectProvider(currentProvider.value); 
  }
};

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
    if (records.value.length > 0 && records.value[0].type === 'ERROR') {
      addLog(`加载记录失败: ${records.value[0].content}`, 'error');
    } else {
      addLog(`记录加载成功，共 ${records.value.length} 条`, 'success');
    }
  }
};

onMounted(() => {
  addLog('终端界面初始化完毕，等待 API 连通测试', 'info');
  selectProvider('huawei');
});
</script>

<style scoped>
.app-container { display: flex; flex-direction: column; height: 100vh; background-color: #f5f7fa; font-family: 'Microsoft YaHei', sans-serif; color: #303133 !important; }
.header-area { display: flex; justify-content: space-between; align-items: center; background: #fff; border-bottom: 1px solid #e4e7ed; padding-right: 20px; }
.provider-tabs { display: flex; }
.provider-tabs button { padding: 15px 25px; border: none; background: transparent; cursor: pointer; color: #606266; font-size: 15px; }
.provider-tabs button:hover { color: #409eff; }
.provider-tabs button.active { background: #fff; color: #409eff; border-bottom: 2px solid #409eff; font-weight: bold; }
.settings-btn { padding: 8px 15px; cursor: pointer; border: 1px solid #dcdfe6; background: #fff; border-radius: 4px; color: #606266; font-size: 13px; }

.main-content { display: flex; flex: 1; overflow: hidden; position: relative; }
.domain-list { width: 240px; background: #fff; border-right: 1px solid #e4e7ed; overflow-y: auto; color: #303133; }
.domain-list ul { list-style: none; margin: 0; padding: 0; }
.domain-list li { padding: 15px 20px; cursor: pointer; border-bottom: 1px solid #ebeef5; font-size: 14px; color: #303133; }
.domain-list li.active { background: #ecf5ff; color: #409eff; border-right: 2px solid #409eff; }
.domain-list li.warning { color: #f56c6c; font-size: 13px; }

.record-manager { flex: 1; padding: 25px; overflow-y: auto; color: #303133; }
.manager-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 15px; }
.manager-header h3 { margin: 0; font-size: 18px; color: #303133; }
.record-table { width: 100%; border-collapse: collapse; background: #fff; box-shadow: 0 1px 4px rgba(0,0,0,0.05); }
.record-table th, .record-table td { border: 1px solid #ebeef5; padding: 12px 15px; text-align: left; color: #606266; font-size: 13px; }
.record-table th { background: #f5f7fa; font-weight: bold; color: #303133; }
.content-cell { word-break: break-all; }
.badge { background: #f4f4f5; color: #909399; padding: 3px 8px; border-radius: 4px; font-size: 12px; border: 1px solid #e9e9eb; }
.badge.highlight { background: #ecf5ff; color: #409eff; border-color: #d9ecff; font-weight: bold; }
.action-btn { padding: 5px 10px; cursor: pointer; border: 1px solid #dcdfe6; background: #fff; border-radius: 4px; color: #606266; margin-right: 5px; font-size: 12px; }
.action-btn:hover { color: #409eff; border-color: #c6e2ff; background: #ecf5ff; }
.action-btn.danger:hover { color: #f56c6c; border-color: #fbc4c4; background: #fef0f0; }

/* 拖拽式日志面板修正 */
.log-panel-container { display: flex; flex-direction: column; width: 100%; z-index: 10; background: #1e1e1e; position: relative; }
.resizer { height: 12px; background: #dcdfe6; cursor: ns-resize; display: flex; justify-content: center; align-items: center; border-top: 1px solid #c0c4cc; }
.resizer:hover { background: #c0c4cc; }
.resizer-line { width: 40px; height: 3px; background: #909399; border-radius: 2px; }
.log-panel { flex: 1; display: flex; flex-direction: column; overflow: hidden; }
.log-header { padding: 8px 15px; background: #2d2d2d; color: #ccc; font-size: 13px; border-bottom: 1px solid #1e1e1e; display: flex; justify-content: space-between; align-items: center; }
.hide-log-btn { background: none; border: 1px solid #555; color: #ccc; cursor: pointer; border-radius: 3px; padding: 2px 8px; font-size: 12px; }
.hide-log-btn:hover { background: #444; }
.log-content { flex: 1; padding: 10px 15px; overflow-y: auto; font-family: 'Consolas', monospace; font-size: 13px; }
.log-item { margin-bottom: 5px; line-height: 1.4; }
.log-time { color: #888; margin-right: 8px; }
.log-item.info { color: #d4d4d4; }
.log-item.success { color: #67c23a; }
.log-item.error { color: #f56c6c; }
.floating-log-btn { position: absolute; bottom: 15px; right: 20px; z-index: 5; padding: 8px 15px; border: none; background: #303133; color: #fff; border-radius: 20px; cursor: pointer; box-shadow: 0 2px 10px rgba(0,0,0,0.2); font-size: 13px; }

/* 弹窗通用样式 */
.modal-overlay { position: fixed; top: 0; left: 0; width: 100%; height: 100%; background: rgba(0,0,0,0.5); display: flex; justify-content: center; align-items: center; z-index: 100; }
.modal-content { background: #fff; width: 500px; border-radius: 4px; box-shadow: 0 4px 16px rgba(0,0,0,0.2); display: flex; flex-direction: column; }
.modal-header { padding: 15px 20px; border-bottom: 1px solid #ebeef5; display: flex; justify-content: space-between; align-items: center; }
.modal-header h3 { margin: 0; font-size: 16px; color: #303133; font-weight: bold; }
.close-btn { background: none; border: none; font-size: 20px; color: #909399; cursor: pointer; }
.modal-body { padding: 20px 25px; }
.form-row { display: flex; gap: 15px; }
.form-group { margin-bottom: 20px; }
.form-group.half { flex: 1; }
.form-group label { display: block; margin-bottom: 8px; font-size: 13px; color: #606266; }
.required { color: #f56c6c; margin-right: 4px; }
.form-group input, .form-group select { width: 100%; padding: 10px; border: 1px solid #dcdfe6; border-radius: 4px; box-sizing: border-box; outline: none; font-size: 13px; color: #303133; background: #fff; }
.form-group input:focus, .form-group select:focus { border-color: #409eff; }
.modal-footer { padding: 15px 20px; border-top: 1px solid #ebeef5; display: flex; justify-content: flex-end; gap: 10px; }
.btn-cancel { padding: 9px 20px; cursor: pointer; border: 1px solid #dcdfe6; background: #fff; border-radius: 4px; color: #606266; font-size: 13px; }
.btn-primary { padding: 9px 20px; cursor: pointer; border: none; background: #409eff; border-radius: 4px; color: #fff; font-size: 13px; }
.btn-primary:hover { background: #66b1ff; }
</style>
