#!/usr/bin/make
# Makefile readme (ru): <http://linux.yaroslavl.ru/docs/prog/gnu_make_3-79_russian_manual.html>
# Makefile readme (en): <https://www.gnu.org/software/make/manual/html_node/index.html#SEC_Contents>

SHELL = /bin/sh

test_coverage:
	docker-compose -f tests/env/docker-compose.yaml up -d --remove-orphans
	rm -rf coverage-ci
	mkdir ./coverage-ci
#	go test -v -race -cover -tags=debug -coverpkg=./... -failfast -coverprofile=./coverage-ci/ws_origin.out -covermode=atomic ./http/middleware/websockets
#	go test -v -race -cover -tags=debug -coverpkg=./... -failfast -coverprofile=./coverage-ci/new_relic_mdw.out -covermode=atomic ./http/middleware/new_relic
#	go test -v -race -cover -tags=debug -coverpkg=./... -failfast -coverprofile=./coverage-ci/http_config.out -covermode=atomic ./http/config
#	go test -v -race -cover -tags=debug -coverpkg=./... -failfast -coverprofile=./coverage-ci/server_cmd.out -covermode=atomic ./server
#	go test -v -race -cover -tags=debug -coverpkg=./... -failfast -coverprofile=./coverage-ci/struct_jobs.out -covermode=atomic ./jobs
	#go test -v -race -cover -tags=debug -coverpkg=./... -failfast -coverprofile=./coverage-ci/directives.out -covermode=atomic ./http/middleware/cache/directives
	go test -v -race -cover -tags=debug -coverpkg=./... -failfast -coverprofile=./coverage-ci/service.out -covermode=atomic ./plugins/service
	go test -timeout 20m -v -race -cover -tags=debug -failfast -coverpkg=./... -coverprofile=./coverage-ci/jobs_core.out -covermode=atomic ./plugins/jobs
	go test -v -race -cover -tags=debug -coverpkg=./... -failfast -coverprofile=./coverage-ci/kv_plugin.out -covermode=atomic ./plugins/kv
	go test -v -race -cover -tags=debug -coverpkg=./... -failfast -coverprofile=./coverage-ci/tcp_plugin.out -covermode=atomic ./plugins/tcp
	go test -v -race -cover -tags=debug -coverpkg=./... -failfast -coverprofile=./coverage-ci/reload.out -covermode=atomic ./plugins/reload
	go test -v -race -cover -tags=debug -coverpkg=./... -failfast -coverprofile=./coverage-ci/grpc_codec.out -covermode=atomic ./grpc/codec
	go test -v -race -cover -tags=debug -coverpkg=./... -failfast -coverprofile=./coverage-ci/grpc_parser.out -covermode=atomic ./grpc/parser
	go test -v -race -cover -tags=debug -coverpkg=./... -failfast -coverprofile=./coverage-ci/broadcast_plugin.out -covermode=atomic ./plugins/broadcast
	go test -v -race -cover -tags=debug -coverpkg=./... -failfast -coverprofile=./coverage-ci/websockets.out -covermode=atomic ./plugins/websockets
	go test -v -race -cover -tags=debug -coverpkg=./... -failfast -coverprofile=./coverage-ci/http.out -covermode=atomic ./plugins/http
	go test -v -race -cover -tags=debug -coverpkg=./... -failfast -coverprofile=./coverage-ci/grpc_plugin.out -covermode=atomic ./plugins/grpc
	go test -v -race -cover -tags=debug -coverpkg=./... -failfast -coverprofile=./coverage-ci/informer.out -covermode=atomic ./plugins/informer
	go test -v -race -cover -tags=debug -coverpkg=./... -failfast -coverprofile=./coverage-ci/server.out -covermode=atomic ./plugins/server
	go test -v -race -cover -tags=debug -coverpkg=./... -failfast -coverprofile=./coverage-ci/status.out -covermode=atomic ./plugins/status
	go test -v -race -cover -tags=debug -coverpkg=./... -failfast -coverprofile=./coverage-ci/config.out -covermode=atomic ./plugins/config
	go test -v -race -cover -tags=debug -coverpkg=./... -failfast -coverprofile=./coverage-ci/gzip.out -covermode=atomic ./plugins/gzip
	go test -v -race -cover -tags=debug -coverpkg=./... -failfast -coverprofile=./coverage-ci/headers.out -covermode=atomic ./plugins/headers
	go test -v -race -cover -tags=debug -coverpkg=./... -failfast -coverprofile=./coverage-ci/logger.out -covermode=atomic ./plugins/logger
	go test -v -race -cover -tags=debug -coverpkg=./... -failfast -coverprofile=./coverage-ci/metrics.out -covermode=atomic ./plugins/metrics
	go test -v -race -cover -tags=debug -coverpkg=./... -failfast -coverprofile=./coverage-ci/resetter.out -covermode=atomic ./plugins/resetter
	go test -v -race -cover -tags=debug -coverpkg=./... -failfast -coverprofile=./coverage-ci/rpc.out -covermode=atomic ./plugins/rpc
	echo 'mode: atomic' > ./coverage-ci/summary.txt
	tail -q -n +2 ./coverage-ci/*.out >> ./coverage-ci/summary.txt
	docker-compose -f tests/env/docker-compose.yaml down

test: ## Run application tests
	docker compose -f tests/env/docker-compose.yaml up -d --remove-orphans
	sleep 10
#	go test -v -race -tags=debug ./jobs
#	go test -v -race -tags=debug ./http/config
#	go test -v -race -tags=debug ./http/middleware/new_relic
#	go test -v -race -tags=debug ./http/middleware/websockets
#	go test -v -race -tags=debug ./server
#	go test -v -race -tags=debug ./grpc/codec
#	go test -v -race -tags=debug ./grpc/parser
#	go test -v -race -tags=debug ./http/middleware/cache/directives
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
	docker compose -f tests/env/docker-compose.yaml down

generate-proto:
	protoc -I./api/proto/jobs/v1beta --go_out=./api/proto/jobs/v1beta jobs.proto
	protoc -I./api/proto/kv/v1beta --go_out=./api/proto/kv/v1beta kv.proto
	protoc -I./api/proto/websockets/v1beta --go_out=./api/proto/websockets/v1beta websockets.proto
	protoc -I./api/proto/cache/v1beta --go_out=./api/proto/cache/v1beta response.proto
