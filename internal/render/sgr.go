package render

import (
	"fmt"
	"strconv"
	"strings"
)

// sgrState tracks the color/bold state across SGR sequences.
type sgrState struct {
	boldActive bool
	currentFG  int // 0 = unset
	currentBG  int // 0 = unset
}

// remapSGR rewrites an SGR escape sequence to use 24-bit true-color VGA values.
// seq is the full sequence including ESC[ and m (e.g. "\x1b[1;31m").
func (s *sgrState) remapSGR(seq string) string {
	// Extract parameter string between ESC[ and m
	inner := seq[2 : len(seq)-1]
	var params []int
	if inner == "" {
		params = []int{0}
	} else {
		for _, p := range strings.Split(inner, ";") {
			n, _ := strconv.Atoi(p) // empty/invalid → 0
			params = append(params, n)
		}
	}

	var out []string
	i := 0
	for i < len(params) {
		p := params[i]
		switch {
		case p == 0: // reset
			s.boldActive = false
			s.currentFG = 0
			s.currentBG = 0
			out = append(out, "0")

		case p == 1: // bold
			s.boldActive = true
			// Upgrade dark FG to bright if already set
			if s.currentFG >= 30 && s.currentFG <= 37 {
				bright := s.currentFG + 60
				if rgb, ok := vgaFG[bright]; ok {
					out = append(out, fgRGB(rgb))
					s.currentFG = bright
				}
			}

		case p == 22: // normal intensity
			s.boldActive = false
			// Downgrade bright FG to dark
			if s.currentFG >= 90 && s.currentFG <= 97 {
				dark := s.currentFG - 60
				if rgb, ok := vgaFG[dark]; ok {
					out = append(out, fgRGB(rgb))
					s.currentFG = dark
				}
			}

		case isFG(p): // foreground color
			actual := p
			if s.boldActive && p >= 30 && p <= 37 {
				actual = p + 60
			}
			if rgb, ok := vgaFG[actual]; ok {
				out = append(out, fgRGB(rgb))
				s.currentFG = actual
			} else {
				out = append(out, strconv.Itoa(p))
			}

		case isBG(p): // background color
			if rgb, ok := vgaBG[p]; ok {
				out = append(out, bgRGB(rgb))
				s.currentBG = p
			} else {
				out = append(out, strconv.Itoa(p))
			}

		case p == 38 && i+1 < len(params) && params[i+1] == 5: // 256-color FG passthrough
			end := min(i+3, len(params))
			parts := make([]string, end-i)
			for j := i; j < end; j++ {
				parts[j-i] = strconv.Itoa(params[j])
			}
			out = append(out, strings.Join(parts, ";"))
			i = end - 1

		case p == 48 && i+1 < len(params) && params[i+1] == 5: // 256-color BG passthrough
			end := min(i+3, len(params))
			parts := make([]string, end-i)
			for j := i; j < end; j++ {
				parts[j-i] = strconv.Itoa(params[j])
			}
			out = append(out, strings.Join(parts, ";"))
			i = end - 1

		default:
			out = append(out, strconv.Itoa(p))
		}
		i++
	}

	return "\x1b[" + strings.Join(out, ";") + "m"
}

func isFG(p int) bool { return (p >= 30 && p <= 37) || (p >= 90 && p <= 97) }
func isBG(p int) bool { return (p >= 40 && p <= 47) || (p >= 100 && p <= 107) }

func fgRGB(c RGB) string { return fmt.Sprintf("38;2;%d;%d;%d", c.R, c.G, c.B) }
func bgRGB(c RGB) string { return fmt.Sprintf("48;2;%d;%d;%d", c.R, c.G, c.B) }
