// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package batch

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	awstypes "github.com/aws/aws-sdk-go-v2/service/batch/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
)

const (
	ImagePullPolicyAlways       = "Always"
	ImagePullPolicyIfNotPresent = "IfNotPresent"
	ImagePullPolicyNever        = "Never"
)

func ImagePullPolicy_Values() []string {
	return []string{
		ImagePullPolicyAlways,
		ImagePullPolicyIfNotPresent,
		ImagePullPolicyNever,
	}
}

const (
	DNSPolicyDefault                 = "Default"
	DNSPolicyClusterFirst            = "ClusterFirst"
	DNSPolicyClusterFirstWithHostNet = "ClusterFirstWithHostNet"
)

func DNSPolicy_Values() []string {
	return []string{
		DNSPolicyDefault,
		DNSPolicyClusterFirst,
		DNSPolicyClusterFirstWithHostNet,
	}
}

func expandEKSPodProperties(podPropsMap map[string]interface{}) *awstypes.EksPodProperties {
	podProps := &awstypes.EksPodProperties{}

	if v, ok := podPropsMap["containers"]; ok {
		containers := v.([]interface{})
		podProps.Containers = expandContainers(containers)
	}

	if v, ok := podPropsMap["dns_policy"].(string); ok && v != "" {
		podProps.DnsPolicy = aws.String(v)
	}

	if v, ok := podPropsMap["host_network"]; ok {
		podProps.HostNetwork = aws.Bool(v.(bool))
	}
	if m, ok := podPropsMap["metadata"].([]interface{}); ok && len(m) > 0 {
		if v, ok := m[0].(map[string]interface{})["labels"]; ok {
			podProps.Metadata = &awstypes.EksMetadata{}
			podProps.Metadata.Labels = flex.ExpandStringValueMap(v.(map[string]interface{}))
		}
	}
	if v, ok := podPropsMap["service_account_name"].(string); ok && v != "" {
		podProps.ServiceAccountName = aws.String(v)
	}
	if v, ok := podPropsMap["volumes"]; ok {
		podProps.Volumes = expandVolumes(v.([]interface{}))
	}

	return podProps
}

func expandContainers(containers []interface{}) []awstypes.EksContainer {
	var result []awstypes.EksContainer

	for _, v := range containers {
		containerMap := v.(map[string]interface{})
		container := awstypes.EksContainer{}

		if v, ok := containerMap["args"]; ok {
			container.Args = flex.ExpandStringValueList(v.([]interface{}))
		}

		if v, ok := containerMap["command"]; ok {
			container.Command = flex.ExpandStringValueList(v.([]interface{}))
		}

		if v, ok := containerMap["env"].(*schema.Set); ok && v.Len() > 0 {
			env := []awstypes.EksContainerEnvironmentVariable{}
			for _, e := range v.List() {
				environment := awstypes.EksContainerEnvironmentVariable{}
				environ := e.(map[string]interface{})
				if v, ok := environ["name"].(string); ok && v != "" {
					environment.Name = aws.String(v)
				}
				if v, ok := environ["value"].(string); ok && v != "" {
					environment.Value = aws.String(v)
				}
				env = append(env, environment)
			}
			container.Env = env
		}

		if v, ok := containerMap["image"]; ok {
			container.Image = aws.String(v.(string))
		}
		if v, ok := containerMap["image_pull_policy"].(string); ok && v != "" {
			container.ImagePullPolicy = aws.String(v)
		}

		if v, ok := containerMap["name"].(string); ok && v != "" {
			container.Name = aws.String(v)
		}
		if r, ok := containerMap["resources"].([]interface{}); ok && len(r) > 0 {
			resources := &awstypes.EksContainerResourceRequirements{}
			res := r[0].(map[string]interface{})
			if v, ok := res["limits"]; ok {
				resources.Limits = flex.ExpandStringValueMap(v.(map[string]interface{}))
			}
			if v, ok := res["requests"]; ok {
				resources.Requests = flex.ExpandStringValueMap(v.(map[string]interface{}))
			}
			container.Resources = resources
		}

		if s, ok := containerMap["security_context"].([]interface{}); ok && len(s) > 0 {
			securityContext := &awstypes.EksContainerSecurityContext{}
			security := s[0].(map[string]interface{})
			if v, ok := security["privileged"]; ok {
				securityContext.Privileged = aws.Bool(v.(bool))
			}
			if v, ok := security["run_as_user"]; ok {
				securityContext.RunAsUser = aws.Int64(int64(v.(int)))
			}
			if v, ok := security["run_as_group"]; ok {
				securityContext.RunAsGroup = aws.Int64(int64(v.(int)))
			}
			if v, ok := security["read_only_root_file_system"]; ok {
				securityContext.ReadOnlyRootFilesystem = aws.Bool(v.(bool))
			}
			if v, ok := security["run_as_non_root"]; ok {
				securityContext.RunAsNonRoot = aws.Bool(v.(bool))
			}
			container.SecurityContext = securityContext
		}
		if v, ok := containerMap["volume_mounts"]; ok {
			container.VolumeMounts = expandVolumeMounts(v.([]interface{}))
		}

		result = append(result, container)
	}

	return result
}

