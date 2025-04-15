// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package workspaces

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/workspaces"
	awstypes "github.com/aws/aws-sdk-go-v2/service/workspaces/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_workspaces_connection_alias", name="Connection Alias")
// @Tags(identifierAttribute="id")
func newConnectionAliasResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &connectionAliasResource{}

	r.SetDefaultCreateTimeout(30 * time.Minute)
	r.SetDefaultDeleteTimeout(30 * time.Minute)

	return r, nil
}

type connectionAliasResource struct {
	framework.ResourceWithConfigure
	framework.WithNoOpUpdate[connectionAliasResourceModel]
	framework.WithImportByID
	framework.WithTimeouts
}

func (r *connectionAliasResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"connection_string": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Description: "The connection string specified for the connection alias. The connection string must be in the form of a fully qualified domain name (FQDN), such as www.example.com.",
			},
			names.AttrID: framework.IDAttribute(),
			names.AttrOwnerAccountID: schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
				Description: "The identifier of the Amazon Web Services account that owns the connection alias.",
			},
			names.AttrState: schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
				Description: "The current state of the connection alias.",
			},
			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
		},
		Blocks: map[string]schema.Block{
			names.AttrTimeouts: timeouts.Block(ctx, timeouts.Opts{
				Create: true,
				Delete: true,
			}),
		},
	}
}

func (r *connectionAliasResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data connectionAliasResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().WorkSpacesClient(ctx)

	input := &workspaces.CreateConnectionAliasInput{
		ConnectionString: fwflex.StringFromFramework(ctx, data.ConnectionString),
		Tags:             getTagsIn(ctx),
	}

	output, err := conn.CreateConnectionAlias(ctx, input)

	if err != nil {
		response.Diagnostics.AddError("creating WorkSpaces Connection Alias", err.Error())

		return
	}

	// Set values for unknowns.
	data.ID = fwflex.StringToFramework(ctx, output.AliasId)

	alias, err := waitConnectionAliasCreated(ctx, conn, data.ID.ValueString(), r.CreateTimeout(ctx, data.Timeouts))

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("waiting for WorkSpaces Connection Alias (%s) create", data.ID.ValueString()), err.Error())

		return
	}

	// Set values for unknowns.
	data.OwnerAccountId = fwflex.StringToFramework(ctx, alias.OwnerAccountId)
	data.State = fwflex.StringValueToFramework(ctx, alias.State)

	response.Diagnostics.Append(response.State.Set(ctx, data)...)
}

func (r *connectionAliasResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data connectionAliasResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().WorkSpacesClient(ctx)

	alias, err := findConnectionAliasByID(ctx, conn, data.ID.ValueString())

	if tfresource.NotFound(err) {
		response.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)
		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading WorkSpaces Connection Alias (%s)", data.ID.ValueString()), err.Error())

		return
	}

	// Set attributes for import.
	data.ConnectionString = fwflex.StringToFramework(ctx, alias.ConnectionString)
	data.OwnerAccountId = fwflex.StringToFramework(ctx, alias.OwnerAccountId)
	data.State = fwflex.StringValueToFramework(ctx, alias.State)

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *connectionAliasResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data connectionAliasResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().WorkSpacesClient(ctx)

	input := workspaces.DeleteConnectionAliasInput{
		AliasId: data.ID.ValueStringPointer(),
	}
	_, err := conn.DeleteConnectionAlias(ctx, &input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("deleting WorkSpaces Connection Alias (%s)", data.ID.ValueString()), err.Error())

		return
	}

	if _, err := waitConnectionAliasDeleted(ctx, conn, data.ID.ValueString(), r.DeleteTimeout(ctx, data.Timeouts)); err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("waiting for WorkSpaces Connection Alias (%s) delete", data.ID.ValueString()), err.Error())

		return
	}
}

func findConnectionAliasByID(ctx context.Context, conn *workspaces.Client, id string) (*awstypes.ConnectionAlias, error) {
	input := &workspaces.DescribeConnectionAliasesInput{
		AliasIds: []string{id},
	}

	return findConnectionAlias(ctx, conn, input)
}

func findConnectionAlias(ctx context.Context, conn *workspaces.Client, input *workspaces.DescribeConnectionAliasesInput) (*awstypes.ConnectionAlias, error) {
	output, err := findConnectionAliases(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findConnectionAliases(ctx context.Context, conn *workspaces.Client, input *workspaces.DescribeConnectionAliasesInput) ([]awstypes.ConnectionAlias, error) {
	var output []awstypes.ConnectionAlias

	err := describeConnectionAliasesPages(ctx, conn, input, func(page *workspaces.DescribeConnectionAliasesOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		output = append(output, page.ConnectionAliases...)

		return !lastPage
	})

	if err != nil {
		return nil, err
	}

	return output, nil
}

func statusConnectionAlias(ctx context.Context, conn *workspaces.Client, id string) retry.StateRefreshFunc {
	return func() (any, string, error) {
		output, err := findConnectionAliasByID(ctx, conn, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.State), nil
	}
}

func waitConnectionAliasCreated(ctx context.Context, conn *workspaces.Client, id string, timeout time.Duration) (*awstypes.ConnectionAlias, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.ConnectionAliasStateCreating),
		Target:  enum.Slice(awstypes.ConnectionAliasStateCreated),
		Refresh: statusConnectionAlias(ctx, conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.ConnectionAlias); ok {
		return output, err
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

	if output, ok := outputRaw.(*awstypes.ConnectionAlias); ok {
		return output, err
	}

	return nil, err
}

type connectionAliasResourceModel struct {
	ConnectionString types.String   `tfsdk:"connection_string"`
	ID               types.String   `tfsdk:"id"`
	OwnerAccountId   types.String   `tfsdk:"owner_account_id"`
	State            types.String   `tfsdk:"state"`
	Tags             tftags.Map     `tfsdk:"tags"`
	TagsAll          tftags.Map     `tfsdk:"tags_all"`
	Timeouts         timeouts.Value `tfsdk:"timeouts"`
}
