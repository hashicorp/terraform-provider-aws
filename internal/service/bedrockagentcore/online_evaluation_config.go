// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package bedrockagentcore

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/YakDriver/smarterr"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/bedrockagentcorecontrol"
	awstypes "github.com/aws/aws-sdk-go-v2/service/bedrockagentcorecontrol/types"
	"github.com/hashicorp/aws-sdk-go-base/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/float64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
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

// @FrameworkResource("aws_bedrockagentcore_online_evaluation_config", name="Online Evaluation Config")
// @Tags(identifierAttribute="online_evaluation_config_arn")
// @IdentityAttribute("online_evaluation_config_id")
// @Testing(existsType="github.com/aws/aws-sdk-go-v2/service/bedrockagentcorecontrol;bedrockagentcorecontrol.GetOnlineEvaluationConfigOutput")
// @Testing(generator="testAccRandomOnlineEvaluationConfigName(t)")
// @Testing(tagsTest=false)
// @Testing(hasNoPreExistingResource=true)
// @Testing(importStateIdAttribute="online_evaluation_config_id")
// @Testing(importIgnore="enable_on_create")
func newOnlineEvaluationConfigResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &onlineEvaluationConfigResource{}

	r.SetDefaultCreateTimeout(30 * time.Minute)
	r.SetDefaultUpdateTimeout(30 * time.Minute)
	r.SetDefaultDeleteTimeout(30 * time.Minute)

	return r, nil
}

type onlineEvaluationConfigResource struct {
	framework.ResourceWithModel[onlineEvaluationConfigResourceModel]
	framework.WithTimeouts
	framework.WithImportByIdentity
}

func (r *onlineEvaluationConfigResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	const (
		samplingPercentageMin = 0.01
		samplingPercentageMax = 100.0
	)
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrDescription: schema.StringAttribute{
				Optional: true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 200),
				},
			},
			"enable_on_create": schema.BoolAttribute{
				Required: true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"evaluation_execution_role_arn": schema.StringAttribute{
				CustomType: fwtypes.ARNType,
				Required:   true,
			},
			"execution_status": schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.OnlineEvaluationExecutionStatus](),
				Optional:   true,
				Computed:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"online_evaluation_config_arn": framework.ARNAttributeComputedOnly(),
			"online_evaluation_config_id":  framework.IDAttribute(),
			"online_evaluation_config_name": schema.StringAttribute{
				Required: true,
				Validators: []validator.String{
					stringvalidator.RegexMatches(regexache.MustCompile(`^[a-zA-Z][a-zA-Z0-9_]{0,47}$`), "must start with a letter and contain only alphanumeric characters and underscores, up to 48 characters"),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"output_config":   framework.ResourceComputedListOfObjectsAttribute[outputConfigModel](ctx, listplanmodifier.UseStateForUnknown()),
			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
		},
		Blocks: map[string]schema.Block{
			"data_source_config": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[dataSourceConfigModel](ctx),
				Validators: []validator.List{
					listvalidator.SizeBetween(1, 1),
					listvalidator.IsRequired(),
				},
				NestedObject: schema.NestedBlockObject{
					Blocks: map[string]schema.Block{
						names.AttrCloudWatchLogs: schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[cloudWatchLogsInputConfigModel](ctx),
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"log_group_names": schema.ListAttribute{
										CustomType: fwtypes.ListOfStringType,
										Required:   true,
									},
									"service_names": schema.ListAttribute{
										CustomType: fwtypes.ListOfStringType,
										Required:   true,
									},
								},
							},
						},
					},
				},
			},
			"evaluator": schema.SetNestedBlock{
				CustomType: fwtypes.NewSetNestedObjectTypeOf[evaluatorReferenceModel](ctx),
				Validators: []validator.Set{
					setvalidator.SizeBetween(1, 10),
					setvalidator.IsRequired(),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"evaluator_id": schema.StringAttribute{
							Required: true,
							Validators: []validator.String{
								stringvalidator.RegexMatches(regexache.MustCompile(`^(Builtin\.[a-zA-Z0-9_-]+|[a-zA-Z][a-zA-Z0-9-_]{0,99}-[a-zA-Z0-9]{10})$`), "must be a builtin evaluator (e.g. Builtin.Helpfulness) or a custom evaluator ID"),
							},
						},
					},
				},
			},
			names.AttrRule: schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[ruleModel](ctx),
				Validators: []validator.List{
					listvalidator.SizeBetween(1, 1),
					listvalidator.IsRequired(),
				},
				NestedObject: schema.NestedBlockObject{
					Blocks: map[string]schema.Block{
						"sampling_config": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[samplingConfigModel](ctx),
							Validators: []validator.List{
								listvalidator.SizeBetween(1, 1),
								listvalidator.IsRequired(),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"sampling_percentage": schema.Float64Attribute{
										Required: true,
										Validators: []validator.Float64{
											float64validator.Between(samplingPercentageMin, samplingPercentageMax),
										},
									},
								},
							},
						},
						names.AttrFilter: schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[filterModel](ctx),
							Validators: []validator.List{
								listvalidator.SizeAtMost(5),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									names.AttrKey: schema.StringAttribute{
										Required: true,
										Validators: []validator.String{
											stringvalidator.LengthBetween(1, 256),
										},
									},
									"operator": schema.StringAttribute{
										CustomType: fwtypes.StringEnumType[awstypes.FilterOperator](),
										Required:   true,
									},
								},
								Blocks: map[string]schema.Block{
									names.AttrValue: schema.ListNestedBlock{
										CustomType: fwtypes.NewListNestedObjectTypeOf[filterValueModel](ctx),
										Validators: []validator.List{
											listvalidator.SizeBetween(1, 1),
											listvalidator.IsRequired(),
										},
										NestedObject: schema.NestedBlockObject{
											Attributes: map[string]schema.Attribute{
												"boolean_value": schema.BoolAttribute{
													Optional: true,
												},
												"double_value": schema.Float64Attribute{
													Optional: true,
												},
												"string_value": schema.StringAttribute{
													Optional: true,
													Validators: []validator.String{
														stringvalidator.LengthBetween(1, 1024),
													},
												},
											},
										},
									},
								},
							},
						},
						"session_config": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[sessionConfigModel](ctx),
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"session_timeout_minutes": schema.Int64Attribute{
										Required: true,
										Validators: []validator.Int64{
											int64validator.Between(1, 60),
										},
									},
								},
							},
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

