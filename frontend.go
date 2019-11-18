package main

import (
	"net/http"
	"time"

	"bytes"
	"fmt"
	"io"
	"net/http/httputil"

	opentracing "github.com/opentracing/opentracing-go"

	"github.com/gorilla/mux"
	"github.com/ksaritek/opentracing-sample/cmd"
	"github.com/ksaritek/opentracing-sample/pkg/tracing"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus/promhttp"

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

	client := &http.Client{}

	router := mux.NewRouter()
	router.Handle("/metrics", promhttp.Handler())
	router.HandleFunc("/hello/{user}", func(w http.ResponseWriter, r *http.Request) {
		body := &bytes.Buffer{}
		io.Copy(body, r.Body)
		req, _ := http.NewRequestWithContext(r.Context(), http.MethodPut, fmt.Sprintf("http://localhost:9090%s", r.URL), body)
		req.Header = r.Header.Clone()
		header, _ := httputil.DumpRequest(r, true)
		log.Infof("front-end request header >>> %s", string(header))

		res, _ := client.Do(req)

		w.WriteHeader(res.StatusCode)
		io.Copy(w, res.Body)

		tracer := opentracing.GlobalTracer()
		oCtx, _ := tracer.Extract(opentracing.HTTPHeaders, opentracing.HTTPHeadersCarrier(r.Header))

		if oCtx != nil {
			ch := tracer.StartSpan("front-end-childof", opentracing.ChildOf(oCtx))
			time.Sleep(1 * time.Millisecond)
			ch.Finish()

			f := tracer.StartSpan("front-end-followsFrom", opentracing.FollowsFrom(ch.Context()))
			time.Sleep(2 * time.Millisecond)
			f.Finish()
		}

	}).Methods("PUT")

	router.HandleFunc("/hello/{user}", func(w http.ResponseWriter, r *http.Request) {
		body := &bytes.Buffer{}
		io.Copy(body, r.Body)
		req, _ := http.NewRequestWithContext(r.Context(), http.MethodGet, fmt.Sprintf("http://localhost:9090%s", r.URL), body)
		req.Header = r.Header.Clone()
		header, _ := httputil.DumpRequest(r, true)
		log.Infof("front-end request header >>> %s", string(header))
		res, _ := client.Do(req)
		w.WriteHeader(res.StatusCode)
		io.Copy(w, res.Body)

		tracer := opentracing.GlobalTracer()
		oCtx, _ := tracer.Extract(opentracing.HTTPHeaders, opentracing.HTTPHeadersCarrier(r.Header))

		if oCtx != nil {
			ch := tracer.StartSpan("front-end-childof", opentracing.ChildOf(oCtx))
			time.Sleep(1 * time.Millisecond)
			ch.Finish()

			f := tracer.StartSpan("front-end-followsFrom", opentracing.FollowsFrom(ch.Context()))
			time.Sleep(2 * time.Millisecond)
			f.Finish()
		}
	}).Methods("GET")

	srv := &http.Server{
		Addr:    ":8080",
		Handler: tracing.HTTPMiddleware(router, "frontend"),
		// Good practice: enforce timeouts for servers you create!
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	log.Info("HTTP server is started")
	log.Fatal(srv.ListenAndServe())
}
