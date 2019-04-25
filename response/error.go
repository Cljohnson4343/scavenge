// Package response provides a type and functions for http response errors
package response

import (
	"encoding/json"
	"net/http"
	"strings"
)

// Error is a custom types that can accumulate errors.
type Error struct {
	errors []*error
}

// error wraps the strings.Builder type and implements the Error interface. It also
// contains a http return code. The strings.Builder is used to build the error msg used
// in the "detail" field of the generated json. See the JSON() method docs for an example.
type error struct {
	sb   strings.Builder
	code int
}

// NewError returns a pointer to a Error that is initialized with the given arguments
func NewError(msg string, httpCode int) *Error {
	es := make([]*error, 0)
	newErr := error{code: httpCode}
	_, err := newErr.sb.WriteString(msg)
	if err != nil {
		panic(err)
	}

	es = append(es, &newErr)

	e := Error{errors: es}

	return &e
}

// NewNilError returns an Error object that has not been populated with any errors.
// This can be used by consumers in place of an error flag.
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
//	return teams, e.GetErrors()
func NewNilError() *Error {
	return &Error{errors: nil}
}

// GetError returns nil if there aren't any errors in
// the Error yet. This is only useful if you need to instantiate
// nil Error using NewNilError. This is a special use case, for
// details see NewNilError's docs.
func (err *Error) GetError() *Error {
	if err.errors == nil {
		return nil
	}

	return err
}

// Add adds an additional error to the Error. The added error
// will generate its own json object.
func (err *Error) Add(msg string, httpCode int) {
	newErr := error{code: httpCode}
	newErr.sb.WriteString(msg)

	if err.errors == nil {
		err.errors = make([]*error, 0)
	}

	err.errors = append(err.errors, &newErr)
}

// AddError allows all the errors given to be added to the reciever Error
func (err *Error) AddError(e *Error) {
	if e.errors == nil {
		return
	}

	if err.errors == nil {
		err.errors = make([]*error, 0, len(e.errors))
	}

	err.errors = append(err.errors, e.errors...)
}

// res is an internal struct used to map an error's data to json
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

// Handle writes an Error's highest priority return status code to the header and writes its generated
// json to the body. The highest priority header is determined by the lowest valued status code.
func (err *Error) Handle(w http.ResponseWriter) {
	if err.errors == nil {
		panic("tried to handle a nil error")
	}

	highestPriorityCode := 4343
	for _, e := range err.errors {
		if e.code < highestPriorityCode {
			highestPriorityCode = e.code
		}
	}

	w.WriteHeader(highestPriorityCode)
	w.Write(err.JSON())
}

// JSON returns a []byte of the errors for Error, for example:
//  [{
// 			"code": 400,
//			"status": "Bad Request",
//			"detail": "The request body does not contain a required 'hunt_id' field"
// 		},{
// 			"code": 404,
//			"status": "Unauthorized",
//			"detail": "The user is Unauthorized."
//  	}
//  ]
func (err *Error) JSON() []byte {
	rs := make([]*res, 0, len(err.errors))
	for _, e := range err.errors {
		r := res{Code: e.code, Status: httpCodeMap[e.code], Detail: e.sb.String()}
		rs = append(rs, &r)
	}

	json, e := json.Marshal(&rs)
	if e != nil {
		panic(e)
	}

	return json
}
