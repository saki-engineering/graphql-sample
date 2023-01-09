run:
	go run server.go

test:
	go test

setup:
	./setup.sh

generate:
	gqlgen generate
	sqlboiler sqlite3
	go generate ./...