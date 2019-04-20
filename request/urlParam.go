package request

import (
	"net/http"
	"strconv"

	"github.com/cljohnson4343/scavenge/response"
	"github.com/go-chi/chi"
)

// GetIntURLParam wraps chi.URLParam and conversion to a function that returns
// a response.Error
func GetIntURLParam(r *http.Request, str string) (int, *response.Error) {
	v, err := strconv.Atoi(chi.URLParam(r, str))
	if err != nil {
		return 0, response.NewError(err.Error(), http.StatusBadRequest)
	}
	return v, nil
}
