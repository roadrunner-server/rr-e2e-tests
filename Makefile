#!/usr/bin/make
# Makefile readme (ru): <http://linux.yaroslavl.ru/docs/prog/gnu_make_3-79_russian_manual.html>
# Makefile readme (en): <https://www.gnu.org/software/make/manual/html_node/index.html#SEC_Contents>

SHELL = /bin/sh

temporal_docker_up:
	docker-compose -f env/docker-compose-temporal.yaml up -d --remove-orphans

jobs_docker_up:
	docker-compose -f env/docker-compose-jobs.yaml up -d --remove-orphans

remove_prev_ci:
	rm -rf coverage-ci
	mkdir ./coverage-ci

temporal_test: temporal_docker_up remove_prev_ci
	sleep 30
	go test -v -race -cover -tags=debug -coverpkg=all -failfast -coverprofile=./coverage-ci/temporal.out -covermode=atomic ./plugins/temporal

service_test: remove_prev_ci
	go test -v -race -cover -tags=debug -coverpkg=all -failfast -coverprofile=./coverage-ci/service.out -covermode=atomic ./plugins/service

jobs_test: jobs_docker_up remove_prev_ci
	sleep 30
	go test -timeout 20m -v -race -cover -tags=debug -failfast -coverpkg=all -coverprofile=./coverage-ci/jobs_core.out -covermode=atomic ./plugins/jobs

test_coverage:
	docker-compose -f env/docker-compose.yaml up -d --remove-orphans
	rm -rf coverage-ci
	mkdir ./coverage-ci
	sleep 30
	go test -v -race -cover -tags=debug -coverpkg=all -failfast -coverprofile=./coverage-ci/kv_plugin.out -covermode=atomic ./plugins/kv
	go test -v -race -cover -tags=debug -coverpkg=all -failfast -coverprofile=./coverage-ci/tcp_plugin.out -covermode=atomic ./plugins/tcp
	go test -v -race -cover -tags=debug -coverpkg=all -failfast -coverprofile=./coverage-ci/reload.out -covermode=atomic ./plugins/reload
	go test -v -race -cover -tags=debug -coverpkg=all -failfast -coverprofile=./coverage-ci/broadcast_plugin.out -covermode=atomic ./plugins/broadcast
	go test -v -race -cover -tags=debug -coverpkg=all -failfast -coverprofile=./coverage-ci/websockets.out -covermode=atomic ./plugins/websockets
	go test -v -race -cover -tags=debug -coverpkg=all -failfast -coverprofile=./coverage-ci/http.out -covermode=atomic ./plugins/http
	go test -v -race -cover -tags=debug -coverpkg=all -failfast -coverprofile=./coverage-ci/grpc_plugin.out -covermode=atomic ./plugins/grpc
	go test -v -race -cover -tags=debug -coverpkg=all -failfast -coverprofile=./coverage-ci/informer.out -covermode=atomic ./plugins/informer
	go test -v -race -cover -tags=debug -coverpkg=all -failfast -coverprofile=./coverage-ci/server.out -covermode=atomic ./plugins/server
	go test -v -race -cover -tags=debug -coverpkg=all -failfast -coverprofile=./coverage-ci/status.out -covermode=atomic ./plugins/status
	go test -v -race -cover -tags=debug -coverpkg=all -failfast -coverprofile=./coverage-ci/config.out -covermode=atomic ./plugins/config
	go test -v -race -cover -tags=debug -coverpkg=all -failfast -coverprofile=./coverage-ci/gzip.out -covermode=atomic ./plugins/gzip
	go test -v -race -cover -tags=debug -coverpkg=all -failfast -coverprofile=./coverage-ci/headers.out -covermode=atomic ./plugins/headers
	go test -v -race -cover -tags=debug -coverpkg=all -failfast -coverprofile=./coverage-ci/logger.out -covermode=atomic ./plugins/logger
	go test -v -race -cover -tags=debug -coverpkg=all -failfast -coverprofile=./coverage-ci/metrics.out -covermode=atomic ./plugins/metrics
	go test -v -race -cover -tags=debug -coverpkg=all -failfast -coverprofile=./coverage-ci/resetter.out -covermode=atomic ./plugins/resetter
	go test -v -race -cover -tags=debug -coverpkg=all -failfast -coverprofile=./coverage-ci/rpc.out -covermode=atomic ./plugins/rpc
	echo 'mode: atomic' > ./coverage-ci/summary.txt
	tail -q -n +2 ./coverage-ci/*.out >> ./coverage-ci/summary.txt
	sed -i '2,$${/roadrunner/!d}' ./coverage-ci/summary.txt
	docker-compose -f env/docker-compose.yaml kill
	docker-compose -f env/docker-compose.yaml down

test: ## Run application tests
	docker compose -f env/docker-compose.yaml up -d --remove-orphans
	sleep 10
	go test -timeout 20m -v -race -tags=debug ./plugins/jobs
	go test -v -race -tags=debug ./plugins/kv
	go test -v -race -tags=debug ./plugins/tcp
	go test -v -race -tags=debug ./plugins/broadcast
	go test -v -race -tags=debug ./plugins/websockets
	go test -v -race -tags=debug ./plugins/http
	go test -v -race -tags=debug ./plugins/informer
	go test -v -race -tags=debug ./plugins/reload
	go test -v -race -tags=debug ./plugins/websockets
	go test -v -race -tags=debug ./plugins/grpc
	go test -v -race -tags=debug ./plugins/server
	go test -v -race -tags=debug ./plugins/service
	go test -v -race -tags=debug ./plugins/status
	go test -v -race -tags=debug ./plugins/config
	go test -v -race -tags=debug ./plugins/gzip
	go test -v -race -tags=debug ./plugins/headers
	go test -v -race -tags=debug ./plugins/logger
	go test -v -race -tags=debug ./plugins/metrics
	go test -v -race -tags=debug ./plugins/resetter
	go test -v -race -tags=debug ./plugins/rpc
	docker compose -f env/docker-compose.yaml kill
	docker compose -f env/docker-compose.yaml down