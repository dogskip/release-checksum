package main

import (
	"bufio"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

func writeManifest(w io.Writer, paths []string) error {
	sorted := append([]string(nil), paths...)
	sort.Strings(sorted)

	for _, path := range sorted {
		if strings.ContainsAny(path, "\r\n") {
			return fmt.Errorf("path contains a line break: %q", path)
		}
		sum, err := hashRegularFile(path)
		if err != nil {
			return err
		}
		if _, err := fmt.Fprintf(w, "%x  %s\n", sum, filepath.ToSlash(path)); err != nil {
			return fmt.Errorf("write manifest: %w", err)
		}
	}
	return nil
}

func verifyManifest(r io.Reader, baseDir string) (int, error) {
	scanner := bufio.NewScanner(r)
	verified := 0

	for lineNumber := 1; scanner.Scan(); lineNumber++ {
		line := scanner.Text()
		if strings.TrimSpace(line) == "" {
			continue
		}
		parts := strings.SplitN(line, "  ", 2)
		if len(parts) != 2 || parts[1] == "" {
			return verified, fmt.Errorf("manifest line %d must contain a checksum, two spaces, and a path", lineNumber)
		}

		expected, err := hex.DecodeString(parts[0])
		if err != nil || len(expected) != sha256.Size {
			return verified, fmt.Errorf("manifest line %d has an invalid SHA-256 checksum", lineNumber)
		}

		path := filepath.FromSlash(parts[1])
		if !filepath.IsAbs(path) {
			path = filepath.Join(baseDir, path)
		}
		actual, err := hashRegularFile(path)
		if err != nil {
			return verified, fmt.Errorf("manifest line %d: %w", lineNumber, err)
		}
		if subtle.ConstantTimeCompare(expected, actual[:]) != 1 {
			return verified, fmt.Errorf("checksum mismatch for %q", parts[1])
		}
		verified++
	}
	if err := scanner.Err(); err != nil {
		return verified, fmt.Errorf("read manifest: %w", err)
	}
	if verified == 0 {
		return 0, fmt.Errorf("manifest contains no entries")
	}
	return verified, nil
}

func hashRegularFile(path string) ([sha256.Size]byte, error) {
	var empty [sha256.Size]byte
	file, err := os.Open(path)
	if err != nil {
		return empty, fmt.Errorf("open %q: %w", path, err)
	}
	defer file.Close()

	info, err := file.Stat()
	if err != nil {
		return empty, fmt.Errorf("stat %q: %w", path, err)
	}
	if !info.Mode().IsRegular() {
		return empty, fmt.Errorf("%q is not a regular file", path)
	}

	hasher := sha256.New()
	if _, err := io.Copy(hasher, file); err != nil {
		return empty, fmt.Errorf("hash %q: %w", path, err)
	}
	var sum [sha256.Size]byte
	copy(sum[:], hasher.Sum(nil))
	return sum, nil
}
