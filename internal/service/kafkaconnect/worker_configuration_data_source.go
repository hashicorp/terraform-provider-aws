package kafkaconnect

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/kafkaconnect"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
)

func DataSourceWorkerConfiguration() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceWorkerConfigurationRead,

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"description": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"latest_revision": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"properties_file_content": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceWorkerConfigurationRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).KafkaConnectConn

	configName := d.Get("name")

	input := &kafkaconnect.ListWorkerConfigurationsInput{}

	var config *kafkaconnect.WorkerConfigurationSummary

	err := conn.ListWorkerConfigurationsPages(input, func(page *kafkaconnect.ListWorkerConfigurationsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, configSummary := range page.WorkerConfigurations {
			if aws.StringValue(configSummary.Name) == configName {
				config = configSummary

				return false
			}
		}

		return !lastPage
	})

	if err != nil {
		return fmt.Errorf("error listing MSK Connect Worker Configurations: %w", err)
	}

	if config == nil {
		return fmt.Errorf("error reading MSK Connect Worker Configuration (%s): no results found", configName)
	}

	describeInput := &kafkaconnect.DescribeWorkerConfigurationInput{
		WorkerConfigurationArn: config.WorkerConfigurationArn,
	}

	describeOutput, err := conn.DescribeWorkerConfiguration(describeInput)

	if err != nil {
		return fmt.Errorf("error reading MSK Connect Worker Configuration (%s): %w", configName, err)
	}

	d.SetId(aws.StringValue(config.Name))
	d.Set("arn", config.WorkerConfigurationArn)
	d.Set("description", config.Description)
	d.Set("name", config.Name)

	if config.LatestRevision != nil {
		d.Set("latest_revision", config.LatestRevision.Revision)
	} else {
		d.Set("latest_revision", nil)
	}

	if describeOutput.LatestRevision != nil {
		d.Set("properties_file_content", decodePropertiesFileContent(aws.StringValue(describeOutput.LatestRevision.PropertiesFileContent)))
	} else {
		d.Set("properties_file_content", nil)
	}

	return nil
}
