package main

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestWriteManifestSortsPaths(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)
	writeFile(t, "b.txt", "bravo")
	writeFile(t, "a.txt", "alpha")

	var output bytes.Buffer
	if err := writeManifest(&output, []string{"b.txt", "a.txt"}); err != nil {
		t.Fatalf("writeManifest() error = %v", err)
	}

	aSum := sha256.Sum256([]byte("alpha"))
	bSum := sha256.Sum256([]byte("bravo"))
	want := fmt.Sprintf("%x  a.txt\n%x  b.txt\n", aSum, bSum)
	if output.String() != want {
		t.Fatalf("manifest mismatch\nwant:\n%s\ngot:\n%s", want, output.String())
	}
}

func TestVerifyManifestAcceptsValidFiles(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir+"/artifact.zip", "release bytes")
	sum := sha256.Sum256([]byte("release bytes"))
	manifest := fmt.Sprintf("%x  artifact.zip\n", sum)

	count, err := verifyManifest(strings.NewReader(manifest), dir)
	if err != nil {
		t.Fatalf("verifyManifest() error = %v", err)
	}
	if count != 1 {
		t.Fatalf("verifyManifest() count = %d, want 1", count)
	}
}

func TestVerifyManifestDetectsTampering(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir+"/artifact.zip", "original")
	sum := sha256.Sum256([]byte("original"))
	writeFile(t, dir+"/artifact.zip", "tampered")
	manifest := fmt.Sprintf("%x  artifact.zip\n", sum)

	_, err := verifyManifest(strings.NewReader(manifest), dir)
	if err == nil || !strings.Contains(err.Error(), "checksum mismatch") {
		t.Fatalf("verifyManifest() error = %v, want checksum mismatch", err)
	}
}

func TestVerifyManifestRejectsMalformedEntry(t *testing.T) {
	_, err := verifyManifest(strings.NewReader("not-a-manifest-line\n"), t.TempDir())
	if err == nil || !strings.Contains(err.Error(), "two spaces") {
		t.Fatalf("verifyManifest() error = %v, want format error", err)
	}
}

func TestWriteManifestRejectsDirectory(t *testing.T) {
	var output bytes.Buffer
	err := writeManifest(&output, []string{t.TempDir()})
	if err == nil || !strings.Contains(err.Error(), "not a regular file") {
		t.Fatalf("writeManifest() error = %v, want regular-file error", err)
	}
}

func TestRunCreatesAndVerifiesManifest(t *testing.T) {
	dir := t.TempDir()
	artifact := filepath.Join(dir, "artifact.zip")
	manifestPath := filepath.Join(dir, "SHA256SUMS")
	writeFile(t, artifact, "release bytes")

	var manifest bytes.Buffer
	if err := run([]string{"create", artifact}, &manifest); err != nil {
		t.Fatalf("run(create) error = %v", err)
	}
	writeFile(t, manifestPath, manifest.String())

	var output bytes.Buffer
	if err := run([]string{"verify", manifestPath}, &output); err != nil {
		t.Fatalf("run(verify) error = %v", err)
	}
	if output.String() != "verified 1 file(s)\n" {
		t.Fatalf("run(verify) output = %q", output.String())
	}
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("os.WriteFile(%q) error = %v", path, err)
	}
}
