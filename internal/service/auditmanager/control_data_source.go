// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package auditmanager

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/auditmanager"
	awstypes "github.com/aws/aws-sdk-go-v2/service/auditmanager/types"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkDataSource
func newDataSourceControl(context.Context) (datasource.DataSourceWithConfigure, error) {
	return &dataSourceControl{}, nil
}

type dataSourceControl struct {
	framework.DataSourceWithConfigure
}

func (d *dataSourceControl) Metadata(_ context.Context, request datasource.MetadataRequest, response *datasource.MetadataResponse) { // nosemgrep:ci.meta-in-func-name
	response.TypeName = "aws_auditmanager_control"
}

func (d *dataSourceControl) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"action_plan_instructions": schema.StringAttribute{
				Computed: true,
			},
			"action_plan_title": schema.StringAttribute{
				Computed: true,
			},
			names.AttrARN: framework.ARNAttributeComputedOnly(),
			names.AttrDescription: schema.StringAttribute{
				Computed: true,
			},
			names.AttrID: framework.IDAttribute(),
			names.AttrName: schema.StringAttribute{
				Required: true,
			},
			names.AttrTags: tftags.TagsAttributeComputedOnly(),
			"testing_information": schema.StringAttribute{
				Computed: true,
			},
			names.AttrType: schema.StringAttribute{
				Required: true,
				Validators: []validator.String{
					enum.FrameworkValidate[awstypes.ControlType](),
				},
			},
		},
		Blocks: map[string]schema.Block{
			"control_mapping_sources": schema.SetNestedBlock{
				Validators: []validator.Set{
					setvalidator.SizeAtLeast(1),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"source_description": schema.StringAttribute{
							Computed: true,
						},
						"source_frequency": schema.StringAttribute{
							Computed: true,
						},
						"source_id": framework.IDAttribute(),
						"source_name": schema.StringAttribute{
							Computed: true,
						},
						"source_set_up_option": schema.StringAttribute{
							Computed: true,
						},
						names.AttrSourceType: schema.StringAttribute{
							Computed: true,
						},
						"troubleshooting_text": schema.StringAttribute{
							Computed: true,
						},
					},
					Blocks: map[string]schema.Block{
						"source_keyword": schema.ListNestedBlock{
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"keyword_input_type": schema.StringAttribute{
										Computed: true,
									},
									"keyword_value": schema.StringAttribute{
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

func (d *dataSourceControl) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	conn := d.Meta().AuditManagerClient(ctx)

	var data dataSourceControlData
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	controlMetadata, err := FindControlByName(ctx, conn, data.Name.ValueString(), data.Type.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("finding control by name", err.Error())
		return
	}

	// Control metadata from the ListControls API does not contain all information available
	// about a control. Use control ID to get complete information.
	control, err := FindControlByID(ctx, conn, aws.ToString(controlMetadata.Id))
	if err != nil {
		resp.Diagnostics.AddError("finding control by ID", err.Error())
		return
	}

	resp.Diagnostics.Append(data.refreshFromOutput(ctx, d.Meta(), control)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func FindControlByName(ctx context.Context, conn *auditmanager.Client, name, controlType string) (*awstypes.ControlMetadata, error) {
	in := &auditmanager.ListControlsInput{
		ControlType: awstypes.ControlType(controlType),
	}
	pages := auditmanager.NewListControlsPaginator(conn, in)

	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)
		if err != nil {
			return nil, err
		}

		for _, control := range page.ControlMetadataList {
			if name == aws.ToString(control.Name) {
				return &control, nil
			}
		}
	}

	return nil, &retry.NotFoundError{
		LastRequest: in,
	}
}

type dataSourceControlData struct {
	ActionPlanInstructions types.String `tfsdk:"action_plan_instructions"`
	ActionPlanTitle        types.String `tfsdk:"action_plan_title"`
	ARN                    types.String `tfsdk:"arn"`
	ControlMappingSources  types.Set    `tfsdk:"control_mapping_sources"`
	Description            types.String `tfsdk:"description"`
	ID                     types.String `tfsdk:"id"`
	Name                   types.String `tfsdk:"name"`
	Tags                   types.Map    `tfsdk:"tags"`
	TestingInformation     types.String `tfsdk:"testing_information"`
	Type                   types.String `tfsdk:"type"`
}

// refreshFromOutput writes state data from an AWS response object
func (rd *dataSourceControlData) refreshFromOutput(ctx context.Context, meta *conns.AWSClient, out *awstypes.Control) diag.Diagnostics {
	var diags diag.Diagnostics

	if out == nil {
		return diags
	}

	rd.ID = types.StringValue(aws.ToString(out.Id))
	rd.Name = types.StringValue(aws.ToString(out.Name))
	cms, d := flattenControlMappingSources(ctx, out.ControlMappingSources)
	diags.Append(d...)
	rd.ControlMappingSources = cms

	rd.ActionPlanInstructions = flex.StringToFramework(ctx, out.ActionPlanInstructions)
	rd.ActionPlanTitle = flex.StringToFramework(ctx, out.ActionPlanTitle)
	rd.Description = flex.StringToFramework(ctx, out.Description)
	rd.TestingInformation = flex.StringToFramework(ctx, out.TestingInformation)
	rd.ARN = flex.StringToFramework(ctx, out.Arn)
	rd.Type = types.StringValue(string(out.Type))

	ignoreTagsConfig := meta.IgnoreTagsConfig
	tags := KeyValueTags(ctx, out.Tags).IgnoreAWS().IgnoreConfig(ignoreTagsConfig)
	rd.Tags = flex.FlattenFrameworkStringValueMapLegacy(ctx, tags.Map())

	return diags
}
