//go:build unit
// +build unit

package rpc_test

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/davidrenne/frodo/rpc"
	"github.com/davidrenne/frodo/rpc/authorization"
	"github.com/davidrenne/frodo/rpc/metadata"
	"github.com/stretchr/testify/suite"
)

type ClientSuite struct {
	suite.Suite
}

// Ensure that the default RPC client is valid.
func (suite *ClientSuite) TestNewClient_default() {
	client := rpc.NewClient("FooService", ":9000")
	suite.Require().Equal("FooService", client.Name)
	suite.Require().Equal(":9000", client.BaseURL)
	suite.Require().Equal("", client.PathPrefix)
	suite.Require().NotNil(client.HTTP, "Default HTTP client should be non-nil")
	suite.Require().Equal(30*time.Second, client.HTTP.Timeout, "Default HTTP client timeout should be 30 seconds")

	// Yes, you can leave these blank. You'll have a bad time... but you can do it.
	client = rpc.NewClient("", "")
	suite.Require().Equal("", client.Name)
	suite.Require().Equal("", client.BaseURL)

	client = rpc.NewClient("", "http://foo:9000/trailing/slash/")
	suite.Require().Equal("http://foo:9000/trailing/slash", client.BaseURL, "Should trim trailing slashes in base URL")
}

// Ensure that functional options override client defaults.
func (suite *ClientSuite) TestNewClient_options() {
	client := rpc.NewClient("FooService", ":9000",
		func(c *rpc.Client) { c.Name = "FartService" },
		func(c *rpc.Client) { c.BaseURL = "https://google.com" },
		func(c *rpc.Client) { c.PathPrefix = "/v2/" },
		func(c *rpc.Client) { c.HTTP = nil },
	)
	suite.Require().Equal("FartService", client.Name)
	suite.Require().Equal("https://google.com", client.BaseURL)
	suite.Require().Equal("/v2/", client.PathPrefix)
	suite.Require().Nil(client.HTTP)

	httpClient := &http.Client{}
	client = rpc.NewClient("FooService", ":9000", rpc.WithHTTPClient(httpClient))
	suite.Require().Same(httpClient, client.HTTP, "WithHTTPClient should set the client's HTTP client")
}

// Ensures that an RPC client can invoke an HTTP GET endpoint. All of the service request values should
// be set on the query string.
func (suite *ClientSuite) TestInvoke_get() {
	client := suite.newClient(func(r *http.Request) (*http.Response, error) {
		suite.assertURL(r, "http://localhost:9000/foo")
		suite.assertQuery(r, url.Values{
			"ID":         []string{"123"},
			"Int":        []string{"42"},
			"Inner.Flag": []string{"false"}, // even includes default values for fields not explicitly set
			"Inner.Skip": []string{"100"},
		})
		return suite.respond(200, &clientResponse{ID: "Bob", Name: "Loblaw"})
	})

	in := &clientRequest{ID: "123", Int: 42, Inner: clientInner{Skip: 100}}
	out := &clientResponse{}
	err := client.Invoke(context.Background(), "GET", "/foo", in, out)
	suite.Require().NoError(err)
	suite.Require().Equal("Bob", out.ID)
	suite.Require().Equal("Loblaw", out.Name)
}

// Ensures that an RPC client can invoke an HTTP POST endpoint. All of the service request values should
// be set on the body, not the query string.
func (suite *ClientSuite) TestInvoke_post() {
	client := suite.newClient(func(r *http.Request) (*http.Response, error) {
		suite.Require().Len(r.URL.Query(), 0, "Client.Invoke() - POST should not have a query string")
		suite.assertURL(r, "http://localhost:9000/foo")

		actual, err := suite.unmarshal(r)
		suite.NoError(err, "Client.Invoke() - POST should not send junk JSON")
		suite.Require().Equal("123", actual.ID)
		suite.Require().Equal(42, actual.Int)
		suite.Require().Equal(100, actual.Inner.Skip)
		suite.Require().Equal(false, actual.Inner.Flag)

		return suite.respond(200, &clientResponse{ID: "Bob", Name: "Loblaw"})
	})

	in := &clientRequest{ID: "123", Int: 42, Inner: clientInner{Skip: 100}}
	out := &clientResponse{}
	err := client.Invoke(context.Background(), "POST", "/foo", in, out)
	suite.Require().NoError(err)
	suite.Require().Equal("Bob", out.ID)
	suite.Require().Equal("Loblaw", out.Name)
}

