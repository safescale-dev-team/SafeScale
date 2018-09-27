.DEFAULT_GOAL := help

.PHONY: default
default: help ;

ifndef VERBOSE
MAKEFLAGS += --no-print-directory
endif

GO?=go
GOBIN?=~/go/bin
CP?=cp


# Binaries generated
EXECS=broker/cli/broker/broker broker/cli/brokerd/brokerd deploy/cli/deploy perform/perform

# List of packages
PKG_LIST := $(shell $(GO) list ./... | grep -v /vendor/)
# List of packages to test (nor deploy neither providers are ready for prime time :( )
TESTABLE_PKG_LIST := $(shell $(GO) list ./... | grep -v /vendor/ | grep -v /deploy | grep -v /providers/aws )


# DEPENDENCIES MANAGEMENT
STRINGER := golang.org/x/tools/cmd/stringer
RICE := github.com/GeertJohan/go.rice github.com/GeertJohan/go.rice/rice
PROTOC := github.com/golang/protobuf
PROTOBUF := github.com/golang/protobuf/protoc-gen-go

# Build tools
COVER := golang.org/x/tools/cmd/cover
DEP := github.com/golang/dep/cmd/dep

DEVDEPSLIST := $(STRINGER) $(RICE) $(PROTOBUF) $(DEP) $(COVER)


