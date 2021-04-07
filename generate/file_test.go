package generate_test

import (
	"os"
	"testing"

	"github.com/monadicstack/frodo/generate"
	"github.com/stretchr/testify/suite"
)

type FileTemplateSuite struct {
	suite.Suite
}

// Ensures that Eval() fails if we can't find/read the template on the file system.
func (suite *FileTemplateSuite) TestEval_unableToRead() {
	t := generate.FileTemplate{
		Name:       "fail.txt",
		FileSystem: os.DirFS("testdata"),
		Path:       "doesnotexist.tmpl",
	}

	_, err := t.Eval("hello")
	suite.Require().Error(err, "Eval() should fail if we can't find the template on the FS")
	suite.Require().Contains(err.Error(), "read", "Failure message should indicate a failure to read the template")
}

// Ensures that Eval() fails if we found the template, but can't parse it properly.
func (suite *FileTemplateSuite) TestEval_invalidTemplateSyntax() {
	t := generate.FileTemplate{
		Name:       "fail.txt",
		FileSystem: os.DirFS("testdata"),
		Path:       "invalid.tmpl",
	}

	_, err := t.Eval("hello")
	suite.Require().Error(err, "Eval() should fail if the template has invalid syntax")
	suite.Require().Contains(err.Error(), "parse", "Failure message should indicate a failure to parse the template")
}

// Ensures that Eval() fails we have a valid template, but the data we pass causes it to fail.
func (suite *FileTemplateSuite) TestEval_failedExecute() {
	t := generate.FileTemplate{
		Name:       "fail.txt",
		FileSystem: os.DirFS("testdata"),
		Path:       "valid.tmpl",
	}

	_, err := t.Eval("hello")
	suite.Require().Error(err, "Eval() should fail if the template has invalid syntax")
	suite.Require().Contains(err.Error(), "execute", "Failure message should indicate a failure to execute the template")
}

// Ensures that Eval() properly evaluates the underlying code template with the input value. We currently
// assume that if one template resolves properly that any will. The point of the test is to make sure our
// abstraction succeeds if the standard library's 'template.Text' succeeds; we don't need to run that through
// the ringer - the Go team has done that already.
func (suite *FileTemplateSuite) TestEval_success() {
	type data struct {
		Name   string
		Tokens []string
	}
	t := generate.FileTemplate{
		Name:       "fail.txt",
		FileSystem: os.DirFS("testdata"),
		Path:       "valid.tmpl",
	}

	output, err := t.Eval(data{
		Name:   "Bob",
		Tokens: []string{"A", "B", "C"},
	})
	expected := "Hello Bob\n" +
		"-> A\n" +
		"-> B\n" +
		"-> C\n"

	suite.Require().NoError(err, "Eval() should fail if the template has invalid syntax")
	suite.Require().Equal(expected, string(output))
}

func TestFileTemplateSuite(t *testing.T) {
	suite.Run(t, new(FileTemplateSuite))
}
