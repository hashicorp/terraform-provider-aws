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
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
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
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_grafana_workspace_service_account", name="Workspace Service Account")
func newWorkspaceServiceAccountResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	return &workspaceServiceAccountResource{}, nil
}

type workspaceServiceAccountResource struct {
	framework.ResourceWithConfigure
	framework.WithNoUpdate
	framework.WithImportByID
}

func (*workspaceServiceAccountResource) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = "aws_grafana_workspace_service_account"
}

func (r *workspaceServiceAccountResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"grafana_role": schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.Role](),
				Required:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
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
			"service_account_id": framework.IDAttribute(),
			"workspace_id": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
	}
}

func (r *workspaceServiceAccountResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data workspaceServiceAccountResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().GrafanaClient(ctx)

	name := data.Name.ValueString()
	input := &grafana.CreateWorkspaceServiceAccountInput{}
	response.Diagnostics.Append(fwflex.Expand(ctx, data, input)...)
	if response.Diagnostics.HasError() {
		return
	}

	output, err := conn.CreateWorkspaceServiceAccount(ctx, input)

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("creating Grafana Workspace Service Account (%s)", name), err.Error())

		return
	}

	// Set values for unknowns.
	data.ServiceAccountID = fwflex.StringToFramework(ctx, output.Id)
	data.setID()

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *workspaceServiceAccountResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data workspaceServiceAccountResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	if err := data.InitFromID(); err != nil {
		response.Diagnostics.AddError("parsing resource ID", err.Error())

		return
	}

	conn := r.Meta().GrafanaClient(ctx)

	output, err := findWorkspaceServiceAccountByTwoPartKey(ctx, conn, data.WorkspaceID.ValueString(), data.ServiceAccountID.ValueString())

	if tfresource.NotFound(err) {
		response.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)

		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading Grafana Workspace Service Account (%s)", data.ID.ValueString()), err.Error())

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

	// Role is returned from the API in lowercase.
	data.GrafanaRole = fwtypes.StringEnumValueToUpper(output.GrafanaRole)

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *workspaceServiceAccountResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data workspaceServiceAccountResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().GrafanaClient(ctx)

	input := &grafana.DeleteWorkspaceServiceAccountInput{}
	response.Diagnostics.Append(fwflex.Expand(ctx, data, input)...)
	if response.Diagnostics.HasError() {
		return
	}
	_, err := conn.DeleteWorkspaceServiceAccount(ctx, input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("deleting Grafana Workspace Service Account (%s)", data.ID.ValueString()), err.Error())

		return
	}
}

func findWorkspaceServiceAccount(ctx context.Context, conn *grafana.Client, input *grafana.ListWorkspaceServiceAccountsInput, filter tfslices.Predicate[*awstypes.ServiceAccountSummary]) (*awstypes.ServiceAccountSummary, error) {
	output, err := findWorkspaceServiceAccounts(ctx, conn, input, filter)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findWorkspaceServiceAccounts(ctx context.Context, conn *grafana.Client, input *grafana.ListWorkspaceServiceAccountsInput, filter tfslices.Predicate[*awstypes.ServiceAccountSummary]) ([]awstypes.ServiceAccountSummary, error) {
	var output []awstypes.ServiceAccountSummary

	pages := grafana.NewListWorkspaceServiceAccountsPaginator(conn, input)
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

		for _, v := range page.ServiceAccounts {
			if filter(&v) {
				output = append(output, v)
			}
		}
	}

	return output, nil
}

func findWorkspaceServiceAccountByTwoPartKey(ctx context.Context, conn *grafana.Client, workspaceID, serviceAccountID string) (*awstypes.ServiceAccountSummary, error) {
	input := &grafana.ListWorkspaceServiceAccountsInput{
		WorkspaceId: aws.String(workspaceID),
	}

	return findWorkspaceServiceAccount(ctx, conn, input, func(v *awstypes.ServiceAccountSummary) bool {
		return aws.ToString(v.Id) == serviceAccountID
	})
}

type workspaceServiceAccountResourceModel struct {
	GrafanaRole      fwtypes.StringEnum[awstypes.Role] `tfsdk:"grafana_role"`
	ID               types.String                      `tfsdk:"id"`
	Name             types.String                      `tfsdk:"name"`
	ServiceAccountID types.String                      `tfsdk:"service_account_id"`
	WorkspaceID      types.String                      `tfsdk:"workspace_id"`
}

const (
	workspaceServiceAccountResourceIDPartCount = 2
)

func (data *workspaceServiceAccountResourceModel) InitFromID() error {
	id := data.ID.ValueString()
	parts, err := flex.ExpandResourceId(id, workspaceServiceAccountResourceIDPartCount, false)

	if err != nil {
		return err
	}

	data.WorkspaceID = types.StringValue(parts[0])
	data.ServiceAccountID = types.StringValue(parts[1])

	return nil
}

func (data *workspaceServiceAccountResourceModel) setID() {
	data.ID = types.StringValue(errs.Must(flex.FlattenResourceId([]string{data.WorkspaceID.ValueString(), data.ServiceAccountID.ValueString()}, workspaceServiceAccountResourceIDPartCount, false)))
}
