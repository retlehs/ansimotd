package render

import (
	"bytes"
	"golang.org/x/text/encoding/charmap"
	"io"
	"strconv"
	"strings"

	"github.com/retlehs/ansimotd/internal/sauce"
)

// Render processes raw ANSI art bytes through the full pipeline and writes to w.
// If rec is non-nil, SAUCE bytes are stripped and line wrapping uses rec.Width.
func Render(w io.Writer, raw []byte, rec *sauce.Record) error {
	data := raw
	width := 0

	// 1. Strip SAUCE record
	if rec != nil && rec.DataSize > 0 && rec.DataSize < int64(len(data)) {
		data = data[:rec.DataSize]
		width = rec.Width
	}

	// 2. Decode CP437 → UTF-8
	decoded, err := charmap.CodePage437.NewDecoder().Bytes(data)
	if err != nil {
		return err
	}

	// 3. Strip \r
	decoded = bytes.ReplaceAll(decoded, []byte("\r"), nil)

	// 3b. Strip leading and trailing blank lines
	decoded = trimLeadingBlanks(decoded)
	decoded = trimTrailingBlanks(decoded)

	// 4+5. Remap SGR colors and wrap lines
	output := processANSI(string(decoded), width)

	// 6. Write with trailing newline
	if _, err := io.WriteString(w, output); err != nil {
		return err
	}
	_, err = io.WriteString(w, "\n")
	return err
}

// processANSI operates on a UTF-8 string using rune-level iteration for
// correct handling of multi-byte characters (CP437 box-drawing, etc.)
// while parsing ANSI escape sequences at the byte level (they're always ASCII).
func processANSI(data string, wrapWidth int) string {
	state := &sgrState{}
	lines := strings.Split(data, "\n")
	var buf strings.Builder

	for lineIdx, line := range lines {
		col := 0
		runes := []rune(line)
		i := 0
		for i < len(runes) {
			if runes[i] == '\x1b' && i+1 < len(runes) && runes[i+1] == '[' {
				// ANSI escape sequences are pure ASCII, safe to parse as runes
				j := i + 2
				for j < len(runes) && !isAlphaRune(runes[j]) {
					j++
				}
				if j >= len(runes) {
					buf.WriteRune(runes[i])
					i++
					continue
				}

				seq := string(runes[i : j+1])
				cmd := runes[j]

				if cmd == 'm' {
					buf.WriteString(state.remapSGR(seq))
				} else if isCursorMoveRune(cmd) {
					col = applyCursorMove(seq, byte(cmd), col)
					buf.WriteString(seq)
				} else {
					buf.WriteString(seq)
				}
				i = j + 1
			} else {
				switch runes[i] {
				case '\t':
					nextTab := ((col / 8) + 1) * 8
					buf.WriteRune('\t')
					col = nextTab
				case '\b':
					buf.WriteRune('\b')
					if col > 0 {
						col--
					}
				default:
					if wrapWidth > 0 && col >= wrapWidth {
						buf.WriteString(state.resetBG())
						buf.WriteByte('\n')
						col = 0
					}
					buf.WriteRune(runes[i])
					// Count display width: each rune is 1 column
					// (CP437 chars are all single-width)
					col++
				}
				i++
			}
		}
		if lineIdx < len(lines)-1 {
			// Reset background (only) before the newline so bg doesn't bleed
			// past the art's right edge on terminals wider than wrapWidth.
			// FG and bold persist naturally across the newline.
			buf.WriteString(state.resetBG())
			buf.WriteByte('\n')
		}
	}

	return buf.String()
}

func isAlphaRune(r rune) bool {
	return (r >= 'A' && r <= 'Z') || (r >= 'a' && r <= 'z')
}

func isCursorMoveRune(r rune) bool {
	return r == 'A' || r == 'B' || r == 'C' || r == 'D' ||
		r == 'E' || r == 'F' || r == 'G' || r == 'H' || r == 'f'
}

// trimLeadingBlanks removes leading lines that contain no visible content.
func trimLeadingBlanks(data []byte) []byte {
	lines := bytes.Split(data, []byte("\n"))
	start := 0
	for start < len(lines) && isBlankLine(lines[start]) {
		start++
	}
	if start == 0 {
		return data
	}
	return bytes.Join(lines[start:], []byte("\n"))
}

// trimTrailingBlanks removes trailing lines that contain no visible content
// (only whitespace and/or ANSI escape sequences). This handles padding
// that some files have between the art and the SAUCE record.
func trimTrailingBlanks(data []byte) []byte {
	lines := bytes.Split(data, []byte("\n"))
	end := len(lines)
	for end > 0 && isBlankLine(lines[end-1]) {
		end--
	}
	if end == len(lines) {
		return data
	}
	return bytes.Join(lines[:end], []byte("\n"))
}

// isBlankLine returns true if the line has no visible characters
// (only whitespace and ANSI escape sequences).
func isBlankLine(line []byte) bool {
	i := 0
	for i < len(line) {
		if line[i] == '\x1b' && i+1 < len(line) && line[i+1] == '[' {
			// Skip ANSI escape sequence
			j := i + 2
			for j < len(line) && !isAlpha(line[j]) {
				j++
			}
			if j < len(line) {
				i = j + 1
			} else {
				i++
			}
		} else if line[i] == ' ' || line[i] == '\t' || line[i] == '\x1a' {
			// SUB (0x1A) is the EOF marker often before SAUCE
			i++
		} else {
			return false
		}
	}
	return true
}

func isAlpha(b byte) bool {
	return (b >= 'A' && b <= 'Z') || (b >= 'a' && b <= 'z')
}

func applyCursorMove(seq string, cmd byte, col int) int {
	inner := seq[2 : len(seq)-1]
	parts := strings.Split(inner, ";")

	num1 := 1
	num2 := 1
	if len(parts) >= 1 && parts[0] != "" {
		if n, err := strconv.Atoi(parts[0]); err == nil {
			num1 = n
		}
	}
	if len(parts) >= 2 && parts[1] != "" {
		if n, err := strconv.Atoi(parts[1]); err == nil {
			num2 = n
		}
	}

	switch cmd {
	case 'C': // forward
		col += num1
	case 'D': // backward
		col -= num1
		if col < 0 {
			col = 0
		}
	case 'G': // horizontal absolute
		col = num1 - 1
	case 'H', 'f': // cursor position (row;col)
		if len(parts) >= 2 {
			col = num2 - 1
		}
	}
	return col
}
