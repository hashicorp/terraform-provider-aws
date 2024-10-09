// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ssmcontacts

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_ssmcontacts_plan")
func DataSourcePlan() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourcePlanRead,

		Schema: map[string]*schema.Schema{
			"contact_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			names.AttrStage: {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"duration_in_minutes": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						names.AttrTarget: {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"channel_target_info": {
										Type:     schema.TypeList,
										Computed: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"contact_channel_id": {
													Type:     schema.TypeString,
													Computed: true,
												},
												"retry_interval_in_minutes": {
													Type:     schema.TypeInt,
													Computed: true,
												},
											},
										},
									},
									"contact_target_info": {
										Type:     schema.TypeList,
										Computed: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"is_essential": {
													Type:     schema.TypeBool,
													Computed: true,
												},
												"contact_id": {
													Type:     schema.TypeString,
													Computed: true,
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

const (
	DSNamePlan = "Plan Data Source"
)

func dataSourcePlanRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SSMContactsClient(ctx)

	contactId := d.Get("contact_id").(string)

	out, err := findContactByID(ctx, conn, contactId)
	if err != nil {
		return create.AppendDiagError(diags, names.SSMContacts, create.ErrActionReading, DSNamePlan, contactId, err)
	}

	d.SetId(aws.ToString(out.ContactArn))

	if err := setPlanResourceData(d, out); err != nil {
		return create.AppendDiagError(diags, names.SSMContacts, create.ErrActionReading, DSNamePlan, contactId, err)
	}

	return diags
}
