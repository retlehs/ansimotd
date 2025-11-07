# ansimotd

A shell-agnostic script that displays random ANSI art as a MOTD when you start a shell.

Supports files with valid [SAUCE metadata](https://www.acid.org/info/sauce/sauce.htm), which contain the width and height of the art. This script uses the metadata to ensure the art fits your terminal.

![Example MOTD](./example.png)

## Requirements

* coreutils

## Installation

1. Clone this repository:

```bash
git clone https://github.com/retlehs/ansimotd.git
cd ansimotd
```

2. Source the script in your shell's RC file:

```bash
source /path/to/ansimotd/ansimotd.sh
```

3. Download some ANSI art (see below)

### Getting ANSI art to display

After installation, you'll need to download ANSI art files. This script supports ANSI files from [16colo.rs](https://16colo.rs/).

#### Using the download script

```bash
# Download all packs from a specific year
./download-art.sh 1996

# Download a specific pack from a year
./download-art.sh 1999/bmbook20

# Download all packs from a group for a specific year
./download-art.sh 1999/rmrs

# Download all packs from a group across all years
./download-art.sh ice
./download-art.sh acid

# Download all art (this is large!)
./download-art.sh all

# Show usage information
./download-art.sh
```

## Configuration

The script exports the following environment variables:

* `ANSI_MOTD_ART_DIR` - Directory where ANSI art is stored (default: `${XDG_CONFIG_HOME:-$HOME/.config}/ansimotd`)
* `ANSI_MOTD_FILENAME` - Full path to the last displayed ANSI art file

You can override the art directory by setting `ANSI_MOTD_ART_DIR` before sourcing the script:

```bash
export ANSI_MOTD_ART_DIR="$HOME/my-ansi-art"
source /path/to/ansimotd/ansimotd.sh
```

## Credits

Based on [zsh-ansimotd](https://github.com/yuhonas/zsh-ansimotd) by yuhonas, with significant modifications for shell-agnostic support and dimension filtering.
