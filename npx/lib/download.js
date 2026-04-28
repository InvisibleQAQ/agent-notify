const fs = require('node:fs');
const https = require('node:https');

function removeFileIfExists(filePath, callback) {
  fs.rm(filePath, { force: true }, () => callback());
}

function downloadToFile(url, destinationPath, client = https) {
  return new Promise((resolve, reject) => {
    const file = fs.createWriteStream(destinationPath);

    const request = client.get(url, {
      timeout: 300000, // 300 seconds
    }, (response) => {
      if (response.statusCode >= 300 && response.statusCode < 400 && response.headers.location) {
        const redirectUrl = new URL(response.headers.location, url).toString();
        file.close(() => {
          removeFileIfExists(destinationPath, () => {
            downloadToFile(redirectUrl, destinationPath, client).then(resolve, reject);
          });
        });
        return;
      }

      if (response.statusCode !== 200) {
        file.close(() => {
          removeFileIfExists(destinationPath, () => {
            reject(new Error(`download failed: ${response.statusCode} ${url}`));
          });
        });
        return;
      }

      response.pipe(file);
      file.on('finish', () => file.close(() => resolve(destinationPath)));
    });

    request.on('error', (err) => {
      file.close(() => {
        removeFileIfExists(destinationPath, () => reject(err));
      });
    });

    request.on('timeout', () => {
      request.destroy();
      file.close(() => {
        removeFileIfExists(destinationPath, () => {
          reject(new Error(`download timeout after 300s: ${url}`));
        });
      });
    });
  });
}

module.exports = {
  downloadToFile,
};
