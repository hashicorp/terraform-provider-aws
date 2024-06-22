// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package grafana

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/grafana"
	awstypes "github.com/aws/aws-sdk-go-v2/service/grafana/types"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_grafana_workspace_service_account", name="ServiceAccount")
// @Tags(identifierAttribute="id")
func newWorkspaceServiceAccountResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	return &resourceWorkspaceServiceAccount{}, nil
}

const (
	ResNameServiceAccount = "ServiceAccount"
)

type resourceWorkspaceServiceAccount struct {
	framework.ResourceWithConfigure
	framework.WithNoOpUpdate[workspaceServiceAccountResourceModel]
	framework.WithImportByID
}

func (r *resourceWorkspaceServiceAccount) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = "aws_grafana_workspace_service_account"
}

func (r *resourceWorkspaceServiceAccount) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrID: framework.IDAttribute(),
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
	var data workspaceServiceAccountResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().GrafanaClient(ctx)

	input := &grafana.CreateWorkspaceServiceAccountInput{}
	resp.Diagnostics.Append(fwflex.Expand(ctx, data, input)...)

	if resp.Diagnostics.HasError() {
		return
	}

	output, err := conn.CreateWorkspaceServiceAccount(ctx, input)

	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.Grafana, create.ErrActionCreating, ResNameServiceAccount, "", err),
			err.Error(),
		)
		return
	}

	data.ID = fwflex.StringToFramework(ctx, output.Id)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *resourceWorkspaceServiceAccount) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data workspaceServiceAccountResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().GrafanaClient(ctx)

	output, err := findWorkspaceServiceAccountByID(ctx, conn, data.ID.ValueString(), data.WorkspaceID.ValueString())
	if tfresource.NotFound(err) {
		resp.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		resp.State.RemoveResource(ctx)
		return
	}

	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.Grafana, create.ErrActionSetting, ResNameServiceAccount, data.ID.String(), err),
			err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(fwflex.Flatten(ctx, output, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *resourceWorkspaceServiceAccount) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data workspaceServiceAccountResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().GrafanaClient(ctx)

	_, err := conn.DeleteWorkspaceServiceAccount(ctx, &grafana.DeleteWorkspaceServiceAccountInput{
		ServiceAccountId: fwflex.StringFromFramework(ctx, data.ID),
		WorkspaceId:      fwflex.StringFromFramework(ctx, data.WorkspaceID),
	})

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return
	}

	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.Grafana, create.ErrActionDeleting, ResNameServiceAccount, data.ID.String(), err),
			err.Error(),
		)
		return
	}
}

func findWorkspaceServiceAccountByID(ctx context.Context, conn *grafana.Client, id, workspaceID string) (*awstypes.ServiceAccountSummary, error) {

	input := &grafana.ListWorkspaceServiceAccountsInput{
		WorkspaceId: aws.String(workspaceID),
	}

	output, err := conn.ListWorkspaceServiceAccounts(ctx, input)

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	for _, sa := range output.ServiceAccounts {
		if aws.ToString(sa.Id) == id {
			return &sa, nil
		}
	}

	return nil, errs.NewErrorWithMessage(fmt.Errorf("service account %s on workspaceId %s not found", id, workspaceID))
}

type workspaceServiceAccountResourceModel struct {
	ID                 types.String `tfsdk:"id"`
	ServiceAccountName types.String `tfsdk:"service_account_name"`
	ServiceAccountRole types.String `tfsdk:"service_account_role"`
	WorkspaceID        types.String `tfsdk:"workspace_id"`
}
