package rpc

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/robsignorelli/expose/internal/reflection"
)

func NewClient(name string, addr string, options ...Option) Client {
	defaultTimeout := 30 * time.Second
	client := Client{
		HTTP: &http.Client{
			Transport: &http.Transport{
				DialContext:         (&net.Dialer{Timeout: defaultTimeout}).DialContext,
				TLSHandshakeTimeout: defaultTimeout,
			},
		},
		Name:    name,
		BaseURL: strings.TrimSuffix(addr, "/"),
	}
	for _, option := range options {
		option(&client)
	}
	return client
}

type Option func(*Client)

func WithHTTPClient(httpClient *http.Client) Option {
	return func(rpcClient *Client) {
		rpcClient.HTTP = httpClient
	}
}

type Client struct {
	HTTP    *http.Client
	BaseURL string
	Name    string
}

func (c Client) Invoke(ctx context.Context, method string, path string, serviceRequest interface{}, serviceResponse interface{}) error {
	// Step 1: Fill in the URL path and query string w/ fields from the request. (e.g. /user/:id -> /user/abc)
	address := c.buildURL(method, path, serviceRequest)
	log.Printf("Invoking %s -> %s %s", c.Name, method, address)

	// Step 2: Create a JSON reader for the request body (POST/PUT/PATCH only).
	body, err := c.createRequestBody(method, serviceRequest)
	if err != nil {
		return fmt.Errorf("rpc: unable to create request body: %w", err)
	}

	// Step 3: Form the HTTP request
	request, err := http.NewRequest(method, address, body)
	if err != nil {
		return fmt.Errorf("rpc: unable to create request: %w", err)
	}

	// Step 4: Write the HTTP headers (content type, metadata, etc)
	request, err = c.writeHeaders(ctx, request)
	if err != nil {
		return fmt.Errorf("rpc: unable to write headers: %w", err)
	}

	// Step 5: Dispatch the HTTP request to the other service.
	response, err := c.HTTP.Do(request)
	if err != nil {
		return fmt.Errorf("rpc: unable to dispatch request: %w", err)
	}
	defer response.Body.Close()

	// Step 6: Based on the status code, either fill in the "out" struct (service response) with the
	// unmarshaled JSON or respond a properly formed error.
	if response.StatusCode >= 400 {
		errData, _ := ioutil.ReadAll(response.Body)
		return fmt.Errorf("rpc: http status %v: %s", response.StatusCode, string(errData))
	}

	err = json.NewDecoder(response.Body).Decode(serviceResponse)
	if err != nil {
		return fmt.Errorf("rpc: unable to decode response: %w", err)
	}
	return nil
}

func (c Client) createRequestBody(method string, serviceRequest interface{}) (io.Reader, error) {
	if shouldEncodeUsingQueryString(method) {
		return nil, nil
	}
	body := &bytes.Buffer{}
	err := json.NewEncoder(body).Encode(serviceRequest)
	return body, err
}

func (c Client) writeHeaders(ctx context.Context, request *http.Request) (*http.Request, error) {
	return request, nil
}

func (c Client) buildURL(method string, path string, serviceRequest interface{}) string {
	attributes := reflection.StructToMap(serviceRequest)

	path = strings.TrimPrefix(path, "/")
	path = strings.TrimSuffix(path, "/")
	pathSegments := strings.Split(path, "/")

	for i, pathSegment := range pathSegments {
		if strings.HasPrefix(pathSegment, ":") {
			paramName := pathSegment[1:]
			pathSegments[i] = fmt.Sprintf("%v", attributes[paramName])
			delete(attributes, paramName) // so it doesn't also get encoded in the query string
		}
	}

	// If we're doing a POST/PUT/PATCH, don't bother adding query string arguments.
	address := c.BaseURL + "/" + strings.Join(pathSegments, "/")
	if shouldEncodeUsingBody(method) {
		return address
	}

	// We're doing a GET/DELETE/etc, so all request values must come via query string args
	queryString := url.Values{}
	for name, value := range attributes {
		queryString.Set(name, fmt.Sprintf("%v", value))
	}
	return address + "?" + queryString.Encode()
}

func shouldEncodeUsingBody(method string) bool {
	return method == http.MethodPost || method == http.MethodPut || method == http.MethodPatch
}

func shouldEncodeUsingQueryString(method string) bool {
	return !shouldEncodeUsingBody(method)
}
