const { app, BrowserWindow } = require('electron');
const path = require('path');
const { spawn } = require('child_process');
const fs = require('fs');
const http = require('http');

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
    const logsDir = path.join(__dirname, 'logs');
    if (!fs.existsSync(logsDir)) {
      fs.mkdirSync(logsDir);
    }
    
    const backendPath = process.platform === 'win32'
      ? path.join(__dirname, 'go-backend', 'pgweb.exe')
      : path.join(__dirname, 'go-backend', 'pgweb');
    
    console.log('Backend binary path:', backendPath);
    
    if (!fs.existsSync(backendPath)) {
      console.error('Backend binary not found at:', backendPath);
      console.error('Please compile the backend first:');
      console.error('go build -o electron/go-backend/pgweb.exe main.go');
      reject(new Error('Backend binary not found'));
      return;
    }
    
    console.log(`Starting backend on port ${backendPort}...`);
    
    backendProcess = spawn(backendPath, [
      '--bind=127.0.0.1',
      `--port=${backendPort}`
    ], {
      cwd: __dirname,
      detached: false
    });
    
    backendProcess.stdout.on('data', data => {
      console.log(`Backend stdout: ${data}`);
    });
    
    backendProcess.stderr.on('data', data => {
      console.error(`Backend stderr: ${data}`);
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
    }, 2000); // 给后端进程 2 秒启动时间
  });
}

function createWindow() {
  const win = new BrowserWindow({
    width: 1200,
    height: 800,
    webPreferences: {
      preload: path.join(__dirname, 'preload.js'),
      nodeIntegration: false,
      contextIsolation: true,
      webSecurity: false,
      allowRunningInsecureContent: true
    }
  });
  win.webContents.openDevTools();
  win.loadFile(path.join(__dirname, '..', 'static', 'index.html'));
}

app.whenReady().then(async () => {
  try {
    console.log('Starting backend service...');
    await startBackend();
    console.log('Backend is ready, creating window...');
    createWindow();
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
  if (process.platform !== 'darwin') app.quit();
}); 