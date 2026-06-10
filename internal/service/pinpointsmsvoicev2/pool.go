// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

package pinpointsmsvoicev2

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/aws/aws-sdk-go-v2/service/pinpointsmsvoicev2"
	awstypes "github.com/aws/aws-sdk-go-v2/service/pinpointsmsvoicev2/types"
	"github.com/hashicorp/aws-sdk-go-base/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_pinpointsmsvoicev2_pool", name="Pool")
// @Tags(identifierAttribute="arn")
// @IdentityAttribute("id")
// @Testing(hasNoPreExistingResource=true)
// @Testing(preCheck="testAccPreCheckPool")
// @Testing(existsType="github.com/aws/aws-sdk-go-v2/service/pinpointsmsvoicev2/types;awstypes.PoolInformation")
// @Testing(importIgnore="iso_country_code", plannableImportAction="Replace")
// @Testing(generator=false)
func newPoolResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &poolResource{}

	r.SetDefaultCreateTimeout(10 * time.Minute)
	r.SetDefaultUpdateTimeout(10 * time.Minute)
	r.SetDefaultDeleteTimeout(10 * time.Minute)

	return r, nil
}

type poolResource struct {
	framework.ResourceWithModel[poolResourceModel]
	framework.WithTimeouts
	framework.WithImportByIdentity
}

