test: install-static-check
	go test -failfast -count=1 -race -v ./...
	staticcheck -checks=all ./...

sleep-one-sec:
	sleep 1;

install-static-check:
	go install honnef.co/go/tools/cmd/staticcheck@latest

pq-local:
	docker compose -f compose.yaml up -d postgres

migrate-local:
	docker compose -f compose.yaml up migrate

mongo-local:
	docker compose -f compose.yaml up -d mongo

change-db:
	docker compose -f compose.yaml run --no-deps --rm migrate create --dir=migrations --ext=sql --seq ${name}

migrate-down:
	docker compose -f compose.yaml run --rm migrate down ${step}

tear:
	docker compose -f compose.yaml down

setup: pq-local sleep-one-sec migrate-local mongo-local

tidy:
	go mod tidy
	go mod vendor

update-dependencies:
	go get -u ./...
	go mod tidy
	go mod vendor
