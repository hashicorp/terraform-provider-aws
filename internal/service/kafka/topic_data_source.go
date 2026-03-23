// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

package kafka

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkDataSource("aws_msk_topic", name="Topic")
func newTopicDataSource(_ context.Context) (datasource.DataSourceWithConfigure, error) {
	return &topicDataSource{}, nil
}

type topicDataSource struct {
	framework.DataSourceWithModel[topicDataSourceModel]
}

func (d *topicDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrARN: framework.ARNAttributeComputedOnly(),
			"cluster_arn": schema.StringAttribute{
				CustomType: fwtypes.ARNType,
				Required:   true,
			},
			"configs": schema.StringAttribute{
				Computed: true,
			},
			names.AttrName: schema.StringAttribute{
				Required: true,
			},
			"partition_count": schema.Int64Attribute{
				Computed: true,
			},
			"replication_factor": schema.Int64Attribute{
				Computed: true,
			},
		},
	}
}

func (d *topicDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	conn := d.Meta().KafkaClient(ctx)

	var config topicDataSourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Config.Get(ctx, &config))
	if resp.Diagnostics.HasError() {
		return
	}

	clusterARN, topicName := fwflex.StringValueFromFramework(ctx, config.ClusterARN), fwflex.StringValueFromFramework(ctx, config.TopicName)
	out, err := findTopicByTwoPartKey(ctx, conn, clusterARN, topicName)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, clusterARN, topicName)
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, fwflex.Flatten(ctx, out, &config))
	if resp.Diagnostics.HasError() {
		return
	}

	// Data source's Configs contains server-augmented values.
	v, diags := flattenTopicConfigsActual(ctx, out.Configs)
	smerr.AddEnrich(ctx, &resp.Diagnostics, diags)
	if resp.Diagnostics.HasError() {
		return
	}
	config.Configs = v

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &config))
}

type topicDataSourceModel struct {
	framework.WithRegionModel
	ClusterARN        fwtypes.ARN  `tfsdk:"cluster_arn"`
	Configs           types.String `tfsdk:"configs" autoflex:"-"`
	PartitionCount    types.Int64  `tfsdk:"partition_count"`
	ReplicationFactor types.Int64  `tfsdk:"replication_factor"`
	TopicARN          types.String `tfsdk:"arn"`
	TopicName         types.String `tfsdk:"name"`
}
