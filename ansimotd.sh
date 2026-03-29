#!/bin/sh

# Get terminal dimensions
get_terminal_dimensions() {
  # Try stty with /dev/tty first
  if command -v stty >/dev/null 2>&1; then
    dimensions=$(stty size </dev/tty 2>/dev/null)
    if [ -n "$dimensions" ]; then
      TERM_LINES=$(echo "$dimensions" | cut -d' ' -f1)
      TERM_COLS=$(echo "$dimensions" | cut -d' ' -f2)
      return
    fi
  fi

  # Fallback to tput
  if command -v tput >/dev/null 2>&1; then
    TERM_COLS=$(tput cols 2>/dev/null || echo "${COLUMNS:-80}")
    TERM_LINES=$(tput lines 2>/dev/null || echo "${LINES:-24}")
  else
    TERM_COLS="${COLUMNS:-80}"
    TERM_LINES="${LINES:-24}"
  fi
}

# Parse SAUCE metadata from ANSI file
# SAUCE is a 128-byte metadata block at the end of ANSI art files
# Format: https://www.acid.org/info/sauce/sauce.htm
parse_sauce_dimensions() {
  ansi_file="$1"

  if [ ! -f "$ansi_file" ]; then
    echo "0 0"
    return
  fi

  # Check if it has a SAUCE record (starts with "SAUCE00")
  sauce_id=$(tail -c 128 "$ansi_file" 2>/dev/null | head -c 7)

  if [ "$sauce_id" != "SAUCE00" ]; then
    # No SAUCE record, return 0 0 to exclude this file
    echo "0 0"
    return
  fi

  # Extract TInfo1 (width) and TInfo2 (height) from SAUCE record
  # SAUCE record is last 128 bytes of file
  # Width is at bytes 96-97 (32 bytes from end of SAUCE)
  # Height is at bytes 98-99 (30 bytes from end of SAUCE)
  width=$(tail -c 128 "$ansi_file" 2>/dev/null | tail -c 32 | head -c 2 | od -An -tu2 | tr -d ' ')
  height=$(tail -c 128 "$ansi_file" 2>/dev/null | tail -c 30 | head -c 2 | od -An -tu2 | tr -d ' ')

  # If parsing failed, return 0 0
  if [ -z "$width" ] || [ -z "$height" ]; then
    echo "0 0"
    return
  fi

  echo "$width $height"
}

# Check if ANSI file fits within terminal dimensions
ansi_fits_terminal() {
  ansi_file="$1"

  dimensions=$(parse_sauce_dimensions "$ansi_file")
  width=$(echo "$dimensions" | cut -d' ' -f1)
  height=$(echo "$dimensions" | cut -d' ' -f2)

  # If no valid SAUCE data, exclude the file
  if [ "$width" = "0" ] || [ "$height" = "0" ]; then
    return 1
  fi

  # Check if width fits (height can overflow)
  if [ "$width" -le "$TERM_COLS" ]; then
    return 0
  else
    return 1
  fi
}

# Find a random ANSI art file that fits the terminal
ansi_art_random_file() {
  # Find all files, excluding common non-ANSI extensions
  # We'll filter by SAUCE metadata later
  file_list=$(find "$ANSI_MOTD_ART_DIR" -type f \
                   ! -iname '*.diz' \
                   ! -iname '*.nfo' \
                   ! -iname '*.txt' \
                   ! -iname '*.zip' \
                   ! -iname '*.exe' \
                   ! -iname '*.jpg' \
                   ! -iname '*.jpeg' \
                   ! -iname '*.png' \
                   ! -iname '*.gif' \
                   ! -iname '*.bmp' \
                   ! -iname '*.pcx' \
                   ! -iname '*.lbm' \
                   ! -iname '*.rip' \
                   ! -iname '*.htm' \
                   ! -iname '*.html' \
                   ! -iname '*.doc' \
                   ! -iname '*.com' \
                   ! -iname '*.class' \
                   ! -iname '*.bat' \
                   ! -iname '*.iso' \
                   ! -iname '*.mod' \
                   ! -iname '*.s3m' \
                   ! -iname '*.xm' \
                   ! -iname '*.it' \
                   ! -iname '*.mp3' \
                   ! -iname '*.wav' \
                   ! -iname '*.voc' \
                   ! -iname '.DS_Store' \
                   2>/dev/null)

  # Count total files
  total_files=$(echo "$file_list" | wc -l | tr -d ' ')

  if [ "$total_files" -eq 0 ]; then
    return
  fi

  # Try up to 50 random files to find one that fits
  max_attempts=50
  attempt=0

  while [ "$attempt" -lt "$max_attempts" ]; do
    # Pick a random file
    random_file=$(echo "$file_list" | shuf -n 1)

    # Check if it fits
    if ansi_fits_terminal "$random_file"; then
      echo "$random_file"
      return
    fi

    attempt=$((attempt + 1))
  done

  # If we couldn't find one after max_attempts, return nothing
  return
}

