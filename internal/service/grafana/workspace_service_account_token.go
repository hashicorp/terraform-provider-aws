// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package grafana

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/grafana"
	awstypes "github.com/aws/aws-sdk-go-v2/service/grafana/types"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_grafana_workspace_service_account_token", name="WorkspaceServiceAccountToken")
func newResourceWorkspaceServiceAccountToken(_ context.Context) (resource.ResourceWithConfigure, error) {
	return &resourceWorkspaceServiceAccountToken{}, nil
}

const (
	ResNameServiceAccountToken = "WorkspaceServiceAccountToken"
)

type resourceWorkspaceServiceAccountToken struct {
	framework.ResourceWithConfigure
	framework.WithNoOpUpdate[resourceWorkspaceServiceAccountTokenData]
	framework.WithImportByID
}

func (r *resourceWorkspaceServiceAccountToken) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = "aws_grafana_workspace_service_account_token"
}

func (r *resourceWorkspaceServiceAccountToken) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrCreatedAt: schema.StringAttribute{
				Computed: true,
			},
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
			"expires_at": schema.StringAttribute{
				Computed: true,
			},
			names.AttrID: framework.IDAttribute(),
			names.AttrKey: schema.StringAttribute{
				Computed:  true,
				Sensitive: true,
			},
			"seconds_to_live": schema.Int64Attribute{
				Required: true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
				Validators: []validator.Int64{
					int64validator.Between(1, 2592000),
				},
			},
			"service_account_id": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
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

func (r *resourceWorkspaceServiceAccountToken) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data resourceWorkspaceServiceAccountTokenData
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().GrafanaClient(ctx)

	input := &grafana.CreateWorkspaceServiceAccountTokenInput{
		Name:             aws.String(data.Name.ValueString()),
		SecondsToLive:    aws.Int32(int32(data.SecondsToLive.ValueInt64())),
		ServiceAccountId: aws.String(data.ServiceAccountID.ValueString()),
		WorkspaceId:      aws.String(data.WorkspaceID.ValueString()),
	}

	output, err := conn.CreateWorkspaceServiceAccountToken(ctx, input)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.Grafana, create.ErrActionCreating, ResNameServiceAccountToken, "", err),
			err.Error(),
		)
		return
	}

	//update unknowns
	saTokenID := aws.ToString(output.ServiceAccountToken.Id)
	out, err := findWorkspaceServiceAccountToken(ctx, conn, saTokenID, data.ServiceAccountID.ValueString(), data.WorkspaceID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.Grafana, create.ErrActionReading, ResNameServiceAccountToken, "", err),
			err.Error(),
		)
		return
	}

	data.ID = fwflex.StringValueToFramework(ctx, saTokenID)
	data.Key = fwflex.StringToFramework(ctx, output.ServiceAccountToken.Key)
	data.CreatedAt = fwflex.StringValueToFramework(ctx, out.CreatedAt.Format(time.RFC3339))
	data.ExpiresAt = fwflex.StringValueToFramework(ctx, out.ExpiresAt.Format(time.RFC3339))

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *resourceWorkspaceServiceAccountToken) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data resourceWorkspaceServiceAccountTokenData
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().GrafanaClient(ctx)

	output, err := findWorkspaceServiceAccountToken(ctx, conn, data.ID.ValueString(), data.ServiceAccountID.ValueString(), data.WorkspaceID.ValueString())
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

func (r *resourceWorkspaceServiceAccountToken) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data resourceWorkspaceServiceAccountTokenData
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().GrafanaClient(ctx)

	_, err := conn.DeleteWorkspaceServiceAccountToken(ctx, &grafana.DeleteWorkspaceServiceAccountTokenInput{
		ServiceAccountId: fwflex.StringFromFramework(ctx, data.ServiceAccountID),
		TokenId:          fwflex.StringFromFramework(ctx, data.ID),
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

func findWorkspaceServiceAccountToken(ctx context.Context, conn *grafana.Client, id, serviceAccountID, workspaceID string) (*awstypes.ServiceAccountTokenSummary, error) {
	input := &grafana.ListWorkspaceServiceAccountTokensInput{
		WorkspaceId:      aws.String(workspaceID),
		ServiceAccountId: aws.String(serviceAccountID),
	}

	paginator := grafana.NewListWorkspaceServiceAccountTokensPaginator(conn, input)
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, err
		}

		for _, sa := range page.ServiceAccountTokens {
			if aws.ToString(sa.Id) == id {
				return &sa, nil
			}
		}
	}

	return nil, tfresource.NewEmptyResultError(input)
}

type resourceWorkspaceServiceAccountTokenData struct {
	ID               types.String `tfsdk:"id"`
	CreatedAt        types.String `tfsdk:"created_at"`
	ExpiresAt        types.String `tfsdk:"expires_at"`
	Key              types.String `tfsdk:"key"`
	Name             types.String `tfsdk:"name"`
	SecondsToLive    types.Int64  `tfsdk:"seconds_to_live"`
	ServiceAccountID types.String `tfsdk:"service_account_id"`
	WorkspaceID      types.String `tfsdk:"workspace_id"`
}
