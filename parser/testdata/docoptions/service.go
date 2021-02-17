package docoptions

import "context"

/*
 * This test keeps the models simple since right now we only have doc options on services and functions. We do
 * parse the docs on models/fields, but we don't have any configuration options for them right now. Add them
 * to this test when we do.
 */

// LebowskiService occupies various administration buildings.
// VERSION 999.12
// PREFIX  big
type LebowskiService interface {
	// Dude abides.
	//
	//
	// GET /dude/:id
	// HTTP 202
	Dude(context.Context, *Request) (*Response, error)
	Walter(context.Context, *Request) (*Response, error)
	//
	// HTTP 204
	//
	//
	Donnie(context.Context, *Request) (*Response, error)
	// HTTP 201
	// POST /dude/:id/child
	Maude(context.Context, *Request) (*Response, error)
	// PUT /dude/jail
	Jackie(context.Context, *Request) (*Response, error)
	// Sometimes you eat the bar.
	// Sometimes the bar eats you.
	//
	// PATCH /dude/:id
	Stranger(context.Context, *Request) (*Response, error)
	// RemoveToe attempts to extort $1 million.
	// DELETE /nihilist/:id/toe
	RemoveToe(context.Context, *Request) (*Response, error)
}

type Request struct{}
type Response struct{}
