# Makefile

## constant values
MOCK_SERVER_PORT=9001
SERVER_PORT=8080

## start the application
start:
	echo "Starting the application on $(SERVER_PORT)"
	HEALTH_URL_PATTERN="http://{env_dnd}{BASE_URL}/{service}{version}/health" LOGS_URL_PATTERN="{LOGS_BASE_URL}/{CLUSTER_NAME}/{env}/{service}{version}/logs?project={PROJECT_NAME}" LOGS_BASE_URL="localhost:9001/base/logsbase" CLUSTER_NAME=cluster-name PROJECT_NAME=proj-name SERVICE_NAMES=rest-service,rest-service-2,rest-service-3,rest-service-4,rest-service-5,rest-service-6,rest-service-7,rest-service-8,rest-service-9,rest-service-10,rest-service-11,rest-service-12 VERSIONS=,-v1,-v2 BASE_URL=localhost:9001/base ENVIRONMENTS=dev,staging,prod  REST_PORT=$(SERVER_PORT) go run .

## start mock server for testing
mock:
	echo "Starting mock server on $(MOCK_SERVER_PORT)"
	VERSIONS=,-v1 SERVICE_NAMES=rest-service,rest-service-2,rest-service-3,rest-service-4,rest-service-5,rest-service-6,rest-service-7,rest-service-8,rest-service-9,rest-service-10,rest-service-11,rest-service-12 REST_PORT=$(MOCK_SERVER_PORT) go run cmd/mock/mock-server.go

## stop services and remove environment variables
clean:
	@echo "Cleaning up"
	@unset HEALTH_URL_PATTERN
	@unset LOGS_URL_PATTERN
	@unset LOGS_BASE_URL
	@unset CLUSTER_NAME
	@unset PROJECT_NAME
	@unset SERVICE_NAMES
	@unset VERSIONS
	@unset BASE_URL
	@unset ENVIRONMENTS