package locations

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"net/http"

	"github.com/go-openapi/errors"
	"github.com/go-openapi/runtime/middleware"

	strfmt "github.com/go-openapi/strfmt"
)

// NewWeaviateLocationsDeleteParams creates a new WeaviateLocationsDeleteParams object
// with the default values initialized.
func NewWeaviateLocationsDeleteParams() WeaviateLocationsDeleteParams {
	var ()
	return WeaviateLocationsDeleteParams{}
}

// WeaviateLocationsDeleteParams contains all the bound params for the weaviate locations delete operation
// typically these are obtained from a http.Request
//
// swagger:parameters weaviate.locations.delete
type WeaviateLocationsDeleteParams struct {

	// HTTP Request Object
	HTTPRequest *http.Request

	/*
	  Required: true
	  In: path
	*/
	LocationID string
}

// BindRequest both binds and validates a request, it assumes that complex things implement a Validatable(strfmt.Registry) error interface
// for simple values it will use straight method calls
func (o *WeaviateLocationsDeleteParams) BindRequest(r *http.Request, route *middleware.MatchedRoute) error {
	var res []error
	o.HTTPRequest = r

	rLocationID, rhkLocationID, _ := route.Params.GetOK("locationId")
	if err := o.bindLocationID(rLocationID, rhkLocationID, route.Formats); err != nil {
		res = append(res, err)
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}

func (o *WeaviateLocationsDeleteParams) bindLocationID(rawData []string, hasKey bool, formats strfmt.Registry) error {
	var raw string
	if len(rawData) > 0 {
		raw = rawData[len(rawData)-1]
	}

	o.LocationID = raw

	return nil
}
