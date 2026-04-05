#!/usr/bin/env node
const fs = require('fs');
const path = require('path');
const os = require('os');
const https = require('https');

// 1. Fixed Map to match your release.yml naming exactly
const platformMap = {
  'darwin-x64': 'darwin-amd64',
  'darwin-arm64': 'darwin-arm64',
  'linux-x64': 'linux-amd64',
  'linux-arm64': 'linux-arm64',
  'win32-x64': 'windows-amd64.exe',
  'win32-arm64': 'windows-arm64.exe' // Added for safety
};

const platform = `${os.platform()}-${os.arch()}`;
const suffix = platformMap[platform];

if (!suffix) {
  console.error(`❌ Unsupported platform: ${platform}`);
  process.exit(1);
}

const installDir = path.join(__dirname, '..', 'bin');
const fullBinaryName = `dataset-cli-${suffix}`;
const destPath = path.join(installDir, os.platform() === 'win32' ? 'dataset-cli.exe' : 'dataset-cli');

// 2. Clear previous failed attempts
if (!fs.existsSync(installDir)) {
  fs.mkdirSync(installDir, { recursive: true });
}

// 3. Get version from package.json
const packageJson = require(path.join(__dirname, '..', 'package.json'));
const version = packageJson.version;
const repo = 'darshan192004/cli-project';
const downloadUrl = `https://github.com/${repo}/releases/download/v${version}/${fullBinaryName}`;

console.log(`☁️ Downloading dataset-cli v${version} for ${platform}...`);

function download(url) {
  https.get(url, (res) => {
    // Handle Redirects (Important for GitHub)
    if (res.statusCode === 301 || res.statusCode === 302) {
      return download(res.headers.location);
    }

    if (res.statusCode !== 200) {
      console.error(`❌ Download failed! HTTP Status: ${res.statusCode}`);
      console.error(`URL attempted: ${url}`);
      process.exit(1);
    }

    const file = fs.createWriteStream(destPath);
    res.pipe(file);

    file.on('finish', () => {
      file.close();
      // 4. Only chmod on Non-Windows systems
      if (os.platform() !== 'win32') {
        fs.chmodSync(destPath, 0o755);
      }
      console.log(`✅ Success! Installed to ${destPath}`);
    });
  }).on('error', (err) => {
    console.error(`❌ Request Error: ${err.message}`);
    process.exit(1);
  });
}

download(downloadUrl);