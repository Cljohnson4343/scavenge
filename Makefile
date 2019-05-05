api-test: clean-db
	@echo "Running api tests...\n"
	@go test -tags=apiTest ./... 

clean-db:
	@echo "\nCleaning the database...\n"
	psql -d scavengedb_test -f ./db/scavenge_schema.sql

unit:
	@echo "Running unit tests...\n"
	@go test -tags=unit ./...

db: clean-db
	@echo "Opening test db..."
	@psql -d scavengedb_test