// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package pinpointsmsvoicev2

import (
	"context"
	"fmt"
	"strings"

	"github.com/YakDriver/regexache"
	"github.com/YakDriver/smarterr"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/pinpointsmsvoicev2"
	awstypes "github.com/aws/aws-sdk-go-v2/service/pinpointsmsvoicev2/types"
	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework-validators/mapvalidator"
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

// @FrameworkResource("aws_pinpointsmsvoicev2_protect_configuration", name="Protect Configuration")
// @Tags(identifierAttribute="arn")
// @IdentityAttribute("id")
// @Testing(hasNoPreExistingResource=true)
// @Testing(existsType="github.com/aws/aws-sdk-go-v2/service/pinpointsmsvoicev2/types;awstypes;awstypes.ProtectConfigurationInformation")
// @Testing(generator=false)
// @Testing(preCheck="testAccPreCheckProtectConfiguration")
func newProtectConfigurationResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	return &protectConfigurationResource{}, nil
}

// protectConfigurationNumberCapabilities are the number capabilities that have a country rule set.
// Each one maps to a separate *_country_rules attribute.
var protectConfigurationNumberCapabilities = []awstypes.NumberCapability{
	awstypes.NumberCapabilitySms,
	awstypes.NumberCapabilityVoice,
	awstypes.NumberCapabilityMms,
	awstypes.NumberCapabilityRcs,
}

// SMS supports every protect status. VOICE, MMS, and RCS support only ALLOW and BLOCK.
var (
	protectConfigurationSMSStatuses    = enum.EnumValues[awstypes.ProtectStatus]()
	protectConfigurationNonSMSStatuses = []awstypes.ProtectStatus{awstypes.ProtectStatusAllow, awstypes.ProtectStatusBlock}
)

var _ resource.ResourceWithValidateConfig = &protectConfigurationResource{}

type protectConfigurationResource struct {
	framework.ResourceWithModel[protectConfigurationResourceModel]
	framework.WithImportByIdentity
}

func (r *protectConfigurationResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var data protectConfigurationResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Config.Get(ctx, &data))
	if resp.Diagnostics.HasError() {
		return
	}

	for _, capability := range protectConfigurationNumberCapabilities {
		defaultStatus, rules := data.defaultStatus(capability), *data.countryRules(capability)
		if defaultStatus.IsUnknown() || rules.IsNull() || rules.IsUnknown() {
			continue
		}

		want := protectConfigurationEffectiveDefaultStatus(defaultStatus)
		for country, status := range rules.Elements() {
			if v, ok := status.(types.String); ok && !v.IsUnknown() && v.ValueString() == want {
				resp.Diagnostics.AddAttributeError(
					path.Root(protectConfigurationCountryRulesAttributeName(capability)).AtMapKey(country),
					"Invalid Country Rule",
					fmt.Sprintf("Status %q is the default status for %s. Remove the country instead.", want, capability),
				)
			}
		}
	}
}

func (r *protectConfigurationResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"account_default": schema.BoolAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			names.AttrARN: framework.ARNAttributeComputedOnly(),
			"created_timestamp": schema.StringAttribute{
				CustomType: timetypes.RFC3339Type{},
				Computed:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"deletion_protection_enabled": schema.BoolAttribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			names.AttrID:           framework.IDAttribute(),
			"mms_country_rules":    protectConfigurationCountryRulesAttribute(protectConfigurationNonSMSStatuses...),
			"mms_default_status":   protectConfigurationDefaultStatusAttribute(protectConfigurationNonSMSStatuses...),
			"rcs_country_rules":    protectConfigurationCountryRulesAttribute(protectConfigurationNonSMSStatuses...),
			"rcs_default_status":   protectConfigurationDefaultStatusAttribute(protectConfigurationNonSMSStatuses...),
			"sms_country_rules":    protectConfigurationCountryRulesAttribute(protectConfigurationSMSStatuses...),
			"sms_default_status":   protectConfigurationDefaultStatusAttribute(protectConfigurationSMSStatuses...),
			names.AttrTags:         tftags.TagsAttribute(),
			names.AttrTagsAll:      tftags.TagsAttributeComputedOnly(),
			"voice_country_rules":  protectConfigurationCountryRulesAttribute(protectConfigurationNonSMSStatuses...),
			"voice_default_status": protectConfigurationDefaultStatusAttribute(protectConfigurationNonSMSStatuses...),
		},
	}
}

// protectConfigurationDefaultStatusAttribute returns the protect status for every country of a number
// capability that isn't listed in its country rules. Null means ALLOW, the service default. It is
// Optional rather than Optional+Computed with a default so that state from earlier provider versions
// doesn't show a diff.
func protectConfigurationDefaultStatusAttribute(statuses ...awstypes.ProtectStatus) schema.StringAttribute {
	return schema.StringAttribute{
		Optional: true,
		Validators: []validator.String{
			stringvalidator.OneOf(enum.Slice(statuses...)...),
		},
	}
}

