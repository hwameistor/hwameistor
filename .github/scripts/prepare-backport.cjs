'use strict';

const targetByLabel = new Map([
  ['cherry-pick-v1', 'release/v1.0.0'], // Legacy label.
  ['cherry-pick-v1.0', 'release/v1.0.0'],
  ['cherry-pick-v1.1', 'release/v1.1'],
]);

function targetsForLabels(labels) {
  return [...new Set(labels.map(label => targetByLabel.get(label.name)).filter(Boolean))];
}

function isBackportOf(candidate, pullNumber, target, repository) {
  if (candidate.base.ref !== target || (candidate.state !== 'open' && !candidate.merged_at)) {
    return false;
  }

  const body = candidate.body || '';
  const marker = `<!-- hwameistor-backport:${pullNumber}:${target} -->`;
  const branch = `backport-${pullNumber}-to-${target}`;
  if (body.includes(marker) ||
      (candidate.head.repo?.full_name === repository && candidate.head.ref === branch)) {
    return true;
  }

  // Also recognize manual backports. Only the first reference on the source
  // line identifies the source PR; later references may refer to fixed issues.
  const source = body.match(/^\s*Backport of [^#\n]*#([1-9]\d*)\b/im);
  return source !== null && source[1] === String(pullNumber);
}

async function prepareBackport({ github, context, core, pullNumber }) {
  const number = String(pullNumber);
  if (!/^[1-9]\d*$/.test(number) || !Number.isSafeInteger(Number(number))) {
    throw new Error('pull_number must be a positive integer without leading zeros');
  }

  const { owner, repo } = context.repo;
  const repository = `${owner}/${repo}`;
  const { data: pull } = await github.rest.pulls.get({ owner, repo, pull_number: Number(number) });
  if (!pull.merged || pull.base.ref !== 'main' || pull.base.repo.full_name !== repository) {
    throw new Error('Only merged PRs targeting main in this repository can be backported');
  }

  // Read current labels because queued events may contain labels since removed.
  const targets = targetsForLabels(pull.labels);
  const pending = [];
  const summary = [];
  for (const target of targets) {
    const candidates = await github.paginate(github.rest.pulls.list, {
      owner, repo, base: target, state: 'all', per_page: 100,
    });
    const existing = candidates.find(candidate => isBackportOf(candidate, number, target, repository));
    if (existing) {
      core.info(`Skipping ${target}: backport #${existing.number} is already open or merged`);
      summary.push(`- Skipped \`${target}\`: existing [#${existing.number}](${existing.html_url}).`);
    } else {
      pending.push(target);
      summary.push(`- Ready to backport to \`${target}\`.`);
    }
  }

  core.setOutput('pull_number', number);
  core.setOutput('targets', pending.join(' '));
  await core.summary
    .addHeading(`Backport of #${number}`)
    .addRaw(summary.length ? summary.join('\n') : 'No supported cherry-pick labels remain.', true)
    .write();
}

module.exports = { targetsForLabels, isBackportOf, prepareBackport };
