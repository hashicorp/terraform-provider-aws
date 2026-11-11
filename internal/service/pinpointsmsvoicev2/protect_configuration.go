// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package pinpointsmsvoicev2

import (
	"context"

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

type protectConfigurationResource struct {
	framework.ResourceWithModel[protectConfigurationResourceModel]
	framework.WithImportByIdentity
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
			names.AttrID:          framework.IDAttribute(),
			"mms_country_rules":   protectConfigurationCountryRulesAttribute(awstypes.ProtectStatusBlock),
			"rcs_country_rules":   protectConfigurationCountryRulesAttribute(awstypes.ProtectStatusBlock),
			"sms_country_rules":   protectConfigurationCountryRulesAttribute(awstypes.ProtectStatusBlock, awstypes.ProtectStatusMonitor, awstypes.ProtectStatusFilter),
			names.AttrTags:        tftags.TagsAttribute(),
			names.AttrTagsAll:     tftags.TagsAttributeComputedOnly(),
			"voice_country_rules": protectConfigurationCountryRulesAttribute(awstypes.ProtectStatusBlock),
		},
	}
}

// protectConfigurationCountryRulesAttribute returns a map of ISO country code to protect status.
// ALLOW is the service default for every country and is represented by omitting the country,
// so it isn't an accepted value. An empty map would read back as null, so at least one entry is required.
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
		updates := protectConfigurationCountryRuleSetUpdates(ctx, fwtypes.NewMapValueOfNull[types.String](ctx), *plan.countryRules(capability))
		if err := updateProtectConfigurationCountryRuleSet(ctx, conn, id, capability, updates); err != nil {
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
		updates := protectConfigurationCountryRuleSetUpdates(ctx, *state.countryRules(capability), *plan.countryRules(capability))
		if err := updateProtectConfigurationCountryRuleSet(ctx, conn, id, capability, updates); err != nil {
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

// protectConfigurationCountryRuleSetUpdates returns the changes that turn the old rules into the new ones.
// The update API merges, so a country removed from the rules is reset to ALLOW.
func protectConfigurationCountryRuleSetUpdates(ctx context.Context, old, new fwtypes.MapOfString) map[string]awstypes.ProtectConfigurationCountryRuleSetInformation {
	oldRules := fwflex.ExpandFrameworkStringValueMap(ctx, old)
	newRules := fwflex.ExpandFrameworkStringValueMap(ctx, new)

	updates := make(map[string]awstypes.ProtectConfigurationCountryRuleSetInformation)
	for country, status := range newRules {
		if oldRules[country] != status {
			updates[country] = awstypes.ProtectConfigurationCountryRuleSetInformation{
				ProtectStatus: awstypes.ProtectStatus(status),
			}
		}
	}
	for country := range oldRules {
		if _, ok := newRules[country]; !ok {
			updates[country] = awstypes.ProtectConfigurationCountryRuleSetInformation{
				ProtectStatus: awstypes.ProtectStatusAllow,
			}
		}
	}

	return updates
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

// findProtectConfigurationCountryRules returns the countries whose protect status for the number
// capability is not ALLOW, the default.
func findProtectConfigurationCountryRules(ctx context.Context, conn *pinpointsmsvoicev2.Client, id string, capability awstypes.NumberCapability) (map[string]string, error) {
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

	rules := make(map[string]string)
	for country, v := range out.CountryRuleSet {
		if v.ProtectStatus != awstypes.ProtectStatusAllow {
			rules[country] = string(v.ProtectStatus)
		}
	}

	return rules, nil
}

type protectConfigurationResourceModel struct {
	framework.WithRegionModel
	AccountDefault            types.Bool          `tfsdk:"account_default"`
	ARN                       types.String        `tfsdk:"arn"`
	CreatedTimestamp          timetypes.RFC3339   `tfsdk:"created_timestamp"`
	DeletionProtectionEnabled types.Bool          `tfsdk:"deletion_protection_enabled"`
	ID                        types.String        `tfsdk:"id"`
	MMSCountryRules           fwtypes.MapOfString `tfsdk:"mms_country_rules" autoflex:"-"`
	RCSCountryRules           fwtypes.MapOfString `tfsdk:"rcs_country_rules" autoflex:"-"`
	SMSCountryRules           fwtypes.MapOfString `tfsdk:"sms_country_rules" autoflex:"-"`
	Tags                      tftags.Map          `tfsdk:"tags"`
	TagsAll                   tftags.Map          `tfsdk:"tags_all"`
	VoiceCountryRules         fwtypes.MapOfString `tfsdk:"voice_country_rules" autoflex:"-"`
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

func (m *protectConfigurationResourceModel) flatten(ctx context.Context, conn *pinpointsmsvoicev2.Client, out *awstypes.ProtectConfigurationInformation) diag.Diagnostics {
	var diags diag.Diagnostics

	smerr.AddEnrich(ctx, &diags, fwflex.Flatten(ctx, out, m, fwflex.WithFieldNamePrefix("ProtectConfiguration")))
	if diags.HasError() {
		return diags
	}

	id := aws.ToString(out.ProtectConfigurationId)
	for _, capability := range protectConfigurationNumberCapabilities {
		rules, err := findProtectConfigurationCountryRules(ctx, conn, id, capability)
		if err != nil {
			smerr.AddError(ctx, &diags, err, smerr.ID, id)
			return diags
		}

		*m.countryRules(capability) = fwflex.FlattenFrameworkStringValueMapOfString(ctx, rules)
	}

	return diags
}
