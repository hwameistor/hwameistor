#!/usr/bin/env python3
"""Publication-stage companion, copied into the consuming repository's CI.

Never called by Release Assistant's generate/validate commands. The release
workflow invokes this only after the reviewed release PR has been merged.
"""
import hashlib
import io
import json
import os
from pathlib import Path
import re
import sys
import tarfile
import urllib.error
import urllib.parse
import urllib.request


class SafeRedirect(urllib.request.HTTPRedirectHandler):
    def redirect_request(self, req, fp, code, msg, headers, newurl):
        redirected = super().redirect_request(req, fp, code, msg, headers, newurl)
        if urllib.parse.urlsplit(req.full_url).netloc != urllib.parse.urlsplit(newurl).netloc:
            redirected.remove_header('Authorization')
        return redirected


class GitHub:
    def __init__(self, repository, token):
        if not re.fullmatch(r'[\w.-]+/[\w.-]+', repository) or not token:
            raise ValueError('Repository and publishing token are required')
        self.root = 'https://api.github.com/repos/' + repository
        self.token = token
        self.opener = urllib.request.build_opener(SafeRedirect())

    def request(self, method, path, data=None, binary=False, missing_ok=False):
        url = path if path.startswith('https://') else self.root + path
        if urllib.parse.urlsplit(url).hostname not in ('api.github.com', 'uploads.github.com'):
            raise ValueError('Unexpected GitHub endpoint')
        headers = {'Authorization': 'Bearer ' + self.token, 'User-Agent': 'hwameistor-release',
                   'X-GitHub-Api-Version': '2022-11-28',
                   'Accept': 'application/octet-stream' if binary else 'application/vnd.github+json'}
        if isinstance(data, bytes):
            payload = data
            headers['Content-Type'] = 'application/gzip'
        else:
            payload = json.dumps(data).encode() if data is not None else None
            headers['Content-Type'] = 'application/json'
        request = urllib.request.Request(url, data=payload, headers=headers, method=method)
        try:
            with self.opener.open(request, timeout=120) as response:
                content = response.read()
                return content if binary else json.loads(content)
        except urllib.error.HTTPError as error:
            if error.code == 404 and missing_ok:
                return None
            raise RuntimeError(f'GitHub {method} failed with HTTP {error.code}') from None

    def tag(self, version):
        ref = self.request('GET', '/git/ref/tags/' + urllib.parse.quote(version, safe=''), missing_ok=True)
        if ref is None:
            return None
        obj = ref['object']
        for _ in range(8):
            if obj['type'] != 'tag':
                break
            obj = self.request('GET', '/git/tags/' + obj['sha'])['object']
        if obj['type'] != 'commit':
            raise ValueError('Release tag does not resolve to a commit')
        return obj['sha']


def archive_contents(data):
    # Helm package timestamps vary across builds. Compare contents without
    # extracting paths, then retain the already published bytes for index hashes.
    result = {}
    with tarfile.open(fileobj=io.BytesIO(data), mode='r:gz') as archive:
        for member in archive:
            if member.name in result:
                raise ValueError('Duplicate chart archive entry')
            stream = archive.extractfile(member) if member.isfile() else None
            result[member.name] = (member.type, member.mode, member.linkname,
                                   hashlib.sha256(stream.read()).hexdigest() if stream else None)
    return result


def publish(api, version, source, body, prerelease, latest, package):
    if not re.fullmatch(r'v\d+\.\d+\.\d+(?:-[0-9A-Za-z.-]+)?', version):
        raise ValueError('Invalid canonical release version')
    if not re.fullmatch(r'[a-f0-9]{40}', source) or not body.strip():
        raise ValueError('Exact source SHA and reviewed release notes are required')
    if prerelease != ('-' in version) or prerelease and latest:
        raise ValueError('Inconsistent prerelease/latest flags')
    # Read the package before creating a release, so a missing file cannot
    # leave an empty release behind.
    Path(package).read_bytes()
    tag = api.tag(version)
    if tag is not None and tag != source:
        raise ValueError('Existing tag points at a different source commit')
    release = api.request('GET', '/releases/tags/' + urllib.parse.quote(version, safe=''), missing_ok=True)
    if release is None:
        release = api.request('POST', '/releases', {
            'tag_name': version, 'target_commitish': source, 'name': version, 'body': body,
            'draft': False, 'prerelease': prerelease, 'make_latest': str(latest).lower(),
        })
    elif release['body'] != body or release['prerelease'] != prerelease or release['draft']:
        raise ValueError('Existing release differs from the reviewed publication; refusing to overwrite it')
    if api.tag(version) != source:
        raise ValueError('Release tag changed concurrently')
    upload_asset(api, release, package, chart=True)
    return release['html_url']


def upload_asset(api, release, package, chart=False):
    package = Path(package)
    content = package.read_bytes()
    digest = 'sha256:' + hashlib.sha256(content).hexdigest()
    assets = release.get('assets', [])
    # A chart release has one package. Fail instead of silently missing assets if
    # a previous publisher attached more than GitHub includes in this response.
    if len(assets) >= 100:
        raise ValueError('Too many existing release assets for safe retry')
    existing = next((asset for asset in assets if asset['name'] == package.name), None)
    if existing:
        actual = existing.get('digest')
        if actual != digest:
            downloaded = api.request('GET', '/releases/assets/' + str(existing['id']), binary=True)
            actual = 'sha256:' + hashlib.sha256(downloaded).hexdigest()
            if actual != digest:
                try:
                    equivalent = chart and archive_contents(content) == archive_contents(downloaded)
                except (tarfile.TarError, ValueError, TypeError):
                    equivalent = False
                if not equivalent:
                    raise ValueError('Existing release asset has different content; refusing to overwrite it')
            package.write_bytes(downloaded)
    else:
        upload = release['upload_url'].split('{', 1)[0] + '?name=' + urllib.parse.quote(package.name, safe='')
        api.request('POST', upload, content)


def publish_binary(api, version, source, package):
    if not re.fullmatch(r'v\d+\.\d+\.\d+(?:-[0-9A-Za-z.-]+)?', version) or not re.fullmatch(r'[a-f0-9]{40}', source):
        raise ValueError('Canonical version and exact source SHA are required')
    if api.tag(version) != source:
        raise ValueError('Existing tag does not match the binary source')
    release = api.request('GET', '/releases/tags/' + urllib.parse.quote(version, safe=''), missing_ok=True)
    if release is None or release['draft']:
        raise ValueError('Publish the reviewed Chart release before uploading binaries')
    upload_asset(api, release, package)
    return release['html_url']


def main():
    def flag(key):
        value = os.environ[key]
        if value not in ('true', 'false'):
            raise ValueError(key + ' must be true or false')
        return value == 'true'
    api = GitHub(os.environ['GITHUB_REPOSITORY'], os.environ['GH_TOKEN'])
    if sys.argv[1:] == ['upload-binary']:
        print(publish_binary(api, os.environ['RELEASE_VERSION'], os.environ['RELEASE_SHA'], os.environ['RELEASE_ASSET']))
        return
    if sys.argv[1:]:
        raise ValueError('Unknown publication command')
    print(publish(api, os.environ['RELEASE_VERSION'], os.environ['RELEASE_SHA'],
                  Path(os.environ['RELEASE_NOTES']).read_text(), flag('RELEASE_PRERELEASE'),
                  flag('RELEASE_LATEST'), os.environ['CHART_PACKAGE']))


if __name__ == '__main__':
    try:
        main()
    except Exception as error:
        print('publish-chart: ' + str(error), file=sys.stderr)
        sys.exit(1)
