package main

import (
	"context"
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
)

var (
	ServerPings = promauto.NewCounter(prometheus.CounterOpts{
		Name:      "server_total_ping_hits",
		Namespace: "server",
		Help:      "The total ammount of hits on the ping route",
	})
)

type Application struct {
	server *http.Server
}

func GetServerAddr() string {
	host, found := os.LookupEnv("SERVER_HOST")
	if !found {
		host = "0.0.0.0"
	}
	port, found := os.LookupEnv("SERVER_PORT")
	if !found {
		port = "9090"
	}

	return fmt.Sprintf("%s:%s", host, port)
}

func main() {

	// setup structured Logging
	jsonLogger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(jsonLogger)

	// setup mux and routes
	mux := http.NewServeMux()
	mux.Handle("GET /metrics", promhttp.Handler())
	mux.HandleFunc("GET /ping", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)

		// Increment and log ping metric
		ServerPings.Inc()
		slog.Info("Ping hit")
		w.Write([]byte("Pong mermo"))
	})

	// initialize the mux and server in the Application struct
	app := Application{
		server: &http.Server{
			Addr:    GetServerAddr(),
			Handler: mux,
		},
	}

	// create graceful shutdown channel
	shutdownChannel := make(chan struct{})

	go func() {
		sigInt := make(chan os.Signal, 1)
		signal.Notify(sigInt, os.Interrupt)
		<-sigInt

		slog.Info("Server: SigInt received, Gracefully shutting down")
		if err := app.server.Shutdown(context.Background()); err != nil {
			slog.Error(fmt.Sprintf("Server: Shutdown error: %s\n", err.Error()))
		}
		close(shutdownChannel)
	}()

	slog.Info("Listening on " + app.server.Addr)
	if err := app.server.ListenAndServe(); err != nil {
		slog.Error(fmt.Sprintf("Server: error on ListenAndServe: %s\n", err.Error()))
	}

	// wait on graceful shutdown channel
	<-shutdownChannel
	return
}
