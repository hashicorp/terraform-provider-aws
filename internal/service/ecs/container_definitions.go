// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ecs

import (
	"cmp"
	"fmt"
	"slices"
	_ "unsafe" // Required for go:linkname

	"github.com/aws/aws-sdk-go-v2/aws"
	_ "github.com/aws/aws-sdk-go-v2/service/ecs" // Required for go:linkname
	awstypes "github.com/aws/aws-sdk-go-v2/service/ecs/types"
	smithyjson "github.com/aws/smithy-go/encoding/json"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	tfjson "github.com/hashicorp/terraform-provider-aws/internal/json"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	itypes "github.com/hashicorp/terraform-provider-aws/internal/types"
)

func containerDefinitionsAreEquivalent(def1, def2 string, isAWSVPC bool) (bool, error) {
	var obj1 containerDefinitions
	err := tfjson.DecodeFromString(def1, &obj1)
	if err != nil {
		return false, err
	}
	obj1.reduce(isAWSVPC)
	b1, err := tfjson.EncodeToBytes(obj1)
	if err != nil {
		return false, err
	}

	var obj2 containerDefinitions
	err = tfjson.DecodeFromString(def2, &obj2)
	if err != nil {
		return false, err
	}
	obj2.reduce(isAWSVPC)
	b2, err := tfjson.EncodeToBytes(obj2)
	if err != nil {
		return false, err
	}

	return tfjson.EqualBytes(b1, b2), nil
}

type containerDefinitions []awstypes.ContainerDefinition

func (cd containerDefinitions) reduce(isAWSVPC bool) {
	// Deal with fields which may be re-ordered in the API.
	cd.orderContainers()
	cd.orderEnvironmentVariables()
	cd.orderSecrets()

	// Compact any sparse lists.
	cd.compactArrays()

	// Deal with special fields which have defaults.
	// See https://docs.aws.amazon.com/AmazonECS/latest/developerguide/task_definition_parameters.html#container_definitions.
	for i, def := range cd {
		if def.Essential == nil {
			cd[i].Essential = aws.Bool(true)
		}

		if hc := def.HealthCheck; hc != nil {
			if hc.Interval == nil {
				hc.Interval = aws.Int32(30)
			}
			if hc.Retries == nil {
				hc.Retries = aws.Int32(3)
			}
			if hc.Timeout == nil {
				hc.Timeout = aws.Int32(5)
			}
		}

		for j, pm := range def.PortMappings {
			if pm.Protocol == awstypes.TransportProtocolTcp {
				cd[i].PortMappings[j].Protocol = ""
			}
			if aws.ToInt32(pm.HostPort) == 0 {
				cd[i].PortMappings[j].HostPort = nil
			}
			if isAWSVPC && cd[i].PortMappings[j].HostPort == nil {
				cd[i].PortMappings[j].HostPort = cd[i].PortMappings[j].ContainerPort
			}
		}

		// Set all empty slices to nil.
		if len(def.Command) == 0 {
			cd[i].Command = nil
		}
		if len(def.CredentialSpecs) == 0 {
			cd[i].CredentialSpecs = nil
		}
		if len(def.DependsOn) == 0 {
			cd[i].DependsOn = nil
		}
		if len(def.DnsSearchDomains) == 0 {
			cd[i].DnsSearchDomains = nil
		}
		if len(def.DnsServers) == 0 {
			cd[i].DnsServers = nil
		}
		if len(def.DockerSecurityOptions) == 0 {
			cd[i].DockerSecurityOptions = nil
		}
		if len(def.EntryPoint) == 0 {
			cd[i].EntryPoint = nil
		}
		if len(def.Environment) == 0 {
			cd[i].Environment = nil
		}
		if len(def.EnvironmentFiles) == 0 {
			cd[i].EnvironmentFiles = nil
		}
		if len(def.ExtraHosts) == 0 {
			cd[i].ExtraHosts = nil
		}
		if len(def.Links) == 0 {
			cd[i].Links = nil
		}
		if len(def.MountPoints) == 0 {
			cd[i].MountPoints = nil
		}
		if len(def.PortMappings) == 0 {
			cd[i].PortMappings = nil
		}
		if len(def.ResourceRequirements) == 0 {
			cd[i].ResourceRequirements = nil
		}
		if len(def.Secrets) == 0 {
			cd[i].Secrets = nil
		}
		if len(def.SystemControls) == 0 {
			cd[i].SystemControls = nil
		}
		if len(def.Ulimits) == 0 {
			cd[i].Ulimits = nil
		}
		if len(def.VolumesFrom) == 0 {
			cd[i].VolumesFrom = nil
		}
	}
}

