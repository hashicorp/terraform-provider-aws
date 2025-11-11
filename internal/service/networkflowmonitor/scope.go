// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package networkflowmonitor

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/networkflowmonitor"
	awstypes "github.com/aws/aws-sdk-go-v2/service/networkflowmonitor/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	sdkid "github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	fwvalidators "github.com/hashicorp/terraform-provider-aws/internal/framework/validators"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_networkflowmonitor_scope", name="Scope")
// @Tags(identifierAttribute="scope_arn")
func newScopeResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &scopeResource{}

	r.SetDefaultCreateTimeout(30 * time.Minute)
	r.SetDefaultUpdateTimeout(30 * time.Minute)
	r.SetDefaultDeleteTimeout(30 * time.Minute)

	return r, nil
}

type scopeResource struct {
	framework.ResourceWithModel[scopeResourceModel]
	framework.WithTimeouts
}

func (r *scopeResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"scope_arn":       framework.ARNAttributeComputedOnly(),
			"scope_id":        framework.IDAttribute(),
			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
		},
		Blocks: map[string]schema.Block{
			"target": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[targetResourceModel](ctx),
				Validators: []validator.List{
					listvalidator.SizeAtLeast(1),
					listvalidator.IsRequired(),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						names.AttrRegion: schema.StringAttribute{
							Required: true,
							Validators: []validator.String{
								fwvalidators.AWSRegion(),
							},
						},
					},
					Blocks: map[string]schema.Block{
						"target_identifier": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[targetIdentifierModel](ctx),
							Validators: []validator.List{
								listvalidator.SizeAtLeast(1),
								listvalidator.SizeAtMost(1),
								listvalidator.IsRequired(),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"target_type": schema.StringAttribute{
										CustomType: fwtypes.StringEnumType[awstypes.TargetType](),
										Required:   true,
									},
								},
								Blocks: map[string]schema.Block{
									"target_id": schema.ListNestedBlock{
										CustomType: fwtypes.NewListNestedObjectTypeOf[targetIdModel](ctx),
										Validators: []validator.List{
											listvalidator.SizeAtLeast(1),
											listvalidator.SizeAtMost(1),
											listvalidator.IsRequired(),
										},
										NestedObject: schema.NestedBlockObject{
											Attributes: map[string]schema.Attribute{
												"account_id": schema.StringAttribute{
													Required: true,
													Validators: []validator.String{
														fwvalidators.AWSAccountID(),
													},
												},
											},
										},
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

func (r *scopeResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data scopeResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().NetworkFlowMonitorClient(ctx)

	var input networkflowmonitor.CreateScopeInput
	response.Diagnostics.Append(fwflex.Expand(ctx, data, &input)...)
	if response.Diagnostics.HasError() {
		return
	}

	// Additional fields.
	input.ClientToken = aws.String(sdkid.UniqueId())
	input.Tags = getTagsIn(ctx)

	output, err := conn.CreateScope(ctx, &input)

	if err != nil {
		response.Diagnostics.AddError("creating Network Flow Monitor Scope", err.Error())
		return
	}

	// Set values for unknowns.
	data.ScopeARN = fwflex.StringToFramework(ctx, output.ScopeArn)
	data.ScopeID = fwflex.StringToFramework(ctx, output.ScopeId)

	scopeID := fwflex.StringValueFromFramework(ctx, data.ScopeID)
	if _, err := waitScopeCreated(ctx, conn, scopeID, r.CreateTimeout(ctx, data.Timeouts)); err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("waiting for Network Flow Monitor Scope (%s) create", scopeID), err.Error())
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, data)...)
}

func (r *scopeResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data scopeResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().NetworkFlowMonitorClient(ctx)

	scopeID := fwflex.StringValueFromFramework(ctx, data.ScopeID)
	output, err := findScopeByID(ctx, conn, scopeID)

	if tfresource.NotFound(err) {
		response.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)
		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading Network Flow Monitor Scope (%s)", scopeID), err.Error())
		return
	}

	response.Diagnostics.Append(fwflex.Flatten(ctx, output, &data)...)
	if response.Diagnostics.HasError() {
		return
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

	//conn := r.Meta().NetworkFlowMonitorClient(ctx)

	// Handle targets updates
	// if !new.Targets.Equal(old.Targets) {
	// 	// Calculate targets to add and remove
	// 	resourcesToAdd, resourcesToDelete, diags := r.calculateTargetChanges(ctx, old.Targets, new.Targets)
	// 	response.Diagnostics.Append(diags...)
	// 	if response.Diagnostics.HasError() {
	// 		return
	// 	}

	// 	// Update scope targets if there are changes
	// 	if len(resourcesToAdd) > 0 || len(resourcesToDelete) > 0 {
	// 		input := networkflowmonitor.UpdateScopeInput{
	// 			ScopeId:           old.ScopeID.ValueStringPointer(),
	// 			ResourcesToAdd:    resourcesToAdd,
	// 			ResourcesToDelete: resourcesToDelete,
	// 		}

	// 		_, err := conn.UpdateScope(ctx, &input)
	// 		if err != nil {
	// 			response.Diagnostics.AddError(fmt.Sprintf("updating Network Flow Monitor Scope (%s) targets", new.ID.ValueString()), err.Error())
	// 			return
	// 		}

	// 		// Wait for scope to be updated
	// 		_, err = waitScopeUpdated(ctx, conn, old.ScopeID.ValueString(), r.UpdateTimeout(ctx, new.Timeouts))
	// 		if err != nil {
	// 			response.Diagnostics.AddError(fmt.Sprintf("waiting for Network Flow Monitor Scope (%s) update", new.ID.ValueString()), err.Error())
	// 			return
	// 		}
	// 	}
	// }

	response.Diagnostics.Append(response.State.Set(ctx, &new)...)
}

/*
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
*/

func (r *scopeResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data scopeResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().NetworkFlowMonitorClient(ctx)

	scopeID := fwflex.StringValueFromFramework(ctx, data.ScopeID)
	input := networkflowmonitor.DeleteScopeInput{
		ScopeId: aws.String(scopeID),
	}
	_, err := conn.DeleteScope(ctx, &input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("deleting Network Flow Monitor Scope (%s)", scopeID), err.Error())
		return
	}

	if _, err := waitScopeDeleted(ctx, conn, scopeID, r.DeleteTimeout(ctx, data.Timeouts)); err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("waiting for Network Flow Monitor Scope (%s) delete", scopeID), err.Error())
		return
	}
}

