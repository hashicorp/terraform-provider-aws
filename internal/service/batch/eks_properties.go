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

type eksProperties awstypes.EksProperties

func (ep *eksProperties) reduce() {
	if ep.PodProperties == nil {
		return
	}
	ep.orderContainers()
	ep.orderEnvironmentVariables()

	// Set all empty slices to nil.
	if len(ep.PodProperties.Containers) == 0 {
		ep.PodProperties.Containers = nil
	} else {
		for j, container := range ep.PodProperties.Containers {
			if len(container.Args) == 0 {
				container.Args = nil
			}
			if len(container.Command) == 0 {
				container.Command = nil
			}
			if len(container.Env) == 0 {
				container.Env = nil
			}
			if len(container.VolumeMounts) == 0 {
				container.VolumeMounts = nil
			}
			ep.PodProperties.Containers[j] = container
		}
	}
	if len(ep.PodProperties.InitContainers) == 0 {
		ep.PodProperties.InitContainers = nil
	} else {
		for j, container := range ep.PodProperties.InitContainers {
			if len(container.Args) == 0 {
				container.Args = nil
			}
			if len(container.Command) == 0 {
				container.Command = nil
			}
			if len(container.Env) == 0 {
				container.Env = nil
			}
			if len(container.VolumeMounts) == 0 {
				container.VolumeMounts = nil
			}
			ep.PodProperties.InitContainers[j] = container
		}
	}
	if ep.PodProperties.DnsPolicy == nil {
		ep.PodProperties.DnsPolicy = aws.String("ClusterFirst")
	}
	if ep.PodProperties.HostNetwork == nil {
		ep.PodProperties.HostNetwork = aws.Bool(true)
	}
	if len(ep.PodProperties.Volumes) == 0 {
		ep.PodProperties.Volumes = nil
	}
	if len(ep.PodProperties.ImagePullSecrets) == 0 {
		ep.PodProperties.ImagePullSecrets = nil
	}
}

func (ep *eksProperties) orderContainers() {
	slices.SortFunc(ep.PodProperties.Containers, func(a, b awstypes.EksContainer) int {
		return cmp.Compare(aws.ToString(a.Name), aws.ToString(b.Name))
	})
}

func (ep *eksProperties) orderEnvironmentVariables() {
	for j, container := range ep.PodProperties.Containers {
		// Remove environment variables with empty values.
		container.Env = tfslices.Filter(container.Env, func(kvp awstypes.EksContainerEnvironmentVariable) bool {
			return aws.ToString(kvp.Value) != ""
		})

		slices.SortFunc(container.Env, func(a, b awstypes.EksContainerEnvironmentVariable) int {
			return cmp.Compare(aws.ToString(a.Name), aws.ToString(b.Name))
		})

		ep.PodProperties.Containers[j].Env = container.Env
	}
}

