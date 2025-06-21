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
		return written, fmt.Errorf("error writing uint64: %w", err)
	}
	return written, nil
}

func readUint64(r io.Reader) (uint64, error) {
	buf := make([]byte, U64size)
	n, err := r.Read(buf)
	if err != nil {
		return 0, fmt.Errorf("error reading uint64: %w", err)
	}
	if n != U64size {
		return 0, fmt.Errorf("error reading uint64: expected %d bytes, got %d", U64size, n)
	}

	return binary.LittleEndian.Uint64(buf), nil
}

func writeString(w io.Writer, s string) (int64, error) {
	var written int64
	n, err := writeUint64(w, uint64(len(s)))
	written += n
	if err != nil {
		return written, fmt.Errorf("error writing string size: %w", err)
	}

	wn, err := w.Write([]byte(s))
	written += int64(wn)
	if err != nil {
		return written, fmt.Errorf("error writing string: %w", err)
	}
	return written, nil
}

func readString(r io.Reader) (string, error) {
	size, err := readUint64(r)
	if err != nil {
		return "", fmt.Errorf("error reading size: %w", err)
	}
	buf := make([]byte, size)
	n, err := r.Read(buf)
	if err != nil {
		return "", fmt.Errorf("error reading string size: %w", err)
	}
	return string(buf[:n]), nil
}

type HostCompositor struct {
	XdgRuntimeDir  string
	WaylandDisplay string
}

func (h *HostCompositor) WriteTo(w io.Writer) (int64, error) {
	var written int64
	n, err := writeString(w, h.XdgRuntimeDir)
	written += n
	if err != nil {
		return written, fmt.Errorf("error writing string: %w", err)
	}
	n, err = writeString(w, h.WaylandDisplay)
	written += n
	if err != nil {
		return written, fmt.Errorf("error writing string: %w", err)
	}
	return written, nil
}

type RegisterHost struct {
	HostCompositor
	PID uint64
}

func (r *RegisterHost) WriteTo(w io.Writer) (int64, error) {
	var written int64
	n, err := r.HostCompositor.WriteTo(w)
	written += n
	if err != nil {
		return written, fmt.Errorf("error writing string: %w", err)
	}

	n, err = writeUint64(w, r.PID)
	written += n
	if err != nil {
		return written, fmt.Errorf("error writing string: %w", err)
	}
	return written, nil
}

type StopHost struct{}

func (s *StopHost) WriteTo(w io.Writer) (int64, error) {
	return 0, nil
}
