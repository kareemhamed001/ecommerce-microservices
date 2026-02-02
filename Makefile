proto:
	protoc --go_out=. --go-grpc_out=. --proto_path=. shared/proto/v1/*.proto

up:
	docker compose  up --build

down:
	docker compose  down