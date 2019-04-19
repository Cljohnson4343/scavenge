// Package response provides a type and functions for http response errors
package response

import (
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"strings"
)

// Error wraps the strings.Builder type and implements the error interface. It also
// contains a http return code. The strings.Builder is used to build the error msg used
// in the "detail" field of the generated json. See the JSON() method docs for an example.
type Error struct {
	strings.Builder
	code int
	flag bool
}

// Error returns the error msg that has been built by the Error's embedded
// strings.Builder
func (err *Error) Error() string {
	return err.String()
}

// NewError returns a pointer to a Error that is initialized with the given arguments
func NewError(msg string, httpCode int) *Error {
	newErr := Error{code: httpCode, flag: true}
	_, err := newErr.WriteString(msg)
	if err != nil {
		panic(err)
	}

	return &newErr
}

// NewNilError returns an Error object that has not been populated with a message or
// http return status code yet. This can be used by consumers in place of an error flag.
// For example the following code:
//
//	e := response.NewError("", response.LowestPriorityCode)
//	encounteredErr := false
//
//	team := Team{}
//	for rows.Next() {
//		err = rows.Scan(&team.Name, &team.ID, &team.HuntID)
//		if err != nil {
//			encounteredErr = true
//			e.Add(err.Error(), http.StatusInternalServerError)
//		}
//
//		*teams = append(*teams, team)
//	}
//
//	if encounteredErr {
//		return teams, e
//	}
//
// Could be replaced with:
//
//	e := response.NewNilError()
//
//	team := Team{}
//	for rows.Next() {
//		err = rows.Scan(&team.Name, &team.ID, &team.HuntID)
//		if err != nil {
//			e.Add(err.Error(), http.StatusInternalServerError)
//		}
//
//		*teams = append(*teams, team)
//	}
//
//	return teams, e.GetError()
func NewNilError() *Error {
	return &Error{code: LowestPriorityCode, flag: false}
}

// GetError returns nil if there hasn't been an error added to
// the Error yet. This is only useful if you need Error's nil
// error feature. This is a special use case, for details see
// NewNilError's docs.
func (err *Error) GetError() *Error {
	if err.flag {
		return err
	}

	return nil
}

// Add adds an additional error to the Error. The msg
// will be appended to the generated json's 'detail' field. The
// Error's http status code might be updated depending on
// priority. Each Error only has one status code with more
// general codes being given higher priority.
func (err *Error) Add(msg string, httpCode int) {
	err.flag = true
	if err.code > httpCode {
		err.code = httpCode
	}

	// handle the case where NewNilError was used and there isn't a msg
	// in the buffer
	var sep string
	if err.Len() == 0 {
		sep = ""
	} else {
		sep = "; "
	}
	err.WriteString(fmt.Sprintf("%s%s", sep, msg))
}

// AddError is a convenience func for Add()
func (err *Error) AddError(e *Error) {
	err.Add(e.Error(), e.Code())
}

// LowestPriorityCode returns an int that will always be overridden when Add()
// is used.
const LowestPriorityCode = math.MaxInt32

// Code returns the http status code for this Error
func (err *Error) Code() int {
	return err.code
}

// res is an internal struct used to map a Error's data to json
type res struct {
	Code   int    `json:"code"`
	Status string `json:"status"`
	Detail string `json:"detail"`
}

// httpCodeMap is used to map a return code to its textual description
var httpCodeMap = map[int]string{
	400: "Bad Request",
	401: "Unauthorized",
	403: "Forbidden",
	404: "Not Found",
	500: "Internal Server Error",
}

// Handle writes a Error's return status code to the header and writes its generated
// json to the body.
func (err *Error) Handle(w http.ResponseWriter) {
	w.WriteHeader(err.Code())
	w.Write(err.JSON())
}

// JSON returns a []byte of the Error, for example:
//  {
// 		"code": 400,
//		"status": "Bad Request",
//		"detail": "The request body does not contain a required 'hunt_id' field"
// 	}
func (err *Error) JSON() []byte {
	r := res{Code: err.Code(), Status: httpCodeMap[err.Code()], Detail: err.String()}

	json, e := json.Marshal(&r)
	if e != nil {
		panic(e)
	}

	return json
}
