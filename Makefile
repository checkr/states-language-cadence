GO=go
GO_BUILD_OPTS?=

# Test options. -count 1 disables test result caching.
GO_TEST_OPTS?=-v --race -count 1

BINDIR?=bin
GENDIR?=gen

.PHONY: all
all: clean build lint test

.PHONY: clean
clean:
	rm -fR "$(BINDIR)/*"
	rm -fR "$(GENDIR)/*"

# a rule to force phony pattern rules to build always.
.PHONY: force
force:

.PHONY: build
build: build-all bin

.PHONY: build-all
build-all:
	$(GO) build -mod vendor -o $(BINDIR) $(GO_BUILD_OPTS) ./...

.PHONY: $(BINDIR)
$(BINDIR): $(shell find cmd/* -type d | sed -e 's/^cmd/$(BINDIR)/')

# Build a specific binary. Binaries are generated from ./cmd/ subdirs.
$(BINDIR)/%: force
	$(GO) build -mod vendor -o $(BINDIR)/$* $(GO_BUILD_OPTS) ./cmd/$*

.PHONY: test
test: test-unit test-integration

.PHONY: test-unit
test-unit:
	$(GO) test -mod vendor -tags unit -coverprofile="$(GENDIR)/unit.cov" $(GO_TEST_OPTS) ./...

.PHONY: test-integration
test-integration: check-pg # remove check-pg if pq is not required.
	$(GO) test -mod vendor -tags integration -coverprofile="$(GENDIR)/int.cov" $(GO_TEST_OPTS) ./...

# check for pg connection only if has pg_isready utility.
# in ci probably doesn't have this installed.
.PHONY: check-pg
check-pg:
	@(! which pg_isready || pg_isready -h localhost -p 5432) 2>&1 >/dev/null || \
		(echo 'postgres not ready! run "docker-compose start" first.' && exit 1)

# update all dependencies.
.PHONY: mod-update
mod-update:
	go get -u all
	make mod-vendor

.PHONY: mod-vendor
mod-vendor: mod-tidy
	rm -fR vendor
	$(GO) mod vendor
	make mod-vendor-pack

.PHONY: mod-vendor-pack
mod-vendor-pack:
	tar czf vendor.tar.gz vendor

.PHONY: mod-vendor-unpack
mod-vendor-unpack:
	rm -fR vendor
	tar xf vendor.tar.gz

.PHONY: mod-tidy
mod-tidy:
	$(GO) mod tidy

.PHONY: lint
lint:
	# see https://github.com/golangci/golangci-lint
	golangci-lint run ./...

.PHONY: docker
docker:
	docker build -t states-language-cadence -f Dockerfile .

CADENCE_HOST := host.docker.internal:7933

.PHONY: cadence-create-domain
cadence-create-domain:
	docker run \
	--rm ubercadence/cli:master \
	--address ${CADENCE_HOST} \
	--domain samples-domain domain register \
	--global_domain false

.PHONY: run-trigger
run-trigger:
	docker run \
	--rm ubercadence/cli:master \
	--address ${CADENCE_HOST} \
	--domain samples-domain workflow run \
	--tl WorkflowDemo \
	--wt example:workflow:ExampleWorkflow \
	--et 60 \
	-i '{"example":"example"}'
