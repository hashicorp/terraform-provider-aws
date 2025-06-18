package internal

import (
	"embed"
	"sync"

	"github.com/YakDriver/smarterr"
	"github.com/YakDriver/smarterr/filesystem"
)

//go:embed service/smarterr.hcl
//go:embed service/*/smarterr.hcl
var SmarterrFS embed.FS

var smarterrInitOnce sync.Once

func init() {
	smarterrInitOnce.Do(func() {
		smarterr.SetFS(&filesystem.WrappedFS{FS: &SmarterrFS}, "internal")
	})
}
