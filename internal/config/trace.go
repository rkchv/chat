package config

import "time"

// Trace настройки трейсинга
type Trace struct {
	ServiceName         string        `yaml:"service_name" env:"TRACE_SERVICE_NAME" env-default:"auth-service"`
	BatchTimeout        time.Duration `yaml:"batch_timeout" env:"TRACE_BATCH_TIMEOUT" env-default:"1s"`
	ExporterGRPCAddress string        `yaml:"exporter_grpc_address" env:"TRACE_EXPORTER_GRPC_ADDRESS" env-default:"localhost:4317"`
}
