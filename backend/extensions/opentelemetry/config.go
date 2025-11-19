package opentelemetry

type Config struct {
	ServiceName string `envconfig:"SERVICE_NAME" default:"lnk-backend"`
	Endpoint    string `envconfig:"OTEL_EXPORTER_OTLP_ENDPOINT" default:"localhost:4317"`
}
