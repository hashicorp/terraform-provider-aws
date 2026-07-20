// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package data

import (
	_ "embed"
	"fmt"
	"log"
	"maps"
	"slices"
	"strings"

	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/hashicorp/hcl/v2/hclparse"
)

type ServiceRecord struct {
	service Service
}

func (sr ServiceRecord) AWSCLIV2Command() string {
	result := sr.service.ProviderPackage
	if sr.service.ServiceCli != nil {
		result = sr.service.ServiceCli.AWSCLIV2Command
	}
	return result
}

func (sr ServiceRecord) AWSCLIV2CommandNoDashes() string {
	result := sr.service.ProviderPackage
	if sr.service.ServiceCli != nil {
		result = sr.service.ServiceCli.AWSCLIV2CommandNoDashes
	}
	return result
}

func (sr ServiceRecord) GoPackageName() string {
	switch sr.SDKVersion() {
	case 1:
		return sr.GoV1Package()
	case 2:
		return sr.GoV2Package()
	}
	return ""
}

func (sr ServiceRecord) GoV1Package() string {
	result := sr.service.ProviderPackage
	if sr.service.ServiceGoPackages != nil {
		result = sr.service.ServiceGoPackages.V1Package
	}
	return result
}

func (sr ServiceRecord) GoV2Package() string {
	result := sr.service.ProviderPackage
	if sr.service.ServiceGoPackages != nil {
		result = sr.service.ServiceGoPackages.V2Package
	}
	return result
}

func (sr ServiceRecord) ProviderPackage() string {
	return sr.service.ProviderPackage
}

func (sr ServiceRecord) ProviderPackageCorrect() string {
	result := sr.service.ProviderPackage
	if len(sr.service.ServiceProviderPackageCorrect) > 0 {
		result = sr.service.ServiceProviderPackageCorrect
	}
	return result
}

func (sr ServiceRecord) SplitPackageRealPackage() string {
	return sr.service.ServiceSplitPackage
}

func (sr ServiceRecord) Aliases() []string {
	if len(sr.service.ServiceNames.Aliases) > 0 {
		return slices.Clone(sr.service.ServiceNames.Aliases)
	}
	return nil
}

func (sr ServiceRecord) ProviderNameUpper() string {
	return sr.service.ServiceNames.ProviderNameUpper
}

func (sr ServiceRecord) ClientTypeName(version int) (s string) {
	switch version {
	case 1:
		return sr.GoV1ClientTypeName()
	}
	return "Client"
}

func (sr ServiceRecord) GoV1ClientTypeName() string {
	if sr.service.ServiceClient != nil {
		return sr.service.ServiceClient.GoV1ClientTypeName
	}
	return ""
}

func (sr ServiceRecord) skipClientGenerate() bool {
	if sr.service.ServiceClient != nil {
		return sr.service.ServiceClient.SkipClientGenerate
	}
	return false
}

func (sr ServiceRecord) GenerateClient() bool {
	return !sr.skipClientGenerate()
}

func (sr ServiceRecord) IsClientSDKV1() bool {
	return sr.SDKVersion() == 1
}

func (sr ServiceRecord) IsClientSDKV2() bool {
	return sr.SDKVersion() == 2 //nolint:mnd
}

func (sr ServiceRecord) SDKVersion() int {
	if sr.service.ServiceSDK != nil {
		return sr.service.ServiceSDK.Version
	}
	return 0
}

func (sr ServiceRecord) ResourcePrefix() string {
	prefix := sr.ResourcePrefixCorrect()
	if sr.ResourcePrefixActual() != "" {
		prefix = sr.ResourcePrefixActual()
	}
	return prefix
}

func (sr ServiceRecord) ResourcePrefixActual() string {
	return sr.service.ServiceResourcePrefix.ResourcePrefixActual
}

func (sr ServiceRecord) ResourcePrefixCorrect() string {
	return sr.service.ServiceResourcePrefix.ResourcePrefixCorrect
}

