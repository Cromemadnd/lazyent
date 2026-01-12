package gen

import (
	"github.com/Cromemadnd/lazyent/internal/types"
)

// Config defines the configuration for the generator
type Config struct {
	ProtoOut     string // Proto Output directory (e.g. "api/v1")
	ProtoPackage string // Proto Package name
	GoPackage    string // Protobuf go_package
	BizOut       string // Biz Output directory
	ServiceOut   string // Service Output directory
	DataOut      string // Data Output directory
	SingleFile   bool   // Single file mode

	// Optional configuration (Internal use)
	BizBaseFileName    string
	BizEntityFileName  string
	SvcMapperFileName  string
	DataMapperFileName string
	ProtoFileName      string
	ProtoValidator     types.ProtoValidator
}
