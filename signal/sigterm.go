package signal

import (
	"os"
	"os/signal"
	"syscall"
)

// OSShutdownSignal happens when OS sends a signal term signal.
func OSShutdownSignal() <-chan struct{} {
	termChan := make(chan os.Signal, 1)
	signal.Notify(termChan, syscall.SIGINT, syscall.SIGTERM)

	stop := make(chan struct{})
	go func() {
		<-termChan
		stop <- struct{}{}
	}()

	return stop
}
