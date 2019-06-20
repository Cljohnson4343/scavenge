integration: clean-db-test
	@echo "Running integration tests...\n"
	@go test -tags=integration -count=1 ./... 

clean-db-test:
	@echo "\nCleaning the test database...\n"
	psql -d scavengedb_test -f ./db/scavenge_schema.sql

clean-db:
	@echo "\nCleaning the database...\n"
	psql -d scavengedb -f ./db/scavenge_schema.sql

unit: 
	@echo "Running unit tests...\n"
	@go test -tags=unit ./...

start-db-test: clean-db-test
	@echo "Opening test db..."
	@psql -d scavengedb_test

start-db:
	@echo "Opening db..."
	@psql -d scavengedb

tests: integration unit

start: 
	@echo "\nStarting the server...\n"
	@go build && ./scavenge serve --dev-mode

build:
	go build -a
	docker build -t scavenge .
