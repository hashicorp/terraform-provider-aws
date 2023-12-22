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

func (sr ServiceRecord) ClientSDKV1() string {
	return sr[colClientSDKV1]
}

func (sr ServiceRecord) ClientSDKV2() string {
	return sr[colClientSDKV2]
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

func (sr ServiceRecord) TfAwsEnvVar() string {
	return sr[colTfAwsEnvVar]
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
	colDeprecatedEnvVar // Deprecated `AWS_<service>_ENDPOINT` envvar defined for some services
	colTfAwsEnvVar      // `TF_AWS_<service>_ENDPOINT` envvar defined for some services
	colNote
)
