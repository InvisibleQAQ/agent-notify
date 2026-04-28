const { REPO_OWNER, REPO_NAME } = require('./constants');

function buildTag(version) {
  return version.startsWith('v') ? version : `v${version}`;
}

function buildAssetName(version, target) {
  const tag = buildTag(version);
  return `agent-notify-${tag}-${target.goos}-${target.goarch}.tar.gz`;
}

function buildDownloadUrl(version, assetName) {
  const tag = buildTag(version);
  return `https://github.com/${REPO_OWNER}/${REPO_NAME}/releases/download/${tag}/${assetName}`;
}

module.exports = {
  buildTag,
  buildAssetName,
  buildDownloadUrl,
};
