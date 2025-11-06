// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package networkflowmonitor

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/aws/aws-sdk-go-v2/service/networkflowmonitor"
	awstypes "github.com/aws/aws-sdk-go-v2/service/networkflowmonitor/types"
	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_networkflowmonitor_scope", name="Scope")
// @Tags(identifierAttribute="arn")
func newScopeResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &scopeResource{}

	r.SetDefaultCreateTimeout(30 * time.Minute)
	r.SetDefaultUpdateTimeout(30 * time.Minute)
	r.SetDefaultDeleteTimeout(30 * time.Minute)

	return r, nil
}

type scopeResource struct {
	framework.ResourceWithConfigure
	framework.ResourceWithModel[scopeResourceModel]
	framework.WithTimeouts
}

func (r *scopeResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrARN: framework.ARNAttributeComputedOnly(),
			names.AttrID:  framework.IDAttribute(),
			"scope_id": schema.StringAttribute{
				Computed: true,
			},
			names.AttrStatus: schema.StringAttribute{
				Computed: true,
			},
			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
		},
		Blocks: map[string]schema.Block{
			"targets": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[targetResourceModel](ctx),
				Validators: []validator.List{
					listvalidator.SizeAtLeast(1),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						names.AttrRegion: schema.StringAttribute{
							Required: true,
						},
					},
					Blocks: map[string]schema.Block{
						"target_identifier": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[targetIdentifierModel](ctx),
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
								listvalidator.IsRequired(),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"target_id": schema.StringAttribute{
										Required: true,
									},
									"target_type": schema.StringAttribute{
										Required: true,
									},
								},
							},
						},
					},
				},
			},
			names.AttrTimeouts: timeouts.Block(ctx, timeouts.Opts{
				Create: true,
				Update: true,
				Delete: true,
			}),
		},
	}
}

type scopeResourceModel struct {
	framework.WithRegionModel
	ARN      types.String                                         `tfsdk:"arn"`
	ID       types.String                                         `tfsdk:"id"`
	ScopeId  types.String                                         `tfsdk:"scope_id"`
	Status   types.String                                         `tfsdk:"status"`
	Targets  fwtypes.ListNestedObjectValueOf[targetResourceModel] `tfsdk:"targets"`
	Tags     tftags.Map                                           `tfsdk:"tags"`
	TagsAll  tftags.Map                                           `tfsdk:"tags_all"`
	Timeouts timeouts.Value                                       `tfsdk:"timeouts"`
}

type targetResourceModel struct {
	Region           types.String                                           `tfsdk:"region"`
	TargetIdentifier fwtypes.ListNestedObjectValueOf[targetIdentifierModel] `tfsdk:"target_identifier"`
}

type targetIdentifierModel struct {
	TargetId   types.String `tfsdk:"target_id"`
	TargetType types.String `tfsdk:"target_type"`
}

