'use strict';

const fs = require('node:fs');
const path = require('node:path');
const { execFileSync } = require('node:child_process');

function classifyChanges(files) {
  return {
    code: files.some(file => /^(pkg|cmd|test|build|helm|deploy|vendor|hack|third_party)\//.test(file)
      || /\.go$/.test(file)
      || /^(go\.(mod|sum|work|work\.sum)|Makefile(?:\..*)?|\.golangci\.yml)$/.test(file)),
    workflows: files.filter(file => /^\.github\/workflows\/[^/]+\.ya?ml$/.test(file)),
  };
}

function changedFiles(base, head, cwd = process.cwd()) {
  if (![base, head].every(sha => /^[0-9a-f]{40}$/.test(sha || ''))) {
    throw new Error('Both PR base and head must be full commit SHAs');
  }
  // Use Git instead of the PR-files API, which truncates large PRs at 3,000 files.
  // Disabling rename detection includes both the old and new paths in coverage.
  return execFileSync('git', ['diff', '--name-only', '--no-renames', '-z', `${base}...${head}`], {
    cwd, encoding: 'utf8', maxBuffer: 16 * 1024 * 1024, stdio: ['ignore', 'pipe', 'pipe'],
  }).split('\0').filter(Boolean);
}

function checkResults(jobs) {
  for (const name of ['changes', 'workflow-checks']) {
    if (jobs[name]?.result !== 'success') throw new Error(`${name} did not succeed`);
  }
  const code = jobs.changes.outputs?.code;
  if (!['true', 'false'].includes(code)) throw new Error('Missing change classification');
  for (const name of ['unit-test', 'pr-test']) {
    const expected = code === 'true' ? 'success' : 'skipped';
    if (jobs[name]?.result !== expected) throw new Error(`${name}: expected ${expected}, got ${jobs[name]?.result}`);
  }
}

function main(command) {
  if (command === 'changes') {
    const changes = classifyChanges(changedFiles(process.env.PR_BASE_SHA, process.env.PR_HEAD_SHA));
    fs.appendFileSync(process.env.GITHUB_OUTPUT,
      `code=${changes.code}\nworkflows=${JSON.stringify(changes.workflows)}\n`);
    console.log(JSON.stringify(changes, null, 2));
  } else if (command === 'lint') {
    const files = JSON.parse(process.env.CHANGED_WORKFLOWS).filter(file => fs.existsSync(file));
    if (files.length) {
      execFileSync(path.join(process.env.RUNNER_TEMP, 'actionlint'),
        ['-shellcheck=', '-pyflakes=', ...files], { stdio: 'inherit' });
    }
  } else if (command === 'gate') {
    checkResults(JSON.parse(process.env.PR_JOB_RESULTS));
    console.log('All applicable PR checks passed');
  } else {
    throw new Error(`Unknown command: ${command}`);
  }
}

if (require.main === module) main(process.argv[2]);
module.exports = { classifyChanges, changedFiles, checkResults };
