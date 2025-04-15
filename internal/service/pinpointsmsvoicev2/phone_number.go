// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package pinpointsmsvoicev2

import (
	"context"
	"fmt"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/pinpointsmsvoicev2"
	awstypes "github.com/aws/aws-sdk-go-v2/service/pinpointsmsvoicev2/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/boolvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
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
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_pinpointsmsvoicev2_phone_number", name="Phone Number")
// @Tags(identifierAttribute="arn")
func newPhoneNumberResource(context.Context) (resource.ResourceWithConfigure, error) {
	r := &phoneNumberResource{}

	r.SetDefaultCreateTimeout(30 * time.Minute)
	r.SetDefaultUpdateTimeout(30 * time.Minute)
	r.SetDefaultDeleteTimeout(30 * time.Minute)

	return r, nil
}

type phoneNumberResource struct {
	framework.ResourceWithConfigure
	framework.WithImportByID
	framework.WithTimeouts
}

func (r *phoneNumberResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrARN: framework.ARNAttributeComputedOnly(),
			"deletion_protection_enabled": schema.BoolAttribute{
				Optional: true,
				Computed: true,
				Default:  booldefault.StaticBool(false),
			},
			names.AttrID: framework.IDAttribute(),
			"iso_country_code": schema.StringAttribute{
				Required: true,
				Validators: []validator.String{
					stringvalidator.RegexMatches(regexache.MustCompile(`^[A-Z]{2}$`), "must be in ISO 3166-1 alpha-2 format"),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"message_type": schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.MessageType](),
				Required:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"monthly_leasing_price": schema.StringAttribute{
				Computed: true,
			},
			"number_capabilities": schema.SetAttribute{
				CustomType:  fwtypes.NewSetTypeOf[fwtypes.StringEnum[awstypes.NumberCapability]](ctx),
				Required:    true,
				ElementType: fwtypes.StringEnumType[awstypes.NumberCapability](),
				PlanModifiers: []planmodifier.Set{
					setplanmodifier.RequiresReplace(),
				},
			},
			"number_type": schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.RequestableNumberType](),
				Required:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"opt_out_list_name": schema.StringAttribute{
				Optional: true,
				Computed: true,
				Default:  stringdefault.StaticString("Default"),
			},
			"phone_number": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"registration_id": schema.StringAttribute{
				Optional: true,
			},
			"self_managed_opt_outs_enabled": schema.BoolAttribute{
				Optional: true,
				Computed: true,
				Default:  booldefault.StaticBool(false),
			},
			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
			"two_way_channel_arn": schema.StringAttribute{
				CustomType: fwtypes.ARNType,
				Optional:   true,
				Validators: []validator.String{
					stringvalidator.AlsoRequires(
						path.MatchRelative().AtParent().AtName("two_way_channel_enabled"),
					),
				},
			},
			"two_way_channel_enabled": schema.BoolAttribute{
				Optional: true,
				Computed: true,
				Default:  booldefault.StaticBool(false),
				Validators: []validator.Bool{
					boolvalidator.AlsoRequires(
						path.MatchRelative().AtParent().AtName("two_way_channel_enabled"),
					),
				},
			},
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

func (r *phoneNumberResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data phoneNumberResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().PinpointSMSVoiceV2Client(ctx)

	input := &pinpointsmsvoicev2.RequestPhoneNumberInput{}
	response.Diagnostics.Append(fwflex.Expand(ctx, data, input)...)
	if response.Diagnostics.HasError() {
		return
	}

	// Additional fields.
	input.ClientToken = aws.String(sdkid.UniqueId())
	input.Tags = getTagsIn(ctx)

	output, err := conn.RequestPhoneNumber(ctx, input)

	if err != nil {
		response.Diagnostics.AddError("requesting End User Messaging SMS Phone Number", err.Error())

		return
	}

	// Set values for unknowns.
	data.PhoneNumberID = fwflex.StringToFramework(ctx, output.PhoneNumberId)
	response.State.SetAttribute(ctx, path.Root(names.AttrID), data.PhoneNumberID) // Set 'id' so as to taint the resource.

	out, err := waitPhoneNumberActive(ctx, conn, data.PhoneNumberID.ValueString(), r.CreateTimeout(ctx, data.Timeouts))

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("waiting for End User Messaging SMS Phone Number (%s) create", data.PhoneNumberID.ValueString()), err.Error())

		return
	}

	if (!data.SelfManagedOptOutsEnabled.IsNull() && data.SelfManagedOptOutsEnabled.ValueBool()) ||
		!data.TwoWayChannelARN.IsNull() ||
		(!data.TwoWayEnabled.IsNull() && data.TwoWayEnabled.ValueBool()) {
		input := &pinpointsmsvoicev2.UpdatePhoneNumberInput{
			PhoneNumberId:             fwflex.StringFromFramework(ctx, data.PhoneNumberID),
			SelfManagedOptOutsEnabled: fwflex.BoolFromFramework(ctx, data.SelfManagedOptOutsEnabled),
			TwoWayChannelArn:          fwflex.StringFromFramework(ctx, data.TwoWayChannelARN),
			TwoWayEnabled:             fwflex.BoolFromFramework(ctx, data.TwoWayEnabled),
		}

		_, err := conn.UpdatePhoneNumber(ctx, input)

		if err != nil {
			response.Diagnostics.AddError(fmt.Sprintf("updating End User Messaging SMS Phone Number (%s)", data.PhoneNumberID.ValueString()), err.Error())

			return
		}

		out, err = waitPhoneNumberActive(ctx, conn, data.PhoneNumberID.ValueString(), r.CreateTimeout(ctx, data.Timeouts))

		if err != nil {
			response.Diagnostics.AddError(fmt.Sprintf("waiting for End User Messaging SMS Phone Number (%s) create", data.PhoneNumberID.ValueString()), err.Error())

			return
		}
	}

	response.Diagnostics.Append(fwflex.Flatten(ctx, out, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, data)...)
}

