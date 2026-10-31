'use strict';

const fs = require('node:fs');
const path = require('node:path');
const crypto = require('node:crypto');
const { execFileSync } = require('node:child_process');

function readVersion(text) {
  const version = text.trim();
  const match = /^v(0|[1-9]\d*)\.(0|[1-9]\d*)\.(0|[1-9]\d*)(?:-([0-9A-Za-z-]+(?:\.[0-9A-Za-z-]+)*))?$/.exec(version);
  if (!match || version.length > 122 || match[4]?.split('.').some(p => /^0\d+$/.test(p))) {
    throw new Error('VERSION must contain one SemVer such as v1.1.3 or v1.1.3-alpha1 (without build metadata)');
  }
  return version;
}

function compareVersions(a, b) {
  const parts = v => readVersion(v).slice(1).split(/-(.*)/s);
  const [ac, ap] = parts(a);
  const [bc, bp] = parts(b);
  const an = ac.split('.').map(BigInt);
  const bn = bc.split('.').map(BigInt);
  for (let i = 0; i < 3; i++) {
    if (an[i] !== bn[i]) return an[i] > bn[i] ? 1 : -1;
  }
  if (ap === bp) return 0;
  if (ap === undefined || bp === undefined) return ap === undefined ? 1 : -1;
  const aa = ap.split('.');
  const ba = bp.split('.');
  for (let i = 0; i < Math.max(aa.length, ba.length); i++) {
    if (aa[i] === ba[i]) continue;
    if (aa[i] === undefined || ba[i] === undefined) return aa[i] === undefined ? -1 : 1;
    const numericA = /^\d+$/.test(aa[i]);
    const numericB = /^\d+$/.test(ba[i]);
    if (numericA && numericB) return BigInt(aa[i]) > BigInt(ba[i]) ? 1 : -1;
    if (numericA !== numericB) return numericA ? -1 : 1;
    return aa[i] > ba[i] ? 1 : -1;
  }
  return 0;
}

async function optional(request) {
  try { return (await request()).data; }
  catch (error) { if (error.status === 404) return null; throw error; }
}

async function tagCommit(github, repo, version) {
  const ref = await optional(() => github.rest.git.getRef({ ...repo, ref: `tags/${version}` }));
  if (!ref) return null;
  let object = ref.object;
  for (let depth = 0; object.type === 'tag' && depth < 10; depth++) {
    object = (await github.rest.git.getTag({ ...repo, tag_sha: object.sha })).data.object;
  }
  if (object.type !== 'commit') throw new Error(`Tag ${version} does not resolve to a commit`);
  return object.sha;
}

function checkTag(version, existing, sha) {
  if (existing && existing !== sha) throw new Error(`Tag ${version} already points to ${existing}, not ${sha}; refusing to move it`);
}

async function assertUnreleased({ github, context, version }) {
  const existing = await tagCommit(github, context.repo, version);
  checkTag(version, existing, context.sha);
  if (existing) throw new Error(`Tag ${version} already exists; rerun only the failed publish-release job to reuse its verified artifacts`);
  const release = await optional(() => github.rest.repos.getReleaseByTag({ ...context.repo, tag: version }));
  if (release) throw new Error(`Release ${version} already exists without its tag; refusing to overwrite it`);
}

async function prepareRelease({ github, context, core, git = args => execFileSync('git', args, { encoding: 'utf8' }).trim() }) {
  const sha = context.sha;
  if (git(['rev-parse', 'HEAD']) !== sha) throw new Error('Checkout does not match the triggering commit');
  const version = readVersion(git(['show', `${sha}:VERSION`]));
  core.setOutput('version', version);
  core.setOutput('source-sha', sha);
  core.setOutput('release', 'false');
  const before = context.payload.before;
  // Adding the workflow and VERSION, or creating a branch, is not a release.
  if (!before || /^0+$/.test(before) || context.payload.created) {
    core.info('Skipping branch initialization');
    return;
  }
  if (!git(['ls-tree', '--name-only', before, '--', 'VERSION'])) {
    core.info('Skipping initial VERSION creation');
    return;
  }
  const previous = readVersion(git(['show', `${before}:VERSION`]));
  if (previous === version) { core.info('VERSION is unchanged'); return; }
  if (compareVersions(version, previous) <= 0) throw new Error(`VERSION must advance beyond ${previous}`);
  const branch = /^refs\/heads\/release\/v(\d+)\.(\d+)(?:\.\d+)?$/.exec(context.ref);
  if (branch && !version.startsWith(`v${branch[1]}.${branch[2]}.`)) {
    throw new Error(`VERSION ${version} does not match release branch ${context.ref}`);
  }
  const existing = await tagCommit(github, context.repo, version);
  checkTag(version, existing, sha);
  if (existing) {
    const release = await optional(() => github.rest.repos.getReleaseByTag({ ...context.repo, tag: version }));
    if (release && !release.draft) { core.info(`${version} is already published from this commit`); return; }
  }
  await assertUnreleased({ github, context, version });
  core.setOutput('release', 'true');
}

