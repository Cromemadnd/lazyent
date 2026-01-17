package gen

import (
	"bytes"
	"embed"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"
	"text/template"

	entgen "entgo.io/ent/entc/gen"
	"golang.org/x/mod/modfile"
	"golang.org/x/tools/imports"
)

//go:embed templates/*
var templates embed.FS

// --- Descriptor Definitions (Renamed to Pb* to avoid conflict with ProtoMessage constant) ---

// PbFile represents the structure of a generated .proto file.
type PbFile struct {
	Package   string
	GoPackage string
	Imports   []string
	Elements  []PbElement // Unified list
}

type PbElement struct {
	Enum    *PbEnum
	Message *PbMessage
}

type PbMessage struct {
	Name    string
	Fields  []*PbField
	Comment string
}

type PbField struct {
	Name          string
	Type          string
	Tag           int
	Repeated      bool
	Comment       string
	ValidateRules string
}

type PbEnum struct {
	Name   string
	Values []*PbEnumValue
}

type PbEnumValue struct {
	Name   string
	Number int32
}

// AddImport adds an import path if it doesn't exist.
func (f *PbFile) AddImport(path string) {
	for _, i := range f.Imports {
		if i == path {
			return
		}
	}
	f.Imports = append(f.Imports, path)
	sort.Strings(f.Imports)
}

// --- End Descriptor Definitions ---

func Generate(conf Config, g *entgen.Graph) error {
	e := &Generator{
		conf: conf,
	}
	return e.generate(g)
}

type Generator struct {
	conf Config
}

func (e *Generator) generate(g *entgen.Graph) error {
	modPath, moduleRoot, err := getModulePath()
	if err != nil {
		return fmt.Errorf("failed to get module path: %w", err)
	}

	// 1. Resolve Defaults
	e.resolveDefaults(g)

	// 2. Prepare Template Data
	protoPkg := e.conf.ProtoPackage
	if protoPkg == "" {
		protoPkg = g.Package
	}

	commonData := map[string]interface{}{
		"Package":    e.conf.ProtoPackage,
		"Module":     modPath,
		"BizPackage": path.Join(modPath, filepath.ToSlash(e.conf.BizOut)),
		"ApiPackage": path.Join(modPath, filepath.ToSlash(e.conf.ProtoOut)),
		"GoPackage":  e.conf.GoPackage,
		"EntPackage": g.Config.Package,
	}

	// 3. Prepare Nodes Data
	var allNodes []interface{}
	for _, n := range g.Nodes {
		var enums []*entgen.Field
		for _, f := range n.Fields {
			if f.IsEnum() {
				enums = append(enums, f)
			}
		}
		nodeData := map[string]interface{}{
			"Name":   n.Name,
			"ID":     n.ID,
			"Fields": n.Fields,
			"Edges":  n.Edges,
			"Enums":  enums,
		}
		allNodes = append(allNodes, nodeData)
	}

	// 4. Generate
	var generatedProtoFiles []string

	if e.conf.SingleFile {
		// --- Phase 1: Proto Generation ---
		// Proto Files (Using new Builder)
		protoDesc, err := e.buildProtoFile(g) // Build Descriptor
		if err != nil {
			return err
		}
		// Reset GoPackage if needed
		if protoDesc.GoPackage == "" {
			protoDesc.GoPackage = e.conf.GoPackage
		}

		protoPath := filepath.Join(moduleRoot, e.conf.ProtoOut, e.conf.ProtoFileName)
		if err := e.render(nil, "templates/proto.tmpl", protoPath, protoDesc); err != nil {
			return err
		}
		generatedProtoFiles = append(generatedProtoFiles, protoPath)

		// --- Phase 3: Go Generation ---
		// Single file generation data
		data := make(map[string]interface{})
		for k, v := range commonData {
			data[k] = v
		}
		data["Nodes"] = allNodes

		// Biz Base
		if err := e.render(nil, "templates/base.tmpl", filepath.Join(moduleRoot, e.conf.BizOut, e.conf.BizBaseFileName), data); err != nil {
			return err
		}

		// Biz Entities (Scaffold)
		scaffoldPath := filepath.Join(moduleRoot, e.conf.BizOut, e.conf.BizEntityFileName)
		if _, err := os.Stat(scaffoldPath); os.IsNotExist(err) {
			if err := e.render(nil, "templates/scaffold.tmpl", scaffoldPath, data); err != nil {
				return err
			}
		}

		// Service Mappers
		if err := e.render(nil, "templates/service_mapper.tmpl", filepath.Join(moduleRoot, e.conf.ServiceOut, e.conf.SvcMapperFileName), data); err != nil {
			return err
		}

		// Data Mappers (Ent)
		if err := e.render(nil, "templates/data_mapper.tmpl", filepath.Join(moduleRoot, e.conf.DataOut, e.conf.DataMapperFileName), data); err != nil {
			return err
		}

	} else {
		// Multiple files generation

		// --- Phase 1: Proto Generation ---
		for _, nd := range allNodes {
			ndMap := nd.(map[string]interface{})
			name := ndMap["Name"].(string)
			lName := strings.ToLower(name)

			singleNodeProto, err := e.buildSingleNodeProto(g, name)
			if err != nil {
				return err
			}
			protoPath := filepath.Join(moduleRoot, e.conf.ProtoOut, lName+".proto")
			if err := e.render(nil, "templates/proto.tmpl", protoPath, singleNodeProto); err != nil {
				return err
			}
			generatedProtoFiles = append(generatedProtoFiles, protoPath)
		}

		// --- Phase 3: Go Generation ---
		for _, nd := range allNodes {
			// Basic template data (legacy templates)
			data := make(map[string]interface{})
			for k, v := range commonData {
				data[k] = v
			}
			data["Nodes"] = []interface{}{nd}
			ndMap := nd.(map[string]interface{})
			name := ndMap["Name"].(string)
			lName := strings.ToLower(name)

			// 1. Biz Base
			if err := e.render(nil, "templates/base.tmpl", filepath.Join(moduleRoot, e.conf.BizOut, lName+"_base_gen.go"), data); err != nil {
				return err
			}

			// 2. Biz Scaffold
			scaffoldPath := filepath.Join(moduleRoot, e.conf.BizOut, lName+".go")
			if _, err := os.Stat(scaffoldPath); os.IsNotExist(err) {
				if err := e.render(nil, "templates/scaffold.tmpl", scaffoldPath, data); err != nil {
					return err
				}
			}

			// 3. Service Mapper
			if err := e.render(nil, "templates/mapper.tmpl", filepath.Join(moduleRoot, e.conf.ServiceOut, lName+"_mapper.go"), data); err != nil {
				return err
			}

			// 4. Data Mapper
			if err := e.render(nil, "templates/data.tmpl", filepath.Join(moduleRoot, e.conf.DataOut, lName+"_ent.go"), data); err != nil {
				return err
			}
		}
	}

	return nil
}

