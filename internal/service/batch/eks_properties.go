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
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
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

func eksPropertiesSchema() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"pod_properties": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Required: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"containers": {
							Type:     schema.TypeList,
							Required: true,
							MaxItems: 10,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"args": {
										Type:     schema.TypeList,
										Optional: true,
										Elem:     &schema.Schema{Type: schema.TypeString},
									},
									"command": {
										Type:     schema.TypeList,
										Optional: true,
										Elem:     &schema.Schema{Type: schema.TypeString},
									},
									"env": {
										Type:     schema.TypeSet,
										Optional: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												names.AttrName: {
													Type:     schema.TypeString,
													Required: true,
												},
												names.AttrValue: {
													Type:     schema.TypeString,
													Required: true,
												},
											},
										},
									},
									"image": {
										Type:     schema.TypeString,
										Required: true,
									},
									"image_pull_policy": {
										Type:         schema.TypeString,
										Optional:     true,
										ValidateFunc: validation.StringInSlice(imagePullPolicy_Values(), false),
									},
									names.AttrName: {
										Type:     schema.TypeString,
										Optional: true,
									},
									names.AttrResources: {
										Type:     schema.TypeList,
										Optional: true,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"limits": {
													Type:     schema.TypeMap,
													Optional: true,
													Elem:     &schema.Schema{Type: schema.TypeString},
												},
												"requests": {
													Type:     schema.TypeMap,
													Optional: true,
													Elem:     &schema.Schema{Type: schema.TypeString},
												},
											},
										},
									},
									"security_context": {
										Type:     schema.TypeList,
										Optional: true,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"privileged": {
													Type:     schema.TypeBool,
													Optional: true,
												},
												"read_only_root_file_system": {
													Type:     schema.TypeBool,
													Optional: true,
												},
												"run_as_group": {
													Type:     schema.TypeInt,
													Optional: true,
												},
												"run_as_non_root": {
													Type:     schema.TypeBool,
													Optional: true,
												},
												"run_as_user": {
													Type:     schema.TypeInt,
													Optional: true,
												},
											},
										},
									},
									"volume_mounts": {
										Type:     schema.TypeList,
										Optional: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"mount_path": {
													Type:     schema.TypeString,
													Required: true,
												},
												names.AttrName: {
													Type:     schema.TypeString,
													Required: true,
												},
												"read_only": {
													Type:     schema.TypeBool,
													Optional: true,
												},
											},
										},
									},
								},
							},
						},
						"dns_policy": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringInSlice(dnsPolicy_Values(), false),
						},
						"host_network": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  true,
						},
						"image_pull_secret": {
							Type:     schema.TypeList,
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									names.AttrName: {
										Type:     schema.TypeString,
										Required: true,
									},
								},
							},
						},
						"init_containers": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 10,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"args": {
										Type:     schema.TypeList,
										Optional: true,
										Elem:     &schema.Schema{Type: schema.TypeString},
									},
									"command": {
										Type:     schema.TypeList,
										Optional: true,
										Elem:     &schema.Schema{Type: schema.TypeString},
									},
									"env": {
										Type:     schema.TypeSet,
										Optional: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												names.AttrName: {
													Type:     schema.TypeString,
													Required: true,
												},
												names.AttrValue: {
													Type:     schema.TypeString,
													Required: true,
												},
											},
										},
									},
									"image": {
										Type:     schema.TypeString,
										Required: true,
									},
									"image_pull_policy": {
										Type:         schema.TypeString,
										Optional:     true,
										ValidateFunc: validation.StringInSlice(imagePullPolicy_Values(), false),
									},
									names.AttrName: {
										Type:     schema.TypeString,
										Optional: true,
									},
									names.AttrResources: {
										Type:     schema.TypeList,
										Optional: true,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"limits": {
													Type:     schema.TypeMap,
													Optional: true,
													Elem:     &schema.Schema{Type: schema.TypeString},
												},
												"requests": {
													Type:     schema.TypeMap,
													Optional: true,
													Elem:     &schema.Schema{Type: schema.TypeString},
												},
											},
										},
									},
									"security_context": {
										Type:     schema.TypeList,
										Optional: true,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"privileged": {
													Type:     schema.TypeBool,
													Optional: true,
												},
												"read_only_root_file_system": {
													Type:     schema.TypeBool,
													Optional: true,
												},
												"run_as_group": {
													Type:     schema.TypeInt,
													Optional: true,
												},
												"run_as_non_root": {
													Type:     schema.TypeBool,
													Optional: true,
												},
												"run_as_user": {
													Type:     schema.TypeInt,
													Optional: true,
												},
											},
										},
									},
									"volume_mounts": {
										Type:     schema.TypeList,
										Optional: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"mount_path": {
													Type:     schema.TypeString,
													Required: true,
												},
												names.AttrName: {
													Type:     schema.TypeString,
													Required: true,
												},
												"read_only": {
													Type:     schema.TypeBool,
													Optional: true,
												},
											},
										},
									},
								},
							},
						},
						"metadata": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"labels": {
										Type:     schema.TypeMap,
										Optional: true,
										Elem:     &schema.Schema{Type: schema.TypeString},
									},
								},
							},
						},
						"service_account_name": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"share_process_namespace": {
							Type:     schema.TypeBool,
							Optional: true,
						},
						"volumes": {
							Type:     schema.TypeList,
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"empty_dir": {
										Type:     schema.TypeList,
										MaxItems: 1,
										Optional: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"medium": {
													Type:         schema.TypeString,
													Optional:     true,
													Default:      "",
													ValidateFunc: validation.StringInSlice([]string{"", "Memory"}, true),
												},
												"size_limit": {
													Type:     schema.TypeString,
													Required: true,
												},
											},
										},
									},
									"host_path": {
										Type:     schema.TypeList,
										Optional: true,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												names.AttrPath: {
													Type:     schema.TypeString,
													Required: true,
												},
											},
										},
									},
									names.AttrName: {
										Type:     schema.TypeString,
										Optional: true,
										Default:  "Default",
									},
									"secret": {
										Type:     schema.TypeList,
										Optional: true,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"optional": {
													Type:     schema.TypeBool,
													Optional: true,
												},
												"secret_name": {
													Type:     schema.TypeString,
													Required: true,
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}
}
