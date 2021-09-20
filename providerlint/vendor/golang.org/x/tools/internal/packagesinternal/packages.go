// Package packagesinternal exposes internal-only fields from go/packages.
package packagesinternal

import (
	"golang.org/x/tools/internal/gocommand"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

var GetForTest = func(p interface{}) string { return "" }

var GetGoCmdRunner = func(config interface{}) *gocommand.Runner { return nil }

var SetGoCmdRunner = func(config interface{}, runner *gocommand.Runner) {}

var TypecheckCgo int

var SetModFlag = func(config interface{}, value string) {}
var SetModFile = func(config interface{}, value string) {}
