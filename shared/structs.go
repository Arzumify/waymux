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
		return written, fmt.Errorf("failed to write XDG RuntimeDir: %w", err)
	}

	n, err = writeString(w, h.WaylandDisplay)
	written += n
	if err != nil {
		return written, fmt.Errorf("failed to write Wayland Display: %w", err)
	}

	n, err = writeUint64(w, uint64(h.PID))
	written += n
	if err != nil {
		return written, fmt.Errorf("failed to write PID: %w", err)
	}
	return written, nil
}

func ReadHostCompositorFrom(r io.Reader) (*HostCompositor, error) {
	var h HostCompositor
	var err error
	h.XdgRuntimeDir, err = readString(r)
	if err != nil {
		return nil, fmt.Errorf("failed to read XDG RuntimeDir: %w", err)
	}

	h.WaylandDisplay, err = readString(r)
	if err != nil {
		return nil, fmt.Errorf("failed to read Wayland Display: %w", err)
	}

	pidUint64, err := readUint64(r)
	if err != nil {
		return nil, fmt.Errorf("failed to read PID: %w", err)
	}

	h.PID = int(pidUint64)

	return &h, nil
}
