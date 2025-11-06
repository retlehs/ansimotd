#!/bin/sh

set -e

ANSI_MOTD_ART_DIR="${XDG_CONFIG_HOME:-$HOME/.config}/ansimotd"

# Create directory if it doesn't exist
mkdir -p "$ANSI_MOTD_ART_DIR"

echo "Downloading ANSI art from 16colo.rs..."
echo "This may take a while depending on your selection."
echo ""

# Check if rsync is available
if ! command -v rsync >/dev/null 2>&1; then
  echo "Error: rsync is required but not installed." >&2
  echo "Please install rsync and try again." >&2
  exit 1
fi

# Parse arguments
if [ -z "$1" ]; then
  cat <<EOF
Usage: $0 [year|all|year/pack|year/group|group]

Examples:
  $0 1996                # Download all packs from 1996
  $0 1997/bmbook14       # Download specific pack 'bmbook14' from 1997
  $0 1999/ice            # Download all 'ice' packs from 1999
  $0 1997/acid           # Download all 'acid' packs from 1997
  $0 ice                 # Download all 'ice' packs from all years
  $0 acid                # Download all 'acid' packs from all years
  $0 all                 # Download all ANSI art (large!)

Available years: 1990-present
Common groups: ice, acid, fire, blocktronics, impure, etc.

The art will be downloaded to: $ANSI_MOTD_ART_DIR
EOF
  exit 0
fi

if [ "$1" = "all" ]; then
  echo "Downloading all ANSI art from 16colo.rs..."
  rsync -azvhP \
    --prune-empty-dirs \
    --include='*/' \
    --exclude='*.diz' \
    --exclude='*.nfo' \
    --exclude='*.txt' \
    --exclude='*.zip' \
    rsync://16colo.rs/pack/ "$ANSI_MOTD_ART_DIR/"
elif echo "$1" | grep -q '/'; then
  # Downloading from a specific year with pattern (format: year/group or year/pack)
  input="$1"
  year=$(echo "$input" | cut -d'/' -f1)
  group=$(echo "$input" | cut -d'/' -f2)

  echo "Downloading '$group' packs from year $year..."
  # Use rsync's filter rules to match group prefix
  # Download all files - we'll filter by SAUCE metadata when displaying
  rsync -azvhP \
    --prune-empty-dirs \
    --include="${group}*/" \
    --include="${group}*/**/" \
    --exclude='*.diz' \
    --exclude='*.nfo' \
    --exclude='*.txt' \
    --exclude='*.zip' \
    --exclude='*/' \
    rsync://16colo.rs/pack/"$year"/ "$ANSI_MOTD_ART_DIR/$year/"
else
  # Check if it's a year (4 digits) or a group name
  if echo "$1" | grep -qE '^[0-9]{4}$'; then
    # Downloading all packs from a year
    year="$1"
    echo "Downloading all packs from year $year..."
    rsync -azvhP \
      --prune-empty-dirs \
      --include='*/' \
      --exclude='*.diz' \
      --exclude='*.nfo' \
      --exclude='*.txt' \
      --exclude='*.zip' \
      rsync://16colo.rs/pack/"$year" "$ANSI_MOTD_ART_DIR/"
  else
    # Downloading a group from all years
    group="$1"
    echo "Downloading all '$group' packs from all years..."
    rsync -azvhP \
      --prune-empty-dirs \
      --include='*/' \
      --include="${group}*/" \
      --include="${group}*/**/" \
      --exclude='*.diz' \
      --exclude='*.nfo' \
      --exclude='*.txt' \
      --exclude='*.zip' \
      --exclude='*/' \
      rsync://16colo.rs/pack/ "$ANSI_MOTD_ART_DIR/"
  fi
fi

echo ""
echo "Download complete!"
echo "Art saved to: $ANSI_MOTD_ART_DIR"
