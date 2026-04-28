const test = require('node:test');
const assert = require('node:assert/strict');
const fs = require('node:fs');
const os = require('node:os');
const path = require('node:path');
const { main } = require('../bin/agent-notify');

test('downloads when installed binary is missing', async (t) => {
  const calls = [];
  const root = fs.mkdtempSync(path.join(os.tmpdir(), 'agent-notify-launcher-'));
  t.after(() => fs.rmSync(root, { recursive: true, force: true }));
  const binaryPath = path.join(root, 'agent-notify');

  const exitCode = await main(['doctor'], {
    getDesiredVersion: () => '0.2.3',
    getPlatformTarget: () => ({ goos: 'linux', goarch: 'amd64', ext: '' }),
    getInstalledBinaryPath: () => binaryPath,
    getInstalledVersion: () => null,
    downloadAndInstall: async () => {
      calls.push('download');
      fs.writeFileSync(binaryPath, 'binary');
      return binaryPath;
    },
    runBinary: async (targetPath, args) => {
      calls.push(['run', targetPath, args]);
      return 0;
    },
    pathExists: (value) => fs.existsSync(value),
    warn: () => {},
  });

  assert.equal(exitCode, 0);
  assert.deepEqual(calls, ['download', ['run', binaryPath, ['doctor']]]);
});

test('reuses installed binary when version is current', async () => {
  const calls = [];

  const exitCode = await main(['doctor'], {
    getDesiredVersion: () => '0.2.3',
    getPlatformTarget: () => ({ goos: 'linux', goarch: 'amd64', ext: '' }),
    getInstalledBinaryPath: () => '/tmp/agent-notify',
    getInstalledVersion: () => '0.2.3',
    downloadAndInstall: async () => {
      throw new Error('should not download');
    },
    runBinary: async (targetPath, args) => {
      calls.push(['run', targetPath, args]);
      return 0;
    },
    pathExists: () => true,
    compareVersions: (left, right) => (left === right ? 0 : -1),
    warn: () => {},
  });

  assert.equal(exitCode, 0);
  assert.deepEqual(calls, [['run', '/tmp/agent-notify', ['doctor']]]);
});

test('reuses installed binary when local version is newer', async () => {
  const calls = [];

  const exitCode = await main(['doctor'], {
    getDesiredVersion: () => '0.2.3',
    getPlatformTarget: () => ({ goos: 'linux', goarch: 'amd64', ext: '' }),
    getInstalledBinaryPath: () => '/tmp/agent-notify',
    getInstalledVersion: () => '0.2.4',
    downloadAndInstall: async () => {
      throw new Error('should not download');
    },
    runBinary: async (targetPath, args) => {
      calls.push(['run', targetPath, args]);
      return 0;
    },
    pathExists: () => true,
    compareVersions: () => 1,
    warn: () => {},
  });

  assert.equal(exitCode, 0);
  assert.deepEqual(calls, [['run', '/tmp/agent-notify', ['doctor']]]);
});

test('falls back to old binary when download fails but installed binary exists', async () => {
  const warnings = [];
  const calls = [];

  const exitCode = await main(['doctor'], {
    getDesiredVersion: () => '0.2.3',
    getPlatformTarget: () => ({ goos: 'linux', goarch: 'amd64', ext: '' }),
    getInstalledBinaryPath: () => '/tmp/agent-notify',
    getInstalledVersion: () => '0.2.2',
    compareVersions: () => -1,
    downloadAndInstall: async () => {
      throw new Error('network down');
    },
    runBinary: async (targetPath, args) => {
      calls.push(['run', targetPath, args]);
      return 0;
    },
    pathExists: () => true,
    warn: (message) => warnings.push(message),
  });

  assert.equal(exitCode, 0);
  assert.equal(warnings.length, 1);
  assert.match(warnings[0], /network down/);
  assert.deepEqual(calls, [['run', '/tmp/agent-notify', ['doctor']]]);
});

test('fails when download fails and no installed binary exists', async () => {
  await assert.rejects(
    main(['doctor'], {
      getDesiredVersion: () => '0.2.3',
      getPlatformTarget: () => ({ goos: 'linux', goarch: 'amd64', ext: '' }),
      getInstalledBinaryPath: () => '/tmp/agent-notify',
      getInstalledVersion: () => null,
      compareVersions: () => -1,
      downloadAndInstall: async () => {
        throw new Error('network down');
      },
      runBinary: async () => 0,
      pathExists: () => false,
      warn: () => {},
    }),
    /network down/,
  );
});

test('fails when installed binary is unreadable and update also fails', async () => {
  const warnings = [];

  await assert.rejects(
    main(['doctor'], {
      getDesiredVersion: () => '0.2.3',
      getPlatformTarget: () => ({ goos: 'linux', goarch: 'amd64', ext: '' }),
      getInstalledBinaryPath: () => '/tmp/agent-notify',
      getInstalledVersion: () => null,
      compareVersions: () => -1,
      downloadAndInstall: async () => {
        throw new Error('network down');
      },
      runBinary: async () => 0,
      pathExists: () => true,
      warn: (message) => warnings.push(message),
    }),
    /network down/,
  );

  assert.deepEqual(warnings, []);
});
