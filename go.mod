module github.com/kernelplex/ubase

go 1.24.0

tool github.com/kernelplex/evercore/cmd/evercoregen

require (
	github.com/joho/godotenv v1.5.1
	github.com/kernelplex/evercore v0.0.36
	github.com/lib/pq v1.10.9
	github.com/mattn/go-sqlite3 v1.14.24
	github.com/olekukonko/tablewriter v1.0.8
	github.com/pquerna/otp v1.5.0
	github.com/pressly/goose/v3 v3.24.1
	github.com/xo/dburl v0.23.8
	golang.org/x/crypto v0.37.0
	golang.org/x/term v0.31.0
)

replace github.com/kernelplex/evercore => /mnt/data/seddy/proj/evercore

require (
	github.com/bmatcuk/doublestar/v4 v4.8.1 // indirect
	github.com/boombuler/barcode v1.0.1-0.20190219062509-6c824513bacc // indirect
	github.com/fatih/color v1.15.0 // indirect
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/mattn/go-runewidth v0.0.16 // indirect
	github.com/mfridman/interpolate v0.0.2 // indirect
	github.com/olekukonko/errors v0.0.0-20250405072817-4e6d85265da6 // indirect
	github.com/olekukonko/ll v0.0.8 // indirect
	github.com/rivo/uniseg v0.2.0 // indirect
	github.com/sethvargo/go-retry v0.3.0 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	golang.org/x/sync v0.15.0 // indirect
	golang.org/x/sys v0.32.0 // indirect
)
