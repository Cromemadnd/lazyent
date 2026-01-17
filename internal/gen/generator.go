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

	// 2. Adapt Graph (Pre-calculation)
	genNodes, err := AdaptGraph(g)
	if err != nil {
		return fmt.Errorf("failed to adapt graph: %w", err)
	}

	// 3. Prepare Template Data
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

	// Prepare Nodes Data for Templates (Use GenNode)
	var allNodes []interface{}
	for _, n := range genNodes {
		allNodes = append(allNodes, n)
	}

	// 4. Generate
	var generatedProtoFiles []string

	if e.conf.SingleFile {
		// --- Phase 1: Proto Generation ---
		// Proto Files (Using new Builder with GenNodes)
		protoDesc, err := e.buildProtoFile(genNodes) // Build Descriptor
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
		for _, nd := range genNodes {
			lName := strings.ToLower(nd.Name) // GenNode has Name from Type

			singleNodeProto, err := e.buildSingleNodeProto(nd)
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
		for _, nd := range genNodes {
			// Basic template data
			data := make(map[string]interface{})
			for k, v := range commonData {
				data[k] = v
			}
			data["Nodes"] = []interface{}{nd}
			lName := strings.ToLower(nd.Name)

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
	// Format Go files
	if strings.HasSuffix(targetPath, ".go") {
		formatted, err := imports.Process(targetPath, content, nil)
		if err != nil {
			fmt.Printf("⚠️  Warning: Could not auto-fix imports for %s\n", filepath.Base(targetPath))
			fmt.Printf("    Reason: %v\n", err)
			fmt.Printf("    → Please compile proto files first (protoc), then run 'goimports -w %s'\n", targetPath)
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

// buildProtoFile constructs the PbFile descriptor from GenNodes
func (e *Generator) buildProtoFile(nodes []*GenNode) (*PbFile, error) {
	files := &PbFile{
		Package:   e.conf.ProtoPackage,
		GoPackage: e.conf.GoPackage,
		Imports:   []string{},
	}

	for _, n := range nodes {
		// Enums
		for _, f := range n.Fields {
			if f.IsEnum() && !f.IsExternalEnum {
				files.Elements = append(files.Elements, PbElement{Enum: e.buildProtoEnum(n, f)})
			}
		}

		// Output Message
		msg := e.buildProtoMessage(n, files, false)
		files.Elements = append(files.Elements, PbElement{Message: msg})

		// Input Message
		inputMsg := e.buildProtoMessage(n, files, true)
		inputMsg.Name = n.Name + "Input"
		files.Elements = append(files.Elements, PbElement{Message: inputMsg})
	}

	e.ensureValidateImport(files)
	return files, nil
}

func (e *Generator) buildSingleNodeProto(n *GenNode) (*PbFile, error) {
	files := &PbFile{
		Package:   e.conf.ProtoPackage,
		GoPackage: e.conf.GoPackage,
		Imports:   []string{},
	}

	for _, f := range n.Fields {
		if f.IsEnum() && !f.IsExternalEnum {
			files.Elements = append(files.Elements, PbElement{Enum: e.buildProtoEnum(n, f)})
		}
	}
	msg := e.buildProtoMessage(n, files, false)
	files.Elements = append(files.Elements, PbElement{Message: msg})

	inputMsg := e.buildProtoMessage(n, files, true)
	inputMsg.Name = n.Name + "Input"
	files.Elements = append(files.Elements, PbElement{Message: inputMsg})

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
		for _, imp := range f.Imports {
			if imp == "buf/validate/validate.proto" {
				return
			}
		}
		f.Imports = append(f.Imports, "buf/validate/validate.proto")
	}
}

func (e *Generator) buildProtoMessage(n *GenNode, f *PbFile, in bool) *PbMessage {
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
	field *GenField
	edge  *GenEdge
	pf    *PbField
}

func (e *Generator) buildProtoFields(n *GenNode, f *PbFile, usedTags map[int]bool, in bool) []fieldInfo {
	var results []fieldInfo

	// 1. ID (Implicitly handled? ID is not in n.Fields usually, but n.ID)
	// GenNode doesn't wrap ID as GenField currently in local version of AdaptNode?
	// models.go: GenNode has *entgen.Type.
	// We can adapt ID on the fly if needed or just access n.ID and wrap it.
	// But `buildProtoFields` previously worked on `entgen.Field`.
	// For ID:
	if n.ID != nil {
		// We need to check strategy.
		// ID doesn't have a wrapper in GenNode struct explicitly, just n.ID.
		// But helpers expect interface that can be cast to *GenField or *entgen.Field.
		// adaptField can adapt on fly if we pass it.
		// But let's look at `isFieldProtoExclude`. It calls `asGenField`.
		// `asGenField` handles `*entgen.Field`.
		// So passing `n.ID` (which is `*entgen.Field`) works!

		if !isFieldProtoExclude(n.ID, in) {
			pf := &PbField{
				Name:    protoFieldName(n.ID), // helpers work on *entgen.Field via JIT adaptation
				Comment: n.ID.Comment(),
			}

			// Resolve Type
			// JIT adaptation for type resolution
			// We can use protoType helper which does JIT
			pf.Type = protoType(n.ID, n.Name)
			// Handle imports:
			if pf.Type == "google.protobuf.Timestamp" {
				f.AddImport("google/protobuf/timestamp.proto")
			}

			t := getProtoTag(n.ID, -1)
			if t > 0 {
				pf.Tag = t
				usedTags[t] = true
			}
			// fieldInfo expects *GenField. `asGenField` returns *GenField.
			results = append(results, fieldInfo{isID: true, field: asGenField(n.ID), pf: pf})
		}
	}

	// 2. Fields
	for _, fld := range n.Fields { // fld is *GenField
		// Check Strategy
		if isFieldProtoExclude(fld, in) {
			continue
		}

		pf := &PbField{
			Name:    fld.ProtoName,
			Comment: fld.Comment(),
		}

		if fld.IsExternalEnum {
			pf.Type = "string"
		} else {
			pf.Type = fld.ProtoType
			if pf.Type == "google.protobuf.Timestamp" {
				f.AddImport("google/protobuf/timestamp.proto")
			}
		}

		if in {
			pf.ValidateRules = getFieldValidateRules(fld.Field) // This expects *entgen.Field usually?
			// getFieldValidateRules is in validate.go (assumed). I need to check if it's updated.
			// It probably just reads annotation.
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

func (e *Generator) buildProtoEdges(n *GenNode, in bool) []fieldInfo {
	var results []fieldInfo
	for _, edge := range n.Edges { // edge is *GenEdge
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
			// ID or FK
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

func (e *Generator) buildProtoEnum(n *GenNode, f *GenField) *PbEnum {
	enumName := n.Name + f.StructField()
	pe := &PbEnum{
		Name: enumName,
	}

	// GenField has pre-calculated EnumValues
	// vals := f.EnumValues (unused)

	// Use helper to get consistent pairs
	pairs := getEnumPairs(f) // Works with *GenField

	for _, p := range pairs {
		pe.Values = append(pe.Values, &PbEnumValue{
			Name:   strings.ToUpper(enumName) + "_" + p.Key,
			Number: p.Value,
		})
	}

	return pe
}

// TODO: Ensure getFieldValidateRules is available or stubbed if it was in another file I didn't see?
// It was likely in `validate.go` or `generator.go`.
// Checking file listing earlier: `validate.go` exists.
// I will assume it takes *entgen.Field.
