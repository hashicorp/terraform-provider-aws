// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

package paymentcryptography

import (
	"context"
	"errors"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/paymentcryptography"
	awstypes "github.com/aws/aws-sdk-go-v2/service/paymentcryptography/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_paymentcryptography_key", name="Key")
// @Tags(identifierAttribute="arn")
// @ArnIdentity(identityDuplicateAttributes="id")
// @Testing(existsType="github.com/aws/aws-sdk-go-v2/service/paymentcryptography;paymentcryptography.GetKeyOutput")
// @Testing(generator=false)
// @Testing(importIgnore="deletion_window_in_days")
// @Testing(preIdentityVersion="v5.100.0")
func newKeyResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &keyResource{}

	r.SetDefaultCreateTimeout(30 * time.Minute)
	r.SetDefaultUpdateTimeout(30 * time.Minute)
	r.SetDefaultDeleteTimeout(30 * time.Minute)

	return r, nil
}

const (
	ResNameKey                  = "Key"
	defaultDeletionWindowInDays = 7
)

type keyResource struct {
	framework.ResourceWithModel[keyResourceModel]
	framework.WithImportByIdentity
	framework.WithTimeouts
}

func (r *keyResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Version: 1,
		Attributes: map[string]schema.Attribute{
			names.AttrARN: framework.ARNAttributeComputedOnly(),
			names.AttrID:  framework.IDAttribute(),
			"deletion_window_in_days": schema.Int64Attribute{
				Optional: true,
				Computed: true,
				Default:  int64default.StaticInt64(defaultDeletionWindowInDays),
				Validators: []validator.Int64{
					int64validator.Between(3, 180),
				},
			},
			names.AttrEnabled: schema.BoolAttribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
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
				Computed:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
					stringplanmodifier.UseStateForUnknown(),
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
		},
		Blocks: map[string]schema.Block{
			"key_attributes": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[keyAttributesModel](ctx),
				NestedObject: schema.NestedBlockObject{
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
						"key_modes_of_use": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[keyModesOfUseModel](ctx),
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"decrypt": schema.BoolAttribute{
										Optional: true,
										Computed: true,
										PlanModifiers: []planmodifier.Bool{
											boolplanmodifier.RequiresReplace(),
											boolplanmodifier.UseStateForUnknown(),
										},
									},
									"derive_key": schema.BoolAttribute{
										Optional: true,
										Computed: true,
										PlanModifiers: []planmodifier.Bool{
											boolplanmodifier.RequiresReplace(),
											boolplanmodifier.UseStateForUnknown(),
										},
									},
									"encrypt": schema.BoolAttribute{
										Optional: true,
										Computed: true,
										PlanModifiers: []planmodifier.Bool{
											boolplanmodifier.RequiresReplace(),
											boolplanmodifier.UseStateForUnknown(),
										},
									},
									"generate": schema.BoolAttribute{
										Optional: true,
										Computed: true,
										PlanModifiers: []planmodifier.Bool{
											boolplanmodifier.RequiresReplace(),
											boolplanmodifier.UseStateForUnknown(),
										},
									},
									"no_restrictions": schema.BoolAttribute{
										Optional: true,
										Computed: true,
										PlanModifiers: []planmodifier.Bool{
											boolplanmodifier.RequiresReplace(),
											boolplanmodifier.UseStateForUnknown(),
										},
									},
									"sign": schema.BoolAttribute{
										Optional: true,
										Computed: true,
										PlanModifiers: []planmodifier.Bool{
											boolplanmodifier.RequiresReplace(),
											boolplanmodifier.UseStateForUnknown(),
										},
									},
									"unwrap": schema.BoolAttribute{
										Optional: true,
										Computed: true,
										PlanModifiers: []planmodifier.Bool{
											boolplanmodifier.RequiresReplace(),
											boolplanmodifier.UseStateForUnknown(),
										},
									},
									"verify": schema.BoolAttribute{
										Optional: true,
										Computed: true,
										PlanModifiers: []planmodifier.Bool{
											boolplanmodifier.RequiresReplace(),
											boolplanmodifier.UseStateForUnknown(),
										},
									},
									"wrap": schema.BoolAttribute{
										Optional: true,
										Computed: true,
										PlanModifiers: []planmodifier.Bool{
											boolplanmodifier.RequiresReplace(),
											boolplanmodifier.UseStateForUnknown(),
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

func (r *keyResource) UpgradeState(ctx context.Context) map[int64]resource.StateUpgrader {
	schemaV0 := keySchemaV0(ctx)

	return map[int64]resource.StateUpgrader{
		0: {
			PriorSchema:   &schemaV0,
			StateUpgrader: upgradeKeyStateV0toV1,
		},
	}
}

func (r *keyResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	conn := r.Meta().PaymentCryptographyClient(ctx)

	var plan keyResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}

	in := &paymentcryptography.CreateKeyInput{}
	response.Diagnostics.Append(flex.Expand(ctx, &plan, in)...)

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

	plan.KeyARN = flex.StringToFramework(ctx, out.Key.KeyArn)
	plan.setId()

	createTimeout := r.CreateTimeout(ctx, plan.Timeouts)
	created, err := waitKeyCreated(ctx, conn, plan.ID.ValueString(), createTimeout)
	if err != nil {
		response.Diagnostics.AddError(
			create.ProblemStandardMessage(names.PaymentCryptography, create.ErrActionWaitingForCreation, ResNameKey, plan.KeyARN.String(), err),
			err.Error(),
		)
		return
	}

	response.Diagnostics.Append(flex.Flatten(ctx, created, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &plan)...)
}

func (r *keyResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	conn := r.Meta().PaymentCryptographyClient(ctx)

	var state keyResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	out, err := findKeyByID(ctx, conn, state.ID.ValueString())
	if retry.NotFound(err) {
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

func (r *keyResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var old, new keyResourceModel
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
					create.ProblemStandardMessage(names.PaymentCryptography, create.ErrActionUpdating, ResNameKey, new.KeyARN.String(), err),
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
					create.ProblemStandardMessage(names.PaymentCryptography, create.ErrActionUpdating, ResNameKey, new.KeyARN.String(), err),
					err.Error(),
				)
				return
			}
		}
		out, err := findKeyByID(ctx, conn, new.ID.ValueString())
		if err != nil {
			response.Diagnostics.AddError(
				create.ProblemStandardMessage(names.PaymentCryptography, create.ErrActionSetting, ResNameKey, new.ID.String(), err),
				err.Error(),
			)
			return
		}
		response.Diagnostics.Append(flex.Flatten(ctx, out, &new)...)
		if response.Diagnostics.HasError() {
			return
		}
	}

	response.Diagnostics.Append(response.State.Set(ctx, &new)...)
}

func (r *keyResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	conn := r.Meta().PaymentCryptographyClient(ctx)

	var state keyResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	in := &paymentcryptography.DeleteKeyInput{
		KeyIdentifier: state.ID.ValueStringPointer(),
	}

	_, err := conn.DeleteKey(ctx, in)
	if err != nil {
		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return
		}
		// Check if the errors is about the key not being in CREATE_COMPLETE, if it's been deleted outside.
		if errs.IsAErrorMessageContains[*awstypes.ValidationException](err, "not in CREATE_COMPLETE state.") {
			return
		}
		response.Diagnostics.AddError(
			create.ProblemStandardMessage(names.PaymentCryptography, create.ErrActionDeleting, ResNameKey, state.ID.String(), err),
			err.Error(),
		)
		return
	}

	deleteTimeout := r.DeleteTimeout(ctx, state.Timeouts)
	_, err = waitKeyDeleted(ctx, conn, state.ID.ValueString(), deleteTimeout)
	if err != nil {
		response.Diagnostics.AddError(
			create.ProblemStandardMessage(names.PaymentCryptography, create.ErrActionWaitingForDeletion, ResNameKey, state.ID.String(), err),
			err.Error(),
		)
		return
	}
}

func waitKeyCreated(ctx context.Context, conn *paymentcryptography.Client, id string, timeout time.Duration) (*awstypes.Key, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   enum.Slice(awstypes.KeyStateCreateInProgress),
		Target:                    enum.Slice(awstypes.KeyStateCreateComplete),
		Refresh:                   statusKey(conn, id),
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
		Target:  []string{},
		Refresh: statusKey(conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*awstypes.Key); ok {
		return out, err
	}

	return nil, err
}

func statusKey(conn *paymentcryptography.Client, id string) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		out, err := findKeyByID(ctx, conn, id)
		if retry.NotFound(err) {
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
				LastError: err,
			}
		}

		return nil, err
	}

	if out == nil || out.Key == nil {
		return nil, tfresource.NewEmptyResultError()
	}

	// If the key is either Pending or Complete deletion state Terraform considers it logically deleted.
	if state := out.Key.KeyState; state == awstypes.KeyStateDeletePending || state == awstypes.KeyStateDeleteComplete {
		return nil, &retry.NotFoundError{
			Message: string(state),
		}
	}

	return out.Key, nil
}

