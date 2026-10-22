#!/usr/bin/env python3
"""Check shell argument boundaries and exit propagation without a cluster."""
import json
import os
from pathlib import Path
import subprocess
import tempfile
import unittest


SCRIPTS = Path(__file__).resolve().parent


class SyncScriptsTest(unittest.TestCase):
    def test_arguments_and_exit_status(self):
        for checked in (False, True):
            for code in (0, 37):
                with self.subTest(checked=checked, code=code), tempfile.TemporaryDirectory() as directory:
                    root = Path(directory)
                    stub = root / "juicesync"
                    stub.write_text("#!/usr/bin/env python3\nimport json, os, sys\n"
                                    "json.dump({'args':sys.argv[1:], 'key':os.environ.get('SSH_PRIVATE_KEY_PATH')}, open(os.environ['CAPTURE'], 'w'))\n"
                                    "sys.exit(int(os.environ['COPY_EXIT']))\n")
                    stub.chmod(0o755)
                    capture = root / "arguments.json"
                    env = dict(os.environ, PATH=str(root) + os.pathsep + os.environ["PATH"],
                               CAPTURE=str(capture), COPY_EXIT=str(code),
                               sourceNode="192.0.2.1", targetNode="192.0.2.2",
                               sourceMountPoint="/source/space 中文 *", targetMountPoint="/target/space 中文 *")
                    script = "sync_hwameistor_volumes_check.sh" if checked else "sync_hwameistor_volumes.sh"
                    result = subprocess.run(["sh", str(SCRIPTS / script)], env=env, capture_output=True, text=True)
                    self.assertEqual(result.returncode, code, result.stderr)
                    actual = json.loads(capture.read_text())
                    expected = ["--links", "--dirs", "--perms", "--force-update"]
                    if checked:
                        expected.append("--check-all")
                    expected += ["root@192.0.2.1:/source/space 中文 */", "root@192.0.2.2:/target/space 中文 */"]
                    self.assertEqual(actual["args"], expected)
                    self.assertEqual(actual["key"], "/root/.ssh/id_rsa")

    def test_missing_source_fails_before_copy(self):
        env = dict(os.environ, sourceNode="", sourceMountPoint="/source", targetNode="192.0.2.2", targetMountPoint="/target")
        for script in ("sync_hwameistor_volumes.sh", "sync_hwameistor_volumes_check.sh"):
            result = subprocess.run(["sh", str(SCRIPTS / script)], env=env, capture_output=True)
            self.assertNotEqual(result.returncode, 0)

    def test_invalid_version_does_not_fallback(self):
        for version in ("latest", "v1.2.0", "../../bin/sh"):
            result = subprocess.run(["sh", str(SCRIPTS / "juicesync"), "--help"],
                                    env=dict(os.environ, JUICEFS_VERSION=version), capture_output=True, text=True)
            self.assertEqual(result.returncode, 2)
            self.assertIn("Unsupported JUICEFS_VERSION", result.stderr)


if __name__ == "__main__":
    unittest.main()
