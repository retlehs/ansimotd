package sauce

import (
	"encoding/binary"
	"io"
	"os"
)

const recordSize = 128

// Record holds parsed SAUCE metadata.
type Record struct {
	Width    int // TInfo1: width in characters
	Height   int // TInfo2: height in characters
	DataSize int64
}

// Parse reads the SAUCE record from the end of a file.
// Returns nil (not an error) if no valid SAUCE record is found.
func Parse(path string) (*Record, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer func() { _ = f.Close() }()

	return ParseReader(f)
}

// ParseReader reads the SAUCE record from a ReadSeeker.
// Returns nil (not an error) if no valid SAUCE record is found.
func ParseReader(r io.ReadSeeker) (*Record, error) {
	size, err := r.Seek(0, io.SeekEnd)
	if err != nil {
		return nil, err
	}
	if size < recordSize {
		return nil, nil
	}

	if _, err := r.Seek(-recordSize, io.SeekEnd); err != nil {
		return nil, err
	}

	buf := make([]byte, recordSize)
	if _, err := io.ReadFull(r, buf); err != nil {
		return nil, err
	}

	// Validate "SAUCE00" magic at offset 0
	if string(buf[:7]) != "SAUCE00" {
		return nil, nil
	}

	// TInfo1 (width) at offset 96, TInfo2 (height) at offset 98 — little-endian uint16
	width := int(binary.LittleEndian.Uint16(buf[96:98]))
	height := int(binary.LittleEndian.Uint16(buf[98:100]))

	return &Record{
		Width:    width,
		Height:   height,
		DataSize: size - recordSize,
	}, nil
}

// ParseBytes parses a SAUCE record from raw file bytes.
// Returns nil if no valid SAUCE record is found.
func ParseBytes(data []byte) *Record {
	if len(data) < recordSize {
		return nil
	}
	sauce := data[len(data)-recordSize:]
	if string(sauce[:7]) != "SAUCE00" {
		return nil
	}
	width := int(binary.LittleEndian.Uint16(sauce[96:98]))
	height := int(binary.LittleEndian.Uint16(sauce[98:100]))
	return &Record{
		Width:    width,
		Height:   height,
		DataSize: int64(len(data) - recordSize),
	}
}
