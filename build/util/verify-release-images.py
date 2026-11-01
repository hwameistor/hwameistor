#!/usr/bin/env python3
"""Verify published multi-architecture images against the built images or a saved report."""

import argparse
import json
import subprocess
from pathlib import Path


def run(*args):
    return subprocess.check_output(args, text=True).strip()


def inspect(reference):
    return json.loads(run('docker', 'buildx', 'imagetools', 'inspect', '--raw', reference))


def platforms(manifest):
    result = {}
    for entry in manifest.get('manifests', []):
        platform = entry.get('platform', {})
        if entry.get('annotations', {}).get('vnd.docker.reference.type') == 'attestation-manifest':
            continue
        arch = platform.get('architecture')
        if platform.get('os') != 'linux' or arch not in ('amd64', 'arm64') or arch in result:
            raise ValueError('Expected exactly one linux/amd64 and one linux/arm64 image')
        result[arch] = entry['digest']
    if set(result) != {'amd64', 'arm64'}:
        raise ValueError('Missing linux/amd64 or linux/arm64 image')
    return result


def verify_image(image, version, expected=None):
    reference = f'{image}:{version}'
    manifests = platforms(inspect(reference))
    verified = {}
    for arch, digest in sorted(manifests.items()):
        config = inspect(f'{image}@{digest}')['config']['digest']
        if expected is None:
            local = run('docker', 'image', 'inspect', f'{reference}-{arch}', '--format', '{{.Id}}')
            if local != config:
                raise ValueError(f'{reference}: {arch} does not match the image built by this run')
        elif expected.get(arch) != {'digest': digest, 'config': config}:
            raise ValueError(f'{reference}: {arch} changed after publication verification')
        verified[arch] = {'digest': digest, 'config': config}
    return {'reference': reference, 'platforms': verified}


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('--registry', required=True)
    parser.add_argument('--version', required=True)
    parser.add_argument('--source-sha', required=True)
    parser.add_argument('--report', required=True)
    parser.add_argument('--verify-report', action='store_true')
    args = parser.parse_args()
    if run('git', 'rev-parse', 'HEAD') != args.source_sha:
        raise ValueError('Checkout does not match release source SHA')
    images = run('make', '--no-print-directory', '-s', 'release-image-list', f'IMAGE_REGISTRY={args.registry}').splitlines()
    if not images or len(set(images)) != len(images):
        raise ValueError('Invalid release image list')
    previous = None
    if args.verify_report:
        previous = json.loads(Path(args.report).read_text())
        if previous['version'] != args.version or previous['sourceSha'] != args.source_sha:
            raise ValueError('Report does not match this version and source SHA')
        if [item['reference'] for item in previous['images']] != [f'{image}:{args.version}' for image in images]:
            raise ValueError('Report does not cover the complete release image list')
    results = []
    for index, image in enumerate(images):
        expected = previous['images'][index]['platforms'] if previous else None
        results.append(verify_image(image, args.version, expected))
        print(f'Verified {image}:{args.version} (linux/amd64, linux/arm64)', flush=True)
    if not args.verify_report:
        destination = Path(args.report)
        destination.parent.mkdir(parents=True, exist_ok=True)
        destination.write_text(json.dumps({'version': args.version, 'sourceSha': args.source_sha, 'images': results}, indent=2) + '\n')


if __name__ == '__main__':
    main()
