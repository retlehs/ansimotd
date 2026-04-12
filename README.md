# ansimotd

[![Website](https://img.shields.io/badge/ansimotd.com-4a4a4a?style=flat-square)](https://ansimotd.com) [![X](https://img.shields.io/badge/@ansimotd-000000?style=flat-square&logo=x)](https://x.com/ansimotd) [![Bluesky](https://img.shields.io/badge/@ansimotd.com-0085ff?style=flat-square&logo=bluesky&logoColor=white)](https://bsky.app/profile/ansimotd.com)

Display random ANSI art in your terminal as a message of the day. For a browser version, visit [ansimotd.com](https://ansimotd.com).

- Selects a random ANSI art file with valid [SAUCE metadata](https://www.acid.org/info/sauce/sauce.htm)
- Ensures art fits your terminal width
- Converts ANSI color codes to 24-bit true-color escape sequences using exact VGA RGB values, so art displays accurately regardless of your terminal's color scheme
- Art sourced from [16colo.rs](https://16colo.rs/)

![Example MOTD](./example.png)

## Installation

### macOS

```bash
brew tap retlehs/tap
brew install retlehs/tap/ansimotd
```

### Ubuntu/Debian

```
curl -fsSL https://apt.ansimotd.com/install.sh | sudo sh
```

### Go

```bash
go install github.com/retlehs/ansimotd@latest
```

### Download art

ansimotd ships without any art — you need to download at least one pack before it can display anything. A good starting point:

```bash
ansimotd download 1996
ansimotd download 1997
ansimotd download 1998
ansimotd download 1999
```

### Shell integration

Add to your shell RC file (`.bashrc`, `.zshrc`, etc.) to display a random piece of art each time you open a terminal:

```bash
ansimotd
```

## Usage

```bash
# Display a random ANSI art file
ansimotd

# Display a specific file
ansimotd display --file /path/to/art.ans
```

### Targeted downloads

```bash
# Filter by group
ansimotd download 1996 --group ice

# Download a specific pack
ansimotd download 1999 --pack bmbook20
```

### Other commands

```bash
# Print the path of the last displayed file
ansimotd last

# Print version
ansimotd --version
```

## Configuration

| Environment variable | Description | Default |
|---|---|---|
| `ANSIMOTD_DIR` | Root directory for all ansimotd data | `$XDG_CONFIG_HOME/ansimotd` or `~/.config/ansimotd` |

All paths derive from the root:

| Path | Purpose |
|---|---|
| `$ROOT/art/` | Downloaded art packs |
| `$ROOT/last` | Path to the last displayed file |

## Credits

Based on [zsh-ansimotd](https://github.com/yuhonas/zsh-ansimotd) by yuhonas, with significant modifications including accurate VGA color rendering, SAUCE-aware line wrapping, and the 16colo.rs download client.
