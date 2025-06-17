const { contextBridge } = require('electron');
contextBridge.exposeInMainWorld('pgwebConfig', {
  apiBase: 'http://localhost:3000'
}); 