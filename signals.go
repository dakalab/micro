package micro

import (
	"os"
	"syscall"
)

// InterruptSignals are the interrupt signals to catch
var InterruptSignals = []os.Signal{
	syscall.SIGSTOP,
	syscall.SIGINT,
	syscall.SIGTERM,
	syscall.SIGQUIT,
}
