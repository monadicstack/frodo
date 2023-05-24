// Package metadata provides request-scoped values for your RPC calls. You can use the
// standard context.Context package to store values for the request, true, but those will
// not follow you when you make an RPC call to another service. Metadata values *do* follow
// you as you hop from service to service, so they're ideal for trace data, identity data, etc.
//
// The API for interacting with metadata values is similar to dealing w/ context values with the
// small modification that when you fetch a value, it accepts an 'out' parameter (like json.Unmarshal)
// instead of returning an interface{} value. This is a necessary evil due to limitations with
// Go's type system and reflection, but should only result in one extra line of code when fetching
// values from the metadata/context.
package metadata

import (
	"bytes"
	"context"
	"encoding/json"
	"log"

	"github.com/monadicstack/frodo/internal/reflection"
)

// The context value entry for our metadata map.
type contextKey struct{}

// RequestHeader is the custom HTTP header that we use to encode your metadata values as
// you make RPC calls from service A to service B. Service B will decode this header to
// restore all of the metadata that you had while code was still running in service A.
const RequestHeader = "X-RPC-Values"

// Values provides a lookup for all of the request-scoped data that you want to follow
// you as you make RPC calls from service to service (or client to service).
type Values map[string]valuesEntry

// valuesEntry is a single value in our metadata "Values" map. It tracks two different representations
// of the value. The Value is the actual Go primitive/struct that you're storing. The JSON field is the
// marshaled version of Value that we use when sending this value over the wire.
//
// The idea is that the entire Values map is converted to JSON and sent as an X-RPC-Values header to the
// service we're invoking. When that service receives the header, it will keep the raw JSON for a while
// because it doesn't have any type information about the values' underlying types - this mainly due to the fact
// that Go reflection does not support looking up type information given a package and type name like
// many other languages. For example even if we encoded that the JSON was the type "Baz" from the
// package "github.com/foo/bar", we wouldn't be able to get a reflect.Type to actually construct a new Baz{}
// when we needed it.
//
// Due to that limitation, the receiving service won't actually know the types of the values until some
// code looks one of the values up and feeds an empty instance into the 'out' parameter of the Value() call.
// At that point, we'll have a strongly typed Go value in hand, so we'll perform some lazy
// unmarshaling of the JSON back into the Value. Now you'll have the real value available for the duration
// of the call.
type valuesEntry struct {
	// JSON is the representation of the value we'll use temporarily when sending the metadata from
	// service A to service B. It will be unmarshaled back into Value once we have type information later.
	JSON string `json:"value"`
	// Value is the actual Go value for this piece of RPC metadata.
	Value interface{}
}

// MarshalJSON encodes the entry as a JSON object that encodes the value so that it can be embedded
// in an X-RPC-Values header.
func (v valuesEntry) MarshalJSON() ([]byte, error) {
	buf := &bytes.Buffer{}
	buf.WriteString(`{"value":`)
	err := json.NewEncoder(buf).Encode(v.Value)
	if err != nil {
		return nil, err
	}
	buf.WriteString(`}`)
	return buf.Bytes(), nil
}

// UnmarshalJSON receives the JSON for a single metadata entry and creates it. Since we do
// not have type information yet (see docs for valuesEntry above) we can't actually unmarshal
// the raw value. We'll just strip off the right-hand-side of the JSON attribute since that's
// the marshaled raw value. The first time that someone calls metadata.Value() for this key we
// will unmarshal the value for reals.
func (v *valuesEntry) UnmarshalJSON(data []byte) error {
	length := len(data)
	if length < 11 {
		return nil
	}
	v.JSON = string(data[9 : length-1])
	return nil
}

// unmarshal uses the type information from 'out' to reconstitute the entry's JSON into a real value.
func (v *valuesEntry) unmarshal(out interface{}) error {
	err := json.Unmarshal([]byte(v.JSON), out)
	if err != nil {
		return err
	}
	v.Value = out
	v.JSON = ""
	return nil
}

// Value looks up a single piece of metadata on the specified context. The 'key' is the name of
// the value you're looking for and 'out' is a pointer to the value you want us to fill in - the
// mechanics are similar to json.Unmarshal().
func Value(ctx context.Context, key string, out interface{}) bool {
	if ctx == nil {
		return false
	}
	if key == "" {
		return false
	}

	// Make sure that we even *have* scope values on the context first.
	scope, ok := ctx.Value(contextKey{}).(Values)
	if !ok {
		return false
	}

	// We have a scope but nothing for this key
	entry, ok := scope[key]
	if !ok {
		return false
	}

	// We have already reconstituted the raw value from the header json, so just assign
	// the value to the "out" pointer.
	if entry.Value != nil {
		return reflection.Assign(entry.Value, out)
	}

	// You are likely on the server side and are attempting to access a value for the first time,
	// so we need to unmarshal the JSON to get the value to the caller.
	err := entry.unmarshal(out)
	if err != nil {
		log.Printf("error: unmarshal rpc value '%s': %v", key, err)
		return false
	}
	return true
}

// WithValue stores a key/value pair in the context metadata. It returns a new context that contains
// the metadata map with your value.
func WithValue(ctx context.Context, key string, value interface{}) context.Context {
	if ctx == nil {
		return ctx
	}
	if key == "" {
		return ctx
	}
	meta, ok := ctx.Value(contextKey{}).(Values)
	if !ok {
		meta = Values{}
		ctx = WithValues(ctx, meta)
	}

	// At some point I would love for this to be not rely on side-effects and mutating a map. I'd
	// rather that adding new values create copies of the metadata structure so it's more thread
	// safe like normal context values. But this will work for now...
	if meta != nil {
		meta[key] = valuesEntry{Value: value}
	}
	return ctx
}

// WithValues does a wholesale replacement of ALL metadata values stored on the context. Usually you
// will not call this yourself - you should interact with individual values via Value()/WithValue(). This
// is typically just used by frodo internals to preserve your metadata across RPC calls.
func WithValues(ctx context.Context, meta Values) context.Context {
	return context.WithValue(ctx, contextKey{}, meta)
}

// ToJSON serializes all of the metadata values into a single JSON string. You won't usually call this
// yourself as this is mainly used to set the X-RPC-Values header when making RPC calls to preserve your
// data across RPC calls.
func ToJSON(ctx context.Context) (string, error) {
	meta, _ := ctx.Value(contextKey{}).(Values)
	metaJSON, err := json.Marshal(meta)
	return string(metaJSON), err
}

// FromJSON rebuilds a Values map from the JSON embedded in an X-RPC-Values header. While it does
// make the Values map, all of the individual value entries will still be in their json formats rather
// than their unmarshaled raw Go values until they're explicitly used. See docs for Values for more info.
func FromJSON(rpcValuesHeader string) (Values, error) {
	if rpcValuesHeader == "" {
		return Values{}, nil
	}

	meta := Values{}
	err := json.Unmarshal([]byte(rpcValuesHeader), &meta)
	return meta, err
}
