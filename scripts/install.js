#!/usr/bin/env node
const fs = require('fs');
const path = require('path');
const os = require('os');
const https = require('https');
const crypto = require('crypto');

const platformMap = {
  'darwin-x64': 'darwin-amd64',
  'darwin-arm64': 'darwin-arm64',
  'linux-x64': 'linux-amd64',
  'linux-arm64': 'linux-arm64',
  'win32-x64': 'windows-amd64.exe',
  'win32-arm64': 'windows-arm64.exe'
};

const platform = `${os.platform()}-${os.arch()}`;
const suffix = platformMap[platform];

if (!suffix) {
  console.error(`❌ Unsupported platform: ${platform}`);
  console.error('Supported platforms: darwin-x64, darwin-arm64, linux-x64, linux-arm64, win32-x64, win32-arm64');
  process.exit(1);
}

const installDir = path.join(__dirname, '..', 'bin');
const fullBinaryName = `dataset-cli-${suffix}`;
const destPath = path.join(installDir, os.platform() === 'win32' ? 'dataset-cli.exe' : 'dataset-cli');

if (!fs.existsSync(installDir)) {
  fs.mkdirSync(installDir, { recursive: true });
}

const packageJson = require(path.join(__dirname, '..', 'package.json'));
const version = packageJson.version;
const repo = 'darshan192004/cli-project';
const downloadUrl = `https://github.com/${repo}/releases/download/v${version}/${fullBinaryName}`;
const checksumUrl = `https://github.com/${repo}/releases/download/v${version}/${fullBinaryName}.sha256`;

const MAX_RETRIES = 3;
const RETRY_DELAY = 1000;

function delay(ms) {
  return new Promise(resolve => setTimeout(resolve, ms));
}

async function downloadFile(url, isChecksum = false) {
  let lastError;
  
  for (let attempt = 1; attempt <= MAX_RETRIES; attempt++) {
    try {
      const result = await new Promise((resolve, reject) => {
        https.get(url, (res) => {
          if (res.statusCode === 301 || res.statusCode === 302) {
            return resolve(downloadFile(res.headers.location, isChecksum));
          }
          
          if (res.statusCode !== 200) {
            reject(new Error(`HTTP ${res.statusCode}`));
            return;
          }
          
          const chunks = [];
          res.on('data', chunk => chunks.push(chunk));
          res.on('end', () => {
            const data = Buffer.concat(chunks);
            if (isChecksum) {
              resolve(data.toString('utf8').trim());
            } else {
              resolve(data);
            }
          });
          res.on('error', reject);
        }).on('error', reject);
      });
      return result;
    } catch (err) {
      lastError = err;
      if (attempt < MAX_RETRIES) {
        const waitTime = RETRY_DELAY * attempt;
        console.log(`⚠️  Attempt ${attempt} failed, retrying in ${waitTime}ms...`);
        await delay(waitTime);
      }
    }
  }
  
  throw lastError;
}

function verifyChecksum(filePath, expectedChecksum) {
  const fileBuffer = fs.readFileSync(filePath);
  const hash = crypto.createHash('sha256').update(fileBuffer).digest('hex');
  
  const parts = expectedChecksum.split(/\s+/);
  const checksumParts = parts[0];
  
  if (hash !== checksumParts) {
    console.error('❌ Checksum verification failed!');
    console.error(`Expected: ${checksumParts}`);
    console.error(`Got: ${hash}`);
    return false;
  }
  return true;
}

async function install() {
  console.log(`☁️  Downloading dataset-cli v${version} for ${platform}...`);
  
  let checksum;
  try {
    checksum = await downloadFile(checksumUrl, true);
  } catch (err) {
    console.warn(`⚠️  Warning: Could not download checksum file: ${err.message}`);
    console.warn('Proceeding without checksum verification...');
    checksum = null;
  }
  
  try {
    const binaryData = await downloadFile(downloadUrl);
    
    fs.writeFileSync(destPath, binaryData);
    
    if (checksum && !verifyChecksum(destPath, checksum)) {
      fs.unlinkSync(destPath);
      console.error('❌ Installation failed: checksum mismatch');
      process.exit(1);
    }
    
    if (os.platform() !== 'win32') {
      try {
        fs.chmodSync(destPath, 0o755);
      } catch (chmodErr) {
        console.warn(`⚠️  Warning: Could not set executable permissions: ${chmodErr.message}`);
      }
    }
    
    const stat = fs.statSync(destPath);
    if (stat.size === 0) {
      fs.unlinkSync(destPath);
      console.error('❌ Installation failed: downloaded file is empty');
      process.exit(1);
    }
    
    console.log(`✅ Success! Installed to ${destPath}`);
    console.log(`   Size: ${(stat.size / 1024 / 1024).toFixed(2)} MB`);
    
  } catch (err) {
    console.error(`❌ Download failed: ${err.message}`);
    console.error(`URL attempted: ${downloadUrl}`);
    process.exit(1);
  }
}

install();
