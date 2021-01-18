.PHONY: test
test:
	go test ./...

.PHONY: cover
cover:
	go test -race -covermode atomic -coverprofile=.profile.cov ./...
