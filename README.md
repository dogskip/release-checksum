# release-checksum

A small Go command-line tool for creating and verifying SHA-256 manifests for
release artifacts. It uses only the Go standard library and produces the common
`<sha256>  <path>` format used by `SHA256SUMS` files.

## Build

```bash
go build -o release-checksum .
```

## Create a manifest

Run the tool from the directory containing the release artifacts so the manifest
contains portable relative paths.

```bash
cd dist
../release-checksum create app-darwin.zip app-linux.tar.gz > SHA256SUMS
```

Entries are sorted by path, so identical inputs produce a stable manifest.

## Verify a manifest

Relative paths are resolved beside the manifest file, so verification works from
any current directory.

```bash
release-checksum verify dist/SHA256SUMS
```

The command exits non-zero for malformed entries, missing files, and checksum
mismatches.

## Development

```bash
gofmt -w .
go vet ./...
go test ./...
```
