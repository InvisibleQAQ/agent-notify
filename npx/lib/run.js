const { spawn } = require('node:child_process');

function runBinary(binaryPath, args) {
  return new Promise((resolve, reject) => {
    const child = spawn(binaryPath, args, {
      stdio: 'inherit',
      env: { ...process.env },
    });

    child.on('error', reject);
    child.on('close', (code) => resolve(code ?? 1));
  });
}

module.exports = {
  runBinary,
};