func (sr ServiceRecord) FilePrefix() string {
	return sr.service.FilePrefix
}

func (sr ServiceRecord) DocPrefix() []string {
	return sr.service.DocPrefix
}

func (sr ServiceRecord) HumanFriendly() string {
	return sr.service.ServiceNames.HumanFriendly
}

func (sr ServiceRecord) HumanFriendlyShort() string {
	if sr.service.ServiceNames.HumanFriendlyShort != "" {
		return sr.service.ServiceNames.HumanFriendlyShort
	}
	return sr.service.ServiceNames.HumanFriendly
}

func (sr ServiceRecord) FullHumanFriendly() string {
	if sr.Brand() == "" {
		return sr.HumanFriendly()
	}

	return fmt.Sprintf("%s %s", sr.Brand(), sr.HumanFriendly())
}

func (sr ServiceRecord) Brand() string {
	return sr.service.Brand
}

func (sr ServiceRecord) Exclude() bool {
	return sr.service.Exclude
}

func (sr ServiceRecord) IsGlobal() bool {
	return sr.service.IsGlobal
}

func (sr ServiceRecord) NotImplemented() bool {
	return sr.service.NotImplemented
}

func (sr ServiceRecord) EndpointOnly() bool {
	if sr.service.ServiceEndpoints != nil {
		return sr.service.ServiceEndpoints.EndpointOnly
	}
	return false
}

func (sr ServiceRecord) AllowedSubcategory() bool {
	return sr.service.AllowedSubcategory
}

func (sr ServiceRecord) DeprecatedEnvVar() string {
	if sr.service.ServiceEnvVars != nil {
		return sr.service.ServiceEnvVars.DeprecatedEnvVar
	}
	return ""
}

func (sr ServiceRecord) TFAWSEnvVar() string {
	if sr.service.ServiceEnvVars != nil {
		return sr.service.ServiceEnvVars.TFAWSEnvVar
	}
	return ""
}

func (sr ServiceRecord) SDKID() string {
	if sr.service.ServiceSDK != nil {
		return sr.service.ServiceSDK.ID
	}
	return ""
}

func (sr ServiceRecord) ARNNamespace() string {
	if sr.service.ServiceSDK != nil {
		return sr.service.ServiceSDK.ARNNamespace
	}
	return ""
}

func (sr ServiceRecord) AWSServiceEnvVar() string {
	return "AWS_ENDPOINT_URL_" + strings.ReplaceAll(strings.ToUpper(sr.SDKID()), " ", "_")
}

func (sr ServiceRecord) AWSConfigParameter() string {
	return strings.ReplaceAll(strings.ToLower(sr.SDKID()), " ", "_")
}

func (sr ServiceRecord) EndpointAPICall() string {
	if sr.service.ServiceEndpoints != nil {
		return sr.service.ServiceEndpoints.EndpointAPICall
	}
	return ""
}

func (sr ServiceRecord) EndpointAPIParams() string {
	if sr.service.ServiceEndpoints != nil {
		return sr.service.ServiceEndpoints.EndpointAPIParams
	}
	return ""
}

func (sr ServiceRecord) EndpointRegionOverrides() map[string]string {
	if sr.service.ServiceEndpoints != nil && len(sr.service.ServiceEndpoints.EndpointRegionOverrides) > 0 {
		return maps.Clone(sr.service.ServiceEndpoints.EndpointRegionOverrides)
	}
	return nil
}

func (sr ServiceRecord) Note() string {
	return sr.service.Note
}

func parseService(curr Service) ServiceRecord {
	return ServiceRecord{
		service: curr,
	}
}