func (r *onlineEvaluationConfigResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data onlineEvaluationConfigResourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.Plan.Get(ctx, &data))
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().BedrockAgentCoreClient(ctx)

	var input bedrockagentcorecontrol.CreateOnlineEvaluationConfigInput
	smerr.AddEnrich(ctx, &response.Diagnostics, fwflex.Expand(ctx, data, &input))
	if response.Diagnostics.HasError() {
		return
	}

	// Additional fields.
	input.ClientToken = aws.String(create.UniqueId(ctx))
	input.Tags = getTagsIn(ctx)

	var (
		out *bedrockagentcorecontrol.CreateOnlineEvaluationConfigOutput
		err error
	)
	err = tfresource.Retry(ctx, propagationTimeout, func(ctx context.Context) *tfresource.RetryError {
		out, err = conn.CreateOnlineEvaluationConfig(ctx, &input)

		if tfawserr.ErrMessageContains(err, errCodeValidationException, "The provided execution role cannot be assumed") {
			return tfresource.RetryableError(err)
		}

		if err != nil {
			return tfresource.NonRetryableError(err)
		}

		return nil
	})
	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, data.OnlineEvaluationConfigName.String())
		return
	}

	configID := aws.ToString(out.OnlineEvaluationConfigId)

	if _, err := waitOnlineEvaluationConfigCreated(ctx, conn, configID, r.CreateTimeout(ctx, data.Timeouts)); err != nil {
		// Taint the resource.
		response.State.SetAttribute(ctx, path.Root("online_evaluation_config_id"), configID)
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, configID)
		return
	}

	config, err := findOnlineEvaluationConfigByID(ctx, conn, configID)
	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, configID)
		return
	}

	// Set values for unknowns.
	smerr.AddEnrich(ctx, &response.Diagnostics, fwflex.Flatten(ctx, config, &data))
	if response.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, response.State.Set(ctx, data))
}

func (r *onlineEvaluationConfigResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data onlineEvaluationConfigResourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.State.Get(ctx, &data))
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().BedrockAgentCoreClient(ctx)

	configID := fwflex.StringValueFromFramework(ctx, data.OnlineEvaluationConfigID)
	out, err := findOnlineEvaluationConfigByID(ctx, conn, configID)
	if retry.NotFound(err) {
		smerr.AddOne(ctx, &response.Diagnostics, fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, configID)
		return
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, fwflex.Flatten(ctx, out, &data))
	if response.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, response.State.Set(ctx, &data))
}

