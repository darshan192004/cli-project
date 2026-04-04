#!/usr/bin/env node
const fs = require('fs');
const path = require('path');
const os = require('os');
const https = require('https');
const { execSync } = require('child_process');

const platformMap = {
  'darwin-x64': 'darwin-amd64',
  'darwin-arm64': 'darwin-arm64',
  'linux-x64': 'linux-amd64',
  'linux-arm64': 'linux-arm64',
  'win32-x64': 'windows-amd64.exe'
};

const platform = `${os.platform()}-${os.arch()}`;
const binaryName = platformMap[platform];

if (!binaryName) {
  console.error(`Unsupported platform: ${platform}`);
  process.exit(1);
}

const installDir = path.join(__dirname, 'bin');
const binaryPath = path.join(installDir, binaryName);

if (fs.existsSync(binaryPath)) {
  return;
}

const version = require(path.join(__dirname, '..', 'package.json')).version;
const repo = 'darshan192004/cli-project';
const downloadUrl = `https://github.com/${repo}/releases/download/v${version}/dataset-cli-${binaryName}`;

console.log(`Downloading dataset-cli v${version} for ${platform}...`);

const file = fs.createWriteStream(binaryPath);

https.get(downloadUrl, (response) => {
  if (response.statusCode === 302 || response.statusCode === 301) {
    https.get(response.headers.location, (redirectResponse) => {
      redirectResponse.pipe(file);
      file.on('finish', () => {
        file.close();
        fs.chmodSync(binaryPath, 0o755);
        console.log(`Installed dataset-cli v${version}`);
      });
    });
  } else {
    response.pipe(file);
    file.on('finish', () => {
      file.close();
      fs.chmodSync(binaryPath, 0o755);
      console.log(`Installed dataset-cli v${version}`);
    });
  }
}).on('error', (err) => {
  fs.unlink(binaryPath, () => {});
  console.error(`Failed to download: ${err.message}`);
  process.exit(1);
});
