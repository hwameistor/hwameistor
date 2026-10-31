'use strict';

const assert = require('node:assert/strict');
const { test } = require('node:test');
const fs = require('node:fs');
const os = require('node:os');
const path = require('node:path');
const { readVersion, compareVersions, prepareRelease, publishRelease, assertUnreleased, tagCommit } = require('./release.cjs');

const SHA = 'a'.repeat(40);
const BEFORE = 'b'.repeat(40);
const context = { sha: SHA, ref: 'refs/heads/release/v1.1', repo: { owner: 'hwameistor', repo: 'hwameistor' }, payload: { before: BEFORE } };
const version = 'v1.1.3';
const missing = () => { throw Object.assign(new Error('Not found'), { status: 404 }); };

function fixture(options = {}) {
  const state = { tag: options.tag || null, release: options.release || null, assets: [], events: [], failUpload: false };
  const outputs = {};
  const core = { setOutput: (key, value) => { outputs[key] = value; }, info() {}, summary: { addRaw() { return this; }, async write() {} } };
  const github = { rest: {
    git: {
      async getRef() { return state.tag ? { data: { object: { type: 'commit', sha: state.tag } } } : missing(); },
      async createRef(args) { state.events.push('tag'); state.tag = args.sha; return { data: {} }; },
    },
    repos: {
      async getReleaseByTag() { return state.release ? { data: state.release } : missing(); },
      async getLatestRelease() { return options.latest ? { data: { tag_name: options.latest } } : missing(); },
      async createRelease(args) { state.events.push('draft'); state.release = { ...args, id: 1 }; return { data: state.release }; },
      async listReleaseAssets() { return { data: state.assets }; },
      async deleteReleaseAsset(args) { state.events.push('delete-asset'); state.assets = state.assets.filter(a => a.id !== args.asset_id); },
      async uploadReleaseAsset(args) {
        state.events.push('upload');
        if (state.failUpload) throw new Error('Registry-independent asset upload failure');
        state.assets.push({ name: args.name, id: state.assets.length + 1 });
        return { data: {} };
      },
      async updateRelease(args) {
        state.events.push('publish');
        state.release = { ...state.release, ...args, html_url: 'https://github.com/hwameistor/hwameistor/releases/tag/' + version };
        return { data: state.release };
      },
    },
  }, async paginate(method, args) { return (await method(args)).data; } };
  return { github, core, state, outputs };
}

function gitFixture({ current = version, previous = 'v1.1.2', head = SHA } = {}) {
  return args => {
    if (args[0] === 'rev-parse') return head;
    if (args[0] === 'ls-tree') return previous === null ? '' : 'VERSION';
    if (args[0] === 'show') return args[1] === `${SHA}:VERSION` ? current : previous;
    throw new Error(`Unexpected git invocation: ${args}`);
  };
}

async function withArtifacts(callback, releaseVersion = version) {
  const cwd = process.cwd();
  const temporary = fs.mkdtempSync(path.join(os.tmpdir(), 'hwameistor-release-test-'));
  try {
    process.chdir(temporary);
    fs.mkdirSync('assets');
    for (const target of ['linux-amd64', 'linux-arm64', 'darwin-amd64', 'darwin-arm64', 'windows-amd64.exe', 'windows-arm64.exe']) {
      fs.writeFileSync(`assets/hwameictl-${target}`, `binary-${target}`);
    }
    fs.writeFileSync('assets/images.json', JSON.stringify({ version: releaseVersion, sourceSha: SHA, images: [{ reference: 'example/image:' + releaseVersion }] }));
    fs.mkdirSync(`changelogs/released/${releaseVersion}`, { recursive: true });
    fs.writeFileSync(`changelogs/released/${releaseVersion}/changelog`, 'Reviewed release notes\n');
    fs.writeFileSync('notes.md', 'Reviewed release notes\n');
    await callback({ context, version: releaseVersion, directory: 'assets', notesPath: 'notes.md' });
  } finally {
    process.chdir(cwd);
    fs.rmSync(temporary, { recursive: true, force: true });
  }
}

test('accepts stable and prerelease versions, rejects ambiguous or invalid image tags', () => {
  for (const v of ['v1.1.3', 'v1.1.3-alpha1', 'v1.1.3-rc.2']) assert.equal(readVersion(v + '\n'), v);
  for (const v of ['1.1.3', 'v1.01.3', 'v1.1.3+build', 'v1.1.3\nv1.1.4', 'v1.1.3-rc.01', 'v1.1.3; echo bad']) {
    assert.throws(() => readVersion(v), /VERSION/);
  }
});

test('compares SemVer numerically, including prerelease precedence', () => {
  for (const [a, b] of [['v1.1.3', 'v1.1.3-rc.1'], ['v1.10.0', 'v1.9.0'], ['v1.1.3-rc.10', 'v1.1.3-rc.2'], ['v1.1.3-alpha.1', 'v1.1.3-alpha'], ['v1.1.3-alpha', 'v1.1.3-1']]) {
    assert.equal(compareVersions(a, b), 1);
    assert.equal(compareVersions(b, a), -1);
  }
  assert.equal(compareVersions(version, version), 0);
});

test('initial VERSION creation and branch creation never publish', async () => {
  for (const [ctx, git] of [[context, gitFixture({ previous: null })], [{ ...context, payload: { before: '0'.repeat(40), created: true } }, gitFixture()]]) {
    const f = fixture();
    await prepareRelease({ ...f, context: ctx, git });
    assert.equal(f.outputs.release, 'false');
    assert.deepEqual(f.state.events, []);
  }
});

