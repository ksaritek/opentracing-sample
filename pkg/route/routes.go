package route

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/ksaritek/opentracing-sample/pkg/db"

	"github.com/go-redis/redis"
	"github.com/pkg/errors"

	"github.com/gorilla/mux"
	opentracing "github.com/opentracing/opentracing-go"
	otlog "github.com/opentracing/opentracing-go/log"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

//NewRoutes returns defined routes
func NewRoutes(repository db.Repository) *mux.Router {
	router := mux.NewRouter()

	router.Handle("/metrics", promhttp.Handler())
	router.HandleFunc("/healthz", healthz(repository)).Methods("GET")
	router.HandleFunc("/readiness", readiness(repository)).Methods("GET")
	router.HandleFunc("/hello/{user}", upsertDateOfBirth(repository)).Methods("PUT")
	router.HandleFunc("/hello/{user}", getBirthDate(repository)).Methods("GET")

	return router
}

//DTO is incoming upsert request for Date of Birth & reply to user request
type DTO struct {
	DateOfBirth string `json:"dateOfBirth,omitempty"`
	Message     string `json:"message,omitempty"`
	ErrMessage  string `json:"error,omitempty"`
}

func healthz(repository db.Repository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		if err := repository.IsOk(); err != nil {
			http.Error(w, "", http.StatusServiceUnavailable)
			return
		}

		w.WriteHeader(http.StatusOK)
		return
	}
}

func readiness(repository db.Repository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		if err := repository.IsReady(); err != nil {
			http.Error(w, "", http.StatusServiceUnavailable)
			return
		}

		w.WriteHeader(http.StatusOK)
		return
	}
}

func upsertDateOfBirth(repository db.Repository) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {
		tracer := opentracing.GlobalTracer()
		oCtx, err := tracer.Extract(opentracing.HTTPHeaders, opentracing.HTTPHeadersCarrier(r.Header))

		var span opentracing.Span
		if err == nil {
			span = tracer.StartSpan("upsert-birthday", opentracing.ChildOf(oCtx))
		} else {
			span = tracer.StartSpan("upsert-birthday")
		}
		defer span.Finish()
		spanCtx := opentracing.ContextWithSpan(r.Context(), span)
		span.SetTag("http.method", r.Method)

		vars := mux.Vars(r)
		user := vars["user"]

		// Read body
		b, err := ioutil.ReadAll(r.Body)
		if err != nil {
			span.SetTag("error", true)
			span.SetTag("http.status_code", http.StatusBadRequest)
			span.LogFields(
				otlog.String("event", "error"),
				otlog.String("message", "could not read request body"),
			)
			http.Error(w, fmt.Sprintf("{\"error\": \"%s\"}", err.Error()), http.StatusBadRequest)
			return
		}
		defer r.Body.Close()

		// Unmarshal
		var dto DTO
		err = json.Unmarshal(b, &dto)
		if err != nil {
			span.SetTag("error", true)
			span.SetTag("http.status_code", http.StatusBadRequest)
			span.LogFields(
				otlog.String("event", "error"),
				otlog.String("message", "could not unmarshall json body"),
			)
			http.Error(w, fmt.Sprintf("{\"error\": \"%s\"}", err.Error()), http.StatusBadRequest)
			return
		}

		if isValidDateFormat(dto.DateOfBirth) == false {
			err := errors.New(fmt.Sprintf("%s date format is invalid, date must be in yyyy-mm-dd format", dto.DateOfBirth))
			span.SetTag("error", true)
			span.SetTag("http.status_code", http.StatusBadRequest)
			span.LogFields(
				otlog.String("event", "error"),
				otlog.String("message", err.Error()),
			)
			http.Error(w, fmt.Sprintf("{\"error\": \"%s\"}", err.Error()), http.StatusBadRequest)
			return
		}

		if err := repository.Upsert(spanCtx, user, dto.DateOfBirth); err != nil {
			span.SetTag("error", true)
			span.SetTag("http.status_code", http.StatusInternalServerError)
			span.LogFields(
				otlog.String("event", "error"),
				otlog.String("message", "could not upsert user date of birth"),
			)
			http.Error(w, fmt.Sprintf("{\"error\": \"%s\"}", err.Error()), http.StatusInternalServerError)
			return
		}

		span.LogFields(
			otlog.String("event", "upsert-birthday"),
			otlog.String("value", user),
		)
		w.WriteHeader(http.StatusNoContent)
		return
	}
}

func getBirthDate(repository db.Repository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tracer := opentracing.GlobalTracer()
		oCtx, err := tracer.Extract(opentracing.HTTPHeaders, opentracing.HTTPHeadersCarrier(r.Header))

		var span opentracing.Span
		if err == nil {
			span = tracer.StartSpan("check-birthday", opentracing.ChildOf(oCtx))
		} else {
			span = tracer.StartSpan("check-birthday")
		}
		defer span.Finish()
		spanCtx := opentracing.ContextWithSpan(r.Context(), span)

		span.SetTag("http.method", r.Method)

		vars := mux.Vars(r)
		user := vars["user"]
		span.LogFields(
			otlog.String("event", "check-birthday"),
			otlog.String("value", user),
		)

		dateOfBirth, err := repository.Get(spanCtx, user)
		if err == redis.Nil {
			span.SetTag("http.status_code", http.StatusNotFound)
			w.WriteHeader(http.StatusNotFound)
			return
		} else if err != nil {
			span.SetTag("error", err)
			span.SetTag("http.status_code", http.StatusInternalServerError)
			http.Error(w, fmt.Sprintf("{\"error\": \"%s\"}", err.Error()), http.StatusInternalServerError)
			return
		}

		msg, err := msgOfBirthday(user, dateOfBirth, time.Now())
		if err != nil {
			http.Error(w, fmt.Sprintf("{\"error\": \"%s\"}", err.Error()), http.StatusInternalServerError)
			return
		}

		res, err := json.Marshal(DTO{Message: msg})
		if err != nil {
			http.Error(w, fmt.Sprintf("{\"error\": \"%s\"}", err.Error()), http.StatusInternalServerError)
			return
		}
		span.SetTag("http.status_code", http.StatusOK)
		w.Write(res)
		return
	}
}

func msgOfBirthday(user string, dateOfBirth string, now time.Time) (string, error) {
	now = now.Truncate(24 * time.Hour)

	layout := "2006-01-02"
	bd, err := time.Parse(layout, dateOfBirth)
	if err != nil {
		return "", errors.Wrap(err, "failed to parse dateOfBirth:")
	}

	bdThisYear, err := time.Parse(layout, fmt.Sprintf("%04d-%02d-%02d", now.Year(), bd.Month(), bd.Day()))
	if err != nil {
		return "", errors.Wrap(err, "failed to parse dateOfBirth:")
	}

	bdYearDay := bdThisYear.YearDay()
	nowYearDay := now.YearDay()

	dayDiff := bdYearDay - nowYearDay

	switch {
	case dayDiff == 0:
		return fmt.Sprintf("Hello, %s! Happy birthday", user), nil
	case dayDiff == 1:
		return fmt.Sprintf("Hello, %s! Your birthday is in 1 day", user), nil
	case dayDiff > 1:
		return fmt.Sprintf("Hello, %s! Your birthday is in %d days", user, dayDiff), nil
	default:
		return fmt.Sprintf("Hello, %s! Your birthday is passed this year", user), nil
	}
}

func isValidDateFormat(date string) bool {
	layout := "2006-01-02"
	_, err := time.Parse(layout, date)
	if err != nil {
		return false
	}

	return true
}
