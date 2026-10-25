"""Bounded child-session execution for macOS qualification and metadata reads."""
import os
import signal
import subprocess
import time


class LifecycleError(RuntimeError):
    """Cannot verify completion of an invocation's owned process group."""


def group_running(pgid, observations=None):
    probe_error = None
    try:
        os.killpg(pgid, 0)
    except ProcessLookupError:
        return False
    except PermissionError as error:
        # A just-exited group may reject signal probes. Inspect state; EPERM
        # alone never means either "running" or "gone".
        probe_error = str(error)
    # Orphaned zombies cannot run; report only this group's live members.
    states = subprocess.run(["ps", "-g", str(pgid), "-o", "pgid=,stat="], check=True,
                            stdout=subprocess.PIPE, stderr=subprocess.PIPE, timeout=5)
    own_states = []
    for line in states.stdout.decode().splitlines():
        if not line.strip():
            continue
        parts = line.split()
        if len(parts) != 2 or not parts[0].isdigit() or parts[1][0] not in "RSDITZUWtX":
            error = LifecycleError("malformed owned-group state query: " + repr(line))
            error.stdout, error.stderr, error.returncode = states.stdout, states.stderr, states.returncode
            raise error
        if int(parts[0]) == pgid:
            own_states.append(parts[1])
    if observations is not None:
        observations.append({"pgid": pgid, "signal_probe_error": probe_error,
                             "query_exit": states.returncode, "own_states": own_states})
    return any(not state.startswith("Z") for state in own_states)


def stop_group(process):
    actions = []
    observations = []
    for sig in (signal.SIGTERM, signal.SIGKILL):
        process.poll()  # Reap an exited direct child before probing a zombie-only group.
        if not group_running(process.pid, observations):
            break
        try:
            os.killpg(process.pid, sig)
            actions.append(sig.name)
        except ProcessLookupError:
            break
        deadline = time.monotonic() + 1
        while group_running(process.pid, observations) and time.monotonic() < deadline:
            process.poll()
            time.sleep(.01)
        process.poll()
        if not group_running(process.pid, observations):
            break
    if group_running(process.pid, observations):
        raise LifecycleError("owned process group still running: " + str(process.pid))
    return {"pgid": process.pid, "signals": actions, "running_members_after": False,
            "liveness_observations": observations}


def run(args, *, timeout=60, **kwargs):
    start = time.perf_counter_ns()
    process = subprocess.Popen(args, stdout=subprocess.PIPE, stderr=subprocess.PIPE,
                               start_new_session=True, **kwargs)
    error = None
    stdout, stderr = b"", b""
    try:
        stdout, stderr = process.communicate(timeout=timeout)
    except (Exception, KeyboardInterrupt) as failure:
        error = failure
        stdout = getattr(failure, "output", None) or stdout
        stderr = getattr(failure, "stderr", None) or stderr
    process_elapsed_ms = (time.perf_counter_ns() - start) / 1_000_000
    cleanup_start = time.perf_counter_ns()
    try:
        cleanup = stop_group(process)
        # This also reaps the direct child after an interrupted communicate.
        stdout, stderr = process.communicate(timeout=2)
    except (Exception, KeyboardInterrupt) as failure:
        prior_error = error
        error = LifecycleError("owned process cleanup failed: " + str(failure))
        error.__cause__ = prior_error or failure
        cleanup = {"pgid": process.pid, "verified": False, "error": str(failure)}
        try:
            os.killpg(process.pid, signal.SIGKILL)
        except OSError as kill_error:
            cleanup["kill_error"] = str(kill_error)
        try:
            stdout, stderr = process.communicate(timeout=2)
        except subprocess.TimeoutExpired as partial:
            stdout, stderr = partial.output or stdout, partial.stderr or stderr
    if error is None and process.returncode:
        error = subprocess.CalledProcessError(process.returncode, args)
    if error is not None:
        error.stdout, error.stderr = stdout, stderr
        error.returncode = process.returncode
        error.cmd = args
        error.elapsed_ms = process_elapsed_ms
        error.cleanup_elapsed_ms = (time.perf_counter_ns() - cleanup_start) / 1_000_000
        error.cleanup = cleanup
        raise error
    result = subprocess.CompletedProcess(args, process.returncode, stdout, stderr)
    result.process_elapsed_ms = process_elapsed_ms
    result.cleanup_elapsed_ms = (time.perf_counter_ns() - cleanup_start) / 1_000_000
    result.cleanup = cleanup
    return result
