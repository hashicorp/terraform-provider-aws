package kafka

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/kafka"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
)

func DataSourceConfiguration() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceConfigurationRead,

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"description": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"kafka_versions": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"latest_revision": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringLenBetween(1, 64),
			},
			"server_properties": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceConfigurationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).KafkaConn()

	listConfigurationsInput := &kafka.ListConfigurationsInput{}

	var configuration *kafka.Configuration
	err := conn.ListConfigurationsPagesWithContext(ctx, listConfigurationsInput, func(page *kafka.ListConfigurationsOutput, lastPage bool) bool {
		for _, config := range page.Configurations {
			if aws.StringValue(config.Name) == d.Get("name").(string) {
				configuration = config
				break
			}
		}

		return !lastPage
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "listing MSK Configurations: %s", err)
	}

	if configuration == nil {
		return sdkdiag.AppendErrorf(diags, "reading MSK Configuration: no results found")
	}

	if configuration.LatestRevision == nil {
		return sdkdiag.AppendErrorf(diags, "describing MSK Configuration (%s): missing latest revision", d.Id())
	}

	revision := configuration.LatestRevision.Revision
	revisionInput := &kafka.DescribeConfigurationRevisionInput{
		Arn:      configuration.Arn,
		Revision: revision,
	}

	revisionOutput, err := conn.DescribeConfigurationRevisionWithContext(ctx, revisionInput)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "describing MSK Configuration (%s) Revision (%d): %s", d.Id(), aws.Int64Value(revision), err)
	}

	if revisionOutput == nil {
		return sdkdiag.AppendErrorf(diags, "describing MSK Configuration (%s) Revision (%d): missing result", d.Id(), aws.Int64Value(revision))
	}

	d.Set("arn", configuration.Arn)
	d.Set("description", configuration.Description)

	if err := d.Set("kafka_versions", aws.StringValueSlice(configuration.KafkaVersions)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting kafka_versions: %s", err)
	}

	d.Set("latest_revision", revision)
	d.Set("name", configuration.Name)
	d.Set("server_properties", string(revisionOutput.ServerProperties))

	d.SetId(aws.StringValue(configuration.Arn))

	return diags
}