func expandVolumes(volumes []interface{}) []awstypes.EksVolume {
	var result []awstypes.EksVolume
	for _, v := range volumes {
		volume := awstypes.EksVolume{}
		volumeMap := v.(map[string]interface{})
		if v, ok := volumeMap["name"].(string); ok {
			volume.Name = aws.String(v)
		}
		if e, ok := volumeMap["empty_dir"].([]interface{}); ok && len(e) > 0 {
			if empty, ok := e[0].(map[string]interface{}); ok {
				volume.EmptyDir = &awstypes.EksEmptyDir{
					Medium:    aws.String(empty["medium"].(string)),
					SizeLimit: aws.String(empty["size_limit"].(string)),
				}
			}
		}
		if h, ok := volumeMap["host_path"].([]interface{}); ok && len(h) > 0 {
			volume.HostPath = &awstypes.EksHostPath{}
			if host, ok := h[0].(map[string]interface{}); ok {
				if v, ok := host["path"]; ok {
					volume.HostPath.Path = aws.String(v.(string))
				}
			}
		}
		if s, ok := volumeMap["secret"].([]interface{}); ok && len(s) > 0 {
			volume.Secret = &awstypes.EksSecret{}
			if secret := s[0].(map[string]interface{}); ok {
				if v, ok := secret["secret_name"]; ok {
					volume.Secret.SecretName = aws.String(v.(string))
				}
				if v, ok := secret["optional"]; ok {
					volume.Secret.Optional = aws.Bool(v.(bool))
				}
			}
		}
		result = append(result, volume)
	}

	return result
}

func expandVolumeMounts(volumeMounts []interface{}) []awstypes.EksContainerVolumeMount {
	var result []awstypes.EksContainerVolumeMount
	for _, v := range volumeMounts {
		volumeMount := awstypes.EksContainerVolumeMount{}
		volumeMountMap := v.(map[string]interface{})
		if v, ok := volumeMountMap["name"]; ok {
			volumeMount.Name = aws.String(v.(string))
		}
		if v, ok := volumeMountMap["mount_path"]; ok {
			volumeMount.MountPath = aws.String(v.(string))
		}
		if v, ok := volumeMountMap["read_only"]; ok {
			volumeMount.ReadOnly = aws.Bool(v.(bool))
		}
		result = append(result, volumeMount)
	}

	return result
}

func flattenEKSProperties(eksProperties *awstypes.EksProperties) []interface{} {
	var eksPropertiesList []interface{}
	if eksProperties == nil {
		return eksPropertiesList
	}
	if v := eksProperties.PodProperties; v != nil {
		eksPropertiesList = append(eksPropertiesList, map[string]interface{}{
			"pod_properties": flattenEKSPodProperties(eksProperties.PodProperties),
		})
	}

	return eksPropertiesList
}

func flattenEKSPodProperties(podProperties *awstypes.EksPodProperties) (tfList []interface{}) {
	tfMap := make(map[string]interface{}, 0)
	if v := podProperties.Containers; v != nil {
		tfMap["containers"] = flattenEKSContainers(v)
	}

	if v := podProperties.DnsPolicy; v != nil {
		tfMap["dns_policy"] = aws.ToString(v)
	}

	if v := podProperties.HostNetwork; v != nil {
		tfMap["host_network"] = aws.ToBool(v)
	}

	if v := podProperties.Metadata; v != nil {
		metaData := make([]map[string]interface{}, 0)
		if v := v.Labels; v != nil {
			metaData = append(metaData, map[string]interface{}{"labels": v})
		}
		tfMap["metadata"] = metaData
	}

	if v := podProperties.ServiceAccountName; v != nil {
		tfMap["service_account_name"] = aws.ToString(v)
	}

	if v := podProperties.Volumes; v != nil {
		tfMap["volumes"] = flattenEKSVolumes(v)
	}

	tfList = append(tfList, tfMap)
	return tfList
}

