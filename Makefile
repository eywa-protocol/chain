.PHONY: wrappers test

all: wrappers test

test:
	go test -failfast -v ./...

test_block:
	go test -v ./core/store/ledgerstore -run TestBlock

test_tx:
	go test -v ./core/store/ledgerstore -run TestSaveTransaction

test_verify:
	go test -v ./core/test/ -run TestVerifyTx

wrappers:
	cd gen && go generate
