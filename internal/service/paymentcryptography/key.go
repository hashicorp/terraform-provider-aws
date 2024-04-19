// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package paymentcryptography

import (
	"context"
	"errors"
	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/paymentcryptography"
	awstypes "github.com/aws/aws-sdk-go-v2/service/paymentcryptography/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// Function annotations are used for resource registration to the Provider. DO NOT EDIT.
// @FrameworkResource("aws_paymentcryptography_key", name="Key")
// @Tags(identifierAttribute="arn")
func newResourceKey(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &resourceKey{}

	r.SetDefaultCreateTimeout(30 * time.Minute)
	r.SetDefaultUpdateTimeout(30 * time.Minute)
	r.SetDefaultDeleteTimeout(30 * time.Minute)

	return r, nil
}

const (
	ResNameKey = "Key"
)

type resourceKey struct {
	framework.ResourceWithConfigure
	framework.WithTimeouts
}

func (r *resourceKey) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = "aws_paymentcryptography_key"
}

func (r *resourceKey) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"arn": framework.ARNAttributeComputedOnly(),
			"id":  framework.IDAttribute(),
			"create_timestamp": schema.StringAttribute{
				Computed:   true,
				CustomType: timetypes.RFC3339Type{},
			},
			"deletion_window_in_days": schema.Int64Attribute{
				Validators: []validator.Int64{
					int64validator.Between(3, 180),
				},
			},
			"delete_pending_timestamp": schema.StringAttribute{
				Computed:   true,
				CustomType: timetypes.RFC3339Type{},
			},
			"delete_timestamp": schema.StringAttribute{
				Computed:   true,
				CustomType: timetypes.RFC3339Type{},
			},
			"enabled": schema.BoolAttribute{
				Optional: true,
				Default:  booldefault.StaticBool(true),
			},
			"exportable": schema.BoolAttribute{
				Required: true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.RequiresReplace(),
				},
			},
			"key_check_value": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"key_check_value_algorithm": schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.KeyCheckValueAlgorithm](),
				Optional:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"key_origin": schema.StringAttribute{
				Computed:   true,
				CustomType: fwtypes.StringEnumType[awstypes.KeyOrigin](),
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"key_state": schema.StringAttribute{
				Computed:   true,
				CustomType: fwtypes.StringEnumType[awstypes.KeyState](),
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
			"usage_start_timestamp": schema.StringAttribute{
				Computed:   true,
				CustomType: timetypes.RFC3339Type{},
			},
			"usage_stop_timestamp": schema.StringAttribute{
				Computed:   true,
				CustomType: timetypes.RFC3339Type{},
			},
		},
		Blocks: map[string]schema.Block{
			"key_attributes": schema.SingleNestedBlock{
				CustomType: fwtypes.NewObjectTypeOf[keyAttributesModel](ctx),
				PlanModifiers: []planmodifier.Object{
					objectplanmodifier.RequiresReplace(),
				},
				Attributes: map[string]schema.Attribute{
					"key_algorithm": schema.StringAttribute{
						Required:   true,
						CustomType: fwtypes.StringEnumType[awstypes.KeyAlgorithm](),
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.RequiresReplace(),
						},
					},
					"key_class": schema.StringAttribute{
						Required:   true,
						CustomType: fwtypes.StringEnumType[awstypes.KeyClass](),
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.RequiresReplace(),
						},
					},
					"key_usage": schema.StringAttribute{
						Required:   true,
						CustomType: fwtypes.StringEnumType[awstypes.KeyUsage](),
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.RequiresReplace(),
						},
					},
				},
				Blocks: map[string]schema.Block{
					"key_modes_of_use": schema.SingleNestedBlock{
						CustomType: fwtypes.NewObjectTypeOf[keyModesOfUseModel](ctx),
						PlanModifiers: []planmodifier.Object{
							objectplanmodifier.RequiresReplace(),
						},
						Attributes: map[string]schema.Attribute{
							"decrypt": schema.BoolAttribute{
								Optional: true,
								Default:  booldefault.StaticBool(false),
							},
							"derive_key": schema.BoolAttribute{
								Optional: true,
								Default:  booldefault.StaticBool(false),
							},
							"encrypt": schema.BoolAttribute{
								Optional: true,
								Default:  booldefault.StaticBool(false),
							},
							"generate": schema.BoolAttribute{
								Optional: true,
								Default:  booldefault.StaticBool(false),
							},
							"no_restrictions": schema.BoolAttribute{
								Optional: true,
								Default:  booldefault.StaticBool(false),
							},
							"sign": schema.BoolAttribute{
								Optional: true,
								Default:  booldefault.StaticBool(false),
							},
							"unwrap": schema.BoolAttribute{
								Optional: true,
								Default:  booldefault.StaticBool(false),
							},
							"verify": schema.BoolAttribute{
								Optional: true,
								Default:  booldefault.StaticBool(false),
							},
							"wrap": schema.BoolAttribute{
								Optional: true,
								Default:  booldefault.StaticBool(false),
							},
						},
					},
				},
			},
			"timeouts": timeouts.Block(ctx, timeouts.Opts{
				Create: true,
				Update: true,
				Delete: true,
			}),
		},
	}
}

