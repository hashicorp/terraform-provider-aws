package internal

import (
	"embed"
	"sync"

	"github.com/YakDriver/smarterr"
)

//go:embed service/smarterr.hcl
//go:embed service/*/smarterr.hcl
var SmarterrFS embed.FS

var smarterrInitOnce sync.Once

func init() {
	smarterrInitOnce.Do(func() {
		smarterr.SetLogger(smarterr.TFLogLogger{})
		smarterr.SetFS(&smarterr.WrappedFS{FS: &SmarterrFS}, "internal")
	})
}
