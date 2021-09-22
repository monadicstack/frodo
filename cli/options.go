package cli

import (
	"errors"
	"log"
	"os"

	"github.com/monadicstack/frodo/generate"
	"github.com/monadicstack/frodo/parser"
)

// templateOption can be embedded on a command request struct to give it the option to supply your
// own template for performing artifact generation.
type templateOption struct {
	// Template is the path to a custom Go template that we'll use to render this artifact. By leaving
	// this option blank you will just use the standard embedded template for that type of artifact.
	Template string
}

// ToFileTemplate constructs a new 'FileTemplate' based on the 'Template' option on this command. When
// it's blank, you'll get a version pointing at one of our standard templates. If it has a value, you'll
// get a version that looks for the template on the local file system.
//
// The 'name' argument specifies the suffix we will use for the generated artifact. For instance, if you are
// generating the artifact named "client.js" for the Go service definition file "foo_service.go" then you will
// generate a file named "foo_service.gen.client.js".
func (opt templateOption) ToFileTemplate(name string) generate.FileTemplate {
	if opt.Template == "" {
		return generate.NewStandardTemplate(name, "templates/"+name+".tmpl")
	}
	return generate.NewCustomTemplate(name, opt.Template)
}

// crapPants is a catch-all handler for dealing with errors parsing code files and generating artifacts. It
// tries to give helpful, descriptive error messages that instruct the user how to address the issue in addition
// to notifying them about the failure.
func crapPants(err error) {
	if err == nil {
		return
	}

	log.Println(err.Error())
	switch {
	case errors.Is(err, parser.ErrNoServices):
		log.Println("")
		log.Println("  * Your service interface must end with 'Service' (e.g. 'UserService')")
		log.Println("  * Your interface must be exported (e.g. 'UserService', not 'userService')")
		log.Println("")
	case errors.Is(err, parser.ErrMultipleServices):
		log.Println("")
		log.Println("  * Separate services into their own files (e.g. 'UserService' in user_service.go")
		log.Println("    and 'OrderService' in order_service.go)")
		log.Println("  * It's usually more idiomatic to have one service per package (e.g. 'UserService'")
		log.Println("    goes in the 'users' package and 'OrderService' in the 'orders' package)")
		log.Println("")
	case errors.Is(err, parser.ErrMissingGoMod):
		log.Println("")
		log.Println("  * Frodo only works with projects that use go modules")
		log.Println("")
	// We want all signature-related errors to give instructions about what you need.
	case errors.Is(err, parser.ErrTypeNotStructPointer),
		errors.Is(err, parser.ErrTypeNotError),
		errors.Is(err, parser.ErrTypeNotContext),
		errors.Is(err, parser.ErrTypeNotTwoParams),
		errors.Is(err, parser.ErrTypeNotTwoReturns):
		log.Println("")
		log.Println("  * All service functions must accept two parameters.")
		log.Println("    * The 1st parameter must be 'context.Context'")
		log.Println("    * The 2nd parameter must be a pointer to a struct type")
		log.Println("  * All service functions must return two values.")
		log.Println("    * The 1st return value must be a pointer to a struct type")
		log.Println("    * The 2nd return value must be 'error'")
		log.Println("  * Example: Login(context.Context, *LoginRequest) (*LoginResponse, error)")
		log.Println("")
	}
	os.Exit(1)
}