// Ensures that an RPC client fills in path params (e.g. "/:id"->"/1234"). We will make sure
// that path param substitutions:
//
// * Param name matches request attribute exactly
// * Param name can be case insensitive (e.g. ":id" should match field "ID")
// * Support nested params (e.g. ":Criteria.Paging.Limit")
// * Should ignore param names that don't match anything
// * If you include a request attribute in the path, do NOT include it in the query string, too.
func (suite *ClientSuite) TestInvoke_pathParams() {
	client := suite.newClient(func(r *http.Request) (*http.Response, error) {
		suite.assertURL(r, "http://localhost:9000/foo/123/42//100") // missing :bar is between 42 and 100

		// Make sure that the path params are not in the query string, too. Values not in the path should
		// still be in there, however.
		query := r.URL.Query()
		suite.Require().Empty(query.Get("ID"), "Should not be in the query string when field is in the path")
		suite.Require().Empty(query.Get("Int"), "Should not be in the query string when field is in the path")
		suite.Require().Empty(query.Get("Inner.Skip"), "Should not be in the query string when field is in the path")
		suite.Require().Equal("true", query.Get("Inner.Flag"))
		suite.Require().Equal("", query.Get("Inner.Test"))

		return suite.respond(200, &clientResponse{ID: "Bob", Name: "Loblaw"})
	})

	in := &clientRequest{ID: "123", Int: 42, Inner: clientInner{Skip: 100, Flag: true}}
	out := &clientResponse{}
	err := client.Invoke(context.Background(), "GET", "/foo/:id/:Int/:bar/:Inner.Skip/:Nope", in, out)
	suite.Require().NoError(err)
	suite.Require().Equal("Bob", out.ID)
	suite.Require().Equal("Loblaw", out.Name)
}

// Ensures that an RPC client will translate 4XX/5XX errors into the
// equivalent status-coded error.
func (suite *ClientSuite) TestInvoke_httpStatusError() {
	client := suite.newClient(func(r *http.Request) (*http.Response, error) {
		typeText := http.Header{"Content-Type": []string{"text/plain"}}
		typeJSON := http.Header{"Content-Type": []string{"application/json"}}
		switch r.URL.Path {
		case "/404":
			// A non-json response where the body is the error message
			body := "not here, dude"
			return &http.Response{StatusCode: 404, Header: typeText, Body: io.NopCloser(strings.NewReader(body))}, nil
		case "/409":
			body := `"already did that"`
			return &http.Response{StatusCode: 409, Header: typeJSON, Body: io.NopCloser(strings.NewReader(body))}, nil
		case "/500":
			body := `{"message": "broke as hell"}`
			return &http.Response{StatusCode: 500, Header: typeJSON, Body: io.NopCloser(strings.NewReader(body))}, nil
		case "/504":
			body := `{"foo": "broke as hell"}`
			return &http.Response{StatusCode: 504, Header: typeJSON, Body: io.NopCloser(strings.NewReader(body))}, nil
		}
		panic("how did you get here?")
	})

	out := &clientResponse{}
	err := client.Invoke(context.Background(), "POST", "/404", &clientRequest{}, out)
	suite.Require().Error(err, "Client.Invoke() - 404 status code should return an error")
	suite.Require().Contains(err.Error(), "not here, dude", "Client.Invoke() - should include plain text message")

	out = &clientResponse{}
	err = client.Invoke(context.Background(), "POST", "/409", &clientRequest{}, out)
	suite.Require().Error(err, "Client.Invoke() - 409 status code should return an RPC error")
	suite.Require().Contains(err.Error(), "already did that", "Client.Invoke() - should include json string message")

	// A json response where the body is the Responder error struct (i.e. message is a json attribute)
	out = &clientResponse{}
	err = client.Invoke(context.Background(), "POST", "/500", &clientRequest{}, out)
	suite.Require().Error(err, "Client.Invoke() - 500 status code should return an error")
	suite.Require().Contains(err.Error(), "broke as hell", "Client.Invoke() - should include json error struct message")

	// A json response but the body doesn't look like our normal JSON error structure.
	out = &clientResponse{}
	err = client.Invoke(context.Background(), "POST", "/504", &clientRequest{}, out)
	suite.Require().Error(err, "Client.Invoke() - 504 status code should return an error")
	suite.Require().NotContains(err.Error(), "broke as hell", "Client.Invoke() - not include unknown error message formats")
}

