package main

import (
	"fmt"

	"github.com/gocraft/web"
)

// Holds per-request application state, e.g. from middleware
type RequestContext struct {
	app *AppContext
}

func setupRoutes(app *AppContext) *web.Router {

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
		Get("/", (*RequestContext).Index).
		Get("/error/", (*RequestContext).Error)

	return router
}

func (c *RequestContext) Index(res web.ResponseWriter, req *web.Request) {
	c.app.logger.Println("got a request!")
	fmt.Fprintf(res, "%#v", c.app.config)
}

func (c *RequestContext) Error(res web.ResponseWriter, req *web.Request) {
	returnError(JSON{
		"error": "there was an error",
	}, 400)
}
