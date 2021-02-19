package main

import (
	"flag"
	"net/http"
	"net/url"
	"time"

	"github.com/gorilla/handlers"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rguilmont/deepdetect-exporter/pkg/ddclient"
	"github.com/sirupsen/logrus"
)

const readinessTime = 360

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Do stuff here
		logrus.Info("%v ")
		// Call the next handler, which can be another middleware in the chain, or the final handler.
		next.ServeHTTP(w, r)
	})
}

func initCollector(ddURL *url.URL) *ddclient.DeepDetectCollector {
	var err error
	for i := 0; i < readinessTime; i++ {
		ddCollector, err := ddclient.NewDeepDetectCollector(*ddURL)
		if err == nil {
			return ddCollector
		}
		logrus.Infof("Waiting for deepdetect to be available ( Currently %v )", err)
		time.Sleep(time.Second)
	}
	// Finally just panic
	logrus.Panicln(err)
	// Return nil because well... We've panic
	return nil
}

func main() {
	listen := flag.String("listen", "0.0.0.0:8081", "host:port to listen")
	monitor := flag.String("monitor", "http://localhost:8080", "DeepDetect URL to monitor")
	flag.Parse()

	logrus.Info("Starting DeepDetect exporter")
	ddURL, err := url.Parse(*monitor)
	if err != nil {
		logrus.Panicln("Impossible to parse URL ", *monitor, err)
	}

	ddCollector := initCollector(ddURL)
	logrus.Info("Monitoring DeepDetect at ", *monitor)

	prometheus.MustRegister(ddCollector)
	http.Handle("/metrics", promhttp.Handler())

	logrus.Info("Listening ", *listen)
	logrus.Fatal(http.ListenAndServe(*listen, handlers.LoggingHandler(logrus.New().Out, http.DefaultServeMux)))
}
