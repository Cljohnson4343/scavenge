// Package err provides a type and functions for http response errors
package err

import (
	"encoding/json"
	"strings"
)

// ResponseError wraps the strings.Builder type and implements the error interface. It also
// contains a http return code. The strings.Builder is used to build the error msg used
// in the "detail" field of the generated json. See the JSON() method docs for an example.
type ResponseError struct {
	strings.Builder
	code int
}

// Error returns the error msg that has been built by the ResponseError's embedded
// strings.Builder
func (err *ResponseError) Error() string {
	return err.String()
}

// New returns a pointer to a ResponseError that is initialized with the given arguments
func New(msg string, httpCode int) *ResponseError {
	newErr := ResponseError{code: httpCode}
	_, err := newErr.WriteString(msg)
	if err != nil {
		panic(err)
	}

	return &newErr
}

// Code returns the http status code for this ResponseError
func (err *ResponseError) Code() int {
	return err.code
}

// res is an internal struct used to map a ResponseError's data to json
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

// JSON returns a []byte of the ResponseError, for example:
//  {
// 		"code": 400,
//		"status": "Bad Request",
//		"detail": "The request body does not contain a required 'hunt_id' field"
// 	}
func (err *ResponseError) JSON() []byte {
	r := res{Code: err.Code(), Status: httpCodeMap[err.Code()], Detail: err.String()}

	json, e := json.Marshal(&r)
	if e != nil {
		panic(e)
	}

	return json
}
