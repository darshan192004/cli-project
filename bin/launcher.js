#!/usr/bin/env node
const { spawn } = require('child_process');
const path = require('path');
const os = require('os');
const fs = require('fs');

// 1. Determine the extension based on the OS
const ext = os.platform() === 'win32' ? '.exe' : '';
// 2. Look for the binary in the same 'bin' folder where launcher.js lives
const binaryPath = path.join(__dirname, 'dataset-cli' + ext);

// 3. Safety check: Does the binary actually exist?
if (!fs.existsSync(binaryPath)) {
  console.error(`❌ Error: dataset-cli binary not found at ${binaryPath}`);
  console.error(`Please try reinstalling: npm install -g dataset-cli@latest`);
  process.exit(1);
}

// 4. Run the Go binary and pass through all arguments
const child = spawn(binaryPath, process.argv.slice(2), {
  stdio: 'inherit',
  cwd: process.cwd(),
  shell: true // Added for better compatibility with Windows shells
});

child.on('exit', (code) => {
  process.exit(code || 0);
});

child.on('error', (err) => {
  console.error(`❌ Failed to start dataset-cli: ${err.message}`);
  process.exit(1);
});