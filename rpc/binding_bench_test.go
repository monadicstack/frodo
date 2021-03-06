package rpc_test

import (
	"context"
	"net/http"
	"net/url"
	"testing"

	"github.com/dimfeld/httptreemux/v5"
	"github.com/monadicstack/frodo/rpc"
)

func BenchmarkJsonBinder_Bind(b *testing.B) {
	type benchmarkFlipper bool
	type benchmarkPaging struct {
		Limit  int
		Offset int
		Sort   string `json:"order"`
	}
	type benchmarkRequest struct {
		ID   string           `json:"id"`
		Flag bool             `json:"flaggeroo"`
		Flip benchmarkFlipper `json:"flip"`
		Page benchmarkPaging  `json:"p"`
	}

	binder := rpc.NewGateway().Binder
	address, _ := url.Parse("http://localhost:8080/group/abcdef?flaggeroo=true&p.limit=42&p.offset=3&p.order=crap&p.missing=goo")
	request := &http.Request{
		URL: address,
	}

	ctx := context.Background()
	ctx = httptreemux.AddRouteDataToContext(ctx, mockRouteData{
		params: map[string]string{"id": "abcdef"},
	})
	request = request.WithContext(ctx)
	output := benchmarkRequest{}

	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = binder.Bind(request, &output)
	}
}

type mockRouteData struct {
	route  string
	params map[string]string
}

func (m mockRouteData) Route() string {
	return m.route
}

func (m mockRouteData) Params() map[string]string {
	return m.params
}