func (e *Generator) resolveDefaults(g *entgen.Graph) {
	if e.conf.BizOut == "" {
		e.conf.BizOut = "internal/biz"
	}
	if e.conf.ServiceOut == "" {
		e.conf.ServiceOut = "internal/service"
	}
	if e.conf.DataOut == "" {
		e.conf.DataOut = "internal/data"
	}
	if e.conf.ProtoOut == "" {
		e.conf.ProtoOut = "api/v1"
	}

	if e.conf.ProtoPackage == "" {
		e.conf.ProtoPackage = g.Package
		if e.conf.ProtoPackage == "ent" {
			e.conf.ProtoPackage = "api.v1"
		}
	}

	if e.conf.BizBaseFileName == "" {
		e.conf.BizBaseFileName = "entities_base_gen.go"
	}
	if e.conf.BizEntityFileName == "" {
		e.conf.BizEntityFileName = "entities.go"
	}
	if e.conf.SvcMapperFileName == "" {
		e.conf.SvcMapperFileName = "service_mappers_gen.go"
	}
	if e.conf.DataMapperFileName == "" {
		e.conf.DataMapperFileName = "data_mappers_gen.go"
	}
	if e.conf.ProtoFileName == "" {
		e.conf.ProtoFileName = "dtos_gen.proto"
	}
}

func (e *Generator) render(n *entgen.Type, tmplName string, targetPath string, data interface{}) error {
	t, err := template.New(filepath.Base(tmplName)).Funcs(funcMap).ParseFS(templates, tmplName)
	if err != nil {
		return fmt.Errorf("failed to parse template %s: %w", tmplName, err)
	}

	var buf bytes.Buffer
	if err := t.Execute(&buf, data); err != nil {
		return fmt.Errorf("failed to execute template %s: %w", tmplName, err)
	}

	if err := os.MkdirAll(filepath.Dir(targetPath), 0755); err != nil {
		return fmt.Errorf("failed to create directory for %s: %w", targetPath, err)
	}

	content := buf.Bytes()
	// Format Go files with goimports (auto-fix imports)
	if strings.HasSuffix(targetPath, ".go") {
		// Try goimports first (handles imports + formatting)
		formatted, err := imports.Process(targetPath, content, nil)
		if err != nil {
			// Fallback: proto files might not be compiled yet, just save raw generated code
			// User will need to run protoc first, then goimports manually or via their build process
			fmt.Printf("⚠️  Warning: Could not auto-fix imports for %s\n", filepath.Base(targetPath))
			fmt.Printf("    Reason: %v\n", err)
			fmt.Printf("    → Please compile proto files first (protoc), then run 'goimports -w %s'\n", targetPath)
			// Don't format at all to preserve the template output for debugging
		} else {
			content = formatted
		}
	}

	return os.WriteFile(targetPath, content, 0644)
}

