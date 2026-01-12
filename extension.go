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
	ProtoOut       string         // Proto 文件输出目录 (e.g. "api/v1")
	ProtoPackage   string         // Proto 文件中的 package 定义
	GoPackage      string         // Protobuf 中的 go_package 定义
	BizOut         string         // Biz 层输出目录 (e.g. "internal/biz")
	ServiceOut     string         // Service 层输出目录 (e.g. "internal/service")
	DataOut        string         // Data 层输出目录 (e.g. "internal/data")
	SingleFile     bool           // 是否启用单文件生成模式
	ProtoValidator ProtoValidator // Proto 校验器类型

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

func (e *Extension) Hooks() []gen.Hook {
	return []gen.Hook{
		e.GenerateFiles,
	}
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

func (e *Extension) GenerateFiles(next gen.Generator) gen.Generator {
	return gen.GenerateFunc(func(g *gen.Graph) error {
		if err := next.Generate(g); err != nil {
			return err
		}
		// Convert config to internal config
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
			ProtoValidator:     e.conf.ProtoValidator,
		}

		return lg.Generate(iConf, g)
	})
}
