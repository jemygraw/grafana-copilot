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

func init() {
	jsonHandler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})
	logger := slog.New(jsonHandler)
	slog.SetDefault(logger)
}

func main() {
	// parse flags
	var listenHost string
	var listenPort int
	flag.StringVar(&listenHost, "host", "0.0.0.0", "The host to listen on")
	flag.IntVar(&listenPort, "port", 8080, "The port to listen on")
	flag.Parse()
	// parse envs
	conf.MustParseConfigFromEnvs()
	// listen server
	http.HandleFunc("/infoflow-robot-callback", controllers.ReceiveInfoflowRobotMessage)
	slog.Info(fmt.Sprintf("Starting grafana copilot server on %s:%d ...", listenHost, listenPort))
	err := http.ListenAndServe(fmt.Sprintf("%s:%d", listenHost, listenPort), nil)
	if err != nil {
		log.Fatal(err.Error())
	}
}
