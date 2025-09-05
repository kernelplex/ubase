
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
	go test -count=1 -tags sqlite,postgresql ./integration_tests/...

.PHONY: test-postgresql
test-postgresql: sqlc evercoregen
	go test -count=1 -tags postgresql ./integration_tests/...

.PHONY: test-sqlite
test-sqlite: sqlc evercoregen
	go test -count=1 -tags sqlite ./integration_tests/...

