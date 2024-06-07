package batch

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/batch"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
)

const (
	IpcModeHost                                  = "host"
	IpcModeTask                                  = "task"
	IpcModeNone                                  = "none"
	PicModeHost                                  = "host"
	PicModeTask                                  = "task"
	ResourceRequirementTypeGPU                   = "GPU"
	ResourceRequirementTypeVCPU                  = "VCPU"
	ResourceRequirementTypeMEMORY                = "MEMORY"
	RunTimePlatFormCpuArchX86_64                 = "X86_64"
	RunTimePlatFormCpuArchARM64                  = "ARM64"
	RunTimePlatformOSFamilyLinux                 = "LINUX"
	RunTimePlatformOSFamilyWindowsServer2019Core = "WINDOWS_SERVER_2019_CORE"
	RunTimePlatformOSFamilyWindowsServer2019Full = "WINDOWS_SERVER_2019_FULL"
	RunTimePlatformOSFamilyWindowsServer2022Core = "WINDOWS_SERVER_2022_CORE"
	RunTimePlatformOSFamilyWindowsServer2022Full = "WINDOWS_SERVER_2022_FULL"
	EFSVolumeTransitEncryptionEnabled            = "ENABLED"
	EFSVolumeTransitEncryptionDisabled           = "DISABLED"
	NetworkConfigurationAssignPublicIpEnabled    = "ENABLED"
	NetworkConfigurationAssignPublicIpDisabled   = "DISABLED"
)

func IpcMode_Values() []string {
	return []string{IpcModeHost, IpcModeTask, IpcModeNone}
}

func PidMode_Values() []string {
	return []string{PicModeHost, PicModeTask}
}

func ResourceRequirementType_Values() []string {
	return []string{ResourceRequirementTypeGPU, ResourceRequirementTypeVCPU, ResourceRequirementTypeMEMORY}
}

func RunTimePlatFormCpuArch_Values() []string {
	return []string{RunTimePlatFormCpuArchX86_64, RunTimePlatFormCpuArchARM64}
}

func RunTimePlatformOSFamily_Values() []string {
	return []string{
		RunTimePlatformOSFamilyLinux,
		RunTimePlatformOSFamilyWindowsServer2019Core,
		RunTimePlatformOSFamilyWindowsServer2019Full,
		RunTimePlatformOSFamilyWindowsServer2022Core,
		RunTimePlatformOSFamilyWindowsServer2022Full,
	}
}

func EFSVolumeTransitEncryption_Values() []string {
	return []string{EFSVolumeTransitEncryptionEnabled, EFSVolumeTransitEncryptionDisabled}
}

func NetworkConfigurationAssignPublicIp_Values() []string {
	return []string{NetworkConfigurationAssignPublicIpEnabled, NetworkConfigurationAssignPublicIpDisabled}
}