func (r *scopeResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	response.Diagnostics.Append(response.State.SetAttribute(ctx, path.Root("scope_id"), request.ID)...)
}

func findScopeByID(ctx context.Context, conn *networkflowmonitor.Client, id string) (*networkflowmonitor.GetScopeOutput, error) {
	input := networkflowmonitor.GetScopeInput{
		ScopeId: aws.String(id),
	}

	return findScope(ctx, conn, &input)
}

func findScope(ctx context.Context, conn *networkflowmonitor.Client, input *networkflowmonitor.GetScopeInput) (*networkflowmonitor.GetScopeOutput, error) {
	output, err := conn.GetScope(ctx, input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
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

type scopeResourceModel struct {
	framework.WithRegionModel
	ScopeARN types.String                                         `tfsdk:"scope_arn"`
	ScopeID  types.String                                         `tfsdk:"scope_id"`
	Tags     tftags.Map                                           `tfsdk:"tags"`
	TagsAll  tftags.Map                                           `tfsdk:"tags_all"`
	Targets  fwtypes.ListNestedObjectValueOf[targetResourceModel] `tfsdk:"target"`
	Timeouts timeouts.Value                                       `tfsdk:"timeouts"`
}

type targetResourceModel struct {
	Region           types.String                                           `tfsdk:"region"`
	TargetIdentifier fwtypes.ListNestedObjectValueOf[targetIdentifierModel] `tfsdk:"target_identifier"`
}

type targetIdentifierModel struct {
	TargetID   fwtypes.ListNestedObjectValueOf[targetIdModel] `tfsdk:"target_id"`
	TargetType fwtypes.StringEnum[awstypes.TargetType]        `tfsdk:"target_type"`
}

type targetIdModel struct {
	AccountID types.String `tfsdk:"account_id"`
}

var (
	_ fwflex.Expander  = targetIdModel{}
	_ fwflex.Flattener = &targetIdModel{}
)

func (m targetIdModel) Expand(ctx context.Context) (any, diag.Diagnostics) {
	var diags diag.Diagnostics
	switch {
	case !m.AccountID.IsNull():
		var r awstypes.TargetIdMemberAccountId
		r.Value = fwflex.StringValueFromFramework(ctx, m.AccountID)
		return &r, diags
	}
	return nil, diags
}

func (m *targetIdModel) Flatten(ctx context.Context, v any) diag.Diagnostics {
	var diags diag.Diagnostics
	switch t := v.(type) {
	case awstypes.TargetIdMemberAccountId:
		m.AccountID = fwflex.StringValueToFramework(ctx, t.Value)
	default:
		diags.AddError(
			"Unsupported Type",
			fmt.Sprintf("target ID flatten: %T", v),
		)
	}
	return diags
}
