"""Deterministic orchestration for the adopted macOS startup protocol.

The observer supplies measurements, provenance and load observations. No result
is discarded: invalid attempts and incomplete command output remain in report.
"""
import statistics

from startup_process import LifecycleError


class ProvenanceError(RuntimeError):
    """A qualification input changed; the entire batch is invalid."""


def recorded_execution(action, report, persist, failure_record):
    """Persist setup, gap and interruption failures without losing raw sessions."""
    try:
        return action()
    except (Exception, KeyboardInterrupt) as error:
        report.update(schema_version=2, outcome="INCONCLUSIVE", fatal_failure=failure_record(error))
        if report.get("sessions"):
            session = report["sessions"][-1]
            # A gap failure must not relabel the already completed prior session.
            if not session["valid"]:
                session["invalid_reasons"].append(type(error).__name__ + ": " + str(error))
        persist(report)
        return 1


def verified_identity(observer):
    try:
        return observer.identity()
    except Exception as error:
        raise ProvenanceError("cannot verify qualification identity") from error


def pair_order(attempt, pair):
    return ("version", "empty_go_control") if (attempt + pair) % 2 == 0 else (
        "empty_go_control", "version")


def summarize(report):
    valid = [s for s in report["sessions"] if s["valid"]][:3]
    report["selected_attempts"] = [s["attempt"] for s in valid]
    if len(valid) < 3:
        report["outcome"] = "INCONCLUSIVE"
        return
    deltas = [s["startup_delta_ms"] for s in valid]
    report["aggregate_delta_ms"] = statistics.median(deltas)
    report["criteria"] = {
        "startup": report["aggregate_delta_ms"] <= 3,
        "build": all(statistics.median(s["warm"]["build_10_modules"]) < 50 for s in valid),
        "memory": max(m["peak_footprint_bytes"] for s in valid
                      for m in s["memory"]["version"]) < 10_000_000,
    }
    report["outcome"] = "PASS" if all(report["criteria"].values()) else "FAIL"


def collect(observer, report, persist):
    """At most five attempts; stop after first three valid sessions.

    observer.checkpoint must report load, logical CPUs, scoped heavy-job witness,
    and before/after provenance. Failures are retained, never retried silently.
    """
    report.update(schema_version=2, sessions=[], outcome="INCONCLUSIVE")
    report["identity"] = verified_identity(observer)
    persist(report)
    for attempt in range(5):
        if attempt:
            observer.wait_gap(60)
        session = {"attempt": attempt + 1, "valid": False, "invalid_reasons": [],
                   "checkpoints": [], "first_observed": {}, "pairs": [],
                   "warm": {k: [] for k in ("version", "empty_go_control", "help", "build_10_modules")},
                   "memory": {}, "fresh_copy": []}
        report["sessions"].append(session)

        def checkpoint(label):
            observation = observer.checkpoint(label)
            session["checkpoints"].append(observation)
            reasons = session["invalid_reasons"]
            cpus = observation["logical_cpus"]
            if cpus <= 0 or observation["load"][0] / cpus > 1:
                reasons.append("load ceiling exceeded at " + label)
            if not observation["own_heavy_witness_valid"] or observation["active_heavy_jobs"]:
                reasons.append("own-heavy witness unavailable, expired or busy at " + label)
            persist(report)
            if verified_identity(observer) != report["identity"]:
                raise ProvenanceError("provenance changed at " + label)
            if reasons:
                raise ValueError("checkpoint admission failed: " + label)

        try:
            before = verified_identity(observer)
            session["identity_before"] = before
            if before != report["identity"]:
                raise ProvenanceError("provenance changed between attempts")
            checkpoint("session-start")
            # First-observed remains visible; these are not cold-cache claims.
            for name in session["warm"]:
                checkpoint("first-observed-" + name)
                session["first_observed"][name] = observer.elapsed(name)
                persist(report)
            for pair in range(30):
                checkpoint("warm-pair-" + str(pair))
                item = {"order": list(pair_order(attempt, pair)), "values_ms": {}}
                session["pairs"].append(item)
                for name in item["order"]:
                    value = observer.elapsed(name)
                    item["values_ms"][name] = value
                    session["warm"][name].append(value)
                    persist(report)
            for name in ("help", "build_10_modules"):
                checkpoint("workload-" + name)
                for _ in range(30):
                    session["warm"][name].append(observer.elapsed(name))
                    persist(report)
            for name in session["warm"]:
                checkpoint("memory-" + name)
                session["memory"][name] = []
                for _ in range(5):
                    metric = observer.memory(name)
                    if any(metric[k] < 0 for k in ("peak_footprint_bytes", "maximum_rss_bytes")):
                        raise ValueError("negative memory observation")
                    session["memory"][name].append(metric)
                    persist(report)
            checkpoint("fresh-copy")
            for pair in range(5):
                item = {"order": list(pair_order(attempt, pair)), "values_ms": {}}
                session["fresh_copy"].append(item)
                for name in item["order"]:
                    item["values_ms"][name] = observer.fresh(name, attempt, pair)
                    persist(report)
            checkpoint("session-end")
            session["identity_after"] = verified_identity(observer)
            if before != session["identity_after"]:
                raise ProvenanceError("provenance changed during session")
            version = session["warm"]["version"]
            control = session["warm"]["empty_go_control"]
            session["startup_delta_ms"] = statistics.median(version) - statistics.median(control)
            session["paired_difference_median_ms_reference"] = statistics.median(
                p["values_ms"]["version"] - p["values_ms"]["empty_go_control"] for p in session["pairs"])
            session["valid"] = not session["invalid_reasons"]
        except (ProvenanceError, LifecycleError):
            raise
        except Exception as error:
            session["invalid_reasons"].append(type(error).__name__ + ": " + str(error))
            session["failure"] = observer.failure(error)
        summarize(report)
        persist(report)
        if len(report["selected_attempts"]) == 3:
            break
    return report
