package main

import (
	"encoding/base64"
	"encoding/json"
	"io/ioutil"
	"net"
	"net/http"
	"strings"
)

const EMPTY_GIF = "GIF89a\x01\x00\x01\x00\x80\x00\x00\xff\xff\xff\x00\x00\x00!\xf9\x04\x00\x00\x00\x00\x00,\x00\x00\x00\x00\x01\x00\x01\x00\x00\x02\x02D\x01\x00;"

func (app *AppContext) setupRoutes() http.Handler {

	/* Create a new router. */
	router := http.NewServeMux()

	// Map routes
	router.HandleFunc("/track.gif", app.TrackQueryParams)
	router.HandleFunc("/track/", app.TrackPostedJSON)
	router.HandleFunc("/track/base64/", app.TrackEncodedJSON)

	// TODO: setup expvar metrics to replace request stats
	// TODO: use an expvar metric for logging uploads, failures, etc.
	// TODO: expose expvar handler through the router

	return router
}

// TrackQueryParams decodes the "data" query parameter as base64, and logs the
// resulting data directly to the S3 logger.
func (app *AppContext) TrackEncodedJSON(res http.ResponseWriter, req *http.Request) {
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
}

// TrackPostedJSON logs the body of the request directly to the S3 logger.
func (app *AppContext) TrackPostedJSON(res http.ResponseWriter, req *http.Request) {
	var body []byte
	var err error

	defer req.Body.Close()
	if body, err = ioutil.ReadAll(req.Body); err != nil {
		res.WriteHeader(400)
		return
	}

	app.Logf(string(body))
}

// TrackQueryParams parses the request parameters from the query and serializes
// the result into JSON, then logs a message to the S3 logger.
func (app *AppContext) TrackQueryParams(res http.ResponseWriter, req *http.Request) {
	if err := req.ParseForm(); err != nil {
		res.WriteHeader(400)
		return
	}

	// Flatten query params using only the first value for each key
	data := map[string]string{}
	for key, value := range req.Form {
		data[key] = value[0]
	}

	result, err := json.Marshal(data)
	if err != nil {
		res.WriteHeader(400)
		return
	}

	app.Logf(string(result))

	if strings.HasSuffix(req.URL.Path, ".gif") {
		res.Header().Set("Content-Type", "image/gif")
		res.Write([]byte(EMPTY_GIF))
	}
}

// clientError writes a HTTP client error status code and textual response to the ResponseWriter.
func (app *AppContext) clientError(res http.ResponseWriter, message string) {
	res.WriteHeader(400)
	res.Write([]byte(message))
}

// isLocalRequest returns true if the request came from localhost (127.0.0.1).
func isLocalRequest(req *http.Request) bool {
	remote_ip := net.ParseIP(strings.Split(req.RemoteAddr, ":")[0])

	if remote_ip.String() != "127.0.0.1" {
		return false
	}

	return true
}