func (r *onlineEvaluationConfigResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var new, old onlineEvaluationConfigResourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.Plan.Get(ctx, &new))
	smerr.AddEnrich(ctx, &response.Diagnostics, request.State.Get(ctx, &old))
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().BedrockAgentCoreClient(ctx)

	diff, d := fwflex.Diff(ctx, new, old)
	smerr.AddEnrich(ctx, &response.Diagnostics, d)
	if response.Diagnostics.HasError() {
		return
	}

	if diff.HasChanges() {
		configID := fwflex.StringValueFromFramework(ctx, new.OnlineEvaluationConfigID)

		var input bedrockagentcorecontrol.UpdateOnlineEvaluationConfigInput
		smerr.AddEnrich(ctx, &response.Diagnostics, fwflex.Expand(ctx, new, &input))
		if response.Diagnostics.HasError() {
			return
		}

		// Additional fields.
		input.ClientToken = aws.String(create.UniqueId(ctx))

		_, err := conn.UpdateOnlineEvaluationConfig(ctx, &input)
		if err != nil {
			smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, configID)
			return
		}

		if _, err := waitOnlineEvaluationConfigUpdated(ctx, conn, configID, r.UpdateTimeout(ctx, new.Timeouts)); err != nil {
			smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, configID)
			return
		}
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, response.State.Set(ctx, &new))
}

func (r *onlineEvaluationConfigResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data onlineEvaluationConfigResourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.State.Get(ctx, &data))
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().BedrockAgentCoreClient(ctx)

	configID := fwflex.StringValueFromFramework(ctx, data.OnlineEvaluationConfigID)
	input := bedrockagentcorecontrol.DeleteOnlineEvaluationConfigInput{
		OnlineEvaluationConfigId: aws.String(configID),
	}
	_, err := conn.DeleteOnlineEvaluationConfig(ctx, &input)
	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return
	}
	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, configID)
		return
	}

	if _, err := waitOnlineEvaluationConfigDeleted(ctx, conn, configID, r.DeleteTimeout(ctx, data.Timeouts)); err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, configID)
		return
	}
}

// Waiters.

func waitOnlineEvaluationConfigCreated(ctx context.Context, conn *bedrockagentcorecontrol.Client, id string, timeout time.Duration) (*bedrockagentcorecontrol.GetOnlineEvaluationConfigOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   enum.Slice(awstypes.OnlineEvaluationConfigStatusCreating),
		Target:                    enum.Slice(awstypes.OnlineEvaluationConfigStatusActive),
		Refresh:                   statusOnlineEvaluationConfig(conn, id),
		Timeout:                   timeout,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*bedrockagentcorecontrol.GetOnlineEvaluationConfigOutput); ok {
		retry.SetLastError(err, errors.New(aws.ToString(out.FailureReason)))
		return out, smarterr.NewError(err)
	}

	return nil, smarterr.NewError(err)
}

func waitOnlineEvaluationConfigUpdated(ctx context.Context, conn *bedrockagentcorecontrol.Client, id string, timeout time.Duration) (*bedrockagentcorecontrol.GetOnlineEvaluationConfigOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   enum.Slice(awstypes.OnlineEvaluationConfigStatusUpdating),
		Target:                    enum.Slice(awstypes.OnlineEvaluationConfigStatusActive),
		Refresh:                   statusOnlineEvaluationConfig(conn, id),
		Timeout:                   timeout,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*bedrockagentcorecontrol.GetOnlineEvaluationConfigOutput); ok {
		retry.SetLastError(err, errors.New(aws.ToString(out.FailureReason)))
		return out, smarterr.NewError(err)
	}

	return nil, smarterr.NewError(err)
}

func waitOnlineEvaluationConfigDeleted(ctx context.Context, conn *bedrockagentcorecontrol.Client, id string, timeout time.Duration) (*bedrockagentcorecontrol.GetOnlineEvaluationConfigOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.OnlineEvaluationConfigStatusDeleting, awstypes.OnlineEvaluationConfigStatusActive),
		Target:  []string{},
		Refresh: statusOnlineEvaluationConfig(conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*bedrockagentcorecontrol.GetOnlineEvaluationConfigOutput); ok {
		retry.SetLastError(err, errors.New(aws.ToString(out.FailureReason)))
		return out, smarterr.NewError(err)
	}

	return nil, smarterr.NewError(err)
}

