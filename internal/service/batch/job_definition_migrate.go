// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package batch

import (
	"context"

	"github.com/YakDriver/regexache"
	awstypes "github.com/aws/aws-sdk-go-v2/service/batch/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/structure"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/sdkv2"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func resourceJobDefinitionV0() *schema.Resource {
	return &schema.Resource{
		SchemaVersion: 0,
		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"arn_prefix": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"container_properties": {
				Type:          schema.TypeString,
				Optional:      true,
				ConflictsWith: []string{"ecs_properties", "eks_properties", "node_properties"},
				StateFunc: func(v any) string {
					json, _ := structure.NormalizeJsonString(v)
					return json
				},
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					equal, _ := equivalentContainerPropertiesJSON(old, new)
					return equal
				},
				DiffSuppressOnRefresh: true,
				ValidateFunc:          validJobContainerProperties,
			},
			"deregister_on_new_revision": {
				Type:     schema.TypeBool,
				Default:  true,
				Optional: true,
			},
			"ecs_properties": {
				Type:          schema.TypeString,
				Optional:      true,
				ConflictsWith: []string{"container_properties", "eks_properties", "node_properties"},
				StateFunc: func(v any) string {
					json, _ := structure.NormalizeJsonString(v)
					return json
				},
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					equal, _ := equivalentECSPropertiesJSON(old, new)
					return equal
				},
				DiffSuppressOnRefresh: true,
				ValidateFunc:          validJobECSProperties,
			},
			"eks_properties": {
				Type:          schema.TypeList,
				MaxItems:      1,
				Optional:      true,
				ConflictsWith: []string{"ecs_properties", "container_properties", "node_properties"},
				Elem: &schema.Resource{
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
				},
			},
			names.AttrName: {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validName,
			},
			"node_properties": {
				Type:          schema.TypeString,
				Optional:      true,
				ConflictsWith: []string{"container_properties", "ecs_properties", "eks_properties"},
				StateFunc: func(v any) string {
					json, _ := structure.NormalizeJsonString(v)
					return json
				},
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					equal, _ := equivalentNodePropertiesJSON(old, new)
					return equal
				},
				DiffSuppressOnRefresh: true,
				ValidateFunc:          validJobNodeProperties,
			},
			names.AttrParameters: {
				Type:     schema.TypeMap,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"platform_capabilities": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Schema{
					Type:             schema.TypeString,
					ValidateDiagFunc: enum.Validate[awstypes.PlatformCapability](),
				},
			},
			names.AttrPropagateTags: {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"retry_strategy": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"attempts": {
							Type:         schema.TypeInt,
							Optional:     true,
							ValidateFunc: validation.IntBetween(1, 10),
						},
						"evaluate_on_exit": {
							Type:     schema.TypeList,
							Optional: true,
							ForceNew: true,
							MinItems: 0,
							MaxItems: 5,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									names.AttrAction: {
										Type:             schema.TypeString,
										Required:         true,
										StateFunc:        sdkv2.ToLowerSchemaStateFunc,
										ValidateDiagFunc: enum.ValidateIgnoreCase[awstypes.RetryAction](),
									},
									"on_exit_code": {
										Type:     schema.TypeString,
										Optional: true,
										ValidateFunc: validation.All(
											validation.StringLenBetween(1, 512),
											validation.StringMatch(regexache.MustCompile(`^[0-9]*\*?$`), "must contain only numbers, and can optionally end with an asterisk"),
										),
									},
									"on_reason": {
										Type:     schema.TypeString,
										Optional: true,
										ValidateFunc: validation.All(
											validation.StringLenBetween(1, 512),
											validation.StringMatch(regexache.MustCompile(`^[0-9A-Za-z.:\s]*\*?$`), "must contain letters, numbers, periods, colons, and white space, and can optionally end with an asterisk"),
										),
									},
									"on_status_reason": {
										Type:     schema.TypeString,
										Optional: true,
										ValidateFunc: validation.All(
											validation.StringLenBetween(1, 512),
											validation.StringMatch(regexache.MustCompile(`^[0-9A-Za-z.:\s]*\*?$`), "must contain letters, numbers, periods, colons, and white space, and can optionally end with an asterisk"),
										),
									},
								},
							},
						},
					},
				},
			},
			"revision": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"scheduling_priority": {
				Type:     schema.TypeInt,
				Optional: true,
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
			names.AttrTimeout: {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"attempt_duration_seconds": {
							Type:         schema.TypeInt,
							Optional:     true,
							ValidateFunc: validation.IntAtLeast(60),
						},
					},
				},
			},
			names.AttrType: {
				Type:             schema.TypeString,
				Required:         true,
				ValidateDiagFunc: enum.ValidateIgnoreCase[awstypes.JobDefinitionType](),
			},
		},
	}
}

func jobDefinitionStateUpgradeV0(_ context.Context, rawState map[string]any, meta any) (map[string]any, error) {
	if rawState == nil {
		rawState = map[string]any{}
	}

	// TODO

	// container_properties from JSON to block

	// ecs_properties from JSON to block

	// node_properties from JSON to block

	return rawState, nil
}
