// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package pinpointsmsvoicev2

import (
	"context"
	"fmt"
	"strings"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/pinpointsmsvoicev2"
	awstypes "github.com/aws/aws-sdk-go-v2/service/pinpointsmsvoicev2/types"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	sdkid "github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

const (
	senderIDResourceIDSeparator = ","
)

// @FrameworkResource("aws_pinpointsmsvoicev2_sender_id", name="Sender ID")
// @Tags(identifierAttribute="arn")
func newSenderIDResource(context.Context) (resource.ResourceWithConfigure, error) {
	r := &senderIDResource{}

	return r, nil
}

type senderIDResource struct {
	framework.ResourceWithModel[senderIDResourceModel]
	framework.WithImportByID
}

func (r *senderIDResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
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
			"message_types": schema.SetAttribute{
				CustomType:  fwtypes.NewSetTypeOf[fwtypes.StringEnum[awstypes.MessageType]](ctx),
				Optional:    true,
				Computed:    true,
				ElementType: fwtypes.StringEnumType[awstypes.MessageType](),
				PlanModifiers: []planmodifier.Set{
					setplanmodifier.UseStateForUnknown(),
					setplanmodifier.RequiresReplace(),
				},
			},
			"monthly_leasing_price": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"registered": schema.BoolAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"registration_id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"sender_id": schema.StringAttribute{
				CustomType: fwtypes.CaseInsensitiveStringType,
				Required:   true,
				Validators: []validator.String{
					stringvalidator.RegexMatches(regexache.MustCompile(`^[A-Za-z0-9-]{3,11}$`), "must be between 3 and 11 characters and contain only letters, numbers, and dashes"),
					stringvalidator.RegexMatches(regexache.MustCompile(`[A-Za-z]`), "must contain at least one letter (numeric-only sender IDs are not supported)"),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
		},
	}
}

func (r *senderIDResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data senderIDResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().PinpointSMSVoiceV2Client(ctx)

	input := &pinpointsmsvoicev2.RequestSenderIdInput{
		SenderId:       data.SenderID.ValueStringPointer(),
		IsoCountryCode: data.ISOCountryCode.ValueStringPointer(),
		ClientToken:    aws.String(sdkid.UniqueId()),
		Tags:           getTagsIn(ctx),
	}

	if !data.DeletionProtectionEnabled.IsNull() {
		input.DeletionProtectionEnabled = data.DeletionProtectionEnabled.ValueBoolPointer()
	}

	if !data.MessageTypes.IsNull() && !data.MessageTypes.IsUnknown() {
		var messageTypes []fwtypes.StringEnum[awstypes.MessageType]
		response.Diagnostics.Append(data.MessageTypes.ElementsAs(ctx, &messageTypes, false)...)
		if response.Diagnostics.HasError() {
			return
		}
		for _, mt := range messageTypes {
			input.MessageTypes = append(input.MessageTypes, mt.ValueEnum())
		}
	}

	output, err := conn.RequestSenderId(ctx, input)

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("requesting End User Messaging SMS Sender ID (%s)", data.SenderID.ValueString()), err.Error())

		return
	}

	// Use API-returned values (AWS may uppercase sender_id).
	data.SenderID = fwtypes.CaseInsensitiveStringValue(aws.ToString(output.SenderId))
	data.ISOCountryCode = fwflex.StringToFramework(ctx, output.IsoCountryCode)
	data.SenderIDARN = fwflex.StringToFramework(ctx, output.SenderIdArn)
	data.MonthlyLeasingPrice = fwflex.StringToFramework(ctx, output.MonthlyLeasingPrice)
	data.Registered = types.BoolValue(output.Registered)
	data.RegistrationID = types.StringNull()
	data.setID()

	messageTypeValues := make([]attr.Value, len(output.MessageTypes))
	for i, mt := range output.MessageTypes {
		messageTypeValues[i] = fwtypes.StringEnumValue(awstypes.MessageType(strings.ToUpper(string(mt))))
	}
	data.MessageTypes = fwtypes.NewSetValueOfMust[fwtypes.StringEnum[awstypes.MessageType]](ctx, messageTypeValues)

	response.Diagnostics.Append(response.State.Set(ctx, data)...)
}

func (r *senderIDResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data senderIDResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	if err := data.InitFromID(); err != nil {
		response.Diagnostics.AddError("parsing resource ID", err.Error())

		return
	}

	conn := r.Meta().PinpointSMSVoiceV2Client(ctx)

	out, err := findSenderIDByTwoPartKey(ctx, conn, data.SenderID.ValueString(), data.ISOCountryCode.ValueString())

	if retry.NotFound(err) {
		response.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)

		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading End User Messaging SMS Sender ID (%s)", data.ID.ValueString()), err.Error())

		return
	}

	data.flattenSenderIdInformation(ctx, out)

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *senderIDResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var old, new senderIDResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &new)...)
	if response.Diagnostics.HasError() {
		return
	}
	response.Diagnostics.Append(request.State.Get(ctx, &old)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().PinpointSMSVoiceV2Client(ctx)

	if !new.DeletionProtectionEnabled.Equal(old.DeletionProtectionEnabled) {
		input := &pinpointsmsvoicev2.UpdateSenderIdInput{
			SenderId:                  new.SenderID.ValueStringPointer(),
			IsoCountryCode:            new.ISOCountryCode.ValueStringPointer(),
			DeletionProtectionEnabled: new.DeletionProtectionEnabled.ValueBoolPointer(),
		}

		_, err := conn.UpdateSenderId(ctx, input)

		if err != nil {
			response.Diagnostics.AddError(fmt.Sprintf("updating End User Messaging SMS Sender ID (%s)", new.ID.ValueString()), err.Error())

			return
		}
	}

	response.Diagnostics.Append(response.State.Set(ctx, &new)...)
}