func (r *poolResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrARN: framework.ARNAttributeComputedOnly(),
			"deletion_protection_enabled": schema.BoolAttribute{
				Description: "Whether deletion protection is enabled. When `true`, the pool cannot be deleted.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			names.AttrID: framework.IDAttribute(),
			"iso_country_code": schema.StringAttribute{
				Description: "Two-character code, in ISO 3166-1 alpha-2 format, for the country or region of the pool. This field is optional for origination identity types that are not country-specific.",
				Optional:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.RegexMatches(regexache.MustCompile(`^[A-Z]{2}$`), "must be a two-character ISO 3166-1 alpha-2 country code in upper case"),
				},
			},
			"message_type": schema.StringAttribute{
				Description: "Type of message.",
				CustomType:  fwtypes.StringEnumType[awstypes.MessageType](),
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"opt_out_list_name": schema.StringAttribute{
				Description: "Name of the opt-out list to associate with the pool. Inherited from the initial origination identity when omitted.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"origination_identities": schema.SetAttribute{
				Description: "Set of origination identity ARNs to associate with the pool. At least one origination identity is required at creation.",
				CustomType:  fwtypes.SetOfStringType,
				ElementType: types.StringType,
				Required:    true,
				Validators: []validator.Set{
					setvalidator.SizeAtLeast(1),
				},
			},
			"self_managed_opt_outs_enabled": schema.BoolAttribute{
				Description: "Whether the pool relies on self-managed opt-out handling. When `false`, AWS auto-replies to HELP/STOP requests and manages the opt-out list. Inherited from the initial origination identity when omitted.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"shared_routes_enabled": schema.BoolAttribute{
				Description: "Whether shared routes are enabled for the pool. When `true`, messages may use shared phone numbers or sender IDs in countries that allow it.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
			"two_way_channel_arn": schema.StringAttribute{
				Description: "ARN of the two-way channel that receives inbound messages.",
				Optional:    true,
				Validators: []validator.String{
					stringvalidator.AlsoRequires(
						path.MatchRelative().AtParent().AtName("two_way_enabled"),
					),
				},
			},
			"two_way_channel_role": schema.StringAttribute{
				Description: "ARN of the IAM role that End User Messaging SMS assumes to publish inbound messages to the two-way channel.",
				CustomType:  fwtypes.ARNType,
				Optional:    true,
				Validators: []validator.String{
					stringvalidator.AlsoRequires(
						path.MatchRelative().AtParent().AtName("two_way_channel_arn"),
						path.MatchRelative().AtParent().AtName("two_way_enabled"),
					),
				},
			},
			"two_way_enabled": schema.BoolAttribute{
				Description: "Whether inbound message reception is enabled for the pool.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
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

func (r *poolResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := r.Meta().PinpointSMSVoiceV2Client(ctx)

	var plan poolResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Plan.Get(ctx, &plan))
	if resp.Diagnostics.HasError() {
		return
	}

	identities, d := sortOriginationIdentities(ctx, plan)
	smerr.AddEnrich(ctx, &resp.Diagnostics, d)
	if resp.Diagnostics.HasError() {
		return
	}

	// Validate the full origination_identities set against the pool's intended configuration before any AWS write.
	smerr.AddEnrich(ctx, &resp.Diagnostics, validateOriginationIdentities(ctx, conn, identities, plan.MessageType.ValueEnum(), plan.IsoCountryCode.ValueString()))
	if resp.Diagnostics.HasError() {
		return
	}

	// Deterministic seed selection: sort lexicographically and pick the first element.
	// The seed identity is opaque to the user; they always interact with `origination_identities` as a symmetric set.
	seed, extras := identities[0], identities[1:]

	input := &pinpointsmsvoicev2.CreatePoolInput{
		ClientToken:         aws.String(create.UniqueId(ctx)),
		OriginationIdentity: aws.String(seed),
		Tags:                getTagsIn(ctx),
	}
	smerr.AddEnrich(ctx, &resp.Diagnostics, fwflex.Expand(ctx, plan, input, fwflex.WithIgnoredFieldNamesAppend("OriginationIdentities")))
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := conn.CreatePool(ctx, input)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, seed)
		return
	}
	if out == nil || out.PoolId == nil {
		smerr.AddError(ctx, &resp.Diagnostics, fmt.Errorf("empty CreatePool output"), smerr.ID, seed)
		return
	}

	poolID := aws.ToString(out.PoolId)
	plan.PoolID = types.StringValue(poolID)

	// Check pool's ID into state immediately after CreatePool returns to avoid orphaning.
	// Any subsequent error leaves the pool in state so Delete can reach it
	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.SetAttribute(ctx, path.Root(names.AttrID), poolID))
	if resp.Diagnostics.HasError() {
		return
	}

	createTimeout := r.CreateTimeout(ctx, plan.Timeouts)
	pool, err := waitPoolActive(ctx, conn, poolID, createTimeout)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, poolID)
		return
	}

	isoCC := fwflex.StringFromFramework(ctx, plan.IsoCountryCode)
	if err := associateOriginationIdentities(ctx, conn, poolID, isoCC, extras...); err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, poolID)
		return
	}

	// Account for any update-based arguments set in the plan.
	if needsPostCreateUpdate(plan, pool) {
		var updateInput pinpointsmsvoicev2.UpdatePoolInput
		smerr.AddEnrich(ctx, &resp.Diagnostics, fwflex.Expand(ctx, plan, &updateInput, fwflex.WithIgnoredFieldNames([]string{"DeletionProtectionEnabled"})))
		if resp.Diagnostics.HasError() {
			return
		}

		if err := updatePoolWithIAMPropagation(ctx, conn, &updateInput); err != nil {
			smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, poolID)
			return
		}
		pool, err = waitPoolActive(ctx, conn, poolID, createTimeout)
		if err != nil {
			smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, poolID)
			return
		}
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, fwflex.Flatten(ctx, pool, &plan))
	if resp.Diagnostics.HasError() {
		return
	}

	associated, err := waitPoolOriginationIdentitiesConverged(ctx, conn, poolID, identities, createTimeout)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, poolID)
		return
	}
	plan.OriginationIdentities = fwflex.FlattenFrameworkStringValueSetOfString(ctx, associated)

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &plan))
}

func (r *poolResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().PinpointSMSVoiceV2Client(ctx)

	var state poolResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	pool, err := findPoolByID(ctx, conn, state.PoolID.ValueString())
	if retry.NotFound(err) {
		resp.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, state.PoolID.String())
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, fwflex.Flatten(ctx, pool, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	poolID := state.PoolID.ValueString()
	associated, err := findPoolOriginationIdentities(ctx, conn, poolID)
	if retry.NotFound(err) {
		resp.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, poolID)
		return
	}
	state.OriginationIdentities = fwflex.FlattenFrameworkStringValueSetOfString(ctx, associated)

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &state))
}

