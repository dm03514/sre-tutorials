LOAD_TEST_RATE=50
LOAD_TEST_TARGET=http://localhost:11111
WORKER_NUM_DECRYPTERS=4
PKGS = $(shell go list ./... | grep -v /vendor/)

fmt:
	go fmt $(PKGS)

test-unit:
	go test $(PKGS) -v

stack:
	WORKER_NUM_DECRYPTERS=$(WORKER_NUM_DECRYPTERS) \
	docker-compose down && docker-compose up

load-test:
	echo "POST $(LOAD_TEST_TARGET)/decrypt" | vegeta attack -body tests/fixtures/password_no_match.json -rate=$(LOAD_TEST_RATE) -duration=0 | tee results.bin | vegeta report

ping-server:
	curl \
		-X POST \
		-H "Content-Type: application/json" \
		-d @tests/fixtures/password_no_match.json \
		$(LOAD_TEST_TARGET)/decrypt

.PHONY: stack load-test fmt test-unit
