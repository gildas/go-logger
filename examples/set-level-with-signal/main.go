// +build linux darwin
package main

import (
	"context"
	"os"
	"os/signal"
	"time"

	"github.com/gildas/go-logger"
	"golang.org/x/sys/unix"
)

func generateSomeLogs(logger *logger.Logger) (stop chan struct{}) {
	stop = make(chan struct{}, 1)

	go func() {
		logger.Infof("Logging stuff every 5 seconds...")
		for {
			select {
			case <-stop:
				logger.Infof("Stopping...")
				return
			case <-time.After(5*time.Second):
				logger.Fatalf("Something went very wrong")
				logger.Errorf("Something went wrong")
				logger.Warnf("You should pay attention")
				logger.Infof("Logging stuff")
				logger.Debugf("Debugging stuff")
				logger.Tracef("Tracing stuff")
			}
		}
	}()

	return stop
}

func handleSignals(logger *logger.Logger) (stopChannel chan struct{}) {
	stopChannel = make(chan struct{}, 1)
	signalChannel := make(chan os.Signal, 1)
	signal.Notify(signalChannel, unix.SIGUSR1, unix.SIGUSR2)

	go func() {
		for {
			select {
			case sig := <-signalChannel:
				switch sig {
				case unix.SIGUSR1:
					logger.Warnf("*********** Received SIGUSR1 -> Logging less stuff")
					logger.FilterMore()
					logger.Warnf("%s", logger)
				case unix.SIGUSR2:
					logger.Warnf("*********** Received SIGUSR2 -> Logging more stuff")
					logger.FilterLess()
					logger.Warnf("Logger: %s", logger)
				default:
					logger.Errorf("Received unsupported %s", sig)
				}
			case <-stopChannel:
				logger.Warnf("Signals are not processed anymore")
				return
			}
		}
	}()
	return stopChannel
}

func main() {
	logger := logger.Create("signal", &logger.StdoutStream{Unbuffered: true})

	logger.Infof("%s", logger)
	stopHandlingChannel := handleSignals(logger)
	stopGeneratingChannel := generateSomeLogs(logger)

	interruptChannel := make(chan os.Signal, 1)
	exitChannel := make(chan struct{})
	signal.Notify(interruptChannel, os.Interrupt, unix.SIGTERM)

	go func() {
		sig := <-interruptChannel // Block until we have to stop
		_, cancel := context.WithTimeout(context.Background(), time.Second*5)
		defer cancel()

		logger.Infof("Stopping after receiving %s", sig)
		close(stopGeneratingChannel)
		close(stopHandlingChannel)

		close(exitChannel)
	}()

	<-exitChannel
}
