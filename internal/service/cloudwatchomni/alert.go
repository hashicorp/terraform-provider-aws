// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package cloudwatchomni

import (
	"context"
	"fmt"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/YakDriver/smarterr"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatchomni"
	awstypes "github.com/aws/aws-sdk-go-v2/service/cloudwatchomni/types"
	"github.com/hashicorp/terraform-plugin-framework-validators/int32validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/mapvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int32planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
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

// @FrameworkResource("aws_cloudwatchomni_alert", name="Alert")
// @Tags(identifierAttribute="arn")
// @IdentityAttribute("space_id")
// @IdentityAttribute("alert_id")
// @ImportIDHandler("alertImportID")
// @Testing(existsType="github.com/aws/aws-sdk-go-v2/service/cloudwatchomni/types;awstypes;awstypes.Alert")
// @Testing(preCheck="testAccPreCheck")
// @Testing(hasNoPreExistingResource=true)
// @Testing(serialize=true)
// @Testing(importStateIdAttributes="space_id;alert_id", importStateIdAttributesSep="flex.ResourceIdSeparator")
func newAlertResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	return &alertResource{}, nil
}

type alertResource struct {
	framework.ResourceWithModel[alertResourceModel]
	framework.WithImportByIdentity
}

func (r *alertResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"alert_id":    framework.IDAttribute(),
			names.AttrARN: framework.ARNAttributeComputedOnly(),
			names.AttrDescription: schema.StringAttribute{
				Optional: true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
				PlanModifiers: []planmodifier.String{
					// UpdateAlert leaves an omitted description unchanged, so removing the description requires replacement.
					stringplanmodifier.RequiresReplaceIf(func(ctx context.Context, request planmodifier.StringRequest, response *stringplanmodifier.RequiresReplaceIfFuncResponse) {
						response.RequiresReplace = request.PlanValue.IsNull() && !request.StateValue.IsNull()
					}, "Removing the description requires replacement.", "Removing the description requires replacement."),
				},
			},
			names.AttrName: schema.StringAttribute{
				Required: true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 256),
				},
			},
			"notifications_enabled": schema.BoolAttribute{
				Optional: true,
				Computed: true,
				Default:  booldefault.StaticBool(true),
			},
			"profile_id": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					// UpdateAlert's access check fails for existing alerts whenever profileId is sent, so changing it requires replacement.
					stringplanmodifier.RequiresReplace(),
				},
			},
			"space_id": schema.StringAttribute{
				Required: true,
				Validators: []validator.String{
					stringvalidator.RegexMatches(regexache.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`), "must be a space ID"),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
		},
		Blocks: map[string]schema.Block{
			"notification_rule": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[notificationRuleModel](ctx),
				NestedObject: schema.NestedBlockObject{
					Blocks: map[string]schema.Block{
						names.AttrTarget: schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[notificationTargetModel](ctx),
							Validators: []validator.List{
								listvalidator.IsRequired(),
								listvalidator.SizeAtMost(1),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									names.AttrARN: schema.StringAttribute{
										CustomType: fwtypes.ARNType,
										Required:   true,
									},
									"metadata": schema.MapAttribute{
										CustomType: fwtypes.MapOfStringType,
										Optional:   true,
										Validators: []validator.Map{
											mapvalidator.SizeAtLeast(1),
										},
									},
									names.AttrType: schema.StringAttribute{
										CustomType: fwtypes.StringEnumType[awstypes.NotificationTargetType](),
										Required:   true,
									},
								},
							},
						},
						"trigger": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[notificationTriggerModel](ctx),
							Validators: []validator.List{
								listvalidator.IsRequired(),
								listvalidator.SizeAtMost(1),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"state_values": schema.SetAttribute{
										CustomType: fwtypes.SetOfStringEnumType[awstypes.AlertState](),
										Optional:   true,
										Validators: []validator.Set{
											setvalidator.SizeAtLeast(1),
										},
									},
								},
							},
						},
					},
				},
			},
			names.AttrRule: schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[ruleModel](ctx),
				Validators: []validator.List{
					listvalidator.IsRequired(),
					listvalidator.SizeAtMost(1),
				},
				NestedObject: schema.NestedBlockObject{
					Blocks: map[string]schema.Block{
						"telemetry_rule": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[telemetryRuleModel](ctx),
							Validators: []validator.List{
								listvalidator.IsRequired(),
								listvalidator.SizeAtMost(1),
							},
							NestedObject: schema.NestedBlockObject{
								Blocks: map[string]schema.Block{
									names.AttrCondition: schema.ListNestedBlock{
										CustomType: fwtypes.NewListNestedObjectTypeOf[alertConditionModel](ctx),
										Validators: []validator.List{
											listvalidator.IsRequired(),
											listvalidator.SizeAtMost(1),
										},
										NestedObject: schema.NestedBlockObject{
											Attributes: map[string]schema.Attribute{
												"comparator": schema.StringAttribute{
													CustomType: fwtypes.StringEnumType[awstypes.Comparator](),
													Optional:   true,
													Computed:   true,
													PlanModifiers: []planmodifier.String{
														stringplanmodifier.UseStateForUnknown(),
													},
												},
												"critical_threshold": schema.Float64Attribute{
													Optional: true,
												},
												"threshold_field": schema.StringAttribute{
													Optional: true,
												},
												"threshold_mode": schema.StringAttribute{
													CustomType: fwtypes.StringEnumType[awstypes.ThresholdMode](),
													Optional:   true,
													Computed:   true,
													PlanModifiers: []planmodifier.String{
														stringplanmodifier.UseStateForUnknown(),
													},
												},
												"warning_threshold": schema.Float64Attribute{
													Optional: true,
												},
											},
										},
									},
									"evaluation": schema.ListNestedBlock{
										CustomType: fwtypes.NewListNestedObjectTypeOf[alertEvaluationModel](ctx),
										Validators: []validator.List{
											listvalidator.IsRequired(),
											listvalidator.SizeAtMost(1),
										},
										NestedObject: schema.NestedBlockObject{
											Attributes: map[string]schema.Attribute{
												"interval_seconds": schema.Int32Attribute{
													Required: true,
													Validators: []validator.Int32{
														int32validator.AtLeast(1),
													},
												},
												"pending_duration_seconds": schema.Int32Attribute{
													Optional: true,
													Computed: true,
													PlanModifiers: []planmodifier.Int32{
														int32planmodifier.UseStateForUnknown(),
													},
													Validators: []validator.Int32{
														int32validator.AtLeast(0),
													},
												},
												"recovery_duration_seconds": schema.Int32Attribute{
													Optional: true,
													Computed: true,
													PlanModifiers: []planmodifier.Int32{
														int32planmodifier.UseStateForUnknown(),
													},
													Validators: []validator.Int32{
														int32validator.AtLeast(0),
													},
												},
											},
										},
									},
									"no_data": schema.ListNestedBlock{
										CustomType: fwtypes.NewListNestedObjectTypeOf[noDataModel](ctx),
										Validators: []validator.List{
											listvalidator.IsRequired(),
											listvalidator.SizeAtMost(1),
										},
										NestedObject: schema.NestedBlockObject{
											Attributes: map[string]schema.Attribute{
												"treat_as": schema.StringAttribute{
													CustomType: fwtypes.StringEnumType[awstypes.AlertState](),
													Required:   true,
												},
											},
										},
									},
									"query": schema.ListNestedBlock{
										CustomType: fwtypes.NewListNestedObjectTypeOf[alertRuleQueryModel](ctx),
										Validators: []validator.List{
											listvalidator.IsRequired(),
											listvalidator.SizeAtMost(1),
										},
										NestedObject: schema.NestedBlockObject{
											Attributes: map[string]schema.Attribute{
												names.AttrExpression: schema.StringAttribute{
													Required: true,
													Validators: []validator.String{
														stringvalidator.LengthAtLeast(1),
													},
												},
												"language": schema.StringAttribute{
													CustomType: fwtypes.StringEnumType[awstypes.QueryLanguage](),
													Required:   true,
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

func (r *alertResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan alertResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Plan.Get(ctx, &plan))
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().CloudWatchOmniClient(ctx)

	name := fwflex.StringValueFromFramework(ctx, plan.Name)
	var input cloudwatchomni.CreateAlertInput
	smerr.AddEnrich(ctx, &resp.Diagnostics, fwflex.Expand(ctx, plan, &input))
	if resp.Diagnostics.HasError() {
		return
	}

	// Additional fields.
	input.ClientToken = aws.String(create.UniqueId(ctx))

	// CreateAlert takes a tag map, but the tag operations live in the cloudwatch package and use a slice.
	if tags := keyValueTags(ctx, getTagsIn(ctx)).Map(); len(tags) > 0 {
		input.Tags = tags
	}

	// A new access grant can take a few minutes to let alerts assume its access profile.
	const (
		alertPropagationTimeout         = 5 * time.Minute
		errMsgCannotAssumeAccessProfile = "cannot assume access profile"
	)
	out, err := tfresource.RetryWhenIsAErrorMessageContains[*cloudwatchomni.CreateAlertOutput, *awstypes.ValidationException](ctx, alertPropagationTimeout,
		func(ctx context.Context) (*cloudwatchomni.CreateAlertOutput, error) {
			return conn.CreateAlert(ctx, &input)
		},
		errMsgCannotAssumeAccessProfile,
	)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.Name, name)
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, r.flatten(ctx, out.Alert, &plan))
	if resp.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, plan))
}

func (r *alertResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state alertResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().CloudWatchOmniClient(ctx)

	spaceID, alertID := fwflex.StringValueFromFramework(ctx, state.SpaceID), fwflex.StringValueFromFramework(ctx, state.AlertID)
	out, err := findAlertByTwoPartKey(ctx, conn, spaceID, alertID)
	if retry.NotFound(err) {
		smerr.AddOne(ctx, &resp.Diagnostics, fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, alertID)
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, r.flatten(ctx, out, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &state))
}

func (r *alertResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state alertResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Plan.Get(ctx, &plan))
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().CloudWatchOmniClient(ctx)

	diff, d := fwflex.Diff(ctx, plan, state)
	smerr.AddEnrich(ctx, &resp.Diagnostics, d)
	if resp.Diagnostics.HasError() {
		return
	}

	if diff.HasChanges() {
		alertID := fwflex.StringValueFromFramework(ctx, plan.AlertID)
		var input cloudwatchomni.UpdateAlertInput
		smerr.AddEnrich(ctx, &resp.Diagnostics, fwflex.Expand(ctx, plan, &input))
		if resp.Diagnostics.HasError() {
			return
		}

		// UpdateAlert re-checks that the alert can assume its access profile whenever profileId is present,
		// and that check fails for existing alerts. profile_id forces replacement, so never send it.
		input.ProfileId = nil

		// UpdateAlert leaves omitted notification rules unchanged; an empty list clears them.
		if input.NotificationRules == nil {
			input.NotificationRules = []awstypes.NotificationRule{}
		}

		if _, err := conn.UpdateAlert(ctx, &input); err != nil {
			smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, alertID)
			return
		}
	}

	// UpdateAlert returns no output; the plan has no unknown values to resolve.
	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &plan))
}

func (r *alertResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state alertResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().CloudWatchOmniClient(ctx)

	spaceID, alertID := fwflex.StringValueFromFramework(ctx, state.SpaceID), fwflex.StringValueFromFramework(ctx, state.AlertID)
	input := cloudwatchomni.DeleteAlertInput{
		AlertId: aws.String(alertID),
		SpaceId: aws.String(spaceID),
	}
	_, err := conn.DeleteAlert(ctx, &input)
	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return
	}
	if err != nil {
		// A missing space is reported as AccessDeniedException.
		if errs.IsA[*awstypes.AccessDeniedException](err) {
			_, findErr := findAlertByTwoPartKey(ctx, conn, spaceID, alertID)
			if retry.NotFound(findErr) {
				return
			}
			if findErr != nil {
				smerr.AddError(ctx, &resp.Diagnostics, findErr, smerr.ID, alertID)
				return
			}
		}

		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, alertID)
		return
	}
}

func (r *alertResource) flatten(ctx context.Context, alert *awstypes.Alert, data *alertResourceModel) diag.Diagnostics {
	var diags diag.Diagnostics

	// The API returns an empty metadata map for targets configured without metadata; treat it as absent.
	for i, v := range alert.NotificationRules {
		if v.Target != nil && len(v.Target.Metadata) == 0 {
			alert.NotificationRules[i].Target.Metadata = nil
		}
	}

	smerr.AddEnrich(ctx, &diags, fwflex.Flatten(ctx, alert, data))
	if diags.HasError() {
		return diags
	}

	// The API takes notificationsEnabled but returns notificationStatus.
	data.NotificationsEnabled = types.BoolValue(alert.NotificationStatus != awstypes.NotificationStatusDisabled)

	return diags
}

var (
	_ inttypes.ImportIDParser = alertImportID{}
)

type alertImportID struct{}

func (alertImportID) Parse(id string) (string, map[string]any, error) {
	const (
		importIDPartCount = 2
	)
	parts, err := intflex.ExpandResourceId(id, importIDPartCount, false)
	if err != nil {
		return "", nil, fmt.Errorf("unexpected format for ID (%[1]s), expected space_id%[2]salert_id: %w", id, intflex.ResourceIdSeparator, err)
	}

	result := map[string]any{
		"alert_id": parts[1],
		"space_id": parts[0],
	}

	return id, result, nil
}

// findAlertByTwoPartKey returns the alert, or a NotFoundError if the alert or its space doesn't exist.
func findAlertByTwoPartKey(ctx context.Context, conn *cloudwatchomni.Client, spaceID, alertID string) (*awstypes.Alert, error) {
	input := cloudwatchomni.GetAlertInput{
		AlertId: aws.String(alertID),
		SpaceId: aws.String(spaceID),
	}

	out, err := conn.GetAlert(ctx, &input)
	switch {
	case errs.IsA[*awstypes.ResourceNotFoundException](err):
		return nil, smarterr.NewError(&retry.NotFoundError{
			LastError: err,
		})
	case errs.IsA[*awstypes.AccessDeniedException](err):
		// The service returns AccessDeniedException instead of ResourceNotFoundException for a missing space,
		// so only treat the error as not found once the space's absence is confirmed.
		_, spaceErr := findSpaceByID(ctx, conn, spaceID)
		if retry.NotFound(spaceErr) {
			return nil, smarterr.NewError(&retry.NotFoundError{
				LastError: err,
			})
		}
		if spaceErr != nil {
			return nil, smarterr.NewError(spaceErr)
		}

		return nil, smarterr.NewError(err)
	case err != nil:
		return nil, smarterr.NewError(err)
	}

	if out == nil || out.Alert == nil {
		return nil, smarterr.NewError(tfresource.NewEmptyResultError())
	}

	return out.Alert, nil
}

type alertResourceModel struct {
	framework.WithRegionModel
	AlertARN             types.String                                           `tfsdk:"arn"`
	AlertID              types.String                                           `tfsdk:"alert_id"`
	Description          types.String                                           `tfsdk:"description"`
	Name                 types.String                                           `tfsdk:"name"`
	NotificationRules    fwtypes.ListNestedObjectValueOf[notificationRuleModel] `tfsdk:"notification_rule"`
	NotificationsEnabled types.Bool                                             `tfsdk:"notifications_enabled"`
	ProfileID            types.String                                           `tfsdk:"profile_id"`
	Rule                 fwtypes.ListNestedObjectValueOf[ruleModel]             `tfsdk:"rule"`
	SpaceID              types.String                                           `tfsdk:"space_id"`
	Tags                 tftags.Map                                             `tfsdk:"tags"`
	TagsAll              tftags.Map                                             `tfsdk:"tags_all"`
}

type notificationRuleModel struct {
	Target  fwtypes.ListNestedObjectValueOf[notificationTargetModel]  `tfsdk:"target"`
	Trigger fwtypes.ListNestedObjectValueOf[notificationTriggerModel] `tfsdk:"trigger"`
}

type notificationTargetModel struct {
	ARN      fwtypes.ARN                                         `tfsdk:"arn"`
	Metadata fwtypes.MapOfString                                 `tfsdk:"metadata"`
	Type     fwtypes.StringEnum[awstypes.NotificationTargetType] `tfsdk:"type"`
}

type notificationTriggerModel struct {
	StateValues fwtypes.SetOfStringEnum[awstypes.AlertState] `tfsdk:"state_values"`
}

type ruleModel struct {
	TelemetryRule fwtypes.ListNestedObjectValueOf[telemetryRuleModel] `tfsdk:"telemetry_rule"`
}

var (
	_ fwflex.Expander  = ruleModel{}
	_ fwflex.Flattener = &ruleModel{}
)

func (m *ruleModel) Flatten(ctx context.Context, v any) diag.Diagnostics {
	var diags diag.Diagnostics

	switch t := v.(type) {
	case awstypes.RuleMemberTelemetryRule:
		var model telemetryRuleModel
		smerr.AddEnrich(ctx, &diags, fwflex.Flatten(ctx, t.Value, &model))
		if diags.HasError() {
			return diags
		}

		var d diag.Diagnostics
		m.TelemetryRule, d = fwtypes.NewListNestedObjectValueOfPtr(ctx, &model)
		smerr.AddEnrich(ctx, &diags, d)
	default:
		diags.AddError(
			"Unsupported Type",
			fmt.Sprintf("ruleModel.Flatten: %T", v),
		)
	}

	return diags
}

func (m ruleModel) Expand(ctx context.Context) (any, diag.Diagnostics) {
	var diags diag.Diagnostics

	switch {
	case !m.TelemetryRule.IsNull():
		model, d := m.TelemetryRule.ToPtr(ctx)
		smerr.AddEnrich(ctx, &diags, d)
		if diags.HasError() {
			return nil, diags
		}

		var r awstypes.RuleMemberTelemetryRule
		smerr.AddEnrich(ctx, &diags, fwflex.Expand(ctx, model, &r.Value))
		if diags.HasError() {
			return nil, diags
		}

		return &r, diags
	}

	return nil, diags
}

type telemetryRuleModel struct {
	Condition  fwtypes.ListNestedObjectValueOf[alertConditionModel]  `tfsdk:"condition"`
	Evaluation fwtypes.ListNestedObjectValueOf[alertEvaluationModel] `tfsdk:"evaluation"`
	NoData     fwtypes.ListNestedObjectValueOf[noDataModel]          `tfsdk:"no_data"`
	Query      fwtypes.ListNestedObjectValueOf[alertRuleQueryModel]  `tfsdk:"query"`
}

type alertConditionModel struct {
	Comparator        fwtypes.StringEnum[awstypes.Comparator]    `tfsdk:"comparator"`
	CriticalThreshold types.Float64                              `tfsdk:"critical_threshold"`
	ThresholdField    types.String                               `tfsdk:"threshold_field"`
	ThresholdMode     fwtypes.StringEnum[awstypes.ThresholdMode] `tfsdk:"threshold_mode"`
	WarningThreshold  types.Float64                              `tfsdk:"warning_threshold"`
}

type alertEvaluationModel struct {
	IntervalSeconds         types.Int32 `tfsdk:"interval_seconds"`
	PendingDurationSeconds  types.Int32 `tfsdk:"pending_duration_seconds"`
	RecoveryDurationSeconds types.Int32 `tfsdk:"recovery_duration_seconds"`
}

type noDataModel struct {
	TreatAs fwtypes.StringEnum[awstypes.AlertState] `tfsdk:"treat_as"`
}

type alertRuleQueryModel struct {
	Expression types.String                               `tfsdk:"expression"`
	Language   fwtypes.StringEnum[awstypes.QueryLanguage] `tfsdk:"language"`
}
