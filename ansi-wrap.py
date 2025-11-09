#!/usr/bin/env python3
import sys
import re

width = int(sys.argv[1])
data = sys.stdin.read()

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
            seq_start = j
            while j < len(line) and not line[j].isalpha():
                j += 1

            # Extract the escape sequence
            seq = line[i:j+1]

            # Check if this is a cursor movement sequence
            if j < len(line):
                # Cursor position with row;col
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

            # Print the whole escape sequence
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
