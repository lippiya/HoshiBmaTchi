PROTO_DIR = proto
PROTO_OUT_DIR = proto

USER_PROTO = $(PROTO_DIR)/user.proto
EMAIL_PROTO = $(PROTO_DIR)/email.proto
PHOTO_PROTO = $(PROTO_DIR)/photo.proto
POST_PROTO = $(PROTO_DIR)/post.proto
SOCIAL_PROTO = $(PROTO_DIR)/social.proto

ALL_PROTOS = $(USER_PROTO) $(EMAIL_PROTO) $(PHOTO_PROTO) $(POST_PROTO) $(SOCIAL_PROTO)

.PHONY: all proto clean install-tools help

all: proto

install-tools:
	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest


prepare_proto_dirs:
ifeq ($(OS),Windows_NT)
	@if not exist "$(PROTO_OUT_DIR)" mkdir "$(PROTO_OUT_DIR)"
else
	@mkdir -p $(PROTO_OUT_DIR)
endif

proto: prepare_proto_dirs
	protoc --go_out=$(PROTO_OUT_DIR) \
	   --go_opt=paths=source_relative \
	   --go-grpc_out=$(PROTO_OUT_DIR) \
	   --go-grpc_opt=paths=source_relative \
	   --proto_path=$(PROTO_DIR) \
	   $(ALL_PROTOS)


clean:
	@rm -f $(PROTO_OUT_DIR)/*.pb.go
