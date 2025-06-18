
main: evercoregen sqlc
	go build -o build/example examples/*.go

.PHONY: evercoregen
evercoregen:
	go tool evercoregen -output-dir=lib/evercoregen/ -output-pkg=evercoregen

.PHONY: sqlc
sqlc: 
	sqlc generate

