package main

import (
	"net/http"
	"time"

	"github.com/ksaritek/opentracing-sample/cmd"
	"github.com/ksaritek/opentracing-sample/pkg/db"
	"github.com/ksaritek/opentracing-sample/pkg/route"
	"github.com/ksaritek/opentracing-sample/pkg/tracing"

	"github.com/pkg/errors"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

func init() {
	viper.SetDefault("address", ":8080")
}

func main() {
	cmd := cmd.NewRoot()
	if err := cmd.Execute(); err != nil {
		panic(errors.Wrap(err, "could not execute command:"))
	}

	log.Infof("config_file: %s", viper.ConfigFileUsed())

	_, closer, err := tracing.Init("opentracing-sample")
	if err != nil {
		panic(errors.Wrap(err, "could not start opentracing:"))
	}
	defer closer.Close()
	log.Info("opentracing is configured")

	router := route.NewRoutes(db.NewRepository())
	srv := &http.Server{
		Addr:    ":9090",
		Handler: tracing.HTTPMiddleware(router, "backend"),
		// Good practice: enforce timeouts for servers you create!
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	log.Info("HTTP server is started")
	log.Fatal(srv.ListenAndServe())
}
