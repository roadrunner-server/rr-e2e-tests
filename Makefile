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
	go test -v -race -cover -tags=debug -coverpkg=all -failfast -coverprofile=./coverage-ci/beanstalk.out -covermode=atomic ./plugins/jobs/beanstalk
	go test -v -race -cover -tags=debug -coverpkg=all -failfast -coverprofile=./coverage-ci/boltdb.out -covermode=atomic ./plugins/jobs/boltdb
	go test -v -race -cover -tags=debug -coverpkg=all -failfast -coverprofile=./coverage-ci/durability.out -covermode=atomic ./plugins/jobs/durability
	go test -v -race -cover -tags=debug -coverpkg=all -failfast -coverprofile=./coverage-ci/general.out -covermode=atomic ./plugins/jobs/general
	go test -v -race -cover -tags=debug -coverpkg=all -failfast -coverprofile=./coverage-ci/memory.out -covermode=atomic ./plugins/jobs/memory
	go test -v -race -cover -tags=debug -coverpkg=all -failfast -coverprofile=./coverage-ci/nats.out -covermode=atomic ./plugins/jobs/nats
	go test -v -race -cover -tags=debug -coverpkg=all -failfast -coverprofile=./coverage-ci/sqs.out -covermode=atomic ./plugins/jobs/sqs
	go test -v -race -cover -tags=debug -coverpkg=all -failfast -coverprofile=./coverage-ci/kv_plugin.out -covermode=atomic ./plugins/kv
	go test -v -race -cover -tags=debug -coverpkg=all -failfast -coverprofile=./coverage-ci/tcp_plugin.out -covermode=atomic ./plugins/tcp
	go test -v -race -cover -tags=debug -coverpkg=all -failfast -coverprofile=./coverage-ci/proxy_ip.out -covermode=atomic ./plugins/proxy_ip_parser
	go test -v -race -cover -tags=debug -coverpkg=all -failfast -coverprofile=./coverage-ci/reload.out -covermode=atomic ./plugins/reload
	go test -v -race -cover -tags=debug -coverpkg=all -failfast -coverprofile=./coverage-ci/attributes.out -covermode=atomic ./plugins/http/attributes
	go test -v -race -cover -tags=debug -coverpkg=all -failfast -coverprofile=./coverage-ci/handler.out -covermode=atomic ./plugins/http/handler
	go test -v -race -cover -tags=debug -coverpkg=all -failfast -coverprofile=./coverage-ci/http.out -covermode=atomic ./plugins/http/http
	go test -v -race -cover -tags=debug -coverpkg=all -failfast -coverprofile=./coverage-ci/otlp.out -covermode=atomic ./plugins/http/otlp
	go test -v -race -cover -tags=debug -coverpkg=all -failfast -coverprofile=./coverage-ci/uploads.out -covermode=atomic ./plugins/http/uploads
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
	go test -v -race -tags=debug ./plugins/temporal
	go test -v -race -tags=debug ./plugins/service
	go test -v -race -tags=debug ./plugins/jobs/amqp
	go test -v -race -tags=debug ./plugins/jobs/beanstalk
	go test -v -race -tags=debug ./plugins/jobs/boltdb
	go test -v -race -tags=debug ./plugins/jobs/durability
	go test -v -race -tags=debug ./plugins/jobs/general
	go test -v -race -tags=debug ./plugins/jobs/memory
	go test -v -race -tags=debug ./plugins/jobs/kafka
	go test -v -race -tags=debug ./plugins/jobs/nats
	go test -v -race -tags=debug ./plugins/jobs/sqs
	go test -v -race -tags=debug ./plugins/kv
	go test -v -race -tags=debug ./plugins/tcp
	go test -v -race -tags=debug ./plugins/reload
	go test -v -race -tags=debug ./plugins/proxy_ip_parser
	go test -v -race -tags=debug ./plugins/http/attributes
	go test -v -race -tags=debug ./plugins/http/handler
	go test -v -race -tags=debug ./plugins/http/http
	go test -v -race -tags=debug ./plugins/http/otlp
	go test -v -race -tags=debug ./plugins/http/uploads
	go test -v -race -tags=debug ./plugins/grpc
	go test -v -race -tags=debug ./plugins/informer
	go test -v -race -tags=debug ./plugins/server
	go test -v -race -tags=debug ./plugins/status
	go test -v -race -tags=debug ./plugins/config
	go test -v -race -tags=debug ./plugins/gzip
	go test -v -race -tags=debug ./plugins/headers
	go test -v -race -tags=debug ./plugins/logger
	go test -v -race -tags=debug ./plugins/metrics
	go test -v -race -tags=debug ./plugins/resetter
	go test -v -race -tags=debug ./plugins/rpc

test_nightly:
	go test -v -race -cover -tags=debug,nightly ./plugins/http

	docker compose -f env/docker-compose.yaml kill
	docker compose -f env/docker-compose.yaml down

# only 1 sample here
regenerate_test_proto:
	protoc --plugin=protoc-gen-php-grpc --proto_path=plugins/grpc/proto/service --php_out=php_test_files/src --php-grpc_out=php_test_files/src service.proto

# local generate certs
generate-test-local-certs:
	mkcert localhost 127.0.0.1 ::1
	mkcert -client localhost 127.0.0.1 ::1
	cp (mkcert -CAROOT)/rootCA.pem test-certs
