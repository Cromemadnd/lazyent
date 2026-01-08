package lazyent

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"entgo.io/ent/entc/gen"
	"golang.org/x/mod/modfile"
)

func (e *Extension) GenerateFiles(next gen.Generator) gen.Generator {
	return gen.GenerateFunc(func(g *gen.Graph) error {
		if err := next.Generate(g); err != nil {
			return err
		}
		return e.generate(g)
	})
}

func (e *Extension) generate(g *gen.Graph) error {
	modPath, moduleRoot, err := getModulePath()
	if err != nil {
		return fmt.Errorf("failed to get module path: %w", err)
	}

	// default fallback for protoPkg
	protoPkg := e.protoPkg
	if protoPkg == "" {
		protoPkg = g.Package
		if protoPkg == "ent" {
			protoPkg = "api.v1"
		}
	}

	goPkgPrefix := e.goPkgPrefix
	if goPkgPrefix == "" {
		goPkgPrefix = modPath
	}

	goPackage := fmt.Sprintf("%s/%s;%sv1", goPkgPrefix, strings.ReplaceAll(e.protoOut, "\\", "/"), g.Package)
	// modPath includes the module name. e.bizOut is relative to module root.
	// We need to construct the Go import path.
	// Imports should be "github.com/my/mod/internal/biz".
	// e.bizOut e.g. "app/user/internal/biz".
	bizPackage := fmt.Sprintf("%s/%s", modPath, strings.ReplaceAll(e.bizOut, "\\", "/"))
	apiPackage := fmt.Sprintf("%s/%s", modPath, strings.ReplaceAll(e.protoOut, "\\", "/"))
	entPackage := g.Config.Package

	// Prepare data for all nodes
	var allNodes []interface{}
	for _, n := range g.Nodes {
		var enums []*gen.Field
		for _, f := range n.Fields {
			if f.IsEnum() {
				enums = append(enums, f)
			}
		}
		nodeData := map[string]interface{}{
			"Name":   n.Name,
			"Fields": n.Fields,
			"Edges":  n.Edges,
			"Enums":  enums,
		}
		allNodes = append(allNodes, nodeData)
	}

	commonData := map[string]interface{}{
		"Package":    protoPkg,
		"Module":     modPath,
		"BizPackage": bizPackage,
		"ApiPackage": apiPackage,
		"GoPackage":  goPackage,
		"EntPackage": entPackage,
	}

	if e.singleFile {
		// Single file generation
		data := make(map[string]interface{})
		for k, v := range commonData {
			data[k] = v
		}
		data["Nodes"] = allNodes

		// 1. Biz Base Entities
		if err := e.render(nil, "templates/base.tmpl", filepath.Join(moduleRoot, e.bizOut, e.bizBaseFileName), data); err != nil {
			return err
		}

		// 2. Biz Entities (Scaffold) - Check if exists first
		scaffoldPath := filepath.Join(moduleRoot, e.bizOut, e.bizEntityFileName)
		if _, err := os.Stat(scaffoldPath); os.IsNotExist(err) {
			if err := e.render(nil, "templates/scaffold.tmpl", scaffoldPath, data); err != nil {
				return err
			}
		}

		// 3. Service Mappers
		if err := e.render(nil, "templates/mapper.tmpl", filepath.Join(moduleRoot, e.serviceOut, e.svcMapperFileName), data); err != nil {
			return err
		}

		// 4. Data Mappers (Ent)
		if err := e.render(nil, "templates/data.tmpl", filepath.Join(moduleRoot, e.dataOut, e.dataMapperFileName), data); err != nil {
			return err
		}

		// Proto files are usually per-service/message, but user asked for consolidation of Go files.
		// Proto generation currently uses separate files. The user request example listing implied separate proto files?
		// Actually user listing: "data\mappers_gen.go, biz\entities_base_gen.go..."
		// It didn't mention proto consolidation.
		// However, my current proto template is per-node.
		// If I keep it per-node, I need to iterate.
		// Let's create individual proto files even in singleFile mode for now, unless requested otherwise.
		// BUT wait, if I change the template to Expect `.Nodes` list, then I must iterate and generate logic.
		// Existing proto.tmpl generates `message Post {}`.
		// If I wrap it in range, it generates multiple messages in one proto file.
		// That is VALID proto.
		// So I CAN consolidate proto files if I want.
		// But let's look at the user request again.
		// "Enable: separate files... Disable: single files..."
		// It implies consistent behavior.
		// However, consolidation of Proto files into one `api.proto` might conflict with imports if not careful.
		// Let's stick to generating separate Proto files for now as it wasn't strictly asked to consolidate protos (only Go files listed).
		// Wait, if I change templates to iterate `.Nodes`, then I CANNOT use the same template for single file generation AND multiple file generation unless I always pass a list.
		// So if I am in singleFile mode, I pass all nodes.
		// For Proto, if I want separate files, I must call render individually.

		// Let's iterate for Proto files individually regardless of mode OR support separate proto files.
		// Since user didn't explicitly ask for proto consolidation, I will generate separate proto files.
		// To do this reuse the "allNodes" but render individually.
		// BUT the template will be updated to expect `Nodes` (plural).
		// So I pass `Nodes: [n]` (one element).

		for _, nd := range allNodes {
			pData := make(map[string]interface{})
			for k, v := range commonData {
				pData[k] = v
			}
			pData["Nodes"] = []interface{}{nd} // Single node list
			// Name is needed for filename construction in loop
			ndMap := nd.(map[string]interface{})
			name := ndMap["Name"].(string)
			if err := e.render(nil, "templates/proto.tmpl", filepath.Join(moduleRoot, e.protoOut, strings.ToLower(name)+".proto"), pData); err != nil {
				return err
			}
		}

	} else {
		// Multiple files generation (Legacy/Split mode)
		for _, nd := range allNodes {
			data := make(map[string]interface{})
			for k, v := range commonData {
				data[k] = v
			}
			data["Nodes"] = []interface{}{nd}
			ndMap := nd.(map[string]interface{})
			name := ndMap["Name"].(string)

			// 1. Biz Base
			if err := e.render(nil, "templates/base.tmpl", filepath.Join(moduleRoot, e.bizOut, strings.ToLower(name)+"_base_gen.go"), data); err != nil {
				return err
			}

			// 2. Biz Scaffold
			scaffoldPath := filepath.Join(moduleRoot, e.bizOut, strings.ToLower(name)+".go")
			if _, err := os.Stat(scaffoldPath); os.IsNotExist(err) {
				if err := e.render(nil, "templates/scaffold.tmpl", scaffoldPath, data); err != nil {
					return err
				}
			}

			// 3. Proto
			if err := e.render(nil, "templates/proto.tmpl", filepath.Join(moduleRoot, e.protoOut, strings.ToLower(name)+".proto"), data); err != nil {
				return err
			}

			// 4. Service Mapper
			if err := e.render(nil, "templates/mapper.tmpl", filepath.Join(moduleRoot, e.serviceOut, strings.ToLower(name)+"_mapper.go"), data); err != nil {
				return err
			}

			// 5. Data Mapper
			if err := e.render(nil, "templates/data.tmpl", filepath.Join(moduleRoot, e.dataOut, strings.ToLower(name)+"_ent.go"), data); err != nil {
				return err
			}
		}
	}

	return nil
}

func (e *Extension) render(n *gen.Type, tmplName string, targetPath string, data interface{}) error {
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

	return os.WriteFile(targetPath, buf.Bytes(), 0644)
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
