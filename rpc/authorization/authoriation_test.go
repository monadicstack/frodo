//go:build unit
// +build unit

package authorization_test

import (
	"context"
	"testing"

	"github.com/davidrenne/frodo/rpc/authorization"
	"github.com/stretchr/testify/suite"
)

type HeaderSuite struct {
	suite.Suite
}

func (suite *HeaderSuite) TestHeader_String() {
	header := authorization.New("")
	suite.Require().Equal("", header.String())

	header = authorization.New("  \n \t")
	suite.Require().Equal("", header.String())

	header = authorization.New("  \n x\t")
	suite.Require().Equal("x", header.String())

	header = authorization.New("x")
	suite.Require().Equal("x", header.String())

	header = authorization.New("Bearer xxx")
	suite.Require().Equal("Bearer xxx", header.String())
}

func (suite *HeaderSuite) TestHeader_Empty() {
	header := authorization.New("")
	suite.Require().True(header.Empty())

	header = authorization.New("  \n \t")
	suite.Require().True(header.Empty())

	header = authorization.New("  \n x\t")
	suite.Require().False(header.Empty())

	header = authorization.New("x")
	suite.Require().False(header.Empty())

	header = authorization.New("Bearer xxxx")
	suite.Require().False(header.Empty())
}

func (suite *HeaderSuite) TestHeader_NotEmpty() {
	header := authorization.New("")
	suite.Require().False(header.NotEmpty())

	header = authorization.New("  \n \t")
	suite.Require().False(header.NotEmpty())

	header = authorization.New("  \n x\t")
	suite.Require().True(header.NotEmpty())

	header = authorization.New("x")
	suite.Require().True(header.NotEmpty())

	header = authorization.New("Bearer xxxx")
	suite.Require().True(header.NotEmpty())
}

func (suite *HeaderSuite) TestWithHeader() {
	var ctx context.Context
	var headerA = authorization.New("Dog Woof")
	var headerB = authorization.New("Cat Meow")

	suite.Panics(func() {
		authorization.WithHeader(nil, headerA)
	})

	// No header on the context
	suite.Equal(authorization.None, authorization.FromContext(nil))

	// No header on the context
	ctx = context.Background()
	suite.Equal(authorization.None, authorization.FromContext(ctx))

	// Simple case where there's one authorization on the context.
	ctx = context.Background()
	ctx = authorization.WithHeader(ctx, headerA)
	suite.Equal(headerA, authorization.FromContext(ctx))

	// Header has been applied multiple times. Last in wins.
	ctx = context.Background()
	ctx = authorization.WithHeader(ctx, headerA)
	ctx = authorization.WithHeader(ctx, headerB)
	ctx = authorization.WithHeader(ctx, headerA)
	ctx = authorization.WithHeader(ctx, headerA)
	ctx = authorization.WithHeader(ctx, headerB)
	suite.Equal(headerB, authorization.FromContext(ctx))
}

func TestHeaderSuite(t *testing.T) {
	suite.Run(t, new(HeaderSuite))
}
