package main

import (
	"log"
	"net/http"
	"strconv"

	"github.com/prometheus/client_golang/prometheus"
)

var (
	zombies = prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: namespace,
		Subsystem: subsystem,
		Name:      "zombies_reaped",
		Help:      "The number of zombies reaped.",
	})
)

func init() {
	prometheus.MustRegister(zombies)
}

func metrics() {
	http.Handle("/metrics", prometheus.Handler())
	go func() {
		err := http.ListenAndServe(":"+strconv.Itoa(port), nil)
		if err != nil {
			log.Fatal("dinit: %s", err)
		}
	}()
}
