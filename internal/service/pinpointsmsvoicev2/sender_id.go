// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package pinpointsmsvoicev2

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/YakDriver/smarterr"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/pinpointsmsvoicev2"
	awstypes "github.com/aws/aws-sdk-go-v2/service/pinpointsmsvoicev2/types"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	intflex "github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

const (
	// A sender ID re-requested immediately after being released returns a
	// ValidationException with Reason="SENDER_ID_REQUIRES_REGISTRATION" until the
	// release propagates. This is the window Create retries over, e.g. when a change
	// to an immutable attribute forces a destroy-then-create on the same sender ID.
	senderIDRegistrationPropagationTimeout = 2 * time.Minute
)

// @FrameworkResource("aws_pinpointsmsvoicev2_sender_id", name="Sender ID")
// @Tags(identifierAttribute="arn")
// @IdentityAttribute("sender_id")
// @IdentityAttribute("iso_country_code")
// @ImportIDHandler("senderIDImportID")
// @Testing(hasNoPreExistingResource=true)
// @Testing(preCheck="testAccPreCheckSenderID")
// @Testing(existsType="github.com/aws/aws-sdk-go-v2/service/pinpointsmsvoicev2/types;awstypes.SenderIdInformation")
// @Testing(generator="testAccRandomSenderID(t)")
// @Testing(importStateIdAttributes="sender_id;iso_country_code", importStateIdAttributesSep="flex.ResourceIdSeparator")
func newSenderIDResource(context.Context) (resource.ResourceWithConfigure, error) {
	r := &senderIDResource{}

	return r, nil
}

type senderIDResource struct {
	framework.ResourceWithModel[senderIDResourceModel]
	framework.WithImportByIdentity
}

func (r *senderIDResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrARN: framework.ARNAttributeComputedOnly(),
			"deletion_protection_enabled": schema.BoolAttribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
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
				Required: true,
				Validators: []validator.String{
					stringvalidator.RegexMatches(regexache.MustCompile(`^[A-Z0-9-]{3,11}$`), "must be between 3 and 11 characters and contain only uppercase letters, numbers, and dashes"),
					stringvalidator.RegexMatches(regexache.MustCompile(`[A-Z]`), "must contain at least one letter (numeric-only sender IDs are not supported)"),
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
	smerr.AddEnrich(ctx, &response.Diagnostics, request.Plan.Get(ctx, &data))
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().PinpointSMSVoiceV2Client(ctx)

	var input pinpointsmsvoicev2.RequestSenderIdInput
	smerr.AddEnrich(ctx, &response.Diagnostics, fwflex.Expand(ctx, data, &input))
	if response.Diagnostics.HasError() {
		return
	}
	input.Tags = getTagsIn(ctx)

	// A fresh ClientToken is generated on each attempt: AWS pins the result of a
	// request to its ClientToken, so reusing one token would replay the cached
	// SENDER_ID_REQUIRES_REGISTRATION failure instead of re-evaluating the request.
	outputRaw, err := tfresource.RetryWhenAWSErrMessageContains(ctx, senderIDRegistrationPropagationTimeout, func(ctx context.Context) (any, error) {
		input.ClientToken = aws.String(create.UniqueId(ctx))
		return conn.RequestSenderId(ctx, &input)
	}, "ValidationException", "SENDER_ID_REQUIRES_REGISTRATION")

	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, data.SenderID.String())
		return
	}

	output := outputRaw.(*pinpointsmsvoicev2.RequestSenderIdOutput)

	smerr.AddEnrich(ctx, &response.Diagnostics, fwflex.Flatten(ctx, output, &data))
	if response.Diagnostics.HasError() {
		return
	}

	// RequestSenderId does not return a registration ID, so set it explicitly after
	// flattening the response.
	data.RegistrationID = types.StringNull()

	// DescribeSenderIds (used on Read/import) returns message types in mixed case
	// (e.g. "Transactional"), while the canonical enum values are upper case. Normalize
	// here so Create-time state matches what Read produces and import stays consistent.
	data.MessageTypes = normalizeMessageTypes(ctx, output.MessageTypes)

	smerr.AddEnrich(ctx, &response.Diagnostics, response.State.Set(ctx, &data))
}

func (r *senderIDResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data senderIDResourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.State.Get(ctx, &data))
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().PinpointSMSVoiceV2Client(ctx)

	out, err := findSenderIDByTwoPartKey(ctx, conn, data.SenderID.ValueString(), data.ISOCountryCode.ValueString())

	if retry.NotFound(err) {
		smerr.AddOne(ctx, &response.Diagnostics, fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)

		return
	}

	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, data.SenderID.String())
		return
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, data.flatten(ctx, out))
	if response.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, response.State.Set(ctx, &data))
}

func (r *senderIDResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var old, new senderIDResourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.Plan.Get(ctx, &new))
	if response.Diagnostics.HasError() {
		return
	}
	smerr.AddEnrich(ctx, &response.Diagnostics, request.State.Get(ctx, &old))
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().PinpointSMSVoiceV2Client(ctx)

	if !new.DeletionProtectionEnabled.Equal(old.DeletionProtectionEnabled) {
		input := pinpointsmsvoicev2.UpdateSenderIdInput{
			SenderId:                  new.SenderID.ValueStringPointer(),
			IsoCountryCode:            new.ISOCountryCode.ValueStringPointer(),
			DeletionProtectionEnabled: new.DeletionProtectionEnabled.ValueBoolPointer(),
		}

		_, err := conn.UpdateSenderId(ctx, &input)

		if err != nil {
			smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, new.SenderID.String())
			return
		}
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, response.State.Set(ctx, &new))
}