func statusOnlineEvaluationConfig(conn *bedrockagentcorecontrol.Client, id string) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		out, err := findOnlineEvaluationConfigByID(ctx, conn, id)
		if retry.NotFound(err) {
			return nil, "", nil
		}
		if err != nil {
			return nil, "", smarterr.NewError(err)
		}
		return out, string(out.Status), nil
	}
}

// Finders.

func findOnlineEvaluationConfigByID(ctx context.Context, conn *bedrockagentcorecontrol.Client, id string) (*bedrockagentcorecontrol.GetOnlineEvaluationConfigOutput, error) {
	input := bedrockagentcorecontrol.GetOnlineEvaluationConfigInput{
		OnlineEvaluationConfigId: aws.String(id),
	}
	return findOnlineEvaluationConfig(ctx, conn, &input)
}

func findOnlineEvaluationConfig(ctx context.Context, conn *bedrockagentcorecontrol.Client, input *bedrockagentcorecontrol.GetOnlineEvaluationConfigInput) (*bedrockagentcorecontrol.GetOnlineEvaluationConfigOutput, error) {
	out, err := conn.GetOnlineEvaluationConfig(ctx, input)
	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, smarterr.NewError(&retry.NotFoundError{LastError: err})
	}
	if err != nil {
		return nil, smarterr.NewError(err)
	}
	if out == nil {
		return nil, smarterr.NewError(tfresource.NewEmptyResultError())
	}
	return out, nil
}

// Models.

type onlineEvaluationConfigResourceModel struct {
	framework.WithRegionModel
	DataSourceConfig           fwtypes.ListNestedObjectValueOf[dataSourceConfigModel]       `tfsdk:"data_source_config"`
	Description                types.String                                                 `tfsdk:"description"`
	EnableOnCreate             types.Bool                                                   `tfsdk:"enable_on_create"`
	EvaluationExecutionRoleARN fwtypes.ARN                                                  `tfsdk:"evaluation_execution_role_arn"`
	Evaluators                 fwtypes.SetNestedObjectValueOf[evaluatorReferenceModel]      `tfsdk:"evaluator"`
	ExecutionStatus            fwtypes.StringEnum[awstypes.OnlineEvaluationExecutionStatus] `tfsdk:"execution_status"`
	OnlineEvaluationConfigARN  types.String                                                 `tfsdk:"online_evaluation_config_arn"`
	OnlineEvaluationConfigID   types.String                                                 `tfsdk:"online_evaluation_config_id"`
	OnlineEvaluationConfigName types.String                                                 `tfsdk:"online_evaluation_config_name"`
	OutputConfig               fwtypes.ListNestedObjectValueOf[outputConfigModel]           `tfsdk:"output_config"`
	Rule                       fwtypes.ListNestedObjectValueOf[ruleModel]                   `tfsdk:"rule"`
	Tags                       tftags.Map                                                   `tfsdk:"tags"`
	TagsAll                    tftags.Map                                                   `tfsdk:"tags_all"`
	Timeouts                   timeouts.Value                                               `tfsdk:"timeouts"`
}

type dataSourceConfigModel struct {
	CloudWatchLogs fwtypes.ListNestedObjectValueOf[cloudWatchLogsInputConfigModel] `tfsdk:"cloudwatch_logs"`
}

var (
	_ fwflex.Expander  = dataSourceConfigModel{}
	_ fwflex.Flattener = &dataSourceConfigModel{}
)

func (m dataSourceConfigModel) Expand(ctx context.Context) (any, diag.Diagnostics) {
	var diags diag.Diagnostics
	switch {
	case !m.CloudWatchLogs.IsNull():
		data, d := m.CloudWatchLogs.ToPtr(ctx)
		diags.Append(d...)
		if diags.HasError() {
			return nil, diags
		}
		var r awstypes.DataSourceConfigMemberCloudWatchLogs
		diags.Append(fwflex.Expand(ctx, data, &r.Value)...)
		if diags.HasError() {
			return nil, diags
		}
		return &r, diags
	}
	return nil, diags
}

func (m *dataSourceConfigModel) Flatten(ctx context.Context, v any) diag.Diagnostics {
	var diags diag.Diagnostics
	switch t := v.(type) {
	case awstypes.DataSourceConfigMemberCloudWatchLogs:
		var data cloudWatchLogsInputConfigModel
		diags.Append(fwflex.Flatten(ctx, t.Value, &data)...)
		if diags.HasError() {
			return diags
		}
		m.CloudWatchLogs = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &data)
	default:
		diags.AddError("Unsupported Type", fmt.Sprintf("data_source_config flatten: %T", v))
	}
	return diags
}

