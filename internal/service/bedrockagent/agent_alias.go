// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package bedrockagent

import (
	"context"
	"fmt"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/bedrockagent"
	awstypes "github.com/aws/aws-sdk-go-v2/service/bedrockagent/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource(name="Agent Alias")
// @Tags(identifierAttribute="agent_alias_arn")
func newAgentAliasResource(context.Context) (resource.ResourceWithConfigure, error) {
	r := &agentAliasResource{}

	r.SetDefaultCreateTimeout(5 * time.Minute)
	r.SetDefaultUpdateTimeout(5 * time.Minute)
	r.SetDefaultDeleteTimeout(5 * time.Minute)

	return r, nil
}

type agentAliasResource struct {
	framework.ResourceWithConfigure
	framework.WithImportByID
	framework.WithTimeouts
}

func (*agentAliasResource) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = "aws_bedrockagent_agent_alias"
}

func (r *agentAliasResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"agent_alias_arn": framework.ARNAttributeComputedOnly(),
			"agent_alias_id":  framework.IDAttribute(),
			"agent_alias_name": schema.StringAttribute{
				Required: true,
				Validators: []validator.String{
					stringvalidator.RegexMatches(regexache.MustCompile(`^([0-9a-zA-Z][_-]?){1,100}$`), "valid characters are a-z, A-Z, 0-9, _ (underscore) and - (hyphen). The name can have up to 100 characters"),
				},
			},
			"agent_id": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			names.AttrDescription: schema.StringAttribute{
				Optional: true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 200),
				},
			},
			names.AttrID: framework.IDAttribute(),
			"routing_configuration": schema.ListAttribute{
				CustomType: fwtypes.NewListNestedObjectTypeOf[agentAliasRoutingConfigurationListItemModel](ctx),
				Optional:   true,
				Computed:   true,
				PlanModifiers: []planmodifier.List{
					listplanmodifier.UseStateForUnknown(),
				},
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
				},
				ElementType: types.ObjectType{
					AttrTypes: fwtypes.AttributeTypesMust[agentAliasRoutingConfigurationListItemModel](ctx),
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

func (r *agentAliasResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data agentAliasResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().BedrockAgentClient(ctx)

	input := &bedrockagent.CreateAgentAliasInput{}
	response.Diagnostics.Append(fwflex.Expand(ctx, data, input)...)
	if response.Diagnostics.HasError() {
		return
	}

	input.ClientToken = aws.String(id.UniqueId())
	input.Tags = getTagsIn(ctx)

	output, err := conn.CreateAgentAlias(ctx, input)

	if err != nil {
		response.Diagnostics.AddError("creating Bedrock Agent Alias", err.Error())

		return
	}

	// Set values for unknowns.
	data.AgentAliasID = fwflex.StringToFramework(ctx, output.AgentAlias.AgentAliasId)
	data.setID()

	alias, err := waitAgentAliasCreated(ctx, conn, data.AgentAliasID.ValueString(), data.AgentID.ValueString(), r.CreateTimeout(ctx, data.Timeouts))

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("waiting for Bedrock Agent Alias (%s) create", data.ID.ValueString()), err.Error())

		return
	}

	if _, err := waitAgentVersioned(ctx, conn, data.AgentID.ValueString(), r.CreateTimeout(ctx, data.Timeouts)); err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("waiting for Bedrock Agent (%s) version", data.ID.ValueString()), err.Error())

		return
	}

	// Set values for unknowns.
	response.Diagnostics.Append(fwflex.Flatten(ctx, alias, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *agentAliasResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data agentAliasResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	if err := data.InitFromID(); err != nil {
		response.Diagnostics.AddError("parsing resource ID", err.Error())

		return
	}

	conn := r.Meta().BedrockAgentClient(ctx)

	output, err := findAgentAliasByTwoPartKey(ctx, conn, data.AgentAliasID.ValueString(), data.AgentID.ValueString())

	if tfresource.NotFound(err) {
		response.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)

		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading Bedrock Agent Alias (%s)", data.ID.String()), err.Error())

		return
	}

	response.Diagnostics.Append(fwflex.Flatten(ctx, output, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *agentAliasResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var old, new agentAliasResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &old)...)
	if response.Diagnostics.HasError() {
		return
	}
	response.Diagnostics.Append(request.Plan.Get(ctx, &new)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().BedrockAgentClient(ctx)

	if !new.AgentAliasName.Equal(old.AgentAliasName) ||
		!new.Description.Equal(old.Description) ||
		!new.RoutingConfiguration.Equal(old.RoutingConfiguration) {
		input := &bedrockagent.UpdateAgentAliasInput{}
		response.Diagnostics.Append(fwflex.Expand(ctx, new, input)...)
		if response.Diagnostics.HasError() {
			return
		}

		_, err := conn.UpdateAgentAlias(ctx, input)

		if err != nil {
			response.Diagnostics.AddError(fmt.Sprintf("reading Bedrock Agent Alias (%s)", new.ID.String()), err.Error())

			return
		}

		if _, err := waitAgentAliasUpdated(ctx, conn, new.AgentAliasID.ValueString(), new.AgentID.ValueString(), r.CreateTimeout(ctx, new.Timeouts)); err != nil {
			response.Diagnostics.AddError(fmt.Sprintf("waiting for Bedrock Agent Alias (%s) update", new.ID.ValueString()), err.Error())

			return
		}
	}

	response.Diagnostics.Append(response.State.Set(ctx, &new)...)
}

func (r *agentAliasResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data agentAliasResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().BedrockAgentClient(ctx)

	_, err := conn.DeleteAgentAlias(ctx, &bedrockagent.DeleteAgentAliasInput{
		AgentAliasId: fwflex.StringFromFramework(ctx, data.AgentAliasID),
		AgentId:      fwflex.StringFromFramework(ctx, data.AgentID),
	})

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("deleting Bedrock Agent Alias (%s)", data.ID.ValueString()), err.Error())

		return
	}
}

