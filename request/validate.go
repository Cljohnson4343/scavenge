package request

import (
	"encoding/json"
	"net/http"

	"github.com/cljohnson4343/scavenge/response"
)

// Validater is used by structs that can be input from a request
type Validater interface {
	Validate(r *http.Request) *response.Error
}

// DecodeAndValidate is the entry point for deserialization and validation of request json.
// It decodes the json-encoded body of the request and stores it into the value pointed to
// by v. v is then validated.
func DecodeAndValidate(r *http.Request, v Validater) *response.Error {
	err := json.NewDecoder(r.Body).Decode(v)
	if err != nil {
		return response.NewError(err.Error(), http.StatusBadRequest)
	}
	defer r.Body.Close()

	return v.Validate(r)
}
