BUILD_OS_TARGETS := "linux darwin freebsd windows netbsd"
export GO111MODULE := on

.PHONY: test
test: lint testgo

.PHONY: testgo
testgo: testdeps
	go test ./...

.PHONY: testdeps
testdeps:
	go get -d -v -t ./...
	GO111MODULE=off \
	go get golang.org/x/lint/golint \
		golang.org/x/tools/cmd/cover \
		github.com/pierrre/gotestcover \
		github.com/mattn/goveralls

.PHONY: lint
lint: testdeps
	go vet -all .
	_tools/go-linter $(BUILD_OS_TARGETS)

.PHONY: cover
cover: testdeps
	gotestcover -v -covermode=count -coverprofile=.profile.cov -parallelpackages=4 ./...
