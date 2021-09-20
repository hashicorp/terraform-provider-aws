package aws

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/kafka"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func DataSourceConfiguration() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceConfigurationRead,

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

func dataSourceConfigurationRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).KafkaConn

	listConfigurationsInput := &kafka.ListConfigurationsInput{}

	var configuration *kafka.Configuration
	err := conn.ListConfigurationsPages(listConfigurationsInput, func(page *kafka.ListConfigurationsOutput, lastPage bool) bool {
		for _, config := range page.Configurations {
			if aws.StringValue(config.Name) == d.Get("name").(string) {
				configuration = config
				break
			}
		}

		return !lastPage
	})

	if err != nil {
		return fmt.Errorf("error listing MSK Configurations: %w", err)
	}

	if configuration == nil {
		return fmt.Errorf("error reading MSK Configuration: no results found")
	}

	if configuration.LatestRevision == nil {
		return fmt.Errorf("error describing MSK Configuration (%s): missing latest revision", d.Id())
	}

	revision := configuration.LatestRevision.Revision
	revisionInput := &kafka.DescribeConfigurationRevisionInput{
		Arn:      configuration.Arn,
		Revision: revision,
	}

	revisionOutput, err := conn.DescribeConfigurationRevision(revisionInput)

	if err != nil {
		return fmt.Errorf("error describing MSK Configuration (%s) Revision (%d): %w", d.Id(), aws.Int64Value(revision), err)
	}

	if revisionOutput == nil {
		return fmt.Errorf("error describing MSK Configuration (%s) Revision (%d): missing result", d.Id(), aws.Int64Value(revision))
	}

	d.Set("arn", configuration.Arn)
	d.Set("description", configuration.Description)

	if err := d.Set("kafka_versions", aws.StringValueSlice(configuration.KafkaVersions)); err != nil {
		return fmt.Errorf("error setting kafka_versions: %w", err)
	}

	d.Set("latest_revision", revision)
	d.Set("name", configuration.Name)
	d.Set("server_properties", string(revisionOutput.ServerProperties))

	d.SetId(aws.StringValue(configuration.Arn))

	return nil
}
