// 预加载脚本
const { contextBridge } = require('electron');

// 暴露安全的 API 到渲染进程
contextBridge.exposeInMainWorld('electron', {
  // 这里可以添加其他需要的 API
}); 