package domain

import (
	"bytes"
	"context"
	"errors"
	"io"
	"os"
	"runtime"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type streamingRunner struct {
	started chan struct{}
	release chan struct{}
}

func (*streamingRunner) Run(context.Context, string, ...string) ([]byte, error) {
	return nil, errors.New("streaming runner Run should not be called")
}

func (r *streamingRunner) RunWithOutput(ctx context.Context, output io.Writer, _ string, _ ...string) error {
	if _, err := output.Write([]byte("installing package\n")); err != nil {
		return err
	}
	close(r.started)
	select {
	case <-r.release:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

type observingWriter struct {
	mu       sync.Mutex
	output   bytes.Buffer
	ready    chan struct{}
	notified bool
}

func (w *observingWriter) Write(p []byte) (int, error) {
	w.mu.Lock()
	defer w.mu.Unlock()
	n, err := w.output.Write(p)
	if !w.notified && strings.Contains(w.output.String(), "stdout") && strings.Contains(w.output.String(), "stderr") {
		close(w.ready)
		w.notified = true
	}
	return n, err
}

func (w *observingWriter) String() string {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.output.String()
}

// fakeRunner lets tests control command output without invoking brew/apt.
type fakeRunner struct {
	// ok maps "name arg0 arg1..." to whether the command should succeed.
	ok map[string]bool
	// calls records every command invocation for assertions.
	calls [][]string
	ctxs  []context.Context
}

func (f *fakeRunner) Run(ctx context.Context, name string, args ...string) ([]byte, error) {
	call := append([]string{name}, args...)
	f.calls = append(f.calls, call)
	f.ctxs = append(f.ctxs, ctx)
	if f.ok[joinCall(call)] {
		return []byte("ok"), nil
	}
	return []byte(""), errors.New("not found")
}

func joinCall(call []string) string {
	return strings.Join(call, " ")
}

func TestBrewFormulaManager(t *testing.T) {
	runner := &fakeRunner{ok: map[string]bool{
		"brew list --formula --versions mise": true,
		"brew install missing-pkg":            true,
	}}
	m := &BrewFormulaManager{Runner: runner}

	assert.Equal(t, "brew", m.Name())

	installed, err := m.IsInstalled(context.Background(), "mise")
	require.NoError(t, err)
	assert.True(t, installed)

	installed, err = m.IsInstalled(context.Background(), "missing-pkg")
	require.NoError(t, err)
	assert.False(t, installed)

	require.NoError(t, m.Install(context.Background(), "missing-pkg"))
}

func TestBrewFormulaManager_InstallFailure(t *testing.T) {
	runner := &fakeRunner{ok: map[string]bool{}}
	m := &BrewFormulaManager{Runner: runner}

	err := m.Install(context.Background(), "broken-pkg")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "broken-pkg")
}

func TestPrepareStreamsInstallOutput(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	runner := &streamingRunner{started: make(chan struct{}), release: make(chan struct{})}
	var output bytes.Buffer
	manager := &BrewFormulaManager{Runner: runner, Output: &output}
	done := make(chan error, 1)
	completed := false
	t.Cleanup(func() {
		cancel()
		if !completed {
			<-done
		}
	})

	go func() { done <- manager.Install(ctx, "mise") }()
	select {
	case <-runner.started:
		assert.Contains(t, output.String(), "installing package", "output must arrive before Install returns")
	case <-ctx.Done():
		err := <-done
		completed = true
		require.ErrorIs(t, err, context.DeadlineExceeded)
		t.Fatal("install output did not arrive before the process timeout")
	}
	close(runner.release)
	err := <-done
	completed = true
	require.NoError(t, err)
}

func TestOsCommandRunnerStreamsOutput(t *testing.T) {
	if os.Getenv("GZH_TEST_STREAM_HELPER") == "1" {
		_, _ = os.Stdout.WriteString("stdout")
		_, _ = os.Stderr.WriteString("stderr")
		for {
			runtime.Gosched()
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	output := &observingWriter{ready: make(chan struct{})}
	done := make(chan error, 1)
	t.Setenv("GZH_TEST_STREAM_HELPER", "1")
	completed := false
	t.Cleanup(func() {
		cancel()
		if !completed {
			<-done
		}
	})

	go func() {
		done <- OsCommandRunner{}.RunWithOutput(ctx, output, os.Args[0], "-test.run=^TestOsCommandRunnerStreamsOutput$", "--")
	}()

	select {
	case <-output.ready:
		assert.Contains(t, output.String(), "stdout")
		assert.Contains(t, output.String(), "stderr")
	case <-ctx.Done():
		err := <-done
		completed = true
		require.ErrorIs(t, err, context.DeadlineExceeded)
		t.Fatal("stdout and stderr were not streamed before the process timeout")
	}
	cancel()
	err := <-done
	completed = true
	require.ErrorIs(t, err, context.Canceled)
}

func TestBrewCaskManager(t *testing.T) {
	runner := &fakeRunner{ok: map[string]bool{
		"brew list --cask --versions google-cloud-sdk": true,
	}}
	m := &BrewCaskManager{Runner: runner}

	assert.Equal(t, "cask", m.Name())

	installed, err := m.IsInstalled(context.Background(), "google-cloud-sdk")
	require.NoError(t, err)
	assert.True(t, installed)
}

func TestAptManager(t *testing.T) {
	runner := &fakeRunner{ok: map[string]bool{
		"dpkg-query -W -f=${Status} chezmoi": true,
	}}
	m := &AptManager{Runner: runner}

	assert.Equal(t, "apt", m.Name())

	installed, err := m.IsInstalled(context.Background(), "chezmoi")
	require.NoError(t, err)
	// fakeRunner returns "ok" as output, which does not contain
	// "install ok installed", so IsInstalled must report false here —
	// this exercises the output-parsing branch, not just the error branch.
	assert.False(t, installed)
}

func TestAptManager_IsInstalled_ParsesDpkgStatus(t *testing.T) {
	runner := &dpkgStatusRunner{status: "install ok installed"}
	m := &AptManager{Runner: runner}

	installed, err := m.IsInstalled(context.Background(), "chezmoi")
	require.NoError(t, err)
	assert.True(t, installed)
}

func TestAptManager_IsInstalled_UnknownPackageIsMissingNotError(t *testing.T) {
	runner := &commandResultRunner{err: commandExitError{code: 1}}
	m := &AptManager{Runner: runner}

	installed, err := m.IsInstalled(context.Background(), "does-not-exist")
	require.NoError(t, err)
	assert.False(t, installed)
}

func TestAptManager_IsInstalled_FatalDpkgErrorIsReported(t *testing.T) {
	cause := commandExitError{code: 2}
	runner := &commandResultRunner{output: []byte("database is locked"), err: cause}
	m := &AptManager{Runner: runner}

	installed, err := m.IsInstalled(context.Background(), "chezmoi")

	assert.False(t, installed)
	require.Error(t, err)
	assert.ErrorIs(t, err, cause)
	assert.Contains(t, err.Error(), "chezmoi")
	assert.Contains(t, err.Error(), "database is locked")
}

func TestAptManager_IsInstalled_CommandFailureIsReported(t *testing.T) {
	cause := errors.New("dpkg-query executable not found")
	runner := &commandResultRunner{err: cause}
	m := &AptManager{Runner: runner}

	installed, err := m.IsInstalled(context.Background(), "chezmoi")

	assert.False(t, installed)
	require.Error(t, err)
	assert.ErrorIs(t, err, cause)
	assert.Contains(t, err.Error(), "chezmoi")
}

// dpkgStatusRunner returns a fixed dpkg-query style status line regardless of args.
type dpkgStatusRunner struct{ status string }

func (r *dpkgStatusRunner) Run(context.Context, string, ...string) ([]byte, error) {
	return []byte(r.status), nil
}

type commandResultRunner struct {
	output []byte
	err    error
}

func (r *commandResultRunner) Run(context.Context, string, ...string) ([]byte, error) {
	return r.output, r.err
}

type commandExitError struct{ code int }

func (e commandExitError) Error() string { return "command failed" }
func (e commandExitError) ExitCode() int { return e.code }

func TestBrewFormulaManager_ForwardsContextToRunner(t *testing.T) {
	runner := &fakeRunner{ok: map[string]bool{"brew list --formula --versions mise": true}}
	manager := &BrewFormulaManager{Runner: runner}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	installed, err := manager.IsInstalled(ctx, "mise")

	require.NoError(t, err)
	assert.True(t, installed)
	require.Len(t, runner.ctxs, 1)
	assert.Same(t, ctx, runner.ctxs[0])
}

func TestBrewManagers_PropagateContextCancellation(t *testing.T) {
	tests := []struct {
		name    string
		manager PackageManager
		wantErr error
	}{
		{name: "formula", manager: &BrewFormulaManager{Runner: &commandResultRunner{err: context.Canceled}}, wantErr: context.Canceled},
		{name: "cask", manager: &BrewCaskManager{Runner: &commandResultRunner{err: context.DeadlineExceeded}}, wantErr: context.DeadlineExceeded},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			installed, err := tt.manager.IsInstalled(context.Background(), "pkg")

			assert.False(t, installed)
			assert.ErrorIs(t, err, tt.wantErr)
		})
	}
}

func TestPackageManagers_PreferCanceledContextToCommandExitError(t *testing.T) {
	tests := []struct {
		name    string
		manager PackageManager
	}{
		{name: "formula", manager: &BrewFormulaManager{Runner: &commandResultRunner{err: commandExitError{code: 1}}}},
		{name: "cask", manager: &BrewCaskManager{Runner: &commandResultRunner{err: commandExitError{code: 1}}}},
		{name: "apt", manager: &AptManager{Runner: &commandResultRunner{err: commandExitError{code: 1}}}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			cancel()

			installed, err := tt.manager.IsInstalled(ctx, "pkg")

			assert.False(t, installed)
			assert.ErrorIs(t, err, context.Canceled)
		})
	}
}
