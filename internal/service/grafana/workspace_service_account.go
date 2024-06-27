// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package grafana

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/grafana"
	awstypes "github.com/aws/aws-sdk-go-v2/service/grafana/types"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_grafana_workspace_service_account", name="WorkspaceServiceAccount")
func newResourceWorkspaceServiceAccount(_ context.Context) (resource.ResourceWithConfigure, error) {
	return &resourceWorkspaceServiceAccount{}, nil
}

const (
	ResNameServiceAccount = "WorkspaceServiceAccount"
)

type resourceWorkspaceServiceAccount struct {
	framework.ResourceWithConfigure
	framework.WithNoOpUpdate[resourceWorkspaceServiceAccountData]
	framework.WithImportByID
}

func (r *resourceWorkspaceServiceAccount) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = "aws_grafana_workspace_service_account"
}

func (r *resourceWorkspaceServiceAccount) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrID: framework.IDAttribute(),
			names.AttrName: schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
					stringvalidator.LengthAtMost(128),
				},
			},
			"grafana_role": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					enum.FrameworkValidate[awstypes.Role](),
				},
			},
			names.AttrWorkspaceID: schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
	}
}

func (r *resourceWorkspaceServiceAccount) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data resourceWorkspaceServiceAccountData
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().GrafanaClient(ctx)

	input := &grafana.CreateWorkspaceServiceAccountInput{
		Name:        aws.String(data.Name.ValueString()),
		GrafanaRole: awstypes.Role(data.ServiceAccountRole.ValueString()),
		WorkspaceId: aws.String(data.WorkspaceID.ValueString()),
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
	var data resourceWorkspaceServiceAccountData
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().GrafanaClient(ctx)

	output, err := findWorkspaceServiceAccount(ctx, conn, data.ID.ValueString(), data.WorkspaceID.ValueString())
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
	var data resourceWorkspaceServiceAccountData
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

func (r *resourceWorkspaceServiceAccount) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	const (
		partCount = 3
	)
	parts, err := flex.ExpandResourceId(req.ID, partCount, false)
	if err != nil {
		resp.Diagnostics.AddError(fmt.Sprintf("importing Workspace Service Account ID (%s)", req.ID), err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root(names.AttrID), parts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("grafana_role"), parts[1])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root(names.AttrWorkspaceID), parts[2])...)
}

func findWorkspaceServiceAccount(ctx context.Context, conn *grafana.Client, id, workspaceID string) (*awstypes.ServiceAccountSummary, error) {
	if workspaceID == "" {
		return nil, errs.NewErrorWithMessage(fmt.Errorf("workspace_id is required to find the service account"))
	}
	input := &grafana.ListWorkspaceServiceAccountsInput{
		WorkspaceId: aws.String(workspaceID),
	}

	paginator := grafana.NewListWorkspaceServiceAccountsPaginator(conn, input)
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, err
		}

		for _, sa := range page.ServiceAccounts {
			if aws.ToString(sa.Id) == id {
				return &sa, nil
			}
		}
	}

	return nil, tfresource.NewEmptyResultError(input)
}

type resourceWorkspaceServiceAccountData struct {
	ID                 types.String `tfsdk:"id"`
	Name               types.String `tfsdk:"name"`
	ServiceAccountRole types.String `tfsdk:"grafana_role"`
	WorkspaceID        types.String `tfsdk:"workspace_id"`
}