# Life is better with colors
COM_COLOR   = \033[0;34m
OBJ_COLOR   = \033[0;36m
OK_COLOR    = \033[0;32m
GOLD_COLOR  = \033[0;93m
ERROR_COLOR = \033[0;31m
WARN_COLOR  = \033[0;33m
NO_COLOR    = \033[m

OK_STRING    = "[OK]"
INFO_STRING  = "[INFO]"
ERROR_STRING = "[ERROR]"
WARN_STRING  = "[WARNING]"

all: begin ground getdevdeps ensure generate providers broker system deploy perform utils
	@printf "%b" "$(OK_COLOR)$(OK_STRING) Build SUCCESSFUL $(NO_COLOR)\n";

begin:
	@printf "%b" "$(OK_COLOR)$(INFO_STRING) Build begins...$(NO_COLOR)\n";

ground:
	@printf "%b" "$(OK_COLOR)$(INFO_STRING) Testing tool prerequisites, $(NO_COLOR)target $(OBJ_COLOR)$(@)$(NO_COLOR)\n";
	@command -v git >/dev/null 2>&1 || { printf "%b" "$(ERROR_COLOR)$(ERROR_STRING) git is required but it's not installed.  Aborting.$(NO_COLOR)\n" >&2; exit 1; }
	@command -v $(GO) >/dev/null 2>&1 || { printf "%b" "$(ERROR_COLOR)$(ERROR_STRING) go is required but it's not installed.  Aborting.$(NO_COLOR)\n" >&2; exit 1; }
	@command -v protoc >/dev/null 2>&1 || { printf "%b" "$(ERROR_COLOR)$(ERROR_STRING) protoc is required but it's not installed.  Aborting.$(NO_COLOR)\n" >&2; exit 1; }

getdevdeps:
	@printf "%b" "$(OK_COLOR)$(INFO_STRING) Testing prerequisites, $(NO_COLOR)target $(OBJ_COLOR)$(@)$(NO_COLOR)\n";
	@which dep rice stringer protoc-gen-go cover > /dev/null; if [ $$? -ne 0 ]; then \
    	  $(GO) get -u $(STRINGER) $(RICE) $(PROTOBUF) $(COVER) $(DEP); \
    fi

ensure:
	@printf "%b" "$(OK_COLOR)$(INFO_STRING) Checking versions, $(NO_COLOR)target $(OBJ_COLOR)$(@)$(NO_COLOR)\n";
	@($(GOBIN)/dep ensure)

utils:
	@printf "%b" "$(OK_COLOR)$(INFO_STRING) Building utils, $(NO_COLOR)target $(OBJ_COLOR)$(@)$(NO_COLOR)\n";
	@(cd utils && $(MAKE) all)

providers:
	@printf "%b" "$(OK_COLOR)$(INFO_STRING) Building providers, $(NO_COLOR)target $(OBJ_COLOR)$(@)$(NO_COLOR)\n";
	@(cd providers && $(MAKE) all)

system:
	@printf "%b" "$(OK_COLOR)$(INFO_STRING) Building system, $(NO_COLOR)target $(OBJ_COLOR)$(@)$(NO_COLOR)\n";
	@(cd system && $(MAKE) all)

broker: utils system providers
	@printf "%b" "$(OK_COLOR)$(INFO_STRING) Building service broker, $(NO_COLOR)target $(OBJ_COLOR)$(@)$(NO_COLOR)\n";
	@(cd broker && $(MAKE) all)

deploy: utils system providers broker
	@printf "%b" "$(OK_COLOR)$(INFO_STRING) Building service deploy, $(NO_COLOR)target $(OBJ_COLOR)$(@)$(NO_COLOR)\n";
	@(cd deploy && $(MAKE) all)

perform: utils system providers broker
	@printf "%b" "$(OK_COLOR)$(INFO_STRING) Building service perform, $(NO_COLOR)target $(OBJ_COLOR)$(@)$(NO_COLOR)\n";
	@(cd perform && $(MAKE) all)# List of packages

clean:
	@printf "%b" "$(OK_COLOR)$(INFO_STRING) Cleaning..., $(NO_COLOR)target $(OBJ_COLOR)$(@)$(NO_COLOR)\n";
	@(cd providers && $(MAKE) $@)
	@(cd system && $(MAKE) $@)
	@(cd broker && $(MAKE) $@)
	@(cd deploy && $(MAKE) $@)
	@(cd perform && $(MAKE) $@)

broker/client/broker: broker
	@printf "%b" "$(OK_COLOR)$(INFO_STRING) Building service broker (client) , $(NO_COLOR)target $(OBJ_COLOR)$(@)$(NO_COLOR)\n";

broker/daemon/brokerd: broker
	@printf "%b" "$(OK_COLOR)$(INFO_STRING) Building service broker (daemon) , $(NO_COLOR)target $(OBJ_COLOR)$(@)$(NO_COLOR)\n";

deploy/cli/deploy: deploy
	@printf "%b" "$(OK_COLOR)$(INFO_STRING) Building service deploy, $(NO_COLOR)target $(OBJ_COLOR)$(@)$(NO_COLOR)\n";

perform/perform: perform
	@printf "%b" "$(OK_COLOR)$(INFO_STRING) Building service perform, $(NO_COLOR)target $(OBJ_COLOR)$(@)$(NO_COLOR)\n";

install: $(EXECS)
	@($(CP) -f $^ $(GOBIN))

docs:
	@printf "%b" "$(OK_COLOR)$(INFO_STRING) Running godocs in background, $(NO_COLOR)target $(OBJ_COLOR)$(@)$(NO_COLOR)\n";
	@(godoc -http=:6060 &)

devdeps:
	@printf "%b" "$(OK_COLOR)$(INFO_STRING) Getting dev dependencies, $(NO_COLOR)target $(OBJ_COLOR)$(@)$(NO_COLOR)\n";
	@($(GO) get -u $(DEVDEPSLIST))

depclean:
	@printf "%b" "$(OK_COLOR)$(INFO_STRING) Cleaning vendor and redownloading deps, $(NO_COLOR)target $(OBJ_COLOR)$(@)$(NO_COLOR)\n";
	@rm ./Gopkg.lock
	@rm -rf ./vendor
	@(dep ensure)

generate: # Run unit tests
	@printf "%b" "$(OK_COLOR)$(INFO_STRING) Running code generation, $(NO_COLOR)target $(OBJ_COLOR)$(@)$(NO_COLOR)\n";
	@($(GO) generate ./...) | tee generation_results.log

test: # Run unit tests
	@printf "%b" "$(OK_COLOR)$(INFO_STRING) Running unit tests, $(NO_COLOR)target $(OBJ_COLOR)$(@)$(NO_COLOR)\n";
	@$(GO) test -short ${TESTABLE_PKG_LIST} | tee test_results.log

vet:
	@printf "%b" "$(OK_COLOR)$(INFO_STRING) Running vet checks (with restrictions), $(NO_COLOR)target $(OBJ_COLOR)$(@)$(NO_COLOR)\n";
	@$(GO) vet ${TESTABLE_PKG_LIST} | tee vet_results.log

truevet:
	@printf "%b" "$(OK_COLOR)$(INFO_STRING) Running vet checks, $(NO_COLOR)target $(OBJ_COLOR)$(@)$(NO_COLOR)\n";
	@$(GO) vet ${PKG_LIST} | tee vet_results.log

coverage:
	@printf "%b" "$(OK_COLOR)$(INFO_STRING) Collecting coverage data, $(NO_COLOR)target $(OBJ_COLOR)$(@)$(NO_COLOR)\n";
	@printf "%b" "$(WARN_COLOR)$(WARN_STRING) Not ready, coming soon ;) , $(NO_COLOR)target $(OBJ_COLOR)$(@)$(NO_COLOR)\n";

help:
	@echo ''
	@printf "%b" "$(GOLD_COLOR) **************** SAFESCALE BUILD ****************$(NO_COLOR)\n";
	@echo ' If in doubt, try "make all"'
	@echo ''
	@printf "%b" "$(OK_COLOR)BUILD TARGETS:$(NO_COLOR)\n";
	@printf "%b" "  $(GOLD_COLOR)all          - Builds all binaries$(NO_COLOR)\n";
	@echo '  help         - Prints this help message'
	@echo '  docs         - Runs godoc in background at port 6060.'
	@echo '                 Go to (http://localhost:6060/pkg/github.com/CS-SI/)'
	@echo '  install      - Copies all binaries to $(GOBIN)'
	@echo ''
	@printf "%b" "$(OK_COLOR)TESTING TARGETS:$(NO_COLOR)\n";
	@echo '  vet          - Runs all checks (with restrictions)'
	@echo '  truevet      - Runs all checks'
	@echo '  test         - Runs all tests'
	@echo '  coverage     - Collects coverage info from unit tests'
	@echo ''
	@printf "%b" "$(OK_COLOR)DEV TARGETS:$(NO_COLOR)\n";
	@echo '  clean        - Removes files generated by build (obsolete, running again "make all" should overwrite everything)'
	@echo '  depclean     - Rebuilds vendor dependencies'
	@echo ''
	@echo
