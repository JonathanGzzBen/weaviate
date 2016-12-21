package models




import (
	strfmt "github.com/go-openapi/strfmt"
	"github.com/go-openapi/swag"

	"github.com/go-openapi/errors"
)

// DevicesCreateLocalAuthTokensResponse devices create local auth tokens response
// swagger:model DevicesCreateLocalAuthTokensResponse
type DevicesCreateLocalAuthTokensResponse struct {

	// Identifies what kind of resource this is. Value: the fixed string "weave#devicesCreateLocalAuthTokensResponse".
	Kind *string `json:"kind,omitempty"`

	// Minted device and client tokens.
	MintedLocalAuthTokens []*MintedLocalAuthInfo `json:"mintedLocalAuthTokens"`
}

// Validate validates this devices create local auth tokens response
func (m *DevicesCreateLocalAuthTokensResponse) Validate(formats strfmt.Registry) error {
	var res []error

	if err := m.validateMintedLocalAuthTokens(formats); err != nil {
		// prop
		res = append(res, err)
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}

func (m *DevicesCreateLocalAuthTokensResponse) validateMintedLocalAuthTokens(formats strfmt.Registry) error {

	if swag.IsZero(m.MintedLocalAuthTokens) { // not required
		return nil
	}

	for i := 0; i < len(m.MintedLocalAuthTokens); i++ {

		if swag.IsZero(m.MintedLocalAuthTokens[i]) { // not required
			continue
		}

		if m.MintedLocalAuthTokens[i] != nil {

			if err := m.MintedLocalAuthTokens[i].Validate(formats); err != nil {
				return err
			}
		}

	}

	return nil
}