func expandECSTaskProperties(taskPropsMap map[string]interface{}) *batch.EcsTaskProperties {
	taskProps := &batch.EcsTaskProperties{}

	if v, ok := taskPropsMap["containers"]; ok {
		containers := v.([]interface{})
		taskProps.Containers = expandEcsTaskContainers(containers)
	}

	if v, ok := taskPropsMap["ephemeral_storage"].(*schema.Set); ok && v.Len() > 0 {
		ephemeralStorage := &batch.EphemeralStorage{}
		for _, e := range v.List() {
			if v, ok := e.(map[string]interface{})["size_in_gib"].(int); ok {
				ephemeralStorage.SizeInGiB = aws.Int64(int64(v))
			}
		}
		taskProps.EphemeralStorage = ephemeralStorage
	}

	if v, ok := taskPropsMap["execution_role_arn"].(string); ok && v != "" {
		taskProps.ExecutionRoleArn = aws.String(v)
	}

	if v, ok := taskPropsMap["ipc_mode"].(string); ok && v != "" {
		taskProps.IpcMode = aws.String(v)
	}

	if v, ok := taskPropsMap["network_configuration"].([]interface{}); ok && len(v) > 0 {
		networkConfig := &batch.NetworkConfiguration{}
		if v, ok := v[0].(map[string]interface{})["assign_public_ip"].(string); ok && v != "" {
			networkConfig.AssignPublicIp = aws.String(v)
		}
		taskProps.NetworkConfiguration = networkConfig
	}

	if v, ok := taskPropsMap["pid_mode"].(string); ok && v != "" {
		taskProps.PidMode = aws.String(v)
	}

	if v, ok := taskPropsMap["platform_version"].(string); ok && v != "" {
		taskProps.PlatformVersion = aws.String(v)
	}

	if v, ok := taskPropsMap["runtime_platform"].(*schema.Set); ok && v.Len() > 0 {
		runtimePlatform := &batch.RuntimePlatform{}
		for _, r := range v.List() {
			if v, ok := r.(map[string]interface{})["cpu_architecture"].(string); ok && v != "" {
				runtimePlatform.CpuArchitecture = aws.String(v)
			}
			if v, ok := r.(map[string]interface{})["operation_system_family"].(string); ok && v != "" {
				runtimePlatform.OperatingSystemFamily = aws.String(v)
			}
		}
		taskProps.RuntimePlatform = runtimePlatform
	}

	if v, ok := taskPropsMap["task_role_arn"].(string); ok && v != "" {
		taskProps.TaskRoleArn = aws.String(v)
	}

	//volumes
	if v, ok := taskPropsMap["volumes"].(*schema.Set); ok && v.Len() > 0 {
		volumes := []*batch.Volume{}
		for _, vol := range v.List() {
			volume := &batch.Volume{}

			if v, ok := vol.(map[string]interface{})["efs_volume_configuration"].(map[string]interface{}); ok && len(v) > 0 {
				efs := &batch.EFSVolumeConfiguration{}
				if v, ok := v["file_system_id"].(string); ok && v != "" {
					efs.FileSystemId = aws.String(v)
				}

				if v, ok := v["authorization_config"].(map[string]interface{}); ok && len(v) > 0 {
					authConfig := &batch.EFSAuthorizationConfig{}
					if v, ok := v["access_point_id"].(string); ok && v != "" {
						authConfig.AccessPointId = aws.String(v)
					}
					if v, ok := v["iam"].(string); ok && v != "" {
						authConfig.Iam = aws.String(v)
					}
					efs.AuthorizationConfig = authConfig
				}

				if v, ok := v["root_directory"].(string); ok && v != "" {
					efs.RootDirectory = aws.String(v)
				}
				if v, ok := v["transit_encryption"].(string); ok && v != "" {
					efs.TransitEncryption = aws.String(v)
				}
				if v, ok := v["transit_encryption_port"].(int); ok {
					efs.TransitEncryptionPort = aws.Int64(int64(v))
				}
			}

			if v, ok := vol.(map[string]interface{})["host"].(map[string]interface{}); ok && len(v) > 0 {
				host := &batch.Host{}
				if v, ok := v["source_path"].(string); ok && v != "" {
					host.SourcePath = aws.String(v)
				}
				volume.Host = host
			}
			if v, ok := vol.(map[string]interface{})["name"].(string); ok && v != "" {
				volume.Name = aws.String(v)
			}
		}

		taskProps.Volumes = volumes
	}

	return taskProps
}

func flattenECSProperties(ecsProperties *batch.EcsProperties) []interface{} {
	var ecsPropertiesList []interface{}
	if ecsProperties == nil {
		return ecsPropertiesList
	}
	if v := ecsProperties.TaskProperties; v != nil {
		ecsPropertiesList = append(ecsPropertiesList, map[string]interface{}{
			"task_properties": flattenECSTaskProperties(v[0]),
		})
	}
	return ecsPropertiesList
}