# Display ANSI art (random or from specified file)
ansi_art_display() {
  # Use provided file or find a random one
  if [ -n "$1" ]; then
    ansi_filename="$1"
    if [ ! -f "$ansi_filename" ]; then
      cat >&2 <<EOF
ansimotd.sh: Error: File not found: $ansi_filename
EOF
      return 1
    fi
  else
    ansi_filename=$(ansi_art_random_file)
  fi

  if [ -n "$ansi_filename" ]; then
    # Get file size and calculate art size (total - 128 byte SAUCE record)
    file_size=$(wc -c < "$ansi_filename" | tr -d ' ')

    # Check if file has SAUCE record and get dimensions
    sauce_id=$(tail -c 128 "$ansi_filename" 2>/dev/null | head -c 7)
    if [ "$sauce_id" = "SAUCE00" ]; then
      art_size=$((file_size - 128))
      # Parse SAUCE width for line wrapping
      dimensions=$(parse_sauce_dimensions "$ansi_filename")
      sauce_width=$(echo "$dimensions" | cut -d' ' -f1)
    else
      art_size=$file_size
      sauce_width=0
    fi

    # Display only the art portion (exclude SAUCE metadata)
    # Convert from Code Page 437 character encoding
    # see https://en.wikipedia.org/wiki/Code_page_437
    # Remove CR characters and wrap at SAUCE width if available
    if [ "$sauce_width" -gt 0 ]; then
      head -c "$art_size" "$ansi_filename" | iconv -f 437 2>/dev/null | tr -d '\r' | \
        python3 -c "
import sys
import re

width = int(sys.argv[1])
data = sys.stdin.read()

# VGA palette - the exact RGB values used by DOS/VGA hardware
# These are the colors ANSI artists designed for
VGA_FG = {
    30: (0,0,0),       # black
    31: (170,0,0),     # red
    32: (0,170,0),     # green
    33: (170,85,0),    # brown/yellow
    34: (0,0,170),     # blue
    35: (170,0,170),   # magenta
    36: (0,170,170),   # cyan
    37: (170,170,170), # light gray
    90: (85,85,85),    # dark gray
    91: (255,85,85),   # light red
    92: (85,255,85),   # light green
    93: (255,255,85),  # yellow
    94: (85,85,255),   # light blue
    95: (255,85,255),  # light magenta
    96: (85,255,255),  # light cyan
    97: (255,255,255), # white
}
VGA_BG = {
    40: (0,0,0),
    41: (170,0,0),
    42: (0,170,0),
    43: (170,85,0),
    44: (0,0,170),
    45: (170,0,170),
    46: (0,170,170),
    47: (170,170,170),
    100: (85,85,85),
    101: (255,85,85),
    102: (85,255,85),
    103: (255,255,85),
    104: (85,85,255),
    105: (255,85,255),
    106: (85,255,255),
    107: (255,255,255),
}

# Bold flag turns dark colors (30-37) into bright colors
bold_active = False
current_fg = None
current_bg = None

def remap_sgr(seq):
    \"\"\"Remap an SGR (color/style) escape sequence to use 24-bit true color.\"\"\"
    global bold_active, current_fg, current_bg
    # Extract the parameter part between ESC[ and m
    inner = seq[2:-1]  # strip ESC[ and m
    if not inner:
        params = [0]
    else:
        params = [int(p) if p else 0 for p in inner.split(';')]

    out_parts = []
    i = 0
    while i < len(params):
        p = params[i]
        if p == 0:  # reset
            bold_active = False
            current_fg = None
            current_bg = None
            out_parts.append('0')
        elif p == 1:  # bold
            bold_active = True
            # If we already have a dark FG set, upgrade it to bright
            if current_fg is not None and 30 <= current_fg <= 37:
                bright = current_fg + 60  # 30->90, etc.
                if bright in VGA_FG:
                    r, g, b = VGA_FG[bright]
                    out_parts.append(f'38;2;{r};{g};{b}')
                    current_fg = bright
        elif p == 22:  # normal intensity
            bold_active = False
            # Downgrade bright FG back to dark if needed
            if current_fg is not None and 90 <= current_fg <= 97:
                dark = current_fg - 60
                if dark in VGA_FG:
                    r, g, b = VGA_FG[dark]
                    out_parts.append(f'38;2;{r};{g};{b}')
                    current_fg = dark
        elif p in VGA_FG:
            # Foreground color
            actual = p
            # If bold is active and this is a dark color, use bright variant
            if bold_active and 30 <= p <= 37:
                actual = p + 60
            if actual in VGA_FG:
                r, g, b = VGA_FG[actual]
                out_parts.append(f'38;2;{r};{g};{b}')
                current_fg = actual
            else:
                out_parts.append(str(p))
        elif p in VGA_BG:
            r, g, b = VGA_BG[p]
            out_parts.append(f'48;2;{r};{g};{b}')
            current_bg = p
        elif p == 38 and i + 1 < len(params) and params[i+1] == 5:
            # 256-color FG - pass through
            end = min(i + 3, len(params))
            out_parts.append(';'.join(str(x) for x in params[i:end]))
            i = end - 1
        elif p == 48 and i + 1 < len(params) and params[i+1] == 5:
            # 256-color BG - pass through
            end = min(i + 3, len(params))
            out_parts.append(';'.join(str(x) for x in params[i:end]))
            i = end - 1
        else:
            out_parts.append(str(p))
        i += 1

    return f'\x1b[{\";\".join(out_parts)}m'

# Split on existing newlines first
lines = data.split('\n')

for line in lines:
    col = 0
    i = 0
    while i < len(line):
        # Check for ANSI escape sequence
        if line[i] == '\033' or line[i] == '\x1b':
            # Find the end of the escape sequence
            j = i + 1
            while j < len(line) and not line[j].isalpha():
                j += 1

            # Extract the escape sequence
            seq = line[i:j+1]

            # Remap SGR (color) sequences to true-color VGA palette
            if j < len(line) and line[j] == 'm':
                seq = remap_sgr(seq)
            # Check if this is a cursor movement sequence
            elif j < len(line):
                match = re.match(r'\x1b\[(\d+)?;?(\d+)?([HfABCDEFG])', seq)
                if match:
                    num1 = int(match.group(1)) if match.group(1) else 1
                    num2 = int(match.group(2)) if match.group(2) else 1
                    cmd = match.group(3)

                    if cmd == 'C':  # Cursor forward
                        col += num1
                    elif cmd == 'D':  # Cursor backward
                        col = max(0, col - num1)
                    elif cmd == 'G':  # Cursor horizontal absolute
                        col = num1 - 1
                    elif cmd in 'Hf':  # Cursor position (row;col)
                        if match.group(2):  # Has column specified
                            col = num2 - 1

            # Print the (possibly remapped) escape sequence
            sys.stdout.write(seq)
            i = j + 1
        else:
            # Handle special characters
            if line[i] == '\t':
                # Tab moves to next multiple of 8
                next_tab = ((col // 8) + 1) * 8
                sys.stdout.write(line[i])
                col = next_tab
                i += 1
            elif line[i] == '\b':
                # Backspace moves back one column
                sys.stdout.write(line[i])
                col = max(0, col - 1)
                i += 1
            else:
                # Regular character
                if col >= width:
                    sys.stdout.write('\n')
                    col = 0
                sys.stdout.write(line[i])
                col += 1
                i += 1
    sys.stdout.write('\n')
" "$sauce_width"
    else
      head -c "$art_size" "$ansi_filename" | iconv -f 437 2>/dev/null | tr -d '\r' | \
        python3 -c "
import sys
import re

data = sys.stdin.read()

# VGA palette for non-SAUCE path
VGA_FG = {
    30: (0,0,0), 31: (170,0,0), 32: (0,170,0), 33: (170,85,0),
    34: (0,0,170), 35: (170,0,170), 36: (0,170,170), 37: (170,170,170),
    90: (85,85,85), 91: (255,85,85), 92: (85,255,85), 93: (255,255,85),
    94: (85,85,255), 95: (255,85,255), 96: (85,255,255), 97: (255,255,255),
}
VGA_BG = {
    40: (0,0,0), 41: (170,0,0), 42: (0,170,0), 43: (170,85,0),
    44: (0,0,170), 45: (170,0,170), 46: (0,170,170), 47: (170,170,170),
    100: (85,85,85), 101: (255,85,85), 102: (85,255,85), 103: (255,255,85),
    104: (85,85,255), 105: (255,85,255), 106: (85,255,255), 107: (255,255,255),
}
bold_active = False
current_fg = None

def remap_sgr(seq):
    global bold_active, current_fg
    inner = seq[2:-1]
    if not inner:
        params = [0]
    else:
        params = [int(p) if p else 0 for p in inner.split(';')]
    out_parts = []
    i = 0
    while i < len(params):
        p = params[i]
        if p == 0:
            bold_active = False; current_fg = None; out_parts.append('0')
        elif p == 1:
            bold_active = True
            if current_fg is not None and 30 <= current_fg <= 37:
                bright = current_fg + 60
                if bright in VGA_FG:
                    r, g, b = VGA_FG[bright]; out_parts.append(f'38;2;{r};{g};{b}'); current_fg = bright
        elif p == 22:
            bold_active = False
        elif p in VGA_FG:
            actual = p
            if bold_active and 30 <= p <= 37: actual = p + 60
            if actual in VGA_FG:
                r, g, b = VGA_FG[actual]; out_parts.append(f'38;2;{r};{g};{b}'); current_fg = actual
            else: out_parts.append(str(p))
        elif p in VGA_BG:
            r, g, b = VGA_BG[p]; out_parts.append(f'48;2;{r};{g};{b}')
        else: out_parts.append(str(p))
        i += 1
    return f'\x1b[{\";\".join(out_parts)}m'

def remap_line(line):
    result = []
    i = 0
    while i < len(line):
        if line[i] == '\x1b' and i + 1 < len(line) and line[i+1] == '[':
            j = i + 2
            while j < len(line) and not line[j].isalpha():
                j += 1
            if j < len(line):
                seq = line[i:j+1]
                if line[j] == 'm':
                    seq = remap_sgr(seq)
                result.append(seq)
                i = j + 1
            else:
                result.append(line[i]); i += 1
        else:
            result.append(line[i]); i += 1
    return ''.join(result)

for line in data.split('\n'):
    sys.stdout.write(remap_line(line) + '\n')
"
    fi

    # Ensure output ends with a newline
    echo ""

    # Record the filename for later reference
    ANSI_MOTD_FILENAME="$ansi_filename"
    export ANSI_MOTD_FILENAME
  else
    cat >&2 <<EOF
ansimotd.sh:
I couldn't find any ANSI art to display that fits your terminal width (${TERM_COLS} columns).
I tried looking in '$ANSI_MOTD_ART_DIR'.
For help on getting ANSI art, see:
https://github.com/retlehs/ansimotd#getting-ansi-art
EOF
  fi
}

# Main execution
ANSI_MOTD_ART_DIR="${XDG_CONFIG_HOME:-$HOME/.config}/ansimotd"
export ANSI_MOTD_ART_DIR

# Parse command-line arguments
MANUAL_FILE=""
while [ $# -gt 0 ]; do
  case "$1" in
    -f|--file)
      MANUAL_FILE="$2"
      shift 2
      ;;
    *)
      shift
      ;;
  esac
done

# Create art directory if it doesn't exist
[ -d "$ANSI_MOTD_ART_DIR" ] || mkdir -p "$ANSI_MOTD_ART_DIR"

# Get terminal dimensions
get_terminal_dimensions

# Display ANSI art
ansi_art_display "$MANUAL_FILE"
