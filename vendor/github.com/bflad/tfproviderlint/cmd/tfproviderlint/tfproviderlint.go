// The tfproviderlint command is a static checker for Terraform Providers.
//
// Each analyzer flag name is preceded by the analyzer name: -NAME.flag.
// In addition, the -NAME flag itself controls whether the
// diagnostics of that analyzer are displayed. (A disabled analyzer may yet
// be run if it is required by some other analyzer that is enabled.)
package main

import (
	"golang.org/x/tools/go/analysis/multichecker"

	"github.com/bflad/tfproviderlint/passes/AT001"
	"github.com/bflad/tfproviderlint/passes/AT002"
	"github.com/bflad/tfproviderlint/passes/AT003"
	"github.com/bflad/tfproviderlint/passes/AT004"
	"github.com/bflad/tfproviderlint/passes/R001"
	"github.com/bflad/tfproviderlint/passes/R002"
	"github.com/bflad/tfproviderlint/passes/R003"
	"github.com/bflad/tfproviderlint/passes/R004"
	"github.com/bflad/tfproviderlint/passes/S001"
	"github.com/bflad/tfproviderlint/passes/S002"
	"github.com/bflad/tfproviderlint/passes/S003"
	"github.com/bflad/tfproviderlint/passes/S004"
	"github.com/bflad/tfproviderlint/passes/S005"
	"github.com/bflad/tfproviderlint/passes/S006"
	"github.com/bflad/tfproviderlint/passes/S007"
	"github.com/bflad/tfproviderlint/passes/S008"
	"github.com/bflad/tfproviderlint/passes/S009"
	"github.com/bflad/tfproviderlint/passes/S010"
	"github.com/bflad/tfproviderlint/passes/S011"
	"github.com/bflad/tfproviderlint/passes/S012"
	"github.com/bflad/tfproviderlint/passes/S013"
	"github.com/bflad/tfproviderlint/passes/S014"
	"github.com/bflad/tfproviderlint/passes/S015"
	"github.com/bflad/tfproviderlint/passes/S016"
	"github.com/bflad/tfproviderlint/passes/S017"
	"github.com/bflad/tfproviderlint/passes/S018"
	"github.com/bflad/tfproviderlint/passes/S019"
)

func main() {
	multichecker.Main(
		AT001.Analyzer,
		AT002.Analyzer,
		AT003.Analyzer,
		AT004.Analyzer,
		R001.Analyzer,
		R002.Analyzer,
		R003.Analyzer,
		R004.Analyzer,
		S001.Analyzer,
		S002.Analyzer,
		S003.Analyzer,
		S004.Analyzer,
		S005.Analyzer,
		S006.Analyzer,
		S007.Analyzer,
		S008.Analyzer,
		S009.Analyzer,
		S010.Analyzer,
		S011.Analyzer,
		S012.Analyzer,
		S013.Analyzer,
		S014.Analyzer,
		S015.Analyzer,
		S016.Analyzer,
		S017.Analyzer,
		S018.Analyzer,
		S019.Analyzer,
	)
}
