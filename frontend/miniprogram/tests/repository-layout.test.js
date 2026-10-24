const test = require('node:test');
const assert = require('node:assert/strict');
const fs = require('node:fs');
const path = require('node:path');
const { execFileSync } = require('node:child_process');

const root = execFileSync('git', ['rev-parse', '--show-toplevel'], {
  cwd: __dirname,
  encoding: 'utf8',
}).trim();

test('repository has independent frontend and backend roots', () => {
  for (const relativePath of [
    'frontend/package.json',
    'frontend/project.config.json',
    'frontend/miniprogram/app.json',
    'backend/go.mod',
    'backend/cmd',
    'backend/internal',
    'backend/migrations',
  ]) {
    assert.ok(fs.existsSync(path.join(root, relativePath)), `${relativePath} is missing`);
  }
});

test('obsolete application roots are gone', () => {
  for (const relativePath of ['miniprogram', 'server', 'package.json', 'project.config.json']) {
    assert.equal(fs.existsSync(path.join(root, relativePath)), false, `${relativePath} still exists`);
  }
});

test('tracked repository root contains only delivery files', () => {
  const trackedRootEntries = execFileSync('git', ['ls-files'], {
    cwd: root,
    encoding: 'utf8',
  })
    .trim()
    .split('\n')
    .filter(Boolean)
    .map((file) => file.split('/')[0]);

  assert.deepEqual(
    [...new Set(trackedRootEntries)].sort(),
    ['.gitignore', 'README.md', 'backend', 'docs', 'frontend'],
  );
});