func (r *senderIDResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data senderIDResourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.State.Get(ctx, &data))
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().PinpointSMSVoiceV2Client(ctx)

	input := pinpointsmsvoicev2.ReleaseSenderIdInput{
		SenderId:       data.SenderID.ValueStringPointer(),
		IsoCountryCode: data.ISOCountryCode.ValueStringPointer(),
	}

	_, err := conn.ReleaseSenderId(ctx, &input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return
	}

	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, data.SenderID.String())
		return
	}
}

type senderIDResourceModel struct {
	framework.WithRegionModel
	DeletionProtectionEnabled types.Bool                                    `tfsdk:"deletion_protection_enabled"`
	ISOCountryCode            types.String                                  `tfsdk:"iso_country_code"`
	MessageTypes              fwtypes.SetOfStringEnum[awstypes.MessageType] `tfsdk:"message_types"`
	MonthlyLeasingPrice       types.String                                  `tfsdk:"monthly_leasing_price"`
	Registered                types.Bool                                    `tfsdk:"registered"`
	RegistrationID            types.String                                  `tfsdk:"registration_id"`
	SenderID                  types.String                                  `tfsdk:"sender_id"`
	SenderIDARN               types.String                                  `tfsdk:"arn"`
	Tags                      tftags.Map                                    `tfsdk:"tags"`
	TagsAll                   tftags.Map                                    `tfsdk:"tags_all"`
}

var (
	_ inttypes.ImportIDParser = senderIDImportID{}
)

type senderIDImportID struct{}

func (senderIDImportID) Parse(id string) (string, map[string]any, error) {
	senderID, isoCountryCode, found := strings.Cut(id, intflex.ResourceIdSeparator)
	if !found || senderID == "" || isoCountryCode == "" {
		return "", nil, smarterr.NewError(fmt.Errorf("unexpected format for ID (%[1]s), expected SenderID%[2]sISOCountryCode", id, intflex.ResourceIdSeparator))
	}

	return id, map[string]any{
		"sender_id":        senderID,
		"iso_country_code": isoCountryCode,
	}, nil
}

func (model *senderIDResourceModel) flatten(ctx context.Context, out *awstypes.SenderIdInformation) diag.Diagnostics {
	var diags diag.Diagnostics

	smerr.AddEnrich(ctx, &diags, fwflex.Flatten(ctx, out, model))
	if diags.HasError() {
		return diags
	}

	// DescribeSenderIds returns message types in mixed case (e.g. "Transactional"),
	// while the canonical enum values are upper case. Normalize so state matches the
	// configured value and stays stable across refresh and import.
	model.MessageTypes = normalizeMessageTypes(ctx, out.MessageTypes)

	return diags
}

// normalizeMessageTypes canonicalizes AWS message types to their upper case enum
// values. The DescribeSenderIds and RequestSenderId APIs are inconsistent about casing.
func normalizeMessageTypes(ctx context.Context, messageTypes []awstypes.MessageType) fwtypes.SetOfStringEnum[awstypes.MessageType] {
	values := make([]attr.Value, len(messageTypes))
	for i, mt := range messageTypes {
		values[i] = fwtypes.StringEnumValue(awstypes.MessageType(strings.ToUpper(string(mt))))
	}

	return fwtypes.NewSetValueOfMust[fwtypes.StringEnum[awstypes.MessageType]](ctx, values)
}

func findSenderIDByTwoPartKey(ctx context.Context, conn *pinpointsmsvoicev2.Client, senderID, isoCountryCode string) (*awstypes.SenderIdInformation, error) {
	input := pinpointsmsvoicev2.DescribeSenderIdsInput{
		SenderIds: []awstypes.SenderIdAndCountry{
			{
				SenderId:       aws.String(senderID),
				IsoCountryCode: aws.String(isoCountryCode),
			},
		},
	}

	return findSenderID(ctx, conn, &input)
}

func findSenderID(ctx context.Context, conn *pinpointsmsvoicev2.Client, input *pinpointsmsvoicev2.DescribeSenderIdsInput) (*awstypes.SenderIdInformation, error) {
	output, err := findSenderIDs(ctx, conn, input)

	if err != nil {
		return nil, smarterr.NewError(err)
	}

	return smarterr.Assert(tfresource.AssertSingleValueResult(output))
}

func findSenderIDs(ctx context.Context, conn *pinpointsmsvoicev2.Client, input *pinpointsmsvoicev2.DescribeSenderIdsInput) ([]awstypes.SenderIdInformation, error) {
	var output []awstypes.SenderIdInformation

	pages := pinpointsmsvoicev2.NewDescribeSenderIdsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return nil, smarterr.NewError(&retry.NotFoundError{
				LastError: err,
			})
		}

		if err != nil {
			return nil, smarterr.NewError(err)
		}

		output = append(output, page.SenderIds...)
	}

	return output, nil
}
