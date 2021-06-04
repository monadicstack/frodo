package generate

import (
	"bytes"
	"embed"
	"fmt"
	"go/format"
	"io/fs"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"text/template"

	"github.com/monadicstack/frodo/internal/naming"
	"github.com/monadicstack/frodo/parser"
)

//go:embed templates/*
// StandardTemplates provides access to all of the code generation templates that Frodo ships with out of the box.
var StandardTemplates embed.FS

// File runs the parsed service context through the given file template, generating the appropriate
// code/project file. The 'ctx' will be fed in as the root data to the Go template represented by
// the fileTemplate parameter.
func File(ctx *parser.Context, fileTemplate FileTemplate) error {
	inputFileName := filepath.Base(ctx.Path)
	inputDir := filepath.Dir(ctx.Path)

	outputFileName := strings.TrimSuffix(inputFileName, ".go") + ".gen." + fileTemplate.Name
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
	sourceCode, err := fileTemplate.Eval(ctx)
	if err != nil {
		return fmt.Errorf("template eval error: %s: %v", fileTemplate.Name, err)
	}

	// Step 4: Run the generated source code through "go fmt" (if generating a Go artifact)
	original := sourceCode
	sourceCode, err = prettify(fileTemplate, sourceCode)
	if err != nil {
		fmt.Println(string(original))
		return fmt.Errorf("error running 'go fmt': %s: %v", fileTemplate.Name, err)
	}

	// Step 5: Write your cleaned up code to the actual output file.
	_, err = outputFile.Write(sourceCode)
	if err != nil {
		return fmt.Errorf("error writing generated code: %s: %w", fileTemplate.Name, err)
	}
	return nil
}

// NewStandardTemplate creates the metadata that points to one of our standard, built-in
// templates for a gateway, client, etc.
func NewStandardTemplate(name string, path string) FileTemplate {
	return FileTemplate{
		Name:       name,
		FileSystem: StandardTemplates,
		Path:       path,
	}
}

// NewCustomTemplate creates the metadata that points to a custom template defined by the user
// running one of our CLI commands. They might have their own ".tmpl" file somewhere on their hard
// drive and this allows you to swap that into our artifact generation logic in place of
// one of our built-in templates.
func NewCustomTemplate(name string, path string) FileTemplate {
	// DirFS() is wired to not let you navigate to a parent directory of the location you pass into the constructor
	// function. We want to support either using a relative directory or an absolute one; either of which could
	// point to a file anywhere on the developer's hard drive.
	//
	// To work around this, we'll only work in absolute paths. This isn't a web server so it's up to the dev where
	// they want to load template files from w/o worrying about security. By expanding relative paths to absolute
	// we can root the DirFS at "/" and everything should work out. The only quirky thing (and maybe I'm doing
	// something wrong) is that I need to strip the leading "/" off of our absolute path because I think DirFS
	// will otherwise try to load the file "//foo/bar/baz.txt" when we Open("/foo/bar/baz.txt"). So we need to turn
	// it into Open("foo/bar/baz.txt") for the paths to work out nicely.
	absolutePath, _ := filepath.Abs(path)
	return FileTemplate{
		Name:       name,
		FileSystem: os.DirFS(""),
		Path:       absolutePath[1:],
	}
}

// FileTemplate tracks the data needed to load a code generation template for one of our output artifacts.
// This can be one of our built-in templates (using embed.FS) or a
type FileTemplate struct {
	// Name is the identifier used to indicate "which" file you're generating. For example you might set
	// this to "client.go" when generating a Go RPC client file or "gateway.go" when generating the API
	// gateway. In practice this is generally used when building the file name for the generated file.
	Name string
	// FileSystem is the store where we can look up the code template.
	FileSystem fs.FS
	// Path is the location on the FileSystem where this template is located.
	Path string
}

// Eval runs the given value through the Go template resolved by looking up Path in the FileSystem. The 'data'
// value is the root context value we'll pass to the template when running Execute(). This will return the complete
// set of bytes for the output file contents.
func (t FileTemplate) Eval(data interface{}) ([]byte, error) {
	templateData, err := fs.ReadFile(t.FileSystem, t.Path)
	if err != nil {
		return nil, fmt.Errorf("unable to read template: %w", err)
	}

	templateText := string(templateData)
	codeTemplate, err := template.New(t.Name).Funcs(templateFuncs).Parse(templateText)

	if err != nil {
		return nil, fmt.Errorf("unable to parse template: %w", err)
	}

	buf := &bytes.Buffer{}
	err = codeTemplate.Execute(buf, data)
	if err != nil {
		return nil, fmt.Errorf("unable to execute template: %w", err)
	}
	return buf.Bytes(), nil
}

// prettify runs your generated Go code through 'go fmt'. If the template is for some
// language other than Go, we'll return the source code as-is.
func prettify(t FileTemplate, sourceCode []byte) ([]byte, error) {
	if !strings.HasSuffix(t.Name, ".go") {
		return sourceCode, nil
	}
	return format.Source(sourceCode)
}