func (r *resourceKey) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	conn := r.Meta().PaymentCryptographyClient(ctx)

	var plan resourceKeyData
	response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}

	in := &paymentcryptography.CreateKeyInput{}
	response.Diagnostics.Append(flex.Expand(ctx, plan, in)...)

	if response.Diagnostics.HasError() {
		return
	}

	in.Tags = getTagsIn(ctx)

	out, err := conn.CreateKey(ctx, in)
	if err != nil {
		response.Diagnostics.AddError(
			create.ProblemStandardMessage(names.PaymentCryptography, create.ErrActionCreating, ResNameKey, "FIXME", err),
			err.Error(),
		)
		return
	}
	if out == nil || out.Key == nil {
		response.Diagnostics.AddError(
			create.ProblemStandardMessage(names.PaymentCryptography, create.ErrActionCreating, ResNameKey, "FIXME", nil),
			errors.New("empty output").Error(),
		)
		return
	}

	plan.KeyArn = flex.StringToFramework(ctx, out.Key.KeyArn)
	plan.setId()

	createTimeout := r.CreateTimeout(ctx, plan.Timeouts)
	created, err := waitKeyCreated(ctx, conn, plan.ID.ValueString(), createTimeout)
	if err != nil {
		response.Diagnostics.AddError(
			create.ProblemStandardMessage(names.PaymentCryptography, create.ErrActionWaitingForCreation, ResNameKey, plan.KeyArn.String(), err),
			err.Error(),
		)
		return
	}

	response.Diagnostics.Append(flex.Flatten(ctx, created, plan)...)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, plan)...)
}

func (r *resourceKey) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	conn := r.Meta().PaymentCryptographyClient(ctx)

	var state resourceKeyData
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	out, err := findKeyByID(ctx, conn, state.ID.ValueString())
	if tfresource.NotFound(err) {
		response.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		response.Diagnostics.AddError(
			create.ProblemStandardMessage(names.PaymentCryptography, create.ErrActionSetting, ResNameKey, state.ID.String(), err),
			err.Error(),
		)
		return
	}

	response.Diagnostics.Append(flex.Flatten(ctx, out, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *resourceKey) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var old, new resourceKeyData
	response.Diagnostics.Append(request.Plan.Get(ctx, &new)...)
	if response.Diagnostics.HasError() {
		return
	}
	response.Diagnostics.Append(request.State.Get(ctx, &old)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().PaymentCryptographyClient(ctx)

	if !old.Enabled.Equal(new.Enabled) {
		if new.Enabled.ValueBool() {
			in := &paymentcryptography.StartKeyUsageInput{
				KeyIdentifier: flex.StringFromFramework(ctx, new.ID),
			}
			_, err := conn.StartKeyUsage(ctx, in)
			if err != nil {
				response.Diagnostics.AddError(
					create.ProblemStandardMessage(names.PaymentCryptography, create.ErrActionUpdating, ResNameKey, new.KeyArn.String(), err),
					err.Error(),
				)
				return
			}
		} else {
			in := &paymentcryptography.StopKeyUsageInput{
				KeyIdentifier: flex.StringFromFramework(ctx, new.ID),
			}
			_, err := conn.StopKeyUsage(ctx, in)
			if err != nil {
				response.Diagnostics.AddError(
					create.ProblemStandardMessage(names.PaymentCryptography, create.ErrActionUpdating, ResNameKey, new.KeyArn.String(), err),
					err.Error(),
				)
				return
			}

		}
	}

	response.Diagnostics.Append(response.State.Set(ctx, &new)...)
}

func (r *resourceKey) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := r.Meta().PaymentCryptographyClient(ctx)

	var state resourceKeyData
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	in := &paymentcryptography.DeleteKeyInput{
		KeyIdentifier: aws.String(state.ID.ValueString()),
	}

	_, err := conn.DeleteKey(ctx, in)
	if err != nil {
		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return
		}
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.PaymentCryptography, create.ErrActionDeleting, ResNameKey, state.ID.String(), err),
			err.Error(),
		)
		return
	}

	deleteTimeout := r.DeleteTimeout(ctx, state.Timeouts)
	_, err = waitKeyDeleted(ctx, conn, state.ID.ValueString(), deleteTimeout)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.PaymentCryptography, create.ErrActionWaitingForDeletion, ResNameKey, state.ID.String(), err),
			err.Error(),
		)
		return
	}
}

func (r *resourceKey) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
func (r *resourceKey) ModifyPlan(ctx context.Context, request resource.ModifyPlanRequest, response *resource.ModifyPlanResponse) {
	r.SetTagsAll(ctx, request, response)
}

const (
	statusChangePending = "Pending"
	statusDeleting      = "Deleting"
	statusNormal        = "Normal"
	statusUpdated       = "Updated"
)