func (r *scopeResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data scopeResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().NetworkFlowMonitorClient(ctx)

	// Manual parameter mapping (AutoFlex can't handle union types)
	input := networkflowmonitor.CreateScopeInput{}

	// Map targets manually since they contain union types
	if !data.Targets.IsNull() && !data.Targets.IsUnknown() {
		targets, diags := data.Targets.ToSlice(ctx)
		response.Diagnostics.Append(diags...)
		if response.Diagnostics.HasError() {
			return
		}

		input.Targets = make([]awstypes.TargetResource, len(targets))
		for i, target := range targets {
			input.Targets[i].Region = target.Region.ValueStringPointer()

			if !target.TargetIdentifier.IsNull() && !target.TargetIdentifier.IsUnknown() {
				targetIds, diags := target.TargetIdentifier.ToSlice(ctx)
				response.Diagnostics.Append(diags...)
				if response.Diagnostics.HasError() {
					return
				}

				if len(targetIds) > 0 {
					targetId := targetIds[0]
					input.Targets[i].TargetIdentifier = &awstypes.TargetIdentifier{
						TargetId: &awstypes.TargetIdMemberAccountId{
							Value: targetId.TargetId.ValueString(),
						},
						TargetType: awstypes.TargetType(targetId.TargetType.ValueString()),
					}
				}
			}
		}
	}

	// Set additional fields that need special handling
	input.ClientToken = aws.String(uuid.New().String())
	input.Tags = getTagsIn(ctx)

	output, err := conn.CreateScope(ctx, &input)
	if err != nil {
		response.Diagnostics.AddError("creating Network Flow Monitor Scope", err.Error())
		return
	}

	// Set ID and computed attributes from create output
	data.ID = types.StringValue(aws.ToString(output.ScopeArn))
	data.ARN = types.StringValue(aws.ToString(output.ScopeArn))

	// Wait for scope to be ready
	scope, err := waitScopeCreated(ctx, conn, aws.ToString(output.ScopeId), r.CreateTimeout(ctx, data.Timeouts))
	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("waiting for Network Flow Monitor Scope (%s) create", data.ID.ValueString()), err.Error())
		return
	}

	// Set all attributes from final scope state manually
	data.ScopeId = types.StringValue(aws.ToString(scope.ScopeId))
	data.Status = types.StringValue(string(scope.Status))

	// Handle targets with union types manually
	if len(scope.Targets) > 0 {
		targetModels := make([]targetResourceModel, len(scope.Targets))
		for i, target := range scope.Targets {
			targetModels[i].Region = types.StringValue(aws.ToString(target.Region))

			if target.TargetIdentifier != nil {
				var targetIdValue string
				if accountId, ok := target.TargetIdentifier.TargetId.(*awstypes.TargetIdMemberAccountId); ok {
					targetIdValue = accountId.Value
				}

				targetIdModel := targetIdentifierModel{
					TargetId:   types.StringValue(targetIdValue),
					TargetType: types.StringValue(string(target.TargetIdentifier.TargetType)),
				}

				targetIdentifierList, diags := fwtypes.NewListNestedObjectValueOfValueSlice(ctx, []targetIdentifierModel{targetIdModel})
				response.Diagnostics.Append(diags...)
				if response.Diagnostics.HasError() {
					return
				}
				targetModels[i].TargetIdentifier = targetIdentifierList
			}
		}

		targetsList, diags := fwtypes.NewListNestedObjectValueOfValueSlice(ctx, targetModels)
		response.Diagnostics.Append(diags...)
		if response.Diagnostics.HasError() {
			return
		}
		data.Targets = targetsList
	}

	// Set tags - preserve the original plan's tags state
	// If tags were null in the plan, keep them null in the state
	if !data.Tags.IsNull() {
		tags := tftags.New(ctx, scope.Tags)
		data.Tags = tftags.FlattenStringValueMap(ctx, tags.Map())
	}
	// TagsAll should always reflect the actual AWS state
	tags := tftags.New(ctx, scope.Tags)
	data.TagsAll = tftags.FlattenStringValueMap(ctx, tags.Map())

	response.Diagnostics.Append(response.State.Set(ctx, data)...)
}

