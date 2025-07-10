// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package batch

import (
	"cmp"
	"slices"
	_ "unsafe" // Required for go:linkname

	"github.com/aws/aws-sdk-go-v2/aws"
	_ "github.com/aws/aws-sdk-go-v2/service/batch" // Required for go:linkname
	awstypes "github.com/aws/aws-sdk-go-v2/service/batch/types"
	smithyjson "github.com/aws/smithy-go/encoding/json"
	tfjson "github.com/hashicorp/terraform-provider-aws/internal/json"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
)

type ecsProperties awstypes.EcsProperties

func (ep *ecsProperties) reduce() {
	ep.orderContainers()
	ep.orderEnvironmentVariables()
	ep.orderSecrets()

	// Set all empty slices to nil.
	// Deal with special fields which have defaults.
	for i, taskProps := range ep.TaskProperties {
		for j, container := range taskProps.Containers {
			if container.Essential == nil {
				container.Essential = aws.Bool(true)
			}

			if len(container.Command) == 0 {
				container.Command = nil
			}
			if len(container.DependsOn) == 0 {
				container.DependsOn = nil
			}
			if len(container.Environment) == 0 {
				container.Environment = nil
			}
			if container.LogConfiguration != nil && len(container.LogConfiguration.SecretOptions) == 0 {
				container.LogConfiguration.SecretOptions = nil
			}
			if len(container.MountPoints) == 0 {
				container.MountPoints = nil
			}
			if len(container.Secrets) == 0 {
				container.Secrets = nil
			}
			if len(container.Ulimits) == 0 {
				container.Ulimits = nil
			}

			taskProps.Containers[j] = container
		}

		if taskProps.PlatformVersion == nil {
			taskProps.PlatformVersion = aws.String(fargatePlatformVersionLatest)
		}

		if len(taskProps.Volumes) == 0 {
			taskProps.Volumes = nil
		}

		ep.TaskProperties[i] = taskProps
	}
}

func (ep *ecsProperties) orderContainers() {
	for i, taskProps := range ep.TaskProperties {
		slices.SortFunc(taskProps.Containers, func(a, b awstypes.TaskContainerProperties) int {
			return cmp.Compare(aws.ToString(a.Name), aws.ToString(b.Name))
		})

		ep.TaskProperties[i].Containers = taskProps.Containers
	}
}

func (ep *ecsProperties) orderEnvironmentVariables() {
	for i, taskProps := range ep.TaskProperties {
		for j, container := range taskProps.Containers {
			// Remove environment variables with empty values.
			container.Environment = tfslices.Filter(container.Environment, func(kvp awstypes.KeyValuePair) bool {
				return aws.ToString(kvp.Value) != ""
			})

			slices.SortFunc(container.Environment, func(a, b awstypes.KeyValuePair) int {
				return cmp.Compare(aws.ToString(a.Name), aws.ToString(b.Name))
			})

			ep.TaskProperties[i].Containers[j].Environment = container.Environment
		}
	}
}

func (ep *ecsProperties) orderSecrets() {
	for i, taskProps := range ep.TaskProperties {
		for j, container := range taskProps.Containers {
			slices.SortFunc(container.Secrets, func(a, b awstypes.Secret) int {
				return cmp.Compare(aws.ToString(a.Name), aws.ToString(b.Name))
			})

			ep.TaskProperties[i].Containers[j].Secrets = container.Secrets
		}
	}
}

func equivalentECSPropertiesJSON(str1, str2 string) (bool, error) {
	if str1 == "" {
		str1 = "{}"
	}

	if str2 == "" {
		str2 = "{}"
	}

	var ep1 ecsProperties
	err := tfjson.DecodeFromString(str1, &ep1)
	if err != nil {
		return false, err
	}
	ep1.reduce()
	b1, err := tfjson.EncodeToBytes(ep1)
	if err != nil {
		return false, err
	}

	var ep2 ecsProperties
	err = tfjson.DecodeFromString(str2, &ep2)
	if err != nil {
		return false, err
	}
	ep2.reduce()
	b2, err := tfjson.EncodeToBytes(ep2)
	if err != nil {
		return false, err
	}

	return tfjson.EqualBytes(b1, b2), nil
}

func expandECSProperties(tfString string) (*awstypes.EcsProperties, error) {
	apiObject := &awstypes.EcsProperties{}

	if err := tfjson.DecodeFromString(tfString, apiObject); err != nil {
		return nil, err
	}

	return apiObject, nil
}

// Dirty hack to avoid any backwards compatibility issues with the AWS SDK for Go v2 migration.
// Reach down into the SDK and use the same serialization function that the SDK uses.
//
//go:linkname serializeECSPProperties github.com/aws/aws-sdk-go-v2/service/batch.awsRestjson1_serializeDocumentEcsProperties
func serializeECSPProperties(v *awstypes.EcsProperties, value smithyjson.Value) error

func flattenECSProperties(apiObject *awstypes.EcsProperties) (string, error) {
	if apiObject == nil {
		return "", nil
	}

	jsonEncoder := smithyjson.NewEncoder()
	err := serializeECSPProperties(apiObject, jsonEncoder.Value)

	if err != nil {
		return "", err
	}

	return jsonEncoder.String(), nil
}