func equivalentEKSPropertiesJSON(str1, str2 string) (bool, error) {
	if str1 == "" {
		str1 = "{}"
	}

	if str2 == "" {
		str2 = "{}"
	}

	var ep1 eksProperties
	err := tfjson.DecodeFromString(str1, &ep1)
	if err != nil {
		return false, err
	}
	ep1.reduce()
	b1, err := tfjson.EncodeToBytes(ep1)
	if err != nil {
		return false, err
	}

	var ep2 eksProperties
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

func serializeEKSProperties(v *awstypes.EksProperties, value smithyjson.Value) {
	o := value.Object()
	defer o.Close()

	if v.PodProperties != nil {
		serializeEKSPodProperties(v.PodProperties, o.Key("podProperties"))
	}
}

func serializeEKSPodProperties(v *awstypes.EksPodProperties, value smithyjson.Value) {
	o := value.Object()
	defer o.Close()

	if v.Containers != nil {
		serializeEKSContainers(v.Containers, o.Key("containers"))
	}
	if v.DnsPolicy != nil {
		o.Key("dnsPolicy").String(*v.DnsPolicy)
	}
	if v.HostNetwork != nil {
		o.Key("hostNetwork").Boolean(*v.HostNetwork)
	}
	if v.ImagePullSecrets != nil {
		a := o.Key("imagePullSecrets").Array()
		for _, secret := range v.ImagePullSecrets {
			s := a.Value().Object()
			if secret.Name != nil {
				s.Key(names.AttrName).String(*secret.Name)
			}
			s.Close()
		}
		a.Close()
	}
	if v.InitContainers != nil {
		serializeEKSContainers(v.InitContainers, o.Key("initContainers"))
	}
	if v.Metadata != nil {
		m := o.Key("metadata").Object()
		if v.Metadata.Annotations != nil {
			serializeStringMap(v.Metadata.Annotations, m.Key("annotations"))
		}
		if v.Metadata.Labels != nil {
			serializeStringMap(v.Metadata.Labels, m.Key("labels"))
		}
		if v.Metadata.Namespace != nil {
			m.Key(names.AttrNamespace).String(*v.Metadata.Namespace)
		}
		m.Close()
	}
	if v.ServiceAccountName != nil {
		o.Key("serviceAccountName").String(*v.ServiceAccountName)
	}
	if v.ShareProcessNamespace != nil {
		o.Key("shareProcessNamespace").Boolean(*v.ShareProcessNamespace)
	}
	if v.Volumes != nil {
		serializeEKSVolumes(v.Volumes, o.Key("volumes"))
	}
}

func serializeEKSContainers(v []awstypes.EksContainer, value smithyjson.Value) {
	a := value.Array()
	defer a.Close()

	for _, container := range v {
		serializeEKSContainer(&container, a.Value())
	}
}

func serializeEKSContainer(v *awstypes.EksContainer, value smithyjson.Value) {
	o := value.Object()
	defer o.Close()

	if v.Args != nil {
		serializeStringList(v.Args, o.Key("args"))
	}
	if v.Command != nil {
		serializeStringList(v.Command, o.Key("command"))
	}
	if v.Env != nil {
		a := o.Key("env").Array()
		for _, variable := range v.Env {
			e := a.Value().Object()
			if variable.Name != nil {
				e.Key(names.AttrName).String(*variable.Name)
			}
			if variable.Value != nil {
				e.Key(names.AttrValue).String(*variable.Value)
			}
			e.Close()
		}
		a.Close()
	}
	if v.Image != nil {
		o.Key("image").String(*v.Image)
	}
	if v.ImagePullPolicy != nil {
		o.Key("imagePullPolicy").String(*v.ImagePullPolicy)
	}
	if v.Name != nil {
		o.Key(names.AttrName).String(*v.Name)
	}
	if v.Resources != nil {
		r := o.Key(names.AttrResources).Object()
		if v.Resources.Limits != nil {
			serializeStringMap(v.Resources.Limits, r.Key("limits"))
		}
		if v.Resources.Requests != nil {
			serializeStringMap(v.Resources.Requests, r.Key("requests"))
		}
		r.Close()
	}
	if v.SecurityContext != nil {
		serializeEKSContainerSecurityContext(v.SecurityContext, o.Key("securityContext"))
	}
	if v.VolumeMounts != nil {
		a := o.Key("volumeMounts").Array()
		for _, mount := range v.VolumeMounts {
			m := a.Value().Object()
			if mount.MountPath != nil {
				m.Key("mountPath").String(*mount.MountPath)
			}
			if mount.Name != nil {
				m.Key(names.AttrName).String(*mount.Name)
			}
			if mount.ReadOnly != nil {
				m.Key("readOnly").Boolean(*mount.ReadOnly)
			}
			if mount.SubPath != nil {
				m.Key("subPath").String(*mount.SubPath)
			}
			m.Close()
		}
		a.Close()
	}
}

func serializeEKSContainerSecurityContext(v *awstypes.EksContainerSecurityContext, value smithyjson.Value) {
	o := value.Object()
	defer o.Close()

	if v.AllowPrivilegeEscalation != nil {
		o.Key("allowPrivilegeEscalation").Boolean(*v.AllowPrivilegeEscalation)
	}
	if v.Privileged != nil {
		o.Key("privileged").Boolean(*v.Privileged)
	}
	if v.ReadOnlyRootFilesystem != nil {
		o.Key("readOnlyRootFilesystem").Boolean(*v.ReadOnlyRootFilesystem)
	}
	if v.RunAsGroup != nil {
		o.Key("runAsGroup").Long(*v.RunAsGroup)
	}
	if v.RunAsNonRoot != nil {
		o.Key("runAsNonRoot").Boolean(*v.RunAsNonRoot)
	}
	if v.RunAsUser != nil {
		o.Key("runAsUser").Long(*v.RunAsUser)
	}
}

func serializeEKSVolumes(v []awstypes.EksVolume, value smithyjson.Value) {
	a := value.Array()
	defer a.Close()

	for _, volume := range v {
		o := a.Value().Object()
		if volume.EmptyDir != nil {
			d := o.Key("emptyDir").Object()
			if volume.EmptyDir.Medium != nil {
				d.Key("medium").String(*volume.EmptyDir.Medium)
			}
			if volume.EmptyDir.SizeLimit != nil {
				d.Key("sizeLimit").String(*volume.EmptyDir.SizeLimit)
			}
			d.Close()
		}
		if volume.HostPath != nil {
			h := o.Key("hostPath").Object()
			if volume.HostPath.Path != nil {
				h.Key(names.AttrPath).String(*volume.HostPath.Path)
			}
			h.Close()
		}
		if volume.Name != nil {
			o.Key(names.AttrName).String(*volume.Name)
		}
		if volume.PersistentVolumeClaim != nil {
			p := o.Key("persistentVolumeClaim").Object()
			if volume.PersistentVolumeClaim.ClaimName != nil {
				p.Key("claimName").String(*volume.PersistentVolumeClaim.ClaimName)
			}
			if volume.PersistentVolumeClaim.ReadOnly != nil {
				p.Key("readOnly").Boolean(*volume.PersistentVolumeClaim.ReadOnly)
			}
			p.Close()
		}
		if volume.Secret != nil {
			s := o.Key("secret").Object()
			if volume.Secret.Optional != nil {
				s.Key("optional").Boolean(*volume.Secret.Optional)
			}
			if volume.Secret.SecretName != nil {
				s.Key("secretName").String(*volume.Secret.SecretName)
			}
			s.Close()
		}
		o.Close()
	}
}