func (r *phoneNumberResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data phoneNumberResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().PinpointSMSVoiceV2Client(ctx)

	out, err := findPhoneNumberByID(ctx, conn, data.PhoneNumberID.ValueString())

	if tfresource.NotFound(err) {
		response.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)

		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading End User Messaging SMS Phone Number (%s)", data.PhoneNumberID.ValueString()), err.Error())

		return
	}

	// Set attributes for import.
	response.Diagnostics.Append(fwflex.Flatten(ctx, out, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *phoneNumberResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var old, new phoneNumberResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &new)...)
	if response.Diagnostics.HasError() {
		return
	}
	response.Diagnostics.Append(request.State.Get(ctx, &old)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().PinpointSMSVoiceV2Client(ctx)

	if !new.DeletionProtectionEnabled.Equal(old.DeletionProtectionEnabled) ||
		!new.OptOutListName.Equal(old.OptOutListName) ||
		!new.SelfManagedOptOutsEnabled.Equal(old.SelfManagedOptOutsEnabled) ||
		!new.TwoWayChannelARN.Equal(old.TwoWayChannelARN) ||
		!new.TwoWayEnabled.Equal(old.TwoWayEnabled) {
		input := &pinpointsmsvoicev2.UpdatePhoneNumberInput{}
		response.Diagnostics.Append(fwflex.Expand(ctx, new, input)...)
		if response.Diagnostics.HasError() {
			return
		}

		_, err := conn.UpdatePhoneNumber(ctx, input)

		if err != nil {
			response.Diagnostics.AddError(fmt.Sprintf("updating End User Messaging SMS Phone Number (%s)", new.PhoneNumberID.ValueString()), err.Error())

			return
		}

		out, err := waitPhoneNumberActive(ctx, conn, new.PhoneNumberID.ValueString(), r.UpdateTimeout(ctx, new.Timeouts))

		if err != nil {
			response.Diagnostics.AddError(fmt.Sprintf("waiting for End User Messaging SMS Phone Number (%s) update", new.PhoneNumberID.ValueString()), err.Error())

			return
		}

		new.MonthlyLeasingPrice = fwflex.StringToFramework(ctx, out.MonthlyLeasingPrice)
	} else {
		new.MonthlyLeasingPrice = old.MonthlyLeasingPrice
	}

	response.Diagnostics.Append(response.State.Set(ctx, &new)...)
}

func (r *phoneNumberResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data phoneNumberResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().PinpointSMSVoiceV2Client(ctx)

	_, err := conn.ReleasePhoneNumber(ctx, &pinpointsmsvoicev2.ReleasePhoneNumberInput{
		PhoneNumberId: data.PhoneNumberID.ValueStringPointer(),
	})

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("releasing End User Messaging SMS Phone Number (%s)", data.PhoneNumberID.ValueString()), err.Error())

		return
	}

	if _, err := waitPhoneNumberDeleted(ctx, conn, data.PhoneNumberID.ValueString(), r.DeleteTimeout(ctx, data.Timeouts)); err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("waiting for End User Messaging SMS Phone Number (%s) delete", data.PhoneNumberID.ValueString()), err.Error())

		return
	}
}

