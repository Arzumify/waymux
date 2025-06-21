package shared

import (
	"encoding/binary"
	"fmt"
	"net"
)

const u64size = 8
const u8size = 1

type Opcode uint8

// RegisterHost and StartHost being removed are blocked on headless

const (
	// RegisterHostOpcode registers the caller's wayland session as the host for waymux sessions
	RegisterHostOpcode Opcode = iota
	// StopHostOpcode stops the host and exits to TTY
	StopHostOpcode
	// StartSessionOpcode starts a waymux session
	StartSessionOpcode
	// StopSessionOpcode kills a waymux session
	StopSessionOpcode
	// StopAllSessionsOpcode kills all waymux sessions
	StopAllSessionsOpcode
	// ListSessionsOpcode lists all waymux sessions
	ListSessionsOpcode
	// WhoAmIOpcode tells the caller the current wayland session, if any
	WhoAmIOpcode
	/*
		HideSessionOpcode // Blocked on session switcher
		ShowSessionOpcode // Blocked on session switcher
	*/
)

type Message struct {
	Opcode Opcode
	Data   []byte
}

type MessageSocket struct {
	net.Conn
}

func (m *MessageSocket) Next() (*Message, error) {
	opcodeBuffer := make([]byte, u8size)
	_, err := m.Read(opcodeBuffer)
	if err != nil {
		return nil, fmt.Errorf("failed to read opcode: %w", err)
	}

	sizeBuffer := make([]byte, u64size)
	_, err = m.Read(sizeBuffer)
	if err != nil {
		return nil, fmt.Errorf("failed to read size: %w", err)
	}
	size := int64(binary.BigEndian.Uint64(sizeBuffer))

	var msg Message
	msg.Opcode = Opcode(opcodeBuffer[0])
	msg.Data = make([]byte, size)
	_, err = m.Read(msg.Data)
	if err != nil {
		return nil, fmt.Errorf("failed to read data: %w", err)
	}

	return &msg, nil
}

func (m *MessageSocket) WriteMessage(msg *Message) (int, error) {
	var written int
	n, err := m.Write([]byte{byte(msg.Opcode)})
	written += n
	if err != nil {
		return written, fmt.Errorf("failed to write opcode: %w", err)
	}

	sizeBuffer := make([]byte, u64size)
	binary.BigEndian.PutUint64(sizeBuffer, uint64(len(msg.Data)))
	n, err = m.Write(sizeBuffer)
	written += n
	if err != nil {
		return written, fmt.Errorf("failed to write size: %w", err)
	}

	n, err = m.Write(msg.Data)
	written += n
	if err != nil {
		return written, fmt.Errorf("failed to write json: %w", err)
	}

	return written, nil
}
