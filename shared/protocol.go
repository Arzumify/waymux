package shared

import (
	"encoding/binary"
	"fmt"
	"io"
	"net"
)

const u64size = 8
const u8size = 1

type Opcode uint8

// RegisterHost and StartHost being removed are blocked on headless

const (
	/*
		Client -> Server opcodes
	*/

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

	/*
		Server reply opcodes
	*/

	// SuccessOpcode indicates that the request was successful
	SuccessOpcode
	// ErrorOpcode indicates that the request has failed
	ErrorOpcode
)

type Message struct {
	Opcode Opcode
	Data   *io.LimitedReader
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
	size := binary.BigEndian.Uint64(sizeBuffer)

	var msg Message
	msg.Opcode = Opcode(opcodeBuffer[0])
	if size > 0 {
		msg.Data = &io.LimitedReader{
			R: m.Conn,
			N: int64(size),
		}
	}

	return &msg, nil
}

func (m *MessageSocket) WriteMessage(msg *Message) (int64, error) {
	var written int64
	n, err := m.Write([]byte{byte(msg.Opcode)})
	written += int64(n)
	if err != nil {
		return written, fmt.Errorf("failed to write opcode: %w", err)
	}

	sizeBuffer := make([]byte, u64size)
	binary.BigEndian.PutUint64(sizeBuffer, uint64(msg.Data.N))
	n, err = m.Write(sizeBuffer)
	written += int64(n)
	if err != nil {
		return written, fmt.Errorf("failed to write size: %w", err)
	}

	if msg.Data.N > 0 {
		nc, err := io.CopyN(m, msg.Data, msg.Data.N)
		if err != nil {
			return written, fmt.Errorf("failed to write size: %w", err)
		}
		written += nc
	}

	return written, nil
}
