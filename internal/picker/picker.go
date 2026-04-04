package picker

import (
	"errors"
	"math/rand/v2"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/term"

	"github.com/retlehs/ansimotd/internal/sauce"
)

var excludedExts = map[string]bool{
	".diz": true, ".nfo": true, ".txt": true, ".zip": true, ".exe": true,
	".jpg": true, ".jpeg": true, ".png": true, ".gif": true, ".bmp": true,
	".pcx": true, ".lbm": true, ".rip": true, ".htm": true, ".html": true,
	".doc": true, ".com": true, ".class": true, ".bat": true, ".iso": true,
	".mod": true, ".s3m": true, ".xm": true, ".it": true,
	".mp3": true, ".wav": true, ".voc": true,
}

// Pick returns the path to a random ANSI art file that fits the terminal width.
// artDir is the root directory to search. Returns an error if no file fits.
func Pick(artDir string) (string, error) {
	cols := termWidth()

	var files []string
	err := filepath.WalkDir(artDir, func(path string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		ext := strings.ToLower(filepath.Ext(path))
		if !excludedExts[ext] && d.Name() != ".DS_Store" {
			files = append(files, path)
		}
		return nil
	})
	if err != nil {
		return "", err
	}
	if len(files) == 0 {
		return "", errors.New("no ANSI art files found")
	}

	// Fisher-Yates shuffle
	rand.Shuffle(len(files), func(i, j int) {
		files[i], files[j] = files[j], files[i]
	})

	for _, path := range files {
		rec, err := sauce.Parse(path)
		if err != nil {
			continue
		}
		// Skip: no SAUCE, or invalid dimensions
		if rec == nil || rec.Width == 0 || rec.Height == 0 {
			continue
		}
		// Skip: too wide
		if rec.Width > cols {
			continue
		}
		return path, nil
	}

	return "", errors.New("no ANSI art fits the terminal width")
}

func termWidth() int {
	if w, _, err := term.GetSize(int(os.Stdout.Fd())); err == nil && w > 0 {
		return w
	}
	if s := os.Getenv("COLUMNS"); s != "" {
		if n := 0; func() bool { var err error; n, err = parseInt(s); return err == nil }() && n > 0 {
			return n
		}
	}
	return 80
}

func parseInt(s string) (int, error) {
	n := 0
	for _, c := range s {
		if c < '0' || c > '9' {
			return 0, errors.New("not a number")
		}
		n = n*10 + int(c-'0')
	}
	return n, nil
}
