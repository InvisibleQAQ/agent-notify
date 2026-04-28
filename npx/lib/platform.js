function getPlatformTarget({ platform = process.platform, arch = process.arch } = {}) {
  const goosMap = {
    darwin: 'darwin',
    linux: 'linux',
    win32: 'windows',
  };

  const goarchMap = {
    x64: 'amd64',
    arm64: 'arm64',
  };

  const goos = goosMap[platform];
  const goarch = goarchMap[arch];

  if (!goos || !goarch) {
    throw new Error(`unsupported platform: ${platform}/${arch}`);
  }

  return {
    goos,
    goarch,
    ext: goos === 'windows' ? '.exe' : '',
  };
}

module.exports = {
  getPlatformTarget,
};
