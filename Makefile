#!/usr/bin/make
# Makefile readme (ru): <http://linux.yaroslavl.ru/docs/prog/gnu_make_3-79_russian_manual.html>
# Makefile readme (en): <https://www.gnu.org/software/make/manual/html_node/index.html#SEC_Contents>

SHELL = /bin/sh

run_docker:
	docker-compose -f env/docker-compose.yaml up -d --remove-orphans

sleep-%:
	sleep $(@:sleep-%=%)

test_coverage: run_docker sleep-30
	rm -rf coverage-ci
	mkdir ./coverage-ci
	go test -v -race -cover -tags=debug -coverpkg=all -failfast -coverprofile=./coverage-ci/temporal.out -covermode=atomic ./plugins/temporal
	go test -v -race -cover -tags=debug -coverpkg=all -failfast -coverprofile=./coverage-ci/service.out -covermode=atomic ./plugins/service
	go test -v -race -cover -tags=debug -coverpkg=all -failfast -coverprofile=./coverage-ci/amqp.out -covermode=atomic ./plugins/jobs/amqp
	go test -v -race -cover -tags=debug -coverpkg=all -failfast -coverprofile=./coverage-ci/amqp.out -covermode=atomic ./plugins/jobs/beanstalk
	go test -v -race -cover -tags=debug -coverpkg=all -failfast -coverprofile=./coverage-ci/amqp.out -covermode=atomic ./plugins/jobs/boltdb
	go test -v -race -cover -tags=debug -coverpkg=all -failfast -coverprofile=./coverage-ci/amqp.out -covermode=atomic ./plugins/jobs/durability
	go test -v -race -cover -tags=debug -coverpkg=all -failfast -coverprofile=./coverage-ci/amqp.out -covermode=atomic ./plugins/jobs/general
	go test -v -race -cover -tags=debug -coverpkg=all -failfast -coverprofile=./coverage-ci/amqp.out -covermode=atomic ./plugins/jobs/memory
	go test -v -race -cover -tags=debug -coverpkg=all -failfast -coverprofile=./coverage-ci/amqp.out -covermode=atomic ./plugins/jobs/nats
	go test -v -race -cover -tags=debug -coverpkg=all -failfast -coverprofile=./coverage-ci/amqp.out -covermode=atomic ./plugins/jobs/sqs
	go test -v -race -cover -tags=debug -coverpkg=all -failfast -coverprofile=./coverage-ci/kv_plugin.out -covermode=atomic ./plugins/kv
	go test -v -race -cover -tags=debug -coverpkg=all -failfast -coverprofile=./coverage-ci/tcp_plugin.out -covermode=atomic ./plugins/tcp
	go test -v -race -cover -tags=debug -coverpkg=all -failfast -coverprofile=./coverage-ci/proxy_ip.out -covermode=atomic ./plugins/proxy_ip_parser
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

test:
#	go test -v -race -cover -tags=debug ./plugins/temporal
#	go test -v -race -cover -tags=debug ./plugins/service
	go test -v -race -cover -tags=debug ./plugins/jobs/amqp
	go test -v -race -cover -tags=debug ./plugins/jobs/beanstalk
	go test -v -race -cover -tags=debug ./plugins/jobs/boltdb
	go test -v -race -cover -tags=debug ./plugins/jobs/durability
	go test -v -race -cover -tags=debug ./plugins/jobs/general
	go test -v -race -cover -tags=debug ./plugins/jobs/memory
	go test -v -race -cover -tags=debug ./plugins/jobs/nats
	go test -v -race -cover -tags=debug ./plugins/jobs/sqs
#	go test -v -race -cover -tags=debug ./plugins/kv
#	go test -v -race -cover -tags=debug ./plugins/tcp
#	go test -v -race -cover -tags=debug ./plugins/reload
#	go test -v -race -cover -tags=debug ./plugins/broadcast
#	go test -v -race -cover -tags=debug ./plugins/websockets
#	go test -v -race -cover -tags=debug ./plugins/proxy_ip_parser
#	go test -v -race -cover -tags=debug ./plugins/http
#	go test -v -race -cover -tags=debug ./plugins/grpc
#	go test -v -race -cover -tags=debug ./plugins/informer
#	go test -v -race -cover -tags=debug ./plugins/server
#	go test -v -race -cover -tags=debug ./plugins/status
#	go test -v -race -cover -tags=debug ./plugins/config
#	go test -v -race -cover -tags=debug ./plugins/gzip
#	go test -v -race -cover -tags=debug ./plugins/headers
#	go test -v -race -cover -tags=debug ./plugins/logger
#	go test -v -race -cover -tags=debug ./plugins/metrics
#	go test -v -race -cover -tags=debug ./plugins/resetter
#	go test -v -race -cover -tags=debug ./plugins/rpc

test_nightly:
	go test -v -race -cover -tags=debug,nightly ./plugins/http

	docker compose -f env/docker-compose.yaml kill
	docker compose -f env/docker-compose.yaml down