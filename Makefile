
.PHONY: evercoregen
evercoregen:
	go tool evercoregen -output-dir=lib/evercoregen/ -output-pkg=evercoregen

.PHONY: sqlc
sqlc: 
	sqlc generate

