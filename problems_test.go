package problems

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"reflect"
	"strings"
	"testing"
)

func TestProblemFactory_NewHTTPProblem(t *testing.T) {
	type fields struct {
		TraceKey *string
	}
	type args struct {
		ctx         context.Context
		problemType string
		title       string
		status      int
		detail      string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   *http.Response
	}{
		{
			name: "simple case",
			fields: fields{
				TraceKey: pointyStr(traceId),
			},
			args: args{
				ctx:         context.WithValue(context.Background(), traceId, "00000000-0000-0000-0000-000000000000"),
				problemType: "https://example.com/probs/out-of-credit",
				title:       "You do not have enough credit",
				status:      http.StatusForbidden,
				detail:      "Your current balance is 30, but that costs 50",
			},
			want: &http.Response{
				Status:     http.StatusText(http.StatusForbidden),
				StatusCode: http.StatusForbidden,
				Body: toBody(`{
					"detail":"Your current balance is 30, but that costs 50",
					"instance":"",
					"status":403,
					"title":"You do not have enough credit",
					"type":"https://example.com/probs/out-of-credit"}
				`),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pf := &ProblemFactory{
				TraceKey: tt.fields.TraceKey,
			}

			problem := pf.NewHTTPProblem(
				tt.args.problemType,
				tt.args.title,
				tt.args.status,
				tt.args.detail,
			)
			resp, err := problem.ToHTTPResponse()
			if err != nil {
				t.Fatal(err)
			}
			if resp.StatusCode != tt.want.StatusCode {
				t.Errorf("status code got = '%d', want '%d'", resp.StatusCode, tt.want.StatusCode)
			}
			eq, err := JSONEqual(resp.Body, tt.want.Body)
			if err != nil {
				t.Error(err)
			}
			if !eq {
				t.Errorf("response got = '%s', want '%s'", readBody(resp.Body), readBody(tt.want.Body))
			}
		})
	}
}

func readBody(body io.ReadCloser) string {
	all, err := io.ReadAll(body)
	if err != nil {
		return ""
	}
	return string(all)
}

func JSONEqual(a, b io.Reader) (bool, error) {
	var j, j2 interface{}
	d := json.NewDecoder(a)
	if err := d.Decode(&j); err != nil {
		return false, err
	}
	d = json.NewDecoder(b)
	if err := d.Decode(&j2); err != nil {
		return false, err
	}
	return reflect.DeepEqual(j2, j), nil
}

func toBody(s string) io.ReadCloser {
	return io.NopCloser(strings.NewReader(s))
}

const (
	traceId = "trace-id"
)

func pointyStr(s string) *string {
	return &s
}
