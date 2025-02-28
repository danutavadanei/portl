build:
	go build -o bin/main main.go

run:
	go run cmd/main.go --debug

docker-build:
	docker build -t ghcr.io/danutavadanei/portl .

docker-run:
	docker run -p 8080:8080 -p 2222:2222 -v ./keys:/keys ghcr.io/danutavadanei/portl

docker-push:
	docker push ghcr.io/danutavadanei/portl:latest