// protectConfigurationCountryRulesAttribute returns a map of ISO country code to protect status.
// Each country listed overrides the default status. A status equal to the default would read back as
// absent, so ValidateConfig rejects it. An empty map would read back as null, so at least one entry is required.
func protectConfigurationCountryRulesAttribute(statuses ...awstypes.ProtectStatus) schema.MapAttribute {
	return schema.MapAttribute{
		CustomType:  fwtypes.MapOfStringType,
		ElementType: types.StringType,
		Optional:    true,
		Validators: []validator.Map{
			mapvalidator.SizeAtLeast(1),
			mapvalidator.NoNullValues(),
			mapvalidator.KeysAre(stringvalidator.RegexMatches(regexache.MustCompile(`^[A-Z]{2}$`), "must be a two-character ISO 3166-1 alpha-2 country code in upper case")),
			mapvalidator.ValueStringsAre(stringvalidator.OneOf(enum.Slice(statuses...)...)),
		},
	}
}

func (r *protectConfigurationResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan protectConfigurationResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Plan.Get(ctx, &plan))
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().PinpointSMSVoiceV2Client(ctx)

	var input pinpointsmsvoicev2.CreateProtectConfigurationInput
	smerr.AddEnrich(ctx, &resp.Diagnostics, fwflex.Expand(ctx, plan, &input))
	if resp.Diagnostics.HasError() {
		return
	}
	input.ClientToken = aws.String(create.UniqueId(ctx))
	input.Tags = getTagsIn(ctx)

	// Deletion protection is enabled only after the country rules are set. If setting them fails,
	// the tainted configuration must stay deletable so that Terraform can replace it.
	deletionProtectionEnabled := aws.ToBool(input.DeletionProtectionEnabled)
	input.DeletionProtectionEnabled = nil

	out, err := conn.CreateProtectConfiguration(ctx, &input)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err)
		return
	}

	id := aws.ToString(out.ProtectConfigurationId)

	// Persist the ID so that a failure setting country rules taints the resource instead of orphaning it.
	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.SetAttribute(ctx, path.Root(names.AttrID), id))
	if resp.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, fwflex.Flatten(ctx, out, &plan, fwflex.WithFieldNamePrefix("ProtectConfiguration")), smerr.ID, id)
	if resp.Diagnostics.HasError() {
		return
	}

	for _, capability := range protectConfigurationNumberCapabilities {
		if err := syncProtectConfigurationCountryRuleSet(ctx, conn, id, capability, &plan); err != nil {
			smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, id)
			return
		}
	}

	if deletionProtectionEnabled {
		input := pinpointsmsvoicev2.UpdateProtectConfigurationInput{
			DeletionProtectionEnabled: aws.Bool(true),
			ProtectConfigurationId:    aws.String(id),
		}
		if _, err := conn.UpdateProtectConfiguration(ctx, &input); err != nil {
			smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, id)
			return
		}
		plan.DeletionProtectionEnabled = types.BoolValue(true)
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, plan))
}

func (r *protectConfigurationResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state protectConfigurationResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().PinpointSMSVoiceV2Client(ctx)

	id := fwflex.StringValueFromFramework(ctx, state.ID)
	out, err := findProtectConfigurationByID(ctx, conn, id)
	if retry.NotFound(err) {
		smerr.AddOne(ctx, &resp.Diagnostics, fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, id)
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, state.flatten(ctx, conn, out), smerr.ID, id)
	if resp.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &state))
}

func (r *protectConfigurationResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state protectConfigurationResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Plan.Get(ctx, &plan))
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().PinpointSMSVoiceV2Client(ctx)

	id := fwflex.StringValueFromFramework(ctx, state.ID)

	if !plan.DeletionProtectionEnabled.Equal(state.DeletionProtectionEnabled) {
		input := pinpointsmsvoicev2.UpdateProtectConfigurationInput{
			DeletionProtectionEnabled: fwflex.BoolFromFramework(ctx, plan.DeletionProtectionEnabled),
			ProtectConfigurationId:    aws.String(id),
		}

		if _, err := conn.UpdateProtectConfiguration(ctx, &input); err != nil {
			smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, id)
			return
		}
	}

	for _, capability := range protectConfigurationNumberCapabilities {
		if plan.defaultStatus(capability).Equal(state.defaultStatus(capability)) && plan.countryRules(capability).Equal(*state.countryRules(capability)) {
			continue
		}

		if err := syncProtectConfigurationCountryRuleSet(ctx, conn, id, capability, &plan); err != nil {
			smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, id)
			return
		}
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &plan))
}