type phoneNumberResourceModel struct {
	DeletionProtectionEnabled types.Bool                                                        `tfsdk:"deletion_protection_enabled"`
	ISOCountryCode            types.String                                                      `tfsdk:"iso_country_code"`
	MessageType               fwtypes.StringEnum[awstypes.MessageType]                          `tfsdk:"message_type"`
	MonthlyLeasingPrice       types.String                                                      `tfsdk:"monthly_leasing_price"`
	NumberCapabilities        fwtypes.SetValueOf[fwtypes.StringEnum[awstypes.NumberCapability]] `tfsdk:"number_capabilities"`
	NumberType                fwtypes.StringEnum[awstypes.RequestableNumberType]                `tfsdk:"number_type"`
	OptOutListName            types.String                                                      `tfsdk:"opt_out_list_name"`
	PhoneNumber               types.String                                                      `tfsdk:"phone_number"`
	PhoneNumberARN            types.String                                                      `tfsdk:"arn"`
	PhoneNumberID             types.String                                                      `tfsdk:"id"`
	RegistrationID            types.String                                                      `tfsdk:"registration_id"`
	SelfManagedOptOutsEnabled types.Bool                                                        `tfsdk:"self_managed_opt_outs_enabled"`
	Tags                      tftags.Map                                                        `tfsdk:"tags"`
	TagsAll                   tftags.Map                                                        `tfsdk:"tags_all"`
	Timeouts                  timeouts.Value                                                    `tfsdk:"timeouts"`
	TwoWayChannelARN          fwtypes.ARN                                                       `tfsdk:"two_way_channel_arn"`
	TwoWayEnabled             types.Bool                                                        `tfsdk:"two_way_channel_enabled"`
}

func findPhoneNumberByID(ctx context.Context, conn *pinpointsmsvoicev2.Client, id string) (*awstypes.PhoneNumberInformation, error) {
	input := &pinpointsmsvoicev2.DescribePhoneNumbersInput{
		PhoneNumberIds: []string{id},
	}

	output, err := findPhoneNumber(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	if status := output.Status; status == awstypes.NumberStatusDeleted {
		return nil, &retry.NotFoundError{
			Message:     string(status),
			LastRequest: input,
		}
	}

	return output, nil
}

func findPhoneNumber(ctx context.Context, conn *pinpointsmsvoicev2.Client, input *pinpointsmsvoicev2.DescribePhoneNumbersInput) (*awstypes.PhoneNumberInformation, error) {
	output, err := findPhoneNumbers(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findPhoneNumbers(ctx context.Context, conn *pinpointsmsvoicev2.Client, input *pinpointsmsvoicev2.DescribePhoneNumbersInput) ([]awstypes.PhoneNumberInformation, error) {
	var output []awstypes.PhoneNumberInformation

	pages := pinpointsmsvoicev2.NewDescribePhoneNumbersPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: input,
			}
		}

		if err != nil {
			return nil, err
		}

		output = append(output, page.PhoneNumbers...)
	}

	return output, nil
}

func statusPhoneNumber(ctx context.Context, conn *pinpointsmsvoicev2.Client, id string) retry.StateRefreshFunc {
	return func() (any, string, error) {
		output, err := findPhoneNumberByID(ctx, conn, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.Status), nil
	}
}

func waitPhoneNumberActive(ctx context.Context, conn *pinpointsmsvoicev2.Client, id string, timeout time.Duration) (*awstypes.PhoneNumberInformation, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.NumberStatusPending, awstypes.NumberStatusAssociating),
		Target:  enum.Slice(awstypes.NumberStatusActive),
		Refresh: statusPhoneNumber(ctx, conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.PhoneNumberInformation); ok {
		return output, err
	}

	return nil, err
}

func waitPhoneNumberDeleted(ctx context.Context, conn *pinpointsmsvoicev2.Client, id string, timeout time.Duration) (*awstypes.PhoneNumberInformation, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.NumberStatusDisassociating),
		Target:  []string{},
		Refresh: statusPhoneNumber(ctx, conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.PhoneNumberInformation); ok {
		return output, err
	}

	return nil, err
}
