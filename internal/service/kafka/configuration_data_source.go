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
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_msk_configuration", name="Configuration")
func dataSourceConfiguration() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceConfigurationRead,

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrDescription: {
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
			names.AttrName: {
				Type:     schema.TypeString,
				Required: true,
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
	conn := meta.(*conns.AWSClient).KafkaClient(ctx)

	input := &kafka.ListConfigurationsInput{}
	configuration, err := findConfiguration(ctx, conn, input, func(v *types.Configuration) bool {
		return aws.ToString(v.Name) == d.Get(names.AttrName).(string)
	})

	if err != nil {
		return sdkdiag.AppendFromErr(diags, tfresource.SingularDataSourceFindError("MSK Configuration", err))
	}

	configurationARN := aws.ToString(configuration.Arn)
	revision := aws.ToInt64(configuration.LatestRevision.Revision)

	revisionOutput, err := findConfigurationRevisionByTwoPartKey(ctx, conn, configurationARN, revision)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading MSK Configuration (%s) revision (%d): %s", d.Id(), revision, err)
	}

	d.SetId(configurationARN)
	d.Set(names.AttrARN, configurationARN)
	d.Set(names.AttrDescription, configuration.Description)
	d.Set("kafka_versions", configuration.KafkaVersions)
	d.Set("latest_revision", revision)
	d.Set(names.AttrName, configuration.Name)
	d.Set("server_properties", string(revisionOutput.ServerProperties))

	return diags
}

func findConfiguration(ctx context.Context, conn *kafka.Client, input *kafka.ListConfigurationsInput, filter tfslices.Predicate[*types.Configuration]) (*types.Configuration, error) {
	output, err := findConfigurations(ctx, conn, input, filter)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findConfigurations(ctx context.Context, conn *kafka.Client, input *kafka.ListConfigurationsInput, filter tfslices.Predicate[*types.Configuration]) ([]types.Configuration, error) {
	var output []types.Configuration

	pages := kafka.NewListConfigurationsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			return nil, err
		}

		for _, v := range page.Configurations {
			if v.LatestRevision == nil {
				continue
			}

			if filter(&v) {
				output = append(output, v)
			}
		}
	}

	return output, nil
}
