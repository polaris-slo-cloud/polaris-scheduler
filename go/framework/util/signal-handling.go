package util

import (
	"context"
	"os"
	"os/signal"
	"syscall"
)

// Creates a Context that gets cancelled upon an interrupt or TERM signal.
// If the signal is received a second time, the program will be exited immediately.
func SetupSignalHandlingContext() context.Context {
	ctx, cancel := context.WithCancel(context.Background())

	signalHandler := make(chan os.Signal, 2)
	signal.Notify(signalHandler, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-signalHandler
		cancel()
		<-signalHandler
		os.Exit(1) // Exit immediately if we receive a second signal.
	}()

	return ctx
}
