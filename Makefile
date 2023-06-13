# Protips:
# - RUN can be overridden for debugging, like:
#   $ RUN="dlv debug --" make -e

VERSION := $(shell git describe --tags --dirty --always 2> /dev/null || echo "dev")
LDFLAGS = "-X main.Version=$(VERSION) -w -s"
SOURCES = $(shell find . -type f -name '*.go')

BINARY = $(notdir $(PWD))
RUN = ./$(BINARY)

all: $(BINARY)

$(BINARY): $(SOURCES)
	GO111MODULE=on go build -ldflags $(LDFLAGS) -o "$@"

deps:
	GO111MODULE=on go get -d

build: $(BINARY)

clean:
	rm $(BINARY)

run: $(BINARY)
	$(RUN) --help

test:
	go test -vet "all" -timeout 5s -race ./...

.PHONY: dshackle-proto-gen
dshackle-proto-gen:
	protoc -I ./emerald-grpc/proto \
		--proto_path=emerald-grpc/proto \
		--go_out=pb \
		--go-grpc_out=pb \
		--go_opt=paths=source_relative \
		--go_opt=Mblockchain.proto=test/ags/pkg/dshackle \
		--go_opt=Mcommon.proto=test/ags/pkg/dshackle \
		--go-grpc_opt=paths=source_relative \
		--go-grpc_opt=Mblockchain.proto=github.com/p2p-org/drpc-provider-estimator/dshackle \
		--go-grpc_opt=Mcommon.proto=github.com/p2p-org/drpc-provider-estimator/dshackle \
		blockchain.proto common.proto
