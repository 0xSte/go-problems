package problems

import (
	"fmt"
	"net/http"
)

func (pf *ProblemFactory) ErrorHandlerMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		path := r.URL.Path
		defer func() {
			if r := recover(); r != nil {
				err, ok := r.(error)
				if !ok {
					err = fmt.Errorf("%v", r)
				}

				// Create a new HTTP Problem with the recovered error
				problem := pf.NewHTTPProblem(
					http.StatusText(http.StatusInternalServerError),
					"An internal server error occurred.",
					http.StatusInternalServerError,
					err.Error(),
				)
				problem.WithContext(ctx)
				_ = problem.WithExtension("instance", path)

				// Write the HTTP Problem as the response
				resp, err := problem.ToHTTPResponse()
				w.WriteHeader(http.StatusInternalServerError)
				resp.Write(w)
			}
		}()
		next.ServeHTTP(w, r)
	})
}
