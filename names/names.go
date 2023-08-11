// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

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
	AccessAnalyzerEndpointID             = "access-analyzer"
	AccountEndpointID                    = "account"
	ACMEndpointID                        = "acm"
	AuditManagerEndpointID               = "auditmanager"
	CleanRoomsEndpointID                 = "cleanrooms"
	CloudWatchLogsEndpointID             = "logs"
	ComprehendEndpointID                 = "comprehend"
	ComputeOptimizerEndpointID           = "computeoptimizer"
	DSEndpointID                         = "ds"
	EMRServerlessEndpointID              = "emrserverless"
	GlacierEndpointID                    = "glacier"
	IdentityStoreEndpointID              = "identitystore"
	Inspector2EndpointID                 = "inspector2"
	InternetMonitorEndpointID            = "internetmonitor"
	IVSChatEndpointID                    = "ivschat"
	KendraEndpointID                     = "kendra"
	KeyspacesEndpointID                  = "keyspaces"
	LambdaEndpointID                     = "lambda"
	MediaLiveEndpointID                  = "medialive"
	ObservabilityAccessManagerEndpointID = "oam"
	OpenSearchServerlessEndpointID       = "aoss"
	PipesEndpointID                      = "pipes"
	PricingEndpointID                    = "pricing"
	QLDBEndpointID                       = "qldb"
	ResourceExplorer2EndpointID          = "resource-explorer-2"
	RolesAnywhereEndpointID              = "rolesanywhere"
	Route53DomainsEndpointID             = "route53domains"
	SchedulerEndpointID                  = "scheduler"
	SESV2EndpointID                      = "sesv2"
	SSMEndpointID                        = "ssm"
	SSMContactsEndpointID                = "ssm-contacts"
	SSMIncidentsEndpointID               = "ssm-incidents"
	SWFEndpointID                        = "swf"
	TimestreamWriteEndpointID            = "ingest.timestream"
	TranscribeEndpointID                 = "transcribe"
	VPCLatticeEndpointID                 = "vpc-lattice"
	XRayEndpointID                       = "xray"
)

// Type ServiceDatum corresponds closely to columns in `names_data.csv` and are
// described in detail in README.md.
type ServiceDatum struct {
	Aliases            []string
	Brand              string
	DeprecatedEnvVar   string
	EnvVar             string
	GoV1ClientTypeName string
	GoV1Package        string
	GoV2Package        string
	HumanFriendly      string
	ProviderNameUpper  string
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

		if l[ColExclude] != "" {
			continue
		}

		if l[ColNotImplemented] != "" {
			continue
		}

		if l[ColProviderPackageActual] == "" && l[ColProviderPackageCorrect] == "" {
			continue
		}

		p := l[ColProviderPackageCorrect]

		if l[ColProviderPackageActual] != "" {
			p = l[ColProviderPackageActual]
		}

		serviceData[p] = &ServiceDatum{
			Brand:              l[ColBrand],
			DeprecatedEnvVar:   l[ColDeprecatedEnvVar],
			EnvVar:             l[ColEnvVar],
			GoV1ClientTypeName: l[ColGoV1ClientTypeName],
			GoV1Package:        l[ColGoV1Package],
			GoV2Package:        l[ColGoV2Package],
			HumanFriendly:      l[ColHumanFriendly],
			ProviderNameUpper:  l[ColProviderNameUpper],
		}

		a := []string{p}

		if l[ColAliases] != "" {
			a = append(a, strings.Split(l[ColAliases], ";")...)
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

func HumanFriendly(service string) (string, error) {
	if v, ok := serviceData[service]; ok {
		return v.HumanFriendly, nil
	}

	if s, err := ProviderPackageForAlias(service); err == nil {
		return HumanFriendly(s)
	}

	return "", fmt.Errorf("no service data found for %s", service)
}

func AWSGoPackage(providerPackage string, version int) (string, error) {
	switch version {
	case 1:
		return AWSGoV1Package(providerPackage)
	case 2:
		return AWSGoV2Package(providerPackage)
	default:
		return "", fmt.Errorf("unsupported AWS SDK Go version: %d", version)
	}
}

func AWSGoV1Package(providerPackage string) (string, error) {
	if v, ok := serviceData[providerPackage]; ok {
		return v.GoV1Package, nil
	}

	return "", fmt.Errorf("getting AWS SDK Go v1 package, %s not found", providerPackage)
}

func AWSGoV2Package(providerPackage string) (string, error) {
	if v, ok := serviceData[providerPackage]; ok {
		return v.GoV2Package, nil
	}

	return "", fmt.Errorf("getting AWS SDK Go v2 package, %s not found", providerPackage)
}

func AWSGoClientTypeName(providerPackage string, version int) (string, error) {
	switch version {
	case 1:
		return AWSGoV1ClientTypeName(providerPackage)
	case 2:
		return "Client", nil
	default:
		return "", fmt.Errorf("unsupported AWS SDK Go version: %d", version)
	}
}

func AWSGoV1ClientTypeName(providerPackage string) (string, error) {
	if v, ok := serviceData[providerPackage]; ok {
		return v.GoV1ClientTypeName, nil
	}

	return "", fmt.Errorf("getting AWS SDK Go v1 client type name, %s not found", providerPackage)
}
