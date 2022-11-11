module github.com/roadrunner-server/rr-e2e-tests

go 1.19

require (
	github.com/Shopify/toxiproxy/v2 v2.5.0
	github.com/aws/aws-sdk-go-v2 v1.17.1
	github.com/aws/aws-sdk-go-v2/config v1.17.11
	github.com/aws/aws-sdk-go-v2/credentials v1.12.24
	github.com/aws/aws-sdk-go-v2/service/sqs v1.19.12
	github.com/beanstalkd/go-beanstalk v0.1.0
	github.com/centrifugal/centrifuge-go v0.9.3
	github.com/fatih/color v1.13.0
	github.com/goccy/go-json v0.9.11
	github.com/google/uuid v1.3.0
	github.com/nats-io/nats.go v1.19.1
	github.com/pborman/uuid v1.2.1
	github.com/prometheus/client_golang v1.14.0
	github.com/rabbitmq/amqp091-go v1.5.0
	github.com/roadrunner-server/amqp/v3 v3.0.0-beta.3
	github.com/roadrunner-server/beanstalk/v3 v3.0.0-beta.3
	github.com/roadrunner-server/boltdb/v3 v3.0.0-beta.2
	github.com/roadrunner-server/centrifuge/v3 v3.0.0-beta.4
	github.com/roadrunner-server/config/v3 v3.0.0-beta.3
	github.com/roadrunner-server/endure v1.4.5
	github.com/roadrunner-server/errors v1.2.0
	github.com/roadrunner-server/fileserver/v3 v3.0.0-beta.3
	github.com/roadrunner-server/goridge/v3 v3.6.1
	github.com/roadrunner-server/grpc/v3 v3.0.0-beta.3
	github.com/roadrunner-server/gzip/v3 v3.0.0-beta.2
	github.com/roadrunner-server/headers/v3 v3.0.0-beta.4
	github.com/roadrunner-server/http/v3 v3.0.0-beta.3
	github.com/roadrunner-server/informer/v3 v3.0.0-beta.3
	github.com/roadrunner-server/jobs/v3 v3.0.0-beta.3
	github.com/roadrunner-server/kafka/v3 v3.0.0-beta.2
	github.com/roadrunner-server/kv/v3 v3.0.0-beta.2
	github.com/roadrunner-server/logger/v3 v3.0.0-beta.2
	github.com/roadrunner-server/memcached/v3 v3.0.0-beta.2
	github.com/roadrunner-server/memory/v3 v3.0.0-beta.3
	github.com/roadrunner-server/metrics/v3 v3.0.0-beta.2
	github.com/roadrunner-server/nats/v3 v3.0.0-beta.2
	github.com/roadrunner-server/otel/v3 v3.0.0-beta.2
	github.com/roadrunner-server/prometheus/v3 v3.0.0-beta.2
	github.com/roadrunner-server/proxy_ip_parser/v3 v3.0.0-beta.2
	github.com/roadrunner-server/redis/v3 v3.0.0-beta.3
	github.com/roadrunner-server/reload/v3 v3.0.0-beta.2
	github.com/roadrunner-server/resetter/v3 v3.0.0-beta.2
	github.com/roadrunner-server/rpc/v3 v3.0.0-beta.2
	github.com/roadrunner-server/sdk/v3 v3.0.0-beta.5
	github.com/roadrunner-server/send/v3 v3.0.0-beta.2
	github.com/roadrunner-server/server/v3 v3.0.0-beta.6
	github.com/roadrunner-server/service/v3 v3.0.0-beta.3
	github.com/roadrunner-server/sqs/v3 v3.0.0-beta.2
	github.com/roadrunner-server/static/v3 v3.0.0-beta.2
	github.com/roadrunner-server/status/v3 v3.0.0-beta.2
	github.com/roadrunner-server/tcp/v3 v3.0.0-beta.3
	github.com/stretchr/testify v1.8.1
	github.com/temporalio/roadrunner-temporal/v2 v2.0.0-beta.4
	github.com/yookoala/gofast v0.7.0
	go.buf.build/protocolbuffers/go/roadrunner-server/api v1.3.27
	go.temporal.io/api v1.12.0
	go.temporal.io/sdk v1.17.0
	go.uber.org/zap v1.23.0
	golang.org/x/net v0.2.0
	google.golang.org/grpc v1.50.1
	google.golang.org/protobuf v1.28.1
)

