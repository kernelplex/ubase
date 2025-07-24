
main: evercoregen sqlc
	go build -o build/ubase cmd/ubase/main.go

.PHONY: evercoregen
evercoregen:
	go tool evercoregen -output-dir=internal/evercoregen/ -output-pkg=evercoregen

.PHONY: sqlc
sqlc: 
	sqlc generate

.PHONY: test
test: sqlc evercoregen
	go test integration_tests/*.go
