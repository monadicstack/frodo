package docoptions

import "context"

/*
 * This test keeps the models simple since right now we only have doc options on services and functions. We do
 * parse the docs on models/fields, but we don't have any configuration options for them right now. Add them
 * to this test when we do. Here are some of the explicit cases this covers:
 *
 * - Being able to set every possible option on services and functions
 * - Spacing between option key and value doesn't matter
 * - Paths do not require leading slashes in the option but will end up there anyway
 * - Paths lose trailing slash automatically
 * - Options can be anywhere in the comment
 * - Leading/trailing blank comment lines are omitted (inner blank lines kept)
 * - Doc option comments don't end up in final documentation lines structure
 * - Can mix functions that do/don't have options
 * - All supported HTTP methods are accounted for
 * - Option key can have leading spaces, but not other leading characters
 * - Option order doesn't matter (can do route then status or status then route)
 */

// LebowskiService occupies various administration buildings.
// VERSION 999.12
// PREFIX  big
type LebowskiService interface {
	// Dude abides.
	//
	//
	// GET /dude/:id/
	// HTTP 202
	Dude(context.Context, *Request) (*Response, error)
	Walter(context.Context, *Request) (*Response, error)
	//
	// HTTP 204
	//
	//
	Donny(context.Context, *Request) (*Response, error)
	// HTTP 201
	// POST /dude/:id/child
	Maude(context.Context, *Request) (*Response, error)
	// PUT       /dude/jail
	Jackie(context.Context, *Request) (*Response, error)
	// Sometimes you eat the bar.
	//
	// PATCH dude/:id
	// Sometimes the bar eats you.
	Stranger(context.Context, *Request) (*Response, error)
	// RemoveToe attempts to extort $1 million.
	// DELETE /nihilist/:id/toe
	RemoveToe(context.Context, *Request) (*Response, error)
	//     HEAD /ties/room/together
	// * HTTP 202
	Rug(context.Context, *Request) (*Response, error)
}

type Request struct{}
type Response struct{}
