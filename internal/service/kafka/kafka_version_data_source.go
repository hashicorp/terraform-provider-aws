package kafka

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/kafka"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
)

func DataSourceVersion() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceVersionRead,

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
				Optional:     true,
				Computed:     true,
				ExactlyOneOf: []string{"version", "preferred_versions"},
			},
		},
	}
}

func findVersion(preferredVersions []interface{}, versions []*kafka.KafkaVersion) *kafka.KafkaVersion {
	var found *kafka.KafkaVersion

	for _, v := range preferredVersions {
		preferredVersion, ok := v.(string)

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

	return found
}

func dataSourceVersionRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).KafkaConn

	var kafkaVersions []*kafka.KafkaVersion

	err := conn.ListKafkaVersionsPages(&kafka.ListKafkaVersionsInput{}, func(page *kafka.ListKafkaVersionsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		kafkaVersions = append(kafkaVersions, page.KafkaVersions...)

		return !lastPage
	})

	if err != nil {
		return fmt.Errorf("error listing Kafka versions: %w", err)
	}

	if len(kafkaVersions) == 0 {
		return fmt.Errorf("no Kafka versions found")
	}

	var found *kafka.KafkaVersion

	if v, ok := d.GetOk("preferred_versions"); ok {
		found = findVersion(v.([]interface{}), kafkaVersions)
	} else if v, ok := d.GetOk("version"); ok {
		found = findVersion([]interface{}{v}, kafkaVersions)
	}

	if found == nil {
		return fmt.Errorf("no Kafka versions match the criteria")
	}

	d.SetId(aws.StringValue(found.Version))

	d.Set("status", found.Status)
	d.Set("version", found.Version)

	return nil
}
