
main: evercoregen sqlc
	go build -o build/example examples/*.go

.PHONY: evercoregen
evercoregen:
	go tool evercoregen -output-dir=internal/evercoregen/ -output-pkg=evercoregen

.PHONY: sqlc
sqlc: 
	sqlc generate

.PHONY: test
test: sqlc evercoregen main
	go test integration_tests/*.go