func (r *poolResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	conn := r.Meta().PinpointSMSVoiceV2Client(ctx)

	var plan, state poolResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Plan.Get(ctx, &plan))
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	poolID := plan.PoolID.ValueString()

	if !plan.OriginationIdentities.Equal(state.OriginationIdentities) {
		add, remove, d := diffOriginationIdentities(ctx, plan, state)
		smerr.AddEnrich(ctx, &resp.Diagnostics, d)
		if resp.Diagnostics.HasError() {
			return
		}

		if len(add) > 0 {
			smerr.AddEnrich(ctx, &resp.Diagnostics, validateOriginationIdentities(ctx, conn, add, plan.MessageType.ValueEnum(), plan.IsoCountryCode.ValueString()))
			if resp.Diagnostics.HasError() {
				return
			}
		}

		isoCC := fwflex.StringFromFramework(ctx, plan.IsoCountryCode)

		// Associate before disassociate to mitigate pool transiently having zero identities.
		if err := associateOriginationIdentities(ctx, conn, poolID, isoCC, add...); err != nil {
			smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, poolID)
			return
		}
		if err := disassociateOriginationIdentities(ctx, conn, poolID, isoCC, remove...); err != nil {
			smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, poolID)
			return
		}
	}

	diff, d := fwflex.Diff(ctx, plan, state, fwflex.WithIgnoredField("OriginationIdentities"))
	smerr.AddEnrich(ctx, &resp.Diagnostics, d)
	if resp.Diagnostics.HasError() {
		return
	}

	if diff.HasChanges() {
		input := &pinpointsmsvoicev2.UpdatePoolInput{
			PoolId: aws.String(poolID),
		}
		smerr.AddEnrich(ctx, &resp.Diagnostics, fwflex.Expand(ctx, plan, input, diff.IgnoredFieldNamesOpts()...))
		if resp.Diagnostics.HasError() {
			return
		}
		if err := updatePoolWithIAMPropagation(ctx, conn, input); err != nil {
			smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, poolID)
			return
		}
	}

	updateTimeout := r.UpdateTimeout(ctx, plan.Timeouts)
	pool, err := waitPoolActive(ctx, conn, poolID, updateTimeout)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, poolID)
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, fwflex.Flatten(ctx, pool, &plan))
	if resp.Diagnostics.HasError() {
		return
	}

	expected, d := sortOriginationIdentities(ctx, plan)
	smerr.AddEnrich(ctx, &resp.Diagnostics, d)
	if resp.Diagnostics.HasError() {
		return
	}

	associated, err := waitPoolOriginationIdentitiesConverged(ctx, conn, poolID, expected, updateTimeout)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, poolID)
		return
	}
	plan.OriginationIdentities = fwflex.FlattenFrameworkStringValueSetOfString(ctx, associated)

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &plan))
}

func (r *poolResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := r.Meta().PinpointSMSVoiceV2Client(ctx)

	var state poolResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	input := &pinpointsmsvoicev2.DeletePoolInput{
		PoolId: state.PoolID.ValueStringPointer(),
	}

	_, err := conn.DeletePool(ctx, input)
	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return
	}

	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, state.PoolID.ValueString())
		return
	}

	if _, err := waitPoolDeleted(ctx, conn, state.PoolID.ValueString(), r.DeleteTimeout(ctx, state.Timeouts)); err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, state.PoolID.ValueString())
		return
	}
}

func sortOriginationIdentities(ctx context.Context, model poolResourceModel) ([]string, diag.Diagnostics) {
	var diags diag.Diagnostics
	var identities []string
	diags.Append(model.OriginationIdentities.ElementsAs(ctx, &identities, false)...)
	if diags.HasError() {
		return nil, diags
	}
	slices.Sort(identities)
	return identities, diags
}

func diffOriginationIdentities(ctx context.Context, plan, state poolResourceModel) (add, remove []string, diags diag.Diagnostics) {
	// get identities sorted
	planSet, d := sortOriginationIdentities(ctx, plan)
	diags.Append(d...)
	if diags.HasError() {
		return nil, nil, diags
	}
	var stateSet []string
	diags.Append(state.OriginationIdentities.ElementsAs(ctx, &stateSet, false)...)
	if diags.HasError() {
		return nil, nil, diags
	}
	slices.Sort(stateSet)

	// build maps for lookup
	planMap := make(map[string]struct{}, len(planSet))
	for _, v := range planSet {
		planMap[v] = struct{}{}
	}
	stateMap := make(map[string]struct{}, len(stateSet))
	for _, v := range stateSet {
		stateMap[v] = struct{}{}
	}

	// add identities in the plan but not in state
	for _, v := range planSet {
		if _, ok := stateMap[v]; !ok {
			add = append(add, v)
		}
	}
	// remove identities in state but not in the plan
	for _, v := range stateSet {
		if _, ok := planMap[v]; !ok {
			remove = append(remove, v)
		}
	}

	return add, remove, diags
}

