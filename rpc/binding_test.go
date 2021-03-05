package rpc_test

import (
	"context"
	"io"
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/dimfeld/httptreemux/v5"
	"github.com/monadicstack/frodo/rpc"
	"github.com/stretchr/testify/suite"
)

type BindingSuite struct {
	suite.Suite
}

// Ensure that binding requests w/ no values works, just leaving the 'out' value as-is.
func (suite *BindingSuite) TestEmptyRequest() {
	_, err := suite.bind(nil)
	suite.Require().Error(err, "Binding nil request should return an error")

	req := suite.newRequest("GET", noBody, noQuery, noPathParams)
	value, err := suite.bind(req)
	suite.Require().NoError(err)
	suite.Require().Equal(serviceRequest{}, value, "Should be a zero-value struct when no request values present.")

	req = suite.newRequest("POST", noBody, noQuery, noPathParams)
	value, err = suite.bind(req)
	suite.Require().NoError(err)
	suite.Require().Equal(serviceRequest{}, value, "Should be a zero-value struct when no request values present.")

	req = suite.newRequest("POST", noBody, noQuery, noPathParams)
	req.Body = nil
	value, err = suite.bind(req)
	suite.Require().NoError(err)
	suite.Require().Equal(serviceRequest{}, value, "Should be a zero-value struct when body is nil.")

	req = suite.newRequest("POST", noBody, noQuery, noPathParams)
	req.Body = http.NoBody
	value, err = suite.bind(req)
	suite.Require().NoError(err)
	suite.Require().Equal(serviceRequest{}, value, "Should be a zero-value struct when body is nil.")
}

// Ensures that we can bind all sorts of field types and recursive values using just path parameters.
func (suite *BindingSuite) TestBind_pathParameters() {
	pathParams := bindingValues{
		"String":          "foo",
		"Int":             "12345",
		"Uint":            "42",
		"Uint64":          "4200",
		"Float32":         "3.14",
		"Float64":         "99.00",
		"Bool":            "true",
		"remapped_string": "bar",
		"remapped_int":    "1",

		"AliasBasic":          "moo",
		"AliasComplex.Offset": "88",
		"AliasComplex.Text":   "goo",
		"AliasDuration":       "1m",
		"AliasYesNo":          "Yes",

		// Don't ignore the criteria fields, just the attempt to bind a struct to a single value. Make sure that
		// we can specify a param like ":Criteria.audit.created" to map a recursive value of various types/depths, too.
		"Criteria":                   "ignore",
		"Criteria.Offset":            "30",
		"Criteria.Limit":             "15",
		"Criteria.Text":              "test",
		"Criteria.audit.CreatedBy":   "Bob",
		"Criteria.audit.created":     "2020-02-20T01:02:03Z05:00",
		"Criteria.audit.CreatedDate": "1984-01-01T01:02:03Z05:00", // should ignore this one for remapped "created"

		"CriteriaPtr.Limit":                "54",
		"CriteriaPtr.audit.CreatedBy":      "Dude",
		"CriteriaPtr.AuditTrail.CreatedBy": "Bob",

		"StringSlice":   "a,b,c",
		"StringMap":     "foo",
		"StringMap.foo": "a",
		"StringMap.bar": "b",
		"StringMap.baz": "c",
		"ChanInt":       "123",
	}

	req := suite.newRequest("GET", noBody, noQuery, pathParams)
	result, err := suite.bind(req)
	suite.Require().NoError(err)
	suite.Require().Equal("foo", result.String)
	suite.Require().Equal(12345, result.Int)
	suite.Require().Equal(uint(42), result.Uint)
	suite.Require().Equal(uint64(4200), result.Uint64)
	suite.Require().Equal(float32(3.14), result.Float32)
	suite.Require().Equal(true, result.Bool)
	suite.Require().Equal("bar", result.RemappedString)
	suite.Require().Equal(1, result.RemappedInt)
	suite.Require().Equal(30, result.Criteria.Offset)
	suite.Require().Equal(15, result.Criteria.Limit)
	suite.Require().Equal("test", result.Criteria.Text)
	suite.Require().Equal("Bob", result.Criteria.AuditTrail.CreatedBy)
	suite.Require().Equal(parseTime("2020-02-20T01:02:03Z05:00"), result.Criteria.AuditTrail.CreatedDate)

	suite.Require().Equal(aliasBasic("moo"), result.AliasBasic)
	suite.Require().Equal(88, result.AliasComplex.Offset)
	suite.Require().Equal("goo", result.AliasComplex.Text)
	suite.Require().Equal(aliasDuration(1*time.Minute), result.AliasDuration)
	suite.Require().Equal(aliasYesNo(true), result.AliasYesNo)

	suite.Require().Equal(54, result.CriteriaPtr.Limit)
	suite.Require().Equal(0, result.CriteriaPtr.Offset)
	suite.Require().Equal("Dude", result.CriteriaPtr.AuditTrail.CreatedBy)

	// Types we know we don't have support for yet.
	suite.Require().Nil(result.StringSlice)
	suite.Require().Nil(result.StringMap)
	suite.Require().Nil(result.ChanInt)
}

