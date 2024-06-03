// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package data

import (
	"bytes"
	_ "embed"
	"encoding/csv"
	"errors"
	"io"
	"strings"
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

func (sr ServiceRecord) Note() string {
	return sr[colNote]
}

func ReadAllServiceData() (results []ServiceRecord, err error) {
	reader := csv.NewReader(bytes.NewReader(namesData))
	// reader.ReuseRecord = true

	// Skip the header
	_, err = reader.Read()
	if err != nil {
		return
	}

	for {
		r, err := reader.Read()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return nil, err
		}
		results = append(results, ServiceRecord(r))
	}

	return
}

//go:embed names_data.csv
var namesData []byte

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
	colNote
)
