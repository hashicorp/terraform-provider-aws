// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

package cloudwatch

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/YakDriver/smarterr"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatch"
	awstypes "github.com/aws/aws-sdk-go-v2/service/cloudwatch/types"
	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework-validators/helpers/validatordiag"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_cloudwatch_alarm_mute_rule", name="Alarm Mute Rule")
// @Tags(identifierAttribute="arn")
// @Testing(importStateIdAttribute="name")
// @Testing(existsType="github.com/aws/aws-sdk-go-v2/service/cloudwatch;cloudwatch.GetAlarmMuteRuleOutput")
func newAlarmMuteRuleResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &alarmMuteRuleResource{}

	return r, nil
}

const (
	ResNameAlarmMuteRule = "Alarm Mute Rule"
)

type alarmMuteRuleResource struct {
	framework.ResourceWithModel[alarmMuteRuleResourceModel]
}

func (r *alarmMuteRuleResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrARN: framework.ARNAttributeComputedOnly(),
			names.AttrName: schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					// https://docs.aws.amazon.com/AmazonCloudWatch/latest/APIReference/API_PutAlarmMuteRule.html#API_PutAlarmMuteRule_RequestParameters
					stringvalidator.LengthBetween(1, 255),
				},
			},
			names.AttrDescription: schema.StringAttribute{
				Optional: true,
				Validators: []validator.String{
					// https://docs.aws.amazon.com/AmazonCloudWatch/latest/APIReference/API_PutAlarmMuteRule.html#API_PutAlarmMuteRule_RequestParameters
					stringvalidator.LengthAtMost(1024),
				},
			},
			"last_updated_timestamp": schema.StringAttribute{
				CustomType: timetypes.RFC3339Type{},
				Computed:   true,
			},
			"mute_type": schema.StringAttribute{
				Computed: true,
			},
			"start_date": schema.StringAttribute{
				CustomType: timetypes.RFC3339Type{},
				Optional:   true,
				Validators: []validator.String{
					validateTimeMinutePrecision(),
				},
			},
			"expire_date": schema.StringAttribute{
				CustomType: timetypes.RFC3339Type{},
				Optional:   true,
				Validators: []validator.String{
					validateTimeMinutePrecision(),
				},
			},
			names.AttrStatus: schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.AlarmMuteRuleStatus](),
				Computed:   true,
			},
			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
		},
		Blocks: map[string]schema.Block{
			"rule": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[ruleModel](ctx),
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
					listvalidator.SizeAtLeast(1),
				},
				NestedObject: schema.NestedBlockObject{
					Blocks: map[string]schema.Block{
						"schedule": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[scheduleModel](ctx),
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
								listvalidator.SizeAtLeast(1),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"duration": schema.StringAttribute{
										Required: true,
										Validators: []validator.String{
											stringvalidator.LengthBetween(1, 50),
										},
									},
									"expression": schema.StringAttribute{
										Required: true,
										Validators: []validator.String{
											stringvalidator.LengthBetween(1, 256),
										},
									},
									"timezone": schema.StringAttribute{
										Optional: true,
										Computed: true,
										Validators: []validator.String{
											stringvalidator.LengthBetween(1, 50),
										},
									},
								},
							},
						},
					},
				},
			},
			"mute_targets": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[muteTargetsModel](ctx),
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"alarm_names": schema.ListAttribute{
							ElementType: types.StringType,
							Required:    true,
							Validators: []validator.List{
								// https://docs.aws.amazon.com/AmazonCloudWatch/latest/monitoring/alarm-mute-rules.html#defining-alarm-mute-rules
								listvalidator.SizeAtMost(100),
							},
						},
					},
				},
			},
		},
	}
}