func flattenECSTaskProperties(taskProperties *batch.EcsTaskProperties) (tfList []interface{}) {
	tfMap := make(map[string]interface{}, 0)

	if v := taskProperties.Containers; v != nil {
		tfMap["containers"] = flattenEcsTaskContainers(v)
	}

	if v := taskProperties.EphemeralStorage; v != nil {
		ephemeralMap := make(map[string]interface{}, 0)
		if v := taskProperties.EphemeralStorage.SizeInGiB; v != nil {
			ephemeralMap["size_in_gib"] = aws.Int64Value(v)
		}
		tfMap["ephemeral_storage"] = []interface{}{ephemeralMap}
	}

	if v := taskProperties.ExecutionRoleArn; v != nil {
		tfMap["execution_role_arn"] = aws.StringValue(v)
	}

	if v := taskProperties.IpcMode; v != nil {
		tfMap["ipc_mode"] = aws.StringValue(v)
	}

	if v := taskProperties.NetworkConfiguration; v != nil {
		networkMap := make(map[string]interface{}, 0)
		if v := taskProperties.NetworkConfiguration.AssignPublicIp; v != nil {
			networkMap["assign_public_ip"] = aws.StringValue(v)
		}
		tfMap["network_configuration"] = []interface{}{networkMap}
	}

	if v := taskProperties.PidMode; v != nil {
		tfMap["pid_mode"] = aws.StringValue(v)
	}

	if v := taskProperties.PlatformVersion; v != nil {
		tfMap["platform_version"] = aws.StringValue(v)
	}

	if v := taskProperties.RuntimePlatform; v != nil {
		runtimeMap := make(map[string]interface{}, 0)
		if v := taskProperties.RuntimePlatform.CpuArchitecture; v != nil {
			runtimeMap["cpu_architecture"] = aws.StringValue(v)
		}
		if v := taskProperties.RuntimePlatform.OperatingSystemFamily; v != nil {
			runtimeMap["operation_system_family"] = aws.StringValue(v)
		}
		tfMap["runtime_platform"] = []interface{}{runtimeMap}
	}

	if v := taskProperties.TaskRoleArn; v != nil {
		tfMap["task_role_arn"] = aws.StringValue(v)
	}

	if v := taskProperties.Volumes; v != nil {
		tfMap["volumes"] = flattenEcsTaskVolumes(v)
	}

	return append(tfList, tfMap)
}

