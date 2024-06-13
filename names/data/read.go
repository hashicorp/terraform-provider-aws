// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package data

import (
	_ "embed"
	"log"
	"strings"

	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/hashicorp/hcl/v2/hclparse"
)

type ServiceRecord []string

func (sr ServiceRecord) AWSCLIV2Command() string {
	return sr[colAWSCLIV2Command]
}

func (sr ServiceRecord) AWSCLIV2CommandNoDashes() string {
	return sr[colAWSCLIV2CommandNoDashes]
}

func (sr ServiceRecord) GoV1Package() string {
	return sr[colGoV1Package]
}

func (sr ServiceRecord) GoV2Package() string {
	return sr[colGoV2Package]
}

func (sr ServiceRecord) ProviderPackage() string {
	pkg := sr.ProviderPackageCorrect()
	if sr.ProviderPackageActual() != "" {
		pkg = sr.ProviderPackageActual()
	}
	return pkg
}

func (sr ServiceRecord) ProviderPackageActual() string {
	return sr[colProviderPackageActual]
}

func (sr ServiceRecord) ProviderPackageCorrect() string {
	return sr[colProviderPackageCorrect]
}

func (sr ServiceRecord) SplitPackageRealPackage() string {
	return sr[colSplitPackageRealPackage]
}

func (sr ServiceRecord) Aliases() []string {
	if sr[colAliases] == "" {
		return nil
	}
	return strings.Split(sr[colAliases], ";")
}

func (sr ServiceRecord) ProviderNameUpper() string {
	return sr[colProviderNameUpper]
}

func (sr ServiceRecord) GoV1ClientTypeName() string {
	return sr[colGoV1ClientTypeName]
}

func (sr ServiceRecord) SkipClientGenerate() bool {
	return sr[colSkipClientGenerate] != ""
}

func (sr ServiceRecord) ClientSDKV1() bool {
	return sr[colClientSDKV1] != ""
}

func (sr ServiceRecord) ClientSDKV2() bool {
	return sr[colClientSDKV2] != ""
}

// SDKVersion returns:
// * "1" if only SDK v1 is implemented
// * "2" if only SDK v2 is implemented
// * "1,2" if both are implemented
func (sr ServiceRecord) SDKVersion() string {
	if sr.ClientSDKV1() && sr.ClientSDKV2() {
		return "1,2"
	} else if sr.ClientSDKV1() {
		return "1"
	} else if sr.ClientSDKV2() {
		return "2"
	}
	return ""
}

func (sr ServiceRecord) ResourcePrefix() string {
	prefix := sr.ResourcePrefixCorrect()
	if sr.ResourcePrefixActual() != "" {
		prefix = sr.ResourcePrefixActual()
	}
	return prefix
}

func (sr ServiceRecord) ResourcePrefixActual() string {
	return sr[colResourcePrefixActual]
}

func (sr ServiceRecord) ResourcePrefixCorrect() string {
	return sr[colResourcePrefixCorrect]
}

func (sr ServiceRecord) FilePrefix() string {
	return sr[colFilePrefix]
}

func (sr ServiceRecord) DocPrefix() []string {
	if sr[colDocPrefix] == "" {
		return nil
	}
	return strings.Split(sr[colDocPrefix], ";")
}

func (sr ServiceRecord) HumanFriendly() string {
	return sr[colHumanFriendly]
}

func (sr ServiceRecord) Brand() string {
	return sr[colBrand]
}

func (sr ServiceRecord) Exclude() bool {
	return sr[colExclude] != ""
}

func (sr ServiceRecord) NotImplemented() bool {
	return sr[colNotImplemented] != ""
}

func (sr ServiceRecord) EndpointOnly() bool {
	return sr[colEndpointOnly] != ""
}

func (sr ServiceRecord) AllowedSubcategory() string {
	return sr[colAllowedSubcategory]
}

func (sr ServiceRecord) DeprecatedEnvVar() string {
	return sr[colDeprecatedEnvVar]
}

func (sr ServiceRecord) TFAWSEnvVar() string {
	return sr[colTFAWSEnvVar]
}

func (sr ServiceRecord) SDKID() string {
	return sr[colSDKID]
}

func (sr ServiceRecord) AWSServiceEnvVar() string {
	return "AWS_ENDPOINT_URL_" + strings.ReplaceAll(strings.ToUpper(sr.SDKID()), " ", "_")
}

func (sr ServiceRecord) AWSConfigParameter() string {
	return strings.ReplaceAll(strings.ToLower(sr.SDKID()), " ", "_")
}

func (sr ServiceRecord) EndpointAPICall() string {
	return sr[colEndpointAPICall]
}

