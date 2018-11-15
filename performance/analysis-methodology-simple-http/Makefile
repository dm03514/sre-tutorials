LOAD_TEST_RATE=50
LOAD_TEST_TARGET=http://localhost:8080
PKGS = $(shell go list ./... | grep -v /vendor/)

fmt:
	go fmt $(PKGS)

stack:
	docker-compose down && docker-compose up

load-test:
	echo "POST $(LOAD_TEST_TARGET)" | vegeta attack -body tests/fixtures/age_no_match.json -rate=$(LOAD_TEST_RATE) -duration=0 | tee results.bin | vegeta report

ping-server:
	curl \
		-X POST \
		-H "Content-Type: application/json" \
		-d @tests/fixtures/age_no_match.json \
		http://localhost:8080 -v

.PHONY: stack load-test fmt test-unit