// order-independent equality test; no duplicates expected
func originationIdentitiesEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	set := make(map[string]struct{}, len(a))
	for _, v := range a {
		set[v] = struct{}{}
	}
	for _, v := range b {
		if _, ok := set[v]; !ok {
			return false
		}
	}
	return true
}

func needsPostCreateUpdate(plan poolResourceModel, pool *awstypes.PoolInformation) bool {
	if !plan.OptOutListName.IsNull() && plan.OptOutListName.ValueString() != aws.ToString(pool.OptOutListName) {
		return true
	}
	if !plan.SelfManagedOptOutsEnabled.IsNull() && plan.SelfManagedOptOutsEnabled.ValueBool() != pool.SelfManagedOptOutsEnabled {
		return true
	}
	if !plan.SharedRoutesEnabled.IsNull() && plan.SharedRoutesEnabled.ValueBool() != pool.SharedRoutesEnabled {
		return true
	}
	if !plan.TwoWayEnabled.IsNull() && plan.TwoWayEnabled.ValueBool() != pool.TwoWayEnabled {
		return true
	}
	if !plan.TwoWayChannelARN.IsNull() && plan.TwoWayChannelARN.ValueString() != aws.ToString(pool.TwoWayChannelArn) {
		return true
	}
	if !plan.TwoWayChannelRole.IsNull() && plan.TwoWayChannelRole.ValueString() != aws.ToString(pool.TwoWayChannelRole) {
		return true
	}
	return false
}

func updatePoolWithIAMPropagation(ctx context.Context, conn *pinpointsmsvoicev2.Client, input *pinpointsmsvoicev2.UpdatePoolInput) error {
	_, err := tfresource.RetryWhen(ctx, iamPropagationTimeout,
		func(ctx context.Context) (*pinpointsmsvoicev2.UpdatePoolOutput, error) {
			return conn.UpdatePool(ctx, input)
		},
		func(err error) (bool, error) {
			if tfawserr.ErrMessageContains(err, "ValidationException", "Could not assume IAM role") ||
				tfawserr.ErrMessageContains(err, "ValidationException", "RESOURCE_NOT_ACCESSIBLE") {
				return true, err
			}
			return false, err
		},
	)
	return err
}

