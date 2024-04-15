// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package vpclattice

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/vpclattice"
	"github.com/aws/aws-sdk-go-v2/service/vpclattice/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// Function annotations are used for datasource registration to the Provider. DO NOT EDIT.
// @SDKDataSource("aws_vpclattice_target_group", name="Target Group")
func DataSourceTargetGroup() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceTargetGroupRead,

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"config": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"health_check": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"enabled": {
										Type:     schema.TypeBool,
										Computed: true,
									},
									"health_check_interval_seconds": {
										Type:     schema.TypeInt,
										Computed: true,
									},
									"health_check_timeout_seconds": {
										Type:     schema.TypeInt,
										Computed: true,
									},
									"healthy_threshold_count": {
										Type:     schema.TypeInt,
										Computed: true,
									},
									"matcher": {
										Type:     schema.TypeList,
										Computed: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"value": {
													Type:     schema.TypeString,
													Computed: true,
												},
											},
										},
									},
									"path": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"port": {
										Type:     schema.TypeInt,
										Computed: true,
									},
									"protocol": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"protocol_version": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"unhealthy_threshold_count": {
										Type:     schema.TypeInt,
										Computed: true,
									},
								},
							},
						},
						"ip_address_type": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"lambda_event_structure_version": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"port": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"protocol": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"protocol_version": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"vpc_identifier": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"created_at": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"last_updated_at": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"name": {
				Type:         schema.TypeString,
				Computed:     true,
				Optional:     true,
				ExactlyOneOf: []string{"name", "target_group_identifier"},
			},
			"service_arns": {
				Type: schema.TypeList,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Computed: true,
			},
			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"tags": tftags.TagsSchemaComputed(),
			"target_group_identifier": {
				Type:         schema.TypeString,
				Optional:     true,
				ExactlyOneOf: []string{"name", "target_group_identifier"},
			},
			"type": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

const (
	DSNameTargetGroup = "Target Group Data Source"
)

func dataSourceTargetGroupRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).VPCLatticeClient(ctx)

	var out *vpclattice.GetTargetGroupOutput
	if v, ok := d.GetOk("target_group_identifier"); ok {
		targetGroupID := v.(string)
		targetGroup, err := FindTargetGroupByID(ctx, conn, targetGroupID)

		if err != nil {
			return sdkdiag.AppendFromErr(diags, err)
		}

		out = targetGroup
	} else if v, ok := d.GetOk("name"); ok {
		filter := func(x types.TargetGroupSummary) bool {
			return aws.ToString(x.Name) == v.(string)
		}
		output, err := findTargetGroup(ctx, conn, filter)

		if err != nil {
			return sdkdiag.AppendFromErr(diags, err)
		}

		targetGroup, err := FindTargetGroupByID(ctx, conn, aws.ToString(output.Id))

		if err != nil {
			return sdkdiag.AppendFromErr(diags, err)
		}

		out = targetGroup
	}

	// Set simple arguments
	d.SetId(aws.ToString(out.Id))
	targetGroupARN := aws.ToString(out.Arn)
	d.Set("arn", targetGroupARN)
	d.Set("created_at", aws.ToTime(out.CreatedAt).String())
	d.Set("last_updated_at", aws.ToTime(out.LastUpdatedAt).String())
	d.Set("name", out.Name)
	d.Set("service_arns", out.ServiceArns)
	d.Set("status", out.Status)
	d.Set("target_group_identifier", out.Id)
	d.Set("type", out.Type)

	// Flatten complex config attribute - uses flatteners from target_group.go
	if err := d.Set("config", []interface{}{
		flattenTargetGroupConfig(out.Config),
	}); err != nil {
		return diag.Errorf("setting config: %s", err)
	}

	// Set tags
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig
	tags, err := listTags(ctx, conn, aws.ToString(out.Arn))

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "listing tags for VPC Lattice Target Group (%s): %s", targetGroupARN, err)
	}

	//lintignore:AWSR002
	if err := d.Set("tags", tags.IgnoreAWS().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting tags for VPC Lattice Target Group (%s): %s", targetGroupARN, err)
	}

	return nil
}
