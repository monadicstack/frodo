package cli

import "github.com/monadicstack/frodo/generate"

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
