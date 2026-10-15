// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package resiliencehubv2

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	awstypes "github.com/aws/aws-sdk-go-v2/service/resiliencehubv2/types"
	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	fwschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkDataSource("aws_resiliencehubv2_service", name="Service")
// @Tags(identifierAttribute="arn")
func newServiceDataSource(context.Context) (datasource.DataSourceWithConfigure, error) {
	return &serviceDataSource{}, nil
}

type serviceDataSource struct {
	framework.DataSourceWithModel[serviceDataSourceModel]
}

func (d *serviceDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = fwschema.Schema{
		Attributes: map[string]fwschema.Attribute{
			names.AttrARN: fwschema.StringAttribute{
				CustomType: fwtypes.ARNType,
				Required:   true,
			},
			"associated_system": framework.DataSourceComputedListOfObjectAttribute[associatedSystemModel](ctx),
			names.AttrDescription: fwschema.StringAttribute{
				Computed: true,
			},
			// Read-side view of the DevOps Agent connection, mirroring the resource's
			// computed status attributes.
			"dev_ops_agent_space_arn": fwschema.StringAttribute{
				CustomType: fwtypes.ARNType,
				Computed:   true,
			},
			"dev_ops_agent_space_id": fwschema.StringAttribute{
				Computed: true,
			},
			"dev_ops_agent_region": fwschema.StringAttribute{
				Computed: true,
			},
			"dev_ops_agent_operator_app_url": fwschema.StringAttribute{
				Computed: true,
			},
			"dev_ops_agent_connected_at": fwschema.StringAttribute{
				CustomType: timetypes.RFC3339Type{},
				Computed:   true,
			},
			"dev_ops_agent_status": fwschema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.DevOpsAgentConnectionStatus](),
				Computed:   true,
			},
			"dev_ops_agent_status_reason": fwschema.StringAttribute{
				Computed: true,
			},
			names.AttrKMSKeyID: fwschema.StringAttribute{
				Computed: true,
			},
			names.AttrName: fwschema.StringAttribute{
				Computed: true,
			},
			"permission_model": framework.DataSourceComputedListOfObjectAttribute[permissionModelModel](ctx),
			"policy_arn": fwschema.StringAttribute{
				Computed: true,
			},
			"regions": fwschema.ListAttribute{
				CustomType: fwtypes.ListOfStringType,
				Computed:   true,
			},
			names.AttrTags: tftags.TagsAttributeComputedOnly(),
		},
	}
}

func (d *serviceDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data serviceDataSourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Config.Get(ctx, &data))
	if resp.Diagnostics.HasError() {
		return
	}

	conn := d.Meta().ResilienceHubV2Client(ctx)

	arn := fwflex.StringValueFromFramework(ctx, data.ServiceARN)
	svc, err := findServiceByARN(ctx, conn, arn)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, arn)
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, fwflex.Flatten(ctx, svc, &data))
	if resp.Diagnostics.HasError() {
		return
	}

	// DevOps Agent status is read from the single read-side view
	// (svc.DevOpsAgentConfiguration). The view is present whenever the service has ever
	// been linked (DISCONNECTED after an opt-out), and omitted only when never linked.
	data.DevOpsAgentSpaceARN = fwtypes.ARNNull()
	data.DevOpsAgentStatus = fwtypes.StringEnumNull[awstypes.DevOpsAgentConnectionStatus]()
	data.DevOpsAgentSpaceID = types.StringNull()
	data.DevOpsAgentRegion = types.StringNull()
	data.DevOpsAgentOperatorAppURL = types.StringNull()
	data.DevOpsAgentConnectedAt = timetypes.NewRFC3339Null()
	data.DevOpsAgentStatusReason = types.StringNull()
	if v := svc.DevOpsAgentConfiguration; v != nil {
		if v.AgentSpaceArn != nil {
			data.DevOpsAgentSpaceARN = fwtypes.ARNValue(aws.ToString(v.AgentSpaceArn))
		}
		if v.Status != "" {
			data.DevOpsAgentStatus = fwtypes.StringEnumValue(v.Status)
		}
		data.DevOpsAgentSpaceID = fwflex.StringToFramework(ctx, v.AgentSpaceId)
		data.DevOpsAgentRegion = fwflex.StringToFramework(ctx, v.Region)
		data.DevOpsAgentOperatorAppURL = fwflex.StringToFramework(ctx, v.OperatorAppUrl)
		data.DevOpsAgentStatusReason = fwflex.StringToFramework(ctx, v.StatusReason)
		if v.ConnectedAt != nil {
			data.DevOpsAgentConnectedAt = timetypes.NewRFC3339TimeValue(aws.ToTime(v.ConnectedAt))
		}
	}

	setTagsOut(ctx, svc.Tags)

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &data))
}

type serviceDataSourceModel struct {
	framework.WithRegionModel
	AssociatedSystems         fwtypes.ListNestedObjectValueOf[associatedSystemModel]   `tfsdk:"associated_system"`
	Description               types.String                                             `tfsdk:"description"`
	DevOpsAgentConnectedAt    timetypes.RFC3339                                        `tfsdk:"dev_ops_agent_connected_at" autoflex:"-"`
	DevOpsAgentOperatorAppURL types.String                                             `tfsdk:"dev_ops_agent_operator_app_url" autoflex:"-"`
	DevOpsAgentRegion         types.String                                             `tfsdk:"dev_ops_agent_region" autoflex:"-"`
	DevOpsAgentSpaceARN       fwtypes.ARN                                              `tfsdk:"dev_ops_agent_space_arn" autoflex:"-"`
	DevOpsAgentSpaceID        types.String                                             `tfsdk:"dev_ops_agent_space_id" autoflex:"-"`
	DevOpsAgentStatus         fwtypes.StringEnum[awstypes.DevOpsAgentConnectionStatus] `tfsdk:"dev_ops_agent_status" autoflex:"-"`
	DevOpsAgentStatusReason   types.String                                             `tfsdk:"dev_ops_agent_status_reason" autoflex:"-"`
	KMSKeyID                  types.String                                             `tfsdk:"kms_key_id"`
	Name                      types.String                                             `tfsdk:"name"`
	PermissionModel           fwtypes.ListNestedObjectValueOf[permissionModelModel]    `tfsdk:"permission_model"`
	PolicyARN                 types.String                                             `tfsdk:"policy_arn"`
	Regions                   fwtypes.ListOfString                                     `tfsdk:"regions"`
	ServiceARN                fwtypes.ARN                                              `tfsdk:"arn"`
	Tags                      tftags.Map                                               `tfsdk:"tags"`
}
