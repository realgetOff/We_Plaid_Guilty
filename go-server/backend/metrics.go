package main

import (
	"github.com/prometheus/client_golang/prometheus"
)

var ( // PROMETHEUS METRICS	
	activeWebsockets = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "active_websockets",
		Help: "The current number of open / active websocket connections.",
	})
	dbRequests = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "db_requests",
		Help: "Number of SQL requests made to the postgresql database.",
	})
	dbRequestsSucessful = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "db_requests_successful",
		Help: "Number of successful SQL requests made to the postgresql database.",
	})
)

func registerMetrics () {
	prometheus.MustRegister(activeWebsockets)
	prometheus.MustRegister(dbRequests)
	prometheus.MustRegister(dbRequestsSucessful)
}