//go:build generate
// +build generate

package main

import (
	_ "embed"
	"fmt"
	"sort"

	"github.com/hashicorp/terraform-provider-aws/internal/generate/common"
	"github.com/hashicorp/terraform-provider-aws/names"
)

type ServiceDatum struct {
	SDKVersion          string
	GoV1Package         string
	GoV2Package         string
	GoV2PackageOverride string
	ProviderNameUpper   string
	ClientTypeName      string
	ProviderPackage     string
}

type TemplateData struct {
	Services []ServiceDatum
}

func main() {
	const (
		filename      = `awsclient_gen.go`
		namesDataFile = "../../names/names_data.csv"
	)
	g := common.NewGenerator()

	g.Infof("Generating internal/conns/%s", filename)

	data, err := common.ReadAllCSVData(namesDataFile)

	if err != nil {
		g.Fatalf("error reading %s: %s", namesDataFile, err)
	}

	td := TemplateData{}

	for i, l := range data {
		if i < 1 { // skip header
			continue
		}

		if l[names.ColExclude] != "" {
			continue
		}

		if l[names.ColProviderPackageActual] == "" && l[names.ColProviderPackageCorrect] == "" {
			continue
		}

		if l[names.ColClientSDKV1] != "" {
			td.Services = append(td.Services, ServiceDatum{
				ProviderNameUpper: l[names.ColProviderNameUpper],
				SDKVersion:        "1",
				GoV1Package:       l[names.ColGoV1Package],
				GoV2Package:       l[names.ColGoV2Package],
				ClientTypeName:    l[names.ColGoV1ClientTypeName],
				ProviderPackage:   l[names.ColProviderPackageCorrect],
			})
		}
		if l[names.ColClientSDKV2] != "" {
			sd := ServiceDatum{
				ProviderNameUpper: l[names.ColProviderNameUpper],
				SDKVersion:        "2",
				GoV1Package:       l[names.ColGoV1Package],
				GoV2Package:       l[names.ColGoV2Package],
				ClientTypeName:    "Client",
				ProviderPackage:   l[names.ColProviderPackageCorrect],
			}
			if l[names.ColClientSDKV1] != "" {
				// Use `sdkv2` instead of `v2` to prevent collisions with e.g., `elbv2`.
				sd.GoV2PackageOverride = fmt.Sprintf("%s_sdkv2", l[names.ColGoV2Package])
				sd.SDKVersion = "1,2"
			}
			td.Services = append(td.Services, sd)
		}
	}

	sort.SliceStable(td.Services, func(i, j int) bool {
		return td.Services[i].ProviderNameUpper < td.Services[j].ProviderNameUpper
	})

	d := g.NewGoFileDestination(filename)

	if err := d.WriteTemplate("awsclient", tmpl, td); err != nil {
		g.Fatalf("generating file (%s): %s", filename, err)
	}

	if err := d.Write(); err != nil {
		g.Fatalf("generating file (%s): %s", filename, err)
	}
}

//go:embed file.tmpl
var tmpl string
