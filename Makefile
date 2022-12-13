.PHONY: swagger
swagger:
	@echo "[SWAGGER] Generating swagger documentation..."
	swag init --parseInternal --parseDepth 1 -g server/rest/v1/handler/router.go -o ./docs/

.PHONY: mock
mock:
	@echo "[SERVICE] mocking task-service ..."
	mockgen -destination ./service/taskservice/mock/mock.go -package mock_taskservice -source ./service/taskservice/interface.go
	@echo "\n[SERVICE] mocking http-client ..."
	mockgen -destination ./service/httpclient/mock/mock.go -package mock_httpclient -source ./service/httpclient/httpclient.go
	@echo "\n[PROVIDER] mocking message broker ..."
	mockgen -destination ./provider/msgbroker/mock/mock.go -package mock_msgbroker -source ./provider/msgbroker/interface.go
	@echo "\n[PROVIDER] mocking task provider (database queries for tasks) ..."
	mockgen -destination ./provider/taskprovider/mock/mock.go -package mock_taskprovider -source ./provider/taskprovider/querier.go
	@echo "\n[PROVIDER] mocking rate limiter (redis queries) ..."
	mockgen -destination ./provider/ratelimiter/mock/mock.go -package mock_ratelimiter -source ./provider/ratelimiter/interface.go

.PHONY: run_interation_tests
run_interation_tests:
	@echo "[IN-MEMORY] running integration tests for in-memory database (Redis)..."
	go test tests/integration/ratelimiter/rate_test.go
	@echo "\n[DATABASE] running integration tests for taskprovider (PostgreSQL)..."
	go test tests/integration/taskprovider/tasks_test.go

.PHONY: run_mock_tests
run_mock_tests:
	@echo "[MOCK][HANDLER] running tests for parsing http request..."
	go test server/rest/v1/handler/*.go
	@echo "\n[MOCK][BACKGROUND] running tests for background worker..."
	go test server/backgroundworker/*.go
	@echo "\n[MOCK][TASKSERVICE] running tests for task service..."
	go test service/taskservice/*.go

.PHONY: all
all:
	@echo "[DOCKER] preparing services..."
	docker-compose -f docker-services.yml up --bu -d
	@echo "[DOCKER] preparing api..."
	docker-compose up --bu -d

.PHONY: api_up
api_up:
	@echo "[DOCKER] preparing backend..."
	docker-compose up --bu -d

.PHONY: services_up
services_up:
	@echo "[DOCKER] preparing services..."
	docker-compose -f docker-services.yml up --bu -d
