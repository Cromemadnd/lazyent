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

		// 5. Proto Files
		if err := e.render(nil, "templates/proto.tmpl", filepath.Join(moduleRoot, e.protoOut, e.protoFileName), data); err != nil {
			return err
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