exclude go.temporal.io/api v1.13.0

require (
	github.com/Shopify/sarama v1.37.2 // indirect
	github.com/andybalholm/brotli v1.0.4 // indirect
	github.com/aws/aws-sdk-go-v2/feature/ec2/imds v1.12.19 // indirect
	github.com/aws/aws-sdk-go-v2/internal/configsources v1.1.25 // indirect
	github.com/aws/aws-sdk-go-v2/internal/endpoints/v2 v2.4.19 // indirect
	github.com/aws/aws-sdk-go-v2/internal/ini v1.3.26 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/presigned-url v1.9.19 // indirect
	github.com/aws/aws-sdk-go-v2/service/sso v1.11.25 // indirect
	github.com/aws/aws-sdk-go-v2/service/ssooidc v1.13.8 // indirect
	github.com/aws/aws-sdk-go-v2/service/sts v1.17.2 // indirect
	github.com/aws/smithy-go v1.13.4 // indirect
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/bradfitz/gomemcache v0.0.0-20221031212613-62deef7fc822 // indirect
	github.com/cactus/go-statsd-client/statsd v0.0.0-20200423205355-cb0885a1018c // indirect
	github.com/caddyserver/certmagic v0.17.2 // indirect
	github.com/cenkalti/backoff/v4 v4.1.3 // indirect
	github.com/centrifugal/protocol v0.8.11 // indirect
	github.com/cespare/xxhash/v2 v2.1.2 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
	github.com/eapache/go-resiliency v1.3.0 // indirect
	github.com/eapache/go-xerial-snappy v0.0.0-20180814174437-776d5712da21 // indirect
	github.com/eapache/queue v1.1.0 // indirect
	github.com/emicklei/proto v1.11.0 // indirect
	github.com/facebookgo/clock v0.0.0-20150410010913-600d898af40a // indirect
	github.com/felixge/httpsnoop v1.0.3 // indirect
	github.com/fsnotify/fsnotify v1.6.0 // indirect
	github.com/go-logr/logr v1.2.3 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/go-ole/go-ole v1.2.6 // indirect
	github.com/go-redis/redis/v8 v8.11.5 // indirect
	github.com/gofiber/fiber/v2 v2.39.0 // indirect
	github.com/gogo/googleapis v1.4.1 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/gogo/status v1.1.1 // indirect
	github.com/golang/mock v1.6.0 // indirect
	github.com/golang/protobuf v1.5.2 // indirect
	github.com/golang/snappy v0.0.4 // indirect
	github.com/gorilla/websocket v1.5.0 // indirect
	github.com/grpc-ecosystem/go-grpc-middleware v1.3.0 // indirect
	github.com/grpc-ecosystem/grpc-gateway/v2 v2.13.0 // indirect
	github.com/hashicorp/errwrap v1.1.0 // indirect
	github.com/hashicorp/go-multierror v1.1.1 // indirect
	github.com/hashicorp/go-uuid v1.0.3 // indirect
	github.com/hashicorp/go-version v1.6.0 // indirect
	github.com/hashicorp/hcl v1.0.0 // indirect
	github.com/jcmturner/aescts/v2 v2.0.0 // indirect
	github.com/jcmturner/dnsutils/v2 v2.0.0 // indirect
	github.com/jcmturner/gofork v1.7.6 // indirect
	github.com/jcmturner/gokrb5/v8 v8.4.3 // indirect
	github.com/jcmturner/rpc/v2 v2.0.3 // indirect
	github.com/josharian/intern v1.0.0 // indirect
	github.com/jpillora/backoff v1.0.0 // indirect
	github.com/klauspost/compress v1.15.12 // indirect
	github.com/klauspost/cpuid/v2 v2.2.0 // indirect
	github.com/libdns/libdns v0.2.1 // indirect
	github.com/magiconair/properties v1.8.6 // indirect
	github.com/mailru/easyjson v0.7.7 // indirect
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.16 // indirect
	github.com/mattn/go-runewidth v0.0.14 // indirect
	github.com/matttproud/golang_protobuf_extensions v1.0.4 // indirect
	github.com/mholt/acmez v1.0.4 // indirect
	github.com/miekg/dns v1.1.50 // indirect
	github.com/minio/highwayhash v1.0.2 // indirect
	github.com/mitchellh/mapstructure v1.5.0 // indirect
	github.com/nats-io/jwt/v2 v2.2.1-0.20220113022732-58e87895b296 // indirect
	github.com/nats-io/nkeys v0.3.0 // indirect
	github.com/nats-io/nuid v1.0.1 // indirect
	github.com/openzipkin/zipkin-go v0.4.1 // indirect
	github.com/pelletier/go-toml v1.9.5 // indirect
	github.com/pelletier/go-toml/v2 v2.0.5 // indirect
	github.com/pierrec/lz4/v4 v4.1.17 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/prometheus/client_model v0.3.0 // indirect
	github.com/prometheus/common v0.37.0 // indirect
	github.com/prometheus/procfs v0.8.0 // indirect
	github.com/rcrowley/go-metrics v0.0.0-20201227073835-cf1acfcdf475 // indirect
	github.com/rivo/uniseg v0.4.2 // indirect
	github.com/roadrunner-server/tcplisten v1.2.0 // indirect
	github.com/robfig/cron v1.2.0 // indirect
	github.com/segmentio/asm v1.2.0 // indirect
	github.com/segmentio/encoding v0.3.6 // indirect
	github.com/shirou/gopsutil v3.21.11+incompatible // indirect
	github.com/spf13/afero v1.9.2 // indirect
	github.com/spf13/cast v1.5.0 // indirect
	github.com/spf13/jwalterweatherman v1.1.0 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	github.com/spf13/viper v1.14.0 // indirect
	github.com/stretchr/objx v0.5.0 // indirect
	github.com/subosito/gotenv v1.4.1 // indirect
	github.com/tklauser/go-sysconf v0.3.10 // indirect
	github.com/tklauser/numcpus v0.6.0 // indirect
	github.com/twmb/murmur3 v1.1.6 // indirect
	github.com/uber-go/tally/v4 v4.1.3 // indirect
	github.com/valyala/bytebufferpool v1.0.0 // indirect
	github.com/valyala/fasthttp v1.41.0 // indirect
	github.com/valyala/tcplisten v1.0.0 // indirect
	github.com/vmihailenco/msgpack/v5 v5.3.5 // indirect
	github.com/vmihailenco/tagparser/v2 v2.0.0 // indirect
	github.com/yusufpapurcu/wmi v1.2.2 // indirect
	go.buf.build/grpc/go/roadrunner-server/api v1.4.27 // indirect
	go.etcd.io/bbolt v1.3.6 // indirect
	go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp v0.36.4 // indirect
	go.opentelemetry.io/contrib/propagators/jaeger v1.11.1 // indirect
	go.opentelemetry.io/otel v1.11.1 // indirect
	go.opentelemetry.io/otel/exporters/jaeger v1.11.1 // indirect
	go.opentelemetry.io/otel/exporters/otlp/internal/retry v1.11.1 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlptrace v1.11.1 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc v1.11.1 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp v1.11.1 // indirect
	go.opentelemetry.io/otel/exporters/stdout/stdouttrace v1.11.1 // indirect
	go.opentelemetry.io/otel/exporters/zipkin v1.11.1 // indirect
	go.opentelemetry.io/otel/metric v0.33.0 // indirect
	go.opentelemetry.io/otel/sdk v1.11.1 // indirect
	go.opentelemetry.io/otel/trace v1.11.1 // indirect
	go.opentelemetry.io/proto/otlp v0.19.0 // indirect
	go.temporal.io/sdk/contrib/tally v0.2.0 // indirect
	go.temporal.io/server v1.18.4 // indirect
	go.uber.org/atomic v1.10.0 // indirect
	go.uber.org/multierr v1.8.0 // indirect
	golang.org/x/crypto v0.1.0 // indirect
	golang.org/x/mod v0.6.0 // indirect
	golang.org/x/sync v0.1.0 // indirect
	golang.org/x/sys v0.2.0 // indirect
	golang.org/x/text v0.4.0 // indirect
	golang.org/x/time v0.2.0 // indirect
	golang.org/x/tools v0.2.0 // indirect
	google.golang.org/genproto v0.0.0-20221107162902-2d387536bcdd // indirect
	gopkg.in/ini.v1 v1.67.0 // indirect
	gopkg.in/natefinch/lumberjack.v2 v2.0.0 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)