func (r *alarmMuteRuleResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := r.Meta().CloudWatchClient(ctx)

	var plan alarmMuteRuleResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Plan.Get(ctx, &plan))
	if resp.Diagnostics.HasError() {
		return
	}

	var input cloudwatch.PutAlarmMuteRuleInput
	smerr.AddEnrich(ctx, &resp.Diagnostics, flex.Expand(ctx, plan, &input))
	if resp.Diagnostics.HasError() {
		return
	}
	input.Tags = getTagsIn(ctx)

	out, err := conn.PutAlarmMuteRule(ctx, &input)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, plan.Name.String())
		return
	}
	if out == nil {
		smerr.AddError(ctx, &resp.Diagnostics, errors.New("empty output"), smerr.ID, plan.Name.String())
		return
	}

	alarmMuteRule, err := findAlarmMuteRuleByName(ctx, conn, plan.Name.ValueString())
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, plan.Name.String())
		return
	}

	// Normalize timestamps to UTC to ensure consistent timezone representation.
	// This is necessary to prevent diffs caused by AWS API response.
	// error sample without normalization: .expire_date: was cty.StringVal("2026-12-31T23:59:00Z"), but now cty.StringVal("2027-01-01T08:59:00+09:00").
	if alarmMuteRule.StartDate != nil {
		utc := alarmMuteRule.StartDate.UTC()
		alarmMuteRule.StartDate = &utc
	}
	if alarmMuteRule.ExpireDate != nil {
		utc := alarmMuteRule.ExpireDate.UTC()
		alarmMuteRule.ExpireDate = &utc
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, flex.Flatten(ctx, alarmMuteRule, &plan))
	if resp.Diagnostics.HasError() {
		return
	}

	plan.ARN = flex.StringToFramework(ctx, alarmMuteRule.AlarmMuteRuleArn)

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, plan))
}

func (r *alarmMuteRuleResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().CloudWatchClient(ctx)

	var state alarmMuteRuleResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := findAlarmMuteRuleByName(ctx, conn, state.Name.ValueString())
	if retry.NotFound(err) {
		resp.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, state.Name.String())
		return
	}

	// Normalize timestamps to UTC to ensure consistent timezone representation.
	// This is necessary to prevent diffs caused by AWS API response.
	// error sample without normalization: .expire_date: was cty.StringVal("2026-12-31T23:59:00Z"), but now cty.StringVal("2027-01-01T08:59:00+09:00").
	if out.StartDate != nil {
		utc := out.StartDate.UTC()
		out.StartDate = &utc
	}
	if out.ExpireDate != nil {
		utc := out.ExpireDate.UTC()
		out.ExpireDate = &utc
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, flex.Flatten(ctx, out, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	state.ARN = flex.StringToFramework(ctx, out.AlarmMuteRuleArn)

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &state))
}

func (r *alarmMuteRuleResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	conn := r.Meta().CloudWatchClient(ctx)

	var plan, state alarmMuteRuleResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Plan.Get(ctx, &plan))
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	diff, d := flex.Diff(ctx, plan, state)
	smerr.AddEnrich(ctx, &resp.Diagnostics, d)
	if resp.Diagnostics.HasError() {
		return
	}

	if diff.HasChanges() {
		var input cloudwatch.PutAlarmMuteRuleInput
		smerr.AddEnrich(ctx, &resp.Diagnostics, flex.Expand(ctx, plan, &input))
		if resp.Diagnostics.HasError() {
			return
		}

		out, err := conn.PutAlarmMuteRule(ctx, &input)
		if err != nil {
			smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, plan.Name.String())
			return
		}
		if out == nil {
			smerr.AddError(ctx, &resp.Diagnostics, errors.New("empty output"), smerr.ID, plan.Name.String())
			return
		}
	}

	// Always read back from AWS, even for tag-only updates
	alarmMuteRule, err := findAlarmMuteRuleByName(ctx, conn, plan.Name.ValueString())
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, plan.Name.String())
		return
	}

	// Normalize timestamps to UTC to ensure consistent timezone representation.
	// This is necessary to prevent diffs caused by AWS API response.
	// error sample without normalization: .expire_date: was cty.StringVal("2026-12-31T23:59:00Z"), but now cty.StringVal("2027-01-01T08:59:00+09:00").
	if alarmMuteRule.StartDate != nil {
		utc := alarmMuteRule.StartDate.UTC()
		alarmMuteRule.StartDate = &utc
	}
	if alarmMuteRule.ExpireDate != nil {
		utc := alarmMuteRule.ExpireDate.UTC()
		alarmMuteRule.ExpireDate = &utc
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, flex.Flatten(ctx, alarmMuteRule, &plan))
	if resp.Diagnostics.HasError() {
		return
	}

	plan.ARN = flex.StringToFramework(ctx, alarmMuteRule.AlarmMuteRuleArn)

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &plan))
}

