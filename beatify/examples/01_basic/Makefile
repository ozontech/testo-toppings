MAKEFLAGS += --always-make

test:
	go test . -v -tags example -count=1

coverage:
	go test . -tags example -count=1 -beatify.coverage-only
