"""Deterministic protocol regressions; no product launches or timing benchmarks."""
import copy
import importlib.util
import json
from pathlib import Path
import subprocess
import sys
import tempfile
from types import SimpleNamespace
import unittest
from unittest.mock import patch

from startup_protocol import collect, pair_order, recorded_execution
from startup_process import group_running, run as process_run
import startup_process

spec = importlib.util.spec_from_file_location("profile_cli", Path(__file__).with_name("startup-profile.py"))
cli = importlib.util.module_from_spec(spec)
spec.loader.exec_module(cli)


class FakeObserver:
    def __init__(self, deltas=(1, 2, 8), invalid=(), failure=False):
        self.attempt = 0
        self.deltas, self.invalid, self.fail = deltas, invalid, failure
        self.gaps, self.saved, self.calls = [], [], 0

    def identity(self):
        return {"binary": "fixed", "compiler": "matched"}

    def wait_gap(self, seconds):
        self.gaps.append(seconds)
        self.attempt += 1

    def checkpoint(self, label):
        return {"label": label, "load": [2.01 if self.attempt in self.invalid else 2, 0, 0],
                "logical_cpus": 2, "own_heavy_witness_valid": True, "active_heavy_jobs": []}

    def elapsed(self, name):
        self.calls += 1
        if self.fail and self.calls == 7:
            raise RuntimeError("injected command failure")
        if name == "version":
            return 4 + self.deltas[min(self.attempt, len(self.deltas) - 1)]
        return 4

    def memory(self, name):
        return {"peak_footprint_bytes": 5_000_000, "maximum_rss_bytes": 12_000_000}

    def fresh(self, *args):
        return 200

    def failure(self, error):
        return {"message": str(error)}

    def persist(self, report):
        self.saved.append(copy.deepcopy(report))


