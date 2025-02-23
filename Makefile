.PHONY: run

dc:
	docker-compose up -d  --remove-orphans --build

run:
	go build -race -o app cmd/main.go && \
	HTTP_ADDR=:8080 \
	DEBUG_ERRORS=1 \
	DSN="postgres://postgres:@127.0.0.1:5432/postgres?sslmode=disable" \
	MIGRATIONS_PATH="file://./internal/app/migrations" \
	RABBITMQ_URL="amqp://guest:guest@127.0.0.1:5672/" \
	API_KEY="hgOzFGgeBgBsmwbxc6fEjPiVcq2QfOV5i2oIl0sK" \
	API_URL="https://api.nasa.gov/mars-photos/api/v1/rovers/curiosity/photos" \
	./app

generate:
	go generate ./...