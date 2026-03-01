
main: templ evercoregen sqlc
	go build -o build/ubase cmd/ubase/main.go

evercoregen:
	go tool evercoregen -output-dir=internal/evercoregen/ -output-pkg=evercoregen

sqlc: 
	sqlc generate

test: sqlc evercoregen
	go test -count=1 ./...
	go test -count=1 -tags sqlite,postgresql ./integration_tests/...

test-postgresql: sqlc evercoregen
	go test -count=1 -tags postgresql ./integration_tests/...

test-sqlite: sqlc evercoregen
	go test -count=1 -tags sqlite ./integration_tests/...

templ:
	go tool templ generate

lint:
	./scripts/check_route_style.sh
