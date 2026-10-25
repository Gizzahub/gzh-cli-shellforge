#!/usr/bin/env python3
"""Observe local macOS launch cost without claiming a cold-cache acceptance gate."""
import argparse
import hashlib
import json
import math
import os
from pathlib import Path
import platform
import re
import shutil
import statistics
import subprocess
import tempfile
import time

from startup_process import run
from startup_protocol import ProvenanceError, collect, recorded_execution


def lifecycle_observation(result, observations):
    if observations is not None:
        observations.append({"process_elapsed_ms": result.process_elapsed_ms,
                             "cleanup_elapsed_ms_unscored": result.cleanup_elapsed_ms,
                             "cleanup": result.cleanup})


def elapsed(args, cwd, observations=None):
    result = run(args, cwd=cwd)
    lifecycle_observation(result, observations)
    return result.process_elapsed_ms


def distribution(samples):
    ordered = sorted(samples)
    return {"samples_ms": samples, "min_ms": min(samples),
            "median_ms": statistics.median(samples),
            "p95_nearest_rank_ms": ordered[math.ceil(.95 * len(samples)) - 1],
            "max_ms": max(samples)}


def memory(args, cwd, observations=None):
    result = run(["/usr/bin/time", "-l", *args], cwd=cwd)
    lifecycle_observation(result, observations)
    text = result.stderr.decode()
    metrics = {}
    for label, key in [("maximum resident set size", "maximum_rss_bytes"),
                       ("peak memory footprint", "peak_footprint_bytes")]:
        match = re.search(r"^\s*(\d+)\s+" + label + r"$", text, re.MULTILINE)
        if not match:
            error = RuntimeError("/usr/bin/time did not report " + label)
            error.stdout, error.stderr = result.stdout, result.stderr
            error.returncode = result.returncode
            raise error
        metrics[key] = int(match[1])
    return metrics


def digest(path):
    return hashlib.sha256(path.read_bytes()).hexdigest()


def build_info(binary):
    text = run(["go", "version", "-m", str(binary)]).stdout.decode()
    settings = {}
    for line in text.splitlines():
        fields = line.strip().split("\t")
        if len(fields) > 1 and fields[0] == "build" and "=" in fields[1]:
            key, value = fields[1].split("=", 1)
            settings[key] = value
    return {"raw": text, "compiler": text.splitlines()[0].rsplit(": ", 1)[1],
            "settings": settings}


def qualify_inputs(args, binary):
    for field in ("expected_binary_sha256", "expected_source_sha"):
        if not re.fullmatch(r"[0-9a-f]{64}" if field.endswith("sha256") else r"[0-9a-f]{40}",
                            getattr(args, field) or ""):
            raise ValueError("qualification requires verified " + field)
    if not args.release_version or not re.fullmatch(r"v\d+\.\d+\.\d+", args.release_version):
        raise ValueError("qualification requires a stable --release-version")
    if not args.own_heavy_state_file or not args.own_heavy_owner:
        raise ValueError("qualification requires a scoped own-heavy witness and owner")
    if args.samples != 30 or args.fresh_copies != 5:
        raise ValueError("qualification fixes samples=30 and fresh-copies=5")
    info = build_info(binary)
    settings = info["settings"]
    compiler = run(["go", "version"]).stdout.decode().split()[2]
    target = run(["go", "env", "GOOS", "GOARCH"]).stdout.decode().splitlines()
    host_arch = {"x86_64": "amd64", "amd64": "amd64", "aarch64": "arm64", "arm64": "arm64"}.get(platform.machine())
    if host_arch is None:
        raise ValueError("unsupported Darwin host architecture")
    expected = {"GOOS": "darwin", "GOARCH": host_arch,
                "CGO_ENABLED": "0", "-trimpath": "true", "vcs.revision": args.expected_source_sha,
                "vcs.modified": "false"}
    if digest(binary) != args.expected_binary_sha256 or info["compiler"] != compiler:
        raise ValueError("released binary hash or control compiler does not match")
    if target != [expected["GOOS"], expected["GOARCH"]]:
        raise ValueError("control target differs from host/release target")
    if any(settings.get(k) != v for k, v in expected.items()):
        raise ValueError("released binary build settings/source do not match")
    module = "mod\tgithub.com/gizzahub/gzh-cli-shellforge\t" + args.release_version + "\t"
    if module not in info["raw"]:
        raise ValueError("binary is not the declared stable module release")
    args.control_trimpath, args.control_ldflags, args.control_cgo = True, "-s -w", "0"
    return info