func (sr ServiceRecord) EndpointAPIParams() string {
	return sr[colEndpointAPIParams]
}

func (sr ServiceRecord) EndpointOverrideRegion() string {
	return sr[colEndpointOverrideRegion]
}

func (sr ServiceRecord) Note() string {
	return sr[colNote]
}

func parseService(curr Service) ServiceRecord {
	record := make(ServiceRecord, colNote+1)

	// provider packages/label
	record[colProviderPackageActual] = curr.Label
	record[colProviderPackageCorrect] = curr.Label

	// cli_v2_command
	if len(curr.ServiceCli) > 0 {
		record[colAWSCLIV2Command] = curr.ServiceCli[0].AWSCLIV2Command
		record[colAWSCLIV2CommandNoDashes] = curr.ServiceCli[0].AWSCLIV2CommandNoDashes
	} else {
		record[colAWSCLIV2Command] = curr.Label
		record[colAWSCLIV2CommandNoDashes] = curr.Label
	}

	// go_packages
	if len(curr.ServiceGoPackages) > 0 {
		record[colGoV1Package] = curr.ServiceGoPackages[0].V1Package
		record[colGoV2Package] = curr.ServiceGoPackages[0].V2Package
	} else {
		record[colGoV1Package] = curr.Label
		record[colGoV2Package] = curr.Label
	}

	// sdk
	record[colSDKID] = curr.ServiceSDK.ID
	for _, i := range curr.ServiceSDK.Version {
		if i == 1 {
			record[colClientSDKV1] = "1"
		}
		if i == 2 {
			record[colClientSDKV2] = "2"
		}
	}

	// names
	record[colAliases] = strings.Join(curr.ServiceNames.Aliases, ";")
	record[colProviderNameUpper] = curr.ServiceNames.ProviderNameUpper
	record[colHumanFriendly] = curr.ServiceNames.HumanFriendly

	// client
	record[colGoV1ClientTypeName] = curr.ServiceClient.GoV1ClientTypeName
	if curr.ServiceClient.SkipClientGenerate {
		record[colSkipClientGenerate] = "x"
	} else {
		record[colSkipClientGenerate] = ""
	}

	// env_var
	record[colDeprecatedEnvVar] = curr.ServiceEnvVars.DeprecatedEnvVar
	record[colTFAWSEnvVar] = curr.ServiceEnvVars.TFAWSEnvVar

	// endpoint_info
	record[colEndpointAPICall] = curr.ServiceEndpoints.EndpointAPICall
	record[colEndpointAPIParams] = curr.ServiceEndpoints.EndpointAPIParams
	record[colEndpointOverrideRegion] = curr.ServiceEndpoints.EndpointRegionOverride
	if curr.ServiceEndpoints.EndpointOnly {
		record[colEndpointOnly] = "x"
	} else {
		record[colEndpointOnly] = ""
	}

	// resource_prefix
	record[colResourcePrefixActual] = curr.ServiceResourcePrefix.ResourcePrefixActual
	record[colResourcePrefixCorrect] = curr.ServiceResourcePrefix.ResourcePrefixCorrect

	// rest
	record[colSplitPackageRealPackage] = curr.ServiceSplitPackage
	record[colFilePrefix] = curr.FilePrefix
	record[colDocPrefix] = strings.Join(curr.DocPrefix, ";")
	record[colBrand] = curr.Brand
	if curr.Exclude {
		record[colExclude] = "x"
	} else {
		record[colExclude] = ""
	}
	if curr.NotImplemented {
		record[colNotImplemented] = "x"
	} else {
		record[colNotImplemented] = ""
	}
	if curr.AllowedSubcategory {
		record[colAllowedSubcategory] = "x"
	} else {
		record[colAllowedSubcategory] = ""
	}
	record[colNote] = curr.Note
	if len(curr.ServiceProviderPackageCorrect) > 0 {
		record[colProviderPackageCorrect] = curr.ServiceProviderPackageCorrect
	}

	return record
}

func ReadAllServiceData() (results []ServiceRecord, err error) {
	var decodedServiceList Services
	parser := hclparse.NewParser()
	toParse, parseErr := parser.ParseHCL(b, "names_data.hcl")
	if parseErr.HasErrors() {
		log.Fatal("Parser error : ", parseErr)
	}
	decodeErr := gohcl.DecodeBody(toParse.Body, nil, &decodedServiceList)
	if decodeErr.HasErrors() {
		log.Fatal("Decode error", decodeErr)
	}
	for _, curr := range decodedServiceList.ServiceList {
		if len(curr.SubService) > 0 {
			for _, sub := range curr.SubService {
				results = append(results, parseService(sub))
			}
		}
		results = append(results, parseService(curr))
	}

	return
}

