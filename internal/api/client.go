package api

import (
	"archive/zip"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"unicode/utf8"

	"golang.org/x/text/encoding/charmap"
)

const baseURL = "https://api.16colo.rs/v1"

type pageInfo struct {
	Total    int `json:"total"`
	PageSize int `json:"pagesize"`
	Page     int `json:"page"`
	Pages    int `json:"pages"`
}

// flexString handles JSON fields that may be a string or number.
type flexString string

func (f *flexString) UnmarshalJSON(b []byte) error {
	// Try string first
	var s string
	if err := json.Unmarshal(b, &s); err == nil {
		*f = flexString(s)
		return nil
	}
	// Fall back to number
	var n json.Number
	if err := json.Unmarshal(b, &n); err == nil {
		*f = flexString(n.String())
		return nil
	}
	return fmt.Errorf("cannot unmarshal %s as string or number", string(b))
}

// Pack represents an art pack from the API.
type Pack struct {
	Year     int          `json:"year"`
	Name     flexString   `json:"name"`
	Download string       `json:"download"`
	Groups   []flexString `json:"groups"`
}

type response struct {
	Page    pageInfo `json:"page"`
	Results []Pack   `json:"results"`
}

// FetchPacks retrieves all packs for a given year, applying optional filters.
func FetchPacks(year int, group, pack string) ([]Pack, error) {
	var all []Pack
	page := 1

	for {
		url := fmt.Sprintf("%s/year/%d?page=%d", baseURL, year, page)
		resp, err := http.Get(url)
		if err != nil {
			return nil, fmt.Errorf("fetching page %d: %w", page, err)
		}

		var r response
		err = json.NewDecoder(resp.Body).Decode(&r)
		_ = resp.Body.Close()
		if err != nil {
			return nil, fmt.Errorf("decoding page %d: %w", page, err)
		}

		all = append(all, r.Results...)

		if page >= r.Page.Pages {
			break
		}
		page++
	}

	// Apply filters
	if group != "" || pack != "" {
		var filtered []Pack
		for _, p := range all {
			if pack != "" && !strings.EqualFold(string(p.Name), pack) {
				continue
			}
			if group != "" && !matchesGroup(p, group) {
				continue
			}
			filtered = append(filtered, p)
		}
		all = filtered
	}

	return all, nil
}

func matchesGroup(p Pack, group string) bool {
	for _, g := range p.Groups {
		if strings.EqualFold(string(g), group) {
			return true
		}
	}
	// Also match by pack name prefix (e.g. "ice" matches "ice-0196")
	return strings.HasPrefix(strings.ToLower(string(p.Name)), strings.ToLower(group))
}

// DownloadAndExtract downloads a pack archive and extracts ANSI files to destDir.
// Returns true if the pack was downloaded, false if skipped (already exists).
func DownloadAndExtract(p Pack, destDir string) (bool, error) {
	packDir := filepath.Join(destDir, fmt.Sprintf("%d", p.Year), string(p.Name))

	// Skip if already downloaded
	if dirHasFiles(packDir) {
		return false, nil
	}

	if p.Download == "" {
		return false, errors.New("pack has no download URL")
	}

	// Clean up any leftover temp files from previous interrupted runs
	tmpPattern := packDir + ".*.tmp"
	if matches, _ := filepath.Glob(tmpPattern); len(matches) > 0 {
		for _, m := range matches {
			_ = os.Remove(m)
		}
	}

	// Ensure parent directory exists before creating temp file
	if err := os.MkdirAll(filepath.Dir(packDir), 0o755); err != nil {
		return false, fmt.Errorf("creating directory: %w", err)
	}

	tmpFile, err := os.CreateTemp(filepath.Dir(packDir), filepath.Base(packDir)+".*.tmp")
	if err != nil {
		return false, fmt.Errorf("creating temp file: %w", err)
	}
	tmpPath := tmpFile.Name()

	// Ensure cleanup on failure
	success := false
	defer func() {
		if !success {
			_ = os.Remove(tmpPath)
			_ = os.RemoveAll(packDir)
		}
	}()

	resp, err := http.Get(p.Download)
	if err != nil {
		_ = tmpFile.Close()
		return false, fmt.Errorf("downloading %s: %w", p.Name, err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		_ = tmpFile.Close()
		return false, fmt.Errorf("downloading %s: HTTP %d", p.Name, resp.StatusCode)
	}

	if _, err := io.Copy(tmpFile, resp.Body); err != nil {
		_ = tmpFile.Close()
		return false, fmt.Errorf("writing %s: %w", p.Name, err)
	}
	_ = tmpFile.Close()

	// Extract zip
	if err := extractZip(tmpPath, packDir); err != nil {
		return false, fmt.Errorf("extracting %s: %w", p.Name, err)
	}

	_ = os.Remove(tmpPath)
	success = true
	return true, nil
}

// ANSI-eligible extensions (inverse of excluded set)
var ansiExts = map[string]bool{
	".ans": true, ".asc": true, ".ice": true, ".bin": true,
}

func extractZip(zipPath, destDir string) error {
	r, err := zip.OpenReader(zipPath)
	if err != nil {
		return err
	}
	defer func() { _ = r.Close() }()

	if err := os.MkdirAll(destDir, 0o755); err != nil {
		return err
	}

	for _, f := range r.File {
		if f.FileInfo().IsDir() {
			continue
		}

		// Decode CP437 filenames to UTF-8 (DOS-era zip archives)
		name := filepath.Base(f.Name)
		if f.NonUTF8 || !utf8.ValidString(name) {
			if decoded, err := charmap.CodePage437.NewDecoder().String(name); err == nil {
				name = decoded
			}
		}

		// Path traversal protection
		if strings.HasPrefix(name, "..") {
			continue
		}

		// Only extract ANSI-eligible files
		ext := strings.ToLower(filepath.Ext(name))
		if !ansiExts[ext] {
			continue
		}

		outPath := filepath.Join(destDir, name)
		if err := extractFile(f, outPath); err != nil {
			return err
		}
	}

	return nil
}

func extractFile(f *zip.File, outPath string) error {
	rc, err := f.Open()
	if err != nil {
		return err
	}
	defer func() { _ = rc.Close() }()

	outFile, err := os.Create(outPath)
	if err != nil {
		return err
	}
	defer func() { _ = outFile.Close() }()

	_, err = io.Copy(outFile, rc)
	return err
}

func dirHasFiles(dir string) bool {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return false
	}
	return len(entries) > 0
}
