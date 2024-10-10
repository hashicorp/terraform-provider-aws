// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package batch

import (
	"sort"
	_ "unsafe" // Required for go:linkname

	"github.com/aws/aws-sdk-go-v2/aws"
	_ "github.com/aws/aws-sdk-go-v2/service/batch" // Required for go:linkname
	awstypes "github.com/aws/aws-sdk-go-v2/service/batch/types"
	smithyjson "github.com/aws/smithy-go/encoding/json"
	tfjson "github.com/hashicorp/terraform-provider-aws/internal/json"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
)

const (
	fargatePlatformVersionLatest = "LATEST"
)

type containerProperties awstypes.ContainerProperties

func (cp *containerProperties) reduce() {
	cp.sortEnvironment()

	// Prevent difference of API response that adds an empty array when not configured during the request.
	if len(cp.Command) == 0 {
		cp.Command = nil
	}

	// Remove environment variables with empty values.
	cp.Environment = tfslices.Filter(cp.Environment, func(kvp awstypes.KeyValuePair) bool {
		return aws.ToString(kvp.Value) != ""
	})

	// Prevent difference of API response that adds an empty array when not configured during the request.
	if len(cp.Environment) == 0 {
		cp.Environment = nil
	}

	// Prevent difference of API response that contains the default Fargate platform configuration.
	if cp.FargatePlatformConfiguration != nil {
		if aws.ToString(cp.FargatePlatformConfiguration.PlatformVersion) == fargatePlatformVersionLatest {
			cp.FargatePlatformConfiguration = nil
		}
	}

	if cp.LinuxParameters != nil {
		if len(cp.LinuxParameters.Devices) == 0 {
			cp.LinuxParameters.Devices = nil
		}

		for i, device := range cp.LinuxParameters.Devices {
			if len(device.Permissions) == 0 {
				cp.LinuxParameters.Devices[i].Permissions = nil
			}
		}

		if len(cp.LinuxParameters.Tmpfs) == 0 {
			cp.LinuxParameters.Tmpfs = nil
		}

		for i, tmpfs := range cp.LinuxParameters.Tmpfs {
			if len(tmpfs.MountOptions) == 0 {
				cp.LinuxParameters.Tmpfs[i].MountOptions = nil
			}
		}
	}

	// Prevent difference of API response that adds an empty array when not configured during the request.
	if cp.LogConfiguration != nil {
		if len(cp.LogConfiguration.Options) == 0 {
			cp.LogConfiguration.Options = nil
		}

		if len(cp.LogConfiguration.SecretOptions) == 0 {
			cp.LogConfiguration.SecretOptions = nil
		}
	}

	// Prevent difference of API response that adds an empty array when not configured during the request.
	if len(cp.MountPoints) == 0 {
		cp.MountPoints = nil
	}

	// Prevent difference of API response that adds an empty array when not configured during the request.
	if len(cp.ResourceRequirements) == 0 {
		cp.ResourceRequirements = nil
	}

	// Prevent difference of API response that adds an empty array when not configured during the request.
	if len(cp.Secrets) == 0 {
		cp.Secrets = nil
	}

	// Prevent difference of API response that adds an empty array when not configured during the request.
	if len(cp.Ulimits) == 0 {
		cp.Ulimits = nil
	}

	// Prevent difference of API response that adds an empty array when not configured during the request.
	if len(cp.Volumes) == 0 {
		cp.Volumes = nil
	}
}

func (cp *containerProperties) sortEnvironment() {
	// Deal with Environment objects which may be re-ordered in the API.
	sort.Slice(cp.Environment, func(i, j int) bool {
		return aws.ToString(cp.Environment[i].Name) < aws.ToString(cp.Environment[j].Name)
	})
}

// equivalentContainerPropertiesJSON determines equality between two Batch ContainerProperties JSON strings
func equivalentContainerPropertiesJSON(str1, str2 string) (bool, error) {
	if str1 == "" {
		str1 = "{}"
	}

	if str2 == "" {
		str2 = "{}"
	}

	var cp1 containerProperties
	err := tfjson.DecodeFromString(str1, &cp1)
	if err != nil {
		return false, err
	}
	cp1.reduce()
	b1, err := tfjson.EncodeToBytes(cp1)
	if err != nil {
		return false, err
	}

	var cp2 containerProperties
	err = tfjson.DecodeFromString(str2, &cp2)
	if err != nil {
		return false, err
	}
	cp2.reduce()
	b2, err := tfjson.EncodeToBytes(cp2)
	if err != nil {
		return false, err
	}

	return tfjson.EqualBytes(b1, b2), nil
}

func expandContainerProperties(tfString string) (*awstypes.ContainerProperties, error) {
	apiObject := &awstypes.ContainerProperties{}

	if err := tfjson.DecodeFromString(tfString, apiObject); err != nil {
		return nil, err
	}

	return apiObject, nil
}

// Dirty hack to avoid any backwards compatibility issues with the AWS SDK for Go v2 migration.
// Reach down into the SDK and use the same serialization function that the SDK uses.
//
//go:linkname serializeContainerProperties github.com/aws/aws-sdk-go-v2/service/batch.awsRestjson1_serializeDocumentContainerProperties
func serializeContainerProperties(v *awstypes.ContainerProperties, value smithyjson.Value) error

func flattenContainerProperties(apiObject *awstypes.ContainerProperties) (string, error) {
	if apiObject == nil {
		return "", nil
	}

	(*containerProperties)(apiObject).sortEnvironment()

	jsonEncoder := smithyjson.NewEncoder()
	err := serializeContainerProperties(apiObject, jsonEncoder.Value)

	if err != nil {
		return "", err
	}

	return jsonEncoder.String(), nil
}
