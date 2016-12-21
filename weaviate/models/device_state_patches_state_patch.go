package models




import (
	strfmt "github.com/go-openapi/strfmt"

	"github.com/go-openapi/errors"
)

// DeviceStatePatchesStatePatch Device state patch with corresponding timestamp.
// swagger:model DeviceStatePatchesStatePatch
type DeviceStatePatchesStatePatch struct {

	// Component name paths separated by '/'.
	Component string `json:"component,omitempty"`

	// State patch.
	Patch JSONObject `json:"patch,omitempty"`

	// Timestamp of a change. Local time, UNIX timestamp or time since last boot can be used.
	TimeMs int64 `json:"timeMs,omitempty"`
}

// Validate validates this device state patches state patch
func (m *DeviceStatePatchesStatePatch) Validate(formats strfmt.Registry) error {
	var res []error

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}
