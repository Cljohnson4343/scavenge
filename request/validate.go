package request

import (
	"encoding/json"
	"net/http"

	"github.com/asaskevich/govalidator"
	"github.com/cljohnson4343/scavenge/pgsql"
	"github.com/cljohnson4343/scavenge/response"
)

// Validater is used by structs that can be input from a request
type Validater interface {
	Validate(r *http.Request) *response.Error
}

// PatchValidater only validates non-zero value fields
type PatchValidater interface {
	PatchValidate(r *http.Request, entityID int) *response.Error
}

// DecodeAndValidate is the entry point for deserialization and validation of request json.
// It decodes the json-encoded body of the request and stores it into the value pointed to
// by v. v is then validated.
func DecodeAndValidate(r *http.Request, v Validater) *response.Error {
	e := decode(r, v)
	if e != nil {
		return e
	}

	return v.Validate(r)
}

func decode(r *http.Request, v interface{}) *response.Error {
	err := json.NewDecoder(r.Body).Decode(v)
	if err != nil {
		return response.NewError(http.StatusBadRequest, err.Error())
	}
	defer r.Body.Close()

	return nil
}

// DecodeAndPatchValidate is the entry point for deserialization and validation of request json.
// It decodes the json-encoded body of the request and stores it into the value pointed to
// by v. v then has only its non-zero value fields validated.
func DecodeAndPatchValidate(r *http.Request, v PatchValidater, entityID int) *response.Error {
	e := decode(r, v)
	if e != nil {
		return e
	}

	return v.PatchValidate(r, entityID)
}

// PatchValidate uses govalidator.ValidateStruct to validate only the non-zero fields for
// the given v which must be type struct
func PatchValidate(colMap pgsql.ColumnMap, v interface{}) *response.Error {
	_, err := govalidator.ValidateStruct(v)
	if err == nil {
		return nil
	}

	e := response.NewNilError()
	for col := range colMap {
		errStr := govalidator.ErrorByField(err, col)
		if errStr != "" {
			e.Add(http.StatusBadRequest, errStr)
		}
	}

	return e.GetError()
}
