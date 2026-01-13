package gen

import (
	"strings"
	"text/template"
)

var funcMap = template.FuncMap{
	"getEnumValues": getEnumValues,
	"getEnumPairs":  getEnumPairs,
	"protoType":     protoType,

	"getProtoTag":          getProtoTag,
	"convertToProto":       convertToProto,
	"convertFromProto":     convertFromProto,
	"add":                  func(a, b int) int { return a + b },
	"lower":                strings.ToLower,
	"upper":                strings.ToUpper,
	"hasTime":              hasTime,
	"hasUUID":              hasUUID,
	"hasTimeNodes":         hasTimeNodes,
	"hasUUIDNodes":         hasUUIDNodes,
	"edgeHasFK":            edgeHasFK,
	"edgeField":            edgeField,
	"hasField":             hasField,
	"protoStructField":     protoStructField,
	"protoGoName":          protoGoName,
	"edgeIDType":           edgeIDType,
	"edgeProtoType":        edgeProtoType,
	"edgeConvertToProto":   edgeConvertToProto,
	"edgeConvertFromProto": edgeConvertFromProto,
	"zeroValue":            zeroValue,

	"isBizIDOnly":  isBizIDOnly,
	"isBizExclude": isBizExclude,
	"isBizPointer": isBizPointer,
	"isSensitive":  isSensitive,

	"isProtoID":              isProtoID,
	"isProtoMessage":         isProtoMessage,
	"isProtoExclude":         isProtoExclude,
	"enumToProtoFunc":        enumToProtoFuncName,
	"enumFromProtoFunc":      enumFromProtoFuncName,
	"getAllEnums":            getAllEnums,
	"bizFieldName":           bizFieldName,
	"isExternalEnum":         isExternalEnum,
	"getExternalEnumPkg":     getExternalEnumPkg,
	"getExternalEnumName":    getExternalEnumName,
	"isSlice":                isSlice,
	"getSliceElementType":    getSliceElementType,
	"getGoProtoType":         getGoProtoType,
	"isSliceTypeMatch":       isSliceTypeMatch,
	"collectExternalImports": collectExternalImports,
	"getEnumLiteralValues":   getEnumLiteralValues,
	"bizFieldType":           bizFieldType,
	"explicitBizType":        explicitBizType,
	"bizEdgeName":            bizEdgeName,
	"pascal":                 pascal,
	"camel":                  camel,
	"convertEntToBiz":        convertEntToBiz,
	"convertBizToEnt":        convertBizToEnt,

	// New Safe Mapper Functions
	"requiresErrorCheck":    requiresErrorCheck,
	"convertFromProtoSetup": convertFromProtoSetup,
	"convertFromProtoUsage": convertFromProtoUsage,
	"convertBizToEntSetup":  convertBizToEntSetup,
	"convertBizToEntUsage":  convertBizToEntUsage,
}
