<template>
  <div class="app-container">
    <!-- 第一行：顶部工具菜单栏 -->
    <div class="menu-bar">
      <span class="menu-item">文件(F)</span>
      <span class="menu-item">API配置(C)</span>
      <span class="menu-item">帮助(H)</span>
    </div>

    <!-- 第二行：四大厂商 Tabs -->
    <div class="provider-tabs">
      <button v-for="tab in providers" :key="tab.id" 
              :class="{ active: currentProvider === tab.id }"
              @click="selectProvider(tab.id)">
        {{ tab.name }}
      </button>
    </div>

    <!-- 主内容区：分左右两栏结构，符合现代 Windows 设置风格 -->
    <div class="main-content">
      <!-- 左侧：该厂商下的所有域名列表 -->
      <div class="domain-list">
        <ul>
          <li v-for="domain in domains" :key="domain" 
              :class="{ active: currentDomain === domain }"
              @click="loadRecords(domain)">
            {{ domain }}
          </li>
        </ul>
      </div>

      <!-- 右侧：选中域名的 DNS 记录管理 (增删改查) -->
      <div class="record-manager" v-if="currentDomain">
        <div class="action-bar">
          <h3>{{ currentDomain }} 解析记录</h3>
          <button @click="showAddModal = true">＋ 添加记录</button>
        </div>
        
        <table class="record-table">
          <thead>
            <tr>
              <th>类型</th>
              <th>主机记录</th>
              <th>记录值</th>
              <th>操作</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="record in records" :key="record.id">
              <td>{{ record.type }}</td>
              <td>{{ record.name }}</td>
              <td>{{ record.content }}</td>
              <td>
                <button @click="editRecord(record)">修改</button>
                <button class="danger" @click="deleteRecord(record.id)">删除</button>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref } from 'vue';
// 导入 Wails 自动生成的 Go 后端绑定函数
import { GetDomains, GetRecords, DeleteRecord } from '../wailsjs/go/main/DNSService';

const providers = [
  { id: 'cloudflare', name: 'Cloudflare' },
  { id: 'huawei', name: '华为云' },
  { id: 'aliyun', name: '阿里云' },
  { id: 'tencent', name: '腾讯云' }
];

const currentProvider = ref('cloudflare');
const currentDomain = ref('');
const domains = ref([]);
const records = ref([]);

// 切换厂商并拉取域名
const selectProvider = async (pid) => {
  currentProvider.value = pid;
  currentDomain.value = '';
  domains.value = await GetDomains(pid);
};

// 点击域名拉取解析记录
const loadRecords = async (domain) => {
  currentDomain.value = domain;
  records.value = await GetRecords(currentProvider.value, domain);
};

// 删除记录逻辑
const deleteRecord = async (id) => {
  if (confirm('确定要彻底删除该解析记录吗？')) {
    await DeleteRecord(currentProvider.value, currentDomain.value, id);
    loadRecords(currentDomain.value); // 刷新列表
  }
};
</script>

<style scoped>
/* 样式部分建议引入 Element Plus 或 Naive UI 组件库以获得更原生的 Windows UI 体验 */
.app-container { display: flex; flex-direction: column; height: 100vh; background-color: #f3f3f3; font-family: 'Segoe UI', Tahoma, Geneva, Verdana, sans-serif; }
.menu-bar { background: #fff; padding: 5px 15px; border-bottom: 1px solid #ddd; font-size: 14px; }
.menu-item { margin-right: 15px; cursor: pointer; }
.provider-tabs { display: flex; background: #eaeaea; }
.provider-tabs button { padding: 10px 20px; border: none; background: transparent; cursor: pointer; }
.provider-tabs button.active { background: #fff; border-top: 3px solid #0078D7; }
.main-content { display: flex; flex: 1; overflow: hidden; }
.domain-list { width: 250px; background: #fff; border-right: 1px solid #ddd; overflow-y: auto; }
.domain-list li { padding: 12px 15px; cursor: pointer; border-bottom: 1px solid #f0f0f0; }
.domain-list li.active { background: #e5f1fb; }
.record-manager { flex: 1; padding: 20px; background: #fff; overflow-y: auto; }
.record-table { width: 100%; border-collapse: collapse; margin-top: 15px; }
.record-table th, .record-table td { border: 1px solid #ddd; padding: 10px; text-align: left; }
.danger { color: red; }
</style>
