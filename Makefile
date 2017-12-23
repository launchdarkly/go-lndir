test:
	go build ./cmd/go-lndir
	cd test; bats test.bats

.PHONY: test
