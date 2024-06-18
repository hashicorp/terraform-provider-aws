// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package appintegrations

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/appintegrations"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_appintegrations_event_integration", name="Event Integration")
// @Tags
func DataSourceEventIntegration() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceEventIntegrationRead,

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrDescription: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"event_filter": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrSource: {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"eventbridge_bus": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrName: {
				Type:     schema.TypeString,
				Required: true,
			},
			names.AttrTags: tftags.TagsSchemaComputed(),
		},
	}
}

func dataSourceEventIntegrationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).AppIntegrationsClient(ctx)

	name := d.Get(names.AttrName).(string)
	output, err := conn.GetEventIntegration(ctx, &appintegrations.GetEventIntegrationInput{
		Name: aws.String(name),
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading AppIntegrations Event Integration (%s): %s", name, err)
	}

	d.SetId(aws.ToString(output.Name))
	d.Set(names.AttrARN, output.EventIntegrationArn)
	d.Set(names.AttrDescription, output.Description)
	if err := d.Set("event_filter", flattenEventFilter(output.EventFilter)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting event_filter: %s", err)
	}
	d.Set("eventbridge_bus", output.EventBridgeBus)
	d.Set(names.AttrName, output.Name)

	setTagsOut(ctx, output.Tags)

	return diags
}
