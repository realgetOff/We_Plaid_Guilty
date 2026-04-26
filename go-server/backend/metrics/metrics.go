package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
)

var ( // PROMETHEUS METRICS	
	ActiveWebsockets = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "active_websockets",
		Help: "The current number of open / active websocket connections.",
	})
	DbRequests = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "db_requests",
		Help: "Number of SQL requests made to the postgresql database.",
	})
	DbRequestsSucessful = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "db_requests_successful",
		Help: "Number of successful SQL requests made to the postgresql database.",
	})
	RoomCountTotal = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "room_count",
		Help: "Total number of rooms that are currently active.",
	})
	RoomCountStandard = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "room_count_standard",
		Help: "Number of standard rooms that are currently active.",
	})
	RoomCountAI = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "room_count_ai",
		Help: "Number of AI game rooms that are currently active.",
	})
)

func RegisterMetrics () {
	// WS METRICS
	prometheus.MustRegister(ActiveWebsockets)
	// DB METRICS
	prometheus.MustRegister(DbRequests)
	prometheus.MustRegister(DbRequestsSucessful)
	// GAME METRICS
	prometheus.MustRegister(RoomCountTotal)
	prometheus.MustRegister(RoomCountStandard)
	prometheus.MustRegister(RoomCountAI)
}