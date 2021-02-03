package generate

import (
	"bytes"
	"fmt"
	"go/format"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/robsignorelli/frodo/parser"
)

// Artifact runs the parsed service information through the code template to
// generate the ".gen.xxx.go" client/gateway code.
func Artifact(ctx *parser.Context, inputPath string, codeTemplate *template.Template) error {
	inputFileName := filepath.Base(inputPath)
	inputDir := filepath.Dir(inputPath)

	outputFileName := strings.TrimSuffix(inputFileName, ".go") + ".gen." + codeTemplate.Name()
	outputDir := filepath.Join(inputDir, "gen")
	outputPath := filepath.Join(outputDir, outputFileName)

	// Step 1: Create the "gen/" directory in the same directory as the file we're parsing.
	err := os.MkdirAll(outputDir, os.ModePerm)
	if err != nil {
		return fmt.Errorf("unable to create directory: %s: %w", outputDir, err)
	}

	// Step 2: Recreate the output ".gen.xxx" file from scratch.
	_ = os.Remove(outputPath)
	outputFile, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("unable to open file: %s: %w", outputPath, err)
	}
	defer outputFile.Close()

	// Step 3: Generate a []byte containing all of the source code bytes that we generated from the template.
	sourceCode, err := eval(ctx, codeTemplate)
	if err != nil {
		return fmt.Errorf("template eval error: %s: %v", codeTemplate.Name(), err)
	}

	// Step 4: Run the generated source code through "go fmt" (if generating a Go artifact)
	sourceCode, err = prettify(codeTemplate, sourceCode)
	if err != nil {
		return fmt.Errorf("error running 'go fmt': %s: %v", codeTemplate.Name(), err)
	}

	// Step 5: Write your cleaned up code to the actual output file.
	_, err = outputFile.Write(sourceCode)
	if err != nil {
		return fmt.Errorf("error writing generated code: %s: %w", codeTemplate.Name(), err)
	}
	return nil
}

func eval(ctx *parser.Context, t *template.Template) ([]byte, error) {
	buf := &bytes.Buffer{}
	err := t.Execute(buf, ctx)
	return buf.Bytes(), err
}

// prettify runs your generated Go code through 'go fmt'. If the template is for some
// language other than Go, we'll return the source code as-is.
func prettify(t *template.Template, sourceCode []byte) ([]byte, error) {
	if !strings.HasSuffix(t.Name(), ".go") {
		return sourceCode, nil
	}
	return format.Source(sourceCode)
}

func parseArtifactTemplate(name string, text string) *template.Template {
	return template.Must(template.New(name).Funcs(templateFuncs).Parse(text))
}

// templateFuncs are all of pipe functions we want available when evaluating the Go template
// to generate an artifact's source code.
var templateFuncs = template.FuncMap{
	"HTTPMethodSupportsBody": func(method string) bool {
		return method == "POST" || method == "PUT" || method == "PATCH"
	},
	"LeadingSlash": func(value string) string {
		if strings.HasPrefix(value, "/") {
			return value
		}
		return "/" + value
	},
	"NonPointer": func(value string) string {
		if strings.HasPrefix(value, "*") {
			return value[1:]
		}
		if strings.HasPrefix(value, "&") {
			return value[1:]
		}
		return value
	},
	"ToLower": func(value string) string {
		return strings.ToLower(value)
	},
	"EmptyString": func(value string) bool {
		return value == ""
	},
	"NotEmptyString": func(value string) bool {
		return value != ""
	},
	"OpenAPIPath": func(path string) string {
		segments := strings.Split(path, "/")
		for i, segment := range segments {
			if strings.HasPrefix(segment, ":") {
				segments[i] = "{" + segment[1:] + "}"
			}
		}
		path = strings.Join(segments, "/")
		if strings.HasPrefix(path, "/") {
			return path
		}
		return "/" + path
	},
}