// Ensures that we can bind all sorts of field types and recursive values using just query string parameters.
func (suite *BindingSuite) TestBind_queryString() {
	queryString := bindingValues{
		"String":          "foo",
		"Int":             "12345",
		"Uint":            "42",
		"Uint64":          "4200",
		"Float32":         "3.14",
		"Float64":         "99.00",
		"Bool":            "true",
		"remapped_string": "bar",
		"remapped_int":    "1",

		"AliasBasic":          "moo",
		"AliasComplex.Offset": "88",
		"AliasComplex.Text":   "goo",
		"AliasDuration":       "1m",
		"AliasYesNo":          "Yes",

		// Don't ignore the criteria fields, just the attempt to bind a struct to a single value. Make sure that
		// we can specify a param like ":Criteria.audit.created" to map a recursive value of various types/depths, too.
		"Criteria":                   "ignore",
		"Criteria.Offset":            "30",
		"Criteria.Limit":             "15",
		"Criteria.Text":              "test",
		"Criteria.audit.CreatedBy":   "Bob",
		"Criteria.audit.created":     "2020-02-20T01:02:03Z05:00",
		"Criteria.audit.CreatedDate": "1984-01-01T01:02:03Z05:00", // should ignore this one for remapped "created"

		"CriteriaPtr.Limit":                "54",
		"CriteriaPtr.audit.CreatedBy":      "Dude",
		"CriteriaPtr.AuditTrail.CreatedBy": "Bob",

		"StringSlice":   "a,b,c",
		"StringMap":     "foo",
		"StringMap.foo": "a",
		"StringMap.bar": "b",
		"StringMap.baz": "c",
		"ChanInt":       "123",
	}

	req := suite.newRequest("GET", noBody, queryString, noPathParams)
	result, err := suite.bind(req)
	suite.Require().NoError(err)
	suite.Require().Equal("foo", result.String)
	suite.Require().Equal(12345, result.Int)
	suite.Require().Equal(uint(42), result.Uint)
	suite.Require().Equal(uint64(4200), result.Uint64)
	suite.Require().Equal(float32(3.14), result.Float32)
	suite.Require().Equal(true, result.Bool)
	suite.Require().Equal("bar", result.RemappedString)
	suite.Require().Equal(1, result.RemappedInt)
	suite.Require().Equal(30, result.Criteria.Offset)
	suite.Require().Equal(15, result.Criteria.Limit)
	suite.Require().Equal("test", result.Criteria.Text)
	suite.Require().Equal("Bob", result.Criteria.AuditTrail.CreatedBy)
	suite.Require().Equal(parseTime("2020-02-20T01:02:03Z05:00"), result.Criteria.AuditTrail.CreatedDate)

	suite.Require().Equal(aliasBasic("moo"), result.AliasBasic)
	suite.Require().Equal(88, result.AliasComplex.Offset)
	suite.Require().Equal("goo", result.AliasComplex.Text)
	suite.Require().Equal(aliasDuration(1*time.Minute), result.AliasDuration)
	suite.Require().Equal(aliasYesNo(true), result.AliasYesNo)

	suite.Require().Equal(54, result.CriteriaPtr.Limit)
	suite.Require().Equal(0, result.CriteriaPtr.Offset)
	suite.Require().Equal("Dude", result.CriteriaPtr.AuditTrail.CreatedBy)

	// Types we know we don't have support for yet.
	suite.Require().Nil(result.StringSlice)
	suite.Require().Nil(result.StringMap)
	suite.Require().Nil(result.ChanInt)
}

