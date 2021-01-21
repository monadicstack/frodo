package rpc

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"testing"

	"github.com/julienschmidt/httprouter"
)

func BenchmarkJsonBinder_Bind(b *testing.B) {
	b.ReportAllocs()
	binder := jsonBinder{}
	address, _ := url.Parse("http://localhost:8080/group/abcdef?flaggeroo=true&page.limit=42&p.offset=3&p.order=crap&p.missing=goo")
	request := &http.Request{
		URL: address,
	}
	params := httprouter.Params{
		httprouter.Param{Key: "id", Value: "abcdef"},
	}
	request = request.WithContext(context.WithValue(context.Background(), httprouter.ParamsKey, params))
	output := benchmarkRequest{}
	for i := 0; i < b.N; i++ {
		_ = binder.Bind(request, &output)
	}
	fmt.Printf(">>>> %+v\n", output)
}

type benchmarkRequest struct {
	ID   string           `json:"id"`
	Flag bool             `json:"flaggeroo"`
	Flip benchmarkFlipper `json:"flip"`
	Page benchmarkPaging  `json:"p"`
}

type benchmarkFlipper bool

type benchmarkPaging struct {
	Limit  int
	Offset int
	Sort   string `json:"order"`
}
