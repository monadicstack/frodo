package naming

import "strings"

// NoPackage strips of any package prefixes from an identifier (e.g. "context.Context" -> "Context")
func NoPackage(ident string) string {
	period := strings.LastIndex(ident, ".")
	if period < 0 {
		return ident
	}
	return ident[period+1:]
}

// NoPointer strips off any "*" prefix your type identifier might have (e.g. "*Foo" -> "Foo")
func NoPointer(ident string) string {
	return strings.TrimLeft(ident, "*")
}

// NoSlice takes a string like "[]Foo" or "[456]Foo", strips off the slice/array braces, leaving you with "Foo".
func NoSlice(ident string) string {
	closeBrace := strings.Index(ident, "]")
	if closeBrace < 0 {
		return ident
	}
	return ident[closeBrace+1:]
}

// JoinPackageName converts a package-qualified type such as "fmt.Stringer" into a single "safe" identifier
// such as "fmtStringer". This is useful when converting types to languages with different naming semantics.
func JoinPackageName(ident string) string {
	return strings.ReplaceAll(ident, ".", "")
}

// NoImport strips off the prefix before "/" in your type identifier (e.g. "github.com/foo/bar/Baz" -> "Baz")
func NoImport(ident string) string {
	if slash := strings.LastIndex(ident, "/"); slash >= 0 {
		ident = ident[slash+1:]
	}
	return ident
}

// CleanPrefix strips the "command-line-arguments." prefix that the Go 'packages' package prepends to type
// identifier for types defined in the source file we're parsing.
func CleanPrefix(ident string) string {
	ident = strings.ReplaceAll(ident, "command-line-arguments.", "")
	ident = strings.ReplaceAll(ident, "command-line-arguments", "")
	return ident
}

// LeadingSlash adds... a leading slash to the given string.
func LeadingSlash(value string) string {
	if strings.HasPrefix(value, "/") {
		return value
	}
	return "/" + value
}

// ToLowerCamel converts the string to lower camel-cased.
func ToLowerCamel(value string) string {
	// This is a shitty implementation.
	if value == "" {
		return ""
	}
	firstChar := value[0:1]
	return strings.ToLower(firstChar) + value[1:]
}

// ToUpperCamel converts the string to upper camel-cased.
func ToUpperCamel(value string) string {
	// This is a shitty implementation.
	if value == "" {
		return ""
	}
	firstChar := value[0:1]
	return strings.ToUpper(firstChar) + value[1:]
}

// EmptyString is a predicate that returns true when the input value is "".
func EmptyString(value string) bool {
	return value == ""
}

// NotEmptyString is a predicate that returns true when the input value is anything but "".
func NotEmptyString(value string) bool {
	return value != ""
}

// PathTokens accepts a path string like "foo/bar/baz" and returns a slice of the individual
// path segment tokens such as ["foo", "bar", "baz"]. This will ignore leading/trailing slashes
// in your path so that you don't get leading/trailing "" tokens in your slice. This does not,
// however, clean up empty tokens caused by "//" somewhere in your path.
func PathTokens(path string) []string {
	path = strings.Trim(path, "/")
	if path == "" {
		return []string{}
	}
	return strings.Split(path, "/")
}

func CleanTypeNameUpper(typeName string) string {
	typeName = CleanPrefix(typeName)
	typeName = NoSlice(typeName)
	typeName = NoPointer(typeName)
	typeName = JoinPackageName(typeName)
	typeName = ToUpperCamel(typeName)
	return typeName
}

func DispositionFileName(contentDisposition string) string {
	// The start or the file name in the header is the index of "filename=" plus the 9
	// characters in that substring.
	fileNameAttrIndex := strings.Index(contentDisposition, "filename=")
	if fileNameAttrIndex < 0 {
		return ""
	}

	// Support the fact that all of these are valid for the disposition header:
	//
	//   attachment; filename=foo.pdf
	//   attachment; filename="foo.pdf"
	//   attachment; filename='foo.pdf'
	//
	// This just makes sure that you don't have any quotes in your final value.
	fileName := contentDisposition[fileNameAttrIndex+9:]
	fileName = strings.Trim(fileName, `"'`)
	fileName = strings.ReplaceAll(fileName, `\"`, `"`)
	return fileName
}