func validateOriginationIdentities(ctx context.Context, conn *pinpointsmsvoicev2.Client, identities []string, intendedMessageType awstypes.MessageType, intendedIsoCountryCode string) diag.Diagnostics {
	var diags diag.Diagnostics

	if len(identities) == 0 {
		return diags
	}

	var phoneARNs []string
	var senderRefs []awstypes.SenderIdAndCountry
	unknownARN := map[string]struct{}{}

	// Parse and group by resource type
	for _, identityARN := range identities {
		parsed, err := arn.Parse(identityARN)
		if err != nil {
			unknownARN[identityARN] = struct{}{}
			continue
		}

		parts := strings.SplitN(parsed.Resource, "/", 3)
		switch parts[0] {
		case "phone-number":
			phoneARNs = append(phoneARNs, identityARN)
		case "sender-id":
			if len(parts) >= 3 {
				senderRefs = append(senderRefs, awstypes.SenderIdAndCountry{
					SenderId:       aws.String(parts[1]),
					IsoCountryCode: aws.String(parts[2]),
				})
			} else {
				unknownARN[identityARN] = struct{}{}
			}
		default:
			unknownARN[identityARN] = struct{}{}
		}
	}

	// Pass the full list of ARNs to avoid N DescribePhoneNumbers calls.
	// Surface errors as one diagnostic rather than attempting per-identity recovery.
	phoneByARN := map[string]awstypes.PhoneNumberInformation{}
	if len(phoneARNs) > 0 {
		pages := pinpointsmsvoicev2.NewDescribePhoneNumbersPaginator(conn, &pinpointsmsvoicev2.DescribePhoneNumbersInput{
			PhoneNumberIds: phoneARNs,
		})
		for pages.HasMorePages() {
			page, err := pages.NextPage(ctx)

			if errs.IsA[*awstypes.ResourceNotFoundException](err) {
				diags.AddError(
					"Invalid origination identity",
					"one or more phone-number origination identities do not exist in the configured AWS account and region. Verify every phone-number ARN in `origination_identities` refers to an existing aws_pinpointsmsvoicev2_phone_number in this region.",
				)
				return diags
			}
			if err != nil {
				diags.AddError(
					"reading phone-number origination identities",
					fmt.Sprintf("DescribePhoneNumbers failed: %s", err),
				)
				return diags
			}

			// Capture PhoneNumberInformation for each ARN.
			for _, p := range page.PhoneNumbers {
				phoneByARN[aws.ToString(p.PhoneNumberArn)] = p
			}
		}
	}

	// Pass the full list of ARNs to avoid N DescribeSenderIds calls.
	// Surface errors as one diagnostic rather than attempting per-identity recovery.
	senderByARN := map[string]awstypes.SenderIdInformation{}
	if len(senderRefs) > 0 {
		pages := pinpointsmsvoicev2.NewDescribeSenderIdsPaginator(conn, &pinpointsmsvoicev2.DescribeSenderIdsInput{
			SenderIds: senderRefs,
		})
		for pages.HasMorePages() {
			page, err := pages.NextPage(ctx)

			if errs.IsA[*awstypes.ResourceNotFoundException](err) {
				diags.AddError(
					"Invalid origination identity",
					"one or more sender-id origination identities do not exist in the configured AWS account and region. Verify every sender-id ARN in `origination_identities` refers to an existing sender ID in this region.",
				)
				return diags
			}
			if err != nil {
				diags.AddError(
					"reading sender-id origination identities",
					fmt.Sprintf("DescribeSenderIds failed: %s", err),
				)
				return diags
			}

			// Capture SenderIdInformation for each ARN.
			for _, s := range page.SenderIds {
				senderByARN[aws.ToString(s.SenderIdArn)] = s
			}
		}
	}

	// Validation checklist for all identities by resource type
	for _, identityARN := range identities {
		if _, ok := unknownARN[identityARN]; ok {
			// Fail-open. We don't recognize the ARN shape; defer to AWS to reject.
			continue
		}

		if p, ok := phoneByARN[identityARN]; ok {
			if p.Status != awstypes.NumberStatusActive {
				diags.AddError(
					fmt.Sprintf("origination identity %s is not ACTIVE", identityARN),
					fmt.Sprintf("the pool requires identities in ACTIVE state; the phone number is currently %q. Ensure the upstream resource has reached ACTIVE status before referencing it from the pool.",
						string(p.Status)),
				)
				continue
			}
			if p.MessageType != intendedMessageType {
				diags.AddError(
					fmt.Sprintf("origination identity %s has mismatched message_type", identityARN),
					fmt.Sprintf("the pool's message_type is %q; the phone number's message_type is %q. All identities in a pool must share the same message_type.",
						string(intendedMessageType), string(p.MessageType)),
				)
			}
			if intendedIsoCountryCode != "" && aws.ToString(p.IsoCountryCode) != intendedIsoCountryCode {
				diags.AddError(
					fmt.Sprintf("origination identity %s has mismatched iso_country_code", identityARN),
					fmt.Sprintf("the pool's iso_country_code is %q; the phone number's iso_country_code is %q. All identities in a pool must share the same country.",
						intendedIsoCountryCode, aws.ToString(p.IsoCountryCode)),
				)
			}
			continue
		}

		if s, ok := senderByARN[identityARN]; ok {
			if !slices.Contains(s.MessageTypes, intendedMessageType) {
				diags.AddError(
					fmt.Sprintf("origination identity %s does not support message_type=%q", identityARN, string(intendedMessageType)),
					fmt.Sprintf("the sender ID supports message types %v; the pool's intended message_type is %q.",
						s.MessageTypes, string(intendedMessageType)),
				)
			}
			if intendedIsoCountryCode != "" && aws.ToString(s.IsoCountryCode) != intendedIsoCountryCode {
				diags.AddError(
					fmt.Sprintf("origination identity %s has mismatched iso_country_code", identityARN),
					fmt.Sprintf("the pool's iso_country_code is %q; the sender ID's iso_country_code is %q.",
						intendedIsoCountryCode, aws.ToString(s.IsoCountryCode)),
				)
			}
			continue
		}

		// ARN shape valid but could not be found
		diags.AddError(
			fmt.Sprintf("origination identity %s not found", identityARN),
			"the identity does not exist in the configured AWS account/region. Ensure the upstream resource is fully provisioned before referencing it from the pool.",
		)
	}

	return diags
}

