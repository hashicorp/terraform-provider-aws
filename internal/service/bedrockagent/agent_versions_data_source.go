// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package bedrockagent

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/bedrockagent"
	awstypes "github.com/aws/aws-sdk-go-v2/service/bedrockagent/types"
	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkDataSource("aws_bedrockagent_agent_versions, name="Agent Versions")
func newDataSourceAgentVersions(context.Context) (datasource.DataSourceWithConfigure, error) {
	return &dataSourceAgentVersions{}, nil
}

const (
	DSNameAgentVersions = "Agent Versions Data Source"
)

type dataSourceAgentVersions struct {
	framework.DataSourceWithConfigure
}

func (d *dataSourceAgentVersions) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"agent_id": schema.StringAttribute{
				Required: true,
			},
		},
		Blocks: map[string]schema.Block{
			"agent_version_summaries": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[dsAgentVersionSummaries](ctx),
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"agent_name": schema.StringAttribute{
							Computed: true,
						},
						"agent_status": schema.StringAttribute{
							Computed: true,
						},
						"agent_version": schema.StringAttribute{
							Computed: true,
						},
						names.AttrCreatedAt: schema.StringAttribute{
							Computed: true,
						},
						"updated_at": schema.StringAttribute{
							Computed: true,
						},
						names.AttrDescription: schema.StringAttribute{
							Computed: true,
						},
					},
					Blocks: map[string]schema.Block{
						"guardrail_configuration": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[guardrailConfigurationData](ctx),
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"guardrail_identifier": schema.StringAttribute{
										Computed: true,
									},
									"guardrail_version": schema.StringAttribute{
										Computed: true,
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

func (d *dataSourceAgentVersions) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	conn := d.Meta().BedrockAgentClient(ctx)

	var data dataSourceAgentVersionsData
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	paginator := bedrockagent.NewListAgentVersionsPaginator(conn, &bedrockagent.ListAgentVersionsInput{
		AgentId: data.AgentID.ValueStringPointer(),
	})

	var out bedrockagent.ListAgentVersionsOutput
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.BedrockAgent, create.ErrActionReading, DSNameAgentVersions, data.AgentID.String(), err),
				err.Error(),
			)
			return
		}

		if page != nil && len(page.AgentVersionSummaries) > 0 {
			out.AgentVersionSummaries = append(out.AgentVersionSummaries, page.AgentVersionSummaries...)
		}
	}

	resp.Diagnostics.Append(flex.Flatten(ctx, out, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

type dataSourceAgentVersionsData struct {
	AgentID               types.String                                             `tfsdk:"agent_id"`
	AgentVersionSummaries fwtypes.ListNestedObjectValueOf[dsAgentVersionSummaries] `tfsdk:"agent_version_summaries"`
}

type dsAgentVersionSummaries struct {
	AgentName              types.String                                                `tfsdk:"agent_name"`
	AgentStatus            fwtypes.StringEnum[awstypes.AgentStatus]                    `tfsdk:"agent_status"`
	AgentVersion           types.String                                                `tfsdk:"agent_version"`
	CreatedAt              timetypes.RFC3339                                           `tfsdk:"created_at"`
	UpdatedAt              timetypes.RFC3339                                           `tfsdk:"updated_at"`
	Description            types.String                                                `tfsdk:"description"`
	GuardrailConfiguration fwtypes.ListNestedObjectValueOf[guardrailConfigurationData] `tfsdk:"guardrail_configuration"`
}

type guardrailConfigurationData struct {
	GuardrailIdentifier types.String `tfsdk:"guardrail_identifier"`
	GuardrailVersion    types.String `tfsdk:"guardrail_version"`
}
