build:
	docker compose --progress plain --profile app build
up:
	docker compose --profile app --profile infra up -d
down:
	docker compose --profile app down -v
start:
	docker compose --profile app --profile infra start
stop:
	docker compose --profile app stop
logs:
	docker compose --profile app logs -f rate-limit-webserver
infra-start:
	docker compose --profile infra start
infra-stop:
	docker compose --profile infra stop
infra-up:
	docker compose --profile infra up -d
infra-down:
	docker compose --profile infra down -v
test-inmemory:
	go test -v -failfast -run ^TestLimitUseCaseTestSuite$$ ./internal/usecase
test-redis:
	go test -v -failfast -run ^TestLimitUseCaseRedisTestSuite$$ ./internal/usecase