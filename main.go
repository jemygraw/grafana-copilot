package main

import (
	"flag"
	"fmt"
	"github.com/jemygraw/grafana-copilot/conf"
	"github.com/jemygraw/grafana-copilot/controllers"
	"log"
	"log/slog"
	"net/http"
	"os"
)

func initLogging(debug bool) {
	var logLevel slog.Level
	if debug {
		logLevel = slog.LevelDebug
	} else {
		logLevel = slog.LevelInfo
	}
	jsonHandler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: logLevel,
	})
	logger := slog.New(jsonHandler)
	slog.SetDefault(logger)
}

func main() {
	// parse flags
	var listenHost string
	var listenPort int
	var debug bool
	flag.StringVar(&listenHost, "host", "0.0.0.0", "The host to listen on")
	flag.IntVar(&listenPort, "port", 8080, "The port to listen on")
	flag.BoolVar(&debug, "debug", false, "Enable debug mode")
	flag.Parse()
	// parse envs
	conf.MustParseConfigFromEnvs()
	// init logging
	initLogging(debug)
	// listen server
	http.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	http.HandleFunc("/api/chatbot/infoflow-robot-callback", controllers.ReceiveInfoflowRobotMessage)
	slog.Info(fmt.Sprintf("Starting grafana copilot server on %s:%d ...", listenHost, listenPort))
	err := http.ListenAndServe(fmt.Sprintf("%s:%d", listenHost, listenPort), nil)
	if err != nil {
		log.Fatal(err.Error())
	}
}
