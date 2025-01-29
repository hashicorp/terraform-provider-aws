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

// @SDKDataSource("aws_mskconnect_connector", name="Connector")
// @Tags(identifierAttribute="arn")
func dataSourceConnector() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceConnectorRead,

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrDescription: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrName: {
				Type:     schema.TypeString,
				Required: true,
			},
			names.AttrVersion: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrTags: tftags.TagsSchemaComputed(),
		},
	}
}

func dataSourceConnectorRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).KafkaConnectClient(ctx)

	connector, err := findConnectorByName(ctx, conn, d.Get(names.AttrName).(string))

	if err != nil {
		return sdkdiag.AppendFromErr(diags, tfresource.SingularDataSourceFindError("MSK Connect Connector", err))
	}

	arn := aws.ToString(connector.ConnectorArn)
	d.SetId(arn)
	d.Set(names.AttrARN, arn)
	d.Set(names.AttrDescription, connector.ConnectorDescription)
	d.Set(names.AttrName, connector.ConnectorName)
	d.Set(names.AttrVersion, connector.CurrentVersion)

	return diags
}

func findConnector(ctx context.Context, conn *kafkaconnect.Client, input *kafkaconnect.ListConnectorsInput, filter tfslices.Predicate[*awstypes.ConnectorSummary]) (*awstypes.ConnectorSummary, error) {
	output, err := findConnectors(ctx, conn, input, filter)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findConnectors(ctx context.Context, conn *kafkaconnect.Client, input *kafkaconnect.ListConnectorsInput, filter tfslices.Predicate[*awstypes.ConnectorSummary]) ([]awstypes.ConnectorSummary, error) {
	var output []awstypes.ConnectorSummary

	pages := kafkaconnect.NewListConnectorsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			return nil, err
		}

		for _, v := range page.Connectors {
			if filter(&v) {
				output = append(output, v)
			}
		}
	}

	return output, nil
}

func findConnectorByName(ctx context.Context, conn *kafkaconnect.Client, name string) (*awstypes.ConnectorSummary, error) {
	input := &kafkaconnect.ListConnectorsInput{}

	return findConnector(ctx, conn, input, func(v *awstypes.ConnectorSummary) bool {
		return aws.ToString(v.ConnectorName) == name
	})
}
