package commands

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the generate command

import (
	"net/http"

	middleware "github.com/go-openapi/runtime/middleware"
)

// WeaviateCommandsGetHandlerFunc turns a function with the right signature into a weaviate commands get handler
type WeaviateCommandsGetHandlerFunc func(WeaviateCommandsGetParams, interface{}) middleware.Responder

// Handle executing the request and returning a response
func (fn WeaviateCommandsGetHandlerFunc) Handle(params WeaviateCommandsGetParams, principal interface{}) middleware.Responder {
	return fn(params, principal)
}

// WeaviateCommandsGetHandler interface for that can handle valid weaviate commands get params
type WeaviateCommandsGetHandler interface {
	Handle(WeaviateCommandsGetParams, interface{}) middleware.Responder
}

// NewWeaviateCommandsGet creates a new http.Handler for the weaviate commands get operation
func NewWeaviateCommandsGet(ctx *middleware.Context, handler WeaviateCommandsGetHandler) *WeaviateCommandsGet {
	return &WeaviateCommandsGet{Context: ctx, Handler: handler}
}

/*WeaviateCommandsGet swagger:route GET /commands/{commandId} commands weaviateCommandsGet

Get a command based on its uuid related to this key.

Returns a particular command.

*/
type WeaviateCommandsGet struct {
	Context *middleware.Context
	Handler WeaviateCommandsGetHandler
}

func (o *WeaviateCommandsGet) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	route, _ := o.Context.RouteInfo(r)
	var Params = NewWeaviateCommandsGetParams()

	uprinc, err := o.Context.Authorize(r, route)
	if err != nil {
		o.Context.Respond(rw, r, route.Produces, route, err)
		return
	}
	var principal interface{}
	if uprinc != nil {
		principal = uprinc
	}

	if err := o.Context.BindValidRequest(r, route, &Params); err != nil { // bind params
		o.Context.Respond(rw, r, route.Produces, route, err)
		return
	}

	res := o.Handler.Handle(Params, principal) // actually handle the request

	o.Context.Respond(rw, r, route.Produces, route, res)

}
