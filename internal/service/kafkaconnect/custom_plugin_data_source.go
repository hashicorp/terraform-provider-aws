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

// @SDKDataSource("aws_mskconnect_custom_plugin", name="Custom Plugin")
// @Tags(identifierAttribute="arn")
func dataSourceCustomPlugin() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceCustomPluginRead,

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
			names.AttrState: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrTags: tftags.TagsSchemaComputed(),
		},
	}
}

func dataSourceCustomPluginRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).KafkaConnectClient(ctx)

	plugin, err := findCustomPluginByName(ctx, conn, d.Get(names.AttrName).(string))

	if err != nil {
		return sdkdiag.AppendFromErr(diags, tfresource.SingularDataSourceFindError("MSK Connect Custom Plugin", err))
	}

	arn := aws.ToString(plugin.CustomPluginArn)
	d.SetId(arn)
	d.Set(names.AttrARN, arn)
	d.Set(names.AttrDescription, plugin.Description)
	d.Set(names.AttrName, plugin.Name)
	d.Set(names.AttrState, plugin.CustomPluginState)

	if plugin.LatestRevision != nil {
		d.Set("latest_revision", plugin.LatestRevision.Revision)
	} else {
		d.Set("latest_revision", nil)
	}

	return diags
}

func findCustomPlugin(ctx context.Context, conn *kafkaconnect.Client, input *kafkaconnect.ListCustomPluginsInput, filter tfslices.Predicate[*awstypes.CustomPluginSummary]) (*awstypes.CustomPluginSummary, error) {
	output, err := findCustomPlugins(ctx, conn, input, filter)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findCustomPlugins(ctx context.Context, conn *kafkaconnect.Client, input *kafkaconnect.ListCustomPluginsInput, filter tfslices.Predicate[*awstypes.CustomPluginSummary]) ([]awstypes.CustomPluginSummary, error) {
	var output []awstypes.CustomPluginSummary

	pages := kafkaconnect.NewListCustomPluginsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			return nil, err
		}

		for _, v := range page.CustomPlugins {
			if filter(&v) {
				output = append(output, v)
			}
		}
	}

	return output, nil
}

func findCustomPluginByName(ctx context.Context, conn *kafkaconnect.Client, name string) (*awstypes.CustomPluginSummary, error) {
	input := &kafkaconnect.ListCustomPluginsInput{}

	return findCustomPlugin(ctx, conn, input, func(v *awstypes.CustomPluginSummary) bool {
		return aws.ToString(v.Name) == name
	})
}
