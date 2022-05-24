package appintegrations

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/appintegrationsservice"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
)

func DataSourceEventIntegration() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceEventIntegrationRead,
		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"description": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"eventbridge_bus": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
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
			"tags": tftags.TagsSchemaComputed(),
		},
	}
}

func dataSourceEventIntegrationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).AppIntegrationsConn
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	name := d.Get("name").(string)

	resp, err := conn.GetEventIntegrationWithContext(ctx, &appintegrationsservice.GetEventIntegrationInput{
		Name: aws.String(name),
	})
	if err != nil {
		return diag.FromErr(fmt.Errorf("error getting AppIntegrations Event Integration: %w", err))
	}

	d.SetId(aws.StringValue(resp.Name))

	d.Set("arn", resp.EventIntegrationArn)
	d.Set("description", resp.Description)
	d.Set("eventbridge_bus", resp.EventBridgeBus)
	d.Set("name", resp.Name)

	if err := d.Set("event_filter", flattenEventFilter(resp.EventFilter)); err != nil {
		return diag.FromErr(fmt.Errorf("error setting event_filter: %w", err))
	}

	if err := d.Set("tags", KeyValueTags(resp.Tags).IgnoreAWS().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return diag.FromErr(fmt.Errorf("error setting tags: %s", err))
	}

	return nil
}
