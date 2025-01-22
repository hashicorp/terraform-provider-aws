// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package kafka

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/kafka"
	"github.com/aws/aws-sdk-go-v2/service/kafka/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_msk_kafka_version", name="Kafka Version")
func dataSourceKafkaVersion() *schema.Resource { // nosemgrep:ci.kafka-in-func-name
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceKafkaVersionRead,

		Schema: map[string]*schema.Schema{
			"preferred_versions": {
				Type:         schema.TypeList,
				Optional:     true,
				Elem:         &schema.Schema{Type: schema.TypeString},
				ExactlyOneOf: []string{names.AttrVersion, "preferred_versions"},
			},
			names.AttrStatus: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrVersion: {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ExactlyOneOf: []string{names.AttrVersion, "preferred_versions"},
			},
		},
	}
}

func dataSourceKafkaVersionRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics { // nosemgrep:ci.kafka-in-func-name
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).KafkaClient(ctx)

	var preferredVersions []string
	if v, ok := d.GetOk("preferred_versions"); ok && len(v.([]interface{})) > 0 {
		preferredVersions = flex.ExpandStringValueList(v.([]interface{}))
	} else if v, ok := d.GetOk(names.AttrVersion); ok {
		preferredVersions = []string{v.(string)}
	}

	kafkaVersion, err := findKafkaVersion(ctx, conn, preferredVersions)

	if err != nil {
		return sdkdiag.AppendFromErr(diags, tfresource.SingularDataSourceFindError("MSK Kafka Version", err))
	}

	version := aws.ToString(kafkaVersion.Version)
	d.SetId(version)
	d.Set(names.AttrStatus, kafkaVersion.Status)
	d.Set(names.AttrVersion, version)

	return diags
}

func findKafkaVersion(ctx context.Context, conn *kafka.Client, preferredVersions []string) (*types.KafkaVersion, error) { // nosemgrep:ci.kafka-in-func-name
	input := &kafka.ListKafkaVersionsInput{}
	output, err := findKafkaVersions(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	var kafkaVersions []types.KafkaVersion
	for _, preferredVersion := range preferredVersions {
		for _, kafkaVersion := range output {
			if preferredVersion == aws.ToString(kafkaVersion.Version) {
				kafkaVersions = append(kafkaVersions, kafkaVersion)
			}
		}
	}

	return tfresource.AssertFirstValueResult(kafkaVersions)
}

func findKafkaVersions(ctx context.Context, conn *kafka.Client, input *kafka.ListKafkaVersionsInput) ([]types.KafkaVersion, error) { // nosemgrep:ci.kafka-in-func-name
	var output []types.KafkaVersion

	pages := kafka.NewListKafkaVersionsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			return nil, err
		}

		output = append(output, page.KafkaVersions...)
	}

	return output, nil
}
