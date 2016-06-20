package main

import (
	"encoding/base64"
	"encoding/json"
	"io/ioutil"
	"net"
	"net/http"
	"strings"
	"time"
)

const EMPTY_GIF = "GIF89a\x01\x00\x01\x00\x80\x00\x00\xff\xff\xff\x00\x00\x00!\xf9\x04\x00\x00\x00\x00\x00,\x00\x00\x00\x00\x01\x00\x01\x00\x00\x02\x02D\x01\x00;"

func (app *AppContext) setupRoutes() http.Handler {

	/* Create a new router. */
	router := http.NewServeMux()

	// Map routes
	router.Handle("/", app.Middleware(app.Track))
	router.Handle("/track.gif", app.Middleware(app.TrackEncodedQueryParam))

	// Serve global handlers (i.e. expvar) only to local
	router.HandleFunc("/debug/vars", func(res http.ResponseWriter, req *http.Request) {
		if !isLocalRequest(req) {
			res.WriteHeader(http.StatusForbidden)
			return
		}
		http.DefaultServeMux.ServeHTTP(res, req)
	})

	return router
}

// Track routes to the appropriate method based on the incoming request's content
func (app *AppContext) Track(res http.ResponseWriter, req *http.Request) {
	switch {
	case req.Method == "POST":
		app.TrackPostedBody(res, req)

	case req.Method == "GET":
		app.TrackQueryParams(res, req)

	default:
		res.WriteHeader(http.StatusMethodNotAllowed)
	}
}

// TrackPostedBody logs the body of the request directly to the S3 logger.
func (app *AppContext) TrackPostedBody(res http.ResponseWriter, req *http.Request) {
	var body []byte
	var err error

	defer req.Body.Close()
	if body, err = ioutil.ReadAll(req.Body); err != nil {
		res.WriteHeader(http.StatusBadRequest)
		return
	}

	app.Logf(string(body))
	res.WriteHeader(http.StatusNoContent)
}

// TrackQueryParams parses the request parameters from the query and serializes
// the result into JSON, then logs a message to the S3 logger.
func (app *AppContext) TrackQueryParams(res http.ResponseWriter, req *http.Request) {
	if err := req.ParseForm(); err != nil {
		res.WriteHeader(http.StatusBadRequest)
		return
	}

	// Flatten query params using only the first value for each key
	data := map[string]string{}
	for key, value := range req.Form {
		data[key] = value[0]
	}

	result, err := json.Marshal(data)
	if err != nil {
		res.WriteHeader(http.StatusBadRequest)
		return
	}

	app.Logf(string(result))

	res.WriteHeader(http.StatusNoContent)
}

// TrackEncodedQueryParam decodes the "data" query parameter as base64, and
// logs the resulting data directly to the S3 logger.
//
// This allows passing JSON-encoded data safely from the browser, or baking a
// request into an image URL.
func (app *AppContext) TrackEncodedQueryParam(res http.ResponseWriter, req *http.Request) {
	var data []byte
	var err error

	param := req.FormValue("data")
	if param == "" {
		app.clientError(res, "Error: missing parameter data")
		return
	}

	// Decode base64-encoded data
	decoder := base64.NewDecoder(base64.StdEncoding, strings.NewReader(param))
	if data, err = ioutil.ReadAll(decoder); err != nil {
		app.clientError(res, "Failed to parse base64-encoded data")
		return
	}

	app.Logf(string(data))

	// Return empty GIF
	res.Header().Set("Content-Type", "image/gif")
	res.WriteHeader(http.StatusOK)
	res.Write([]byte(EMPTY_GIF))
}

// Middleware wraps a HandlerFunc with a middleware handler for metrics, panic
// recovery, and logging.
func (app *AppContext) Middleware(handler http.HandlerFunc) http.Handler {
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {

		// Record response time after the handler returns
		defer func(begin time.Time) {
			app.metrics.ResponseTime.Observe(time.Since(begin))
		}(time.Now())

		// Record total number of requests
		app.metrics.RequestCount.Add(1)

		handler(res, req)
	})
}

// clientError writes a HTTP client error status code and textual response to the ResponseWriter.
func (app *AppContext) clientError(res http.ResponseWriter, message string) {
	res.WriteHeader(http.StatusBadRequest)
	res.Write([]byte(message))
}

// isLocalRequest returns true if the request came from localhost (127.0.0.1).
func isLocalRequest(req *http.Request) bool {
	host, _, err := net.SplitHostPort(req.RemoteAddr)
	if err != nil {
		return false
	}

	switch host {

	case "127.0.0.1":
		return true

	case "::1":
		return true

	default:
		return false
	}
}
