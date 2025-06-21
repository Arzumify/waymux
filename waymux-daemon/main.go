package main

import (
	"errors"
	"fmt"
	"log/slog"
	"net"
	"os"
	"os/signal"
	"strconv"
	"syscall"
)

const (
	lockFile = "/var/run/waymux.lock"
	socket   = "/var/run/waymux.sock"
)

var errWaymuxAlreadyRunning = errors.New("waymux is already running")

func createLockFile() error {
	pid := os.Getpid()
	_, err := os.Stat(lockFile)
	if err == nil {
		return errWaymuxAlreadyRunning
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("failed to check if lock file exists: %w", err)
	}
	handler, err := os.OpenFile(lockFile, os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		return fmt.Errorf("failed to open lock file: %w", err)
	}
	_, err = handler.Write([]byte(strconv.Itoa(pid)))
	if err != nil {
		return fmt.Errorf("failed to write to lock file: %w", err)
	}
	err = handler.Close()
	if err != nil {
		return fmt.Errorf("failed to close lock file: %w", err)
	}
	return nil
}

func listenerLoop(listener net.Listener) {
	for {
		conn, err := listener.Accept()
		if err != nil {
			slog.Error("Failed to accept connection" + err.Error())
			continue
		}

		err = accept(conn)
		if err != nil {
			slog.Error("Failed to process connection" + err.Error())
		}
	}
}

func startListener() error {
	_, err := os.Stat(socket)
	if err == nil {
		err = os.Remove(socket)
		if err != nil {
			return fmt.Errorf("failed to remove socket: %w", err)
		}
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("failed to stat socket: %w", err)
	}

	listener, err := net.Listen("unix", socket)
	if err != nil {
		return fmt.Errorf("failed to create listener: %w", err)
	}

	err = syscall.Chmod(socket, 0777)
	if err != nil {
		return fmt.Errorf("failed to chmod listener: %w", err)
	}

	go listenerLoop(listener)
	return nil
}

func main() {
	slog.Info("Starting waymux-daemon...")
	slog.Info("Creating lockfile...")
	err := createLockFile()
	if err != nil {
		slog.Error("Failed to create lockfile:", err)
		os.Exit(1)
	}
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		err = os.Remove(lockFile)
		if err != nil {
			slog.Error("Failed to remove lock file:", err)
		}
		err = os.Remove(socket)
		if err != nil {
			slog.Error("Failed to remove socket:", err)
		}
		os.Exit(1)
	}()
	slog.Info("Starting listener...")
	err = startListener()
	if err != nil {
		slog.Error("Failed to start listener:", err)
	}
	slog.Info("Listening at /var/run/waymux.sock")
	select {}
}
