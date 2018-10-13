BUILD_OS_TARGETS = "linux darwin freebsd windows netbsd"

test: lint testgo

testgo: testdeps
	go test ./...

deps:
	go get -d -v ./...

testdeps:
	go get -d -v -t ./...
	go get golang.org/x/lint/golint \
		golang.org/x/tools/cmd/cover \
		github.com/pierrre/gotestcover \
		github.com/mattn/goveralls

lint: testdeps
	go tool vet -all .
	_tools/go-linter $(BUILD_OS_TARGETS)

cover: testdeps
	gotestcover -v -covermode=count -coverprofile=.profile.cov -parallelpackages=4 ./...

.PHONY: test testgo deps testdeps lint cover