function releaseAssets(directory) {
  const names = [];
  for (const os of ['linux', 'darwin', 'windows']) {
    for (const arch of ['amd64', 'arm64']) names.push(`hwameictl-${os}-${arch}${os === 'windows' ? '.exe' : ''}`);
  }
  names.push('images.json');
  const assets = names.map(name => {
    const data = fs.readFileSync(path.join(directory, name));
    if (!data.length) throw new Error(`Empty release asset: ${name}`);
    return { name, data };
  });
  const sums = assets.map(({ name, data }) => `${crypto.createHash('sha256').update(data).digest('hex')}  ${name}\n`).join('');
  assets.push({ name: 'SHA256SUMS', data: Buffer.from(sums) });
  return assets;
}

async function publishRelease({ github, context, core, version, directory, notesPath }) {
  version = readVersion(version);
  // Validate every local input before creating a tag or draft release.
  const assets = releaseAssets(directory);
  const images = JSON.parse(assets.find(a => a.name === 'images.json').data.toString());
  if (images.version !== version || images.sourceSha !== context.sha || !images.images?.length) {
    throw new Error('Image report does not match this release');
  }
  const body = fs.readFileSync(notesPath, 'utf8');
  const reviewedBody = fs.readFileSync(`changelogs/released/${version}/changelog`, 'utf8');
  if (!body.trim() || body !== reviewedBody) throw new Error('Release body must match the reviewed file at the source commit');
  const repo = context.repo;
  const existing = await tagCommit(github, repo, version);
  checkTag(version, existing, context.sha);
  let release = await optional(() => github.rest.repos.getReleaseByTag({ ...repo, tag: version }));
  if (release && !existing) throw new Error('An existing release has no matching tag');
  if (release && !release.draft) {
    core.info(`${version} is already published`);
    return;
  }
  if (!existing) {
    try { await github.rest.git.createRef({ ...repo, ref: `refs/tags/${version}`, sha: context.sha }); }
    catch (error) {
      if (error.status !== 422) throw error;
      const raced = await tagCommit(github, repo, version);
      if (!raced) throw error;
      checkTag(version, raced, context.sha);
    }
  }
  const prerelease = version.includes('-');
  if (!release) release = (await github.rest.repos.createRelease({
    ...repo, tag_name: version, target_commitish: context.sha, name: version, body,
    draft: true, prerelease,
  })).data;

  const uploaded = await github.paginate(github.rest.repos.listReleaseAssets, { ...repo, release_id: release.id, per_page: 100 });
  for (const asset of assets) {
    const current = uploaded.find(a => a.name === asset.name);
    if (current) await github.rest.repos.deleteReleaseAsset({ ...repo, asset_id: current.id });
    await github.rest.repos.uploadReleaseAsset({
      ...repo, release_id: release.id, name: asset.name, data: asset.data,
      headers: { 'content-type': 'application/octet-stream', 'content-length': asset.data.length },
    });
  }
  const latest = await optional(() => github.rest.repos.getLatestRelease(repo));
  // Backports and prereleases must not replace a newer stable release as Latest.
  const makeLatest = !prerelease && (!latest || compareVersions(version, latest.tag_name) >= 0);
  const finalTag = await tagCommit(github, repo, version);
  if (!finalTag) throw new Error(`Tag ${version} disappeared during publication`);
  checkTag(version, finalTag, context.sha);
  const published = (await github.rest.repos.updateRelease({
    ...repo, release_id: release.id, name: version, body, draft: false, prerelease,
    make_latest: String(makeLatest),
  })).data;
  core.setOutput('release-url', published.html_url);
  await core.summary.addRaw(`[${version}](${published.html_url})\n`).write();
}

module.exports = { readVersion, compareVersions, tagCommit, assertUnreleased, prepareRelease, publishRelease };
