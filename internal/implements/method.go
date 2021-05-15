package implements

import (
	"go/types"
)

// Method returns true if the given type is a named structure that implements the specified method
// signature. Since you might be referencing types that are hard to look up in the AST packages info,
// you can just supply the qualified names of the types for your params and return values (e.g. "io.Reader"
// or "http.Client").
func Method(t types.Type, name string, paramTypes []string, returnTypes []string) bool {
	named, ok := t.(*types.Named)
	if !ok {
		return false
	}

	for i := 0; i < named.NumMethods(); i++ {
		fn := named.Method(i)
		if Signature(fn, name, paramTypes, returnTypes) {
			return true
		}
	}
	return false
}

// Signature accepts a single method and determines whether or not it has the same name, parameter types, and return
// types as what you provide. Since you might be referencing types that are hard to look up in the AST packages info,
// you can just supply the qualified names of the types for your params and return values (e.g. "io.Reader"
// or "http.Client").
func Signature(method *types.Func, name string, paramTypes []string, resultTypes []string) bool {
	signature, ok := method.Type().(*types.Signature)
	if !ok {
		return false
	}
	if method.Name() != name {
		return false
	}

	for i, paramType := range paramTypes {
		param := signature.Params().At(i)
		if param.Type().String() != paramType {
			return false
		}
	}

	for i, resultType := range resultTypes {
		result := signature.Results().At(i)
		if result.Type().String() != resultType {
			return false
		}
	}

	return true
}
