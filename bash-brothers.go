package main

import (
	"context"
	_ "embed"
	"errors"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"
	"go.uber.org/automaxprocs/maxprocs"
)

const port = "8080"

//nolint:gochecknoglobals // Globals required for embed
var (
	//go:embed web/static/Arma_3_Preset_Bash_Brothers.html
	arma3PresetBashBrothers []byte

	//go:embed web/static/ts3.html
	ts3 []byte
)

func newMux(log *logrus.Entry) *http.ServeMux {
	mux := http.NewServeMux()

	mux.Handle("/metrics", promhttp.Handler())

	mux.HandleFunc("/ts3", func(w http.ResponseWriter, r *http.Request) {
		log.Infof("Handling request for %s", r.URL.Path)

		discardCloseRequestBody(log, r.Body)

		w.Header().Set("Content-Type", "text/html; charset=UTF-8")
		w.Header().Set("Content-Length", strconv.Itoa(len(ts3)))

		_, err := w.Write(ts3)
		if err != nil {
			log.WithError(err).Errorf("Error writing %s response", r.URL.Path)
		}
	})

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		log.Infof("Handling request for %s", r.URL.Path)

		discardCloseRequestBody(log, r.Body)

		w.Header().Set("Content-Type", "text/html; charset=UTF-8")
		w.Header().Set("Content-Length", strconv.Itoa(len(arma3PresetBashBrothers)))

		_, err := w.Write(arma3PresetBashBrothers)
		if err != nil {
			log.WithError(err).Errorf("Error writing %s response", r.URL.Path)
		}
	})

	return mux
}

func discardCloseRequestBody(log *logrus.Entry, body io.ReadCloser) {
	_, err := io.Copy(ioutil.Discard, body)
	if err != nil {
		log.WithError(err).Warn("Error discarding request body")
	}

	err = body.Close()
	if err != nil {
		log.WithError(err).Warn("Error closing request body")
	}
}

func listenAndServer(log *logrus.Entry, httpServer *http.Server) {
	err := httpServer.ListenAndServe()
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.WithError(err).Error("Error serving HTTP")
	}
}

func main() {
	log := logrus.NewEntry(&logrus.Logger{
		Out: os.Stdout,
		Formatter: &logrus.TextFormatter{
			FullTimestamp: true,
		},
		Level: logrus.InfoLevel,
	})

	log.Info("Server starting up")
	defer log.Info("Server stopped")

	_, err := maxprocs.Set()
	if err != nil {
		log.WithError(err).Fatal("Failed to set GOMAXPROCS")
	}

	httpServer := &http.Server{
		Addr:    "0.0.0.0:" + port,
		Handler: newMux(log),
	}

	go listenAndServer(log, httpServer)

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGTERM)
	<-stop // Block until the OS signal

	err = httpServer.Shutdown(context.Background())
	if err != nil {
		log.WithError(err).Error("Error stopping HTTP server gracefully")
	}
}
