package kafkaconnect

import (
	"context"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/kafkaconnect"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
)

func DataSourceConnector() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceConnectorRead,

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"state": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"version": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceConnectorRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).KafkaConnectConn

	connectorName := d.Get("name")

	input := &kafkaconnect.ListConnectorsInput{}

	var connector *kafkaconnect.ConnectorSummary

	err := conn.ListConnectorsPagesWithContext(ctx, input, func(page *kafkaconnect.ListConnectorsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, connectorSummary := range page.Connectors {
			if aws.StringValue(connectorSummary.ConnectorName) == connectorName {
				connector = connectorSummary

				return false
			}
		}

		return !lastPage
	})

	if err != nil {
		return diag.Errorf("error listing MSK Connect Connector: %s", err)
	}

	if connector == nil {
		return diag.Errorf("error reading MSK Connect Connector (%s): no results found", connectorName)
	}

	d.SetId(aws.StringValue(connector.ConnectorArn))
	_ = d.Set("arn", connector.ConnectorArn)
	_ = d.Set("description", connector.ConnectorDescription)
	_ = d.Set("name", connector.ConnectorName)
	_ = d.Set("state", connector.ConnectorState)
	_ = d.Set("version", connector.CurrentVersion)

	return nil
}
