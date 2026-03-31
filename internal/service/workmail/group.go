// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package workmail

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/YakDriver/smarterr"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/workmail"
	awstypes "github.com/aws/aws-sdk-go-v2/service/workmail/types"
	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	intflex "github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_workmail_group", name="Group")
// @IdentityAttribute("organization_id")
// @IdentityAttribute("group_id")
// @ImportIDHandler("groupImportID")
// @Testing(hasNoPreExistingResource=true)
// @Testing(importStateIdAttributes="organization_id;group_id", importStateIdAttributesSep="flex.ResourceIdSeparator")
// @Testing(existsType="github.com/aws/aws-sdk-go-v2/service/workmail;workmail.DescribeGroupOutput")
func newGroupResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	return &groupResource{}, nil
}

const (
	groupPropagationTimeout     = 2 * time.Minute
	groupDeleteTransitionTimeout = 2 * time.Minute
)

type groupResource struct {
	framework.ResourceWithModel[groupResourceModel]
	framework.WithImportByIdentity
}

func (r *groupResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"disabled_date": schema.StringAttribute{
				Description: "Timestamp when the group was disabled from WorkMail use.",
				CustomType:  timetypes.RFC3339Type{},
				Computed:    true,
			},
			names.AttrEmail: schema.StringAttribute{
				Description: "Primary email address used to register the group with WorkMail.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"enabled_date": schema.StringAttribute{
				Description: "Timestamp when the group was enabled for WorkMail use.",
				CustomType:  timetypes.RFC3339Type{},
				Computed:    true,
			},
			"group_id": schema.StringAttribute{
				Description: "Identifier of the group.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"hidden_from_global_address_list": schema.BoolAttribute{
				Description: "Whether to hide the group from the global address list.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			names.AttrName: schema.StringAttribute{
				Description: "Name of the group.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"organization_id": schema.StringAttribute{
				Description: "Identifier of the WorkMail organization where the group is managed.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			names.AttrState: schema.StringAttribute{
				Description: "Current WorkMail state of the group.",
				Computed:    true,
			},
		},
	}
}

func (r *groupResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := r.Meta().WorkMailClient(ctx)

	var plan groupResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Plan.Get(ctx, &plan))
	if resp.Diagnostics.HasError() {
		return
	}

	var input workmail.CreateGroupInput
	smerr.AddEnrich(ctx, &resp.Diagnostics, flex.Expand(ctx, plan, &input))
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := conn.CreateGroup(ctx, &input)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, plan.Name.String())
		return
	}
	if out == nil || out.GroupId == nil {
		smerr.AddError(ctx, &resp.Diagnostics, fmt.Errorf("empty output"), smerr.ID, plan.Name.String())
		return
	}

	plan.GroupId = flex.StringToFramework(ctx, out.GroupId)

	if err := registerGroup(ctx, conn, plan); err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, plan.GroupId.String())
		return
	}

	created, err := waitGroupEnabled(ctx, conn, plan.OrganizationId.ValueString(), plan.GroupId.ValueString(), groupPropagationTimeout)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, plan.GroupId.String())
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, flex.Flatten(ctx, created, &plan))
	if resp.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, plan))
}

func (r *groupResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().WorkMailClient(ctx)

	var state groupResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := findGroupByTwoPartKey(ctx, conn, state.OrganizationId.ValueString(), state.GroupId.ValueString())
	if retry.NotFound(err) {
		resp.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, state.GroupId.String())
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, flex.Flatten(ctx, out, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &state))
}

