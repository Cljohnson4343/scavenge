api-test: clean-db
	@echo "Running api tests...\n"
	@go test -v -tags apiTest ./...

clean-db:
	@echo "\nCleaning the database...\n"
	psql -d scavengedb_test -f ./db/scavenge_schema.sql