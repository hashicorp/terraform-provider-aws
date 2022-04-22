// Package names provides constants for AWS service names that are used as keys
// for the endpoints slice in internal/conns/conns.go. The package also exposes
// access to data found in the names_data.csv file, which provides additional
// service-related name information.
//
// Consumers of the names package include the conns package
// (internal/conn/conns.go), the provider package
// (internal/provider/provider.go), generators, and the skaff tool.
//
// It is very important that information in the names_data.csv be exactly
// correct because the Terrform AWS Provider relies on the information to
// function correctly.

package names

import (
	_ "embed"
	"encoding/csv"
	"fmt"
	"log"
	"strings"
)

// This "should" be defined by the AWS Go SDK v2, but currently isn't.
const (
	Route53DomainsEndpointID = "route53domains"
)

// Type ServiceDatum corresponds closely to columns in `names_data.csv` and are
// described in detail in README.md.
type ServiceDatum struct {
	Aliases           []string
	Brand             string
	DeprecatedEnvVar  string
	EnvVar            string
	GoV1ClientName    string
	GoV1Package       string
	GoV2Package       string
	HumanFriendly     string
	ProviderNameUpper string
}

// serviceData key is the AWS provider service package
var serviceData map[string]*ServiceDatum

func init() {
	serviceData = make(map[string]*ServiceDatum)

	// Data from names_data.csv
	if err := readCSVIntoServiceData(); err != nil {
		log.Fatalf("reading CSV into service data: %s", err)
	}
}

const (
	// column indices of CSV
	//awsCLIV2Command         = 0
	//awsCLIV2CommandNoDashes = 1
	//goV1Package             = 2
	//goV2Package             = 3
	//providerPackageActual   = 4
	//providerPackageCorrect  = 5
	//splitPackageRealPackage = 6
	//aliases                 = 7
	//providerNameUpper       = 8
	//goV1ClientName          = 9
	//skipClientGenerate      = 10
	//sdkVersion              = 11
	//resourcePrefixActual    = 12
	//resourcePrefixCorrect   = 13
	//filePrefix              = 14
	//docPrefix               = 15
	//humanFriendly           = 16
	//brand                   = 17
	//exclude                 = 18
	//allowedSubcategory      = 19
	//deprecatedEnvVar        = 20
	//envVar                  = 21
	//note                    = 22
	goV1Package            = 2
	goV2Package            = 3
	providerPackageActual  = 4
	providerPackageCorrect = 5
	aliases                = 7
	providerNameUpper      = 8
	goV1ClientName         = 9
	humanFriendly          = 16
	brand                  = 17
	exclude                = 18
	deprecatedEnvVar       = 20
	envVar                 = 21
)

//go:embed names_data.csv
var namesData string

func readCSVIntoServiceData() error {
	// names_data.csv is dynamically embedded so changes, additions should be made
	// there also

	r := csv.NewReader(strings.NewReader(namesData))

	d, err := r.ReadAll()
	if err != nil {
		return fmt.Errorf("reading CSV into service data: %w", err)
	}

	for i, l := range d {
		if i < 1 { // omit header line
			continue
		}

		if l[exclude] != "" {
			continue
		}

		if l[providerPackageActual] == "" && l[providerPackageCorrect] == "" {
			continue
		}

		p := l[providerPackageCorrect]

		if l[providerPackageActual] != "" {
			p = l[providerPackageActual]
		}

		serviceData[p] = &ServiceDatum{
			Brand:             l[brand],
			DeprecatedEnvVar:  l[deprecatedEnvVar],
			EnvVar:            l[envVar],
			GoV1ClientName:    l[goV1ClientName],
			GoV1Package:       l[goV1Package],
			GoV2Package:       l[goV2Package],
			HumanFriendly:     l[humanFriendly],
			ProviderNameUpper: l[providerNameUpper],
		}

		a := []string{p}

		if l[aliases] != "" {
			a = append(a, strings.Split(l[aliases], ";")...)
		}

		serviceData[p].Aliases = a
	}

	return nil
}

func ProviderPackageForAlias(serviceAlias string) (string, error) {
	for k, v := range serviceData {
		for _, hclKey := range v.Aliases {
			if serviceAlias == hclKey {
				return k, nil
			}
		}
	}

	return "", fmt.Errorf("unable to find service for service alias %s", serviceAlias)
}

func ProviderPackages() []string {
	keys := make([]string, len(serviceData))

	i := 0
	for k := range serviceData {
		keys[i] = k
		i++
	}

	return keys
}

func Aliases() []string {
	keys := make([]string, 0)

	for _, v := range serviceData {
		keys = append(keys, v.Aliases...)
	}

	return keys
}

func ProviderNameUpper(service string) (string, error) {
	if v, ok := serviceData[service]; ok {
		return v.ProviderNameUpper, nil
	}

	return "", fmt.Errorf("no service data found for %s", service)
}

func DeprecatedEnvVar(service string) string {
	if v, ok := serviceData[service]; ok {
		return v.DeprecatedEnvVar
	}

	return ""
}

func EnvVar(service string) string {
	if v, ok := serviceData[service]; ok {
		return v.EnvVar
	}

	return ""
}

func FullHumanFriendly(service string) (string, error) {
	if v, ok := serviceData[service]; ok {
		if v.Brand == "" {
			return v.HumanFriendly, nil
		}

		return fmt.Sprintf("%s %s", v.Brand, v.HumanFriendly), nil
	}

	if s, err := ProviderPackageForAlias(service); err == nil {
		return FullHumanFriendly(s)
	}

	return "", fmt.Errorf("no service data found for %s", service)
}

func AWSGoV1Package(providerPackage string) (string, error) {
	if v, ok := serviceData[providerPackage]; ok {
		return v.GoV1Package, nil
	}

	return "", fmt.Errorf("getting AWS Go SDK v1 package, %s not found", providerPackage)
}

func AWSGoV1ClientName(providerPackage string) (string, error) {
	if v, ok := serviceData[providerPackage]; ok {
		return v.GoV1ClientName, nil
	}

	return "", fmt.Errorf("getting AWS Go SDK v1 client name, %s not found", providerPackage)
}
