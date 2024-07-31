// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package codebuild

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_codebuild_fleet", name="Fleet")
func DataSourceFleet() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceFleetRead,

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"base_capacity": {
				Type:     schema.TypeInt,
				Computed: true,
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
			"id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"last_modified": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"name": {
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
						"max_capacity": {
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
			"status": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"context": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"message": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"status_code": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"tags": tftags.TagsSchemaComputed(),
		},
	}
}

const (
	DSNameFleet = "Fleet Data Source"
)

func dataSourceFleetRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).CodeBuildClient(ctx)
	name := d.Get("name").(string)

	out, err := findFleetByARNOrNames(ctx, conn, name)
	if err != nil {
		return create.AppendDiagError(diags, names.CodeBuild, create.ErrActionReading, DSNameFleet, name, err)
	}
	d.SetId(aws.StringValue(out.Fleets[0].Arn))

	d.Set("arn", out.Fleets[0].Arn)
	d.Set("base_capacity", out.Fleets[0].BaseCapacity)
	d.Set("compute_type", out.Fleets[0].ComputeType)
	d.Set("created", out.Fleets[0].Created.String())
	d.Set("environment_type", out.Fleets[0].EnvironmentType)
	d.Set("last_modified", out.Fleets[0].LastModified.String())
	d.Set("overflow_behavior", out.Fleets[0].OverflowBehavior)
	d.Set("name", out.Fleets[0].Name)
	if out.Fleets[0].ScalingConfiguration != nil {
		if err := d.Set("scaling_configuration", []interface{}{flattenScalingConfiguration(out.Fleets[0].ScalingConfiguration)}); err != nil {
			return create.AppendDiagError(diags, names.CodeBuild, create.ErrActionSetting, DSNameFleet, d.Id(), err)
		}
	}
	if out.Fleets[0].Status != nil {
		if err := d.Set("status", []interface{}{flattenStatus(out.Fleets[0].Status)}); err != nil {
			return create.AppendDiagError(diags, names.CodeBuild, create.ErrActionSetting, DSNameFleet, d.Id(), err)
		}
	}

	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig
	//lintignore:AWSR002
	if err := d.Set("tags", KeyValueTags(ctx, out.Fleets[0].Tags).IgnoreAWS().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return create.AppendDiagError(diags, names.CodeBuild, create.ErrActionSetting, DSNameFleet, d.Id(), err)
	}

	return diags
}
