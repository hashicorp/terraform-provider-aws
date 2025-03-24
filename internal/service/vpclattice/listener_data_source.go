// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package vpclattice

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_vpclattice_listener", name="Listener")
// @Tags(identifierAttribute="arn")
// @Testing(tagsTest=false)
func dataSourceListener() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceListenerRead,

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrCreatedAt: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrDefaultAction: {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"fixed_response": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									names.AttrStatusCode: {
										Type:     schema.TypeInt,
										Computed: true,
									},
								},
							},
						},
						"forward": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"target_groups": {
										Type:     schema.TypeList,
										Computed: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"target_group_identifier": {
													Type:     schema.TypeString,
													Computed: true,
												},
												names.AttrWeight: {
													Type:     schema.TypeInt,
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
			"last_updated_at": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"listener_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"listener_identifier": {
				Type:     schema.TypeString,
				Required: true,
			},
			names.AttrName: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrPort: {
				Type:     schema.TypeInt,
				Computed: true,
			},
			names.AttrProtocol: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"service_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"service_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"service_identifier": {
				Type:     schema.TypeString,
				Required: true,
			},
			names.AttrTags: tftags.TagsSchemaComputed(),
		},
	}
}

func dataSourceListenerRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).VPCLatticeClient(ctx)

	serviceID, listenerID := d.Get("service_identifier").(string), d.Get("listener_identifier").(string)

	output, err := findListenerByTwoPartKey(ctx, conn, serviceID, listenerID)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading VPCLattice Listener (%s): %s", listenerCreateResourceID(serviceID, listenerID), err)
	}

	d.SetId(aws.ToString(output.Id))
	d.Set(names.AttrARN, output.Arn)
	d.Set(names.AttrCreatedAt, aws.ToTime(output.CreatedAt).String())
	if err := d.Set(names.AttrDefaultAction, flattenListenerRuleActions(output.DefaultAction)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting default_action: %s", err)
	}
	d.Set("last_updated_at", aws.ToTime(output.LastUpdatedAt).String())
	d.Set("listener_id", output.Id)
	d.Set(names.AttrName, output.Name)
	d.Set(names.AttrPort, output.Port)
	d.Set(names.AttrProtocol, output.Protocol)
	d.Set("service_arn", output.ServiceArn)
	d.Set("service_id", output.ServiceId)

	return diags
}
