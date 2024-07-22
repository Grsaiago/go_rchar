package main

import (
	"context"
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/redis/go-redis/v9"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
)

// Variables for exporting metrics to prometheus
var (
	ServerPings = promauto.NewCounter(prometheus.CounterOpts{
		Name:      "total_ping_hits",
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

func GetRedisOptions() *redis.Options {
	var user string
	var password string
	var host string
	var port string
	var redis_db string
	var exists bool
	var err error

	if host, exists = os.LookupEnv("REDIS_HOST"); exists != true {
		slog.Warn("redis host set to default [localhost]")
		host = "localhost"
	}
	if port, exists = os.LookupEnv("REDIS_PORT"); exists != true {
		slog.Warn("redis port set to default [6379]")
		port = "6379"
	}
	if password, exists = os.LookupEnv("REDIS_PASSWORD"); exists != true {
		slog.Warn("redis password set to default [empty]")
	}
	if redis_db, exists = os.LookupEnv("REDIS_DB"); exists != true {
		slog.Warn("redis db set to default [0]")
		redis_db = "0"
	}

	redis_opt, err := redis.ParseURL(fmt.Sprintf("redis://%s:%s@%s:%s/%s", user, password, host, port, redis_db))
	if err != nil {
		slog.Error(fmt.Sprintf("redis.ParseURL: %s", err.Error()))
		panic(err)
	}
	return redis_opt
}

func main() {

	// setup structured Logging
	jsonLogger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(jsonLogger)

	// setup redis connection
	rdb := redis.NewClient(GetRedisOptions())
	if rdb == nil {
		slog.Error(fmt.Sprintf("Server: error connecting to redis"))
		return
	}

	// setup mux and routes
	mux := http.NewServeMux()
	mux.Handle("GET /metrics", promhttp.Handler())
	mux.HandleFunc("GET /ping", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)

		// Increment and log ping metric
		ServerPings.Inc()
		slog.Info("ping hit")
		w.Write([]byte("pong"))
	})

	mux.HandleFunc("POST /key/{key}", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)

		// Increment and log ping metric
		slog.Info(fmt.Sprintf("POST on /key/{%s}", r.PathValue("key")))
		w.Write([]byte(r.PathValue("key")))
	})

	mux.HandleFunc("GET /key/{key}", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)

		// Increment and log ping metric
		slog.Info(fmt.Sprintf("GET on /key/{%s}", r.PathValue("key")))
		w.Write([]byte(r.PathValue("key")))
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
		return
	}()

	slog.Info("Listening on " + app.server.Addr)
	if err := app.server.ListenAndServe(); err != nil {
		slog.Error(fmt.Sprintf("Server: error on ListenAndServe: %s\n", err.Error()))
	}

	// wait on graceful shutdown channel
	<-shutdownChannel
	return
}
