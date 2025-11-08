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
  if command -v fd >/dev/null 2>&1; then
    file_list=$(fd --type f \
                   --absolute-path \
                   --exclude '*.diz' \
                   --exclude '*.nfo' \
                   --exclude '*.txt' \
                   --exclude '*.zip' \
                   --exclude '*.exe' \
                   --search-path "$ANSI_MOTD_ART_DIR" 2>/dev/null)
  else
    file_list=$(find "$ANSI_MOTD_ART_DIR" -type f \
                     ! -iname '*.diz' \
                     ! -iname '*.nfo' \
                     ! -iname '*.txt' \
                     ! -iname '*.zip' \
                     ! -iname '*.exe' \
                     2>/dev/null)
  fi

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

# Display random ANSI art
ansi_art_random() {
  ansi_filename=$(ansi_art_random_file)

  if [ -n "$ansi_filename" ]; then
    # Get file size and calculate art size (total - 128 byte SAUCE record)
    file_size=$(wc -c < "$ansi_filename" | tr -d ' ')
    art_size=$((file_size - 128))

    # Display only the art portion (exclude SAUCE metadata)
    # Convert from Code Page 437 character encoding
    # see https://en.wikipedia.org/wiki/Code_page_437
    head -c "$art_size" "$ansi_filename" | iconv -f 437 2>/dev/null

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

# Create art directory if it doesn't exist
[ -d "$ANSI_MOTD_ART_DIR" ] || mkdir -p "$ANSI_MOTD_ART_DIR"

# Get terminal dimensions
get_terminal_dimensions

# Display random ANSI art
ansi_art_random