class ProtocolTests(unittest.TestCase):
    def evaluate(self, observer):
        report = {}
        recorded_execution(lambda: collect(observer, report, observer.persist), report, observer.persist,
                           lambda error: cli.ProtocolObserver.failure(None, error))
        return report

    def test_order_and_aggregate_not_every_session(self):
        observer = FakeObserver()
        report = self.evaluate(observer)
        self.assertEqual(report["outcome"], "PASS")
        self.assertEqual(report["aggregate_delta_ms"], 2)
        self.assertEqual([s["startup_delta_ms"] for s in report["sessions"]], [1, 2, 8])
        self.assertEqual(observer.gaps, [60, 60])
        for index, session in enumerate(report["sessions"]):
            self.assertEqual(len(session["pairs"]), 30)
            self.assertEqual(len(session["memory"]["version"]), 5)
            self.assertEqual(len(session["fresh_copy"]), 5)
            self.assertEqual(session["pairs"][0]["order"], list(pair_order(index, 0)))
            self.assertEqual(session["pairs"][1]["order"], list(reversed(pair_order(index, 0))))

    def test_first_valid_selection_keeps_invalid_attempts(self):
        report = self.evaluate(FakeObserver(deltas=(0, 8, 0, 9, 10), invalid=(0, 2)))
        self.assertEqual(report["selected_attempts"], [2, 4, 5])
        self.assertEqual(report["outcome"], "FAIL")
        self.assertEqual(len(report["sessions"]), 5)
        self.assertTrue(report["sessions"][0]["checkpoints"])

    def test_fewer_than_three_valid_is_inconclusive(self):
        report = self.evaluate(FakeObserver(invalid=(0, 1, 2)))
        self.assertEqual(report["outcome"], "INCONCLUSIVE")
        self.assertEqual(report["selected_attempts"], [4, 5])

    def test_exact_startup_boundary(self):
        self.assertEqual(self.evaluate(FakeObserver((3, 3, 3)))["outcome"], "PASS")
        self.assertEqual(self.evaluate(FakeObserver((3.001,) * 3))["outcome"], "FAIL")

    def test_difference_of_medians_is_not_paired_median(self):
        class Paired(FakeObserver):
            def __init__(self):
                super().__init__()
                self.n = {"version": -1, "empty_go_control": -1}

            def elapsed(self, name):
                if name not in self.n:
                    return 4
                self.n[name] += 1
                return ((0, 100, 101) if name == "version" else (0, 1, 100))[(self.n[name] - 1) % 3]

        report = self.evaluate(Paired())
        self.assertEqual(report["aggregate_delta_ms"], 99)
        self.assertEqual(report["sessions"][0]["paired_difference_median_ms_reference"], 1)
        self.assertEqual(report["outcome"], "FAIL")

    def test_partial_failure_is_saved_and_not_valid(self):
        observer = FakeObserver(failure=True)
        report = self.evaluate(observer)
        first = report["sessions"][0]
        self.assertFalse(first["valid"])
        self.assertIn("injected command failure", first["failure"]["message"])
        self.assertTrue(first["pairs"])
        self.assertTrue(any(x["sessions"] and x["sessions"][0]["pairs"] for x in observer.saved))

    def test_provenance_change_between_attempts_invalidates(self):
        class Changed(FakeObserver):
            def identity(self):
                return {"binary": self.attempt}
        report = self.evaluate(Changed())
        self.assertEqual(report["outcome"], "INCONCLUSIVE")
        self.assertEqual(report["selected_attempts"], [1])
        self.assertEqual(len(report["sessions"]), 2)
        self.assertIn("provenance", report["fatal_failure"]["message"])

    def test_restored_identity_cannot_recover_same_qualification(self):
        observer = FakeObserver(deltas=(1, 1, 1))
        calls = 0
        def identity():
            nonlocal calls
            calls += 1
            return {"binary": "changed" if calls == 8 else "original"}
        observer.identity = identity
        report = self.evaluate(observer)
        self.assertEqual(report["outcome"], "INCONCLUSIVE")
        self.assertEqual(len(report["sessions"]), 1)
        self.assertEqual(calls, 8)
        self.assertNotIn("aggregate_delta_ms", report)

    def test_missing_or_busy_witness_stops_before_launch(self):
        for valid, jobs in [(False, []), (True, ["owner-build"] )]:
            observer = FakeObserver()
            original = observer.checkpoint
            observer.checkpoint = lambda label: {**original(label), "own_heavy_witness_valid": valid,
                                                  "active_heavy_jobs": jobs}
            report = self.evaluate(observer)
            self.assertEqual(report["outcome"], "INCONCLUSIVE")
            self.assertEqual(observer.calls, 0)

    def test_real_witness_expiry_owner_and_execution_binding(self):
        with tempfile.TemporaryDirectory() as directory:
            path = Path(directory) / "witness.json"
            observer = cli.ProtocolObserver.__new__(cli.ProtocolObserver)
            observer.args = type("Args", (), {"own_heavy_state_file": path, "own_heavy_owner": "codex/mbp"})
            observer.execution_id = None
            base = {"schema_version": 1, "owner": "codex/mbp", "execution_id": "measurement-1",
                    "active_heavy_jobs": [], "valid_until_epoch": 100}
            with patch.object(cli.time, "time", return_value=100):
                self.assertFalse(observer.checkpoint("missing")["own_heavy_witness_valid"])
                for changes, expected in [({}, True), ({"valid_until_epoch": 99}, False),
                                           ({"owner": "other"}, False), ({"execution_id": "other"}, False)]:
                    path.write_text(json.dumps({**base, **changes}))
                    self.assertEqual(observer.checkpoint("test")["own_heavy_witness_valid"], expected)

    def test_strict_build_memory_boundaries_and_atomic_partial(self):
        observer = FakeObserver()
        original = observer.elapsed
        observer.elapsed = lambda name: 50 if name == "build_10_modules" else original(name)
        observer.memory = lambda name: {"peak_footprint_bytes": 10_000_000, "maximum_rss_bytes": 1}
        report = self.evaluate(observer)
        self.assertFalse(report["criteria"]["build"])
        self.assertFalse(report["criteria"]["memory"])
        with tempfile.TemporaryDirectory() as directory:
            path = Path(directory) / "result.json"
            cli.save_report(path, report)
            self.assertEqual(json.loads(path.read_text()), report)
            self.assertFalse(path.with_name("result.json.partial").exists())

    def test_release_preflight_requires_hash_compiler_and_source(self):
        with tempfile.TemporaryDirectory() as directory:
            binary = Path(directory) / "binary"
            binary.write_bytes(b"public release fixture")
            args = SimpleNamespace(expected_binary_sha256=cli.digest(binary), expected_source_sha="a" * 40,
                release_version="v0.6.0", own_heavy_state_file=Path(directory) / "witness",
                own_heavy_owner="codex/mbp", samples=30, fresh_copies=5)
            info = {"raw": "mod\tgithub.com/gizzahub/gzh-cli-shellforge\tv0.6.0\t\n",
                    "compiler": "go1.27.3", "settings": {"GOOS": "darwin", "GOARCH": "arm64",
                    "CGO_ENABLED": "0", "-trimpath": "true", "vcs.modified": "false", "vcs.revision": "a" * 40}}
            def metadata(command):
                return SimpleNamespace(stdout=b"go version go1.27.3 darwin/arm64\n" if command[1] == "version" else b"darwin\narm64\n")
            with patch.object(cli, "build_info", return_value=info), patch.object(cli, "run", side_effect=metadata), patch.object(cli.platform, "machine", return_value="arm64"):
                cli.qualify_inputs(args, binary)
                self.assertEqual(args.control_ldflags, "-s -w")
                for field, wrong in [("expected_binary_sha256", "b" * 64), ("expected_source_sha", "b" * 40), ("release_version", "v0.6.0-SNAPSHOT")]:
                    old = getattr(args, field)
                    setattr(args, field, wrong)
                    with self.assertRaises(ValueError):
                        cli.qualify_inputs(args, binary)
                    setattr(args, field, old)
                info["compiler"] = "go1.26.7"
                with self.assertRaisesRegex(ValueError, "compiler"):
                    cli.qualify_inputs(args, binary)

    def test_subprocess_failure_retains_status_output_and_elapsed(self):
        error = subprocess.CalledProcessError(23, ["fixture"], output=b"partial output", stderr=b"diagnostic")
        error.elapsed_ms = 7
        record = cli.ProtocolObserver.failure(None, error)
        self.assertEqual(record["returncode"], 23)
        self.assertEqual(record["stdout"], "partial output")
        self.assertEqual(record["stderr"], "diagnostic")
        self.assertEqual(record["elapsed_ms"], 7)

    def test_preflight_architecture_aliases_and_wrong_targets(self):
        cases = [
            ("x86_64", "amd64", "amd64", True),
            ("amd64", "amd64", "amd64", True),
            ("arm64", "arm64", "arm64", True),
            ("aarch64", "arm64", "arm64", True),
            ("x86_64", "arm64", "arm64", False),
            ("arm64", "amd64", "amd64", False),
            ("x86_64", "amd64", "arm64", False),
            ("unknown", "arm64", "arm64", False),
            ("mips", "mips", "mips", False),
        ]
        with tempfile.TemporaryDirectory() as directory:
            binary = Path(directory) / "release"
            binary.write_bytes(b"verified fixture bytes")
            for host, control, release, accepted in cases:
                with self.subTest(host=host, control=control, release=release):
                    args = SimpleNamespace(expected_binary_sha256=cli.digest(binary), expected_source_sha="a" * 40,
                        release_version="v0.6.0", own_heavy_state_file=Path(directory) / "witness",
                        own_heavy_owner="codex/mbp", samples=30, fresh_copies=5)
                    info = {"raw": "mod\tgithub.com/gizzahub/gzh-cli-shellforge\tv0.6.0\t\n",
                            "compiler": "go1.27.3", "settings": {"GOOS": "darwin", "GOARCH": release,
                            "CGO_ENABLED": "0", "-trimpath": "true", "vcs.modified": "false", "vcs.revision": "a" * 40}}
                    def metadata(command):
                        return SimpleNamespace(stdout=("go version go1.27.3 darwin/" + control + "\n" if command[1] == "version" else "darwin\n" + control + "\n").encode())
                    with patch.object(cli, "build_info", return_value=info), patch.object(cli, "run", side_effect=metadata), patch.object(cli.platform, "machine", return_value=host):
                        if accepted:
                            cli.qualify_inputs(args, binary)
                        else:
                            with self.assertRaises(ValueError):
                                cli.qualify_inputs(args, binary)

    def test_memory_parse_failure_keeps_successful_command_output(self):
        result = subprocess.CompletedProcess(["time"], 0, b"product stdout", b"42 maximum resident set size\n")
        with patch.object(cli, "run", return_value=result):
            with self.assertRaisesRegex(RuntimeError, "peak memory footprint") as caught:
                cli.memory(["fixture"], Path("."))
        record = cli.ProtocolObserver.failure(None, caught.exception)
        self.assertEqual(record["stdout"], "product stdout")
        self.assertIn("maximum resident", record["stderr"])
        self.assertEqual(record["returncode"], 0)

    def test_release_swap_before_new_observer_cannot_rebind_baseline(self):
        with tempfile.TemporaryDirectory() as directory:
            binary, helper = Path(directory) / "release", Path(directory) / "control"
            binary.write_bytes(b"verified public bytes")
            helper.write_bytes(b"matched control")
            args = SimpleNamespace(expected_binary_sha256=cli.digest(binary),
                expected_control_sha256=cli.digest(helper), expected_harness_sha256={}, witness_execution_id="test")
            binary.write_bytes(b"different bytes after control build")
            report = {}
            status = recorded_execution(lambda: cli.ProtocolObserver(args, Path(directory), binary, helper, Path(directory), {}),
                report, lambda r: None, lambda e: cli.ProtocolObserver.failure(None, e))
            self.assertEqual(status, 1)
            self.assertEqual(report["outcome"], "INCONCLUSIVE")

    def test_identity_setup_gap_and_interrupt_are_persisted(self):
        for stage in ("identity", "gap", "interrupt"):
            observer = FakeObserver(deltas=(1, 1, 1))
            if stage == "identity":
                def fail():
                    raise OSError("identity unavailable")
                observer.identity = fail
            elif stage == "gap":
                def fail(seconds):
                    raise RuntimeError("gap interrupted")
                observer.wait_gap = fail
            else:
                original = observer.elapsed
                def interrupt(name):
                    if observer.calls >= 6:
                        raise KeyboardInterrupt()
                    return original(name)
                observer.elapsed = interrupt
            report = self.evaluate(observer)
            self.assertEqual(report["outcome"], "INCONCLUSIVE")
            self.assertIn("fatal_failure", observer.saved[-1])
            if stage == "interrupt":
                self.assertTrue(report["sessions"][0]["pairs"])
                self.assertIn("KeyboardInterrupt", report["sessions"][0]["invalid_reasons"][-1])

    def test_timeout_stops_own_descendant_group_and_keeps_partial_output(self):
        code = "import subprocess,sys,time; subprocess.Popen([sys.executable,'-c','import time; time.sleep(30)']); print('started',flush=True); time.sleep(30)"
        with self.assertRaises(subprocess.TimeoutExpired) as caught:
            process_run([sys.executable, "-c", code], timeout=.3)
        self.assertIn(b"started", caught.exception.stdout)
        self.assertFalse(group_running(caught.exception.cleanup["pgid"]))
        self.assertFalse(caught.exception.cleanup["running_members_after"])

    def test_keyboard_interrupt_reaps_owned_child(self):
        original = subprocess.Popen.communicate
        interrupted = False
        command = [sys.executable, "-c", "import time; time.sleep(30)"]
        def communicate(process, *args, **kwargs):
            nonlocal interrupted
            if process.args == command and not interrupted:
                interrupted = True
                raise KeyboardInterrupt()
            return original(process, *args, **kwargs)
        with patch.object(subprocess.Popen, "communicate", communicate):
            with self.assertRaises(KeyboardInterrupt) as caught:
                process_run(command)
        self.assertFalse(group_running(caught.exception.cleanup["pgid"]))

    def test_permission_probe_never_implies_group_absence(self):
        observation = []
        with patch.object(startup_process.os, "killpg", side_effect=PermissionError("probe denied")):
            with patch.object(startup_process.subprocess, "run", return_value=subprocess.CompletedProcess([], 0, b"123 R\n", b"")):
                self.assertTrue(group_running(123, observation))
                self.assertIn("probe denied", observation[-1]["signal_probe_error"])
            with patch.object(startup_process.subprocess, "run", side_effect=subprocess.CalledProcessError(1, ["ps"])):
                with self.assertRaises(subprocess.CalledProcessError):
                    group_running(123)

    def test_cleanup_delay_is_verified_but_not_scored(self):
        clock = [0]
        class Child:
            pid, returncode, calls = 123, 0, 0
            def communicate(self, **kwargs):
                self.calls += 1
                if self.calls == 1:
                    clock[0] += 2_000_000
                return b"version output", b""
        child = Child()
        def cleanup(process):
            self.assertIs(process, child)
            clock[0] += 300_000_000
            return {"pgid": 123, "running_members_after": False}
        with patch.object(startup_process.time, "perf_counter_ns", side_effect=lambda: clock[0]), patch.object(startup_process.subprocess, "Popen", return_value=child), patch.object(startup_process, "stop_group", side_effect=cleanup) as cleaned:
            result = process_run(["fixture"])
        self.assertEqual(result.process_elapsed_ms, 2)
        self.assertEqual(result.cleanup_elapsed_ms, 300)
        cleaned.assert_called_once_with(child)
        self.assertEqual(child.calls, 2)
        observations = []
        with patch.object(cli, "run", return_value=result):
            self.assertEqual(cli.elapsed(["fixture"], Path("."), observations), 2)
        self.assertEqual(observations[0]["cleanup_elapsed_ms_unscored"], 300)

    def test_nonempty_malformed_ps_rows_are_unknown_not_absent(self):
        with patch.object(startup_process.os, "killpg", return_value=None):
            for raw in [b"123\n", b"123 R extra\n", b"invalid R\n", b"123 nonsense\n"]:
                with patch.object(startup_process.subprocess, "run", return_value=subprocess.CompletedProcess([], 0, raw, b"")):
                    with self.assertRaises(startup_process.LifecycleError):
                        group_running(123)
            for raw in [b"", b"456 R\n", b"123 Z\n"]:
                with patch.object(startup_process.subprocess, "run", return_value=subprocess.CompletedProcess([], 0, raw, b"")):
                    self.assertFalse(group_running(123))

    def test_nonfinite_witness_remains_serializable_invalid_evidence(self):
        with tempfile.TemporaryDirectory() as directory:
            path = Path(directory) / "witness.json"
            args = SimpleNamespace(own_heavy_state_file=path, own_heavy_owner="codex/mbp")
            for constant in ["NaN", "Infinity", "-Infinity", "1e999"]:
                raw = '{"schema_version":1,"valid_until_epoch":' + constant + '}'
                path.write_text(raw)
                state = cli.scoped_witness(args)
                self.assertFalse(state["own_heavy_witness_valid"])
                self.assertEqual(state["invalid_witness_text"], raw)
                observer = FakeObserver()
                original = observer.checkpoint
                observer.checkpoint = lambda label: {**original(label), **state}
                report = self.evaluate(observer)
                self.assertEqual(report["outcome"], "INCONCLUSIVE")
                output = Path(directory) / "result.json"
                cli.save_report(output, report)
                self.assertEqual(json.loads(output.read_text()), report)


if __name__ == "__main__":
    unittest.main()
