// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package continuum

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/continuum"
	awstypes "github.com/aws/aws-sdk-go-v2/service/continuum/types"
	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkDataSource("aws_continuum_collector", name="Collector")
func newCollectorDataSource(context.Context) (datasource.DataSourceWithConfigure, error) {
	return &collectorDataSource{}, nil
}

type collectorDataSource struct {
	framework.DataSourceWithModel[collectorDataSourceModel]
}

func (d *collectorDataSource) Schema(ctx context.Context, _ datasource.SchemaRequest, response *datasource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrARN:         framework.ARNAttributeComputedOnly(),
			"config_recorder_arn": schema.StringAttribute{Computed: true},
			names.AttrCreatedAt:   schema.StringAttribute{CustomType: timetypes.RFC3339Type{}, Computed: true},
			names.AttrID: schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Collector ID. Omit to look up the account's only collector.",
			},
			"kubernetes_cluster_role_arn": schema.StringAttribute{Computed: true},
			names.AttrStatus:              schema.StringAttribute{Computed: true},
			names.AttrType:                schema.StringAttribute{CustomType: fwtypes.StringEnumType[awstypes.CollectorType](), Computed: true},
			"updated_at":                  schema.StringAttribute{CustomType: timetypes.RFC3339Type{}, Computed: true},
		},
	}
}

func (d *collectorDataSource) Read(ctx context.Context, request datasource.ReadRequest, response *datasource.ReadResponse) {
	var data collectorDataSourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.Config.Get(ctx, &data))
	if response.Diagnostics.HasError() {
		return
	}

	conn := d.Meta().ContinuumClient(ctx)

	id := data.ID.ValueString()
	if id == "" {
		list, err := conn.ListCollectors(ctx, &continuum.ListCollectorsInput{})
		if err != nil {
			smerr.AddError(ctx, &response.Diagnostics, err)
			return
		}
		switch len(list.Items) {
		case 0:
			smerr.AddError(ctx, &response.Diagnostics, &retry.NotFoundError{Message: "no Continuum collector in this account and Region"})
			return
		case 1:
			id = aws.ToString(list.Items[0].CollectorId)
		default:
			smerr.AddError(ctx, &response.Diagnostics, fmt.Errorf("found %d Continuum collectors, set id", len(list.Items)))
			return
		}
	}

	output, err := findCollectorByID(ctx, conn, id)
	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, id)
		return
	}

	data.ARN = types.StringPointerValue(output.CollectorArn)
	data.ConfigRecorderARN = types.StringPointerValue(output.ConfigRecorderArn)
	data.CreatedAt = timetypes.NewRFC3339TimePointerValue(output.CreatedAt)
	data.ID = types.StringPointerValue(output.CollectorId)
	data.KubernetesClusterRoleARN = types.StringPointerValue(output.KubernetesClusterRoleArn)
	data.Status = types.StringValue(string(output.Status))
	data.Type = fwtypes.StringEnumValue(output.Type)
	data.UpdatedAt = timetypes.NewRFC3339TimePointerValue(output.UpdatedAt)

	smerr.AddEnrich(ctx, &response.Diagnostics, response.State.Set(ctx, &data))
}

type collectorDataSourceModel struct {
	framework.WithRegionModel
	ARN                      types.String                               `tfsdk:"arn"`
	ConfigRecorderARN        types.String                               `tfsdk:"config_recorder_arn"`
	CreatedAt                timetypes.RFC3339                          `tfsdk:"created_at"`
	ID                       types.String                               `tfsdk:"id"`
	KubernetesClusterRoleARN types.String                               `tfsdk:"kubernetes_cluster_role_arn"`
	Status                   types.String                               `tfsdk:"status"`
	Type                     fwtypes.StringEnum[awstypes.CollectorType] `tfsdk:"type"`
	UpdatedAt                timetypes.RFC3339                          `tfsdk:"updated_at"`
}
