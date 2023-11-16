GO_LIB_FILES=log.go time.go
GO_BIN_FILES=cmd/calcmetric/calcmetric.go cmd/sync/sync.go
GO_BIN_CMDS=github.com/lukaszgryglicki/calcmetric hithub.com/lukaszgryglicki/sync
#for race CGO_ENABLED=1
GO_ENV=CGO_ENABLED=1
# GO_ENV=CGO_ENABLED=0
GO_BUILD=go build -ldflags '-s -w' -race
# GO_BUILD=go build -ldflags '-s -w'
GO_INSTALL=go install -ldflags '-s'
GO_FMT=gofmt -s -w
GO_LINT=golint -set_exit_status
GO_VET=go vet
GO_CONST=goconst
GO_IMPORTS=goimports -w
GO_USEDEXPORTS=usedexports
BINARIES=calcmetric sync
STRIP=strip

all: check ${BINARIES}

calcmetric: cmd/calcmetric/calcmetric.go ${GO_LIB_FILES}
	 ${GO_ENV} ${GO_BUILD} -o calcmetric cmd/calcmetric/calcmetric.go

sync: cmd/sync/sync.go ${GO_LIB_FILES}
	 ${GO_ENV} ${GO_BUILD} -o sync cmd/sync/sync.go

fmt: ${GO_BIN_FILES} ${GO_LIB_FILES}
	./for_each_go_file.sh "${GO_FMT}"

lint: ${GO_BIN_FILES} ${GO_LIB_FILES}
	./for_each_go_file.sh "${GO_LINT}"

vet: ${GO_BIN_FILES} ${GO_LIB_FILES}
	./vet_files.sh "${GO_VET}"

imports: ${GO_BIN_FILES} ${GO_LIB_FILES}
	./for_each_go_file.sh "${GO_IMPORTS}"

const: ${GO_BIN_FILES} ${GO_LIB_FILES}
	${GO_CONST} ./...

usedexports: ${GO_BIN_FILES} ${GO_LIB_FILES}
	${GO_USEDEXPORTS} ./...

check: fmt lint imports vet usedexports

install: check ${BINARIES}
	${GO_INSTALL} ${GO_BIN_CMDS}

strip: ${BINARIES}
	${STRIP} ${BINARIES}

clean:
	-rm -f ${BINARIES}

.PHONY: all
