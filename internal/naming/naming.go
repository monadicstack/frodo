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
	if strings.HasPrefix(ident, "command-line-arguments.") {
		return ident[23:]
	}
	if strings.HasPrefix(ident, "*command-line-arguments.") {
		return ident[24:]
	}
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

// EmptyString is a predicate that returns true when the input value is "".
func EmptyString(value string) bool {
	return value == ""
}

// NotEmptyString is a predicate that returns true when the input value is anything but "".
func NotEmptyString(value string) bool {
	return value != ""
}

// Indent returns a string with a specific number of spaces left-padded on the content; useful in templates.
func Indent(spaces int, content interface{}) string {
	return strings.Repeat(" ", spaces)
}
