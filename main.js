const { app, BrowserWindow } = require('electron');
const path = require('path');

// 全局配置 - 作为唯一配置源
global.CONFIG = {
  HOST: '127.0.0.1',
  PORT: 3000,
  get API_BASE_URL() {
    return `http://${this.HOST}:${this.PORT}`;
  }
};

function createWindow() {
  const win = new BrowserWindow({
    width: 1200,
    height: 800,
    webPreferences: {
      nodeIntegration: false,
      contextIsolation: true,
      preload: path.join(__dirname, 'preload.js')
    }
  });

  // 加载本地文件
  win.loadFile('static/index.html');
}

app.whenReady().then(() => {
  createWindow();

  app.on('activate', () => {
    if (BrowserWindow.getAllWindows().length === 0) {
      createWindow();
    }
  });
});

app.on('window-all-closed', () => {
  if (process.platform !== 'darwin') {
    app.quit();
  }
}); 