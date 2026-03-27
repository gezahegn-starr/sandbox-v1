package cmd

import (
	"os"
	"os/exec"
	"os/signal"
	"syscall"
)

// forwardSignals relays OS signals to the child process and returns a cleanup
// function that should be deferred by the caller.
//
// It intercepts SIGHUP and SIGPIPE so they are forwarded to the container
// rather than killing the host terminal session.
//
//	defer forwardSignals(c)()
func forwardSignals(c *exec.Cmd) func() {
	ch := make(chan os.Signal, 8)
	signal.Notify(ch,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGWINCH,
		syscall.SIGHUP,
		syscall.SIGPIPE,
	)

	go func() {
		for sig := range ch {
			if c.Process != nil {
				_ = c.Process.Signal(sig)
			}
		}
	}()

	return func() {
		signal.Stop(ch)
		close(ch)
	}
}