func (r *groupResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	conn := r.Meta().WorkMailClient(ctx)

	var old, new groupResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Plan.Get(ctx, &new))
	if resp.Diagnostics.HasError() {
		return
	}
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &old))
	if resp.Diagnostics.HasError() {
		return
	}

	if !new.HiddenFromGlobalAddressList.Equal(old.HiddenFromGlobalAddressList) {
		var updateInput workmail.UpdateGroupInput
		smerr.AddEnrich(ctx, &resp.Diagnostics, flex.Expand(ctx, new, &updateInput))
		if resp.Diagnostics.HasError() {
			return
		}

		if _, err := conn.UpdateGroup(ctx, &updateInput); err != nil {
			smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, new.GroupId.ValueString())
			return
		}
	}

	out, err := findGroupByTwoPartKey(ctx, conn, old.OrganizationId.ValueString(), old.GroupId.ValueString())
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, old.GroupId.String())
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, flex.Flatten(ctx, out, &new))
	if resp.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &new))
}

func (r *groupResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := r.Meta().WorkMailClient(ctx)

	var state groupResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	group, err := findGroupByTwoPartKey(ctx, conn, state.OrganizationId.ValueString(), state.GroupId.ValueString())
	if retry.NotFound(err) {
		return
	}
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, state.GroupId.String())
		return
	}

	if group.State == awstypes.EntityStateEnabled {
		if err := deregisterGroup(ctx, conn, state); err != nil {
			smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, state.GroupId.String())
			return
		}

		if _, err := waitGroupDisabled(ctx, conn, state.OrganizationId.ValueString(), state.GroupId.ValueString(), groupDeleteTransitionTimeout); err != nil {
			smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, state.GroupId.String())
			return
		}
	}

	_, err = tfresource.RetryWhenIsA[any, *awstypes.EntityStateException](ctx, groupDeleteTransitionTimeout, func(ctx context.Context) (any, error) {
		input := workmail.DeleteGroupInput{
			GroupId:        state.GroupId.ValueStringPointer(),
			OrganizationId: state.OrganizationId.ValueStringPointer(),
		}
		_, err := conn.DeleteGroup(ctx, &input)

		return nil, err
	})
	if err != nil && !errs.IsA[*awstypes.EntityNotFoundException](err) && !errs.IsA[*awstypes.ResourceNotFoundException](err) {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, state.GroupId.String())
		return
	}

	if _, err := waitGroupDeleted(ctx, conn, state.OrganizationId.ValueString(), state.GroupId.ValueString(), groupDeleteTransitionTimeout); err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, state.GroupId.String())
		return
	}
}

func registerGroup(ctx context.Context, conn *workmail.Client, data groupResourceModel) error {
	err := tfresource.Retry(ctx, groupPropagationTimeout, func(ctx context.Context) *tfresource.RetryError {
		input := workmail.RegisterToWorkMailInput{
			Email:          data.Email.ValueStringPointer(),
			EntityId:       data.GroupId.ValueStringPointer(),
			OrganizationId: data.OrganizationId.ValueStringPointer(),
		}
		_, err := conn.RegisterToWorkMail(ctx, &input)

		if errs.IsA[*awstypes.MailDomainStateException](err) || errs.IsA[*awstypes.EntityNotFoundException](err) || errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return tfresource.RetryableError(err)
		}

		if err != nil {
			return tfresource.NonRetryableError(err)
		}

		return nil
	})

	return err
}

func deregisterGroup(ctx context.Context, conn *workmail.Client, data groupResourceModel) error {
	_, err := tfresource.RetryWhenIsA[any, *awstypes.EntityStateException](ctx, groupDeleteTransitionTimeout, func(ctx context.Context) (any, error) {
		input := workmail.DeregisterFromWorkMailInput{
			EntityId:       data.GroupId.ValueStringPointer(),
			OrganizationId: data.OrganizationId.ValueStringPointer(),
		}
		_, err := conn.DeregisterFromWorkMail(ctx, &input)

		return nil, err
	})

	if errs.IsA[*awstypes.EntityNotFoundException](err) || errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil
	}

	return err
}

