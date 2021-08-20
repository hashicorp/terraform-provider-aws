package aws

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/kafka"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceAwsMskKafkaVersion() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAwsMskKafkaVersionRead,

		Schema: map[string]*schema.Schema{
			"preferred_versions": {
				Type:         schema.TypeList,
				Optional:     true,
				Elem:         &schema.Schema{Type: schema.TypeString},
				ExactlyOneOf: []string{"version", "preferred_versions"},
			},
			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"version": {
				Type:         schema.TypeString,
				Computed:     true,
				Optional:     true,
				ExactlyOneOf: []string{"version", "preferred_versions"},
			},
		},
	}
}

func findMskKafkaVersion(prefferedVersions []interface{}, versions []*kafka.KafkaVersion) *kafka.KafkaVersion {
	var found *kafka.KafkaVersion
	if l := prefferedVersions; len(l) > 0 {
		for _, elem := range l {
			preferredVersion, ok := elem.(string)

			if !ok {
				continue
			}

			for _, kafkaVersion := range versions {
				if preferredVersion == aws.StringValue(kafkaVersion.Version) {
					found = kafkaVersion
					break
				}
			}

			if found != nil {
				break
			}
		}
	}
	return found
}

func dataSourceAwsMskKafkaVersionRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).kafkaconn

	listKafkaVersionsInput := &kafka.ListKafkaVersionsInput{}

	var kafkaVersions []*kafka.KafkaVersion
	for {
		listKafkaVersionsOutput, err := conn.ListKafkaVersions(listKafkaVersionsInput)

		if err != nil {
			return fmt.Errorf("error listing MSK Kafka versions: %w", err)
		}

		if listKafkaVersionsOutput == nil {
			break
		}

		kafkaVersions = append(kafkaVersions, listKafkaVersionsOutput.KafkaVersions...)

		if aws.StringValue(listKafkaVersionsOutput.NextToken) == "" {
			break
		}

		listKafkaVersionsInput.NextToken = listKafkaVersionsOutput.NextToken
	}

	if len(kafkaVersions) == 0 {
		return fmt.Errorf("error reading MSK Kafka versions: no results found")
	}

	var found *kafka.KafkaVersion
	if v, ok := d.GetOk("version"); ok {
		found = findMskKafkaVersion([]interface{}{v}, kafkaVersions)
	}

	if pv, ok := d.GetOk("preferred_versions"); ok {
		found = findMskKafkaVersion(pv.([]interface{}), kafkaVersions)
	}

	if found == nil && len(kafkaVersions) > 1 {
		return fmt.Errorf("multiple MSK Kafka versions (%v) match the criteria", kafkaVersions)
	}

	if found == nil {
		return fmt.Errorf("no MSK  Kafka versions match the criteria")
	}

	d.SetId(aws.StringValue(found.Version))
	d.Set("status", found.Status)
	d.Set("version", found.Version)

	return nil
}
