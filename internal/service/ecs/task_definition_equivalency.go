// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ecs

import (
	"sort"

	"github.com/aws/aws-sdk-go-v2/aws"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ecs/types"
	tfjson "github.com/hashicorp/terraform-provider-aws/internal/json"
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

	for i, def := range cd {
		// Deal with special fields which have defaults.
		if def.Essential == nil {
			cd[i].Essential = aws.Bool(true)
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
		sort.Slice(def.Environment, func(i, j int) bool {
			return aws.ToString(def.Environment[i].Name) < aws.ToString(def.Environment[j].Name)
		})
		cd[i].Environment = def.Environment
	}
}

func (cd containerDefinitions) orderSecrets() {
	for i, def := range cd {
		sort.Slice(def.Secrets, func(i, j int) bool {
			return aws.ToString(def.Secrets[i].Name) < aws.ToString(def.Secrets[j].Name)
		})
		cd[i].Secrets = def.Secrets
	}
}

func (cd containerDefinitions) orderContainers() {
	sort.Slice(cd, func(i, j int) bool {
		return aws.ToString(cd[i].Name) < aws.ToString(cd[j].Name)
	})
}
