PROTO_DIR= ./
OUT_DIR=../grpc/go
GOOGLEAPIS_DIR?=../third_party/googleapis

.PHONY: lint

lint:
	@echo "Running clang-format on protobufs"
	clang-format --dry-run --Werror *.proto types/*.proto

# generate-protos
generate-protos:
	@for file in $(PROTO_DIR)/*.proto; do \
		base_name=$$(basename "$$file" .proto); \
		out_dir="$(OUT_DIR)/$${base_name}grpc"; \
		mkdir -p "$$out_dir"; \
		protoc \
			--proto_path=$(PROTO_DIR) \
			--proto_path=$(GOOGLEAPIS_DIR) \
			--go_out="$$out_dir" \
			--go_opt=paths=source_relative \
			--go-grpc_out="$$out_dir" \
			--go-grpc_opt=paths=source_relative \
			$$file; \
	done