func associateOriginationIdentities(ctx context.Context, conn *pinpointsmsvoicev2.Client, poolID string, isoCountryCode *string, identities ...string) error {
	for _, identity := range identities {
		input := &pinpointsmsvoicev2.AssociateOriginationIdentityInput{
			ClientToken:         aws.String(create.UniqueId(ctx)),
			PoolId:              aws.String(poolID),
			OriginationIdentity: aws.String(identity),
			IsoCountryCode:      isoCountryCode,
		}

		// AssociateOriginationIdentity call is rate-limited (1 RPS per Pool ops); wrap to handle throttles.
		_, err := tfresource.RetryWhenIsA[*pinpointsmsvoicev2.AssociateOriginationIdentityOutput, *awstypes.ThrottlingException](
			ctx, 2*time.Minute,
			func(ctx context.Context) (*pinpointsmsvoicev2.AssociateOriginationIdentityOutput, error) {
				return conn.AssociateOriginationIdentity(ctx, input)
			},
		)
		if err != nil {
			return fmt.Errorf("associating origination identity %s: %w", identity, err)
		}
	}
	return nil
}

func disassociateOriginationIdentities(ctx context.Context, conn *pinpointsmsvoicev2.Client, poolID string, isoCountryCode *string, identities ...string) error {
	for _, identity := range identities {
		input := &pinpointsmsvoicev2.DisassociateOriginationIdentityInput{
			ClientToken:         aws.String(create.UniqueId(ctx)),
			PoolId:              aws.String(poolID),
			OriginationIdentity: aws.String(identity),
			IsoCountryCode:      isoCountryCode,
		}
		_, err := tfresource.RetryWhenIsA[*pinpointsmsvoicev2.DisassociateOriginationIdentityOutput, *awstypes.ThrottlingException](
			ctx, 2*time.Minute,
			func(ctx context.Context) (*pinpointsmsvoicev2.DisassociateOriginationIdentityOutput, error) {
				return conn.DisassociateOriginationIdentity(ctx, input)
			},
		)
		// Tolerate ResourceNotFoundException and two ConflictExceptions
		// that imply the desired post-condition is already satisfied.
		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			continue
		}
		if ce, ok := errors.AsType[*awstypes.ConflictException](err); ok &&
			(ce.Reason == awstypes.ConflictExceptionReasonPhoneNumberNotAssociatedToPool ||
				ce.Reason == awstypes.ConflictExceptionReasonLastPhoneNumber) {
			continue
		}
		if err != nil {
			return fmt.Errorf("disassociating origination identity %s: %w", identity, err)
		}
	}
	return nil
}

func waitPoolActive(ctx context.Context, conn *pinpointsmsvoicev2.Client, id string, timeout time.Duration) (*awstypes.PoolInformation, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.PoolStatusCreating),
		Target:  enum.Slice(awstypes.PoolStatusActive),
		Refresh: statusPool(conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if output, ok := outputRaw.(*awstypes.PoolInformation); ok {
		return output, err
	}
	return nil, err
}

func waitPoolDeleted(ctx context.Context, conn *pinpointsmsvoicev2.Client, id string, timeout time.Duration) (*awstypes.PoolInformation, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.PoolStatusDeleting, awstypes.PoolStatusActive),
		Target:  []string{},
		Refresh: statusPool(conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if output, ok := outputRaw.(*awstypes.PoolInformation); ok {
		return output, err
	}
	return nil, err
}

type originationIdentitiesStatus string

const (
	originationIdentitiesStatusConverging originationIdentitiesStatus = "CONVERGING"
	originationIdentitiesStatusConverged  originationIdentitiesStatus = "CONVERGED"
)

