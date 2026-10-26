.PHONY: help proto generate test clean

help:
	@echo "Available make targets for share package:"
	@echo "  make proto     - Compile shared protobuf definitions"
	@echo "  make generate  - Run code generators"
	@echo "  make test      - Run unit tests"
	@echo "  make clean     - Clean generated protobuf stubs"

proto:
	@mkdir -p proto/pb
	@protoc --go_out=proto/pb --go_opt=paths=source_relative --go-grpc_out=proto/pb --go-grpc_opt=paths=source_relative -I=proto proto/user.proto
	@echo "Protobuf generated in apps/common/proto/pb."

generate: proto

test:
	@go test -v -race ./...

clean:
	@rm -rf proto/pb/*.pb.go
	@echo "Cleaned share proto stubs."