func flattenEcsTaskContainers(containers []*batch.TaskContainerProperties) (tfList []interface{}) {
	for _, container := range containers {
		tfMap := make(map[string]interface{}, 0)

		if v := container.Image; v != nil {
			tfMap["image"] = aws.StringValue(v)
		}

		if v := container.Command; v != nil {
			tfMap["command"] = flex.FlattenStringList(v)
		}

		if v := container.DependsOn; v != nil {
			tfList := make([]interface{}, 0)
			for _, dep := range v {
				depMap := make(map[string]interface{}, 0)
				if v := dep.Condition; v != nil {
					depMap["condition"] = aws.StringValue(v)
				}
				if v := dep.ContainerName; v != nil {
					depMap["container_name"] = aws.StringValue(v)
				}
				tfList = append(tfList, depMap)
			}
			tfMap["depends_on"] = tfList
		}

		if v := container.Environment; v != nil {
			tfList := make([]interface{}, 0)
			for _, env := range v {
				envMap := make(map[string]interface{}, 0)
				if v := env.Name; v != nil {
					envMap["name"] = aws.StringValue(v)
				}
				if v := env.Value; v != nil {
					envMap["value"] = aws.StringValue(v)
				}
				tfList = append(tfList, envMap)
			}
			tfMap["environment"] = tfList
		}

		if v := container.Essential; v != nil {
			tfMap["essential"] = aws.BoolValue(v)
		}

		if v := container.LinuxParameters; v != nil {
			linuxMap := make(map[string]interface{}, 0)

			if v := v.Devices; v != nil {
				tfList := make([]interface{}, 0)
				for _, dev := range v {
					devMap := make(map[string]interface{}, 0)
					if v := dev.ContainerPath; v != nil {
						devMap["container_path"] = aws.StringValue(v)
					}
					if v := dev.HostPath; v != nil {
						devMap["host_path"] = aws.StringValue(v)
					}
					if v := dev.Permissions; v != nil {
						devMap["permissions"] = flex.FlattenStringList(v)
					}
					tfList = append(tfList, devMap)
				}
				linuxMap["devices"] = tfList
			}

			if v := v.InitProcessEnabled; v != nil {
				linuxMap["init_process_enabled"] = aws.BoolValue(v)
			}

			if v := v.MaxSwap; v != nil {
				linuxMap["max_swap"] = aws.Int64Value(v)
			}

			if v := v.SharedMemorySize; v != nil {
				linuxMap["shared_memory_size"] = aws.Int64Value(v)
			}

			if v := v.Swappiness; v != nil {
				linuxMap["swappiness"] = aws.Int64Value(v)
			}

			if v := v.Tmpfs; v != nil {
				tfList := make([]interface{}, 0)
				for _, tmp := range v {
					tmpMap := make(map[string]interface{}, 0)
					if v := tmp.ContainerPath; v != nil {
						tmpMap["container_path"] = aws.StringValue(v)
					}
					if v := tmp.Size; v != nil {
						tmpMap["size"] = aws.Int64Value(v)
					}
					if v := tmp.MountOptions; v != nil {
						tmpMap["mount_options"] = flex.FlattenStringList(v)
					}
					tfList = append(tfList, tmpMap)
				}
				linuxMap["tmpfs"] = tfList
			}

			tfMap["linux_parameters"] = linuxMap
		}

		if v := container.LogConfiguration; v != nil {
			tfList := make([]interface{}, 0)
			logMap := make(map[string]interface{}, 0)
			if v := v.LogDriver; v != nil {
				logMap["log_driver"] = aws.StringValue(v)
			}
			if v := v.Options; v != nil {
				logMap["options"] = v
			}
			if v := v.SecretOptions; v != nil {
				tfList := make([]interface{}, 0)
				for _, secret := range v {
					secretMap := make(map[string]interface{}, 0)
					if v := secret.Name; v != nil {
						secretMap["name"] = aws.StringValue(v)
					}
					if v := secret.ValueFrom; v != nil {
						secretMap["value_from"] = aws.StringValue(v)
					}
					tfList = append(tfList, secretMap)
				}
				logMap["secret_options"] = tfList
			}
			tfList = append(tfList, logMap)
			tfMap["log_configuration"] = tfList
		}

		if v := container.MountPoints; v != nil {
			tfList := make([]interface{}, 0)
			for _, mp := range v {
				mpMap := make(map[string]interface{}, 0)
				if v := mp.ContainerPath; v != nil {
					mpMap["container_path"] = aws.StringValue(v)
				}
				if v := mp.ReadOnly; v != nil {
					mpMap["read_only"] = aws.BoolValue(v)
				}
				if v := mp.SourceVolume; v != nil {
					mpMap["source_volume"] = aws.StringValue(v)
				}
				tfList = append(tfList, mpMap)
			}
			tfMap["mount_points"] = tfList
		}

		if v := container.Name; v != nil {
			tfMap["name"] = aws.StringValue(v)
		}

		if v := container.Privileged; v != nil {
			tfMap["privileged"] = aws.BoolValue(v)
		}

		if v := container.ReadonlyRootFilesystem; v != nil {
			tfMap["readonly_root_filesystem"] = aws.BoolValue(v)
		}

		if v := container.RepositoryCredentials; v != nil {
			tfList := make([]interface{}, 0)
			repoMap := make(map[string]interface{}, 0)
			if v := v.CredentialsParameter; v != nil {
				repoMap["credentials_parameter"] = aws.StringValue(v)
			}
			tfList = append(tfList, repoMap)
			tfMap["repository_credentials"] = tfList
		}

		if v := container.ResourceRequirements; v != nil {
			tfList := make([]interface{}, 0)
			for _, req := range v {
				reqMap := make(map[string]interface{}, 0)
				if v := req.Type; v != nil {
					reqMap["type"] = aws.StringValue(v)
				}
				if v := req.Value; v != nil {
					reqMap["value"] = aws.StringValue(v)
				}
				tfList = append(tfList, reqMap)
			}
			tfMap["resource_requirements"] = tfList
		}

		if v := container.Secrets; v != nil {
			tfList := make([]interface{}, 0)
			for _, secret := range v {
				secretMap := make(map[string]interface{}, 0)
				if v := secret.Name; v != nil {
					secretMap["name"] = aws.StringValue(v)
				}
				if v := secret.ValueFrom; v != nil {
					secretMap["value_from"] = aws.StringValue(v)
				}
				tfList = append(tfList, secretMap)
			}
			tfMap["secrets"] = tfList
		}

		if v := container.Ulimits; v != nil {
			tfList := make([]interface{}, 0)
			for _, ulimit := range v {
				ulimitMap := make(map[string]interface{}, 0)
				if v := ulimit.HardLimit; v != nil {
					ulimitMap["hard_limit"] = aws.Int64Value(v)
				}
				if v := ulimit.Name; v != nil {
					ulimitMap["name"] = aws.StringValue(v)
				}
				if v := ulimit.SoftLimit; v != nil {
					ulimitMap["soft_limit"] = aws.Int64Value(v)
				}
				tfList = append(tfList, ulimitMap)
			}
			tfMap["ulimits"] = tfList
		}

		if v := container.User; v != nil {
			tfMap["user"] = aws.StringValue(v)
		}

		tfList = append(tfList, tfMap)
	}

	return tfList
}

