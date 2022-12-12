//go:build generate
// +build generate

package main

import (
	_ "embed"
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/hashicorp/terraform-provider-aws/internal/generate/common"
	"github.com/hashicorp/terraform-provider-aws/names"
)

const (
	spFile        = `service_package_gen.go`
	spsFile       = `../../provider/service_packages_gen.go`
	namesDataFile = `../../../names/names_data.csv`
)

func main() {
	g := common.NewGenerator()

	data, err := common.ReadAllCSVData(namesDataFile)

	if err != nil {
		g.Fatalf("error reading %s: %s", namesDataFile, err.Error())
	}

	g.Infof("Generating per-service %s", filepath.Base(spFile))

	td := TemplateData{}

	for i, l := range data {
		if i < 1 { // no header
			continue
		}

		if l[names.ColProviderPackageActual] == "" && l[names.ColProviderPackageCorrect] == "" {
			continue
		}

		p := l[names.ColProviderPackageCorrect]

		if l[names.ColProviderPackageActual] != "" {
			p = l[names.ColProviderPackageActual]
		}

		dir := fmt.Sprintf("../../service/%s", p)

		if _, err := os.Stat(dir); err != nil {
			continue
		}

		if v, err := filepath.Glob(fmt.Sprintf("%s/*.go", dir)); err != nil || len(v) == 0 {
			continue
		}

		s := ServiceDatum{
			ProviderPackage: p,
		}

		if err := g.ApplyAndWriteTemplateGoFormat(fmt.Sprintf("../../service/%s/%s", p, spFile), "servicepackagedata", spdTmpl, s); err != nil {
			g.Fatalf("error generating %s service package data: %s", p, err.Error())
		}

		td.Services = append(td.Services, s)
	}

	sort.SliceStable(td.Services, func(i, j int) bool {
		return td.Services[i].ProviderPackage < td.Services[j].ProviderPackage
	})

	g.Infof("Generating %s", filepath.Base(spsFile))

	if err := g.ApplyAndWriteTemplateGoFormat(spsFile, "servicepackages", spsTmpl, td); err != nil {
		g.Fatalf("error generating service packages list: %s", err.Error())
	}
}

type ServiceDatum struct {
	ProviderPackage string
}

type TemplateData struct {
	Services []ServiceDatum
}

//go:embed spd.tmpl
var spdTmpl string

//go:embed sps.tmpl
var spsTmpl string
