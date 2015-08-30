package main

import (
	"encoding/base64"
	"encoding/json"
	"strings"

	"github.com/gocraft/web"
)

// Holds per-request application state, e.g. from middleware
type RequestContext struct {
	app *AppContext
}

func (app *AppContext) setupRoutes() *web.Router {

	/* Create a new router. RequestContext instance is only passed to tell it
	   what type of context object to pass to the handlers (it's not reused) */
	router := web.New(RequestContext{})

	// Log all requests
	router.Middleware(web.LoggerMiddleware)

	// Use a closure to set RequestContext.app in every request
	router.Middleware(func(c *RequestContext, res web.ResponseWriter, req *web.Request, next web.NextMiddlewareFunc) {
		c.app = app
		next(res, req)
	})

	router.Middleware(APIErrorMiddleware)

	// Map routes
	router = router.
		Get("/track/", (*RequestContext).TrackQueryParams).
		Get("/track.gif", (*RequestContext).TrackQueryParams).
		Post("/track/", (*RequestContext).TrackPostedJSON).
		Get("/track/base64/", (*RequestContext).TrackEncodedJSON)

	return router
}

func (c *RequestContext) TrackEncodedJSON(res web.ResponseWriter, req *web.Request) {
	param := req.FormValue("data")
	if param == "" {
		returnError(JSON{"error": "data parameter must be passed"}, 400)
	}

	var data map[string]interface{}

	// Decode base64-encoded data
	decoder := base64.NewDecoder(base64.StdEncoding, strings.NewReader(param))

	if err := json.NewDecoder(decoder).Decode(&data); err != nil {
		returnError(JSON{
			"error": "failed to parse data parameter, should be base64-encoded JSON",
		}, 400)
	}

	result, err := json.Marshal(data)
	if err != nil {
		returnError(JSON{"error": "failed to marshal JSON"}, 500)
	}

	c.app.Logf(string(result))
}

func (c *RequestContext) TrackPostedJSON(res web.ResponseWriter, req *web.Request) {
	var data map[string]interface{}

	if err := json.NewDecoder(req.Body).Decode(&data); err != nil {
		returnError(JSON{"error": "failed to parse JSON body"}, 400)
	}

	result, err := json.Marshal(data)
	if err != nil {
		returnError(JSON{"error": "failed to marshal JSON"}, 500)
	}

	c.app.Logf(string(result))
}

func (c *RequestContext) TrackQueryParams(res web.ResponseWriter, req *web.Request) {
	if err := req.ParseForm(); err != nil {
		returnError(JSON{"error": "failed to parse request"}, 400)
	}

	// Flatten query params using only the first value for each key
	data := map[string]string{}
	for key, value := range req.Form {
		data[key] = value[0]
	}

	result, err := json.Marshal(data)
	if err != nil {
		returnError(JSON{"error": "failed to marshal JSON"}, 500)
	}

	c.app.Logf(string(result))

	if strings.HasSuffix(req.URL.Path, ".gif") {
		res.Header().Set("Content-Type", "image/gif")
		res.Write(EMPTY_GIF)
	}
}
