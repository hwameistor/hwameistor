import copy
import importlib.util
import json
import subprocess
import tempfile
import unittest
from pathlib import Path
from unittest.mock import patch


ROOT = Path(__file__).resolve().parents[2]
SPEC = importlib.util.spec_from_file_location('release_images', ROOT / 'build/util/verify-release-images.py')
images = importlib.util.module_from_spec(SPEC)
SPEC.loader.exec_module(images)


def manifest():
    return {'manifests': [
        {'platform': {'os': 'linux', 'architecture': arch}, 'digest': f'sha256:{arch}'}
        for arch in ['amd64', 'arm64']
    ]}


class ReleaseImageTests(unittest.TestCase):
    def test_complete_platforms(self):
        self.assertEqual(images.platforms(manifest()), {'amd64': 'sha256:amd64', 'arm64': 'sha256:arm64'})

    def test_missing_or_duplicate_platform_fails(self):
        cases = [[], manifest()['manifests'][:1], manifest()['manifests'] * 2]
        for entries in cases:
            with self.subTest(entries=entries), self.assertRaises(ValueError):
                images.platforms({'manifests': entries})

    def test_wrong_os_and_unexpected_runtime_architecture_fail(self):
        for key, value in [('os', 'windows'), ('architecture', 's390x')]:
            candidate = manifest()
            candidate['manifests'][0]['platform'][key] = value
            with self.subTest(key=key), self.assertRaises(ValueError):
                images.platforms(candidate)

    def test_attestations_do_not_count_as_runtime_platforms(self):
        candidate = manifest()
        candidate['manifests'].append({
            'platform': {'os': 'unknown', 'architecture': 'unknown'},
            'annotations': {'vnd.docker.reference.type': 'attestation-manifest'},
        })
        self.assertEqual(len(images.platforms(candidate)), 2)

    @staticmethod
    def remote(reference):
        if '@' not in reference:
            return manifest()
        return {'config': {'digest': reference.split('@')[1] + '-config'}}

    @staticmethod
    def local(*args):
        arch = 'arm64' if args[3].endswith('-arm64') else 'amd64'
        return f'sha256:{arch}-config'

    def test_published_image_matches_both_local_architectures(self):
        with patch.object(images, 'inspect', side_effect=self.remote), patch.object(images, 'run', side_effect=self.local):
            report = images.verify_image('example/component', 'v1.1.3')
            self.assertEqual(report['platforms']['arm64']['config'], 'sha256:arm64-config')

    def test_stale_remote_image_is_rejected(self):
        with patch.object(images, 'inspect', side_effect=self.remote), patch.object(images, 'run', return_value='sha256:old-image'):
            with self.assertRaisesRegex(ValueError, 'does not match'):
                images.verify_image('example/component', 'v1.1.3')

    def test_final_check_rejects_replaced_digest_even_if_configuration_matches(self):
        with patch.object(images, 'inspect', side_effect=self.remote), patch.object(images, 'run', side_effect=self.local):
            expected = images.verify_image('example/component', 'v1.1.3')['platforms']
            images.verify_image('example/component', 'v1.1.3', expected)
            changed = copy.deepcopy(expected)
            changed['arm64']['digest'] = 'sha256:replaced'
            with self.assertRaisesRegex(ValueError, 'changed after'):
                images.verify_image('example/component', 'v1.1.3', changed)

    def test_registry_failure_is_not_treated_as_success(self):
        with patch.object(images, 'inspect', side_effect=subprocess.CalledProcessError(1, ['docker'])):
            with self.assertRaises(subprocess.CalledProcessError):
                images.verify_image('example/component', 'v1.1.3')

    def test_incomplete_report_blocks_finalization(self):
        with tempfile.TemporaryDirectory() as directory:
            report = Path(directory) / 'images.json'
            report.write_text(json.dumps({'version': 'v1.1.3', 'sourceSha': 'abc', 'images': []}))
            argv = ['verify', '--registry', 'example', '--version', 'v1.1.3', '--source-sha', 'abc', '--report', str(report), '--verify-report']
            with patch('sys.argv', argv), patch.object(images, 'run', side_effect=['abc', 'example/component']):
                with self.assertRaisesRegex(ValueError, 'complete release image list'):
                    images.main()


if __name__ == '__main__':
    unittest.main()
