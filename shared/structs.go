package shared

import (
	"encoding/binary"
	"fmt"
	"io"
)

func writeUint64(w io.Writer, v uint64) (int64, error) {
	buf := make([]byte, U64size)
	binary.LittleEndian.PutUint64(buf, v)

	var written int64
	n, err := w.Write(buf[:])
	written += int64(n)
	if err != nil {
		return written, fmt.Errorf("failed to write uint64 content: %w", err)
	}
	return written, nil
}

func readUint64(r io.Reader) (uint64, error) {
	buf := make([]byte, U64size)
	n, err := r.Read(buf)
	if err != nil {
		return 0, fmt.Errorf("failed to read uint64 content: %w", err)
	}
	if n != U64size {
		return 0, fmt.Errorf("failed to read uint64 content: expected %d bytes, got %d", U64size, n)
	}

	return binary.LittleEndian.Uint64(buf), nil
}

func writeString(w io.Writer, s string) (int64, error) {
	var written int64
	n, err := writeUint64(w, uint64(len(s)))
	written += n
	if err != nil {
		return written, fmt.Errorf("failed to write string size: %w", err)
	}

	wn, err := w.Write([]byte(s))
	written += int64(wn)
	if err != nil {
		return written, fmt.Errorf("failed to write string content: %w", err)
	}
	return written, nil
}

func readString(r io.Reader) (string, error) {
	size, err := readUint64(r)
	if err != nil {
		return "", fmt.Errorf("failed to read size size: %w", err)
	}
	buf := make([]byte, size)
	n, err := r.Read(buf)
	if err != nil {
		return "", fmt.Errorf("failed to read string content: %w", err)
	}
	return string(buf[:n]), nil
}

type HostCompositor struct {
	XdgRuntimeDir  string
	WaylandDisplay string
	PID            int
}

func (h *HostCompositor) WriteTo(w io.Writer) (int64, error) {
	var written int64
	n, err := writeString(w, h.XdgRuntimeDir)
	written += n
	if err != nil {
		return written, fmt.Errorf("failed to write xdg runtime directory: %w", err)
	}

	n, err = writeString(w, h.WaylandDisplay)
	written += n
	if err != nil {
		return written, fmt.Errorf("failed to write wayland display: %w", err)
	}

	n, err = writeUint64(w, uint64(h.PID))
	written += n
	if err != nil {
		return written, fmt.Errorf("failed to write pid: %w", err)
	}
	return written, nil
}

func ReadHostCompositorFrom(r io.Reader) (*HostCompositor, error) {
	var h HostCompositor
	var err error
	h.XdgRuntimeDir, err = readString(r)
	if err != nil {
		return nil, fmt.Errorf("failed to read xdg runtime directory: %w", err)
	}

	h.WaylandDisplay, err = readString(r)
	if err != nil {
		return nil, fmt.Errorf("failed to read wayland display: %w", err)
	}

	pidUint64, err := readUint64(r)
	if err != nil {
		return nil, fmt.Errorf("failed to read pid: %w", err)
	}

	h.PID = int(pidUint64)

	return &h, nil
}

type SessionInit struct {
	Username       string
	Password       string
	CompositorPath string
}

func (s *SessionInit) WriteTo(w io.Writer) (int64, error) {
	var written int64
	n, err := writeString(w, s.Username)
	written += n
	if err != nil {
		return written, fmt.Errorf("failed to write username: %w", err)
	}

	n, err = writeString(w, s.Password)
	written += n
	if err != nil {
		return written, fmt.Errorf("failed to write password: %w", err)
	}

	n, err = writeString(w, s.CompositorPath)
	written += n
	if err != nil {
		return written, fmt.Errorf("failed to write compositor path: %w", err)
	}

	return written, nil
}

func ReadSessionInitFrom(r io.Reader) (*SessionInit, error) {
	var h SessionInit
	var err error
	h.Username, err = readString(r)
	if err != nil {
		return nil, fmt.Errorf("failed to read username: %w", err)
	}

	h.Password, err = readString(r)
	if err != nil {
		return nil, fmt.Errorf("failed to read password: %w", err)
	}

	h.CompositorPath, err = readString(r)
	if err != nil {
		return nil, fmt.Errorf("failed to read compositor path: %w", err)
	}

	return &h, nil
}
