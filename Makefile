test:
	govendor test +local
	go build ./cmd/go-lndir
	cd test; bats test.bats

.PHONY: test
