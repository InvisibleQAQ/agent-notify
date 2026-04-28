function extractSemver(output) {
  const match = String(output).match(/v?(\d+\.\d+\.\d+)/);
  return match ? match[1] : null;
}

function compareVersions(left, right) {
  const l = left.split('.').map(Number);
  const r = right.split('.').map(Number);

  for (let i = 0; i < 3; i += 1) {
    if (l[i] < r[i]) return -1;
    if (l[i] > r[i]) return 1;
  }

  return 0;
}

module.exports = {
  extractSemver,
  compareVersions,
};
