const { app, BrowserWindow } = require('electron');
const path = require('path');
const { spawn } = require('child_process');
const fs = require('fs');
const http = require('http');
const os = require('os');

let backendProcess;
let backendPort = 3000;

// 检测后端服务是否已启动并能响应请求
function waitForBackend(port, maxAttempts = 30) {
  return new Promise((resolve, reject) => {
    let attempts = 0;
    
    function checkBackend() {
      let backendPingUrl =  `http://127.0.0.1:${port}/api/ping`
      attempts++;
      console.log(`Checking backend [ ${backendPingUrl} ] readiness... (${attempts}/${maxAttempts})`);
      
      const req = http.get(backendPingUrl, (res) => {
        console.log('Backend ping ',backendPingUrl, 'is ready!');
        resolve();
      });
      
      req.on('error', (err) => {
        if (attempts >= maxAttempts) {
          console.error('Backend failed to start within timeout');
          reject(new Error('Backend startup timeout'));
          return;
        }
        
        console.log(`Backend not ready yet, retrying in 1s... (${err.code})`);
        setTimeout(checkBackend, 1000);
      });
      
      req.setTimeout(2000, () => {
        req.destroy();
        if (attempts >= maxAttempts) {
          reject(new Error('Backend startup timeout'));
          return;
        }
        setTimeout(checkBackend, 1000);
      });
    }
    
    checkBackend();
  });
}

function startBackend() {
  return new Promise((resolve, reject) => {
    // 根据是否打包选择不同的后端路径和日志目录
    let backendPath;
    let logsDir;
    let workingDir;
    const isDev = !app.isPackaged;
    
    if (isDev) {
      // 开发环境：从go-backend目录
      backendPath = process.platform === 'win32'
        ? path.join(__dirname, 'go-backend', 'pgweb.exe')
        : path.join(__dirname, 'go-backend', 'pgweb');
      logsDir = path.join(__dirname, 'logs');
      workingDir = __dirname;
    } else {
      // 打包环境：从resources目录
      backendPath = process.platform === 'win32'
        ? path.join(process.resourcesPath, 'pgweb.exe')
        : path.join(process.resourcesPath, 'pgweb');
      
      // 使用用户目录作为日志目录，避免权限问题
      const userDataPath = app.getPath('userData');
      logsDir = path.join(userDataPath, 'logs');
      workingDir = userDataPath;
    }
    
    console.log('Environment:', isDev ? 'development' : 'production');
    console.log('Backend binary path:', backendPath);
    console.log('Logs directory:', logsDir);
    console.log('Working directory:', workingDir);
    console.log('Resources path:', process.resourcesPath);
    console.log('User data path:', app.getPath('userData'));
    
    // 创建日志目录
    try {
      if (!fs.existsSync(logsDir)) {
        fs.mkdirSync(logsDir, { recursive: true });
        console.log('Created logs directory:', logsDir);
      }
    } catch (error) {
      console.warn('Warning: Could not create logs directory:', error.message);
      // 不要因为日志目录创建失败而终止，继续执行
    }
    
    if (!fs.existsSync(backendPath)) {
      const errorMsg = `Backend binary not found at: ${backendPath}`;
      console.error(errorMsg);
      if (isDev) {
        console.error('Please compile the backend first:');
        console.error('go build -o electron/go-backend/pgweb.exe main.go');
      }
      reject(new Error('Backend binary not found'));
      return;
    }
    
    console.log(`Starting backend on port ${backendPort}...`);
    
    backendProcess = spawn(backendPath, [
      '--bind', '127.0.0.1',
      '--listen', `${backendPort}`,
      '--bookmarks-dir', './data/bookmarks',
      '--skip-open'
    ], {
      cwd: workingDir,
      detached: false,
      stdio: ['ignore', 'pipe', 'pipe']
    });
    
    backendProcess.stdout.on('data', data => {
      console.log(`Backend stdout: ${data.toString().trim()}`);
    });
    
    backendProcess.stderr.on('data', data => {
      console.error(`Backend stderr: ${data.toString().trim()}`);
    });
    
    backendProcess.on('close', code => {
      console.log(`Backend process closed with code ${code}`);
    });
    
    backendProcess.on('exit', (code, signal) => {
      console.log(`Backend process exited with code ${code}, signal ${signal}`);
    });
    
    backendProcess.on('error', (err) => {
      console.error('Failed to start backend process:', err);
      reject(err);
    });
    
    // 等待后端启动并能响应请求
    setTimeout(async () => {
      try {
        await waitForBackend(backendPort);
        resolve();
      } catch (err) {
        reject(err);
      }
    }, 2000);
  });
}

function createWindow() {
  const win = new BrowserWindow({
    width: 1200,
    height: 800,
    show: true,
    webPreferences: {
      preload: path.join(__dirname, 'preload.js'),
      nodeIntegration: false,
      contextIsolation: true,
      webSecurity: false,
      allowRunningInsecureContent: true
    }
  });
  
  // 只在开发环境打开开发者工具
  if (!app.isPackaged) {
    win.webContents.openDevTools();
  }
  
  // 加载主页面
  const indexPath = path.join(__dirname, '..', 'static', 'index.html');
  console.log('Loading index.html from:', indexPath);
  win.loadFile(indexPath);
  
  // 确保窗口显示在前面
  win.focus();
  win.show();
}

app.whenReady().then(async () => {
  try {
    console.log('App is ready, starting backend service...');
    await startBackend();
    console.log('Backend is ready, creating window...');
    createWindow();
    console.log('Application started successfully!');
  } catch (err) {
    console.error('Failed to start application:', err);
    app.quit();
  }
});

app.on('window-all-closed', () => {
  if (backendProcess) {
    console.log('Killing backend process...');
    backendProcess.kill();
  }
  if (process.platform !== 'darwin') {
    app.quit();
  }
});

app.on('activate', () => {
  if (BrowserWindow.getAllWindows().length === 0) {
    createWindow();
  }
}); 