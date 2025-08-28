// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package batch

import (
	"context"

	awstypes "github.com/aws/aws-sdk-go-v2/service/batch/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/sdkv2"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func computeEnvironmentSchemaV0() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"compute_environment_name": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ForceNew:      true,
				ConflictsWith: []string{"compute_environment_name_prefix"},
			},
			"compute_environment_name_prefix": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ForceNew:      true,
				ConflictsWith: []string{"compute_environment_name"},
			},
			"compute_resources": {
				Type:     schema.TypeList,
				Optional: true,
				ForceNew: true,
				MinItems: 0,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"allocation_strategy": {
							Type:      schema.TypeString,
							Optional:  true,
							StateFunc: sdkv2.ToUpperSchemaStateFunc,
						},
						"bid_percentage": {
							Type:     schema.TypeInt,
							Optional: true,
						},
						"desired_vcpus": {
							Type:     schema.TypeInt,
							Optional: true,
							Computed: true,
						},
						"ec2_configuration": {
							Type:     schema.TypeList,
							Optional: true,
							Computed: true,
							ForceNew: true,
							MaxItems: 2,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"image_id_override": {
										Type:     schema.TypeString,
										Optional: true,
										Computed: true,
									},
									"image_type": {
										Type:     schema.TypeString,
										Optional: true,
									},
								},
							},
						},
						"ec2_key_pair": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"image_id": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"instance_role": {
							Type:     schema.TypeString,
							Optional: true,
						},
						names.AttrInstanceType: {
							Type:     schema.TypeSet,
							Optional: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						names.AttrLaunchTemplate: {
							Type:     schema.TypeList,
							Optional: true,
							ForceNew: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"launch_template_id": {
										Type:          schema.TypeString,
										Optional:      true,
										ConflictsWith: []string{"compute_resources.0.launch_template.0.launch_template_name"},
									},
									"launch_template_name": {
										Type:          schema.TypeString,
										Optional:      true,
										ConflictsWith: []string{"compute_resources.0.launch_template.0.launch_template_id"},
									},
									names.AttrVersion: {
										Type:     schema.TypeString,
										Optional: true,
										Computed: true,
									},
								},
							},
						},
						"max_vcpus": {
							Type:     schema.TypeInt,
							Required: true,
						},
						"min_vcpus": {
							Type:     schema.TypeInt,
							Optional: true,
						},
						"placement_group": {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
						},
						names.AttrSecurityGroupIDs: {
							Type:     schema.TypeSet,
							Optional: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"spot_iam_fleet_role": {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
						},
						names.AttrSubnets: {
							Type:     schema.TypeSet,
							Required: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						names.AttrTags: tftags.TagsSchema(),
						names.AttrType: {
							Type:      schema.TypeString,
							Required:  true,
							StateFunc: sdkv2.ToUpperSchemaStateFunc,
						},
					},
				},
			},
			"ecs_cluster_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"eks_configuration": {
				Type:     schema.TypeList,
				Optional: true,
				ForceNew: true,
				MinItems: 0,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"eks_cluster_arn": {
							Type:     schema.TypeString,
							Required: true,
							ForceNew: true,
						},
						"kubernetes_namespace": {
							Type:     schema.TypeString,
							Required: true,
							ForceNew: true,
						},
					},
				},
			},
			names.AttrServiceRole: {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			names.AttrState: {
				Type:      schema.TypeString,
				Optional:  true,
				Default:   awstypes.CEStateEnabled,
				StateFunc: sdkv2.ToUpperSchemaStateFunc,
			},
			names.AttrStatus: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrStatusReason: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
			names.AttrType: {
				Type:      schema.TypeString,
				Required:  true,
				ForceNew:  true,
				StateFunc: sdkv2.ToUpperSchemaStateFunc,
			},
			"update_policy": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"job_execution_timeout_minutes": {
							Type:     schema.TypeInt,
							Required: true,
						},
						"terminate_jobs_on_update": {
							Type:     schema.TypeBool,
							Required: true,
						},
					},
				},
			},
		},
	}
}

func computeEnvironmentStateUpgradeV0(_ context.Context, rawState map[string]any, meta any) (map[string]any, error) {
	if rawState == nil {
		rawState = map[string]any{}
	}

	if v, ok := rawState["compute_environment_name"].(string); ok && v != "" {
		rawState[names.AttrName] = v
		delete(rawState, "compute_environment_name")
	}

	if v, ok := rawState["compute_environment_name_prefix"].(string); ok && v != "" {
		rawState[names.AttrNamePrefix] = v
		delete(rawState, "compute_environment_name_prefix")
	}

	return rawState, nil
}
