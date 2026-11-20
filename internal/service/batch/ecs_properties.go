// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package batch

import (
	"cmp"
	"slices"

	"github.com/aws/aws-sdk-go-v2/aws"
	awstypes "github.com/aws/aws-sdk-go-v2/service/batch/types"
	smithyjson "github.com/aws/smithy-go/encoding/json"
	tfjson "github.com/hashicorp/terraform-provider-aws/internal/json"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	"github.com/hashicorp/terraform-provider-aws/names"
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

func serializeECSPProperties(v *awstypes.EcsProperties, value smithyjson.Value) error {
	o := value.Object()
	defer o.Close()

	if v.TaskProperties != nil {
		a := o.Key("taskProperties").Array()
		for _, task := range v.TaskProperties {
			serializeECSTaskProperties(&task, a.Value())
		}
		a.Close()
	}

	return nil
}

func serializeECSTaskProperties(v *awstypes.EcsTaskProperties, value smithyjson.Value) {
	o := value.Object()
	defer o.Close()

	if v.Containers != nil {
		a := o.Key("containers").Array()
		for _, container := range v.Containers {
			serializeTaskContainerProperties(&container, a.Value())
		}
		a.Close()
	}
	if v.EnableExecuteCommand != nil {
		o.Key("enableExecuteCommand").Boolean(*v.EnableExecuteCommand)
	}
	if v.EphemeralStorage != nil {
		serializeEphemeralStorage(v.EphemeralStorage, o.Key("ephemeralStorage"))
	}
	if v.ExecutionRoleArn != nil {
		o.Key("executionRoleArn").String(*v.ExecutionRoleArn)
	}
	if v.IpcMode != nil {
		o.Key("ipcMode").String(*v.IpcMode)
	}
	if v.NetworkConfiguration != nil {
		serializeNetworkConfiguration(v.NetworkConfiguration, o.Key("networkConfiguration"))
	}
	if v.NetworkMode != nil {
		o.Key("networkMode").String(*v.NetworkMode)
	}
	if v.PidMode != nil {
		o.Key("pidMode").String(*v.PidMode)
	}
	if v.PlatformVersion != nil {
		o.Key("platformVersion").String(*v.PlatformVersion)
	}
	if v.RuntimePlatform != nil {
		serializeRuntimePlatform(v.RuntimePlatform, o.Key("runtimePlatform"))
	}
	if v.TaskRoleArn != nil {
		o.Key("taskRoleArn").String(*v.TaskRoleArn)
	}
	if v.Volumes != nil {
		serializeVolumes(v.Volumes, o.Key("volumes"))
	}
}

func serializeTaskContainerProperties(v *awstypes.TaskContainerProperties, value smithyjson.Value) {
	o := value.Object()
	defer o.Close()

	if v.Command != nil {
		serializeStringList(v.Command, o.Key("command"))
	}
	if v.DependsOn != nil {
		a := o.Key("dependsOn").Array()
		for _, dependency := range v.DependsOn {
			d := a.Value().Object()
			if dependency.Condition != nil {
				d.Key(names.AttrCondition).String(*dependency.Condition)
			}
			if dependency.ContainerName != nil {
				d.Key("containerName").String(*dependency.ContainerName)
			}
			d.Close()
		}
		a.Close()
	}
	if v.Environment != nil {
		serializeEnvironment(v.Environment, o.Key(names.AttrEnvironment))
	}
	if v.Essential != nil {
		o.Key("essential").Boolean(*v.Essential)
	}
	if v.FirelensConfiguration != nil {
		f := o.Key("firelensConfiguration").Object()
		if v.FirelensConfiguration.Options != nil {
			serializeStringMap(v.FirelensConfiguration.Options, f.Key("options"))
		}
		if v.FirelensConfiguration.Type != "" {
			f.Key(names.AttrType).String(string(v.FirelensConfiguration.Type))
		}
		f.Close()
	}
	if v.Image != nil {
		o.Key("image").String(*v.Image)
	}
	if v.LinuxParameters != nil {
		serializeLinuxParameters(v.LinuxParameters, o.Key("linuxParameters"))
	}
	if v.LogConfiguration != nil {
		serializeLogConfiguration(v.LogConfiguration, o.Key("logConfiguration"))
	}
	if v.MountPoints != nil {
		serializeMountPoints(v.MountPoints, o.Key("mountPoints"))
	}
	if v.Name != nil {
		o.Key(names.AttrName).String(*v.Name)
	}
	if v.Privileged != nil {
		o.Key("privileged").Boolean(*v.Privileged)
	}
	if v.ReadonlyRootFilesystem != nil {
		o.Key("readonlyRootFilesystem").Boolean(*v.ReadonlyRootFilesystem)
	}
	if v.RepositoryCredentials != nil {
		serializeRepositoryCredentials(v.RepositoryCredentials, o.Key("repositoryCredentials"))
	}
	if v.ResourceRequirements != nil {
		serializeResourceRequirements(v.ResourceRequirements, o.Key("resourceRequirements"))
	}
	if v.Secrets != nil {
		serializeSecrets(v.Secrets, o.Key("secrets"))
	}
	if v.StartTimeout != nil {
		o.Key("startTimeout").Integer(*v.StartTimeout)
	}
	if v.StopTimeout != nil {
		o.Key("stopTimeout").Integer(*v.StopTimeout)
	}
	if v.Ulimits != nil {
		serializeUlimits(v.Ulimits, o.Key("ulimits"))
	}
	if v.User != nil {
		o.Key("user").String(*v.User)
	}
}

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