func getModulePath() (string, string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", "", err
	}
	for {
		path := filepath.Join(dir, "go.mod")
		if _, err := os.Stat(path); err == nil {
			data, err := os.ReadFile(path)
			if err != nil {
				return "", "", err
			}
			f, err := modfile.Parse("go.mod", data, nil)
			if err != nil {
				return "", "", err
			}
			return f.Module.Mod.Path, dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return "", "", fmt.Errorf("go.mod not found")
}

// buildProtoFile constructs the PbFile descriptor from entgen.Graph
func (e *Generator) buildProtoFile(g *entgen.Graph) (*PbFile, error) {
	// Ensure protoPkg is set
	e.resolveDefaults(g)

	files := &PbFile{
		Package:   e.conf.ProtoPackage,
		GoPackage: e.conf.GoPackage,
		Imports:   []string{},
	}

	for _, n := range g.Nodes {
		// Collect Enums (skip external enums as they use string type in proto)
		for _, f := range n.Fields {
			if f.IsEnum() && !isExternalEnum(f) {
				files.Elements = append(files.Elements, PbElement{Enum: e.buildProtoEnum(n, f)})
			}
		}

		// Append Output Message
		msg := e.buildProtoMessage(n, files, false)
		files.Elements = append(files.Elements, PbElement{Message: msg})

		// Append Input Message
		inputMsg := e.buildProtoMessage(n, files, true)
		inputMsg.Name = n.Name + "Input"
		files.Elements = append(files.Elements, PbElement{Message: inputMsg})
	}

	e.ensureValidateImport(files)
	return files, nil
}

func (e *Generator) buildSingleNodeProto(g *entgen.Graph, nodeName string) (*PbFile, error) {
	e.resolveDefaults(g)

	files := &PbFile{
		Package:   e.conf.ProtoPackage,
		GoPackage: e.conf.GoPackage,
		Imports:   []string{},
	}
	for _, n := range g.Nodes {
		if n.Name != nodeName {
			continue
		}
		// Enums then Messages
		for _, f := range n.Fields {
			if f.IsEnum() {
				files.Elements = append(files.Elements, PbElement{Enum: e.buildProtoEnum(n, f)})
			}
		}
		msg := e.buildProtoMessage(n, files, false)
		files.Elements = append(files.Elements, PbElement{Message: msg})

		inputMsg := e.buildProtoMessage(n, files, true)
		inputMsg.Name = n.Name + "Input"
		files.Elements = append(files.Elements, PbElement{Message: inputMsg})
	}
	e.ensureValidateImport(files)
	return files, nil
}

func (e *Generator) ensureValidateImport(f *PbFile) {
	hasValidate := false
	for _, el := range f.Elements {
		if el.Message == nil {
			continue
		}
		for _, fd := range el.Message.Fields {
			if fd.ValidateRules != "" {
				hasValidate = true
				break
			}
		}
		if hasValidate {
			break
		}
	}
	if hasValidate {
		// append invalidates if not exists
		found := false
		for _, imp := range f.Imports {
			if imp == "buf/validate/validate.proto" {
				found = true
				break
			}
		}
		if !found {
			f.Imports = append(f.Imports, "buf/validate/validate.proto")
		}
	}
}

func (e *Generator) buildProtoMessage(n *entgen.Type, f *PbFile, in bool) *PbMessage {
	msg := &PbMessage{
		Name: n.Name,
	}

	// 1. Fields (ID + Regular)
	usedTags := make(map[int]bool)
	var allFields []fieldInfo

	fields := e.buildProtoFields(n, f, usedTags, in)
	allFields = append(allFields, fields...)

	// 2. Edges
	edges := e.buildProtoEdges(n, in)
	allFields = append(allFields, edges...)

	// 3. Assign Tags
	e.assignProtoTags(msg, allFields, usedTags)

	return msg
}

type fieldInfo struct {
	isID  bool
	field *entgen.Field
	edge  *entgen.Edge
	pf    *PbField
}

func (e *Generator) buildProtoFields(n *entgen.Type, f *PbFile, usedTags map[int]bool, in bool) []fieldInfo {
	var results []fieldInfo

	// 1. ID
	if n.ID != nil {
		if !isFieldProtoExclude(n.ID, in) {
			pf := &PbField{
				Name:    protoFieldName(n.ID),
				Comment: n.ID.Comment(),
			}
			pf.Type = e.resolveProtoType(n.ID, n.Name, f)

			t := getProtoTag(n.ID, -1)
			if t > 0 {
				pf.Tag = t
				usedTags[t] = true
			}
			results = append(results, fieldInfo{isID: true, field: n.ID, pf: pf})
		}
	}

	// 2. Fields
	for _, fld := range n.Fields {
		// Check Strategy
		if isFieldProtoExclude(fld, in) {
			continue
		}

		pf := &PbField{
			Name:    protoFieldName(fld),
			Comment: fld.Comment(),
		}
		// For external enums, use string type in proto
		if fld.IsEnum() && isExternalEnum(fld) {
			pf.Type = "string"
		} else {
			pf.Type = e.resolveProtoType(fld, n.Name, f)
		}
		if in {
			pf.ValidateRules = getFieldValidateRules(fld)
		}
		if strings.HasPrefix(fld.Type.String(), "[]") && fld.Type.String() != "[]byte" {
			pf.Repeated = true
		}

		t := getProtoTag(fld, -1)
		if t > 0 {
			pf.Tag = t
			usedTags[t] = true
		}
		results = append(results, fieldInfo{field: fld, pf: pf})
	}
	return results
}

func (e *Generator) buildProtoEdges(n *entgen.Type, in bool) []fieldInfo {
	var results []fieldInfo
	for _, edge := range n.Edges {
		if isProtoExclude(edge, in) {
			continue
		}

		if isProtoMessage(edge, in) {
			typeName := edge.Type.Name
			if in {
				typeName += "Input"
			}
			pf := &PbField{
				Name:     protoEdgeFieldName(edge, in),
				Type:     typeName,
				Repeated: !edge.Unique,
			}
			results = append(results, fieldInfo{edge: edge, pf: pf})
		} else {
			if isProtoID(edge, in) || (edgeHasFK(edge) && !hasField(n.Fields, edgeField(edge, in))) {
				pf := &PbField{
					Name:     protoEdgeFieldName(edge, in),
					Type:     edgeProtoType(edge, in),
					Repeated: !edge.Unique,
				}

				results = append(results, fieldInfo{edge: edge, pf: pf})
			}
		}
	}
	return results
}

func (e *Generator) assignProtoTags(msg *PbMessage, allFields []fieldInfo, usedTags map[int]bool) {
	currentTag := 1
	for _, info := range allFields {
		if info.pf.Tag == 0 {
			for usedTags[currentTag] {
				currentTag++
			}
			info.pf.Tag = currentTag
			usedTags[currentTag] = true
		}
		msg.Fields = append(msg.Fields, info.pf)
	}
}

func (e *Generator) buildProtoEnum(n *entgen.Type, f *entgen.Field) *PbEnum {
	enumName := n.Name + f.StructField()
	pe := &PbEnum{
		Name: enumName,
	}

	vals := getEnumValues(f)
	if vals != nil {
		if f.Enums != nil {
			for _, enumItem := range f.Enums {
				if v, ok := vals[enumItem.Value]; ok {
					pe.Values = append(pe.Values, &PbEnumValue{
						Name:   strings.ToUpper(enumName) + "_" + enumItem.Value,
						Number: v,
					})
				}
			}
		} else {
			// Fallback sort by key
			var keys []string
			for k := range vals {
				keys = append(keys, k)
			}
			sort.Strings(keys)
			for _, k := range keys {
				pe.Values = append(pe.Values, &PbEnumValue{
					Name:   strings.ToUpper(enumName) + "_" + k,
					Number: vals[k],
				})
			}
		}
	} else {
		// Auto-generate from 0
		if f.Enums != nil {
			for i, enumItem := range f.Enums {
				pe.Values = append(pe.Values, &PbEnumValue{
					Name:   strings.ToUpper(enumName) + "_" + enumItem.Value,
					Number: int32(i),
				})
			}
		}
	}

	return pe
}

func (e *Generator) resolveProtoType(f *entgen.Field, nodeName string, file *PbFile) string {
	a := getFieldAnnotation(f)
	if a != nil && a.ProtoType != "" {
		return a.ProtoType
	}
	if f.IsEnum() {
		return nodeName + f.StructField()
	}
	t := f.Type.String()
	switch t {
	case "int", "int32":
		return "int32"
	case "int64", "uint64":
		return "int64"
	case "string":
		return "string"
	case "bool":
		return "bool"
	case "time.Time":
		file.AddImport("google/protobuf/timestamp.proto")
		return "google.protobuf.Timestamp"
	case "float64":
		return "double"
	case "float32":
		return "float"
	case "uuid.UUID":
		return "string"
	case "[]byte":
		return "bytes"
	case "[]string":
		return "string"
	default:
		// Check for json types or other
		return "string" // Fallback
	}
}
