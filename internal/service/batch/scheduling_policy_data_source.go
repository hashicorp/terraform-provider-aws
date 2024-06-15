// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package batch

import (
	"bytes"
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_batch_scheduling_policy", name="Scheduling Policy")
// @Tags
func DataSourceSchedulingPolicy() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceSchedulingPolicyRead,
		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Required: true,
			},
			"fair_share_policy": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"compute_reservation": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"share_decay_seconds": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"share_distribution": {
							Type:     schema.TypeSet,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"share_identifier": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"weight_factor": {
										Type:     schema.TypeFloat,
										Computed: true,
									},
								},
							},
							Set: func(v interface{}) int {
								var buf bytes.Buffer
								m := v.(map[string]interface{})
								buf.WriteString(m["share_identifier"].(string))
								if v, ok := m["weight_factor"]; ok {
									buf.WriteString(fmt.Sprintf("%s-", v))
								}
								return create.StringHashcode(buf.String())
							},
						},
					},
				},
			},
			names.AttrName: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrTags: tftags.TagsSchemaComputed(),
		},
	}
}

func dataSourceSchedulingPolicyRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).BatchConn(ctx)

	schedulingPolicy, err := FindSchedulingPolicyByARN(ctx, conn, d.Get(names.AttrARN).(string))

	if err != nil {
		return sdkdiag.AppendFromErr(diags, tfresource.SingularDataSourceFindError("Batch Scheduling Policy", err))
	}

	d.SetId(aws.StringValue(schedulingPolicy.Arn))
	if err := d.Set("fair_share_policy", flattenFairsharePolicy(schedulingPolicy.FairsharePolicy)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting fair_share_policy: %s", err)
	}
	d.Set(names.AttrName, schedulingPolicy.Name)

	setTagsOut(ctx, schedulingPolicy.Tags)

	return diags
}
