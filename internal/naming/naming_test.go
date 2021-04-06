package naming_test

import (
	"testing"

	"github.com/monadicstack/frodo/internal/naming"
	"github.com/stretchr/testify/suite"
)

type NamingSuite struct {
	suite.Suite
}

func (suite *NamingSuite) TestNoPackage() {
	r := suite.Require()
	r.Equal("", naming.NoPackage(""))
	r.Equal("foo", naming.NoPackage("foo"))
	r.Equal("foo_bar", naming.NoPackage("foo_bar"))
	r.Equal("bar", naming.NoPackage("foo.bar"))
	r.Equal("baz", naming.NoPackage("foo.bar.baz"))
	r.Equal("baz", naming.NoPackage("foo.bar...baz"))
	r.Equal("baz", naming.NoPackage("*foo.bar...baz"))
}

func (suite *NamingSuite) TestNoPointer() {
	r := suite.Require()
	r.Equal("", naming.NoPointer(""))
	r.Equal("foo", naming.NoPointer("foo"))
	r.Equal("foo_bar", naming.NoPointer("foo_bar"))
	r.Equal("foo*bar", naming.NoPointer("foo*bar")) // only strip from left
	r.Equal("foo*", naming.NoPointer("foo*"))
	r.Equal("foo", naming.NoPointer("*foo"))
	r.Equal("foo", naming.NoPointer("**foo"))
	r.Equal(" *foo", naming.NoPointer("* *foo")) // only works on single tokens
	r.Equal("foo.bar.baz", naming.NoPointer("**foo.bar.baz"))
	r.Equal("&foo", naming.NoPointer("&foo")) // only strip pointer declarations, not references
	r.Equal("&foo", naming.NoPointer("*&foo"))
}

func (suite *NamingSuite) TestJoinPackageName() {
	r := suite.Require()
	r.Equal("", naming.JoinPackageName(""))
	r.Equal("foo", naming.JoinPackageName("foo"))
	r.Equal("foo bar", naming.JoinPackageName("foo bar"))
	r.Equal("foobar", naming.JoinPackageName("foo.bar"))
	r.Equal("foobarbaz", naming.JoinPackageName("foo.bar.baz"))
	r.Equal("fooBar", naming.JoinPackageName("foo.Bar"))
	r.Equal("fooBar", naming.JoinPackageName("foo..Bar"))
	r.Equal("*fooBar", naming.JoinPackageName("*foo.Bar")) // you have to do your own "un-pointer-ing"
}

func (suite *NamingSuite) TestNoImport() {
	r := suite.Require()
	r.Equal("", naming.NoImport(""))
	r.Equal("foo", naming.NoImport("foo"))
	r.Equal("foo.Bar", naming.NoImport("foo.Bar"))
	r.Equal("baz", naming.NoImport("foo/bar/baz"))
	r.Equal("", naming.NoImport("foo/bar/baz/")) // assume trailing slash means missing identifier name
	r.Equal("baz", naming.NoImport("foo/  /  /bar///baz"))
	r.Equal("baz.Blah", naming.NoImport("foo/bar/baz.Blah"))
}

func (suite *NamingSuite) TestCleanPrefix() {
	r := suite.Require()
	r.Equal("", naming.CleanPrefix(""))
	r.Equal("foo", naming.CleanPrefix("foo"))
	r.Equal("foo.bar", naming.CleanPrefix("foo.bar"))
	r.Equal("*foo.Bar", naming.CleanPrefix("*foo.Bar"))
	r.Equal("*foo/bar.Baz", naming.CleanPrefix("*foo/bar.Baz"))
	r.Equal("*foo/bar.Baz", naming.CleanPrefix("*foo/bar.Baz"))
	r.Equal("", naming.CleanPrefix("command-line-arguments"))
	r.Equal("", naming.CleanPrefix("command-line-arguments."))
	r.Equal("", naming.CleanPrefix("*command-line-arguments"))
	r.Equal("****command-line-arguments.", naming.CleanPrefix("****command-line-arguments."))
	r.Equal("foo.Bar", naming.CleanPrefix("command-line-arguments.foo.Bar"))
	r.Equal("foo.Bar", naming.CleanPrefix("*command-line-arguments.foo.Bar"))
	r.Equal("****command-line-arguments.foo.Bar", naming.CleanPrefix("****command-line-arguments.foo.Bar"))
	r.Equal("COMMAND-line-arguments.foo.Bar", naming.CleanPrefix("COMMAND-line-arguments.foo.Bar")) // case sensitive
}

func (suite *NamingSuite) TestLeadingSlash() {
	r := suite.Require()
	r.Equal("/", naming.LeadingSlash(""))
	r.Equal("/", naming.LeadingSlash("/"))
	r.Equal("//////", naming.LeadingSlash("//////"))
	r.Equal("/foo", naming.LeadingSlash("foo"))
	r.Equal("/foo.bar", naming.LeadingSlash("foo.bar"))
	r.Equal("/foo/bar", naming.LeadingSlash("foo/bar"))
	r.Equal("/foo/bar//", naming.LeadingSlash("/foo/bar//"))
	r.Equal("//foo/bar//", naming.LeadingSlash("//foo/bar//"))
}

func (suite *NamingSuite) TestEmptyString() {
	r := suite.Require()
	r.Equal(true, naming.EmptyString(""))
	r.Equal(false, naming.EmptyString(" "))
	r.Equal(false, naming.EmptyString("/"))
	r.Equal(false, naming.EmptyString("foo"))
	r.Equal(false, naming.EmptyString("🍺"))
}

func (suite *NamingSuite) TestNotEmptyString() {
	r := suite.Require()
	r.Equal(false, naming.NotEmptyString(""))
	r.Equal(true, naming.NotEmptyString(" "))
	r.Equal(true, naming.NotEmptyString("/"))
	r.Equal(true, naming.NotEmptyString("foo"))
	r.Equal(true, naming.NotEmptyString("🍺"))
}

func (suite *NamingSuite) TestPathTokens() {
	r := suite.Require()
	r.Equal([]string{}, naming.PathTokens(""))
	r.Equal([]string{}, naming.PathTokens("/"))
	r.Equal([]string{"*"}, naming.PathTokens("*"))
	r.Equal([]string{"foo"}, naming.PathTokens("foo"))
	r.Equal([]string{"foo"}, naming.PathTokens("/foo"))
	r.Equal([]string{"foo"}, naming.PathTokens("/foo/"))
	r.Equal([]string{"foo"}, naming.PathTokens("foo/"))
	r.Equal([]string{"foo", "bar"}, naming.PathTokens("foo/bar"))
	r.Equal([]string{"foo", ":bar"}, naming.PathTokens("foo/:bar"))
	r.Equal([]string{"foo", ":bar", "baz"}, naming.PathTokens("foo/:bar/baz"))
	r.Equal([]string{"foo", ":bar", "", "", "baz"}, naming.PathTokens("foo/:bar///baz")) // doesn't normalize inner /
	r.Equal([]string{"foo", ":bar", "{baz.Blah}"}, naming.PathTokens("foo/:bar/{baz.Blah}"))
	r.Equal([]string{"foo", ":bar", "{baz.Blah}"}, naming.PathTokens("///foo/:bar/{baz.Blah}///"))
}

func TestNamingSuite(t *testing.T) {
	suite.Run(t, new(NamingSuite))
}