func flattenEKSContainers(containers []awstypes.EksContainer) (tfList []interface{}) {
	for _, container := range containers {
		tfMap := map[string]interface{}{}

		if v := container.Args; v != nil {
			tfMap["args"] = flex.FlattenStringValueList(v)
		}

		if v := container.Command; v != nil {
			tfMap["command"] = flex.FlattenStringValueList(v)
		}

		if v := container.Env; v != nil {
			tfMap["env"] = flattenEKSContainerEnvironmentVariables(v)
		}

		if v := container.Image; v != nil {
			tfMap["image"] = aws.ToString(v)
		}

		if v := container.ImagePullPolicy; v != nil {
			tfMap["image_pull_policy"] = aws.ToString(v)
		}

		if v := container.Name; v != nil {
			tfMap["name"] = aws.ToString(v)
		}

		if v := container.Resources; v != nil {
			tfMap["resources"] = []map[string]interface{}{{
				"limits":   v.Limits,
				"requests": v.Requests,
			}}
		}

		if v := container.SecurityContext; v != nil {
			tfMap["security_context"] = []map[string]interface{}{{
				"privileged":                 aws.ToBool(v.Privileged),
				"run_as_user":                aws.ToInt64(v.RunAsUser),
				"run_as_group":               aws.ToInt64(v.RunAsGroup),
				"read_only_root_file_system": aws.ToBool(v.ReadOnlyRootFilesystem),
				"run_as_non_root":            aws.ToBool(v.RunAsNonRoot),
			}}
		}

		if v := container.VolumeMounts; v != nil {
			tfMap["volume_mounts"] = flattenEKSContainerVolumeMounts(v)
		}
		tfList = append(tfList, tfMap)
	}

	return tfList
}

func flattenEKSContainerEnvironmentVariables(env []awstypes.EksContainerEnvironmentVariable) (tfList []interface{}) {
	for _, e := range env {
		tfMap := map[string]interface{}{}

		if v := e.Name; v != nil {
			tfMap["name"] = aws.ToString(v)
		}

		if v := e.Value; v != nil {
			tfMap["value"] = aws.ToString(v)
		}
		tfList = append(tfList, tfMap)
	}

	return tfList
}

func flattenEKSContainerVolumeMounts(volumeMounts []awstypes.EksContainerVolumeMount) (tfList []interface{}) {
	for _, v := range volumeMounts {
		tfMap := map[string]interface{}{}

		if v := v.Name; v != nil {
			tfMap["name"] = aws.ToString(v)
		}

		if v := v.MountPath; v != nil {
			tfMap["mount_path"] = aws.ToString(v)
		}

		if v := v.ReadOnly; v != nil {
			tfMap["read_only"] = aws.ToBool(v)
		}
		tfList = append(tfList, tfMap)
	}

	return tfList
}

func flattenEKSVolumes(volumes []awstypes.EksVolume) (tfList []interface{}) {
	for _, v := range volumes {
		tfMap := map[string]interface{}{}

		if v := v.Name; v != nil {
			tfMap["name"] = aws.ToString(v)
		}

		if v := v.EmptyDir; v != nil {
			tfMap["empty_dir"] = []map[string]interface{}{{
				"medium":     aws.ToString(v.Medium),
				"size_limit": aws.ToString(v.SizeLimit),
			}}
		}

		if v := v.HostPath; v != nil {
			tfMap["host_path"] = []map[string]interface{}{{
				"path": aws.ToString(v.Path),
			}}
		}

		if v := v.Secret; v != nil {
			tfMap["secret"] = []map[string]interface{}{{
				"secret_name": aws.ToString(v.SecretName),
				"optional":    aws.ToBool(v.Optional),
			}}
		}
		tfList = append(tfList, tfMap)
	}

	return tfList
}
