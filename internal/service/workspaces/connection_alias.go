package workspaces

import (
	"context"
	"errors"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/workspaces"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

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
			"id": framework.IDAttribute(),
			"connection_string": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"owner_account_id": schema.StringAttribute{
				Computed: true,
			},
			"state": schema.StringAttribute{
				Computed: true,
			},
			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
		},
		Blocks: map[string]schema.Block{
			"timeouts": timeouts.Block(ctx, timeouts.Opts{
				Create: true,
				Update: true,
				Delete: true,
			}),
		},
	}
}

func (r *resourceConnectionAlias) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := r.Meta().WorkSpacesConn(ctx)

	var plan resourceConnectionAliasData
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	in := &workspaces.CreateConnectionAliasInput{
		ConnectionString: aws.String(plan.ConnectionString.ValueString()),
		Tags:             getTagsIn(ctx),
	}

	out, err := conn.CreateConnectionAliasWithContext(ctx, in)
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
	_, err = waitConnectionAliasCreated(ctx, conn, plan.ID.ValueString(), createTimeout)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.WorkSpaces, create.ErrActionWaitingForCreation, ResNameConnectionAlias, plan.ID.String(), err),
			err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *resourceConnectionAlias) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().WorkSpacesConn(ctx)

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

	state.ConnectionString = flex.StringToFramework(ctx, out.ConnectionString)
	state.OwnerAccountId = flex.StringToFramework(ctx, out.OwnerAccountId)
	state.State = flex.StringToFramework(ctx, out.State)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *resourceConnectionAlias) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
}

func (r *resourceConnectionAlias) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := r.Meta().WorkSpacesConn(ctx)

	var state resourceConnectionAliasData
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	in := &workspaces.DeleteConnectionAliasInput{
		AliasId: aws.String(state.ID.ValueString()),
	}

	_, err := conn.DeleteConnectionAliasWithContext(ctx, in)
	if err != nil {
		if tfawserr.ErrCodeEquals(err, workspaces.ErrCodeResourceNotFoundException) {
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
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func waitConnectionAliasCreated(ctx context.Context, conn *workspaces.WorkSpaces, id string, timeout time.Duration) (*workspaces.ConnectionAlias, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   []string{workspaces.ConnectionAliasStateCreating},
		Target:                    []string{workspaces.ConnectionAliasStateCreated},
		Refresh:                   statusConnectionAlias(ctx, conn, id),
		Timeout:                   timeout,
		NotFoundChecks:            20,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*workspaces.ConnectionAlias); ok {
		return out, err
	}

	return nil, err
}

func waitConnectionAliasDeleted(ctx context.Context, conn *workspaces.WorkSpaces, id string, timeout time.Duration) (*workspaces.ConnectionAlias, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{workspaces.ConnectionAliasStateDeleting},
		Target:  []string{},
		Refresh: statusConnectionAlias(ctx, conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*workspaces.ConnectionAlias); ok {
		return out, err
	}

	return nil, err
}

func statusConnectionAlias(ctx context.Context, conn *workspaces.WorkSpaces, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		out, err := FindConnectionAliasByID(ctx, conn, id)
		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return out, aws.StringValue(out.State), nil
	}
}

func FindConnectionAliasByID(ctx context.Context, conn *workspaces.WorkSpaces, id string) (*workspaces.ConnectionAlias, error) {
	in := &workspaces.DescribeConnectionAliasesInput{
		AliasIds: aws.StringSlice([]string{id}),
	}

	out, err := conn.DescribeConnectionAliasesWithContext(ctx, in)
	if tfawserr.ErrCodeEquals(err, workspaces.ErrCodeResourceNotFoundException) {
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

	return out.ConnectionAliases[0], nil
}

type resourceConnectionAliasData struct {
	ID               types.String   `tfsdk:"id"`
	ConnectionString types.String   `tfsdk:"name"`
	OwnerAccountId   types.String   `tfsdk:"owner_account_id"`
	State            types.String   `tfsdk:"state"`
	Tags             types.Map      `tfsdk:"tags"`
	TagsAll          types.Map      `tfsdk:"tags_all"`
	Timeouts         timeouts.Value `tfsdk:"timeouts"`
}
