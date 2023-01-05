package actionlint

import (
	"context"
	"fmt"
	"io"
	"os/exec"
	"sync"

	"golang.org/x/sync/errgroup"
	"golang.org/x/sync/semaphore"
	"golang.org/x/sys/execabs"
)

// concurrentProcess is a manager to run process concurrently. Since running process consumes OS
// resources, running too many processes concurrently causes some issues. On macOS, making new
// process hangs (see issue #3). And also running processes which opens files causes an error
// "pipe: too many files to open". To avoid it, this class manages how many processes are run at
// the same time.
type concurrentProcess struct {
	ctx  context.Context
	sema *semaphore.Weighted
	wg   sync.WaitGroup
}

func newConcurrentProcess(par int) *concurrentProcess {
	return &concurrentProcess{
		ctx:  context.Background(),
		sema: semaphore.NewWeighted(int64(par)),
	}
}

func runProcessWithStdin(exe string, args []string, stdin string) ([]byte, error) {
	cmd := exec.Command(exe, args...)
	cmd.Stderr = nil

	p, err := cmd.StdinPipe()
	if err != nil {
		return nil, fmt.Errorf("could not make stdin pipe for %s process: %w", exe, err)
	}
	if _, err := io.WriteString(p, stdin); err != nil {
		p.Close()
		return nil, fmt.Errorf("could not write to stdin of %s process: %w", exe, err)
	}
	p.Close()

	stdout, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			code := exitErr.ExitCode()
			if code < 0 {
				return nil, fmt.Errorf("%s was terminated. stderr: %q", exe, exitErr.Stderr)
			}
			if len(stdout) == 0 {
				return nil, fmt.Errorf("%s exited with status %d but stdout was empty. stderr: %q", exe, code, exitErr.Stderr)
			}
			// Reaches here when exit status is non-zero and stdout is not empty, shellcheck successfully found some errors
		} else {
			return nil, err
		}
	}

	return stdout, nil
}

func (proc *concurrentProcess) run(eg *errgroup.Group, exe string, args []string, stdin string, callback func([]byte, error) error) {
	proc.sema.Acquire(proc.ctx, 1)
	proc.wg.Add(1)
	eg.Go(func() error {
		defer proc.wg.Done()
		stdout, err := runProcessWithStdin(exe, args, stdin)
		proc.sema.Release(1)
		return callback(stdout, err)
	})
}

// wait waits all goroutines started by this concurrentProcess instance finish.
func (proc *concurrentProcess) wait() {
	proc.wg.Wait() // Wait for all goroutines completing to shutdown
}

// newCommandRunner creates new external command runner for given executable. The executable path
// is resolved in this function.
func (proc *concurrentProcess) newCommandRunner(exe string) (*externalCommand, error) {
	p, err := execabs.LookPath(exe)
	if err != nil {
		return nil, err
	}
	cmd := &externalCommand{
		proc: proc,
		exe:  p,
	}
	return cmd, nil
}

// externalCommand is struct to run specific command concurrently with concurrentProcess bounding
// number of processes at the same time. This type manages fatal errors while running the command
// by using errgroup.Group. The wait() method must be called at the end for checking if some fatal
// error occurred.
type externalCommand struct {
	proc *concurrentProcess
	eg   errgroup.Group
	exe  string
}

// run runs the command with given arguments and stdin. The callback function is called after the
// process runs. First argument is stdout and the second argument is an error while running the
// process.
func (cmd *externalCommand) run(args []string, stdin string, callback func([]byte, error) error) {
	cmd.proc.run(&cmd.eg, cmd.exe, args, stdin, callback)
}

// wait waits until all goroutines for this command finish. Note that it does not wait for
// goroutines for other commands.
func (cmd *externalCommand) wait() error {
	return cmd.eg.Wait()
}
