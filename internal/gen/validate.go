package gen

import (
	"entgo.io/ent/entc/gen"
)

// getFieldValidateRules extracts buf validate rules from field annotations.
// In the future, this can be expanded to inspect ent's built-in validators.
func getFieldValidateRules(f *gen.Field) string {
	a := getFieldAnnotation(f)
	if a != nil && a.ProtoValidation != "" {
		return a.ProtoValidation
	}
	// TODO: Automatic extraction from f.Validators
	return ""
}
