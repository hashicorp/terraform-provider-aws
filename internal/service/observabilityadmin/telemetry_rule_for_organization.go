// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package observabilityadmin

import (
	"context"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/observabilityadmin"
	awstypes "github.com/aws/aws-sdk-go-v2/service/observabilityadmin/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	fwvalidators "github.com/hashicorp/terraform-provider-aws/internal/framework/validators"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_observabilityadmin_telemetry_rule_for_organization", name="Telemetry Rule For Organization")
// @Tags(identifierAttribute="rule_arn")
// @Testing(existsType="github.com/aws/aws-sdk-go-v2/service/observabilityadmin;observabilityadmin;observabilityadmin.GetTelemetryRuleForOrganizationOutput")
// @Testing(preCheck="github.com/hashicorp/terraform-provider-aws/internal/acctest;acctest.PreCheckOrganizationManagementAccount")
// @Testing(preCheck="testAccTelemetryRuleForOrganizationPreCheck")
// @IdentityAttribute("rule_name")
// @Testing(tagsTest=false)
// @Testing(importStateIdAttribute="rule_name")
// @Testing(serialize=true)
// @Testing(hasNoPreExistingResource=true)
func newTelemetryRuleForOrganizationResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &telemetryRuleForOrganizationResource{}

	r.SetDefaultCreateTimeout(5 * time.Minute)
	r.SetDefaultUpdateTimeout(5 * time.Minute)
	r.SetDefaultDeleteTimeout(5 * time.Minute)

	return r, nil
}

type telemetryRuleForOrganizationResource struct {
	framework.ResourceWithModel[telemetryRuleForOrganizationResourceModel]
	framework.WithTimeouts
	framework.WithImportByIdentity
}

func (r *telemetryRuleForOrganizationResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"rule_arn": framework.ARNAttributeComputedOnly(),
			"rule_name": schema.StringAttribute{
				Required: true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 100),
					stringvalidator.RegexMatches(regexache.MustCompile(`^[0-9A-Za-z\-_.#/]+$`), ""),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
		},
		Blocks: map[string]schema.Block{
			names.AttrRule: schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[telemetryRuleModel](ctx),
				Validators: []validator.List{
					listvalidator.IsRequired(),
					listvalidator.SizeAtLeast(1),
					listvalidator.SizeAtMost(1),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"all_regions": schema.BoolAttribute{
							Optional: true,
						},
						"allow_field_updates": schema.BoolAttribute{
							Optional: true,
						},
						// regions is treated as an unordered set because the AWS API does
						// not preserve the input order on GET (returns sorted). Using a
						// SetAttribute avoids "provider produced inconsistent result"
						// errors when the planned order differs from the API-returned
						// order. Marked Computed with UseStateForUnknown so that when the
						// field is omitted from config, state retains the API-returned
						// value across refreshes.
						"regions": schema.SetAttribute{
							CustomType:  fwtypes.SetOfStringType,
							ElementType: types.StringType,
							Optional:    true,
							Computed:    true,
							Validators: []validator.Set{
								setvalidator.ValueStringsAre(fwvalidators.AWSRegion()),
							},
							PlanModifiers: []planmodifier.Set{
								setplanmodifier.UseStateForUnknown(),
							},
						},
						names.AttrResourceType: schema.StringAttribute{
							CustomType: fwtypes.StringEnumType[awstypes.ResourceType](),
							Optional:   true,
						},
						names.AttrScope: schema.StringAttribute{
							Optional: true,
						},
						"selection_criteria": schema.StringAttribute{
							Optional: true,
						},
						"telemetry_source_types": schema.ListAttribute{
							CustomType: fwtypes.ListOfStringEnumType[awstypes.TelemetrySourceType](),
							Optional:   true,
							Computed:   true,
							PlanModifiers: []planmodifier.List{
								listplanmodifier.UseStateForUnknown(),
							},
						},
						"telemetry_type": schema.StringAttribute{
							CustomType: fwtypes.StringEnumType[awstypes.TelemetryType](),
							Required:   true,
						},
					},
					Blocks: map[string]schema.Block{
						"destination_configuration": telemetryRuleDestinationConfigurationBlock(ctx),
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

func (r *telemetryRuleForOrganizationResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data telemetryRuleForOrganizationResourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.Plan.Get(ctx, &data))
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().ObservabilityAdminClient(ctx)

	ruleName := fwflex.StringValueFromFramework(ctx, data.RuleName)
	var input observabilityadmin.CreateTelemetryRuleForOrganizationInput
	smerr.AddEnrich(ctx, &response.Diagnostics, fwflex.Expand(ctx, data, &input))
	if response.Diagnostics.HasError() {
		return
	}

	// Additional fields.
	input.Tags = getTagsIn(ctx)

	output, err := conn.CreateTelemetryRuleForOrganization(ctx, &input)
	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, ruleName)
		return
	}

	// Set values for unknowns.
	data.RuleARN = fwflex.StringToFramework(ctx, output.RuleArn)

	// Read back to populate computed fields (e.g. API-derived telemetry_source_types,
	// regions when all_regions is true).
	out, err := findTelemetryRuleForOrganizationByName(ctx, conn, ruleName)
	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, ruleName)
		return
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, r.flatten(ctx, out, &data))
	if response.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, response.State.Set(ctx, data))
}

func (r *telemetryRuleForOrganizationResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data telemetryRuleForOrganizationResourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.State.Get(ctx, &data))
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().ObservabilityAdminClient(ctx)

	ruleName := fwflex.StringValueFromFramework(ctx, data.RuleName)
	output, err := findTelemetryRuleForOrganizationByName(ctx, conn, ruleName)
	if retry.NotFound(err) {
		smerr.AddOne(ctx, &response.Diagnostics, fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)
		return
	}

	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, ruleName)
		return
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, r.flatten(ctx, output, &data))
	if response.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, response.State.Set(ctx, &data))
}

func (r *telemetryRuleForOrganizationResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var new, old telemetryRuleForOrganizationResourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.Plan.Get(ctx, &new))
	smerr.AddEnrich(ctx, &response.Diagnostics, request.State.Get(ctx, &old))
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().ObservabilityAdminClient(ctx)

	diff, d := fwflex.Diff(ctx, new, old)
	smerr.AddEnrich(ctx, &response.Diagnostics, d)
	if response.Diagnostics.HasError() {
		return
	}

	if diff.HasChanges() {
		ruleName := fwflex.StringValueFromFramework(ctx, new.RuleName)
		var input observabilityadmin.UpdateTelemetryRuleForOrganizationInput
		smerr.AddEnrich(ctx, &response.Diagnostics, fwflex.Expand(ctx, new, &input))
		if response.Diagnostics.HasError() {
			return
		}

		// Additional fields.
		input.RuleIdentifier = aws.String(ruleName)

		_, err := conn.UpdateTelemetryRuleForOrganization(ctx, &input)
		if err != nil {
			smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, ruleName)
			return
		}

		// Read back to refresh computed fields after the update.
		out, err := findTelemetryRuleForOrganizationByName(ctx, conn, ruleName)
		if err != nil {
			smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, ruleName)
			return
		}

		smerr.AddEnrich(ctx, &response.Diagnostics, r.flatten(ctx, out, &new))
		if response.Diagnostics.HasError() {
			return
		}
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, response.State.Set(ctx, &new))
}

func (r *telemetryRuleForOrganizationResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data telemetryRuleForOrganizationResourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.State.Get(ctx, &data))
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().ObservabilityAdminClient(ctx)

	ruleName := fwflex.StringValueFromFramework(ctx, data.RuleName)
	input := observabilityadmin.DeleteTelemetryRuleForOrganizationInput{
		RuleIdentifier: aws.String(ruleName),
	}

	_, err := conn.DeleteTelemetryRuleForOrganization(ctx, &input)
	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return
	}
	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, ruleName)
		return
	}
}

func findTelemetryRuleForOrganizationByName(ctx context.Context, conn *observabilityadmin.Client, name string) (*observabilityadmin.GetTelemetryRuleForOrganizationOutput, error) {
	input := observabilityadmin.GetTelemetryRuleForOrganizationInput{
		RuleIdentifier: aws.String(name),
	}

	return findTelemetryRuleForOrganization(ctx, conn, &input)
}

func findTelemetryRuleForOrganization(ctx context.Context, conn *observabilityadmin.Client, input *observabilityadmin.GetTelemetryRuleForOrganizationInput) (*observabilityadmin.GetTelemetryRuleForOrganizationOutput, error) {
	output, err := conn.GetTelemetryRuleForOrganization(ctx, input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) || errs.IsAErrorMessageContains[*awstypes.ValidationException](err, "Telemetry evaluation is not enabled for the organization") {
		return nil, &retry.NotFoundError{
			LastError: err,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.TelemetryRule == nil {
		return nil, tfresource.NewEmptyResultError()
	}

	return output, nil
}

func (r *telemetryRuleForOrganizationResource) flatten(ctx context.Context, telemetryRule *observabilityadmin.GetTelemetryRuleForOrganizationOutput, data *telemetryRuleForOrganizationResourceModel) diag.Diagnostics {
	var diags diag.Diagnostics
	diags.Append(fwflex.Flatten(ctx, telemetryRule, data, fwflex.WithFieldNamePrefix("Telemetry"))...)
	return diags
}

type telemetryRuleForOrganizationResourceModel struct {
	framework.WithRegionModel
	Rule     fwtypes.ListNestedObjectValueOf[telemetryRuleModel] `tfsdk:"rule"`
	RuleARN  types.String                                        `tfsdk:"rule_arn"`
	RuleName types.String                                        `tfsdk:"rule_name"`
	Tags     tftags.Map                                          `tfsdk:"tags"`
	TagsAll  tftags.Map                                          `tfsdk:"tags_all"`
	Timeouts timeouts.Value                                      `tfsdk:"timeouts"`
}