func waitPoolOriginationIdentitiesConverged(ctx context.Context, conn *pinpointsmsvoicev2.Client, poolID string, expected []string, timeout time.Duration) ([]string, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(originationIdentitiesStatusConverging),
		Target:  enum.Slice(originationIdentitiesStatusConverged),
		Refresh: func(ctx context.Context) (any, string, error) {
			got, err := findPoolOriginationIdentities(ctx, conn, poolID)
			if err != nil {
				return nil, "", err
			}
			if !originationIdentitiesEqual(got, expected) {
				return got, string(originationIdentitiesStatusConverging), nil
			}
			return got, string(originationIdentitiesStatusConverged), nil
		},
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if output, ok := outputRaw.([]string); ok {
		return output, err
	}
	return nil, err
}

func statusPool(conn *pinpointsmsvoicev2.Client, id string) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		output, err := findPoolByID(ctx, conn, id)
		if retry.NotFound(err) {
			return nil, "", nil
		}
		if err != nil {
			return nil, "", err
		}
		return output, string(output.Status), nil
	}
}

func findPoolByID(ctx context.Context, conn *pinpointsmsvoicev2.Client, id string) (*awstypes.PoolInformation, error) {
	input := &pinpointsmsvoicev2.DescribePoolsInput{
		PoolIds: []string{id},
	}

	output, err := findPool(ctx, conn, input)
	if err != nil {
		return nil, err
	}

	return output, nil
}

func findPool(ctx context.Context, conn *pinpointsmsvoicev2.Client, input *pinpointsmsvoicev2.DescribePoolsInput) (*awstypes.PoolInformation, error) {
	output, err := findPools(ctx, conn, input)
	if err != nil {
		return nil, err
	}
	return tfresource.AssertSingleValueResult(output)
}

func findPools(ctx context.Context, conn *pinpointsmsvoicev2.Client, input *pinpointsmsvoicev2.DescribePoolsInput) ([]awstypes.PoolInformation, error) {
	var output []awstypes.PoolInformation

	pages := pinpointsmsvoicev2.NewDescribePoolsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return nil, &retry.NotFoundError{LastError: err}
		}
		if err != nil {
			return nil, err
		}

		output = append(output, page.Pools...)
	}

	return output, nil
}

func findPoolOriginationIdentities(ctx context.Context, conn *pinpointsmsvoicev2.Client, poolID string) ([]string, error) {
	var arns []string

	pages := pinpointsmsvoicev2.NewListPoolOriginationIdentitiesPaginator(conn, &pinpointsmsvoicev2.ListPoolOriginationIdentitiesInput{
		PoolId: aws.String(poolID),
	})
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return nil, &retry.NotFoundError{LastError: err}
		}
		if err != nil {
			return nil, err
		}
		for _, identity := range page.OriginationIdentities {
			arns = append(arns, aws.ToString(identity.OriginationIdentityArn))
		}
	}

	return arns, nil
}

type poolResourceModel struct {
	framework.WithRegionModel
	DeletionProtectionEnabled types.Bool                               `tfsdk:"deletion_protection_enabled"`
	IsoCountryCode            types.String                             `tfsdk:"iso_country_code"`
	MessageType               fwtypes.StringEnum[awstypes.MessageType] `tfsdk:"message_type"`
	OptOutListName            types.String                             `tfsdk:"opt_out_list_name"`
	OriginationIdentities     fwtypes.SetOfString                      `tfsdk:"origination_identities"`
	PoolARN                   types.String                             `tfsdk:"arn"`
	PoolID                    types.String                             `tfsdk:"id"`
	SelfManagedOptOutsEnabled types.Bool                               `tfsdk:"self_managed_opt_outs_enabled"`
	SharedRoutesEnabled       types.Bool                               `tfsdk:"shared_routes_enabled"`
	Tags                      tftags.Map                               `tfsdk:"tags"`
	TagsAll                   tftags.Map                               `tfsdk:"tags_all"`
	Timeouts                  timeouts.Value                           `tfsdk:"timeouts"`
	TwoWayChannelARN          types.String                             `tfsdk:"two_way_channel_arn"`
	TwoWayChannelRole         fwtypes.ARN                              `tfsdk:"two_way_channel_role"`
	TwoWayEnabled             types.Bool                               `tfsdk:"two_way_enabled"`
}
