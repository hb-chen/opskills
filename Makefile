.PHONY: proto build run test clean env

# Set local Go 1.25.4
export GOROOT := /usr/local/go1.25.4
export PATH := /usr/local/go1.25.4/bin:$(PATH)

# Generate proto files
proto:
	@echo "Generating proto files..."
	@mkdir -p proto/common proto/ops
	@protoc --go_out=. --go_opt=paths=source_relative \
		--go-grpc_out=. --go-grpc_opt=paths=source_relative \
		--grpc-gateway_out=. --grpc-gateway_opt=paths=source_relative \
		--openapiv2_out=. \
		-I. -Ithird_party \
		proto/common/common.proto \
		proto/ops/ops.proto

# Build the application
build:
	@echo "Info: Building opskills-agent..."
	@go build -o dist/local/opskills-agent ./cmd/server
	@mkdir -p dist/local/configs
	@if [ -f configs/config.yaml ]; then \
		if [ ! -f dist/local/configs/config.yaml ]; then \
			cp configs/config.yaml dist/local/configs/config.yaml; \
			echo "Info: Copied configs/config.yaml to dist/local/configs/"; \
		else \
			echo "Info: dist/local/configs/config.yaml already exists, skipping copy to preserve your configuration"; \
		fi \
	else \
		echo "Warning: configs/config.yaml not found, skipping copy"; \
	fi

# Run the application
run:
	@cd dist/local && ./opskills-agent serve -e --config ./configs/config.yaml

# Run tests
test:
	@go test ./...

# Clean build artifacts
clean:
	@rm -rf dist/local/