test('whitespace-only VERSION changes do not publish', async () => {
  const f = fixture();
  await prepareRelease({ ...f, context, git: gitFixture({ current: version + '\n', previous: version }) });
  assert.equal(f.outputs.release, 'false');
});

test('new release locks source SHA but performs no mutations during preparation', async () => {
  const f = fixture();
  await prepareRelease({ ...f, context, git: gitFixture() });
  assert.deepEqual(f.outputs, { version, 'source-sha': SHA, release: 'true' });
  assert.deepEqual(f.state.events, []);
});

test('rejects a drifting checkout, version rollback, and a version outside its maintenance branch', async () => {
  for (const options of [{ head: BEFORE }, { previous: 'v1.1.4' }, { current: 'v1.2.0' }]) {
    await assert.rejects(prepareRelease({ ...fixture(), context, git: gitFixture(options) }));
  }
});

test('rejects an existing tag pointing at another commit', async () => {
  const f = fixture({ tag: BEFORE });
  await assert.rejects(prepareRelease({ ...f, context, git: gitFixture() }), /refusing to move/);
  assert.deepEqual(f.state.events, []);
});

test('already published releases are skipped rather than rebuilt', async () => {
  const f = fixture({ tag: SHA, release: { draft: false } });
  await prepareRelease({ ...f, context, git: gitFixture() });
  assert.equal(f.outputs.release, 'false');
});

test('a tag created during a failed finalization prevents rebuilding and overwriting its images', async () => {
  const f = fixture({ tag: SHA, release: { draft: true } });
  await assert.rejects(assertUnreleased({ ...f, context, version }), /rerun only the failed publish-release job/);
});

test('resolves annotated tags to their source commit', async () => {
  const f = fixture();
  f.github.rest.git.getRef = async () => ({ data: { object: { type: 'tag', sha: BEFORE } } });
  f.github.rest.git.getTag = async () => ({ data: { object: { type: 'commit', sha: SHA } } });
  assert.equal(await tagCommit(f.github, context.repo, version), SHA);
});

test('authentication failures are not mistaken for missing tags', async () => {
  const f = fixture();
  f.github.rest.git.getRef = async () => { throw Object.assign(new Error('Unauthorized'), { status: 401 }); };
  await assert.rejects(prepareRelease({ ...f, context, git: gitFixture() }), /Unauthorized/);
});

test('missing binaries, wrong image source and changed notes all block tag creation', async () => {
  for (const mutate of [() => fs.unlinkSync('assets/hwameictl-linux-arm64'), () => fs.writeFileSync('assets/images.json', JSON.stringify({ sourceSha: BEFORE, version })), () => fs.writeFileSync('notes.md', 'Unreviewed replacement')]) {
    await withArtifacts(async args => {
      mutate();
      const f = fixture();
      await assert.rejects(publishRelease({ ...f, ...args }));
      assert.deepEqual(f.state.events, []);
    });
  }
});

test('creates the tag at the build SHA and publishes only after every asset uploads', async () => {
  await withArtifacts(async args => {
    const f = fixture();
    await publishRelease({ ...f, ...args });
    assert.equal(f.state.tag, SHA);
    assert.deepEqual(f.state.events, ['tag', 'draft', ...Array(8).fill('upload'), 'publish']);
    assert.equal(f.state.release.draft, false);
    assert.equal(f.state.release.body, 'Reviewed release notes\n');
    assert.equal(f.state.release.make_latest, 'true');
    assert.ok(f.outputs['release-url']);
  });
});

test('asset upload failure leaves a draft; retry reuses the tag and draft', async () => {
  await withArtifacts(async args => {
    const f = fixture();
    f.state.failUpload = true;
    await assert.rejects(publishRelease({ ...f, ...args }), /upload failure/);
    assert.equal(f.state.release.draft, true);
    assert.ok(!f.state.events.includes('publish'));
    f.state.failUpload = false;
    await publishRelease({ ...f, ...args });
    assert.equal(f.state.events.filter(e => e === 'tag').length, 1);
    assert.equal(f.state.events.filter(e => e === 'draft').length, 1);
    assert.equal(f.state.release.draft, false);
  });
});

test('publishing a backport cannot replace a newer stable Latest release', async () => {
  await withArtifacts(async args => {
    const f = fixture({ latest: 'v1.2.0' });
    await publishRelease({ ...f, ...args });
    assert.equal(f.state.release.make_latest, 'false');
  });
});

test('prereleases are marked correctly and never become Latest', async () => {
  await withArtifacts(async args => {
    const f = fixture();
    await publishRelease({ ...f, ...args });
    assert.equal(f.state.release.prerelease, true);
    assert.equal(f.state.release.make_latest, 'false');
  }, 'v1.1.3-alpha1');
});

test('a changed or deleted tag during attachment upload blocks publication', async () => {
  for (const changedTag of [null, BEFORE]) {
    await withArtifacts(async args => {
      const f = fixture();
      const upload = f.github.rest.repos.uploadReleaseAsset;
      f.github.rest.repos.uploadReleaseAsset = async options => {
        const result = await upload(options);
        f.state.tag = changedTag;
        return result;
      };
      await assert.rejects(publishRelease({ ...f, ...args }), /disappeared|refusing to move/);
      assert.equal(f.state.release.draft, true);
      assert.ok(!f.state.events.includes('publish'));
    });
  }
});
