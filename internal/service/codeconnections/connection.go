// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package codeconnections

import (
	"context"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/codeconnections"
	awstypes "github.com/aws/aws-sdk-go-v2/service/codeconnections/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_codeconnections_connection", name="Connection")
// @Tags(identifierAttribute="arn")
func newConnectionResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &connectionResource{}

	r.SetDefaultCreateTimeout(30 * time.Minute)
	r.SetDefaultUpdateTimeout(30 * time.Minute)
	r.SetDefaultDeleteTimeout(30 * time.Minute)

	return r, nil
}

const (
	ResNameConnection = "Connection"
)

type connectionResource struct {
	framework.ResourceWithConfigure
	framework.WithImportByID
	framework.WithTimeouts
}

func (r *connectionResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrARN: framework.ARNAttributeComputedOnly(),
			"connection_status": schema.StringAttribute{
				Computed:   true,
				CustomType: fwtypes.StringEnumType[awstypes.ConnectionStatus](),
			},
			"host_arn": schema.StringAttribute{
				Optional: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.LengthBetween(0, 256),
					stringvalidator.RegexMatches(hostARNRegex, ""),
					stringvalidator.ConflictsWith(path.Expressions{
						path.MatchRoot("provider_type"),
					}...),
				},
			},
			names.AttrID: framework.IDAttribute(),
			names.AttrName: schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 32),
				},
			},
			names.AttrOwnerAccountID: schema.StringAttribute{
				Computed: true,
			},
			"provider_type": schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.ProviderType](),
				Computed:   true,
				Optional:   true,
				Validators: []validator.String{
					stringvalidator.ConflictsWith(path.Expressions{
						path.MatchRoot("host_arn"),
					}...),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
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

var (
	hostARNRegex = regexache.MustCompile("^arn:aws(-[\\w]+)*:(codestar-connections|codeconnections):.+:[0-9]{12}:host\\/.+")
)

func (r *connectionResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := r.Meta().CodeConnectionsClient(ctx)

	var data connectionResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var input codeconnections.CreateConnectionInput

	resp.Diagnostics.Append(fwflex.Expand(ctx, data, &input, fwflex.WithFieldNamePrefix("Connection"))...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Additional fields.
	input.Tags = getTagsIn(ctx)

	output, err := conn.CreateConnection(ctx, &input)

	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.CodeConnections, create.ErrActionCreating, ResNameConnection, data.ConnectionName.String(), err),
			err.Error(),
		)
		return
	}

	// Set values for unknowns
	data.ID = fwflex.StringToFramework(ctx, output.ConnectionArn)

	connection, err := waitConnectionCreated(ctx, conn, data.ID.ValueString(), r.CreateTimeout(ctx, data.Timeouts))

	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.CodeConnections, create.ErrActionWaitingForCreation, ResNameConnection, data.ConnectionName.String(), err),
			err.Error(),
		)
		return
	}

	data.ConnectionArn = fwflex.StringToFramework(ctx, connection.ConnectionArn)
	data.ConnectionName = fwflex.StringToFramework(ctx, connection.ConnectionName)
	data.ConnectionStatus = fwtypes.StringEnumValue(connection.ConnectionStatus)
	data.HostArn = fwflex.StringToFramework(ctx, connection.HostArn)
	data.OwnerAccountId = fwflex.StringToFramework(ctx, connection.OwnerAccountId)
	data.ProviderType = fwtypes.StringEnumValue(connection.ProviderType)

	resp.Diagnostics.Append(resp.State.Set(ctx, data)...)
}

func (r *connectionResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().CodeConnectionsClient(ctx)

	var data connectionResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := findConnectionByARN(ctx, conn, data.ID.ValueString())

	if tfresource.NotFound(err) {
		resp.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.CodeConnections, create.ErrActionSetting, ResNameConnection, data.ID.String(), err),
			err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(fwflex.Flatten(ctx, out, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *connectionResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var old, new connectionResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &old)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(req.Plan.Get(ctx, &new)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Update is only called when `tags` are updated.
	// Set unknowns to the old (in state) values.
	new.ConnectionArn = old.ConnectionArn
	new.ConnectionName = old.ConnectionName
	new.ConnectionStatus = old.ConnectionStatus
	new.HostArn = old.HostArn
	new.OwnerAccountId = old.OwnerAccountId
	new.ProviderType = old.ProviderType

	resp.Diagnostics.Append(resp.State.Set(ctx, &new)...)
}

func (r *connectionResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := r.Meta().CodeConnectionsClient(ctx)

	var data connectionResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	input := codeconnections.DeleteConnectionInput{
		ConnectionArn: data.ID.ValueStringPointer(),
	}

	_, err := conn.DeleteConnection(ctx, &input)

	if err != nil {
		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return
		}
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.CodeConnections, create.ErrActionDeleting, ResNameConnection, data.ID.String(), err),
			err.Error(),
		)
		return
	}

	deleteTimeout := r.DeleteTimeout(ctx, data.Timeouts)
	_, err = waitConnectionDeleted(ctx, conn, data.ID.ValueString(), deleteTimeout)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.CodeConnections, create.ErrActionWaitingForDeletion, ResNameConnection, data.ID.String(), err),
			err.Error(),
		)
		return
	}
}

func (r *connectionResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root(names.AttrID), req, resp)
}

func waitConnectionCreated(ctx context.Context, conn *codeconnections.Client, id string, timeout time.Duration) (*awstypes.Connection, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   []string{},
		Target:                    enum.Slice(awstypes.ConnectionStatusPending, awstypes.ConnectionStatusAvailable),
		Refresh:                   statusConnection(ctx, conn, id),
		Timeout:                   timeout,
		NotFoundChecks:            20,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*awstypes.Connection); ok {
		return out, err
	}

	return nil, err
}

func waitConnectionDeleted(ctx context.Context, conn *codeconnections.Client, id string, timeout time.Duration) (*awstypes.Connection, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.ConnectionStatusPending, awstypes.ConnectionStatusAvailable, awstypes.ConnectionStatusError),
		Target:  []string{},
		Refresh: statusConnection(ctx, conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*awstypes.Connection); ok {
		return out, err
	}

	return nil, err
}

func statusConnection(ctx context.Context, conn *codeconnections.Client, arn string) retry.StateRefreshFunc {
	return func() (any, string, error) {
		out, err := findConnectionByARN(ctx, conn, arn)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return out, aws.ToString((*string)(&out.ConnectionStatus)), nil
	}
}

func findConnectionByARN(ctx context.Context, conn *codeconnections.Client, arn string) (*awstypes.Connection, error) {
	input := &codeconnections.GetConnectionInput{
		ConnectionArn: aws.String(arn),
	}

	output, err := conn.GetConnection(ctx, input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.Connection == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.Connection, nil
}

type connectionResourceModel struct {
	ConnectionArn    types.String                                  `tfsdk:"arn"`
	ConnectionName   types.String                                  `tfsdk:"name"`
	ConnectionStatus fwtypes.StringEnum[awstypes.ConnectionStatus] `tfsdk:"connection_status"`
	HostArn          types.String                                  `tfsdk:"host_arn"`
	ID               types.String                                  `tfsdk:"id"`
	OwnerAccountId   types.String                                  `tfsdk:"owner_account_id"`
	ProviderType     fwtypes.StringEnum[awstypes.ProviderType]     `tfsdk:"provider_type"`
	Tags             tftags.Map                                    `tfsdk:"tags"`
	TagsAll          tftags.Map                                    `tfsdk:"tags_all"`
	Timeouts         timeouts.Value                                `tfsdk:"timeouts"`
}
