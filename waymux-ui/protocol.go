package main

import (
	"errors"
	"fmt"
	"net"
	"os"
	"waymux/shared"
)

var connection shared.MessageSocket

func ConnectToSocket() error {
	_, err := os.Stat("/var/run/waymux.lock")
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("waymux is not running")
		} else {
			return fmt.Errorf("failed to stat lockfile: %w", err)
		}
	}

	conn, err := net.Dial("unix", "/var/run/waymux.sock")
	if err != nil {
		return fmt.Errorf("failed to connect to waymux socket: %w", err)
	}

	connection = shared.MessageSocket{Conn: conn}

	return nil
}
