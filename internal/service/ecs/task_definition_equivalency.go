// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ecs

import (
	"bytes"
	"encoding/json"
	"log"
	"sort"

	"github.com/aws/aws-sdk-go-v2/aws"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ecs/types"
	"github.com/aws/aws-sdk-go/private/protocol/json/jsonutil"
)

// ContainerDefinitionsAreEquivalent determines equality between two ECS container definition JSON strings
// Note: This function will be moved out of the aws package in the future.
func ContainerDefinitionsAreEquivalent(def1, def2 string, isAWSVPC bool) (bool, error) {
	var obj1 containerDefinitions
	err := json.Unmarshal([]byte(def1), &obj1)
	if err != nil {
		return false, err
	}
	err = obj1.Reduce(isAWSVPC)
	if err != nil {
		return false, err
	}
	canonicalJson1, err := jsonutil.BuildJSON(obj1)
	if err != nil {
		return false, err
	}

	var obj2 containerDefinitions
	err = json.Unmarshal([]byte(def2), &obj2)
	if err != nil {
		return false, err
	}
	err = obj2.Reduce(isAWSVPC)
	if err != nil {
		return false, err
	}

	canonicalJson2, err := jsonutil.BuildJSON(obj2)
	if err != nil {
		return false, err
	}

	equal := bytes.Equal(canonicalJson1, canonicalJson2)
	if !equal {
		log.Printf("[DEBUG] Canonical definitions are not equal.\nFirst: %s\nSecond: %s\n",
			canonicalJson1, canonicalJson2)
	}
	return equal, nil
}

type containerDefinitions []awstypes.ContainerDefinition

func (cd containerDefinitions) Reduce(isAWSVPC bool) error {
	// Deal with fields which may be re-ordered in the API
	cd.OrderContainers()
	cd.OrderEnvironmentVariables()
	cd.OrderSecrets()

	for i, def := range cd {
		// Deal with special fields which have defaults
		if def.Essential == nil {
			def.Essential = aws.Bool(true)
		}
		for j, pm := range def.PortMappings {
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
	return nil
}

func (cd containerDefinitions) OrderEnvironmentVariables() {
	for _, def := range cd {
		sort.Slice(def.Environment, func(i, j int) bool {
			return aws.ToString(def.Environment[i].Name) < aws.ToString(def.Environment[j].Name)
		})
	}
}

func (cd containerDefinitions) OrderSecrets() {
	for _, def := range cd {
		sort.Slice(def.Secrets, func(i, j int) bool {
			return aws.ToString(def.Secrets[i].Name) < aws.ToString(def.Secrets[j].Name)
		})
	}
}

func (cd containerDefinitions) OrderContainers() {
	sort.Slice(cd, func(i, j int) bool {
		return aws.ToString(cd[i].Name) < aws.ToString(cd[j].Name)
	})
}
