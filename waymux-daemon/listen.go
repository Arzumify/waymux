package main

import (
	"github.com/msteinert/pam"
	"io"
	"log/slog"
	"os"
	"os/exec"
	"strings"
	"syscall"
	"waymux/shared"

	"fmt"
	"net"
)

var hostCompositor *shared.HostCompositor

func writeBlankResponse(opcode shared.Opcode, messageSocket *shared.MessageSocket) error {
	_, err := messageSocket.WriteMessage(&shared.Message{
		Opcode: opcode,
	})
	if err != nil {
		return fmt.Errorf("failed to write message: %v", err)
	}
	return nil
}

func writeResponse(msg string, opcode shared.Opcode, messageSocket *shared.MessageSocket) error {
	reader := strings.NewReader(msg)
	_, err := messageSocket.WriteMessage(&shared.Message{
		Opcode: opcode,
		Data: &io.LimitedReader{
			R: reader,
			N: int64(reader.Len()),
		},
	})
	if err != nil {
		return fmt.Errorf("failed to write message: %v", err)
	}
	return nil
}

func accept(c net.Conn) error {
	messageSocket := &shared.MessageSocket{Conn: c}
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

		err = os.Chmod(hostCompositor.XdgRuntimeDir, 0777)
		if err != nil {
			return fmt.Errorf("failed to chmod xdg runtime directory: %w", err)
		}

		err = os.Chmod(hostCompositor.WaylandDisplay, 0777)
		if err != nil {
			return fmt.Errorf("failed to chmod wayland display: %w", err)
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
	case shared.StartSessionOpcode:
		if hostCompositor == nil {
			err = writeResponse("host not initialized", shared.ErrorOpcode, messageSocket)
			if err != nil {
				return fmt.Errorf("failed to write response: %w", err)
			}
		}

		readSession, err := shared.ReadSessionInitFrom(message.Data)
		if err != nil {
			return fmt.Errorf("failed to read session init: %w", err)
		}

		transaction, err := pam.StartFunc("passwd", readSession.Username, func(style pam.Style, _ string) (string, error) {
			return readSession.Password, nil
		})
		if err != nil {
			return fmt.Errorf("failed to start transaction: %w", err)
		}

		err = transaction.Authenticate(pam.Silent)
		if err != nil {
			err = writeResponse("authentication failed: "+err.Error(), shared.ErrorOpcode, messageSocket)
			if err != nil {
				return fmt.Errorf("failed to write message: %w", err)
			}
		}

		cmd := exec.Command(
			"su",
			readSession.Username,
			"-c",
			"\"XDG_RUNTIME_DIR=\\\"$XDG_RUNTIME_DIR\\\" $CMD\"",
		)
		cmd.Env = []string{
			"XDG_RUNTIME_DIR=" + hostCompositor.XdgRuntimeDir,
			"CMD=" + readSession.CompositorPath,
		}
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		err = cmd.Start()
		if err != nil {
			return fmt.Errorf("failed to start session: %w", err)
		}

		go func() {
			err = cmd.Wait()
			if err != nil {
				slog.Error("session terminated with error: %v", err)
			}
		}()
	default:
		err = writeResponse("invalid opcode", shared.ErrorOpcode, messageSocket)
		if err != nil {
			return fmt.Errorf("failed to write message: %w", err)
		}
		return nil
	}

	err = writeBlankResponse(shared.SuccessOpcode, messageSocket)
	if err != nil {
		return fmt.Errorf("failed to write message: %w", err)
	}

	return nil
}