// Check all of the different ways that Invoke() can fail.
func (suite *ClientSuite) TestInvoke_roundTripError() {
	client := suite.newClient(func(r *http.Request) (*http.Response, error) {
		switch r.URL.Path {
		case "/ok":
			return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader(`{asdf}`))}, nil
		default:
			return nil, fmt.Errorf("wtf")
		}
	})

	// We failed trying to marshal the service request value as JSON
	err := client.Invoke(context.Background(), "POST", "/fail", &unableToMarshal{}, &clientResponse{})
	suite.Require().Error(err, "Client.Invoke() should return an error when input can't be marshaled")
	suite.Require().NotContains(err.Error(), "wtf", "Client.Invoke() error should not get to handler if failed during setup")

	// We failed creating the request to dispatch (bad http method)
	err = client.Invoke(context.Background(), "🍺", "/fail", &clientRequest{}, &clientResponse{})
	suite.Require().Error(err, "Client.Invoke() should return an error when request can't be constructed")
	suite.Require().NotContains(err.Error(), "wtf", "Client.Invoke() error should not get to handler if failed during setup")

	// Dispatch went ok, but the round-tripper function returned an error
	err = client.Invoke(context.Background(), "POST", "/ok", &clientRequest{}, &clientResponse{})
	suite.Require().Error(err, "Client.Invoke() should return an error when round tripper returns an error")
	suite.Require().NotContains(err.Error(), "wtf", "Client.Invoke() error propagate error returned by round tripper")
}

// Should invoke your middleware in the correct order before dispatching the "real" handler.
func (suite *ClientSuite) TestWithMiddleware() {
	values := []string{"", "", ""}
	client := rpc.NewClient("FooService", "http://localhost:9000", rpc.WithClientMiddleware(
		func(request *http.Request, next rpc.RoundTripperFunc) (*http.Response, error) {
			values[0] = "A"
			return next(request)
		},
		func(request *http.Request, next rpc.RoundTripperFunc) (*http.Response, error) {
			values[1] = "B"
			return next(request)
		},
	))
	client.HTTP.Transport = rpc.RoundTripperFunc(func(r *http.Request) (*http.Response, error) {
		values[2] = "C"
		return suite.respond(200, &clientResponse{ID: "123"})
	})

	out := &clientResponse{}
	err := client.Invoke(context.Background(), "POST", "/foo", &clientRequest{}, out)
	suite.Require().NoError(err)
	suite.Require().Equal("123", out.ID)
	suite.Require().Equal("A", values[0])
	suite.Require().Equal("B", values[1])
	suite.Require().Equal("C", values[2])
}

func (suite *ClientSuite) TestInvoke_includeHeaders() {
	client := suite.newClient(func(r *http.Request) (*http.Response, error) {
		suite.Require().Equal("Hello", r.Header.Get("Authorization"))
		suite.Require().Equal(`{"Foo":{"value":"Bar"}}`, r.Header.Get(metadata.RequestHeader))
		return suite.respond(200, &clientResponse{ID: "123"})
	})

	ctx := context.Background()
	ctx = authorization.WithHeader(ctx, authorization.New("Hello"))
	ctx = metadata.WithValue(ctx, "Foo", "Bar")
	err := client.Invoke(ctx, "POST", "/foo", &clientRequest{}, &clientResponse{})
	suite.Require().NoError(err)
}

func (suite *ClientSuite) newClient(roundTripper rpc.RoundTripperFunc) rpc.Client {
	client := rpc.NewClient("Test", "http://localhost:9000")
	client.HTTP.Transport = roundTripper
	return client
}

func (suite *ClientSuite) assertURL(r *http.Request, expected string) {
	actual := r.URL.String()
	questionIndex := strings.Index(actual, "?")
	if questionIndex >= 0 {
		actual = actual[0:questionIndex]
	}
	suite.Require().Equal(expected, actual, "Client.Invoke() - Incorrect URL/path")
}

func (suite *ClientSuite) assertQuery(r *http.Request, expectedValues url.Values) {
	actual := r.URL.Query()
	for key, expected := range expectedValues {
		suite.Require().ElementsMatch(expected, actual[key], "Client.Invoke() - wrong query string value for %s", key)
	}
}

func (suite *ClientSuite) respond(status int, body *clientResponse) (*http.Response, error) {
	jsonBytes, _ := json.Marshal(body)
	jsonString := strings.NewReader(string(jsonBytes))
	return &http.Response{StatusCode: status, Body: io.NopCloser(jsonString)}, nil
}

func (suite *ClientSuite) unmarshal(r *http.Request) (*clientRequest, error) {
	defer r.Body.Close()
	out := &clientRequest{}
	return out, json.NewDecoder(r.Body).Decode(out)
}

type clientRequest struct {
	ID       string
	Int      int
	Inner    clientInner
	InnerPtr *clientInner
}

type clientInner struct {
	Test string
	Flag bool
	Skip int
}

type clientResponse struct {
	ID   string
	Name string
}

type unableToMarshal struct {
	Channel chan string
}

func TestClientSuite(t *testing.T) {
	suite.Run(t, new(ClientSuite))
}
