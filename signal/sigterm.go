package signal

import (
	"os"
	"os/signal"
	"syscall"
)

// SigtermSignal happens when OS sends a signal term or sigint signal.
func SigtermSignal() chan struct{} {
	termChan := make(chan os.Signal, 1)
	signal.Notify(termChan, syscall.SIGINT, syscall.SIGTERM)

	stop := make(chan struct{})
	go func() {
		<-termChan
		stop <- struct{}{}
	}()

	return stop
}
