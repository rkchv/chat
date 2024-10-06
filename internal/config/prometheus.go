package config

import "net"

// Prometheus настройки работы сервера prometheus
type Prometheus struct {
	Host string `yaml:"host" env:"PROMETHEUS_HOST" env-default:"0.0.0.0"`
	Port string `yaml:"port" env:"PROMETHEUS_PORT" env-default:"9095"`
}

// Address адрес подключения
func (r Prometheus) Address() string {
	return net.JoinHostPort(r.Host, r.Port)
}
