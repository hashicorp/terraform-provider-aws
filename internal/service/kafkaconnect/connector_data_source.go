package kafkaconnect

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/kafkaconnect"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func DataSourceConnector() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceConnectorRead,

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"description": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"version": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceConnectorRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).KafkaConnectConn()

	name := d.Get("name")
	var output []*kafkaconnect.ConnectorSummary

	err := conn.ListConnectorsPagesWithContext(ctx, &kafkaconnect.ListConnectorsInput{}, func(page *kafkaconnect.ListConnectorsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.Connectors {
			if aws.StringValue(v.ConnectorName) == name {
				output = append(output, v)
			}
		}

		return !lastPage
	})

	if err != nil {
		return diag.Errorf("error listing MSK Connect Connectors: %s", err)
	}

	if len(output) == 0 || output[0] == nil {
		err = tfresource.NewEmptyResultError(name)
	} else if count := len(output); count > 1 {
		err = tfresource.NewTooManyResultsError(count, name)
	}

	if err != nil {
		return diag.FromErr(tfresource.SingularDataSourceFindError("MSK Connect Connector", err))
	}

	connector := output[0]

	d.SetId(aws.StringValue(connector.ConnectorArn))

	d.Set("arn", connector.ConnectorArn)
	d.Set("description", connector.ConnectorDescription)
	d.Set("name", connector.ConnectorName)
	d.Set("version", connector.CurrentVersion)

	return nil
}
