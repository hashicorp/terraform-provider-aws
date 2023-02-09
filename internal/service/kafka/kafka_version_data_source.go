package kafka

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/kafka"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
)

func DataSourceVersion() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceVersionRead,

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

func dataSourceVersionRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).KafkaConn()

	var kafkaVersions []*kafka.KafkaVersion

	err := conn.ListKafkaVersionsPagesWithContext(ctx, &kafka.ListKafkaVersionsInput{}, func(page *kafka.ListKafkaVersionsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		kafkaVersions = append(kafkaVersions, page.KafkaVersions...)

		return !lastPage
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "listing Kafka versions: %s", err)
	}

	if len(kafkaVersions) == 0 {
		return sdkdiag.AppendErrorf(diags, "no Kafka versions found")
	}

	var found *kafka.KafkaVersion

	if v, ok := d.GetOk("preferred_versions"); ok {
		found = findVersion(v.([]interface{}), kafkaVersions)
	} else if v, ok := d.GetOk("version"); ok {
		found = findVersion([]interface{}{v}, kafkaVersions)
	}

	if found == nil {
		return sdkdiag.AppendErrorf(diags, "no Kafka versions match the criteria")
	}

	d.SetId(aws.StringValue(found.Version))

	d.Set("status", found.Status)
	d.Set("version", found.Version)

	return diags
}
