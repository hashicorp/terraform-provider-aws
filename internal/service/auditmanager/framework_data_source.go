// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package auditmanager

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/auditmanager"
	awstypes "github.com/aws/aws-sdk-go-v2/service/auditmanager/types"
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
func newDataSourceFramework(context.Context) (datasource.DataSourceWithConfigure, error) {
	return &dataSourceFramework{}, nil
}

type dataSourceFramework struct {
	framework.DataSourceWithConfigure
}

func (d *dataSourceFramework) Metadata(_ context.Context, request datasource.MetadataRequest, response *datasource.MetadataResponse) { // nosemgrep:ci.meta-in-func-name
	response.TypeName = "aws_auditmanager_framework"
}

func (d *dataSourceFramework) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrARN: framework.ARNAttributeComputedOnly(),
			"compliance_type": schema.StringAttribute{
				Computed: true,
			},
			names.AttrDescription: schema.StringAttribute{
				Computed: true,
			},
			"framework_type": schema.StringAttribute{
				Required: true,
				Validators: []validator.String{
					enum.FrameworkValidate[awstypes.FrameworkType](),
				},
			},
			names.AttrID: framework.IDAttribute(),
			names.AttrName: schema.StringAttribute{
				Required: true,
			},
			names.AttrTags: tftags.TagsAttributeComputedOnly(),
		},
		Blocks: map[string]schema.Block{
			"control_sets": schema.SetNestedBlock{
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						names.AttrID: framework.IDAttribute(),
						names.AttrName: schema.StringAttribute{
							Computed: true,
						},
					},
					Blocks: map[string]schema.Block{
						"controls": schema.SetNestedBlock{
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									names.AttrID: schema.StringAttribute{
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

func (d *dataSourceFramework) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	conn := d.Meta().AuditManagerClient(ctx)

	var data dataSourceFrameworkData
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	frameworkMetadata, err := FindFrameworkByName(ctx, conn, data.Name.ValueString(), data.FrameworkType.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("finding framework by name", err.Error())
		return
	}

	// Framework metadata from the ListFrameworks API does not contain all information available
	// about a framework. Use framework ID to get complete information.
	framework, err := FindFrameworkByID(ctx, conn, aws.ToString(frameworkMetadata.Id))
	if err != nil {
		resp.Diagnostics.AddError("finding framework by ID", err.Error())
		return
	}

	resp.Diagnostics.Append(data.refreshFromOutput(ctx, d.Meta(), framework)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func FindFrameworkByName(ctx context.Context, conn *auditmanager.Client, name, frameworkType string) (*awstypes.AssessmentFrameworkMetadata, error) {
	in := &auditmanager.ListAssessmentFrameworksInput{
		FrameworkType: awstypes.FrameworkType(frameworkType),
	}
	pages := auditmanager.NewListAssessmentFrameworksPaginator(conn, in)

	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)
		if err != nil {
			return nil, err
		}

		for _, framework := range page.FrameworkMetadataList {
			if name == aws.ToString(framework.Name) {
				return &framework, nil
			}
		}
	}

	return nil, &retry.NotFoundError{
		LastRequest: in,
	}
}

type dataSourceFrameworkData struct {
	ARN            types.String `tfsdk:"arn"`
	ComplianceType types.String `tfsdk:"compliance_type"`
	ControlSets    types.Set    `tfsdk:"control_sets"`
	Description    types.String `tfsdk:"description"`
	FrameworkType  types.String `tfsdk:"framework_type"`
	ID             types.String `tfsdk:"id"`
	Name           types.String `tfsdk:"name"`
	Tags           types.Map    `tfsdk:"tags"`
}

// refreshFromOutput writes state data from an AWS response object
func (rd *dataSourceFrameworkData) refreshFromOutput(ctx context.Context, meta *conns.AWSClient, out *awstypes.Framework) diag.Diagnostics {
	var diags diag.Diagnostics

	if out == nil {
		return diags
	}

	rd.ID = types.StringValue(aws.ToString(out.Id))
	rd.Name = types.StringValue(aws.ToString(out.Name))
	cs, d := flattenFrameworkControlSets(ctx, out.ControlSets)
	diags.Append(d...)
	rd.ControlSets = cs

	rd.ComplianceType = flex.StringToFramework(ctx, out.ComplianceType)
	rd.Description = flex.StringToFramework(ctx, out.Description)
	rd.FrameworkType = flex.StringValueToFramework(ctx, out.Type)
	rd.ARN = flex.StringToFramework(ctx, out.Arn)

	ignoreTagsConfig := meta.IgnoreTagsConfig
	tags := KeyValueTags(ctx, out.Tags).IgnoreAWS().IgnoreConfig(ignoreTagsConfig)
	rd.Tags = flex.FlattenFrameworkStringValueMapLegacy(ctx, tags.Map())

	return diags
}