func (r *senderIDResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data senderIDResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().PinpointSMSVoiceV2Client(ctx)

	_, err := conn.ReleaseSenderId(ctx, &pinpointsmsvoicev2.ReleaseSenderIdInput{
		SenderId:       data.SenderID.ValueStringPointer(),
		IsoCountryCode: data.ISOCountryCode.ValueStringPointer(),
	})

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("releasing End User Messaging SMS Sender ID (%s)", data.ID.ValueString()), err.Error())

		return
	}
}

type senderIDResourceModel struct {
	framework.WithRegionModel
	DeletionProtectionEnabled types.Bool                                    `tfsdk:"deletion_protection_enabled"`
	ID                        types.String                                  `tfsdk:"id"`
	ISOCountryCode            types.String                                  `tfsdk:"iso_country_code"`
	MessageTypes              fwtypes.SetOfStringEnum[awstypes.MessageType] `tfsdk:"message_types"`
	MonthlyLeasingPrice       types.String                                  `tfsdk:"monthly_leasing_price"`
	Registered                types.Bool                                    `tfsdk:"registered"`
	RegistrationID            types.String                                  `tfsdk:"registration_id"`
	SenderID                  fwtypes.CaseInsensitiveString                 `tfsdk:"sender_id"`
	SenderIDARN               types.String                                  `tfsdk:"arn"`
	Tags                      tftags.Map                                    `tfsdk:"tags"`
	TagsAll                   tftags.Map                                    `tfsdk:"tags_all"`
}

func (model *senderIDResourceModel) InitFromID() error {
	parts := strings.Split(model.ID.ValueString(), senderIDResourceIDSeparator)

	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return fmt.Errorf("unexpected format for ID (%[1]s), expected SenderID%[2]sISOCountryCode", model.ID.ValueString(), senderIDResourceIDSeparator)
	}

	model.SenderID = fwtypes.CaseInsensitiveStringValue(parts[0])
	model.ISOCountryCode = types.StringValue(parts[1])

	return nil
}

func (model *senderIDResourceModel) setID() {
	model.ID = types.StringValue(model.SenderID.ValueString() + senderIDResourceIDSeparator + model.ISOCountryCode.ValueString())
}

func (model *senderIDResourceModel) flattenSenderIdInformation(ctx context.Context, out *awstypes.SenderIdInformation) {
	model.SenderIDARN = fwflex.StringToFramework(ctx, out.SenderIdArn)
	model.SenderID = fwtypes.CaseInsensitiveStringValue(aws.ToString(out.SenderId))
	model.ISOCountryCode = fwflex.StringToFramework(ctx, out.IsoCountryCode)
	model.DeletionProtectionEnabled = types.BoolValue(out.DeletionProtectionEnabled)
	model.MonthlyLeasingPrice = fwflex.StringToFramework(ctx, out.MonthlyLeasingPrice)
	model.Registered = types.BoolValue(out.Registered)
	model.RegistrationID = fwflex.StringToFramework(ctx, out.RegistrationId)

	messageTypeValues := make([]attr.Value, len(out.MessageTypes))
	for i, mt := range out.MessageTypes {
		messageTypeValues[i] = fwtypes.StringEnumValue(awstypes.MessageType(strings.ToUpper(string(mt))))
	}
	model.MessageTypes = fwtypes.NewSetValueOfMust[fwtypes.StringEnum[awstypes.MessageType]](ctx, messageTypeValues)

	model.setID()
}

func findSenderIDByTwoPartKey(ctx context.Context, conn *pinpointsmsvoicev2.Client, senderID, isoCountryCode string) (*awstypes.SenderIdInformation, error) {
	input := &pinpointsmsvoicev2.DescribeSenderIdsInput{
		SenderIds: []awstypes.SenderIdAndCountry{
			{
				SenderId:       aws.String(senderID),
				IsoCountryCode: aws.String(isoCountryCode),
			},
		},
	}

	return findSenderID(ctx, conn, input)
}

func findSenderID(ctx context.Context, conn *pinpointsmsvoicev2.Client, input *pinpointsmsvoicev2.DescribeSenderIdsInput) (*awstypes.SenderIdInformation, error) {
	output, err := findSenderIDs(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findSenderIDs(ctx context.Context, conn *pinpointsmsvoicev2.Client, input *pinpointsmsvoicev2.DescribeSenderIdsInput) ([]awstypes.SenderIdInformation, error) {
	var output []awstypes.SenderIdInformation

	pages := pinpointsmsvoicev2.NewDescribeSenderIdsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return nil, &retry.NotFoundError{
				LastError: err,
			}
		}

		if err != nil {
			return nil, err
		}

		output = append(output, page.SenderIds...)
	}

	return output, nil
}