func (r *protectConfigurationResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state protectConfigurationResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().PinpointSMSVoiceV2Client(ctx)

	id := fwflex.StringValueFromFramework(ctx, state.ID)

	// Deletion fails with a ConflictException while the protect configuration is associated with a
	// configuration set, is the account default, or has deletion protection enabled. The AWS error
	// names the reason, so it is returned as is rather than undoing those settings here.
	input := pinpointsmsvoicev2.DeleteProtectConfigurationInput{
		ProtectConfigurationId: aws.String(id),
	}
	_, err := conn.DeleteProtectConfiguration(ctx, &input)
	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return
	}
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, id)
		return
	}
}

// syncProtectConfigurationCountryRuleSet makes the country rule set for the number capability match the
// model: every country listed in the country rules gets its status, and every other country gets the
// default status. Only countries whose status differs are sent.
func syncProtectConfigurationCountryRuleSet(ctx context.Context, conn *pinpointsmsvoicev2.Client, id string, capability awstypes.NumberCapability, data *protectConfigurationResourceModel) error {
	current, err := findProtectConfigurationCountryRuleSet(ctx, conn, id, capability)
	if err != nil {
		return smarterr.NewError(err)
	}

	defaultStatus := protectConfigurationEffectiveDefaultStatus(data.defaultStatus(capability))
	rules := fwflex.ExpandFrameworkStringValueMap(ctx, *data.countryRules(capability))

	return updateProtectConfigurationCountryRuleSet(ctx, conn, id, capability, protectConfigurationCountryRuleSetUpdates(current, defaultStatus, rules))
}

// protectConfigurationCountryRuleSetUpdates returns the changes that turn the current country rule set
// into the desired one. Countries in rules that aren't in the current set are sent as is, so that AWS
// rejects countries the number capability doesn't support.
func protectConfigurationCountryRuleSetUpdates(current map[string]string, defaultStatus string, rules map[string]string) map[string]awstypes.ProtectConfigurationCountryRuleSetInformation {
	updates := make(map[string]awstypes.ProtectConfigurationCountryRuleSetInformation)
	for country, status := range current {
		want, ok := rules[country]
		if !ok {
			want = defaultStatus
		}
		if status != want {
			updates[country] = awstypes.ProtectConfigurationCountryRuleSetInformation{
				ProtectStatus: awstypes.ProtectStatus(want),
			}
		}
	}
	for country, status := range rules {
		if _, ok := current[country]; !ok {
			updates[country] = awstypes.ProtectConfigurationCountryRuleSetInformation{
				ProtectStatus: awstypes.ProtectStatus(status),
			}
		}
	}

	return updates
}

// protectConfigurationEffectiveDefaultStatus returns the default status, treating null as ALLOW.
func protectConfigurationEffectiveDefaultStatus(v types.String) string {
	if v.IsNull() {
		return string(awstypes.ProtectStatusAllow)
	}

	return v.ValueString()
}

func protectConfigurationCountryRulesAttributeName(capability awstypes.NumberCapability) string {
	return strings.ToLower(string(capability)) + "_country_rules"
}

func updateProtectConfigurationCountryRuleSet(ctx context.Context, conn *pinpointsmsvoicev2.Client, id string, capability awstypes.NumberCapability, updates map[string]awstypes.ProtectConfigurationCountryRuleSetInformation) error {
	if len(updates) == 0 {
		return nil
	}

	input := pinpointsmsvoicev2.UpdateProtectConfigurationCountryRuleSetInput{
		CountryRuleSetUpdates:  updates,
		NumberCapability:       capability,
		ProtectConfigurationId: aws.String(id),
	}
	_, err := conn.UpdateProtectConfigurationCountryRuleSet(ctx, &input)
	if err != nil {
		return smarterr.NewError(err)
	}

	return nil
}

func findProtectConfigurationByID(ctx context.Context, conn *pinpointsmsvoicev2.Client, id string) (*awstypes.ProtectConfigurationInformation, error) {
	input := pinpointsmsvoicev2.DescribeProtectConfigurationsInput{
		ProtectConfigurationIds: []string{id},
	}

	return findProtectConfiguration(ctx, conn, &input)
}

func findProtectConfiguration(ctx context.Context, conn *pinpointsmsvoicev2.Client, input *pinpointsmsvoicev2.DescribeProtectConfigurationsInput) (*awstypes.ProtectConfigurationInformation, error) {
	var output []awstypes.ProtectConfigurationInformation
	for v, err := range listProtectConfigurations(ctx, conn, input) {
		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return nil, smarterr.NewError(&retry.NotFoundError{
				LastError: err,
			})
		}
		if err != nil {
			return nil, smarterr.NewError(err)
		}

		output = append(output, v)
	}

	return smarterr.Assert(tfresource.AssertSingleValueResult(output))
}

