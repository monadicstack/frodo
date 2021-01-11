package generate

import (
	"bytes"
	"fmt"
	"go/format"
	"log"
	"os"
	"strings"
	"text/template"

	"github.com/robsignorelli/frodo/parser"
)

// Artifact runs the parsed service information through the code template to
// generate the ".gen.xxx.go" client/gateway code.
func Artifact(ctx *parser.Context, inputFileName string, codeTemplate *template.Template) error {
	outputFileName := strings.TrimSuffix(inputFileName, ".go") + ".gen." + codeTemplate.Name()
	log.Printf("[frodoc] Generating artifact: %s", outputFileName)

	_ = os.Remove(outputFileName)
	outputFile, err := os.Create(outputFileName)
	if err != nil {
		return fmt.Errorf("unable to open file: %s: %w", outputFileName, err)
	}
	defer outputFile.Close()

	sourceCode, err := eval(ctx, codeTemplate)
	if err != nil {
		return fmt.Errorf("template eval error: %s: %v", codeTemplate.Name(), err)
	}

	sourceCode, err = format.Source(sourceCode)
	if err != nil {
		return fmt.Errorf("error running 'go fmt': %s: %v", codeTemplate.Name(), err)
	}

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