func (r *alarmMuteRuleResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := r.Meta().CloudWatchClient(ctx)

	var state alarmMuteRuleResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	input := cloudwatch.DeleteAlarmMuteRuleInput{
		AlarmMuteRuleName: state.Name.ValueStringPointer(),
	}

	_, err := conn.DeleteAlarmMuteRule(ctx, &input)
	if err != nil {
		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return
		}

		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, state.Name.String())
		return
	}
}

func (r *alarmMuteRuleResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root(names.AttrName), req, resp)
}

func findAlarmMuteRuleByName(ctx context.Context, conn *cloudwatch.Client, name string) (*cloudwatch.GetAlarmMuteRuleOutput, error) {
	input := cloudwatch.GetAlarmMuteRuleInput{
		AlarmMuteRuleName: aws.String(name),
	}

	out, err := conn.GetAlarmMuteRule(ctx, &input)
	if err != nil {
		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return nil, smarterr.NewError(&retry.NotFoundError{
				LastError: err,
			})
		}

		return nil, smarterr.NewError(err)
	}

	if out == nil {
		return nil, smarterr.NewError(tfresource.NewEmptyResultError())
	}

	return out, nil
}

type alarmMuteRuleResourceModel struct {
	framework.WithRegionModel
	ARN                  types.String                                      `tfsdk:"arn"`
	Description          types.String                                      `tfsdk:"description"`
	ExpireDate           timetypes.RFC3339                                 `tfsdk:"expire_date"`
	LastUpdatedTimestamp timetypes.RFC3339                                 `tfsdk:"last_updated_timestamp"`
	MuteTargets          fwtypes.ListNestedObjectValueOf[muteTargetsModel] `tfsdk:"mute_targets" autoflex:",omitempty"`
	MuteType             types.String                                      `tfsdk:"mute_type"`
	Name                 types.String                                      `tfsdk:"name"`
	Rule                 fwtypes.ListNestedObjectValueOf[ruleModel]        `tfsdk:"rule"`
	StartDate            timetypes.RFC3339                                 `tfsdk:"start_date"`
	Status               fwtypes.StringEnum[awstypes.AlarmMuteRuleStatus]  `tfsdk:"status"`
	Tags                 tftags.Map                                        `tfsdk:"tags"`
	TagsAll              tftags.Map                                        `tfsdk:"tags_all"`
}

type ruleModel struct {
	Schedule fwtypes.ListNestedObjectValueOf[scheduleModel] `tfsdk:"schedule"`
}

type scheduleModel struct {
	Duration   types.String `tfsdk:"duration"`
	Expression types.String `tfsdk:"expression"`
	Timezone   types.String `tfsdk:"timezone"`
}

type muteTargetsModel struct {
	AlarmNames fwtypes.ListValueOf[types.String] `tfsdk:"alarm_names"`
}

// This validator ensures the timestamp has seconds set to 00.
// the CloudWatch API truncates timestamps to minute precision.
func validateTimeMinutePrecision() validator.String {
	return timeMinutePrecisionValidator{}
}

var _ validator.String = timeMinutePrecisionValidator{}

type timeMinutePrecisionValidator struct{}

func (v timeMinutePrecisionValidator) Description(_ context.Context) string {
	return "value must have seconds set to 00 (e.g., 2026-01-01T00:01:00Z) because the CloudWatch API truncates to minute precision"
}

func (v timeMinutePrecisionValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v timeMinutePrecisionValidator) ValidateString(ctx context.Context, request validator.StringRequest, response *validator.StringResponse) {
	if request.ConfigValue.IsNull() || request.ConfigValue.IsUnknown() {
		return
	}

	value := request.ConfigValue.ValueString()

	parsedTime, err := time.Parse(time.RFC3339, value)
	if err != nil {
		response.Diagnostics.Append(validatordiag.InvalidAttributeValueDiagnostic(
			request.Path,
			fmt.Sprintf("Could not parse timestamp: %s", err),
			value,
		))
		return
	}

	if parsedTime.Second() != 0 {
		correctedTime := parsedTime.Truncate(time.Minute).Format(time.RFC3339)
		response.Diagnostics.Append(validatordiag.InvalidAttributeValueDiagnostic(
			request.Path,
			fmt.Sprintf("%s. Suggested value: %s", v.Description(ctx), correctedTime),
			value,
		))
	}
}