def scoped_witness(args, execution_id=None):
    witness, error, raw = None, None, None
    def finite_number(value):
        number = float(value)
        if not math.isfinite(number):
            raise ValueError("nonfinite witness number: " + value)
        return number

    def reject_constant(value):
        raise ValueError("nonstandard witness JSON constant: " + value)

    try:
        raw = args.own_heavy_state_file.read_text()
        witness = json.loads(raw, parse_constant=reject_constant, parse_float=finite_number)
        valid = (type(witness["schema_version"]) is int and witness["schema_version"] == 1
                 and witness["owner"] == args.own_heavy_owner
                 and isinstance(witness["execution_id"], str) and bool(witness["execution_id"].strip())
                 and (execution_id is None or execution_id == witness["execution_id"])
                 and type(witness["valid_until_epoch"]) in (int, float)
                 and math.isfinite(witness["valid_until_epoch"])
                 and time.time() <= witness["valid_until_epoch"]
                 and isinstance(witness["active_heavy_jobs"], list)
                 and all(isinstance(j, str) for j in witness["active_heavy_jobs"]))
    except (OSError, ValueError, KeyError, TypeError, OverflowError) as failure:
        valid, error, witness = False, str(failure), None
    return {"own_heavy_witness": witness, "own_heavy_witness_error": error,
            "invalid_witness_text": raw if witness is None else None,
            "own_heavy_witness_valid": valid,
            "active_heavy_jobs": witness.get("active_heavy_jobs", []) if isinstance(witness, dict) else []}


class ProtocolObserver:
    def __init__(self, args, root, binary, helper, work, commands):
        self.args, self.root, self.binary = args, root, binary
        self.helper, self.work, self.commands = helper, work, commands
        self.baseline = self.identity()
        self.execution_id = args.witness_execution_id
        self.process_observations = []

    def identity(self):
        binary_sha, control_sha = digest(self.binary), digest(self.helper)
        if binary_sha != self.args.expected_binary_sha256 or control_sha != self.args.expected_control_sha256:
            raise ProvenanceError("release/control bytes changed from verified build inputs")
        if any(digest(Path(path)) != sha for path, sha in self.args.expected_harness_sha256.items()):
            raise ProvenanceError("harness bytes changed after invocation began")
        return {"binary_sha256": binary_sha, "control_sha256": control_sha,
                "source_head": run(["git", "-C", str(self.root), "rev-parse", "HEAD"]).stdout.decode().strip(),
                "tracked_worktree_diff_sha256": hashlib.sha256(run(
                    ["git", "-C", str(self.root), "diff", "--binary", "HEAD"]).stdout).hexdigest(),
                "harness_sha256": digest(Path(__file__)),
                "protocol_sha256": digest(Path(__file__).with_name("startup_protocol.py")),
                "process_helper_sha256": digest(Path(__file__).with_name("startup_process.py")),
                "control_compiler": build_info(self.helper)["compiler"]}

    def checkpoint(self, label):
        state = scoped_witness(self.args, self.execution_id)
        if state["own_heavy_witness_valid"] and self.execution_id is None:
            self.execution_id = state["own_heavy_witness"]["execution_id"]
        return {"label": label, "epoch": time.time(), "load": list(os.getloadavg()),
                "logical_cpus": os.cpu_count() or 0, **state}

    def wait_gap(self, seconds):
        time.sleep(seconds)

    def elapsed(self, name):
        return elapsed(self.commands[name], self.work, self.process_observations)

    def memory(self, name):
        return memory(self.commands[name], self.work, self.process_observations)

    def fresh(self, name, attempt, pair):
        source = self.binary if name == "version" else self.helper
        copy = self.work / f"fresh-{attempt}-{pair}-{name}"
        shutil.copyfile(source, copy)
        copy.chmod(0o700)
        return elapsed([str(copy), *(["--version"] if name == "version" else [])], self.work,
                       self.process_observations)

    def failure(self, error):
        return {"type": type(error).__name__, "message": str(error),
                "returncode": getattr(error, "returncode", None),
                "elapsed_ms": getattr(error, "elapsed_ms", None),
                "cleanup_elapsed_ms_unscored": getattr(error, "cleanup_elapsed_ms", None),
                "cleanup": getattr(error, "cleanup", None),
                "command": getattr(error, "cmd", None),
                "cause": ProtocolObserver.failure(None, error.__cause__) if error.__cause__ else None,
                "stdout": (getattr(error, "stdout", None) or b"").decode(errors="replace"),
                "stderr": (getattr(error, "stderr", None) or b"").decode(errors="replace")}