func (r *scopeResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data scopeResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().NetworkFlowMonitorClient(ctx)

	output, err := findScopeByID(ctx, conn, data.ScopeId.ValueString())

	if tfresource.NotFound(err) {
		response.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)
		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading Network Flow Monitor Scope (%s)", data.ID.ValueString()), err.Error())
		return
	}

	// Manual parameter mapping (AutoFlex can't handle union types)
	// Set basic attributes
	data.ID = types.StringValue(aws.ToString(output.ScopeArn))
	data.ARN = types.StringValue(aws.ToString(output.ScopeArn))
	data.ScopeId = types.StringValue(aws.ToString(output.ScopeId))
	data.Status = types.StringValue(string(output.Status))

	// Handle targets with union types manually
	if len(output.Targets) > 0 {
		targetModels := make([]targetResourceModel, len(output.Targets))
		for i, target := range output.Targets {
			targetModels[i].Region = types.StringValue(aws.ToString(target.Region))

			if target.TargetIdentifier != nil {
				var targetIdValue string
				if accountId, ok := target.TargetIdentifier.TargetId.(*awstypes.TargetIdMemberAccountId); ok {
					targetIdValue = accountId.Value
				}

				targetIdModel := targetIdentifierModel{
					TargetId:   types.StringValue(targetIdValue),
					TargetType: types.StringValue(string(target.TargetIdentifier.TargetType)),
				}

				targetIdentifierList, diags := fwtypes.NewListNestedObjectValueOfValueSlice(ctx, []targetIdentifierModel{targetIdModel})
				response.Diagnostics.Append(diags...)
				if response.Diagnostics.HasError() {
					return
				}
				targetModels[i].TargetIdentifier = targetIdentifierList
			}
		}

		targetsList, diags := fwtypes.NewListNestedObjectValueOfValueSlice(ctx, targetModels)
		response.Diagnostics.Append(diags...)
		if response.Diagnostics.HasError() {
			return
		}
		data.Targets = targetsList
	}

	// Set tags - only set if there are actual tags, otherwise keep null
	if len(output.Tags) > 0 {
		tags := tftags.New(ctx, output.Tags)
		data.Tags = tftags.FlattenStringValueMap(ctx, tags.Map())
		data.TagsAll = tftags.FlattenStringValueMap(ctx, tags.Map())
	} else {
		// Keep tags null if no tags were originally set and none exist in AWS
		// TagsAll should always reflect AWS state (empty map when no tags)
		tags := tftags.New(ctx, output.Tags)
		data.TagsAll = tftags.FlattenStringValueMap(ctx, tags.Map())
	}

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *scopeResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var old, new scopeResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &old)...)
	if response.Diagnostics.HasError() {
		return
	}
	response.Diagnostics.Append(request.Plan.Get(ctx, &new)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().NetworkFlowMonitorClient(ctx)

	// Handle targets updates
	if !new.Targets.Equal(old.Targets) {
		// Calculate targets to add and remove
		resourcesToAdd, resourcesToDelete, diags := r.calculateTargetChanges(ctx, old.Targets, new.Targets)
		response.Diagnostics.Append(diags...)
		if response.Diagnostics.HasError() {
			return
		}

		// Update scope targets if there are changes
		if len(resourcesToAdd) > 0 || len(resourcesToDelete) > 0 {
			input := networkflowmonitor.UpdateScopeInput{
				ScopeId:           old.ScopeId.ValueStringPointer(),
				ResourcesToAdd:    resourcesToAdd,
				ResourcesToDelete: resourcesToDelete,
			}

			_, err := conn.UpdateScope(ctx, &input)
			if err != nil {
				response.Diagnostics.AddError(fmt.Sprintf("updating Network Flow Monitor Scope (%s) targets", new.ID.ValueString()), err.Error())
				return
			}

			// Wait for scope to be updated
			_, err = waitScopeUpdated(ctx, conn, old.ScopeId.ValueString(), r.UpdateTimeout(ctx, new.Timeouts))
			if err != nil {
				response.Diagnostics.AddError(fmt.Sprintf("waiting for Network Flow Monitor Scope (%s) update", new.ID.ValueString()), err.Error())
				return
			}
		}
	}

	// Handle tag updates
	if !new.Tags.Equal(old.Tags) {
		if err := updateTags(ctx, conn, new.ID.ValueString(), old.Tags, new.Tags); err != nil {
			response.Diagnostics.AddError(fmt.Sprintf("updating Network Flow Monitor Scope (%s) tags", new.ID.ValueString()), err.Error())
			return
		}
	}

	// After updating, read the current state from AWS to get all computed values
	output, err := findScopeByID(ctx, conn, old.ScopeId.ValueString())
	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading Network Flow Monitor Scope (%s) after update", new.ID.ValueString()), err.Error())
		return
	}

	// Update the new model with current AWS state
	new.ScopeId = types.StringValue(aws.ToString(output.ScopeId))
	new.Status = types.StringValue(string(output.Status))

	// Map targets manually from AWS response
	if len(output.Targets) > 0 {
		targetModels := make([]targetResourceModel, len(output.Targets))
		for i, target := range output.Targets {
			targetModels[i].Region = types.StringValue(aws.ToString(target.Region))

			if target.TargetIdentifier != nil {
				var targetIdValue string
				if accountId, ok := target.TargetIdentifier.TargetId.(*awstypes.TargetIdMemberAccountId); ok {
					targetIdValue = accountId.Value
				}

				targetIdModel := targetIdentifierModel{
					TargetId:   types.StringValue(targetIdValue),
					TargetType: types.StringValue(string(target.TargetIdentifier.TargetType)),
				}

				targetIdentifierList, diags := fwtypes.NewListNestedObjectValueOfValueSlice(ctx, []targetIdentifierModel{targetIdModel})
				response.Diagnostics.Append(diags...)
				if response.Diagnostics.HasError() {
					return
				}
				targetModels[i].TargetIdentifier = targetIdentifierList
			}
		}

		targetsList, diags := fwtypes.NewListNestedObjectValueOfValueSlice(ctx, targetModels)
		response.Diagnostics.Append(diags...)
		if response.Diagnostics.HasError() {
			return
		}
		new.Targets = targetsList
	}

	// Set tags based on the updated AWS state
	if !new.Tags.IsNull() && len(output.Tags) > 0 {
		tags := tftags.New(ctx, output.Tags)
		new.Tags = tftags.FlattenStringValueMap(ctx, tags.Map())
	}
	// TagsAll should always reflect AWS state
	tags := tftags.New(ctx, output.Tags)
	new.TagsAll = tftags.FlattenStringValueMap(ctx, tags.Map())

	response.Diagnostics.Append(response.State.Set(ctx, &new)...)
}

