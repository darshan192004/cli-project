#!/usr/bin/env node
const path = require('path');
const os = require('os');

const platformMap = {
  'darwin-x64': 'darwin-amd64',
  'darwin-arm64': 'darwin-arm64',
  'linux-x64': 'linux-amd64',
  'linux-arm64': 'linux-arm64',
  'win32-x64': 'windows-amd64.exe'
};

const platform = `${os.platform()}-${os.arch()}`;
const binaryName = platformMap[platform] || platform;

const binaryPath = path.join(__dirname, 'bin', binaryName);

const { spawn } = require('child_process');
const child = spawn(binaryPath, process.argv.slice(2), {
  stdio: 'inherit',
  cwd: process.cwd()
});

child.on('exit', (code) => {
  process.exit(code);
});
