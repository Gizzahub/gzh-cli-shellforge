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


def run(args, **kwargs):
    return subprocess.run(args, check=True, stdout=subprocess.PIPE,
                          stderr=subprocess.PIPE, timeout=60, **kwargs)


def elapsed(args, cwd):
    start = time.perf_counter_ns()
    run(args, cwd=cwd)
    return (time.perf_counter_ns() - start) / 1_000_000


def distribution(samples):
    ordered = sorted(samples)
    return {"samples_ms": samples, "min_ms": min(samples),
            "median_ms": statistics.median(samples),
            "p95_nearest_rank_ms": ordered[math.ceil(.95 * len(samples)) - 1],
            "max_ms": max(samples)}


def memory(args, cwd):
    result = run(["/usr/bin/time", "-l", *args], cwd=cwd)
    text = result.stderr.decode()
    metrics = {}
    for label, key in [("maximum resident set size", "maximum_rss_bytes"),
                       ("peak memory footprint", "peak_footprint_bytes")]:
        match = re.search(r"^\s*(\d+)\s+" + label + r"$", text, re.MULTILINE)
        if not match:
            raise RuntimeError("/usr/bin/time did not report " + label)
        metrics[key] = int(match[1])
    return metrics


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--binary", type=Path, default=Path("build/gz-shellforge"))
    parser.add_argument("--samples", type=int, default=30)
    parser.add_argument("--fresh-copies", type=int, default=5)
    parser.add_argument("--output", type=Path, required=True)
    args = parser.parse_args()
    if platform.system() != "Darwin":
        parser.error("this workload uses macOS /usr/bin/time -l metrics")
    if args.samples < 1 or args.fresh_copies < 1:
        parser.error("sample counts must be positive")
    root = Path(__file__).resolve().parent.parent
    binary = args.binary.resolve(strict=True)
    attrs = run(["xattr", str(binary)]).stdout.decode().splitlines()
    if "com.apple.quarantine" in attrs:
        parser.error("use a locally built binary; quarantine is never removed by this tool")
    tmp = root / "tmp"
    tmp.mkdir(exist_ok=True)
    report = {
        "schema_version": 1,
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
            "elapsed": "perf_counter_ns around subprocess.run through process exit; includes process creation and output capture",
            "warm": "same executable path repeatedly launched; no reboot or cache purge",
            "fresh_copy": "first launch of a new local copy/inode; NOT proven cold filesystem, dyld or security cache",
            "memory": "separate /usr/bin/time -l processes; RSS and peak footprint are distinct macOS byte metrics",
            "acceptance": "observations only; no threshold or metric reinterpretation"}}
    # Every path below belongs to this invocation; no user dotfile is read or deployed.
    with tempfile.TemporaryDirectory(prefix="startup-profile-", dir=tmp) as directory:
        work = Path(directory)
        helper_source = work / "empty.go"
        helper_source.write_text("package main\nfunc main() {}\n")
        helper = work / "empty-go"
        run(["go", "build", "-trimpath", "-o", str(helper), str(helper_source)], cwd=root)
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
    main()