func expandEcsTaskContainers(containers []interface{}) []*batch.TaskContainerProperties {
	var result []*batch.TaskContainerProperties

	for _, v := range containers {
		containerMap := v.(map[string]interface{})
		container := &batch.TaskContainerProperties{}

		if v, ok := containerMap["image"]; ok {
			container.Image = aws.String(v.(string))
		}

		if v, ok := containerMap["command"]; ok {
			container.Command = flex.ExpandStringList(v.([]interface{}))
		}

		if v, ok := containerMap["depends_on"].(*schema.Set); ok && v.Len() > 0 {
			deps := []*batch.TaskContainerDependency{}
			for _, d := range v.List() {
				dep := &batch.TaskContainerDependency{}
				if v, ok := d.(map[string]interface{})["condition"].(string); ok && v != "" {
					dep.Condition = aws.String(v)
				}
				if v, ok := d.(map[string]interface{})["container_name"].(string); ok && v != "" {
					dep.ContainerName = aws.String(v)
				}
				deps = append(deps, dep)
			}
			container.DependsOn = deps
		}

		if v, ok := containerMap["environment"].(*schema.Set); ok && v.Len() > 0 {
			envs := []*batch.KeyValuePair{}
			for _, e := range v.List() {
				env := &batch.KeyValuePair{}
				if v, ok := e.(map[string]interface{})["name"].(string); ok && v != "" {
					env.Name = aws.String(v)
				}
				if v, ok := e.(map[string]interface{})["value"].(string); ok && v != "" {
					env.Value = aws.String(v)
				}
				envs = append(envs, env)
			}
			container.Environment = envs
		}

		if v, ok := containerMap["essential"]; ok {
			container.Essential = aws.Bool(v.(bool))
		}

		if v, ok := containerMap["linux_parameters"].(map[string]interface{}); ok {
			param := &batch.LinuxParameters{}

			if v, ok := v["devices"].(*schema.Set); ok && v.Len() > 0 {
				devices := []*batch.Device{}
				for _, d := range v.List() {
					device := &batch.Device{}
					if v, ok := d.(map[string]interface{})["container_path"].(string); ok && v != "" {
						device.ContainerPath = aws.String(v)
					}
					if v, ok := d.(map[string]interface{})["host_path"].(string); ok && v != "" {
						device.HostPath = aws.String(v)
					}
					if v, ok := d.(map[string]interface{})["permissions"].(*schema.Set); ok && v.Len() > 0 {
						permissions := []*string{}
						for _, p := range v.List() {
							permissions = append(permissions, aws.String(p.(string)))
						}
						device.Permissions = permissions
					}
					devices = append(devices, device)
				}
				param.Devices = devices
			}

			if v, ok := v["init_process_enabled"].(bool); ok {
				param.InitProcessEnabled = aws.Bool(v)
			}

			if v, ok := v["max_swap"].(int); ok {
				param.MaxSwap = aws.Int64(int64(v))
			}

			if v, ok := v["shared_memory_size"].(int); ok {
				param.SharedMemorySize = aws.Int64(int64(v))
			}

			if v, ok := v["swappiness"].(int); ok {
				param.Swappiness = aws.Int64(int64(v))
			}

			if v, ok := v["tmpfs"].(*schema.Set); ok && v.Len() > 0 {
				tmpfs := []*batch.Tmpfs{}
				for _, t := range v.List() {
					tmp := &batch.Tmpfs{}
					if v, ok := t.(map[string]interface{})["container_path"].(string); ok && v != "" {
						tmp.ContainerPath = aws.String(v)
					}
					if v, ok := t.(map[string]interface{})["size"].(int); ok {
						tmp.Size = aws.Int64(int64(v))
					}
					if v, ok := t.(map[string]interface{})["mount_options"]; ok {
						tmp.MountOptions = flex.ExpandStringList(v.([]interface{}))
					}
					tmpfs = append(tmpfs, tmp)
				}
				param.Tmpfs = tmpfs
			}

			container.LinuxParameters = param
		}

		if v, ok := containerMap["log_configuration"].([]interface{}); ok && len(v) > 0 {
			logConfig := &batch.LogConfiguration{}

			raw := v[0].(map[string]interface{})

			if v, ok := raw["log_driver"].(string); ok && v != "" {
				logConfig.LogDriver = aws.String(v)
			}
			if v, ok := raw["options"].(map[string]interface{}); ok && len(v) > 0 {
				logConfig.Options = flex.ExpandStringMap(v)
			}

			if v, ok := raw["secret_options"].([]interface{}); ok && len(v) > 0 {
				secretOptions := []*batch.Secret{}
				for _, s := range v {
					secret := &batch.Secret{}
					if v, ok := s.(map[string]interface{})["name"].(string); ok && v != "" {
						secret.Name = aws.String(v)
					}
					if v, ok := s.(map[string]interface{})["value_from"].(string); ok && v != "" {
						secret.ValueFrom = aws.String(v)
					}
					secretOptions = append(secretOptions, secret)
				}
				logConfig.SecretOptions = secretOptions
			}

			container.LogConfiguration = logConfig
		}

		if v, ok := containerMap["mount_points"].(*schema.Set); ok && v.Len() > 0 {
			mountPoints := []*batch.MountPoint{}
			for _, mp := range v.List() {
				mountPoint := &batch.MountPoint{}
				if v, ok := mp.(map[string]interface{})["container_path"].(string); ok && v != "" {
					mountPoint.ContainerPath = aws.String(v)
				}
				if v, ok := mp.(map[string]interface{})["read_only"].(bool); ok {
					mountPoint.ReadOnly = aws.Bool(v)
				}
				if v, ok := mp.(map[string]interface{})["source_volume"].(string); ok && v != "" {
					mountPoint.SourceVolume = aws.String(v)
				}
				mountPoints = append(mountPoints, mountPoint)
			}
			container.MountPoints = mountPoints
		}

		if v, ok := containerMap["name"]; ok && v != "" {
			container.Name = aws.String(v.(string))
		}

		if v, ok := containerMap["privileged"].(bool); ok {
			container.Privileged = aws.Bool(v)
		}

		if v, ok := containerMap["readonly_root_filesystem"].(bool); ok {
			container.ReadonlyRootFilesystem = aws.Bool(v)
		}

		if v, ok := containerMap["repository_credentials"].([]interface{}); ok && len(v) > 0 {
			repoCreds := &batch.RepositoryCredentials{}
			if v, ok := v[0].(map[string]interface{})["credentials_parameter"].(string); ok && v != "" {
				repoCreds.CredentialsParameter = aws.String(v)
			}
			container.RepositoryCredentials = repoCreds
		}

		if v, ok := containerMap["resource_requirements"].(*schema.Set); ok && v.Len() > 0 {
			resources := []*batch.ResourceRequirement{}
			for _, r := range v.List() {
				resource := &batch.ResourceRequirement{}
				if v, ok := r.(map[string]interface{})["type"].(string); ok && v != "" {
					resource.Type = aws.String(v)
				}
				if v, ok := r.(map[string]interface{})["value"].(string); ok && v != "" {
					resource.Value = aws.String(v)
				}
				resources = append(resources, resource)
			}
			container.ResourceRequirements = resources
		}

		if v, ok := containerMap["secrets"].(*schema.Set); ok && v.Len() > 0 {
			secrets := []*batch.Secret{}
			for _, s := range v.List() {
				secret := &batch.Secret{}
				if v, ok := s.(map[string]interface{})["name"].(string); ok && v != "" {
					secret.Name = aws.String(v)
				}
				if v, ok := s.(map[string]interface{})["value_from"].(string); ok && v != "" {
					secret.ValueFrom = aws.String(v)
				}
				secrets = append(secrets, secret)
			}
			container.Secrets = secrets
		}

		if v, ok := containerMap["ulimits"].(*schema.Set); ok && v.Len() > 0 {
			ulimits := []*batch.Ulimit{}
			for _, u := range v.List() {
				ulimit := &batch.Ulimit{}
				if v, ok := u.(map[string]interface{})["hard_limit"].(int); ok {
					ulimit.HardLimit = aws.Int64(int64(v))
				}
				if v, ok := u.(map[string]interface{})["name"].(string); ok && v != "" {
					ulimit.Name = aws.String(v)
				}
				if v, ok := u.(map[string]interface{})["soft_limit"].(int); ok {
					ulimit.SoftLimit = aws.Int64(int64(v))
				}
				ulimits = append(ulimits, ulimit)
			}
			container.Ulimits = ulimits
		}

		if v, ok := containerMap["user"].(string); ok && v != "" {
			container.User = aws.String(v)
		}

		result = append(result, container)
	}

	return result
}

