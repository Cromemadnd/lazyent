package lazyent

import (
	"entgo.io/ent/entc"
	"entgo.io/ent/entc/gen"

	lg "github.com/Cromemadnd/lazyent/internal/gen"
)

// Extension 实现 entc.Extension 接口
type Extension struct {
	conf Config
}

// Config 定义了 Extension 的必填配置参数
type Config struct {
	ProtoOut     string // Proto 文件输出目录 (e.g. "api/v1")
	ProtoPackage string // Proto 文件中的 package 定义
	GoPackage    string // Protobuf 中的 go_package 定义
	BizOut       string // Biz 层输出目录 (e.g. "internal/biz")
	ServiceOut   string // Service 层输出目录 (e.g. "internal/service")
	DataOut      string // Data 层输出目录 (e.g. "internal/data")
	SingleFile   bool   // 是否启用单文件生成模式

	// Optional configuration
	BizBaseFileName    string
	BizEntityFileName  string
	SvcMapperFileName  string
	DataMapperFileName string
	ProtoFileName      string
}

func NewExtension(cfg Config) *Extension {
	e := &Extension{
		conf: cfg,
	}
	return e
}

func (e *Extension) Annotations() []entc.Annotation {
	return nil
}

func (e *Extension) Options() []entc.Option {
	return nil
}

func (e *Extension) Templates() []*gen.Template {
	return nil
}

func (e *Extension) Hooks() []gen.Hook {
	return []gen.Hook{
		e.GenerateFiles,
	}
}

func (e *Extension) GenerateFiles(next gen.Generator) gen.Generator {
	return gen.GenerateFunc(func(g *gen.Graph) error {
		// 1. 识别并标记所有的虚拟元素
		vNodes := make(map[*gen.Type]bool)
		vFields := make(map[*gen.Field]bool)
		vEdges := make(map[*gen.Edge]bool)

		for _, n := range g.Nodes {
			// 检查是否标记了 entsql.Skip()
			isVirtualNode := false
			if ant, ok := n.Annotations["EntSQL"]; ok {
				// 检查 entsql.Annotation
				if m, ok := ant.(map[string]interface{}); ok {
					if skip, ok := m["skip"].(bool); ok && skip {
						isVirtualNode = true
					}
				}
			}

			if isVirtualNode {
				vNodes[n] = true
			}
		}

		// 1.1 传播：标记所有与虚拟节点相关的 Field/Edge
		for _, n := range g.Nodes {
			isVNode := vNodes[n]
			for _, f := range n.Fields {
				if isVNode || lg.IsVirtual(f) {
					vFields[f] = true
				}
			}
			for _, ed := range n.Edges {
				// 如果 1. 节点本身是虚拟的 2. 边被标记为虚拟 3. 目标节点是虚拟的
				// 则该边必须被视为虚拟，从 Ent 中清理掉
				if isVNode || lg.IsVirtual(ed) || vNodes[ed.Type] {
					vEdges[ed] = true
					if ed.Ref != nil {
						vEdges[ed.Ref] = true
					}
				}
			}
		}

		// 如果没有任何虚拟元素，直接按标准流程走
		if len(vFields) == 0 && len(vEdges) == 0 && len(vNodes) == 0 {
			if err := next.Generate(g); err != nil {
				return err
			}
			return e.GenerateLazyEntFiles(g)
		}

		// 2. 准备一份“干净”的 Graph 副本
		originalNodes := g.Nodes
		type nodeBackup struct {
			fields  []*gen.Field
			edges   []*gen.Edge
			fks     []*gen.ForeignKey
			idxs    []*gen.Index
			deleted bool
		}
		backups := make(map[string]nodeBackup)
		var regNodes []*gen.Type

		for _, n := range g.Nodes {
			if vNodes[n] {
				backups[n.Name] = nodeBackup{
					fields:  n.Fields,
					edges:   n.Edges,
					fks:     n.ForeignKeys,
					idxs:    n.Indexes,
					deleted: true,
				}
				continue
			}
			regNodes = append(regNodes, n)

			var regFields []*gen.Field
			for _, f := range n.Fields {
				if !vFields[f] {
					regFields = append(regFields, f)
				}
			}
			var regEdges []*gen.Edge
			for _, ed := range n.Edges {
				if !vEdges[ed] {
					regEdges = append(regEdges, ed)
				} else {
					if ed.Ref != nil {
						ed.Ref.Ref = nil
					}
				}
			}
			var regFKs []*gen.ForeignKey
			for _, fk := range n.ForeignKeys {
				if (fk.Field == nil || !vFields[fk.Field]) && (fk.Edge == nil || !vEdges[fk.Edge]) {
					regFKs = append(regFKs, fk)
				}
			}
			var regIdxs []*gen.Index
			for _, idx := range n.Indexes {
				isV := false
				for _, col := range idx.Columns {
					for f := range vFields {
						if f.Type.String() != "" && f.Name == col {
							isV = true
							break
						}
					}
					if isV {
						break
					}
				}
				if !isV {
					regIdxs = append(regIdxs, idx)
				}
			}

			backups[n.Name] = nodeBackup{
				fields: n.Fields,
				edges:  n.Edges,
				fks:    n.ForeignKeys,
				idxs:   n.Indexes,
			}

			n.Fields = regFields
			n.Edges = regEdges
			n.ForeignKeys = regFKs
			n.Indexes = regIdxs
		}

		g.Nodes = regNodes

		// 定义恢复函数
		restore := func() {
			g.Nodes = originalNodes
			for _, n := range g.Nodes {
				if b, ok := backups[n.Name]; ok {
					n.Fields = b.fields
					n.Edges = b.edges
					n.ForeignKeys = b.fks
					n.Indexes = b.idxs
				}
			}
			// 恢复 Ref 链接
			for ed := range vEdges {
				if ed.Ref != nil {
					ed.Ref.Ref = ed
				}
			}
		}

		// 4. 无论生成是否成功，立即恢复 g 的原始状态 (通过 defer 保证即使 panic 也能恢复)
		defer restore()

		// 3. 执行 Ent 标准生成
		// 此时的 g 中已经没有虚拟字段和关联了
		if err := next.Generate(g); err != nil {
			return err
		}

		// 手动显式恢复，确保 LazyEnt 能看到完整 Graph
		restore()

		// 5. 执行 LazyEnt 自己的文件生成逻辑（使用完整的 Graph）
		return e.GenerateLazyEntFiles(g)
	})
}

func (e *Extension) GenerateLazyEntFiles(g *gen.Graph) error {
	iConf := lg.Config{
		ProtoOut:           e.conf.ProtoOut,
		ProtoPackage:       e.conf.ProtoPackage,
		GoPackage:          e.conf.GoPackage,
		BizOut:             e.conf.BizOut,
		ServiceOut:         e.conf.ServiceOut,
		DataOut:            e.conf.DataOut,
		SingleFile:         e.conf.SingleFile,
		BizBaseFileName:    e.conf.BizBaseFileName,
		BizEntityFileName:  e.conf.BizEntityFileName,
		SvcMapperFileName:  e.conf.SvcMapperFileName,
		DataMapperFileName: e.conf.DataMapperFileName,
		ProtoFileName:      e.conf.ProtoFileName,
	}
	return lg.Generate(iConf, g)
}
