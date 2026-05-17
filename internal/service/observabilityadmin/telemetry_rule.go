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
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
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

// @FrameworkResource("aws_observabilityadmin_telemetry_rule", name="Telemetry Rule")
// @Tags(identifierAttribute="rule_arn")
// @IdentityAttribute("rule_name")
// @Testing(tagsTest=false)
// @Testing(hasNoPreExistingResource=true)
// @Testing(importStateIdAttribute="rule_name")
// @Testing(serialize=true)
func newTelemetryRuleResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &telemetryRuleResource{}

	r.SetDefaultCreateTimeout(5 * time.Minute)
	r.SetDefaultUpdateTimeout(5 * time.Minute)
	r.SetDefaultDeleteTimeout(5 * time.Minute)

	return r, nil
}

type telemetryRuleResource struct {
	framework.ResourceWithModel[telemetryRuleResourceModel]
	framework.WithTimeouts
	framework.WithImportByIdentity
}

func (r *telemetryRuleResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
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
						names.AttrResourceType: schema.StringAttribute{
							CustomType: fwtypes.StringEnumType[awstypes.ResourceType](),
							Optional:   true,
						},
						"telemetry_type": schema.StringAttribute{
							CustomType: fwtypes.StringEnumType[awstypes.TelemetryType](),
							Required:   true,
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

func (r *telemetryRuleResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data telemetryRuleResourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.Plan.Get(ctx, &data))
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().ObservabilityAdminClient(ctx)

	ruleName := fwflex.StringValueFromFramework(ctx, data.RuleName)
	var input observabilityadmin.CreateTelemetryRuleInput
	smerr.AddEnrich(ctx, &response.Diagnostics, fwflex.Expand(ctx, data, &input))
	if response.Diagnostics.HasError() {
		return
	}

	// Additional fields.
	input.Tags = getTagsIn(ctx)

	output, err := conn.CreateTelemetryRule(ctx, &input)
	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, ruleName)
		return
	}

	// Set values for unknowns.
	data.RuleARN = fwflex.StringToFramework(ctx, output.RuleArn)

	smerr.AddEnrich(ctx, &response.Diagnostics, response.State.Set(ctx, data))
}

func (r *telemetryRuleResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data telemetryRuleResourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.State.Get(ctx, &data))
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().ObservabilityAdminClient(ctx)

	ruleName := fwflex.StringValueFromFramework(ctx, data.RuleName)
	out, err := findTelemetryRuleByName(ctx, conn, ruleName)
	if retry.NotFound(err) {
		smerr.AddOne(ctx, &response.Diagnostics, fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)
		return
	}

	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, ruleName)
		return
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, r.flatten(ctx, out, &data))
	if response.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, response.State.Set(ctx, &data))
}

func (r *telemetryRuleResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var new, old telemetryRuleResourceModel
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
		var input observabilityadmin.UpdateTelemetryRuleInput
		smerr.AddEnrich(ctx, &response.Diagnostics, fwflex.Expand(ctx, new, &input))
		if response.Diagnostics.HasError() {
			return
		}

		// Additional fields.
		input.RuleIdentifier = aws.String(ruleName)

		_, err := conn.UpdateTelemetryRule(ctx, &input)
		if err != nil {
			smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, ruleName)
			return
		}
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, response.State.Set(ctx, &new))
}

func (r *telemetryRuleResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data telemetryRuleResourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.State.Get(ctx, &data))
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().ObservabilityAdminClient(ctx)

	ruleName := fwflex.StringValueFromFramework(ctx, data.RuleName)
	input := observabilityadmin.DeleteTelemetryRuleInput{
		RuleIdentifier: aws.String(ruleName),
	}

	_, err := conn.DeleteTelemetryRule(ctx, &input)
	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return
	}
	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, ruleName)
		return
	}
}

func (r *telemetryRuleResource) flatten(ctx context.Context, telemetryRule *observabilityadmin.GetTelemetryRuleOutput, data *telemetryRuleResourceModel) diag.Diagnostics {
	var diags diag.Diagnostics
	diags.Append(fwflex.Flatten(ctx, telemetryRule, data, fwflex.WithFieldNamePrefix("Telemetry"))...)
	return diags
}

func findTelemetryRuleByName(ctx context.Context, conn *observabilityadmin.Client, name string) (*observabilityadmin.GetTelemetryRuleOutput, error) {
	input := observabilityadmin.GetTelemetryRuleInput{
		RuleIdentifier: aws.String(name),
	}

	return findTelemetryRule(ctx, conn, &input)
}

func findTelemetryRule(ctx context.Context, conn *observabilityadmin.Client, input *observabilityadmin.GetTelemetryRuleInput) (*observabilityadmin.GetTelemetryRuleOutput, error) {
	output, err := conn.GetTelemetryRule(ctx, input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) || errs.IsAErrorMessageContains[*awstypes.ValidationException](err, "Telemetry evaluation is not enabled") {
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

type telemetryRuleResourceModel struct {
	framework.WithRegionModel
	Rule     fwtypes.ListNestedObjectValueOf[telemetryRuleModel] `tfsdk:"rule"`
	RuleARN  types.String                                        `tfsdk:"rule_arn"`
	RuleName types.String                                        `tfsdk:"rule_name"`
	Tags     tftags.Map                                          `tfsdk:"tags"`
	TagsAll  tftags.Map                                          `tfsdk:"tags_all"`
	Timeouts timeouts.Value                                      `tfsdk:"timeouts"`
}

type telemetryRuleModel struct {
	ResourceType  fwtypes.StringEnum[awstypes.ResourceType]  `tfsdk:"resource_type"`
	TelemetryType fwtypes.StringEnum[awstypes.TelemetryType] `tfsdk:"telemetry_type"`
}
