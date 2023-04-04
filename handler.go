package problems

import (
	"fmt"
	"net/http"
)

func (pf *ProblemFactory) ErrorHandlerMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		defer func() {
			if r := recover(); r != nil {
				err, ok := r.(error)
				if !ok {
					err = fmt.Errorf("%v", r)
				}

				// Create a new HTTP Problem with the recovered error
				problem := pf.NewHTTPProblem(
					ctx,
					http.StatusText(http.StatusInternalServerError),
					"An internal server error occurred.",
					http.StatusInternalServerError,
					err.Error(),
				)

				// Write the HTTP Problem as the response
				resp, err := problem.ToHTTPResponse()
				resp.Write(w)
			}
		}()
		next.ServeHTTP(w, r)
	})
}
