// 预加载脚本
const { contextBridge } = require('electron');

// 暴露安全的 API 到渲染进程
contextBridge.exposeInMainWorld('electron', {
  getConfig: () => {
    return {
      HOST: global.CONFIG.HOST,
      PORT: global.CONFIG.PORT,
      API_BASE_URL: global.CONFIG.API_BASE_URL,
      LOCAL_HOSTS: ['localhost', '127.0.0.1']
    };
  }
}); 