// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package workspaces

import (
	"context"
	"errors"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/workspaces"
	awstypes "github.com/aws/aws-sdk-go-v2/service/workspaces/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

var ResourceConnectionAlias = newResourceConnectionAlias

// @FrameworkResource(name="Connection Alias")
// @Tags(identifierAttribute="id")
func newResourceConnectionAlias(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &resourceConnectionAlias{}

	r.SetDefaultCreateTimeout(30 * time.Minute)
	r.SetDefaultUpdateTimeout(30 * time.Minute)
	r.SetDefaultDeleteTimeout(30 * time.Minute)

	return r, nil
}

const (
	ResNameConnectionAlias = "Connection Alias"
)

type resourceConnectionAlias struct {
	framework.ResourceWithConfigure
	framework.WithTimeouts
}

func (r *resourceConnectionAlias) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "aws_workspaces_connection_alias"
}

func (r *resourceConnectionAlias) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrID: framework.IDAttribute(),
			"connection_string": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Description: "The connection string specified for the connection alias. The connection string must be in the form of a fully qualified domain name (FQDN), such as www.example.com.",
			},
			names.AttrOwnerAccountID: schema.StringAttribute{
				Computed:    true,
				Description: "The identifier of the Amazon Web Services account that owns the connection alias.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			names.AttrState: schema.StringAttribute{
				Computed:    true,
				Description: "The current state of the connection alias.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
		},
		Blocks: map[string]schema.Block{
			names.AttrTimeouts: timeouts.Block(ctx, timeouts.Opts{
				Create: true,
				Update: true,
				Delete: true,
			}),
		},
	}
}

func (r *resourceConnectionAlias) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := r.Meta().WorkSpacesClient(ctx)

	var plan resourceConnectionAliasData
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	in := &workspaces.CreateConnectionAliasInput{
		ConnectionString: aws.String(plan.ConnectionString.ValueString()),
		Tags:             getTagsIn(ctx),
	}

	out, err := conn.CreateConnectionAlias(ctx, in)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.WorkSpaces, create.ErrActionCreating, ResNameConnectionAlias, plan.ConnectionString.String(), err),
			err.Error(),
		)
		return
	}
	if out == nil || out.AliasId == nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.WorkSpaces, create.ErrActionCreating, ResNameConnectionAlias, plan.ConnectionString.String(), nil),
			errors.New("empty output").Error(),
		)
		return
	}

	plan.ID = flex.StringToFramework(ctx, out.AliasId)

	createTimeout := r.CreateTimeout(ctx, plan.Timeouts)
	alias, err := waitConnectionAliasCreated(ctx, conn, plan.ID.ValueString(), createTimeout)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.WorkSpaces, create.ErrActionWaitingForCreation, ResNameConnectionAlias, plan.ID.String(), err),
			err.Error(),
		)
		return
	}

	plan.update(ctx, alias)
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *resourceConnectionAlias) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().WorkSpacesClient(ctx)

	var state resourceConnectionAliasData
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := FindConnectionAliasByID(ctx, conn, state.ID.ValueString())
	if tfresource.NotFound(err) {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.WorkSpaces, create.ErrActionSetting, ResNameConnectionAlias, state.ID.String(), err),
			err.Error(),
		)
		return
	}

	state.update(ctx, out)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *resourceConnectionAlias) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state resourceConnectionAliasData
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *resourceConnectionAlias) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := r.Meta().WorkSpacesClient(ctx)

	var state resourceConnectionAliasData
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	in := &workspaces.DeleteConnectionAliasInput{
		AliasId: aws.String(state.ID.ValueString()),
	}

	_, err := conn.DeleteConnectionAlias(ctx, in)
	if err != nil {
		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return
		}
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.WorkSpaces, create.ErrActionDeleting, ResNameConnectionAlias, state.ID.String(), err),
			err.Error(),
		)
		return
	}

	deleteTimeout := r.DeleteTimeout(ctx, state.Timeouts)
	_, err = waitConnectionAliasDeleted(ctx, conn, state.ID.ValueString(), deleteTimeout)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.WorkSpaces, create.ErrActionWaitingForDeletion, ResNameConnectionAlias, state.ID.String(), err),
			err.Error(),
		)
		return
	}
}

func (r *resourceConnectionAlias) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root(names.AttrID), req, resp)
}

func (r *resourceConnectionAlias) ModifyPlan(ctx context.Context, request resource.ModifyPlanRequest, response *resource.ModifyPlanResponse) {
	r.SetTagsAll(ctx, request, response)
}

func (data *resourceConnectionAliasData) update(ctx context.Context, in *awstypes.ConnectionAlias) {
	data.ConnectionString = flex.StringToFramework(ctx, in.ConnectionString)
	data.OwnerAccountId = flex.StringToFramework(ctx, in.OwnerAccountId)
	data.State = flex.StringValueToFramework(ctx, in.State)
}

func waitConnectionAliasCreated(ctx context.Context, conn *workspaces.Client, id string, timeout time.Duration) (*awstypes.ConnectionAlias, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   enum.Slice(awstypes.ConnectionAliasStateCreating),
		Target:                    enum.Slice(awstypes.ConnectionAliasStateCreated),
		Refresh:                   statusConnectionAlias(ctx, conn, id),
		Timeout:                   timeout,
		NotFoundChecks:            20,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*awstypes.ConnectionAlias); ok {
		return out, err
	}

	return nil, err
}

func waitConnectionAliasDeleted(ctx context.Context, conn *workspaces.Client, id string, timeout time.Duration) (*awstypes.ConnectionAlias, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.ConnectionAliasStateDeleting),
		Target:  []string{},
		Refresh: statusConnectionAlias(ctx, conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*awstypes.ConnectionAlias); ok {
		return out, err
	}

	return nil, err
}

func statusConnectionAlias(ctx context.Context, conn *workspaces.Client, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		out, err := FindConnectionAliasByID(ctx, conn, id)
		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return out, string(out.State), nil
	}
}

func FindConnectionAliasByID(ctx context.Context, conn *workspaces.Client, id string) (*awstypes.ConnectionAlias, error) {
	in := &workspaces.DescribeConnectionAliasesInput{
		AliasIds: []string{id},
	}

	out, err := conn.DescribeConnectionAliases(ctx, in)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: in,
		}
	}

	if err != nil {
		return nil, err
	}

	if out == nil || len(out.ConnectionAliases) == 0 {
		return nil, tfresource.NewEmptyResultError(in)
	}

	return &out.ConnectionAliases[0], nil
}

type resourceConnectionAliasData struct {
	ID               types.String   `tfsdk:"id"`
	ConnectionString types.String   `tfsdk:"connection_string"`
	OwnerAccountId   types.String   `tfsdk:"owner_account_id"`
	State            types.String   `tfsdk:"state"`
	Tags             types.Map      `tfsdk:"tags"`
	TagsAll          types.Map      `tfsdk:"tags_all"`
	Timeouts         timeouts.Value `tfsdk:"timeouts"`
}
