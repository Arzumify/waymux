package main

import (
	"syscall"
	"waymux/shared"

	"fmt"
	"net"
)

var hostCompositor *shared.HostCompositor

func accept(c net.Conn) error {
	messageSocket := shared.MessageSocket{Conn: c}
	message, err := messageSocket.Next()
	if err != nil {
		return fmt.Errorf("failed to read message: %w", err)
	}
	switch message.Opcode {
	case shared.RegisterHostOpcode:
		hostCompositor, err = shared.ReadHostCompositorFrom(message.Data)
		if err != nil {
			return fmt.Errorf("failed to read register host: %w", err)
		}

		_, err := messageSocket.WriteMessage(&shared.Message{
			Opcode: shared.SuccessOpcode,
		})
		if err != nil {
			return fmt.Errorf("failed to write message: %w", err)
		}
	case shared.StopHostOpcode:
		if hostCompositor == nil {
			_, err = messageSocket.WriteMessage(&shared.Message{
				Opcode: shared.ErrorOpcode,
			})
		}
		err := syscall.Kill(hostCompositor.PID, syscall.SIGINT)
		if err != nil {
			return err
		}
	default:
		_, err := messageSocket.WriteMessage(&shared.Message{
			Opcode: shared.ErrorOpcode,
		})
		if err != nil {
			return fmt.Errorf("failed to write message: %w", err)
		}
	}
	return nil
}