type CLIV2Command struct {
	AWSCLIV2Command         string `hcl:"aws_cli_v2_command,attr"`
	AWSCLIV2CommandNoDashes string `hcl:"aws_cli_v2_command_no_dashes,attr"`
}

type GoPackages struct {
	V1Package string `hcl:"v1_package,optional"`
	V2Package string `hcl:"v2_package,optional"`
}

type ResourcePrefix struct {
	ResourcePrefixActual  string `hcl:"actual,attr"`
	ResourcePrefixCorrect string `hcl:"correct,attr"`
}

type SDK struct {
	ID      string `hcl:"id,attr"`
	Version []int  `hcl:"client_version,attr"`
}

type Names struct {
	Aliases           []string `hcl:"aliases,attr"`
	ProviderNameUpper string   `hcl:"provider_name_upper,attr"`
	HumanFriendly     string   `hcl:"human_friendly,attr"`
}

type ProviderPackage struct {
	Actual  string `hcl:"actual,attr"`
	Correct string `hcl:"correct,attr"`
}

type Client struct {
	GoV1ClientTypeName string `hcl:"go_v1_client_typename,attr"`
	SkipClientGenerate bool   `hcl:"skip_client_generate,attr"`
}

type EnvVar struct {
	DeprecatedEnvVar string `hcl:"deprecated_env_var,attr"`
	TFAWSEnvVar      string `hcl:"tf_aws_env_var,attr"`
}

type EndpointInfo struct {
	EndpointAPICall        string `hcl:"endpoint_api_call,attr"`
	EndpointAPIParams      string `hcl:"endpoint_api_params,attr"`
	EndpointRegionOverride string `hcl:"endpoint_region_override,attr"`
	EndpointOnly           bool   `hcl:"endpoint_only,attr"`
}

type Service struct {
	Label                 string         `hcl:"CLIV2Command,label"`
	ServiceCli            []CLIV2Command `hcl:"cli_v2_command,block"`
	ServiceGoPackages     []GoPackages   `hcl:"go_packages,block"`
	ServiceSDK            SDK            `hcl:"sdk,block"`
	ServiceNames          Names          `hcl:"names,block"`
	ServiceClient         Client         `hcl:"client,block"`
	ServiceEnvVars        EnvVar         `hcl:"env_var,block"`
	ServiceEndpoints      EndpointInfo   `hcl:"endpoint_info,block"`
	ServiceResourcePrefix ResourcePrefix `hcl:"resource_prefix,block"`

	SubService []Service `hcl:"sub_service,block"`

	ServiceProviderPackageCorrect string   `hcl:"provider_package_correct,attr"`
	ServiceSplitPackage           string   `hcl:"split_package,attr"`
	FilePrefix                    string   `hcl:"file_prefix,attr"`
	DocPrefix                     []string `hcl:"doc_prefix,attr"`
	Brand                         string   `hcl:"brand,attr"`
	Exclude                       bool     `hcl:"exclude,attr"`
	NotImplemented                bool     `hcl:"not_implemented,attr"`
	AllowedSubcategory            bool     `hcl:"allowed_subcategory,attr"`
	Note                          string   `hcl:"note,attr"`
}

type Services struct {
	ServiceList []Service `hcl:"service,block"`
}

//go:embed names_data.hcl
var b []byte

const (
	colAWSCLIV2Command = iota
	colAWSCLIV2CommandNoDashes
	colGoV1Package
	colGoV2Package
	colProviderPackageActual
	colProviderPackageCorrect
	colSplitPackageRealPackage
	colAliases
	colProviderNameUpper
	colGoV1ClientTypeName
	colSkipClientGenerate
	colClientSDKV1
	colClientSDKV2
	colResourcePrefixActual
	colResourcePrefixCorrect
	colFilePrefix
	colDocPrefix
	colHumanFriendly
	colBrand
	colExclude        // If set, the service is completely ignored
	colNotImplemented // If set, the service will be included in, e.g. labels, but not have a service client
	colEndpointOnly   // If set, the service is included in list of endpoints
	colAllowedSubcategory
	colDeprecatedEnvVar  // Deprecated `AWS_<service>_ENDPOINT` envvar defined for some services
	colTFAWSEnvVar       // `TF_AWS_<service>_ENDPOINT` envvar defined for some services
	colSDKID             // Service SDK ID from AWS SDK for Go v2
	colEndpointAPICall   // API call to use for endpoint tests
	colEndpointAPIParams // Any needed parameters for endpoint tests
	colEndpointOverrideRegion
	colNote
)