func (r *scopeResource) calculateTargetChanges(ctx context.Context, oldTargets, newTargets fwtypes.ListNestedObjectValueOf[targetResourceModel]) ([]awstypes.TargetResource, []awstypes.TargetResource, diag.Diagnostics) {
	var diags diag.Diagnostics
	var resourcesToAdd, resourcesToDelete []awstypes.TargetResource

	// Convert old targets to map for easy lookup
	oldTargetsMap := make(map[string]awstypes.TargetResource)
	if !oldTargets.IsNull() && !oldTargets.IsUnknown() {
		oldTargetsList, d := oldTargets.ToSlice(ctx)
		diags.Append(d...)
		if diags.HasError() {
			return nil, nil, diags
		}

		for _, target := range oldTargetsList {
			awsTarget := awstypes.TargetResource{
				Region: target.Region.ValueStringPointer(),
			}

			// Handle union type for TargetIdentifier
			if !target.TargetIdentifier.IsNull() && !target.TargetIdentifier.IsUnknown() {
				identifiers, d := target.TargetIdentifier.ToSlice(ctx)
				diags.Append(d...)
				if diags.HasError() {
					return nil, nil, diags
				}

				if len(identifiers) > 0 {
					identifier := identifiers[0]
					awsTarget.TargetIdentifier = &awstypes.TargetIdentifier{
						TargetId: &awstypes.TargetIdMemberAccountId{
							Value: identifier.TargetId.ValueString(),
						},
						TargetType: awstypes.TargetType(identifier.TargetType.ValueString()),
					}

					// Create a key for the target (region + target_id + target_type)
					key := fmt.Sprintf("%s:%s:%s",
						target.Region.ValueString(),
						identifier.TargetId.ValueString(),
						identifier.TargetType.ValueString())
					oldTargetsMap[key] = awsTarget
				}
			}
		}
	}

	// Convert new targets to map and identify additions
	newTargetsMap := make(map[string]awstypes.TargetResource)
	if !newTargets.IsNull() && !newTargets.IsUnknown() {
		newTargetsList, d := newTargets.ToSlice(ctx)
		diags.Append(d...)
		if diags.HasError() {
			return nil, nil, diags
		}

		for _, target := range newTargetsList {
			awsTarget := awstypes.TargetResource{
				Region: target.Region.ValueStringPointer(),
			}

			// Handle union type for TargetIdentifier
			if !target.TargetIdentifier.IsNull() && !target.TargetIdentifier.IsUnknown() {
				identifiers, d := target.TargetIdentifier.ToSlice(ctx)
				diags.Append(d...)
				if diags.HasError() {
					return nil, nil, diags
				}

				if len(identifiers) > 0 {
					identifier := identifiers[0]
					awsTarget.TargetIdentifier = &awstypes.TargetIdentifier{
						TargetId: &awstypes.TargetIdMemberAccountId{
							Value: identifier.TargetId.ValueString(),
						},
						TargetType: awstypes.TargetType(identifier.TargetType.ValueString()),
					}

					// Create a key for the target
					key := fmt.Sprintf("%s:%s:%s",
						target.Region.ValueString(),
						identifier.TargetId.ValueString(),
						identifier.TargetType.ValueString())
					newTargetsMap[key] = awsTarget

					// If this target doesn't exist in old targets, it's an addition
					if _, exists := oldTargetsMap[key]; !exists {
						resourcesToAdd = append(resourcesToAdd, awsTarget)
					}
				}
			}
		}
	}

	// Identify deletions
	for key, target := range oldTargetsMap {
		if _, exists := newTargetsMap[key]; !exists {
			resourcesToDelete = append(resourcesToDelete, target)
		}
	}

	return resourcesToAdd, resourcesToDelete, diags
}