func waitKeyCreated(ctx context.Context, conn *paymentcryptography.Client, id string, timeout time.Duration) (*awstypes.Key, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   enum.Slice(awstypes.KeyStateCreateInProgress),
		Target:                    enum.Slice(awstypes.KeyStateCreateComplete),
		Refresh:                   statusKey(ctx, conn, id),
		Timeout:                   timeout,
		NotFoundChecks:            20,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*awstypes.Key); ok {
		return out, err
	}

	return nil, err
}

func waitKeyDeleted(ctx context.Context, conn *paymentcryptography.Client, id string, timeout time.Duration) (*awstypes.Key, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.KeyStateCreateComplete),
		Target:  enum.Slice(awstypes.KeyStateDeletePending, awstypes.KeyStateDeleteComplete),
		Refresh: statusKey(ctx, conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*awstypes.Key); ok {
		return out, err
	}

	return nil, err
}

func statusKey(ctx context.Context, conn *paymentcryptography.Client, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		out, err := findKeyByID(ctx, conn, id)
		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return out, string(out.KeyState), nil
	}
}

func findKeyByID(ctx context.Context, conn *paymentcryptography.Client, id string) (*awstypes.Key, error) {
	in := &paymentcryptography.GetKeyInput{
		KeyIdentifier: aws.String(id),
	}

	out, err := conn.GetKey(ctx, in)
	if err != nil {
		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: in,
			}
		}

		return nil, err
	}

	if out == nil || out.Key == nil {
		return nil, tfresource.NewEmptyResultError(in)
	}

	// If the key is either Pending or Complete deletion it's logically deleted
	if out.Key.KeyState == awstypes.KeyStateDeletePending || out.Key.KeyState == awstypes.KeyStateDeleteComplete {
		return nil, tfresource.NewEmptyResultError(in)
	}

	return out.Key, nil
}

type resourceKeyData struct {
	KeyArn                 types.String                                        `tfsdk:"arn"`
	CreateTimestamp        timetypes.RFC3339Type                               `tfsdk:"create_timestamp"`
	DeletionWindowInDate   types.Int64                                         `tfsdk:"deletion_window_in_days"`
	Enabled                types.Bool                                          `tfsdk:"enabled"`
	Exportable             types.Bool                                          `tfsdk:"exportable"`
	ID                     types.String                                        `tfsdk:"id"`
	KeyAttributes          fwtypes.ListNestedObjectValueOf[keyAttributesModel] `tfsdk:"key_attributes"`
	KeyCheckValue          types.String                                        `tfsdk:"key_check_value"`
	KeyCheckValueAlgorithm fwtypes.StringEnum[awstypes.KeyCheckValueAlgorithm] `tfsdk:"key_check_value_algorithm"`
	KeyOrigin              fwtypes.StringEnum[awstypes.KeyOrigin]              `tfsdk:"key_origin"`
	KeyState               fwtypes.StringEnum[awstypes.KeyState]               `tfsdk:"key_state"`
	Tags                   types.Map                                           `tfsdk:"tags"`
	TagsAll                types.Map                                           `tfsdk:"tags_all"`
	Timeouts               timeouts.Value                                      `tfsdk:"timeouts"`
	UsageStartTimestamp    timetypes.RFC3339Type                               `tfsdk:"usage_start_timestamp"`
	UsageStopTimestamp     timetypes.RFC3339Type                               `tfsdk:"usage_stop_timestamp"`
}

func (k *resourceKeyData) setId() {
	k.ID = k.KeyArn
}

type keyAttributesModel struct {
	KeyAlgorithm  fwtypes.StringEnum[awstypes.KeyAlgorithm]           `tfsdk:"key_algorithm"`
	KeyClass      fwtypes.StringEnum[awstypes.KeyClass]               `tfsdk:"key_class"`
	KeyModesOfUse fwtypes.ListNestedObjectValueOf[keyModesOfUseModel] `tfsdk:"key_modes_of_use"`
	KeyUsage      fwtypes.StringEnum[awstypes.KeyUsage]               `tfsdk:"key_usage"`
}

type keyModesOfUseModel struct {
	Decrypt        types.Bool `tfsdk:"decrypt"`
	DeriveKey      types.Bool `tfsdk:"derive_key"`
	Encrypt        types.Bool `tfsdk:"encrypt"`
	Generate       types.Bool `tfsdk:"generate"`
	NoRestrictions types.Bool `tfsdk:"no_restrictions"`
	Sign           types.Bool `tfsdk:"sign"`
	Unwrap         types.Bool `tfsdk:"unwrap"`
	Verify         types.Bool `tfsdk:"verify"`
	Wrap           types.Bool `tfsdk:"wrap"`
}

type complexArgumentData struct {
	NestedRequired types.String `tfsdk:"nested_required"`
	NestedOptional types.String `tfsdk:"nested_optional"`
}

var complexArgumentAttrTypes = map[string]attr.Type{
	"nested_required": types.StringType,
	"nested_optional": types.StringType,
}
