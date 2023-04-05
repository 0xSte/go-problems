package problems

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"reflect"
)

type ProblemFactory struct {
	TraceKey *string
}

func NewProblemFactory(traceKey string) *ProblemFactory {
	return &ProblemFactory{TraceKey: &traceKey}
}

type HTTPProblem struct {
	Type     string
	Title    string
	Detail   string
	Status   int
	Instance string

	extensions map[string]any
	traceKey   *string
}

func (pf *ProblemFactory) NewHTTPProblem(problemType, title string, status int, detail string) *HTTPProblem {
	return &HTTPProblem{
		Type:     problemType,
		Title:    title,
		Status:   status,
		Detail:   detail,
		traceKey: pf.TraceKey,
	}
}

// ToHTTPResponse implements the https://www.ietf.org/rfc/rfc7807 spec
func (hp *HTTPProblem) ToHTTPResponse() (*http.Response, error) {
	jsonBytes, err := json.Marshal(hp)
	if err != nil {
		// unlikely
		return nil, err
	}
	resp := &http.Response{
		StatusCode: hp.Status,
		Header: http.Header{
			"Content-Type": []string{"application/problem+json"},
		},
		Body: io.NopCloser(bytes.NewReader(jsonBytes)),
	}
	return resp, nil
}

func (hp *HTTPProblem) WithExtension(key string, value any) error {
	if isReservedExtensionName(key) {
		return &ErrReservedField{key}
	}
	hp.extensions[key] = value
	return nil
}

func (hp *HTTPProblem) WithContext(ctx context.Context) {
	traceVal := ctx.Value(hp.traceKey)
	if traceVal != nil {
		s := traceVal.(string)
		hp.traceKey = &s
	}
}

type HttpProblems []HTTPProblem

func (pf *ProblemFactory) NewHttpProblems(problems ...HTTPProblem) HttpProblems {
	return HttpProblems(problems)
}

// ToHTTPResponse implements the https://www.ietf.org/rfc/rfc4918 spec
func (hs *HttpProblems) ToHTTPResponse() (*http.Response, error) {
	jsonBytes, err := json.Marshal(hs)
	if err != nil {
		// unlikely
		return nil, err
	}

	bodyReader := bytes.NewBuffer(jsonBytes)

	resp := &http.Response{
		StatusCode: http.StatusMultiStatus,
		Header: http.Header{
			"Content-Type": []string{"application/problem+json"},
		},
		Body: io.NopCloser(bodyReader),
	}
	return resp, nil
}

func ParseHTTPResponse(resp *http.Response) (*HTTPProblem, error) {
	defer resp.Body.Close()

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// Parse the response body as an HTTP problem
	var problem HTTPProblem
	err = json.Unmarshal(body, &problem)
	if err != nil {
		return nil, err
	}

	return &problem, nil
}

type ErrReservedField struct {
	Field string
}

func (e *ErrReservedField) Error() string {
	return fmt.Sprintf("invalid field used in extensions: %s", e.Field)
}

func (hp *HTTPProblem) MarshalJSON() ([]byte, error) {
	m := map[string]any{
		fieldType:     hp.Type,
		fieldTitle:    hp.Title,
		fieldDetail:   hp.Detail,
		fieldStatus:   hp.Status,
		fieldInstance: hp.Instance,
	}
	for extn, _ := range hp.extensions {
		if isReservedExtensionName(extn) {
			return nil, &ErrReservedField{extn}
		}
	}
	// iterate again so we don't have risk of partial writes
	for k, v := range hp.extensions {
		m[k] = v
	}
	return json.Marshal(m)
}

func (hp *HTTPProblem) Equal(other *HTTPProblem) bool {
	if hp.Title != other.Title {
		return false
	}
	if hp.Detail != other.Detail {
		return false
	}
	if hp.Status != other.Status {
		return false
	}
	if hp.Instance != other.Instance {
		return false
	}
	if !reflect.DeepEqual(hp.extensions, other.extensions) {
		return false
	}
	return true
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
