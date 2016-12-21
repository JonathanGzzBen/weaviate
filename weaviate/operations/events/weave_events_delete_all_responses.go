package events




import (
	"net/http"

	"github.com/go-openapi/runtime"
)

/*WeaveEventsDeleteAllOK Successful response

swagger:response weaveEventsDeleteAllOK
*/
type WeaveEventsDeleteAllOK struct {
}

// NewWeaveEventsDeleteAllOK creates WeaveEventsDeleteAllOK with default headers values
func NewWeaveEventsDeleteAllOK() *WeaveEventsDeleteAllOK {
	return &WeaveEventsDeleteAllOK{}
}

// WriteResponse to the client
func (o *WeaveEventsDeleteAllOK) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.WriteHeader(200)
}