type cloudWatchLogsInputConfigModel struct {
	LogGroupNames fwtypes.ListOfString `tfsdk:"log_group_names"`
	ServiceNames  fwtypes.ListOfString `tfsdk:"service_names"`
}

type evaluatorReferenceModel struct {
	EvaluatorID types.String `tfsdk:"evaluator_id"`
}

var (
	_ fwflex.Expander  = evaluatorReferenceModel{}
	_ fwflex.Flattener = &evaluatorReferenceModel{}
)

func (m evaluatorReferenceModel) Expand(ctx context.Context) (any, diag.Diagnostics) {
	var diags diag.Diagnostics
	switch {
	case !m.EvaluatorID.IsNull():
		return &awstypes.EvaluatorReferenceMemberEvaluatorId{
			Value: m.EvaluatorID.ValueString(),
		}, diags
	}
	return nil, diags
}

func (m *evaluatorReferenceModel) Flatten(ctx context.Context, v any) diag.Diagnostics {
	var diags diag.Diagnostics
	switch t := v.(type) {
	case awstypes.EvaluatorReferenceMemberEvaluatorId:
		m.EvaluatorID = types.StringValue(t.Value)
	default:
		diags.AddError("Unsupported Type", fmt.Sprintf("evaluator flatten: %T", v))
	}
	return diags
}

type outputConfigModel struct {
	CloudWatchConfig fwtypes.ListNestedObjectValueOf[cloudWatchOutputConfigModel] `tfsdk:"cloudwatch_config"`
}

type cloudWatchOutputConfigModel struct {
	LogGroupName types.String `tfsdk:"log_group_name"`
}

type ruleModel struct {
	Filters        fwtypes.ListNestedObjectValueOf[filterModel]         `tfsdk:"filter"`
	SamplingConfig fwtypes.ListNestedObjectValueOf[samplingConfigModel] `tfsdk:"sampling_config"`
	SessionConfig  fwtypes.ListNestedObjectValueOf[sessionConfigModel]  `tfsdk:"session_config"`
}

type samplingConfigModel struct {
	SamplingPercentage types.Float64 `tfsdk:"sampling_percentage"`
}

type filterModel struct {
	Key      types.String                                      `tfsdk:"key"`
	Operator fwtypes.StringEnum[awstypes.FilterOperator]       `tfsdk:"operator"`
	Value    fwtypes.ListNestedObjectValueOf[filterValueModel] `tfsdk:"value"`
}

type filterValueModel struct {
	BooleanValue types.Bool    `tfsdk:"boolean_value"`
	DoubleValue  types.Float64 `tfsdk:"double_value"`
	StringValue  types.String  `tfsdk:"string_value"`
}

var (
	_ fwflex.Expander  = filterValueModel{}
	_ fwflex.Flattener = &filterValueModel{}
)

func (m filterValueModel) Expand(ctx context.Context) (any, diag.Diagnostics) {
	var diags diag.Diagnostics
	switch {
	case !m.StringValue.IsNull() && !m.StringValue.IsUnknown():
		return &awstypes.FilterValueMemberStringValue{
			Value: m.StringValue.ValueString(),
		}, diags
	case !m.BooleanValue.IsNull() && !m.BooleanValue.IsUnknown():
		return &awstypes.FilterValueMemberBooleanValue{
			Value: m.BooleanValue.ValueBool(),
		}, diags
	case !m.DoubleValue.IsNull() && !m.DoubleValue.IsUnknown():
		return &awstypes.FilterValueMemberDoubleValue{
			Value: m.DoubleValue.ValueFloat64(),
		}, diags
	}
	return nil, diags
}

func (m *filterValueModel) Flatten(ctx context.Context, v any) diag.Diagnostics {
	var diags diag.Diagnostics
	switch t := v.(type) {
	case awstypes.FilterValueMemberStringValue:
		m.StringValue = types.StringValue(t.Value)
	case awstypes.FilterValueMemberBooleanValue:
		m.BooleanValue = types.BoolValue(t.Value)
	case awstypes.FilterValueMemberDoubleValue:
		m.DoubleValue = types.Float64Value(t.Value)
	default:
		diags.AddError("Unsupported Type", fmt.Sprintf("filter value flatten: %T", v))
	}
	return diags
}

type sessionConfigModel struct {
	SessionTimeoutMinutes types.Int64 `tfsdk:"session_timeout_minutes"`
}
