const fs = require('fs');
const path = require('path');
const { execSync } = require('child_process');

console.log('🚀 Starting PgWeb build process...');

// 清理函数
function cleanUp() {
  console.log('🧹 Cleaning up previous builds...');
  try {
    // 强制结束可能正在运行的进程
    try {
      execSync('taskkill /f /im PgWeb.exe 2>nul', { stdio: 'ignore' });
    } catch (e) {
      // 忽略错误，进程可能不存在
    }
    
    try {
      execSync('taskkill /f /im pgweb.exe 2>nul', { stdio: 'ignore' });
    } catch (e) {
      // 忽略错误，进程可能不存在
    }
    
    // 等待进程完全关闭
    setTimeout(() => {
      if (fs.existsSync('dist')) {
        try {
          execSync('rmdir /s /q dist 2>nul', { stdio: 'pipe' });
        } catch (e) {
          console.log('⚠️  Warning: Could not fully clean dist directory');
        }
      }
    }, 1000);
    
  } catch (error) {
    console.log('⚠️  Warning during cleanup:', error.message);
  }
}

// 检查静态文件
function checkStaticFiles() {
  console.log('📂 Checking static files...');
  
  const requiredFiles = [
    'static/index.html',
    'static/js/jquery.js',
    'static/js/app.js',
    'static/css/app.css'
  ];
  
  for (const file of requiredFiles) {
    if (!fs.existsSync(file)) {
      throw new Error(`Required static file missing: ${file}`);
    }
  }
  
  console.log('✅ Static files check passed');
}

// 构建Go后端
function buildGoBackend() {
  console.log('🔨 Building Go backend...');
  
  // 确保目录存在
  if (!fs.existsSync('dist')) {
    fs.mkdirSync('dist', { recursive: true });
  }
  
  try {
    // 构建Windows版本
    const buildCmd = 'CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -ldflags "-s -w" -o dist/pgweb.exe main.go';
    console.log('Executing:', buildCmd);
    execSync(buildCmd, { stdio: 'inherit' });
    
    // 验证构建结果
    if (!fs.existsSync('dist/pgweb.exe')) {
      throw new Error('Go backend build failed - executable not found');
    }
    
    const stats = fs.statSync('dist/pgweb.exe');
    console.log(`✅ Go backend built successfully (${Math.round(stats.size / 1024 / 1024)}MB)`);
    
  } catch (error) {
    console.error('❌ Go backend build failed:', error.message);
    process.exit(1);
  }
}

// 构建Electron应用
function buildElectronApp(platform = 'win') {
  console.log(`📦 Building Electron app for ${platform}...`);
  
  try {
    let buildCommand;
    switch (platform) {
      case 'win':
        buildCommand = 'npx electron-builder --win';
        break;
      case 'linux':
        buildCommand = 'npx electron-builder --linux deb';
        break;
      case 'all':
        buildCommand = 'npx electron-builder --win --linux deb';
        break;
      default:
        throw new Error(`Unsupported platform: ${platform}`);
    }
    
    console.log('Executing:', buildCommand);
    execSync(buildCommand, { stdio: 'inherit' });
    
    console.log('✅ Electron app built successfully');
    
  } catch (error) {
    console.error('❌ Electron build failed:', error.message);
    process.exit(1);
  }
}

// 验证构建结果
function verifyBuild() {
  console.log('🔍 Verifying build results...');
  
  const expectedFiles = [
    'dist/win-unpacked/PgWeb.exe',
    'dist/win-unpacked/resources/pgweb.exe'
  ];
  
  let allFilesExist = true;
  for (const file of expectedFiles) {
    if (fs.existsSync(file)) {
      const stats = fs.statSync(file);
      console.log(`✅ ${file} (${Math.round(stats.size / 1024 / 1024)}MB)`);
    } else {
      console.log(`❌ Missing: ${file}`);
      allFilesExist = false;
    }
  }
  
  if (allFilesExist) {
    console.log('🎉 Build verification passed!');
    console.log('📍 Executable location: dist/win-unpacked/PgWeb.exe');
  } else {
    console.log('⚠️  Build verification failed - some files are missing');
  }
}

// 创建图标文件（如果不存在）
function ensureIcons() {
  console.log('🎨 Ensuring icon files exist...');
  
  if (!fs.existsSync('static/img')) {
    fs.mkdirSync('static/img', { recursive: true });
  }
  
  // 如果图标不存在，创建一个简单的SVG图标
  if (!fs.existsSync('static/img/icon.ico') && !fs.existsSync('static/img/icon.png')) {
    console.log('⚠️  No icon files found, creating placeholder...');
    
    // 创建一个简单的SVG图标
    const svgIcon = `<?xml version="1.0" encoding="UTF-8"?>
<svg width="256" height="256" viewBox="0 0 256 256" xmlns="http://www.w3.org/2000/svg">
  <rect width="256" height="256" fill="#2563eb" rx="32"/>
  <text x="128" y="140" font-family="Arial, sans-serif" font-size="72" font-weight="bold" 
        text-anchor="middle" fill="white">PG</text>
  <text x="128" y="200" font-family="Arial, sans-serif" font-size="32" 
        text-anchor="middle" fill="#60a5fa">Web</text>
</svg>`;
    
    fs.writeFileSync('static/img/icon.svg', svgIcon);
    console.log('📝 Created placeholder icon');
  }
}

// 主函数
async function main() {
  try {
    const platform = process.argv[2] || 'win';
    
    console.log(`Building for platform: ${platform}`);
    console.log('='.repeat(50));
    
    // 执行构建步骤
    cleanUp();
    
    // 等待清理完成
    await new Promise(resolve => setTimeout(resolve, 2000));
    
    checkStaticFiles();
    ensureIcons();
    buildGoBackend();
    buildElectronApp(platform);
    verifyBuild();
    
    console.log('='.repeat(50));
    console.log('🎉 Build completed successfully!');
    console.log('');
    console.log('💡 Usage:');
    console.log('  - Run: .\\dist\\win-unpacked\\PgWeb.exe');
    console.log('  - Install: .\\dist\\PgWeb Setup 1.0.0.exe');
    console.log('');
    
  } catch (error) {
    console.error('❌ Build failed:', error.message);
    process.exit(1);
  }
}

// 处理未捕获的异常
process.on('uncaughtException', (error) => {
  console.error('❌ Uncaught exception:', error);
  process.exit(1);
});

process.on('unhandledRejection', (reason, promise) => {
  console.error('❌ Unhandled rejection at:', promise, 'reason:', reason);
  process.exit(1);
});

// 运行主函数
main(); 