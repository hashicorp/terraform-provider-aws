// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package grafana

import (
	"context"
	"go/types"
	// "fmt"
	// "log"
	// "strings"

	// "github.com/aws/aws-sdk-go-v2/aws"
	// "github.com/aws/aws-sdk-go-v2/service/grafana"
	// awstypes "github.com/aws/aws-sdk-go-v2/service/grafana/types"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_grafana_workspace_service_account", name="ServiceAccount")
// @Tags(identifierAttribute="Id")
// @Testing(existsType="github.com/aws/aws-sdk-go-v2/service/grafana/types;types.ServiceAccountSummary")
func newResourceWorkspaceServiceAccount(_ context.Context) (resource.ResourceWithConfigure, error) {
	return &resourceWorkspaceServiceAccount{}, nil
}

type resourceWorkspaceServiceAccount struct {
	framework.ResourceWithConfigure
	framework.WithImportByID
}

func (r *resourceWorkspaceServiceAccount) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = "aws_grafana_workspace_service_account"
}

func (r *resourceWorkspaceServiceAccount) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrID: framework.IDAttribute(),
			names.AttrRoleARN: schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"service_account_name": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"service_account_role": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"workspace_id": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
	}
}

func (r *resourceWorkspaceServiceAccount) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := r.Meta().GrafanaClient()

}

func (r *resourceWorkspaceServiceAccount) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
}

func (r *resourceWorkspaceServiceAccount) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
}

func (r *resourceWorkspaceServiceAccount) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
}

type resourceWorkspaceServiceAccountModel struct {
	Alias               types.String                                             `tfsdk:"alias"`
	ARN                 types.String                                             `tfsdk:"arn"`
	Destination         fwtypes.ListNestedObjectValueOf[scraperDestinationModel] `tfsdk:"destination"`
	ID                  types.String                                             `tfsdk:"id"`
	RoleARN             types.String                                             `tfsdk:"role_arn"`
	ScrapeConfiguration types.String                                             `tfsdk:"scrape_configuration"`
	Source              fwtypes.ListNestedObjectValueOf[scraperSourceModel]      `tfsdk:"source"`
	Tags                types.Map                                                `tfsdk:"tags"`
	TagsAll             types.Map                                                `tfsdk:"tags_all"`
	Timeouts            timeouts.Value                                           `tfsdk:"timeouts"`
}