func (r *scopeResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data scopeResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().NetworkFlowMonitorClient(ctx)

	input := networkflowmonitor.DeleteScopeInput{
		ScopeId: data.ScopeId.ValueStringPointer(),
	}
	_, err := conn.DeleteScope(ctx, &input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("deleting Network Flow Monitor Scope (%s)", data.ID.ValueString()), err.Error())
		return
	}

	if _, err := waitScopeDeleted(ctx, conn, data.ScopeId.ValueString(), r.DeleteTimeout(ctx, data.Timeouts)); err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("waiting for Network Flow Monitor Scope (%s) delete", data.ID.ValueString()), err.Error())
		return
	}
}

func (r *scopeResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	// The import ID can be either an ARN or a scope ID
	id := request.ID

	// If it's an ARN, extract the scope ID
	if arn.IsARN(id) {
		parsedARN, err := arn.Parse(id)
		if err != nil {
			response.Diagnostics.AddError("Invalid ARN", fmt.Sprintf("Unable to parse ARN (%s): %s", id, err))
			return
		}

		if parsedARN.Service != "networkflowmonitor" {
			response.Diagnostics.AddError("Invalid ARN", fmt.Sprintf("Expected networkflowmonitor service ARN, got: %s", parsedARN.Service))
			return
		}

		// ARN format: arn:partition:networkflowmonitor:region:account:scope/scope-id
		parts := strings.Split(parsedARN.Resource, "/")
		if len(parts) != 2 || parts[0] != names.AttrScope {
			response.Diagnostics.AddError("Invalid import ID", fmt.Sprintf("Expected ARN format 'arn:partition:networkflowmonitor:region:account:scope/scope-id', got: %s", id))
			return
		}
		scopeId := parts[1]

		// Set both ID (ARN) and scope_id
		response.Diagnostics.Append(response.State.SetAttribute(ctx, path.Root(names.AttrID), id)...)
		response.Diagnostics.Append(response.State.SetAttribute(ctx, path.Root("scope_id"), scopeId)...)
	} else {
		// Assume it's a scope ID, we'll need to construct the ARN during read
		response.Diagnostics.Append(response.State.SetAttribute(ctx, path.Root("scope_id"), id)...)
		// ID will be set during the subsequent Read operation
	}
}

func findScopeByID(ctx context.Context, conn *networkflowmonitor.Client, id string) (*networkflowmonitor.GetScopeOutput, error) {
	input := networkflowmonitor.GetScopeInput{
		ScopeId: aws.String(id),
	}

	output, err := conn.GetScope(ctx, &input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output, nil
}

func statusScope(ctx context.Context, conn *networkflowmonitor.Client, id string) retry.StateRefreshFunc {
	return func() (any, string, error) {
		output, err := findScopeByID(ctx, conn, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.Status), nil
	}
}

func waitScopeCreated(ctx context.Context, conn *networkflowmonitor.Client, id string, timeout time.Duration) (*networkflowmonitor.GetScopeOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.ScopeStatusInProgress),
		Target:  enum.Slice(awstypes.ScopeStatusSucceeded),
		Refresh: statusScope(ctx, conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*networkflowmonitor.GetScopeOutput); ok {
		return output, err
	}

	return nil, err
}

func waitScopeUpdated(ctx context.Context, conn *networkflowmonitor.Client, id string, timeout time.Duration) (*networkflowmonitor.GetScopeOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.ScopeStatusInProgress),
		Target:  enum.Slice(awstypes.ScopeStatusSucceeded),
		Refresh: statusScope(ctx, conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*networkflowmonitor.GetScopeOutput); ok {
		return output, err
	}

	return nil, err
}

func waitScopeDeleted(ctx context.Context, conn *networkflowmonitor.Client, id string, timeout time.Duration) (*networkflowmonitor.GetScopeOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.ScopeStatusDeactivating),
		Target:  []string{},
		Refresh: statusScope(ctx, conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*networkflowmonitor.GetScopeOutput); ok {
		return output, err
	}

	return nil, err
}
