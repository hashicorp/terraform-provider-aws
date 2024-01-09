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
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"golang.org/x/exp/slices"
)

// @SDKDataSource("aws_msk_kafka_version", name="Kafka Version")
func dataSourceKafkaVersion() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceKafkaVersionRead,

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

func dataSourceKafkaVersionRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).KafkaClient(ctx)

	var preferredVersions []string
	if v, ok := d.GetOk("preferred_versions"); ok && len(v.([]interface{})) > 0 {
		preferredVersions = flex.ExpandStringValueList(v.([]interface{}))
	} else if v, ok := d.GetOk("version"); ok {
		preferredVersions = tfslices.Of(v.(string))
	}

	input := &kafka.ListKafkaVersionsInput{}
	kafkaVersion, err := findKafkaVersion(ctx, conn, input, func(v *types.KafkaVersion) bool {
		return slices.Contains(preferredVersions, aws.ToString(v.Version))
	})

	if err != nil {
		return sdkdiag.AppendFromErr(diags, tfresource.SingularDataSourceFindError("MSK Kafka Version", err))
	}

	version := aws.ToString(kafkaVersion.Version)
	d.SetId(version)
	d.Set("status", kafkaVersion.Status)
	d.Set("version", version)

	return diags
}

func findKafkaVersion(ctx context.Context, conn *kafka.Client, input *kafka.ListKafkaVersionsInput, filter tfslices.Predicate[*types.KafkaVersion]) (*types.KafkaVersion, error) {
	output, err := findKafkaVersions(ctx, conn, input, filter)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertFirstValueResult(output)
}

func findKafkaVersions(ctx context.Context, conn *kafka.Client, input *kafka.ListKafkaVersionsInput, filter tfslices.Predicate[*types.KafkaVersion]) ([]types.KafkaVersion, error) {
	var output []types.KafkaVersion

	pages := kafka.NewListKafkaVersionsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			return nil, err
		}

		for _, v := range page.KafkaVersions {
			if filter(&v) {
				output = append(output, v)
			}
		}
	}

	return output, nil
}
