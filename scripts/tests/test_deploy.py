"""Проверяет порядок деплоя и отказоустойчивость без Docker, SSH и production-секретов."""
import json
import os
from pathlib import Path
import subprocess
import tarfile
import tempfile
import unittest


SCRIPT = Path(__file__).resolve().parents[1] / "deploy.sh"
OLD_SHA = "a" * 40
NEW_SHA = "b" * 40


class DeployTests(unittest.TestCase):
    def setUp(self):
        self.tmp = tempfile.TemporaryDirectory()
        self.addCleanup(self.tmp.cleanup)
        self.root = Path(self.tmp.name)
        (self.root / "shared").mkdir()
        (self.root / "shared/.env.production").write_text("JWT_SECRET=test-only\n")
        self.log = self.root / "docker.jsonl"
        self.bin = self.root / "bin"
        self.bin.mkdir()
        docker = self.bin / "docker"
        docker.write_text("""#!/usr/bin/env python3
import json, os, sys
args = sys.argv[1:]
with open(os.environ['MOCK_DOCKER_LOG'], 'a') as log:
    log.write(json.dumps({'release': os.environ['RELEASE_ID'], 'args': args}) + '\\n')
mode = os.environ.get('MOCK_FAILURE')
if mode == 'migration' and 'migrate' in args and 'run' in args:
    sys.exit(1)
if mode == 'health' and 'exec' in args and os.environ['RELEASE_ID'] == 'b' * 40:
    sys.exit(1)
if mode == 'startup' and 'up' in args and 'frontend' in args:
    sys.exit(1)
""")
        docker.chmod(0o755)
        flock = self.bin / "flock"
        flock.write_text('#!/bin/sh\nexit "${MOCK_LOCK_FAIL:-0}"\n')
        flock.chmod(0o755)
        self.source = self.root / "source"
        (self.source / "backend").mkdir(parents=True)
        (self.source / "backend/go.mod").write_text("module example\n\ngo 1.25.5\n")
        (self.source / "docker-compose.production.yml").write_text("services: {}\n")
        self.archive = self.root / "source.tar.gz"
        with tarfile.open(self.archive, "w:gz") as archive:
            for file in self.source.rglob("*"):
                if file.is_file():
                    archive.add(file, arcname=file.relative_to(self.source))

    def run_deploy(self, action="deploy", release=NEW_SHA, failure="", lock_fail="0"):
        return subprocess.run(
            ["sh", str(SCRIPT), action, release, str(self.archive)],
            env={**os.environ, "DEPLOY_ROOT": str(self.root),
                 "PATH": str(self.bin) + os.pathsep + os.environ["PATH"],
                 "MOCK_DOCKER_LOG": str(self.log), "MOCK_FAILURE": failure,
                 "MOCK_LOCK_FAIL": lock_fail},
            capture_output=True, text=True, check=False,
        )

    def calls(self):
        return [json.loads(line) for line in self.log.read_text().splitlines()]

    def prepare_previous(self):
        result = self.run_deploy(release=OLD_SHA)
        self.assertEqual(result.returncode, 0, result.stderr)
        self.log.write_text("")

    def test_migration_precedes_start_and_success_marks_current(self):
        result = self.run_deploy()
        self.assertEqual(result.returncode, 0, result.stderr)
        calls = self.calls()
        migration = next(i for i, call in enumerate(calls) if "migrate" in call["args"])
        startup = next(i for i, call in enumerate(calls) if "up" in call["args"] and "frontend" in call["args"])
        self.assertLess(migration, startup)
        self.assertTrue(any("/dev/null" in call["args"] for call in calls))
        self.assertEqual((self.root / "current").resolve().name, NEW_SHA)

    def test_failed_migration_leaves_previous_runtime_selected(self):
        self.prepare_previous()
        result = self.run_deploy(failure="migration")
        self.assertNotEqual(result.returncode, 0)
        self.assertEqual((self.root / "current").resolve().name, OLD_SHA)
        self.assertFalse(any("up" in call["args"] and "frontend" in call["args"] for call in self.calls()))

    def test_failed_health_restores_previous_images_and_returns_failure(self):
        self.prepare_previous()
        result = self.run_deploy(failure="health")
        self.assertNotEqual(result.returncode, 0)
        self.assertEqual((self.root / "current").resolve().name, OLD_SHA)
        restored = [call for call in self.calls() if call["release"] == OLD_SHA]
        self.assertTrue(any("--no-build" in call["args"] for call in restored))
        self.assertFalse(any("migrate" in call["args"] for call in restored))

    def test_failed_first_start_does_not_mark_release_current(self):
        result = self.run_deploy(failure="startup")
        self.assertNotEqual(result.returncode, 0)
        self.assertFalse((self.root / "current").exists())

    def test_manual_rollback_never_builds_or_migrates(self):
        self.prepare_previous()
        self.assertEqual(self.run_deploy().returncode, 0)
        self.log.write_text("")
        result = self.run_deploy(action="rollback", release=OLD_SHA)
        self.assertEqual(result.returncode, 0, result.stderr)
        self.assertFalse(any("build" in call["args"] or "migrate" in call["args"] for call in self.calls()))
        self.assertEqual((self.root / "current").resolve().name, OLD_SHA)

    def test_invalid_release_is_rejected_before_docker(self):
        for release in ["main", "../outside", "a" * 39, "a" * 41, "b;echo x"]:
            with self.subTest(release=release):
                self.assertNotEqual(self.run_deploy(release=release).returncode, 0)
        self.assertFalse(self.log.exists())

    def test_lock_prevents_concurrent_deploy(self):
        self.assertNotEqual(self.run_deploy(lock_fail="1").returncode, 0)
        self.assertFalse(self.log.exists())

    def test_missing_environment_is_rejected(self):
        (self.root / "shared/.env.production").unlink()
        self.assertNotEqual(self.run_deploy().returncode, 0)
        self.assertFalse(self.log.exists())


if __name__ == "__main__":
    unittest.main()