def save_report(path, report):
    path.parent.mkdir(parents=True, exist_ok=True)
    temporary = path.with_name(path.name + ".partial")
    temporary.write_text(json.dumps(report, indent=2, allow_nan=False) + "\n")
    temporary.replace(path)


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--binary", type=Path, default=Path("build/gz-shellforge"))
    parser.add_argument("--label", default="local Make development build")
    parser.add_argument("--control-trimpath", action="store_true")
    parser.add_argument("--control-ldflags", default="")
    parser.add_argument("--control-cgo", choices=["0", "1"])
    parser.add_argument("--samples", type=int, default=30)
    parser.add_argument("--fresh-copies", type=int, default=5)
    parser.add_argument("--output", type=Path, required=True)
    parser.add_argument("--qualify", action="store_true", help="adopted three-session release qualification")
    parser.add_argument("--preflight-only", action="store_true", help="qualification metadata only; do not build or measure")
    parser.add_argument("--expected-binary-sha256")
    parser.add_argument("--expected-source-sha")
    parser.add_argument("--release-version")
    parser.add_argument("--own-heavy-state-file", type=Path)
    parser.add_argument("--own-heavy-owner")
    args = parser.parse_args()
    if platform.system() != "Darwin":
        parser.error("this workload uses macOS /usr/bin/time -l metrics")
    if args.samples < 1 or args.fresh_copies < 1:
        parser.error("sample counts must be positive")
    if args.preflight_only and not args.qualify:
        parser.error("--preflight-only requires --qualify")
    report = {}
    return recorded_execution(lambda: profile(args, parser, report), report,
                              lambda value: save_report(args.output, value),
                              lambda error: ProtocolObserver.failure(None, error))