// findProtectConfigurationCountryRuleSet returns the protect status of every country for the number capability.
func findProtectConfigurationCountryRuleSet(ctx context.Context, conn *pinpointsmsvoicev2.Client, id string, capability awstypes.NumberCapability) (map[string]string, error) {
	input := pinpointsmsvoicev2.GetProtectConfigurationCountryRuleSetInput{
		NumberCapability:       capability,
		ProtectConfigurationId: aws.String(id),
	}

	out, err := conn.GetProtectConfigurationCountryRuleSet(ctx, &input)
	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, smarterr.NewError(&retry.NotFoundError{
			LastError: err,
		})
	}
	if err != nil {
		return nil, smarterr.NewError(err)
	}

	if out == nil {
		return nil, smarterr.NewError(tfresource.NewEmptyResultError())
	}

	ruleSet := make(map[string]string, len(out.CountryRuleSet))
	for country, v := range out.CountryRuleSet {
		ruleSet[country] = string(v.ProtectStatus)
	}

	return ruleSet, nil
}

type protectConfigurationResourceModel struct {
	framework.WithRegionModel
	AccountDefault            types.Bool          `tfsdk:"account_default"`
	ARN                       types.String        `tfsdk:"arn"`
	CreatedTimestamp          timetypes.RFC3339   `tfsdk:"created_timestamp"`
	DeletionProtectionEnabled types.Bool          `tfsdk:"deletion_protection_enabled"`
	ID                        types.String        `tfsdk:"id"`
	MMSCountryRules           fwtypes.MapOfString `tfsdk:"mms_country_rules" autoflex:"-"`
	MMSDefaultStatus          types.String        `tfsdk:"mms_default_status" autoflex:"-"`
	RCSCountryRules           fwtypes.MapOfString `tfsdk:"rcs_country_rules" autoflex:"-"`
	RCSDefaultStatus          types.String        `tfsdk:"rcs_default_status" autoflex:"-"`
	SMSCountryRules           fwtypes.MapOfString `tfsdk:"sms_country_rules" autoflex:"-"`
	SMSDefaultStatus          types.String        `tfsdk:"sms_default_status" autoflex:"-"`
	Tags                      tftags.Map          `tfsdk:"tags"`
	TagsAll                   tftags.Map          `tfsdk:"tags_all"`
	VoiceCountryRules         fwtypes.MapOfString `tfsdk:"voice_country_rules" autoflex:"-"`
	VoiceDefaultStatus        types.String        `tfsdk:"voice_default_status" autoflex:"-"`
}

func (m *protectConfigurationResourceModel) countryRules(capability awstypes.NumberCapability) *fwtypes.MapOfString {
	switch capability {
	case awstypes.NumberCapabilityMms:
		return &m.MMSCountryRules
	case awstypes.NumberCapabilityRcs:
		return &m.RCSCountryRules
	case awstypes.NumberCapabilitySms:
		return &m.SMSCountryRules
	case awstypes.NumberCapabilityVoice:
		return &m.VoiceCountryRules
	default:
		return nil
	}
}

func (m *protectConfigurationResourceModel) defaultStatus(capability awstypes.NumberCapability) types.String {
	switch capability {
	case awstypes.NumberCapabilityMms:
		return m.MMSDefaultStatus
	case awstypes.NumberCapabilityRcs:
		return m.RCSDefaultStatus
	case awstypes.NumberCapabilitySms:
		return m.SMSDefaultStatus
	case awstypes.NumberCapabilityVoice:
		return m.VoiceDefaultStatus
	default:
		return types.StringNull()
	}
}

// flatten sets the country rules to the countries whose status differs from the default status.
// AWS doesn't store a default status, so the default status in the model is kept as is.
func (m *protectConfigurationResourceModel) flatten(ctx context.Context, conn *pinpointsmsvoicev2.Client, out *awstypes.ProtectConfigurationInformation) diag.Diagnostics {
	var diags diag.Diagnostics

	smerr.AddEnrich(ctx, &diags, fwflex.Flatten(ctx, out, m, fwflex.WithFieldNamePrefix("ProtectConfiguration")))
	if diags.HasError() {
		return diags
	}

	id := aws.ToString(out.ProtectConfigurationId)
	for _, capability := range protectConfigurationNumberCapabilities {
		ruleSet, err := findProtectConfigurationCountryRuleSet(ctx, conn, id, capability)
		if err != nil {
			smerr.AddError(ctx, &diags, err, smerr.ID, id)
			return diags
		}

		defaultStatus := protectConfigurationEffectiveDefaultStatus(m.defaultStatus(capability))
		rules := make(map[string]string)
		for country, status := range ruleSet {
			if status != defaultStatus {
				rules[country] = status
			}
		}

		*m.countryRules(capability) = fwflex.FlattenFrameworkStringValueMapOfString(ctx, rules)
	}

	return diags
}
