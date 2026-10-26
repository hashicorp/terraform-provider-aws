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
	slices.SortFunc(cp.Environment, func(a, b awstypes.KeyValuePair) int {
		return cmp.Compare(aws.ToString(a.Name), aws.ToString(b.Name))
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

// Preserve the REST JSON representation used in state: omit nil members and empty
// enums, but retain non-nil empty collections and pointers to zero values.
func serializeContainerProperties(v *awstypes.ContainerProperties, value smithyjson.Value) error {
	o := value.Object()
	defer o.Close()

	if v.Command != nil {
		serializeStringList(v.Command, o.Key("command"))
	}
	if v.EnableExecuteCommand != nil {
		o.Key("enableExecuteCommand").Boolean(*v.EnableExecuteCommand)
	}
	if v.Environment != nil {
		serializeEnvironment(v.Environment, o.Key(names.AttrEnvironment))
	}
	if v.EphemeralStorage != nil {
		serializeEphemeralStorage(v.EphemeralStorage, o.Key("ephemeralStorage"))
	}
	if v.ExecutionRoleArn != nil {
		o.Key("executionRoleArn").String(*v.ExecutionRoleArn)
	}
	if v.FargatePlatformConfiguration != nil {
		p := o.Key("fargatePlatformConfiguration").Object()
		if v.FargatePlatformConfiguration.PlatformVersion != nil {
			p.Key("platformVersion").String(*v.FargatePlatformConfiguration.PlatformVersion)
		}
		p.Close()
	}
	if v.Image != nil {
		o.Key("image").String(*v.Image)
	}
	if v.InstanceType != nil {
		o.Key("instanceType").String(*v.InstanceType)
	}
	if v.JobRoleArn != nil {
		o.Key("jobRoleArn").String(*v.JobRoleArn)
	}
	if v.LinuxParameters != nil {
		serializeLinuxParameters(v.LinuxParameters, o.Key("linuxParameters"))
	}
	if v.LogConfiguration != nil {
		serializeLogConfiguration(v.LogConfiguration, o.Key("logConfiguration"))
	}
	if v.Memory != nil {
		o.Key("memory").Integer(*v.Memory)
	}
	if v.MountPoints != nil {
		serializeMountPoints(v.MountPoints, o.Key("mountPoints"))
	}
	if v.NetworkConfiguration != nil {
		serializeNetworkConfiguration(v.NetworkConfiguration, o.Key("networkConfiguration"))
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
	if v.RuntimePlatform != nil {
		serializeRuntimePlatform(v.RuntimePlatform, o.Key("runtimePlatform"))
	}
	if v.Secrets != nil {
		serializeSecrets(v.Secrets, o.Key("secrets"))
	}
	if v.Ulimits != nil {
		serializeUlimits(v.Ulimits, o.Key("ulimits"))
	}
	if v.User != nil {
		o.Key("user").String(*v.User)
	}
	if v.Vcpus != nil {
		o.Key("vcpus").Integer(*v.Vcpus)
	}
	if v.Volumes != nil {
		serializeVolumes(v.Volumes, o.Key("volumes"))
	}

	return nil
}

func serializeStringList(v []string, value smithyjson.Value) {
	a := value.Array()
	defer a.Close()

	for _, s := range v {
		a.Value().String(s)
	}
}

func serializeStringMap(v map[string]string, value smithyjson.Value) {
	o := value.Object()
	defer o.Close()

	for k, s := range v {
		o.Key(k).String(s)
	}
}

func serializeEnvironment(v []awstypes.KeyValuePair, value smithyjson.Value) {
	a := value.Array()
	defer a.Close()

	for _, kv := range v {
		o := a.Value().Object()
		if kv.Name != nil {
			o.Key(names.AttrName).String(*kv.Name)
		}
		if kv.Value != nil {
			o.Key(names.AttrValue).String(*kv.Value)
		}
		o.Close()
	}
}

func serializeEphemeralStorage(v *awstypes.EphemeralStorage, value smithyjson.Value) {
	o := value.Object()
	defer o.Close()

	if v.SizeInGiB != nil {
		o.Key("sizeInGiB").Integer(*v.SizeInGiB)
	}
}

func serializeLinuxParameters(v *awstypes.LinuxParameters, value smithyjson.Value) {
	o := value.Object()
	defer o.Close()

	if v.Devices != nil {
		a := o.Key("devices").Array()
		for _, device := range v.Devices {
			d := a.Value().Object()
			if device.ContainerPath != nil {
				d.Key("containerPath").String(*device.ContainerPath)
			}
			if device.HostPath != nil {
				d.Key("hostPath").String(*device.HostPath)
			}
			if device.Permissions != nil {
				p := d.Key(names.AttrPermissions).Array()
				for _, permission := range device.Permissions {
					p.Value().String(string(permission))
				}
				p.Close()
			}
			d.Close()
		}
		a.Close()
	}
	if v.InitProcessEnabled != nil {
		o.Key("initProcessEnabled").Boolean(*v.InitProcessEnabled)
	}
	if v.MaxSwap != nil {
		o.Key("maxSwap").Integer(*v.MaxSwap)
	}
	if v.SharedMemorySize != nil {
		o.Key("sharedMemorySize").Integer(*v.SharedMemorySize)
	}
	if v.Swappiness != nil {
		o.Key("swappiness").Integer(*v.Swappiness)
	}
	if v.Tmpfs != nil {
		a := o.Key("tmpfs").Array()
		for _, tmpfs := range v.Tmpfs {
			t := a.Value().Object()
			if tmpfs.ContainerPath != nil {
				t.Key("containerPath").String(*tmpfs.ContainerPath)
			}
			if tmpfs.MountOptions != nil {
				serializeStringList(tmpfs.MountOptions, t.Key("mountOptions"))
			}
			if tmpfs.Size != nil {
				t.Key(names.AttrSize).Integer(*tmpfs.Size)
			}
			t.Close()
		}
		a.Close()
	}
}

func serializeLogConfiguration(v *awstypes.LogConfiguration, value smithyjson.Value) {
	o := value.Object()
	defer o.Close()

	if v.LogDriver != "" {
		o.Key("logDriver").String(string(v.LogDriver))
	}
	if v.Options != nil {
		serializeStringMap(v.Options, o.Key("options"))
	}
	if v.SecretOptions != nil {
		serializeSecrets(v.SecretOptions, o.Key("secretOptions"))
	}
}

func serializeMountPoints(v []awstypes.MountPoint, value smithyjson.Value) {
	a := value.Array()
	defer a.Close()

	for _, mount := range v {
		o := a.Value().Object()
		if mount.ContainerPath != nil {
			o.Key("containerPath").String(*mount.ContainerPath)
		}
		if mount.ReadOnly != nil {
			o.Key("readOnly").Boolean(*mount.ReadOnly)
		}
		if mount.SourceVolume != nil {
			o.Key("sourceVolume").String(*mount.SourceVolume)
		}
		o.Close()
	}
}

func serializeNetworkConfiguration(v *awstypes.NetworkConfiguration, value smithyjson.Value) {
	o := value.Object()
	defer o.Close()

	if v.AssignPublicIp != "" {
		o.Key("assignPublicIp").String(string(v.AssignPublicIp))
	}
}

func serializeRepositoryCredentials(v *awstypes.RepositoryCredentials, value smithyjson.Value) {
	o := value.Object()
	defer o.Close()

	if v.CredentialsParameter != nil {
		o.Key("credentialsParameter").String(*v.CredentialsParameter)
	}
}

func serializeResourceRequirements(v []awstypes.ResourceRequirement, value smithyjson.Value) {
	a := value.Array()
	defer a.Close()

	for _, requirement := range v {
		o := a.Value().Object()
		if requirement.Type != "" {
			o.Key(names.AttrType).String(string(requirement.Type))
		}
		if requirement.Value != nil {
			o.Key(names.AttrValue).String(*requirement.Value)
		}
		o.Close()
	}
}

func serializeRuntimePlatform(v *awstypes.RuntimePlatform, value smithyjson.Value) {
	o := value.Object()
	defer o.Close()

	if v.CpuArchitecture != nil {
		o.Key("cpuArchitecture").String(*v.CpuArchitecture)
	}
	if v.OperatingSystemFamily != nil {
		o.Key("operatingSystemFamily").String(*v.OperatingSystemFamily)
	}
}

func serializeSecrets(v []awstypes.Secret, value smithyjson.Value) {
	a := value.Array()
	defer a.Close()

	for _, secret := range v {
		o := a.Value().Object()
		if secret.Name != nil {
			o.Key(names.AttrName).String(*secret.Name)
		}
		if secret.ValueFrom != nil {
			o.Key("valueFrom").String(*secret.ValueFrom)
		}
		o.Close()
	}
}

func serializeUlimits(v []awstypes.Ulimit, value smithyjson.Value) {
	a := value.Array()
	defer a.Close()

	for _, limit := range v {
		o := a.Value().Object()
		if limit.HardLimit != nil {
			o.Key("hardLimit").Integer(*limit.HardLimit)
		}
		if limit.Name != nil {
			o.Key(names.AttrName).String(*limit.Name)
		}
		if limit.SoftLimit != nil {
			o.Key("softLimit").Integer(*limit.SoftLimit)
		}
		o.Close()
	}
}

func serializeVolumes(v []awstypes.Volume, value smithyjson.Value) {
	a := value.Array()
	defer a.Close()

	for _, volume := range v {
		o := a.Value().Object()
		if volume.EfsVolumeConfiguration != nil {
			serializeEFSVolumeConfiguration(volume.EfsVolumeConfiguration, o.Key("efsVolumeConfiguration"))
		}
		if volume.Host != nil {
			h := o.Key("host").Object()
			if volume.Host.SourcePath != nil {
				h.Key("sourcePath").String(*volume.Host.SourcePath)
			}
			h.Close()
		}
		if volume.Name != nil {
			o.Key(names.AttrName).String(*volume.Name)
		}
		if volume.S3filesVolumeConfiguration != nil {
			serializeS3FilesVolumeConfiguration(volume.S3filesVolumeConfiguration, o.Key("s3filesVolumeConfiguration"))
		}
		o.Close()
	}
}

func serializeEFSVolumeConfiguration(v *awstypes.EFSVolumeConfiguration, value smithyjson.Value) {
	o := value.Object()
	defer o.Close()

	if v.AuthorizationConfig != nil {
		a := o.Key("authorizationConfig").Object()
		if v.AuthorizationConfig.AccessPointId != nil {
			a.Key("accessPointId").String(*v.AuthorizationConfig.AccessPointId)
		}
		if v.AuthorizationConfig.Iam != "" {
			a.Key("iam").String(string(v.AuthorizationConfig.Iam))
		}
		a.Close()
	}
	if v.FileSystemId != nil {
		o.Key("fileSystemId").String(*v.FileSystemId)
	}
	if v.RootDirectory != nil {
		o.Key("rootDirectory").String(*v.RootDirectory)
	}
	if v.TransitEncryption != "" {
		o.Key("transitEncryption").String(string(v.TransitEncryption))
	}
	if v.TransitEncryptionPort != nil {
		o.Key("transitEncryptionPort").Integer(*v.TransitEncryptionPort)
	}
}

func serializeS3FilesVolumeConfiguration(v *awstypes.S3FilesVolumeConfiguration, value smithyjson.Value) {
	o := value.Object()
	defer o.Close()

	if v.AccessPointArn != nil {
		o.Key("accessPointArn").String(*v.AccessPointArn)
	}
	if v.FileSystemArn != nil {
		o.Key("fileSystemArn").String(*v.FileSystemArn)
	}
	if v.RootDirectory != nil {
		o.Key("rootDirectory").String(*v.RootDirectory)
	}
	if v.TransitEncryptionPort != nil {
		o.Key("transitEncryptionPort").Integer(*v.TransitEncryptionPort)
	}
}

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