// Ensures that we can bind body params and that it only happens on methods that support it (PUT/POST/PATCH)
func (suite *BindingSuite) TestBind_body() {
	req := suite.newRequest("GET", `{"String": "foo", "Int": 12345}`, noQuery, noPathParams)
	result, err := suite.bind(req)
	suite.NoError(err, "GET request: should ignore body, not fail when present")
	suite.Equal(serviceRequest{}, result, "GET request: should not bind body")

	req = suite.newRequest("DELETE", `{"String": "foo", "Int": 12345}`, noQuery, noPathParams)
	result, err = suite.bind(req)
	suite.NoError(err, "DELETE request: should ignore body, not fail when present")
	suite.Equal(serviceRequest{}, result, "DELETE request: should not bind body")

	req = suite.newRequest("PUT", `{"String": "foo", "Int": 12345}`, noQuery, noPathParams)
	result, err = suite.bind(req)
	suite.NoError(err)
	suite.Equal("foo", result.String, "PUT request: body should bind JSON properly")
	suite.Equal(12345, result.Int, "PUT request: body should bind JSON properly")

	req = suite.newRequest("POST", `{"String": "foo", "Int": 12345}`, noQuery, noPathParams)
	result, err = suite.bind(req)
	suite.NoError(err)
	suite.Equal("foo", result.String, "POST request: body should bind JSON properly")
	suite.Equal(12345, result.Int, "POST request: body should bind JSON properly")

	req = suite.newRequest("PATCH", `{"String": "foo", "Int": 12345}`, noQuery, noPathParams)
	result, err = suite.bind(req)
	suite.NoError(err)
	suite.Equal("foo", result.String, "PATCH request: body should bind JSON properly")
	suite.Equal(12345, result.Int, "PATCH request: body should bind JSON properly")
}

// Make sure that if the same value appears in the body, query string, and path that path always wins, body is next,
// and query string is least "sticky" of the values.
func (suite *BindingSuite) TestBind_bindingOrder() {
	/*
	 * This test is set up so that path should win for String and Int, body should win for Uint64, and
	 * query should win for Float32.
	 */
	queryValues := bindingValues{
		"String":  "A",
		"Int":     "1",
		"Uint64":  "100",
		"Float32": "200.1",
	}
	body := `{
        "String": "B",
        "Int":    2,
		"Uint64": 101
    }`
	pathValues := bindingValues{
		"String": "C",
		"Int":    "3",
	}
	req := suite.newRequest("PATCH", body, queryValues, pathValues)
	result, err := suite.bind(req)
	suite.NoError(err)
	suite.Equal("C", result.String, "Path param should override query string and body.")
	suite.Equal(3, result.Int, "Path param should override query string and body.")
	suite.Equal(uint64(101), result.Uint64, "Body should override query string.")
	suite.Equal(float32(200.1), result.Float32, "Query should be bound when not present in body or path")
}

// Make sure we're throwing errors on all manner of bad data.
func (suite *BindingSuite) TestBind_errors() {
	// Invalid JSON in body (value not quoted for string
	req := suite.newRequest("POST", `{"String": fart}`, noQuery, noPathParams)
	_, err := suite.bind(req)
	suite.Error(err, "Should return an error when body is invalid JSON")

	// Invalid format for query param
	badValues := bindingValues{"Int": "Fail"}
	req = suite.newRequest("POST", "", badValues, noPathParams)
	_, err = suite.bind(req)
	suite.Error(err, "Should return an error when query string value format is incorrect")

	// Invalid format for path param
	req = suite.newRequest("POST", "", noQuery, badValues)
	_, err = suite.bind(req)
	suite.Error(err, "Should return an error when query string value format is incorrect")

	req = suite.newRequest("POST", "", noQuery, badValues)
	req.URL = nil
	_, err = suite.bind(req)
	suite.Error(err, "Should return an error when URL is nil")
}