type keyResourceModel struct {
	framework.WithRegionModel
	KeyARN                 types.String                                        `tfsdk:"arn"`
	DeletionWindowInDays   types.Int64                                         `tfsdk:"deletion_window_in_days"`
	Enabled                types.Bool                                          `tfsdk:"enabled"`
	Exportable             types.Bool                                          `tfsdk:"exportable"`
	ID                     types.String                                        `tfsdk:"id"`
	KeyAttributes          fwtypes.ListNestedObjectValueOf[keyAttributesModel] `tfsdk:"key_attributes"`
	KeyCheckValue          types.String                                        `tfsdk:"key_check_value"`
	KeyCheckValueAlgorithm fwtypes.StringEnum[awstypes.KeyCheckValueAlgorithm] `tfsdk:"key_check_value_algorithm"`
	KeyOrigin              fwtypes.StringEnum[awstypes.KeyOrigin]              `tfsdk:"key_origin"`
	KeyState               fwtypes.StringEnum[awstypes.KeyState]               `tfsdk:"key_state"`
	Tags                   tftags.Map                                          `tfsdk:"tags"`
	TagsAll                tftags.Map                                          `tfsdk:"tags_all"`
	Timeouts               timeouts.Value                                      `tfsdk:"timeouts"`
}

func (k *keyResourceModel) setId() {
	k.ID = k.KeyARN
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
