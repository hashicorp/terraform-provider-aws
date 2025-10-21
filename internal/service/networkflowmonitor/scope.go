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
	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
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
	framework.WithImportByID
	framework.WithTimeouts
}

func (r *scopeResource) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = request.ProviderTypeName + "_networkflowmonitor_scope"
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

	// MANDATORY: Use fwflex.Expand for automatic parameter mapping
	input := &networkflowmonitor.CreateScopeInput{}
	response.Diagnostics.Append(fwflex.Expand(ctx, data, input)...)
	if response.Diagnostics.HasError() {
		return
	}

	// Handle union types that fwflex can't handle automatically
	if !data.Targets.IsNull() && !data.Targets.IsUnknown() {
		targets, diags := data.Targets.ToSlice(ctx)
		response.Diagnostics.Append(diags...)
		if response.Diagnostics.HasError() {
			return
		}

		input.Targets = make([]awstypes.TargetResource, len(targets))
		for i, target := range targets {
			input.Targets[i].Region = aws.String(target.Region.ValueString())

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

	output, err := conn.CreateScope(ctx, input)
	if err != nil {
		response.Diagnostics.AddError("creating Network Flow Monitor Scope", err.Error())
		return
	}

	// Set ID and computed attributes
	data.ID = types.StringValue(aws.ToString(output.ScopeArn))
	data.ARN = types.StringValue(aws.ToString(output.ScopeArn))

	// Wait for scope to be ready
	scope, err := waitScopeCreated(ctx, conn, aws.ToString(output.ScopeId), r.CreateTimeout(ctx, data.Timeouts))
	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("waiting for Network Flow Monitor Scope (%s) create", data.ID.ValueString()), err.Error())
		return
	}

	// Set computed attributes
	data.ScopeId = types.StringValue(aws.ToString(scope.ScopeId))
	data.Status = types.StringValue(string(scope.Status))

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

	// MANDATORY: Use fwflex.Flatten for automatic parameter mapping
	response.Diagnostics.Append(fwflex.Flatten(ctx, output, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	// Handle union types that fwflex can't handle automatically
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

	// Set tags from API response
	setTagsOut(ctx, output.Tags)

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

	// Handle tag updates
	if !new.Tags.Equal(old.Tags) {
		if err := updateTags(ctx, conn, new.ID.ValueString(), old.Tags, new.Tags); err != nil {
			response.Diagnostics.AddError(fmt.Sprintf("updating Network Flow Monitor Scope (%s) tags", new.ID.ValueString()), err.Error())
			return
		}
	}

	response.Diagnostics.Append(response.State.Set(ctx, &new)...)
}

func (r *scopeResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data scopeResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().NetworkFlowMonitorClient(ctx)

	_, err := conn.DeleteScope(ctx, &networkflowmonitor.DeleteScopeInput{
		ScopeId: aws.String(data.ScopeId.ValueString()),
	})

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

func findScopeByID(ctx context.Context, conn *networkflowmonitor.Client, id string) (*networkflowmonitor.GetScopeOutput, error) {
	input := &networkflowmonitor.GetScopeInput{
		ScopeId: aws.String(id),
	}

	output, err := conn.GetScope(ctx, input)

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
	return func() (interface{}, string, error) {
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
