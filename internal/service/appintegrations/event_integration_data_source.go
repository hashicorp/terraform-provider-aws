// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package appintegrations

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/appintegrationsservice"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
)

// @SDKDataSource("aws_appintegrations_event_integration", name="Event Integration")
func DataSourceEventIntegration() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceEventIntegrationRead,

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"description": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"event_filter": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"source": {
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
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"tags": tftags.TagsSchemaComputed(),
		},
	}
}

func dataSourceEventIntegrationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).AppIntegrationsConn(ctx)
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	name := d.Get("name").(string)
	output, err := conn.GetEventIntegrationWithContext(ctx, &appintegrationsservice.GetEventIntegrationInput{
		Name: aws.String(name),
	})

	if err != nil {
		return diag.Errorf("reading AppIntegrations Event Integration (%s): %s", name, err)
	}

	d.SetId(aws.StringValue(output.Name))
	d.Set("arn", output.EventIntegrationArn)
	d.Set("description", output.Description)
	if err := d.Set("event_filter", flattenEventFilter(output.EventFilter)); err != nil {
		return diag.Errorf("setting event_filter: %s", err)
	}
	d.Set("eventbridge_bus", output.EventBridgeBus)
	d.Set("name", output.Name)

	if err := d.Set("tags", KeyValueTags(ctx, output.Tags).IgnoreAWS().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return diag.Errorf("setting tags: %s", err)
	}

	return nil
}
