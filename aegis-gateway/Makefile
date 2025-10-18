.PHONY: build run docker-build docker-run deploy-up test

build:
	go build -o bin/aegis ./cmd/aegis

run: build
	./bin/aegis

docker-build:
	docker build -t aegis:local .

docker-run: docker-build
	docker run --rm -p 8080:8080 aegis:local

deploy-up:
	cd deploy && docker-compose up --build

test:
	go test ./internal/policy
