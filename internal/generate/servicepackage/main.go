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

func main() {
	const (
		spFile        = `service_package_gen.go`
		spsFile       = `../../provider/service_packages_gen.go`
		namesDataFile = `../../../names/names_data.csv`
	)
	g := common.NewGenerator()

	data, err := common.ReadAllCSVData(namesDataFile)

	if err != nil {
		g.Fatalf("error reading %s: %s", namesDataFile, err)
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
		d := g.NewGoFileDestination(fmt.Sprintf("../../service/%s/%s", p, spFile))

		if err := d.WriteTemplate("servicepackagedata", spdTmpl, s); err != nil {
			g.Fatalf("error generating %s service package data: %s", p, err)
		}

		if err := d.Write(); err != nil {
			g.Fatalf("generating file (%s): %s", spFile, err)
		}

		td.Services = append(td.Services, s)
	}

	sort.SliceStable(td.Services, func(i, j int) bool {
		return td.Services[i].ProviderPackage < td.Services[j].ProviderPackage
	})

	g.Infof("Generating %s", filepath.Base(spsFile))

	d := g.NewGoFileDestination(spsFile)

	if err := d.WriteTemplate("servicepackages", spsTmpl, td); err != nil {
		g.Fatalf("error generating service packages list: %s", err)
	}

	if err := d.Write(); err != nil {
		g.Fatalf("generating file (%s): %s", spsFile, err)
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
