package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

const usage = "usage: release-checksum create <file>... | verify <manifest>"

func main() {
	if err := run(os.Args[1:], os.Stdout); err != nil {
		fmt.Fprintf(os.Stderr, "release-checksum: %v\n", err)
		os.Exit(1)
	}
}

func run(args []string, stdout io.Writer) error {
	if len(args) == 0 {
		return fmt.Errorf("%s", usage)
	}

	switch args[0] {
	case "create":
		if len(args) < 2 {
			return fmt.Errorf("create requires at least one file: %s", usage)
		}
		return writeManifest(stdout, args[1:])
	case "verify":
		if len(args) != 2 {
			return fmt.Errorf("verify requires one manifest: %s", usage)
		}
		manifest, err := os.Open(args[1])
		if err != nil {
			return fmt.Errorf("open manifest: %w", err)
		}
		defer manifest.Close()

		count, err := verifyManifest(manifest, filepath.Dir(args[1]))
		if err != nil {
			return err
		}
		fmt.Fprintf(stdout, "verified %d file(s)\n", count)
		return nil
	default:
		return fmt.Errorf("unknown command %q: %s", args[0], usage)
	}
}
