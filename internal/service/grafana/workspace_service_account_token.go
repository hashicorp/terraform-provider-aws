// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package grafana

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/grafana"
	awstypes "github.com/aws/aws-sdk-go-v2/service/grafana/types"
	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_grafana_workspace_service_account_token", name="Workspace Service Account Token")
func newWorkspaceServiceAccountTokenResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	return &workspaceServiceAccountTokenResource{}, nil
}

type workspaceServiceAccountTokenResource struct {
	framework.ResourceWithConfigure
	framework.WithNoUpdate
}

func (r *workspaceServiceAccountTokenResource) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = "aws_grafana_workspace_service_account_token"
}

func (r *workspaceServiceAccountTokenResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrCreatedAt: schema.StringAttribute{
				CustomType: timetypes.RFC3339Type{},
				Computed:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
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
				CustomType: timetypes.RFC3339Type{},
				Computed:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			names.AttrID: framework.IDAttribute(),
			names.AttrKey: schema.StringAttribute{
				Computed:  true,
				Sensitive: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
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
			"service_account_token_id": framework.IDAttribute(),
			"workspace_id": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
	}
}

func (r *workspaceServiceAccountTokenResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data workspaceServiceAccountTokenResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().GrafanaClient(ctx)

	name := data.Name.ValueString()
	input := &grafana.CreateWorkspaceServiceAccountTokenInput{}
	response.Diagnostics.Append(fwflex.Expand(ctx, data, input)...)
	if response.Diagnostics.HasError() {
		return
	}

	output, err := conn.CreateWorkspaceServiceAccountToken(ctx, input)

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("creating Grafana Workspace Service Account Token (%s)", name), err.Error())

		return
	}

	// Set values for unknowns.
	data.Key = fwflex.StringToFramework(ctx, output.ServiceAccountToken.Key)
	data.TokenID = fwflex.StringToFramework(ctx, output.ServiceAccountToken.Id)
	data.setID()

	serviceAccountToken, err := findWorkspaceServiceAccountTokenByThreePartKey(ctx, conn, data.WorkspaceID.ValueString(), data.ServiceAccountID.ValueString(), data.TokenID.ValueString())

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading Grafana Workspace Service Account Token (%s)", data.ID.ValueString()), err.Error())

		return
	}

	data.CreatedAt = fwflex.TimeToFramework(ctx, serviceAccountToken.CreatedAt)
	data.ExpiresAt = fwflex.TimeToFramework(ctx, serviceAccountToken.ExpiresAt)

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *workspaceServiceAccountTokenResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data workspaceServiceAccountTokenResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	if err := data.InitFromID(); err != nil {
		response.Diagnostics.AddError("parsing resource ID", err.Error())

		return
	}

	conn := r.Meta().GrafanaClient(ctx)

	output, err := findWorkspaceServiceAccountTokenByThreePartKey(ctx, conn, data.WorkspaceID.ValueString(), data.ServiceAccountID.ValueString(), data.TokenID.ValueString())

	if tfresource.NotFound(err) {
		response.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)

		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading Grafana Workspace Service Account Token (%s)", data.ID.ValueString()), err.Error())

		return
	}

	// Set attributes for import.
	response.Diagnostics.Append(fwflex.Flatten(ctx, output, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	// Restore resource ID.
	// It has been overwritten by the 'Id' field from the API response.
	data.setID()

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *workspaceServiceAccountTokenResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data workspaceServiceAccountTokenResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().GrafanaClient(ctx)

	input := &grafana.DeleteWorkspaceServiceAccountTokenInput{}
	response.Diagnostics.Append(fwflex.Expand(ctx, data, input)...)
	if response.Diagnostics.HasError() {
		return
	}

	_, err := conn.DeleteWorkspaceServiceAccountToken(ctx, input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("deleting Grafana Workspace Service Account Token (%s)", data.ID.ValueString()), err.Error())

		return
	}
}

func findWorkspaceServiceAccountToken(ctx context.Context, conn *grafana.Client, input *grafana.ListWorkspaceServiceAccountTokensInput, filter tfslices.Predicate[*awstypes.ServiceAccountTokenSummary]) (*awstypes.ServiceAccountTokenSummary, error) {
	output, err := findWorkspaceServiceAccountTokens(ctx, conn, input, filter)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findWorkspaceServiceAccountTokens(ctx context.Context, conn *grafana.Client, input *grafana.ListWorkspaceServiceAccountTokensInput, filter tfslices.Predicate[*awstypes.ServiceAccountTokenSummary]) ([]awstypes.ServiceAccountTokenSummary, error) {
	var output []awstypes.ServiceAccountTokenSummary

	pages := grafana.NewListWorkspaceServiceAccountTokensPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: input,
			}
		}

		if err != nil {
			return nil, err
		}

		for _, v := range page.ServiceAccountTokens {
			if filter(&v) {
				output = append(output, v)
			}
		}
	}

	return output, nil
}

func findWorkspaceServiceAccountTokenByThreePartKey(ctx context.Context, conn *grafana.Client, workspaceID, serviceAccountID, tokenID string) (*awstypes.ServiceAccountTokenSummary, error) {
	input := &grafana.ListWorkspaceServiceAccountTokensInput{
		ServiceAccountId: aws.String(serviceAccountID),
		WorkspaceId:      aws.String(workspaceID),
	}

	return findWorkspaceServiceAccountToken(ctx, conn, input, func(v *awstypes.ServiceAccountTokenSummary) bool {
		return aws.ToString(v.Id) == tokenID
	})
}

type workspaceServiceAccountTokenResourceModel struct {
	CreatedAt        timetypes.RFC3339 `tfsdk:"created_at"`
	ExpiresAt        timetypes.RFC3339 `tfsdk:"expires_at"`
	ID               types.String      `tfsdk:"id"`
	Key              types.String      `tfsdk:"key"`
	Name             types.String      `tfsdk:"name"`
	SecondsToLive    types.Int64       `tfsdk:"seconds_to_live"`
	ServiceAccountID types.String      `tfsdk:"service_account_id"`
	TokenID          types.String      `tfsdk:"service_account_token_id"`
	WorkspaceID      types.String      `tfsdk:"workspace_id"`
}

const (
	workspaceServiceAccountTokenResourceIDPartCount = 3
)

func (data *workspaceServiceAccountTokenResourceModel) InitFromID() error {
	id := data.ID.ValueString()
	parts, err := flex.ExpandResourceId(id, workspaceServiceAccountTokenResourceIDPartCount, false)

	if err != nil {
		return err
	}

	data.WorkspaceID = types.StringValue(parts[0])
	data.ServiceAccountID = types.StringValue(parts[1])
	data.TokenID = types.StringValue(parts[2])

	return nil
}

func (data *workspaceServiceAccountTokenResourceModel) setID() {
	data.ID = types.StringValue(errs.Must(flex.FlattenResourceId([]string{data.WorkspaceID.ValueString(), data.ServiceAccountID.ValueString(), data.TokenID.ValueString()}, workspaceServiceAccountTokenResourceIDPartCount, false)))
}
