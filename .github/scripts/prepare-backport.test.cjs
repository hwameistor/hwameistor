'use strict';

const assert = require('node:assert/strict');
const { test } = require('node:test');
const { targetsForLabels, isBackportOf, prepareBackport } = require('./prepare-backport.cjs');

const repository = 'hwameistor/hwameistor';
const target = 'release/v1.1';
const labels = (...names) => names.map(name => ({ name }));
const manualBody = 'Backport of the scheduler fix from #1864 to `release/v1.1`, addressing #1863.';
function candidate(overrides = {}) {
  return {
    number: 1866,
    html_url: 'https://github.com/hwameistor/hwameistor/pull/1866',
    state: 'open', merged_at: null, body: manualBody,
    base: { ref: target },
    head: { ref: 'codex/backport-1864-v1.1', repo: { full_name: 'contributor/hwameistor' } },
    ...overrides,
  };
}

function harness({ pull = {}, candidates = [], listError } = {}) {
  const calls = [];
  const outputs = {};
  const summary = [];
  const github = {
    rest: { pulls: {
      get: async args => {
        calls.push(['get', args]);
        return { data: {
          merged: true, base: { ref: 'main', repo: { full_name: repository } },
          labels: labels('cherry-pick-v1.1'), ...pull,
        } };
      },
      list: () => {},
    } },
    paginate: async (_, args) => {
      calls.push(['list', args]);
      if (listError) throw listError;
      return candidates;
    },
  };
  const core = {
    info: () => {},
    setOutput: (key, value) => { outputs[key] = value; },
    summary: {
      addHeading(value) { summary.push(value); return this; },
      addRaw(value) { summary.push(value); return this; },
      async write() {},
    },
  };
  return {
    calls, outputs, summary,
    run: (pullNumber = '1864') => prepareBackport({
      github, core, pullNumber,
      context: {
        repo: { owner: 'hwameistor', repo: 'hwameistor' },
        payload: { pull_request: { labels: labels('cherry-pick-v1') } },
      },
    }),
  };
}

test('routes supported labels and deduplicates the legacy v1.0 alias', () => {
  assert.deepEqual(targetsForLabels(labels('cherry-pick-v1.1', 'cherry-pick-v1', 'cherry-pick-v1.0')),
    [target, 'release/v1.0.0']);
  assert.deepEqual(targetsForLabels(labels('cherry-pick', 'cherry-pick-v1.2', 'backport main')), []);
});

test('recognizes manual backports without mistaking other references for the source', () => {
  assert.equal(isBackportOf(candidate(), 1864, target, repository), true);
  assert.equal(isBackportOf(candidate(), 1863, target, repository), false);
  assert.equal(isBackportOf(candidate({ body: 'Backport of #18640' }), 1864, target, repository), false);
  assert.equal(isBackportOf(candidate({ body: 'Fixes #1864' }), 1864, target, repository), false);
});

test('recognizes the generated marker even if the source fork was deleted', () => {
  const pull = candidate({
    body: '<!-- hwameistor-backport:1864:release/v1.1 -->',
    head: { ref: 'deleted-branch', repo: null },
  });
  assert.equal(isBackportOf(pull, 1864, target, repository), true);
  assert.equal(isBackportOf({ ...pull, body: null }, 1864, target, repository), false);
});

test('recognizes the generated branch only in the current repository', () => {
  const head = { ref: 'backport-1864-to-release/v1.1', repo: { full_name: repository } };
  assert.equal(isBackportOf(candidate({ body: null, head }), 1864, target, repository), true);
  head.repo.full_name = 'contributor/hwameistor';
  assert.equal(isBackportOf(candidate({ body: null, head }), 1864, target, repository), false);
});

test('skips open and merged PRs but permits closed unmerged PRs and other targets', () => {
  assert.equal(isBackportOf(candidate({ state: 'closed', merged_at: '2026-01-01' }), 1864, target, repository), true);
  assert.equal(isBackportOf(candidate({ state: 'closed' }), 1864, target, repository), false);
  assert.equal(isBackportOf(candidate(), 1864, 'release/v1.0.0', repository), false);
});

test('reads current API labels instead of stale event labels', async () => {
  const h = harness();
  await h.run();
  assert.deepEqual(h.outputs, { pull_number: '1864', targets: target });
  assert.deepEqual(h.calls[1][1], {
    owner: 'hwameistor', repo: 'hwameistor', base: target, state: 'all', per_page: 100,
  });
});

test('skips only targets with an existing open or merged backport and links to it', async () => {
  const h = harness({
    pull: { labels: labels('cherry-pick-v1', 'cherry-pick-v1.0', 'cherry-pick-v1.1') },
    candidates: [candidate()],
  });
  await h.run();
  assert.equal(h.outputs.targets, 'release/v1.0.0');
  assert.equal(h.calls.filter(([method]) => method === 'list').length, 2);
  assert.match(h.summary.join('\n'), /\[#1866\]\(https:\/\/github.com\/hwameistor\/hwameistor\/pull\/1866\)/);
});

test('does no backport work if labels were removed while the event was queued', async () => {
  const h = harness({ pull: { labels: [] } });
  await h.run();
  assert.equal(h.outputs.targets, '');
  assert.equal(h.calls.length, 1);
});

test('produces no targets when all backports already exist', async () => {
  const h = harness({ candidates: [candidate()] });
  await h.run();
  assert.equal(h.outputs.targets, '');
});

test('allows retrying a closed unmerged backport', async () => {
  const h = harness({ candidates: [candidate({ state: 'closed' })] });
  await h.run();
  assert.equal(h.outputs.targets, target);
});

test('rejects unmerged, non-main and foreign-repository source PRs before target discovery', async () => {
  for (const pull of [
    { merged: false },
    { base: { ref: target, repo: { full_name: repository } } },
    { base: { ref: 'main', repo: { full_name: 'other/repo' } } },
  ]) {
    const h = harness({ pull });
    await assert.rejects(h.run(), /Only merged PRs targeting main/);
    assert.equal(h.calls.length, 1);
    assert.deepEqual(h.outputs, {});
  }
});

test('rejects malformed or unsafe PR numbers before making API calls', async () => {
  for (const number of ['', '0', '-1', '01', '1.5', '1e3', ' 1864 ', '1864; echo nope', '9007199254740992']) {
    const h = harness();
    await assert.rejects(h.run(number), /positive integer/);
    assert.equal(h.calls.length, 0);
  }
});

test('fails closed if duplicate discovery fails', async () => {
  const h = harness({ listError: new Error('API unavailable') });
  await assert.rejects(h.run(), /API unavailable/);
  assert.deepEqual(h.outputs, {});
});
