// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package kafkaconnect

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/kafkaconnect"
	awstypes "github.com/aws/aws-sdk-go-v2/service/kafkaconnect/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_mskconnect_worker_configuration", name="Worker Configuration")
// @Tags(identifierAttribute="arn")
func dataSourceWorkerConfiguration() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceWorkerConfigurationRead,

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrDescription: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"latest_revision": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			names.AttrName: {
				Type:     schema.TypeString,
				Required: true,
			},
			"properties_file_content": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrTags: tftags.TagsSchemaComputed(),
		},
	}
}

func dataSourceWorkerConfigurationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).KafkaConnectClient(ctx)

	output, err := findWorkerConfigurationByName(ctx, conn, d.Get(names.AttrName).(string))

	if err != nil {
		return sdkdiag.AppendFromErr(diags, tfresource.SingularDataSourceFindError("MSK Connect Worker Configuration", err))
	}

	arn := aws.ToString(output.WorkerConfigurationArn)
	config, err := findWorkerConfigurationByARN(ctx, conn, arn)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading MSK Connect Worker Configuration (%s): %s", arn, err)
	}

	name := aws.ToString(config.Name)
	d.SetId(name)
	d.Set(names.AttrARN, config.WorkerConfigurationArn)
	d.Set(names.AttrDescription, config.Description)
	d.Set(names.AttrName, name)

	if config.LatestRevision != nil {
		d.Set("latest_revision", config.LatestRevision.Revision)
		d.Set("properties_file_content", decodePropertiesFileContent(aws.ToString(config.LatestRevision.PropertiesFileContent)))
	} else {
		d.Set("latest_revision", nil)
		d.Set("properties_file_content", nil)
	}

	return diags
}

func findWorkerConfiguration(ctx context.Context, conn *kafkaconnect.Client, input *kafkaconnect.ListWorkerConfigurationsInput, filter tfslices.Predicate[*awstypes.WorkerConfigurationSummary]) (*awstypes.WorkerConfigurationSummary, error) {
	output, err := findWorkerConfigurations(ctx, conn, input, filter)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findWorkerConfigurations(ctx context.Context, conn *kafkaconnect.Client, input *kafkaconnect.ListWorkerConfigurationsInput, filter tfslices.Predicate[*awstypes.WorkerConfigurationSummary]) ([]awstypes.WorkerConfigurationSummary, error) {
	var output []awstypes.WorkerConfigurationSummary

	pages := kafkaconnect.NewListWorkerConfigurationsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			return nil, err
		}

		for _, v := range page.WorkerConfigurations {
			if filter(&v) {
				output = append(output, v)
			}
		}
	}

	return output, nil
}

func findWorkerConfigurationByName(ctx context.Context, conn *kafkaconnect.Client, name string) (*awstypes.WorkerConfigurationSummary, error) {
	input := &kafkaconnect.ListWorkerConfigurationsInput{}

	return findWorkerConfiguration(ctx, conn, input, func(v *awstypes.WorkerConfigurationSummary) bool {
		return aws.ToString(v.Name) == name
	})
}