// Ensures that we can use functional options to set the binder when setting up a gateway.
func (suite *BindingSuite) TestWithBinder() {
	gateway := rpc.NewGateway(rpc.WithBinder(nil))
	suite.NotNil(gateway.Binder, "Should not be allowed to set binder to nil")

	gateway = rpc.NewGateway(rpc.WithBinder(bindingValues{}))
	suite.EqualValues(bindingValues{}, gateway.Binder)
}

// Creates the default binder and binds a 'serviceRequest' with the given request data.
func (suite *BindingSuite) bind(req *http.Request) (serviceRequest, error) {
	value := serviceRequest{}
	err := rpc.NewGateway().Binder.Bind(req, &value)
	return value, err
}

// Creates a new HTTP request with just the handful of request fields filled in that we actually use
// in the binding process.
func (suite *BindingSuite) newRequest(method string, body string, queryValues bindingValues, pathValues bindingValues) *http.Request {
	queryString := queryValues.ToQueryString()

	// The binding process only grabs path params from the context/router, so don't worry about having
	// a meaningful path/endpoint with the data.
	path := "https://foo.services:9000/Some.Function"
	if len(queryString) > 0 {
		path += "?" + queryString.Encode()
	}
	uri, _ := url.Parse(path)

	var bodyReader io.ReadCloser = http.NoBody
	if body != "" {
		bodyReader = io.NopCloser(strings.NewReader(body))
	}

	req := &http.Request{
		Method: method,
		URL:    uri,
		Body:   bodyReader,
	}
	ctx := httptreemux.AddRouteDataToContext(context.Background(), mockRouteData{
		route:  "Some.Function",
		params: pathValues,
	})
	return req.WithContext(ctx)
}

func parseTime(value string) time.Time {
	t, _ := time.Parse(time.RFC3339, value)
	return t
}

// serviceRequest contains all of the different types of fields we want to try binding.
type serviceRequest struct {
	String    string
	StringPtr *string
	Int       int
	Uint      uint
	Uint64    uint64
	Float32   float32
	Float64   float64
	Bool      bool

	AliasBasic    aliasBasic
	AliasComplex  aliasComplex
	AliasDuration aliasDuration
	AliasYesNo    aliasYesNo

	RemappedString string `json:"remapped_string"`
	RemappedInt    int    `json:"remapped_int"`

	Criteria    searchCriteria
	CriteriaPtr *searchCriteria

	// These are types the binder doesn't have support for yet, but
	// include explicit test cases for them so that's known/documented
	// behavior until we address them.

	StringSlice []string
	StringMap   map[string]string
	ChanInt     chan int
}

type searchCriteria struct {
	Limit      int
	Offset     int
	Text       string
	AuditTrail auditTrail `json:"audit"`
}

type auditTrail struct {
	CreatedBy   string
	CreatedDate time.Time `json:"created"`
}

type aliasBasic string
type aliasComplex searchCriteria
type aliasDuration time.Duration
type aliasYesNo bool

func (a *aliasDuration) UnmarshalJSON(data []byte) error {
	isoValue := strings.Trim(string(data), `"`)
	duration, err := time.ParseDuration(isoValue)
	if err != nil {
		return err
	}
	*a = aliasDuration(duration)
	return nil
}

func (a *aliasYesNo) UnmarshalJSON(data []byte) error {
	value := strings.ToLower(strings.Trim(string(data), `"`))
	*a = value == "yes"
	return nil
}

type bindingValues map[string]string

var noQuery = bindingValues{}
var noPathParams = bindingValues{}
var noBody = ""

// Bind does nothing here. It's only used when testing WithBinder so
// that we have something other than the default binder to test against.
func (values bindingValues) Bind(*http.Request, interface{}) error {
	return nil
}

func (values bindingValues) ToQueryString() url.Values {
	queryString := url.Values{}
	for key, value := range values {
		queryString.Set(key, value)
	}
	return queryString
}

func TestBindingSuite(t *testing.T) {
	suite.Run(t, new(BindingSuite))
}
