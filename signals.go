package micro

import (
	"os"
	"syscall"
)

// InterruptSignals are the default interrupt signals to catch
var InterruptSignals = []os.Signal{
	syscall.SIGSTOP,
	syscall.SIGINT,
	syscall.SIGTERM,
	syscall.SIGQUIT,
}