func flattenEcsTaskVolumes(volumes []*batch.Volume) (tfList []interface{}) {
	for _, volume := range volumes {
		tfMap := make(map[string]interface{}, 0)

		if v := volume.EfsVolumeConfiguration; v != nil {
			efsMap := make(map[string]interface{}, 0)
			if v := volume.EfsVolumeConfiguration.FileSystemId; v != nil {
				efsMap["file_system_id"] = aws.StringValue(v)
			}

			if v := volume.EfsVolumeConfiguration.AuthorizationConfig; v != nil {
				authMap := make(map[string]interface{}, 0)
				if v := volume.EfsVolumeConfiguration.AuthorizationConfig.AccessPointId; v != nil {
					authMap["access_point_id"] = aws.StringValue(v)
				}
				if v := volume.EfsVolumeConfiguration.AuthorizationConfig.Iam; v != nil {
					authMap["iam"] = aws.StringValue(v)
				}
				efsMap["authorization_config"] = authMap
			}

			if v := volume.EfsVolumeConfiguration.RootDirectory; v != nil {
				efsMap["root_directory"] = aws.StringValue(v)
			}
			if v := volume.EfsVolumeConfiguration.TransitEncryption; v != nil {
				efsMap["transit_encryption"] = aws.StringValue(v)
			}

			if v := volume.EfsVolumeConfiguration.TransitEncryptionPort; v != nil {
				efsMap["transit_encryption_port"] = aws.Int64Value(v)
			}
			tfMap["efs_volume_configuration"] = efsMap
		}

		if v := volume.Host; v != nil {
			hostMap := make(map[string]interface{}, 0)
			if v := volume.Host.SourcePath; v != nil {
				hostMap["source_path"] = aws.StringValue(v)
			}
			tfMap["host"] = hostMap
		}

		if v := volume.Name; v != nil {
			tfMap["name"] = aws.StringValue(v)
		}

		tfList = append(tfList, tfMap)
	}

	return tfList
}
