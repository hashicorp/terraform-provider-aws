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
	tfjson "github.com/hashicorp/terraform-provider-aws/internal/json"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
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
