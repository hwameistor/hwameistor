'use strict';

const assert = require('node:assert/strict');
const { test } = require('node:test');
const fs = require('node:fs');
const os = require('node:os');
const path = require('node:path');
const { execFileSync } = require('node:child_process');
const { classifyChanges, changedFiles, checkResults } = require('./pr-checks.cjs');

test('dependencies, build inputs and source changes run code tests', () => {
  for (const file of ['pkg/a.go', 'cmd/main.go', 'test/pr-test.sh', 'build/a/Dockerfile',
    'helm/a.yaml', 'deploy/a.yaml', 'vendor/a/b', 'hack/a.sh', 'third_party/a', 'tool.go',
    'go.mod', 'go.sum', 'go.work', 'go.work.sum', 'Makefile', 'Makefile.variables', '.golangci.yml']) {
    assert.equal(classifyChanges([file]).code, true, file);
  }
});

test('release metadata, workflow and documentation changes use lightweight checks', () => {
  assert.deepEqual(classifyChanges(['VERSION', 'release-assistant.yml', 'README.md',
    '.release-assistant/manifest.json', '.github/scripts/release.cjs',
    '.github/workflows/pr.yml', '.github/workflows/release.yaml']), {
    code: false, workflows: ['.github/workflows/pr.yml', '.github/workflows/release.yaml'],
  });
});

test('large PRs retain code changes after thousands of documentation files', () => {
  const files = Array.from({ length: 3100 }, (_, i) => `docs/${i}.md`);
  assert.equal(classifyChanges([...files, 'go.sum']).code, true);
});

test('Git diff includes deleted and renamed source paths, excluding unrelated base changes', t => {
  const cwd = fs.mkdtempSync(path.join(os.tmpdir(), 'hwameistor-pr-checks-'));
  t.after(() => fs.rmSync(cwd, { recursive: true, force: true }));
  const git = (...args) => execFileSync('git', args, { cwd, encoding: 'utf8' }).trim();
  git('init', '-q');
  git('config', 'user.name', 'Test');
  git('config', 'user.email', 'test@example.com');
  git('config', 'commit.gpgsign', 'false');
  fs.mkdirSync(path.join(cwd, 'pkg'));
  fs.writeFileSync(path.join(cwd, 'pkg/old.go'), 'package old\n');
  git('add', '.');
  git('commit', '-qm', 'base');
  const ancestor = git('rev-parse', 'HEAD');
  fs.renameSync(path.join(cwd, 'pkg/old.go'), path.join(cwd, 'notes.md'));
  git('add', '-A');
  git('commit', '-qm', 'move source to docs');
  const head = git('rev-parse', 'HEAD');
  git('checkout', '--quiet', '--detach', ancestor);
  fs.writeFileSync(path.join(cwd, 'unrelated.go'), 'package other\n');
  git('add', '.');
  git('commit', '-qm', 'base advances');
  const files = changedFiles(git('rev-parse', 'HEAD'), head, cwd);
  assert.deepEqual(files.sort(), ['notes.md', 'pkg/old.go']);
  assert.equal(classifyChanges(files).code, true);
});

test('invalid or unavailable diff inputs fail instead of skipping tests', () => {
  assert.throws(() => changedFiles('', 'HEAD'), /full commit SHAs/);
  assert.throws(() => changedFiles('0'.repeat(40), '1'.repeat(40)));
});

function results(code = 'true') {
  return {
    changes: { result: 'success', outputs: { code } },
    'workflow-checks': { result: 'success' },
    'unit-test': { result: code === 'true' ? 'success' : 'skipped' },
    'pr-test': { result: code === 'true' ? 'success' : 'skipped' },
  };
}

test('gate accepts passing code checks and intentional non-code skips', () => {
  checkResults(results());
  checkResults(results('false'));
});

test('gate rejects failed, cancelled, missing and unexpectedly skipped required jobs', () => {
  for (const job of Object.keys(results())) {
    for (const result of ['failure', 'cancelled', 'skipped', undefined]) {
      const jobs = results();
      jobs[job].result = result;
      assert.throws(() => checkResults(jobs), undefined, `${job}: ${result}`);
    }
  }
  const jobs = results('false');
  jobs['workflow-checks'].result = 'failure';
  assert.throws(() => checkResults(jobs));
  assert.throws(() => checkResults(results('')));
});