func waitGroupEnabled(ctx context.Context, conn *workmail.Client, organizationID, groupID string, timeout time.Duration) (*workmail.DescribeGroupOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   enum.Slice(awstypes.EntityStateDisabled),
		Target:                    enum.Slice(awstypes.EntityStateEnabled),
		Refresh:                   statusGroup(conn, organizationID, groupID),
		Timeout:                   timeout,
		NotFoundChecks:            20,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*workmail.DescribeGroupOutput); ok {
		return out, smarterr.NewError(err)
	}

	return nil, smarterr.NewError(err)
}

func waitGroupDisabled(ctx context.Context, conn *workmail.Client, organizationID, groupID string, timeout time.Duration) (*workmail.DescribeGroupOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   enum.Slice(awstypes.EntityStateEnabled),
		Target:                    enum.Slice(awstypes.EntityStateDisabled),
		Refresh:                   statusGroup(conn, organizationID, groupID),
		Timeout:                   timeout,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*workmail.DescribeGroupOutput); ok {
		return out, smarterr.NewError(err)
	}

	return nil, smarterr.NewError(err)
}

func waitGroupDeleted(ctx context.Context, conn *workmail.Client, organizationID, groupID string, timeout time.Duration) (*workmail.DescribeGroupOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.EntityStateDisabled, awstypes.EntityStateEnabled),
		Target:  []string{},
		Refresh: statusGroup(conn, organizationID, groupID),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*workmail.DescribeGroupOutput); ok {
		return out, smarterr.NewError(err)
	}

	return nil, smarterr.NewError(err)
}

func statusGroup(conn *workmail.Client, organizationID, groupID string) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		out, err := findGroupByTwoPartKey(ctx, conn, organizationID, groupID)
		if retry.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", smarterr.NewError(err)
		}

		return out, string(out.State), nil
	}
}

func findGroupByTwoPartKey(ctx context.Context, conn *workmail.Client, organizationID, groupID string) (*workmail.DescribeGroupOutput, error) {
	input := workmail.DescribeGroupInput{
		GroupId:        aws.String(groupID),
		OrganizationId: aws.String(organizationID),
	}

	out, err := conn.DescribeGroup(ctx, &input)
	if err != nil {
		if errs.IsA[*awstypes.EntityNotFoundException](err) || errs.IsA[*awstypes.ResourceNotFoundException](err) || errs.IsA[*awstypes.OrganizationStateException](err) {
			return nil, smarterr.NewError(&retry.NotFoundError{LastError: err})
		}

		return nil, smarterr.NewError(err)
	}

	if out == nil {
		return nil, smarterr.NewError(tfresource.NewEmptyResultError())
	}

	if out.State == awstypes.EntityStateDeleted {
		return nil, smarterr.NewError(&retry.NotFoundError{Message: fmt.Sprintf("WorkMail Group %s is in Deleted state", groupID)})
	}

	return out, nil
}

var (
	_ inttypes.ImportIDParser = groupImportID{}
)

type groupImportID struct{}

func (groupImportID) Parse(id string) (string, map[string]any, error) {
	organizationID, groupID, found := strings.Cut(id, intflex.ResourceIdSeparator)
	if !found {
		return "", nil, fmt.Errorf("id %q should be in the format <organization-id>%s<group-id>", id, intflex.ResourceIdSeparator)
	}

	result := map[string]any{
		"organization_id": organizationID,
		"group_id":        groupID,
	}

	return id, result, nil
}

type groupResourceModel struct {
	framework.WithRegionModel
	DisabledDate                timetypes.RFC3339 `tfsdk:"disabled_date"`
	Email                       types.String      `tfsdk:"email"`
	EnabledDate                 timetypes.RFC3339 `tfsdk:"enabled_date"`
	GroupId                     types.String      `tfsdk:"group_id"`
	HiddenFromGlobalAddressList types.Bool        `tfsdk:"hidden_from_global_address_list"`
	Name                        types.String      `tfsdk:"name"`
	OrganizationId              types.String      `tfsdk:"organization_id"`
	State                       types.String      `tfsdk:"state"`
}
