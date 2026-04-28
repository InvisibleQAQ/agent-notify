const fs = require('node:fs');
const path = require('node:path');
const os = require('node:os');
const tar = require('tar');

function isUnsafeArchivePath(entryPath) {
  return path.isAbsolute(entryPath) || entryPath.split('/').includes('..');
}

async function installFromArchive({ archivePath, installDir, binaryNameInArchive, finalBinaryName }) {
  fs.mkdirSync(installDir, { recursive: true });

  const extractDir = fs.mkdtempSync(path.join(os.tmpdir(), 'agent-notify-extract-'));
  const finalPath = path.join(installDir, finalBinaryName);
  const tempFinalPath = `${finalPath}.tmp`;

  try {
    const entries = [];
    await tar.t({
      file: archivePath,
      gzip: true,
      onReadEntry: (entry) => entries.push({ path: entry.path, type: entry.type }),
    });

    const binaryEntry = entries.find((entry) => entry.path === binaryNameInArchive);
    if (!binaryEntry) {
      throw new Error(`binary not found in archive: ${binaryNameInArchive}`);
    }

    if (binaryEntry.type !== 'File' || entries.some((entry) => isUnsafeArchivePath(entry.path))) {
      throw new Error('unsafe archive contents');
    }

    await tar.x({
      file: archivePath,
      cwd: extractDir,
      gzip: true,
      filter: (entryPath) => entryPath === binaryNameInArchive,
    });

    const extractedPath = path.join(extractDir, binaryNameInArchive);
    if (!fs.existsSync(extractedPath)) {
      throw new Error(`binary not found in archive: ${binaryNameInArchive}`);
    }

    fs.copyFileSync(extractedPath, tempFinalPath);
    if (process.platform !== 'win32') {
      fs.chmodSync(tempFinalPath, 0o755);
    }
    fs.renameSync(tempFinalPath, finalPath);

    return finalPath;
  } finally {
    fs.rmSync(tempFinalPath, { force: true });
    fs.rmSync(extractDir, { recursive: true, force: true });
  }
}

module.exports = {
  installFromArchive,
};
