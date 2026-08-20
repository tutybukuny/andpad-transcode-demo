tdapp:
	go build -o ./bin/tdapp ./cmd
local-up: tdapp
	cd docker && \
	docker compose up -d --build
local-down:
	cd docker && \
	docker compose down