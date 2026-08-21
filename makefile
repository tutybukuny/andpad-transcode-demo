tdapp:
	CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -o ./bin/tdapp ./cmd
local-dev-up: tdapp
	cd docker && \
	docker compose up -d --build db db_migration
local-down:
	cd docker && \
	docker compose down