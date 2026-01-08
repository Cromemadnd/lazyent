package lazyent

import (
	"embed"

	"entgo.io/ent/entc"
	"entgo.io/ent/entc/gen"
)

//go:embed templates/*
var templates embed.FS

// Extension 实现 entc.Extension 接口
type Extension struct {
	protoOut    string
	serviceOut  string
	bizOut      string
	dataOut     string
	protoPkg    string
	goPkgPrefix string

	// Code generation options
	singleFile         bool
	bizBaseFileName    string
	bizEntityFileName  string
	svcMapperFileName  string
	dataMapperFileName string
}

type Option func(*Extension)

// WithSingleFile 设置是否将所有代码生成到单个文件
func WithSingleFile(enable bool) Option {
	return func(e *Extension) {
		e.singleFile = enable
	}
}

// WithBizBaseFileName 设置 Biz Base 文件名 (SingleFile 模式下有效)
func WithBizBaseFileName(name string) Option {
	return func(e *Extension) {
		e.bizBaseFileName = name
	}
}

// WithBizEntityFileName 设置 Biz Entity 文件名 (SingleFile 模式下有效)
func WithBizEntityFileName(name string) Option {
	return func(e *Extension) {
		e.bizEntityFileName = name
	}
}

// WithServiceMapperFileName 设置 Service Mapper 文件名 (SingleFile 模式下有效)
func WithServiceMapperFileName(name string) Option {
	return func(e *Extension) {
		e.svcMapperFileName = name
	}
}

// WithDataMapperFileName 设置 Data Mapper 文件名 (SingleFile 模式下有效)
func WithDataMapperFileName(name string) Option {
	return func(e *Extension) {
		e.dataMapperFileName = name
	}
}

// WithProtoPackage 设置 Proto 包名 (e.g. "my.api.v1")
func WithProtoPackage(pkg string) Option {
	return func(e *Extension) {
		e.protoPkg = pkg
	}
}

// WithProtoOut 设置 Proto 文件输出目录
func WithProtoOut(path string) Option {
	return func(e *Extension) {
		e.protoOut = path
	}
}

// WithBizOut 设置 Biz 层输出目录
func WithBizOut(path string) Option {
	return func(e *Extension) {
		e.bizOut = path
	}
}

// WithDataOut 设置 Data 层输出目录
func WithDataOut(path string) Option {
	return func(e *Extension) {
		e.dataOut = path
	}
}

// WithServiceOut 设置 Service 层输出目录
func WithServiceOut(path string) Option {
	return func(e *Extension) {
		e.serviceOut = path
	}
}

// WithGoPackagePrefix 设置 Proto Go Package 前缀
func WithGoPackagePrefix(prefix string) Option {
	return func(e *Extension) {
		e.goPkgPrefix = prefix
	}
}

// Config 定义了 Extension 的必填配置参数
type Config struct {
	ProtoOut   string // Proto 文件输出目录 (e.g. "api/v1")
	BizOut     string // Biz 层输出目录 (e.g. "internal/biz")
	ServiceOut string // Service 层输出目录 (e.g. "internal/service")
	DataOut    string // Data 层输出目录 (e.g. "internal/data")
	SingleFile bool   // 是否启用单文件生成模式
}

func NewExtension(cfg Config, opts ...Option) *Extension {
	e := &Extension{
		protoOut:           cfg.ProtoOut,
		serviceOut:         cfg.ServiceOut,
		bizOut:             cfg.BizOut,
		dataOut:            cfg.DataOut,
		singleFile:         cfg.SingleFile,
		bizBaseFileName:    "entities_base_gen.go",
		bizEntityFileName:  "entities.go",
		svcMapperFileName:  "mappers_gen.go",
		dataMapperFileName: "mappers_gen.go",
	}
	for _, opt := range opts {
		opt(e)
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
