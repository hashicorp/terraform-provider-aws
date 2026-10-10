// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

package ec2

import (
	"context"
	"errors"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/hashicorp/aws-sdk-go-base/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/mapvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_ec2_instance_event_window_association", name="Instance Event Window Association")
// @IdentityAttribute("id")
// @Testing(hasNoPreExistingResource=true)
func newInstanceEventWindowAssociationResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	return &instanceEventWindowAssociationResource{}, nil
}

const (
	ResNameInstanceEventWindowAssociation = "Instance Event Window Association"
)

type instanceEventWindowAssociationResource struct {
	framework.ResourceWithModel[instanceEventWindowAssociationResourceModel]
	framework.WithImportByIdentity
}

// association_target's three fields are mutually exclusive: AWS allows
// associating an event window with exactly one of a set of instance IDs, a
// set of instance tags, or a set of Dedicated Host IDs. A given target
// (instance ID, Dedicated Host ID, or instance tag) can only be associated
// with one event window at a time. AWS enforces both constraints server-side.
// Limits: up to 100 instance IDs, 50 Dedicated Host IDs, or 50 instance tags
// per event window. See
// https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/event-windows.html.
func (r *instanceEventWindowAssociationResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrID: framework.IDAttribute(),
			"instance_event_window_id": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
		Blocks: map[string]schema.Block{
			"association_target": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[instanceEventWindowAssociationTargetModel](ctx),
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
					listvalidator.SizeAtLeast(1),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"dedicated_host_ids": schema.ListAttribute{
							CustomType:  fwtypes.ListOfStringType,
							ElementType: types.StringType,
							Optional:    true,
							Validators: []validator.List{
								listvalidator.ExactlyOneOf(
									path.MatchRelative().AtParent().AtName("instance_ids"),
									path.MatchRelative().AtParent().AtName("instance_tags"),
								),
							},
						},
						"instance_ids": schema.ListAttribute{
							CustomType:  fwtypes.ListOfStringType,
							ElementType: types.StringType,
							Optional:    true,
							Validators: []validator.List{
								listvalidator.ExactlyOneOf(
									path.MatchRelative().AtParent().AtName("dedicated_host_ids"),
									path.MatchRelative().AtParent().AtName("instance_tags"),
								),
							},
						},
						"instance_tags": schema.MapAttribute{
							CustomType:  fwtypes.MapOfStringType,
							ElementType: types.StringType,
							Optional:    true,
							Validators: []validator.Map{
								mapvalidator.ExactlyOneOf(
									path.MatchRelative().AtParent().AtName("dedicated_host_ids"),
									path.MatchRelative().AtParent().AtName("instance_ids"),
								),
							},
						},
					},
				},
			},
		},
	}
}

func (r *instanceEventWindowAssociationResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := r.Meta().EC2Client(ctx)

	var plan instanceEventWindowAssociationResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Plan.Get(ctx, &plan))
	if resp.Diagnostics.HasError() {
		return
	}

	var target awstypes.InstanceEventWindowAssociationRequest
	smerr.AddEnrich(ctx, &resp.Diagnostics, fwflex.Expand(ctx, plan.AssociationTarget, &target, fwflex.WithIgnoredFieldNamesAppend("InstanceTags")))
	if resp.Diagnostics.HasError() {
		return
	}

	targetModel, d := plan.AssociationTarget.ToPtr(ctx)
	resp.Diagnostics.Append(d...)
	if resp.Diagnostics.HasError() {
		return
	}
	if targetModel != nil {
		target.InstanceTags = instanceEventWindowAssociationTagsFromMap(ctx, targetModel.InstanceTags)
	}

	input := ec2.AssociateInstanceEventWindowInput{
		InstanceEventWindowId: plan.InstanceEventWindowID.ValueStringPointer(),
		AssociationTarget:     &target,
	}

	out, err := conn.AssociateInstanceEventWindow(ctx, &input)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, plan.InstanceEventWindowID.String())
		return
	}
	if out == nil || out.InstanceEventWindow == nil {
		smerr.AddError(ctx, &resp.Diagnostics, errors.New("empty output"), smerr.ID, plan.InstanceEventWindowID.String())
		return
	}

	plan.ID = plan.InstanceEventWindowID

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, plan))
}

func (r *instanceEventWindowAssociationResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().EC2Client(ctx)

	var state instanceEventWindowAssociationResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := findInstanceEventWindowByID(ctx, conn, state.ID.ValueString())
	if retry.NotFound(err) {
		resp.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, state.ID.String())
		return
	}

	// The association target has been disassociated (e.g. via drift) if all
	// three underlying target fields are empty.
	target := out.AssociationTarget
	if target == nil || (len(target.InstanceIds) == 0 && len(target.DedicatedHostIds) == 0 && len(target.Tags) == 0) {
		resp.State.RemoveResource(ctx)
		return
	}

	state.InstanceEventWindowID = types.StringPointerValue(out.InstanceEventWindowId)

	var targetModel instanceEventWindowAssociationTargetModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, fwflex.Flatten(ctx, target, &targetModel))
	if resp.Diagnostics.HasError() {
		return
	}

	// Only one target type is ever populated by AWS; the other two report
	// as empty slices. Represent an unused target type as null (matching an
	// omitted attribute in config) rather than an empty list, to avoid
	// spurious drift between plan and refresh.
	if len(target.InstanceIds) == 0 {
		targetModel.InstanceIDs = fwtypes.NewListValueOfNull[types.String](ctx)
	}
	if len(target.DedicatedHostIds) == 0 {
		targetModel.DedicatedHostIDs = fwtypes.NewListValueOfNull[types.String](ctx)
	}

	// AutoFlex cannot populate this: the describe API names the field "Tags"
	// (not "InstanceTags", as the Associate/Disassociate request types do) and
	// "Tags" is also in AutoFlex's default-ignored-field list. Build the map
	// manually from the same source field the removed-target check above uses.
	tagsMap, d := instanceEventWindowAssociationTagsToMap(ctx, target.Tags)
	smerr.AddEnrich(ctx, &resp.Diagnostics, d)
	if resp.Diagnostics.HasError() {
		return
	}
	targetModel.InstanceTags = tagsMap

	list, d := fwtypes.NewListNestedObjectValueOfPtr(ctx, &targetModel)
	smerr.AddEnrich(ctx, &resp.Diagnostics, d)
	if resp.Diagnostics.HasError() {
		return
	}
	state.AssociationTarget = list

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &state))
}

func (r *instanceEventWindowAssociationResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := r.Meta().EC2Client(ctx)

	var state instanceEventWindowAssociationResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	var target awstypes.InstanceEventWindowDisassociationRequest
	smerr.AddEnrich(ctx, &resp.Diagnostics, fwflex.Expand(ctx, state.AssociationTarget, &target, fwflex.WithIgnoredFieldNamesAppend("InstanceTags")))
	if resp.Diagnostics.HasError() {
		return
	}

	targetModel, d := state.AssociationTarget.ToPtr(ctx)
	resp.Diagnostics.Append(d...)
	if resp.Diagnostics.HasError() {
		return
	}
	if targetModel != nil {
		target.InstanceTags = instanceEventWindowAssociationTagsFromMap(ctx, targetModel.InstanceTags)
	}

	input := ec2.DisassociateInstanceEventWindowInput{
		InstanceEventWindowId: state.InstanceEventWindowID.ValueStringPointer(),
		AssociationTarget:     &target,
	}

	_, err := conn.DisassociateInstanceEventWindow(ctx, &input)
	if err != nil {
		if tfawserr.ErrCodeEquals(err, errCodeInvalidInstanceEventWindowIDNotFound, errCodeInvalidInstanceIDNotFound) {
			return
		}

		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, state.ID.String())
		return
	}
}

type instanceEventWindowAssociationResourceModel struct {
	framework.WithRegionModel
	AssociationTarget     fwtypes.ListNestedObjectValueOf[instanceEventWindowAssociationTargetModel] `tfsdk:"association_target"`
	ID                    types.String                                                               `tfsdk:"id"`
	InstanceEventWindowID types.String                                                               `tfsdk:"instance_event_window_id"`
}

type instanceEventWindowAssociationTargetModel struct {
	DedicatedHostIDs fwtypes.ListValueOf[types.String] `tfsdk:"dedicated_host_ids"`
	InstanceIDs      fwtypes.ListValueOf[types.String] `tfsdk:"instance_ids"`
	InstanceTags     fwtypes.MapValueOf[types.String]  `tfsdk:"instance_tags"`
}

// instanceEventWindowAssociationTagsFromMap converts the association_target's
// instance_tags map to []awstypes.Tag. AutoFlex cannot perform this
// map-to-slice conversion, so it is done manually using the same tag
// marshaling helper (svcTags) the tagging interceptor uses.
func instanceEventWindowAssociationTagsFromMap(ctx context.Context, m fwtypes.MapValueOf[types.String]) []awstypes.Tag {
	if m.IsNull() || m.IsUnknown() {
		return nil
	}

	elements := make(map[string]any, m.Length(fwtypes.CollectionLengthUnhandledAsZero))
	for k, v := range m.Elements() {
		if s, ok := v.(types.String); ok {
			elements[k] = s.ValueString()
		}
	}

	return svcTags(tftags.New(ctx, elements))
}

// instanceEventWindowAssociationTagsToMap converts []awstypes.Tag (the
// describe API's AssociationTarget.Tags field) to the association_target's
// instance_tags map. AutoFlex cannot perform this conversion because the
// source field is named "Tags", not "InstanceTags", and "Tags" is in
// AutoFlex's default-ignored-field list.
func instanceEventWindowAssociationTagsToMap(ctx context.Context, tags []awstypes.Tag) (fwtypes.MapValueOf[types.String], diag.Diagnostics) {
	if len(tags) == 0 {
		return fwtypes.NewMapValueOfNull[types.String](ctx), nil
	}

	elements := make(map[string]attr.Value, len(tags))
	for k, v := range keyValueTags(ctx, tags).Map() {
		elements[k] = types.StringValue(v)
	}

	return fwtypes.NewMapValueOf[types.String](ctx, elements)
}