def profile(args, parser, report):
    root = Path(__file__).resolve().parent.parent
    args.expected_harness_sha256 = {str(root / "scripts" / name): digest(root / "scripts" / name)
                                   for name in ("startup-profile.py", "startup_protocol.py", "startup_process.py")}
    report["invocation_harness_sha256"] = args.expected_harness_sha256
    binary = args.binary.resolve(strict=True)
    attrs = run(["xattr", str(binary)]).stdout.decode().splitlines()
    if "com.apple.quarantine" in attrs:
        parser.error(
            "binary has com.apple.quarantine; use a local build or a release fetched with "
            "`gh release download` or `curl` (quarantine is never removed by this tool)"
        )
    if args.qualify:
        try:
            product_info = qualify_inputs(args, binary)
        except (ValueError, subprocess.SubprocessError) as error:
            save_report(args.output, {"schema_version": 2, "outcome": "INCONCLUSIVE",
                                     "stage": "preflight", "error": str(error), "sessions": []})
            parser.error(str(error))
        if args.preflight_only:
            save_report(args.output, {"schema_version": 2, "outcome": "NOT_MEASURED",
                "stage": "release-metadata-preflight", "binary_sha256": digest(binary),
                "release_buildinfo": product_info, "control_flags": ["-trimpath", "-ldflags", "-s -w", "CGO_ENABLED=0"],
                "limits": "No control built, witness validated or product launched; expected SHA/source must come from verified public release bytes."})
            return 0
    tmp = root / "tmp"
    tmp.mkdir(exist_ok=True)
    report.update({
        "schema_version": 1,
        "binary_label": args.label,
        "timestamp_utc": time.strftime("%Y-%m-%dT%H:%M:%SZ", time.gmtime()),
        "platform": run(["sw_vers"]).stdout.decode().strip(),
        "cpu": run(["sysctl", "-n", "machdep.cpu.brand_string"]).stdout.decode().strip(),
        "python": platform.python_version(),
        "go": run(["go", "version"]).stdout.decode().strip(),
        "source_head": run(["git", "-C", str(root), "rev-parse", "HEAD"]).stdout.decode().strip(),
        "binary_sha256": hashlib.sha256(binary.read_bytes()).hexdigest(),
        "binary_bytes": binary.stat().st_size,
        "binary_buildinfo": run(["go", "version", "-m", str(binary)]).stdout.decode(),
        "binary_xattr_names": attrs,
        "definitions": {
            "elapsed": "perf_counter_ns before Popen through communicate completion; includes process creation and output capture; post-exit cleanup is separately recorded and unscored",
            "warm": "same executable path repeatedly launched; no reboot or cache purge",
            "fresh_copy": "first launch of a new local copy/inode; NOT proven cold filesystem, dyld or security cache",
            "memory": "separate /usr/bin/time -l processes; RSS and peak footprint are distinct macOS byte metrics",
            "acceptance": "observations only; no threshold or metric reinterpretation"}})
    if args.qualify:
        report["binary_label"] = "verified public release " + args.release_version
        report["definitions"]["acceptance"] = "adopted schema2 measurement verdict; final source delivery still requires owner gates"
        state = scoped_witness(args)
        report["pre_control_build_witness"] = state
        if not state["own_heavy_witness_valid"] or state["active_heavy_jobs"]:
            save_report(args.output, {**report, "schema_version": 2, "outcome": "INCONCLUSIVE",
                                     "stage": "own-heavy-admission", "sessions": []})
            return 1
        args.witness_execution_id = state["own_heavy_witness"]["execution_id"]
    # Every path below belongs to this invocation; no user dotfile is read or deployed.
    with tempfile.TemporaryDirectory(prefix="startup-profile-", dir=tmp) as directory:
        work = Path(directory)
        helper_source = work / "empty.go"
        helper_source.write_text("package main\nfunc main() {}\n")
        helper = work / "empty-go"
        helper_args = ["go", "build"]
        if args.control_trimpath:
            helper_args.append("-trimpath")
        if args.control_ldflags:
            helper_args.extend(["-ldflags", args.control_ldflags])
        helper_args.extend(["-o", str(helper), str(helper_source)])
        helper_env = os.environ.copy()
        if args.control_cgo is not None:
            helper_env["CGO_ENABLED"] = args.control_cgo
        report["stage"] = "control-build"
        try:
            run(helper_args, cwd=root, env=helper_env)
        except subprocess.SubprocessError as error:
            if args.qualify:
                save_report(args.output, {**report, "schema_version": 2, "outcome": "INCONCLUSIVE",
                    "stage": "control-build", "error": str(error), "command": helper_args,
                    "stderr": (getattr(error, "stderr", None) or b"").decode(errors="replace"),
                    "sessions": []})
            raise
        args.expected_control_sha256 = digest(helper)
        report["control_build_command"] = helper_args
        report["control_buildinfo"] = run(["go", "version", "-m", str(helper)]).stdout.decode()
        report["control_cgo_override"] = args.control_cgo
        if args.qualify:
            control = build_info(helper)
            product = build_info(binary)
            if control["compiler"] != product["compiler"] or any(
                control["settings"].get(k) != product["settings"].get(k)
                for k in ("GOOS", "GOARCH", "CGO_ENABLED", "-trimpath")):
                save_report(args.output, {**report, "schema_version": 2, "outcome": "INCONCLUSIVE",
                    "stage": "control-build", "error": "built control provenance does not match release", "sessions": []})
                raise ValueError("built control provenance does not match release")
        manifest = work / "manifest.yaml"
        manifest.write_text("modules:\n" + "".join(
            f"  - name: module{i}\n    file: module{i}.sh\n    target: zshrc\n" for i in range(10)))
        for i in range(10):
            (work / f"module{i}.sh").write_text(f"export MODULE_{i}=ready\n")
        commands = {"version": [str(binary), "--version"],
                    "help": [str(binary), "--help"],
                    "build_10_modules": [str(binary), "build", "--manifest", str(manifest),
                                         "--config-dir", str(work), "--output-dir", str(work / "build")],
                    "empty_go_control": [str(helper)]}
        if args.qualify:
            report["qualification"] = {"startup_budget_ms": 3, "load_per_cpu_ceiling": 1,
                "warm_pairs": 30, "valid_sessions": 3, "max_attempts": 5, "gap_seconds": 60,
                "release_version": args.release_version, "expected_binary_sha256": args.expected_binary_sha256,
                "expected_source_sha": args.expected_source_sha,
                "control_ldflags": args.control_ldflags,
                "strip_provenance": "expected public artifact SHA binds release -s -w flags; supplied by operator's verified release manifest",
                "own_heavy_scope": "parent-scoped witness only; not global job detection; measurement itself excluded"}
            observer = ProtocolObserver(args, root, binary, helper, work, commands)
            report["process_lifecycle_unscored"] = observer.process_observations
            report["stage"] = "sessions"
            collect(observer, report, lambda value: save_report(args.output, value))
            print("Qualification:", report["outcome"], "observations:", args.output)
            return 0 if report["outcome"] == "PASS" else 1
        report["workloads"] = {}
        for name, command in commands.items():
            first = elapsed(command, work)
            samples = [elapsed(command, work) for _ in range(args.samples)]
            report["workloads"][name] = {"first_observed_ms": first,
                "warm": distribution(samples),
                "memory_samples": [memory(command, work) for _ in range(5)]}
        fresh = {"shellforge_version": [], "empty_go_control": []}
        for i in range(args.fresh_copies):
            # Alternate which control starts first; retain all samples including outliers.
            cases = [("shellforge_version", binary, ["--version"]),
                     ("empty_go_control", helper, [])]
            for name, source, flags in cases if i % 2 == 0 else reversed(cases):
                copy = work / f"fresh-{name}-{i}"
                shutil.copyfile(source, copy)
                copy.chmod(0o700)
                fresh[name].append(elapsed([str(copy), *flags], work))
        report["fresh_copy_first_launch"] = {name: distribution(values) for name, values in fresh.items()}
        env = os.environ.copy()
        env["GODEBUG"] = "inittrace=1"
        trace = run(commands["version"], cwd=work, env=env).stderr.decode()
        report["inittrace"] = trace
        # Print the exact scope before TemporaryDirectory removes only its own fixture.
        report["temporary_fixture_cleanup"] = {"path": str(work),
            "owned_entries": sorted(str(p.relative_to(work)) for p in work.rglob("*"))}
    args.output.parent.mkdir(parents=True, exist_ok=True)
    args.output.write_text(json.dumps(report, indent=2) + "\n")
    print(report["timestamp_utc"], report["cpu"], report["go"])
    for name, values in report["workloads"].items():
        warm = values["warm"]
        print(f"{name}: first={values['first_observed_ms']:.3f}ms "
              f"warm median={warm['median_ms']:.3f} p95={warm['p95_nearest_rank_ms']:.3f} "
              f"max={warm['max_ms']:.3f}ms")
    for name, values in report["fresh_copy_first_launch"].items():
        print(f"fresh local copy {name}: {values['samples_ms']} ms")
    print("Full observations:", args.output)


if __name__ == "__main__":
    raise SystemExit(main())
