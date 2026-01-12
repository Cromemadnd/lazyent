package gen

import (
	"fmt"
	"strings"

	entgen "entgo.io/ent/entc/gen"
)

func getValidateRules(f *entgen.Field, nodeName string) string {
	// 1. Enum
	if f.IsEnum() && !isExternalEnum(f) {
		return ".enum.defined_only = true"
	}

	// 2. UUID
	if f.Type.String() == "uuid.UUID" {
		return ".string.uuid = true"
	}

	// 3. Annotations
	a := getFieldAnnotation(f)
	if a != nil && a.ProtoValidation != "" {
		val := a.ProtoValidation
		// Basic formatting fix
		val = strings.ReplaceAll(val, ":", ": ")
		val = strings.ReplaceAll(val, ",", ", ")
		val = strings.ReplaceAll(val, ":  ", ": ")
		val = strings.ReplaceAll(val, ",  ", ", ")

		if strings.HasPrefix(val, ".") {
			return val
		}

		if strings.HasPrefix(val, "repeated") {
			if strings.Contains(val, "items:") || strings.Contains(val, "items :") {
				return fmt.Sprintf(".repeated = { %s }", val)
			}
		}

		pType := getProtoType(f)
		return fmt.Sprintf(".%s = { %s }", pType, val)
	}

	return ""
}
