// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

//go:build generate
// +build generate

package main

import (
	_ "embed"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"sort"
	"strings"

	"github.com/hashicorp/hcl/v2/hclsimple"
	"github.com/hashicorp/terraform-provider-aws/internal/generate/common"
	"github.com/hashicorp/terraform-provider-aws/names"
)

type ServiceDatum struct {
	ProviderPackage string
	HumanFriendly   string
	VpcLock         bool
	Parallelism     int
	Region          string
}

type TemplateData struct {
	Services []ServiceDatum
}

func main() {
	const (
		servicesAllFile   = `../../../.teamcity/components/generated/services_all.kt`
		namesDataFile     = "../../../names/names_data.csv"
		serviceConfigFile = "./acctest_services.hcl"
	)
	g := common.NewGenerator()

	g.Infof("Generating %s", strings.TrimPrefix(servicesAllFile, "../../../"))

	serviceConfigs, err := acctestConfigurations(serviceConfigFile)

	if err != nil {
		g.Fatalf("error reading %s: %s", serviceConfigFile, err)
	}

	data, err := common.ReadAllCSVData(namesDataFile)

	if err != nil {
		g.Fatalf("error reading %s: %s", namesDataFile, err)
	}

	td := TemplateData{}

	for i, l := range data {
		if i < 1 { // no header
			continue
		}

		if l[names.ColExclude] != "" {
			continue
		}

		if l[names.ColProviderPackageActual] == "" && l[names.ColProviderPackageCorrect] == "" {
			continue
		}

		p := l[names.ColProviderPackageCorrect]

		if l[names.ColProviderPackageActual] != "" {
			p = l[names.ColProviderPackageActual]
		}

		if _, err := os.Stat(fmt.Sprintf("../../service/%s", p)); err != nil || errors.Is(err, fs.ErrNotExist) {
			continue
		}

		sd := ServiceDatum{
			ProviderPackage: p,
			HumanFriendly:   l[names.ColHumanFriendly],
		}
		serviceConfig, ok := serviceConfigs[p]
		if ok {
			sd.VpcLock = serviceConfig.VpcLock
			sd.Parallelism = serviceConfig.Parallelism
			sd.Region = serviceConfig.Region
		}

		if serviceConfig.Skip {
			g.Infof("Skipping service %q...", p)
			continue
		}

		td.Services = append(td.Services, sd)
	}

	sort.SliceStable(td.Services, func(i, j int) bool {
		return td.Services[i].ProviderPackage < td.Services[j].ProviderPackage
	})

	d := g.NewUnformattedFileDestination(servicesAllFile)

	if err := d.WriteTemplate("teamcity", tmpl, td); err != nil {
		g.Fatalf("generating file (%s): %s", servicesAllFile, err)
	}

	if err := d.Write(); err != nil {
		g.Fatalf("generating file (%s): %s", servicesAllFile, err)
	}
}

//go:embed services.tmpl
var tmpl string

type acctestConfig struct {
	Services []acctestServiceConfig `hcl:"service,block"`
}

type acctestServiceConfig struct {
	Service     string `hcl:",label"`
	VpcLock     bool   `hcl:"vpc_lock,optional"`
	Parallelism int    `hcl:"parallelism,optional"`
	Skip        bool   `hcl:"skip,optional"`
	Region      string `hcl:"region,optional"`
}

func acctestConfigurations(filename string) (map[string]acctestServiceConfig, error) {
	var config acctestConfig

	err := decodeHclFile(filename, &config)
	if err != nil {
		return nil, err
	}

	result := make(map[string]acctestServiceConfig)

	for _, v := range config.Services {
		result[v.Service] = v
	}

	return result, nil
}

func decodeHclFile(filename string, target interface{}) error {
	f, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	b, err := io.ReadAll(f)
	if err != nil {
		return err
	}

	return hclsimple.Decode(filename, b, nil, target)
}
