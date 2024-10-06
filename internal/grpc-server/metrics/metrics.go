package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

const appName = "chat"

type Metrics struct {
	openedChats      prometheus.Gauge
	connectedClients prometheus.Gauge
}

func NewMetrics() *Metrics {
	return &Metrics{
		openedChats: promauto.NewGauge(prometheus.GaugeOpts{
			Namespace: "chat",
			Subsystem: "stream",
			Name:      appName + "_opened_chat_cnt",
			Help:      "Кол-во открытых чатов",
		}),
		connectedClients: promauto.NewGauge(prometheus.GaugeOpts{
			Namespace: "chat",
			Subsystem: "stream",
			Name:      appName + "_connected_clients_cnt",
			Help:      "Кол-во подключенных клиентов",
		}),
	}
}

func (m *Metrics) IncreaseChats() {
	m.openedChats.Inc()
}

func (m *Metrics) DecreaseChats() {
	m.openedChats.Dec()
}

func (m *Metrics) IncreaseClients() {
	m.connectedClients.Inc()
}

func (m *Metrics) DecreaseClients() {
	m.connectedClients.Dec()
}