func (cd containerDefinitions) orderEnvironmentVariables() {
	for i, def := range cd {
		slices.SortFunc(def.Environment, func(a, b awstypes.KeyValuePair) int {
			return cmp.Compare(aws.ToString(a.Name), aws.ToString(b.Name))
		})
		cd[i].Environment = def.Environment
	}
}

func (cd containerDefinitions) orderSecrets() {
	for i, def := range cd {
		slices.SortFunc(def.Secrets, func(a, b awstypes.Secret) int {
			return cmp.Compare(aws.ToString(a.Name), aws.ToString(b.Name))
		})
		cd[i].Secrets = def.Secrets
	}
}

func (cd containerDefinitions) orderContainers() {
	slices.SortFunc(cd, func(a, b awstypes.ContainerDefinition) int {
		return cmp.Compare(aws.ToString(a.Name), aws.ToString(b.Name))
	})
}

// compactArrays removes any zero values from the object arrays in the container definitions.
func (cd containerDefinitions) compactArrays() {
	for i, def := range cd {
		cd[i].DependsOn = compactArray(def.DependsOn)
		cd[i].Environment = compactArray(def.Environment)
		cd[i].EnvironmentFiles = compactArray(def.EnvironmentFiles)
		cd[i].ExtraHosts = compactArray(def.ExtraHosts)
		cd[i].MountPoints = compactArray(def.MountPoints)
		cd[i].PortMappings = compactArray(def.PortMappings)
		cd[i].ResourceRequirements = compactArray(def.ResourceRequirements)
		cd[i].Secrets = compactArray(def.Secrets)
		cd[i].SystemControls = compactArray(def.SystemControls)
		cd[i].Ulimits = compactArray(def.Ulimits)
		cd[i].VolumesFrom = compactArray(def.VolumesFrom)
	}
}

func compactArray[S ~[]E, E any](s S) S {
	if len(s) == 0 {
		return s
	}

	return tfslices.Filter(s, func(e E) bool {
		return !itypes.IsZero(&e)
	})
}

// Dirty hack to avoid any backwards compatibility issues with the AWS SDK for Go v2 migration.
// Reach down into the SDK and use the same serialization function that the SDK uses.
//
//go:linkname serializeContainerDefinitions github.com/aws/aws-sdk-go-v2/service/ecs.awsAwsjson11_serializeDocumentContainerDefinitions
func serializeContainerDefinitions(v []awstypes.ContainerDefinition, value smithyjson.Value) error

func flattenContainerDefinitions(apiObjects []awstypes.ContainerDefinition) (string, error) {
	jsonEncoder := smithyjson.NewEncoder()
	err := serializeContainerDefinitions(apiObjects, jsonEncoder.Value)

	if err != nil {
		return "", err
	}

	return jsonEncoder.String(), nil
}

func expandContainerDefinitions(tfString string) ([]awstypes.ContainerDefinition, error) {
	var apiObjects []awstypes.ContainerDefinition

	if err := tfjson.DecodeFromString(tfString, &apiObjects); err != nil {
		return nil, err
	}

	for i, apiObject := range apiObjects {
		if itypes.IsZero(&apiObject) {
			return nil, fmt.Errorf("invalid container definition supplied at index (%d)", i)
		}
		if !isValidVersionConsistency(apiObject) {
			return nil, fmt.Errorf("invalid version consistency value (%[1]s) for container definition supplied at index (%[2]d)", apiObject.VersionConsistency, i)
		}
	}

	containerDefinitions(apiObjects).compactArrays()

	return apiObjects, nil
}

func isValidVersionConsistency(cd awstypes.ContainerDefinition) bool {
	if cd.VersionConsistency == "" {
		return true
	}

	return slices.Contains(enum.EnumValues[awstypes.VersionConsistency](), cd.VersionConsistency)
}
