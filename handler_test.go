package problems

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestErrorHandlerMiddleware(t *testing.T) {
	// Set up the test case
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic(errors.New("test error"))
	})
	problemFactory := NewProblemFactory(traceId)

	middleware := problemFactory.ErrorHandlerMiddleware(handler)
	req, err := http.NewRequest("GET", "/test", nil)
	if err != nil {
		t.Fatal(err)
	}
	recorder := httptest.NewRecorder()

	// Call the middleware and check the response
	middleware.ServeHTTP(recorder, req)
	if recorder.Code != http.StatusInternalServerError {
		t.Errorf("Expected status code %d but got %d", http.StatusInternalServerError, recorder.Code)
	}
	expectedProblem := problemFactory.NewHTTPProblem(
		http.StatusText(http.StatusInternalServerError),
		"An internal server error occurred.",
		http.StatusInternalServerError,
		"test error",
	)
	expectedProblem.WithContext(req.Context())
	expectedProblem.WithExtension("instance", "/test")

	actualProblem, err := ParseHTTPResponse(recorder.Result())
	if err != nil {
		t.Fatal(err)
	}
	if !actualProblem.Equal(expectedProblem) {
		t.Errorf("Expected HTTP problem %+v but got %+v", expectedProblem, actualProblem)
	}
}
