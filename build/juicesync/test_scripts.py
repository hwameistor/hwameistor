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

    def wrapper_in(self, root):
        # Redirect the fixed installation directory into an isolated filesystem fixture.
        wrapper = root / "juicesync"
        wrapper.write_text((SCRIPTS / "juicesync").read_text().replace("/opt/juicefs", str(root / "binaries")))
        return wrapper

    def test_wrapper_selects_installed_binary(self):
        with tempfile.TemporaryDirectory() as directory:
            root = Path(directory)
            wrapper = self.wrapper_in(root)
            for version in ("1.2.0", "1.4.1", "1.5.0"):
                binary = root / "binaries" / version / "juicefs"
                binary.parent.mkdir(parents=True)
                binary.write_text('#!/bin/sh\nprintf "%s\\n" "$0" "$@"\nexit 37\n')
                binary.chmod(0o755)
            for version in (None, "", "1.2.0", "1.4.1", "1.5.0"):
                with self.subTest(version=version):
                    env = dict(os.environ)
                    env.pop("JUICEFS_VERSION", None)
                    if version is not None:
                        env["JUICEFS_VERSION"] = version
                    args = ["--check-all", "/source/space 中文 *", "/target/path"]
                    result = subprocess.run(["sh", str(wrapper), *args], env=env, capture_output=True, text=True)
                    self.assertEqual(result.returncode, 37, result.stderr)
                    expected = root / "binaries" / (version or "1.2.0") / "juicefs"
                    self.assertEqual(result.stdout.splitlines(), [str(expected), "sync", *args])

    def test_unavailable_binary_does_not_fallback(self):
        with tempfile.TemporaryDirectory() as directory:
            root = Path(directory)
            wrapper = self.wrapper_in(root)
            binary = root / "binaries" / "not-executable" / "juicefs"
            binary.parent.mkdir(parents=True)
            binary.touch(mode=0o644)
            (root / "binaries" / "directory" / "juicefs").mkdir(parents=True)
            for version in ("not-installed", "not-executable", "directory"):
                result = subprocess.run(["sh", str(wrapper), "--help"],
                                        env=dict(os.environ, JUICEFS_VERSION=version), capture_output=True, text=True)
                self.assertEqual(result.returncode, 2)
                self.assertIn("JuiceFS binary not found or not executable:", result.stderr)
                self.assertIn(str(root / "binaries" / version / "juicefs"), result.stderr)

    def test_version_cannot_traverse_directories(self):
        for version in (".", "..", "../../bin/sh", "nested/version"):
            result = subprocess.run(["sh", str(SCRIPTS / "juicesync"), "--help"],
                                    env=dict(os.environ, JUICEFS_VERSION=version), capture_output=True, text=True)
            self.assertEqual(result.returncode, 2)
            self.assertIn("Invalid JUICEFS_VERSION", result.stderr)


if __name__ == "__main__":
    unittest.main()