// templateFuncs are all of pipe functions we want available when evaluating the Go template
// to generate an artifact's source code.
var templateFuncs = template.FuncMap{
	// General purpose string manipulators
	"CleanPrefix":        naming.CleanPrefix,
	"CleanTypeNameUpper": naming.CleanTypeNameUpper,
	"NoPointer":          naming.NoPointer,
	"NoPackage":          naming.NoPackage,
	"JoinPackageName":    naming.JoinPackageName,
	"LeadingSlash":       naming.LeadingSlash,
	"ToLowerCamel":       naming.ToLowerCamel,
	"ToUpperCamel":       naming.ToUpperCamel,
	"EmptyString":        naming.EmptyString,
	"NotEmptyString":     naming.NotEmptyString,
	"PathTokens":         naming.PathTokens,
	"ToLower":            strings.ToLower,
	"ToUpper":            strings.ToUpper,

	// Language/format-specific value conversions
	"JSONType":       jsonFunctions{}.convertType,
	"JSPropertyType": jsFunctions{}.convertPropertyType,
	"JSTypedefType":  jsFunctions{}.convertTypedefType,
	"JavaPackage":    javaFunctions{}.convertPackage,
	"JavaType":       javaFunctions{}.convertType,
	"DartType":       dartFunctions{}.convertType,
	"OpenAPIPath":    openapiFunctions{}.convertPath,
}

type jsFunctions struct{}

func (funcs jsFunctions) convertPropertyType(t *parser.TypeDeclaration) string {
	if !t.Basic {
		return naming.JoinPackageName(naming.NoPointer(t.Name))
	}
	return funcs.convertTypedefType(t)
}

func (funcs jsFunctions) convertTypedefType(t *parser.TypeDeclaration) string {
	switch t.Kind {
	case reflect.String:
		return "string"
	case reflect.Bool:
		return "boolean"
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return "number"
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return "number"
	case reflect.Float32, reflect.Float64:
		return "number"
	case reflect.Complex64, reflect.Complex128:
		return "number"
	case reflect.Array, reflect.Slice:
		elemType := funcs.convertPropertyType(t.Elem)
		return "Array<" + elemType + ">"
	case reflect.Map:
		keyType := funcs.convertPropertyType(t.Key)
		elemType := funcs.convertPropertyType(t.Elem)
		return "Map<" + keyType + "," + elemType + ">"
	case reflect.Struct, reflect.Interface:
		return "object"
	default:
		return "*"
	}
}

type jsonFunctions struct{}

func (funcs jsonFunctions) convertType(t *parser.TypeDeclaration) string {
	switch t.Kind {
	case reflect.String:
		return "string"
	case reflect.Bool:
		return "boolean"
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return "number"
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return "number"
	case reflect.Float32, reflect.Float64, reflect.Complex64, reflect.Complex128:
		return "number"
	case reflect.Array, reflect.Slice:
		return "array"
	default:
		return "object"
	}
}

type javaFunctions struct{}

func (funcs javaFunctions) convertPackage(packageName string) string {
	// Split the package like "github.com/myorg/mymodule/a/b/c" into the segments
	// separated by slashes. Omit the first segment which is the address; regardless
	// of whether it's GitHub, GitLab, or whatever. Then put the remaining segments
	// back together using periods. In the example, the result would be
	// "myorg.mymodule.a.b.c"
	segments := strings.Split(packageName, "/")
	segments = segments[1:]
	return strings.Join(segments, ".")
}

func (funcs javaFunctions) convertType(t *parser.TypeDeclaration) string {
	switch t.Kind {
	case reflect.String:
		return "String"
	case reflect.Bool:
		return "boolean"
	case reflect.Int8, reflect.Uint8:
		return "byte"
	case reflect.Int16, reflect.Uint16:
		return "short"
	case reflect.Int, reflect.Int32, reflect.Uint, reflect.Uint32:
		return "int"
	case reflect.Int64, reflect.Uint64:
		return "long"
	case reflect.Float32:
		return "float"
	case reflect.Float64:
		return "double"
	case reflect.Complex64, reflect.Complex128:
		return "double"
	case reflect.Array, reflect.Slice:
		elemType := funcs.convertType(t.Elem)
		return "java.util.List<" + elemType + ">"
	case reflect.Map:
		elemType := funcs.convertType(t.Elem)
		keyType := funcs.convertType(t.Key)
		return "java.util.Map<" + keyType + "," + elemType + ">"
	case reflect.Struct, reflect.Interface:
		return t.Name
	default:
		return "Object"
	}
}

type dartFunctions struct{}

func (funcs dartFunctions) convertType(t *parser.TypeDeclaration) string {
	if !t.Basic {
		return naming.JoinPackageName(naming.NoPointer(t.Name))
	}
	switch t.Kind {
	case reflect.String:
		return "String"
	case reflect.Bool:
		return "bool"
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return "int"
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return "int"
	case reflect.Float32, reflect.Float64:
		return "double"
	case reflect.Complex64, reflect.Complex128:
		return "double"
	case reflect.Array, reflect.Slice:
		elemType := funcs.convertType(t.Elem)
		return "List<" + elemType + ">"
	case reflect.Map:
		keyType := funcs.convertType(t.Key)
		elemType := funcs.convertType(t.Elem)
		return "Map<" + keyType + "," + elemType + ">"
	case reflect.Struct, reflect.Interface:
		return "dynamic"
	default:
		return "dynamic"
	}
}

type openapiFunctions struct{}

// convertPath converts a router-compatible path pattern like "/foo/:bar/baz/:goo" to the equivalent
// path that OpenAPI/Swagger prefers: "/foo/{bar}/baz/{goo}"
func (funcs openapiFunctions) convertPath(path string) string {
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
}
