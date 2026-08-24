tdapp:
	CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -o ./bin/tdapp ./cmd && \
	docker build . -f ./docker/tdapp_local.DockerFile -t tdapp:local
local-dev-up: tdapp
	cd docker && \
	docker compose up -d db db_migration localstack
local-up: tdapp
	cd docker && \
	docker compose up -d
local-down:
	cd docker && \
	docker compose down
swagger:
	swag init -g cmd/main.go --parseInternal --parseDependency