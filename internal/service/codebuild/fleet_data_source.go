// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package codebuild

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_codebuild_fleet", name="Fleet")
func dataSourceFleet() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceFleetRead,

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"base_capacity": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"compute_configuration": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"disk": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"machine_type": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"memory": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"vcpu": {
							Type:     schema.TypeInt,
							Computed: true,
						},
					},
				},
			},
			"compute_type": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"created": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"environment_type": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"fleet_service_role": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrID: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"image_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"last_modified": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrName: {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringLenBetween(2, 128),
			},
			"overflow_behavior": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"scaling_configuration": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"desired_capacity": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						names.AttrMaxCapacity: {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"scaling_type": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"target_tracking_scaling_configs": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"metric_type": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"target_value": {
										Type:     schema.TypeFloat,
										Computed: true,
									},
								},
							},
						},
					},
				},
			},
			names.AttrStatus: {
				Type:     schema.TypeSet,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"context": {
							Type:     schema.TypeString,
							Computed: true,
						},
						names.AttrMessage: {
							Type:     schema.TypeString,
							Computed: true,
						},
						names.AttrStatusCode: {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			names.AttrTags: tftags.TagsSchemaComputed(),
			names.AttrVPCConfig: {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrSecurityGroupIDs: {
							Type:     schema.TypeSet,
							Computed: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						names.AttrSubnets: {
							Type:     schema.TypeSet,
							Computed: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						names.AttrVPCID: {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
		},
	}
}

const (
	dsNameFleet = "Fleet Data Source"
)

func dataSourceFleetRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).CodeBuildClient(ctx)
	name := d.Get(names.AttrName).(string)

	fleet, err := findFleetByARN(ctx, conn, name)

	if err != nil {
		return create.AppendDiagError(diags, names.CodeBuild, create.ErrActionReading, dsNameFleet, name, err)
	}

	d.SetId(aws.ToString(fleet.Arn))
	d.Set(names.AttrARN, fleet.Arn)
	d.Set("base_capacity", fleet.BaseCapacity)

	if fleet.ComputeConfiguration != nil {
		if err := d.Set("compute_configuration", []any{flattenComputeConfiguration(fleet.ComputeConfiguration)}); err != nil {
			return create.AppendDiagError(diags, names.CodeBuild, create.ErrActionSetting, dsNameFleet, d.Id(), err)
		}
	}

	d.Set("compute_type", fleet.ComputeType)
	d.Set("created", aws.ToTime(fleet.Created).Format(time.RFC3339))
	d.Set("environment_type", fleet.EnvironmentType)
	d.Set("fleet_service_role", fleet.FleetServiceRole)
	d.Set("image_id", fleet.ImageId)
	d.Set("last_modified", aws.ToTime(fleet.LastModified).Format(time.RFC3339))
	d.Set(names.AttrName, fleet.Name)
	d.Set("overflow_behavior", fleet.OverflowBehavior)

	if fleet.ScalingConfiguration != nil {
		if err := d.Set("scaling_configuration", flattenScalingConfiguration(fleet.ScalingConfiguration)); err != nil {
			return create.AppendDiagError(diags, names.CodeBuild, create.ErrActionSetting, dsNameFleet, d.Id(), err)
		}
	}

	if fleet.Status != nil {
		if err := d.Set(names.AttrStatus, []any{flattenStatus(fleet.Status)}); err != nil {
			return create.AppendDiagError(diags, names.CodeBuild, create.ErrActionSetting, dsNameFleet, d.Id(), err)
		}
	}

	if err := d.Set(names.AttrVPCConfig, flattenVPCConfig(fleet.VpcConfig)); err != nil {
		return create.AppendDiagError(diags, names.CodeBuild, create.ErrActionSetting, dsNameFleet, d.Id(), err)
	}

	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig(ctx)
	if err := d.Set(names.AttrTags, KeyValueTags(ctx, fleet.Tags).IgnoreAWS().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return create.AppendDiagError(diags, names.CodeBuild, create.ErrActionSetting, dsNameFleet, d.Id(), err)
	}

	return diags
}
