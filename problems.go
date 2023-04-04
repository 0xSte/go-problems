package problems

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type ProblemFactory struct {
	TraceKey *string
}

type HTTPProblem struct {
	Trace      string
	Type       string
	Title      string
	Detail     string
	Status     int
	Instance   string
	Extensions map[string]any
}

func (pf *ProblemFactory) NewHTTPProblem(ctx context.Context, problemType, title string, status int, detail string) *HTTPProblem {
	traceVal := ctx.Value(pf.TraceKey)
	var trace string
	if pf.TraceKey != nil {
		trace = traceVal.(string)
	}
	return &HTTPProblem{
		Trace:  trace,
		Type:   problemType,
		Title:  title,
		Status: status,
		Detail: detail,
	}
}

// ToHTTPResponse implements the https://www.ietf.org/rfc/rfc7807 spec
func (p *HTTPProblem) ToHTTPResponse() (*http.Response, error) {
	jsonBytes, err := json.Marshal(p)
	if err != nil {
		// unlikely
		return nil, err
	}
	resp := &http.Response{
		StatusCode: p.Status,
		Header: http.Header{
			"Content-Type": []string{"application/problem+json"},
		},
		Body: io.NopCloser(bytes.NewReader(jsonBytes)),
	}
	return resp, nil
}

func (p *HTTPProblem) WithExtension(key string, value any) error {
	if isReservedExtensionName(key) {
		return &ErrReservedField{key}
	}
	p.Extensions[key] = value
	return nil
}

type HttpProblems []HTTPProblem

// ToHTTPResponse implements the https://www.ietf.org/rfc/rfc4918 spec
func (hs *HttpProblems) ToHTTPResponse() (*http.Response, error) {
	jsonBytes, err := json.Marshal(hs)
	if err != nil {
		// unlikely
		return nil, err
	}
	resp := &http.Response{
		StatusCode: http.StatusMultiStatus,
		Header: http.Header{
			"Content-Type": []string{"application/problem+json"},
		},
		Body: io.NopCloser(bytes.NewReader(jsonBytes)),
	}
	return resp, nil

}

type ErrReservedField struct {
	Field string
}

func (e *ErrReservedField) Error() string {
	return fmt.Sprintf("invalid field used in extensions: %s", e.Field)
}

func (u *HTTPProblem) MarshalJSON() ([]byte, error) {
	m := map[string]any{
		fieldType:     u.Type,
		fieldTitle:    u.Title,
		fieldDetail:   u.Detail,
		fieldStatus:   u.Status,
		fieldInstance: u.Instance,
	}
	for extn, _ := range u.Extensions {
		if isReservedExtensionName(extn) {
			return nil, &ErrReservedField{extn}
		}
	}
	// iterate again so we don't have risk of partial writes
	for k, v := range u.Extensions {
		m[k] = v
	}
	return json.Marshal(m)
}

const (
	fieldType     = "type"
	fieldTitle    = "title"
	fieldDetail   = "detail"
	fieldStatus   = "status"
	fieldInstance = "instance"
)

func isReservedExtensionName(field string) bool {
	var fields = []string{fieldType, fieldTitle, fieldDetail, fieldStatus, fieldInstance}
	for _, f := range fields {
		if field == f {
			return true
		}
	}
	return false
}