func ReadAllServiceData() (results []ServiceRecord, err error) {
	var decodedServiceList Services
	parser := hclparse.NewParser()
	toParse, parseErr := parser.ParseHCL(b, "names_data.hcl")
	if parseErr.HasErrors() {
		log.Fatalf("Parser error: %s", parseErr)
	}
	decodeErr := gohcl.DecodeBody(toParse.Body, nil, &decodedServiceList)
	if decodeErr.HasErrors() {
		log.Fatalf("Decode error: %s", decodeErr)
	}
	for _, curr := range decodedServiceList.ServiceList {
		if curr.ServiceSDK != nil && curr.ServiceSDK.Version == 0 {
			curr.ServiceSDK.Version = 2
		}
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
	AWSCLIV2Command         string `hcl:"aws_cli_v2_command,optional"`
	AWSCLIV2CommandNoDashes string `hcl:"aws_cli_v2_command_no_dashes,optional"`
}

type GoPackages struct {
	V1Package string `hcl:"v1_package,optional"`
	V2Package string `hcl:"v2_package,optional"`
}

type ResourcePrefix struct {
	ResourcePrefixActual  string `hcl:"actual,optional"`
	ResourcePrefixCorrect string `hcl:"correct,optional"`
}

type SDK struct {
	ID           string `hcl:"id,optional"`
	Version      int    `hcl:"client_version,optional"`
	ARNNamespace string `hcl:"arn_namespace,optional"`
}

type Names struct {
	Aliases            []string `hcl:"aliases,optional"`
	ProviderNameUpper  string   `hcl:"provider_name_upper,attr"`
	HumanFriendly      string   `hcl:"human_friendly,attr"`
	HumanFriendlyShort string   `hcl:"human_friendly_short,optional"`
}

type ProviderPackage struct {
	Actual  string `hcl:"actual,optional"`
	Correct string `hcl:"correct,optional"`
}

type Client struct {
	GoV1ClientTypeName string `hcl:"go_v1_client_typename,optional"`
	SkipClientGenerate bool   `hcl:"skip_client_generate,optional"`
}

type EnvVar struct {
	DeprecatedEnvVar string `hcl:"deprecated_env_var,optional"`
	TFAWSEnvVar      string `hcl:"tf_aws_env_var,optional"`
}

type EndpointInfo struct {
	EndpointAPICall         string            `hcl:"endpoint_api_call,optional"`
	EndpointAPIParams       string            `hcl:"endpoint_api_params,optional"`
	EndpointRegionOverrides map[string]string `hcl:"endpoint_region_overrides,optional"`
	EndpointOnly            bool              `hcl:"endpoint_only,optional"`
}

type Service struct {
	ProviderPackage       string         `hcl:",label"`
	ServiceCli            *CLIV2Command  `hcl:"cli_v2_command,block"`
	ServiceGoPackages     *GoPackages    `hcl:"go_packages,block"`
	ServiceSDK            *SDK           `hcl:"sdk,block"`
	ServiceNames          Names          `hcl:"names,block"`
	ServiceClient         *Client        `hcl:"client,block"`
	ServiceEnvVars        *EnvVar        `hcl:"env_var,block"`
	ServiceEndpoints      *EndpointInfo  `hcl:"endpoint_info,block"`
	ServiceResourcePrefix ResourcePrefix `hcl:"resource_prefix,block"`

	SubService []Service `hcl:"sub_service,block"`

	ServiceProviderPackageCorrect string   `hcl:"provider_package_correct,optional"`
	ServiceSplitPackage           string   `hcl:"split_package,optional"`
	FilePrefix                    string   `hcl:"file_prefix,optional"`
	DocPrefix                     []string `hcl:"doc_prefix,optional"`
	Brand                         string   `hcl:"brand,optional"`
	Exclude                       bool     `hcl:"exclude,optional"`
	NotImplemented                bool     `hcl:"not_implemented,optional"`
	AllowedSubcategory            bool     `hcl:"allowed_subcategory,optional"`
	Note                          string   `hcl:"note,optional"`
	IsGlobal                      bool     `hcl:"is_global,optional"`
}

type Services struct {
	ServiceList []Service `hcl:"service,block"`
}

//go:embed names_data.hcl
var b []byte
