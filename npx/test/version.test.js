const test = require('node:test');
const assert = require('node:assert/strict');
const { extractSemver, compareVersions } = require('../lib/version');

test('extracts plain semver', () => {
  assert.equal(extractSemver('v0.2.3\n'), '0.2.3');
});

test('extracts semver from prefixed output', () => {
  assert.equal(extractSemver('agent-notify version v1.4.0\n'), '1.4.0');
});

test('returns null when output has no semver', () => {
  assert.equal(extractSemver('development build\n'), null);
});

test('compares versions correctly', () => {
  assert.equal(compareVersions('0.2.3', '0.2.3'), 0);
  assert.equal(compareVersions('0.2.2', '0.2.3'), -1);
  assert.equal(compareVersions('0.3.0', '0.2.9'), 1);
});
