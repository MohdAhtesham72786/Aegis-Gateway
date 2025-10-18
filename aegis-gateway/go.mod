module github.com/example/aegis-gateway

go 1.20

require (
	go.opentelemetry.io/otel v1.20.0
	go.opentelemetry.io/otel/exporters/stdout/stdouttrace v1.20.0
	go.opentelemetry.io/otel/sdk v1.20.0
	gopkg.in/yaml.v3 v3.0.0-20210107192922-496545a6307b
	github.com/fsnotify/fsnotify v1.5.4
	github.com/gorilla/mux v1.8.0
	gopkg.in/natefinch/lumberjack.v2 v2.0.0
	github.com/prometheus/client_golang v1.16.0
)
