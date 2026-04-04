package render

// RGB holds a true-color value.
type RGB struct{ R, G, B uint8 }

// VGA palette — the exact RGB values used by DOS/VGA hardware.
var vgaFG = map[int]RGB{
	30: {0, 0, 0},       // black
	31: {170, 0, 0},     // red
	32: {0, 170, 0},     // green
	33: {170, 85, 0},    // brown
	34: {0, 0, 170},     // blue
	35: {170, 0, 170},   // magenta
	36: {0, 170, 170},   // cyan
	37: {170, 170, 170}, // light gray
	90: {85, 85, 85},    // dark gray
	91: {255, 85, 85},   // light red
	92: {85, 255, 85},   // light green
	93: {255, 255, 85},  // yellow
	94: {85, 85, 255},   // light blue
	95: {255, 85, 255},  // light magenta
	96: {85, 255, 255},  // light cyan
	97: {255, 255, 255}, // white
}

var vgaBG = map[int]RGB{
	40:  {0, 0, 0},
	41:  {170, 0, 0},
	42:  {0, 170, 0},
	43:  {170, 85, 0},
	44:  {0, 0, 170},
	45:  {170, 0, 170},
	46:  {0, 170, 170},
	47:  {170, 170, 170},
	100: {85, 85, 85},
	101: {255, 85, 85},
	102: {85, 255, 85},
	103: {255, 255, 85},
	104: {85, 85, 255},
	105: {255, 85, 255},
	106: {85, 255, 255},
	107: {255, 255, 255},
}
