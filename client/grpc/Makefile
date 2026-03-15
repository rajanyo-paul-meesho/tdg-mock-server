# Common Makefile to generate shared, PDP v2, and For-You v2 protos

.PHONY: all generate install-deps clean check-protoc

all: generate

install-deps:
	@echo "Installing protobuf Go plugins..."
	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
	@echo "Dependencies installed successfully!"

check-protoc:
	@which protoc > /dev/null || (echo "Error: protoc is not installed. Please install Protocol Buffers compiler." && exit 1)
	@echo "protoc found: $$(which protoc)"

generate: check-protoc
	@echo "Checking if protoc-gen-go is available..."
	@export PATH="$$PATH:$$(go env GOPATH)/bin" && \
	which protoc-gen-go > /dev/null || (echo "Installing dependencies..." && $(MAKE) install-deps)
	@echo "Generating shared + PDP v2 + For-You v2 protos..."
	@echo "Note: Proto files use relative paths in go_package, converting to full paths in generated imports"
	@export PATH="$$PATH:$$(go env GOPATH)/bin" && \
	cd ../proto/feedaggregator && \
	protoc \
		-I . \
		--go_out=paths=source_relative:../../grpc \
		--go-grpc_out=paths=source_relative:../../grpc \
		feedcommonresponse/*.proto \
		pdp/v2/*.proto \
		foryou/v2/*.proto && \
	cd ../../grpc && \
	echo "Replacing relative import paths with full module paths..." && \
	find . -name "*.pb.go" -type f -exec sed -i '' 's|"\.\./feedcommonresponse"|"github.com/Meesho/feed-aggregator-go/client/grpc/feedcommonresponse"|g' {} \;

clean:
	@echo "Cleaning generated files..."
	@find . -name "*.pb.go" -delete
	@echo "Generated files cleaned!"