func (r *agentAliasResource) ModifyPlan(ctx context.Context, request resource.ModifyPlanRequest, response *resource.ModifyPlanResponse) {
	r.SetTagsAll(ctx, request, response)
}

func findAgentAliasByTwoPartKey(ctx context.Context, conn *bedrockagent.Client, agentAliasID, agentID string) (*awstypes.AgentAlias, error) {
	input := &bedrockagent.GetAgentAliasInput{
		AgentAliasId: aws.String(agentAliasID),
		AgentId:      aws.String(agentID),
	}

	output, err := conn.GetAgentAlias(ctx, input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.AgentAlias == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.AgentAlias, nil
}

func statusAgentAlias(ctx context.Context, conn *bedrockagent.Client, agentAliasID, agentID string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := findAgentAliasByTwoPartKey(ctx, conn, agentAliasID, agentID)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.AgentAliasStatus), nil
	}
}

func waitAgentAliasCreated(ctx context.Context, conn *bedrockagent.Client, agentAliasID, agentID string, timeout time.Duration) (*awstypes.AgentAlias, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.AgentAliasStatusCreating),
		Target:  enum.Slice(awstypes.AgentAliasStatusPrepared),
		Refresh: statusAgentAlias(ctx, conn, agentAliasID, agentID),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.AgentAlias); ok {
		return output, err
	}

	return nil, err
}

func waitAgentAliasUpdated(ctx context.Context, conn *bedrockagent.Client, agentAliasID, agentID string, timeout time.Duration) (*awstypes.AgentAlias, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.AgentAliasStatusUpdating),
		Target:  enum.Slice(awstypes.AgentAliasStatusPrepared),
		Refresh: statusAgentAlias(ctx, conn, agentAliasID, agentID),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.AgentAlias); ok {
		return output, err
	}

	return nil, err
}

type agentAliasResourceModel struct {
	AgentAliasARN        types.String                                                                 `tfsdk:"agent_alias_arn"`
	AgentAliasID         types.String                                                                 `tfsdk:"agent_alias_id"`
	AgentAliasName       types.String                                                                 `tfsdk:"agent_alias_name"`
	AgentID              types.String                                                                 `tfsdk:"agent_id"`
	Description          types.String                                                                 `tfsdk:"description"`
	ID                   types.String                                                                 `tfsdk:"id"`
	RoutingConfiguration fwtypes.ListNestedObjectValueOf[agentAliasRoutingConfigurationListItemModel] `tfsdk:"routing_configuration"`
	Tags                 types.Map                                                                    `tfsdk:"tags"`
	TagsAll              types.Map                                                                    `tfsdk:"tags_all"`
	Timeouts             timeouts.Value                                                               `tfsdk:"timeouts"`
}

const (
	agentAliasResourceIDPartCount = 2
)

func (m *agentAliasResourceModel) InitFromID() error {
	id := m.ID.ValueString()
	parts, err := flex.ExpandResourceId(id, agentAliasResourceIDPartCount, false)

	if err != nil {
		return err
	}

	m.AgentAliasID = types.StringValue(parts[0])
	m.AgentID = types.StringValue(parts[1])

	return nil
}

func (m *agentAliasResourceModel) setID() {
	m.ID = types.StringValue(errs.Must(flex.FlattenResourceId([]string{m.AgentAliasID.ValueString(), m.AgentID.ValueString()}, agentAliasResourceIDPartCount, false)))
}

type agentAliasRoutingConfigurationListItemModel struct {
	AgentVersion          types.String `tfsdk:"agent_version"`
	ProvisionedThroughput types.String `tfsdk:"provisioned_throughput"`
}
