// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package lambdaweb

import (
	"context"
	"fmt"
	"iter"
	"time"

	"github.com/YakDriver/smarterr"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/lambdaweb"
	awstypes "github.com/aws/aws-sdk-go-v2/service/lambdaweb/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/mapplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	intflex "github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	fwvalidators "github.com/hashicorp/terraform-provider-aws/internal/framework/validators"
	tfiter "github.com/hashicorp/terraform-provider-aws/internal/iter"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_lambdaweb_function_revision", name="Function Revision")
// @IdentityAttribute("function_name")
// @IdentityAttribute("revision_id")
// @ImportIDHandler("functionRevisionImportID")
// @Testing(importStateIdAttribute="revision_id")
// @Testing(hasNoPreExistingResource=true)
// @Testing(preCheck="testAccPreCheck")
// @Testing(existsType="github.com/aws/aws-sdk-go-v2/service/lambdaweb;lambdaweb.GetWebFunctionRevisionOutput")
// @Testing(importStateIdFunc="testAccFunctionRevisionImportStateIDFunc")
func newFunctionRevisionResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &functionRevisionResource{}

	r.SetDefaultCreateTimeout(15 * time.Minute)

	return r, nil
}

type functionRevisionResource struct {
	framework.ResourceWithModel[functionRevisionResourceModel]
	framework.WithNoUpdate
	framework.WithTimeouts
	framework.WithImportByIdentity
}

func (r *functionRevisionResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrARN: framework.ARNAttributeComputedOnly(),
			names.AttrDescription: schema.StringAttribute{
				Optional: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 256),
				},
			},
			"function_name": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					validFunctionName,
				},
			},
			names.AttrKMSKeyARN: schema.StringAttribute{
				CustomType: fwtypes.ARNType,
				Optional:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"revision_id": framework.IDAttribute(),
		},
		Blocks: map[string]schema.Block{
			"build_config": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[buildConfigModel](ctx),
				PlanModifiers: []planmodifier.List{
					listplanmodifier.RequiresReplace(),
				},
				Validators: []validator.List{
					listvalidator.IsRequired(),
					listvalidator.SizeAtLeast(1),
					listvalidator.SizeAtMost(1),
				},
				NestedObject: schema.NestedBlockObject{
					Blocks: map[string]schema.Block{
						"code_config": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[codeConfigModel](ctx),
							PlanModifiers: []planmodifier.List{
								listplanmodifier.RequiresReplace(),
							},
							Validators: []validator.List{
								listvalidator.IsRequired(),
								listvalidator.SizeAtLeast(1),
								listvalidator.SizeAtMost(1),
							},
							NestedObject: schema.NestedBlockObject{
								Blocks: map[string]schema.Block{
									"s3_object": schema.ListNestedBlock{
										CustomType: fwtypes.NewListNestedObjectTypeOf[s3ObjectModel](ctx),
										PlanModifiers: []planmodifier.List{
											listplanmodifier.RequiresReplace(),
										},
										Validators: []validator.List{
											listvalidator.IsRequired(),
											listvalidator.SizeAtLeast(1),
											listvalidator.SizeAtMost(1),
										},
										NestedObject: schema.NestedBlockObject{
											Attributes: map[string]schema.Attribute{
												names.AttrBucket: schema.StringAttribute{
													Required: true,
													PlanModifiers: []planmodifier.String{
														stringplanmodifier.RequiresReplace(),
													},
													Validators: []validator.String{
														fwvalidators.S3BucketName,
													},
												},
												names.AttrKey: schema.StringAttribute{
													Required: true,
													PlanModifiers: []planmodifier.String{
														stringplanmodifier.RequiresReplace(),
													},
												},
												"version_id": schema.StringAttribute{
													Optional: true,
													PlanModifiers: []planmodifier.String{
														stringplanmodifier.RequiresReplace(),
													},
												},
											},
										},
									},
								},
							},
						},
						"runtime_config": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[runtimeConfigModel](ctx),
							PlanModifiers: []planmodifier.List{
								listplanmodifier.RequiresReplace(),
							},
							Validators: []validator.List{
								listvalidator.IsRequired(),
								listvalidator.SizeAtLeast(1),
								listvalidator.SizeAtMost(1),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"runtime": schema.StringAttribute{
										Required: true,
										PlanModifiers: []planmodifier.String{
											stringplanmodifier.RequiresReplace(),
										},
									},
								},
							},
						},
					},
				},
			},
			"service_config": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[serviceConfigModel](ctx),
				PlanModifiers: []planmodifier.List{
					listplanmodifier.RequiresReplace(),
				},
				Validators: []validator.List{
					listvalidator.IsRequired(),
					listvalidator.SizeAtLeast(1),
					listvalidator.SizeAtMost(1),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"egress_network_connector_arn": schema.StringAttribute{
							CustomType: fwtypes.ARNType,
							Optional:   true,
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.RequiresReplace(),
							},
						},
						"environment_variables": schema.MapAttribute{
							CustomType: fwtypes.MapOfStringType,
							Optional:   true,
							Sensitive:  true,
							PlanModifiers: []planmodifier.Map{
								mapplanmodifier.RequiresReplace(),
							},
						},
						names.AttrExecutionRoleARN: schema.StringAttribute{
							CustomType: fwtypes.ARNType,
							Required:   true,
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.RequiresReplace(),
							},
						},
						"max_concurrency_per_environment": schema.Int64Attribute{
							Optional: true,
							Computed: true,
							PlanModifiers: []planmodifier.Int64{
								int64planmodifier.RequiresReplaceIfConfigured(),
								int64planmodifier.UseNonNullStateForUnknown(),
							},
							Validators: []validator.Int64{
								int64validator.Between(1, 128),
							},
						},
						"timeout_seconds": schema.Int64Attribute{
							Optional: true,
							Computed: true,
							PlanModifiers: []planmodifier.Int64{
								int64planmodifier.RequiresReplaceIfConfigured(),
								int64planmodifier.UseNonNullStateForUnknown(),
							},
							Validators: []validator.Int64{
								int64validator.Between(3, 900),
							},
						},
					},
					Blocks: map[string]schema.Block{
						"telemetry_config": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[telemetryConfigModel](ctx),
							PlanModifiers: []planmodifier.List{
								listplanmodifier.RequiresReplace(),
							},
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
							},
							NestedObject: schema.NestedBlockObject{
								Blocks: map[string]schema.Block{
									"logging_config": schema.ListNestedBlock{
										CustomType: fwtypes.NewListNestedObjectTypeOf[loggingConfigModel](ctx),
										PlanModifiers: []planmodifier.List{
											listplanmodifier.RequiresReplaceIfConfigured(),
										},
										Validators: []validator.List{
											listvalidator.SizeAtMost(1),
										},
										NestedObject: schema.NestedBlockObject{
											Attributes: map[string]schema.Attribute{
												"application_log_level": schema.StringAttribute{
													CustomType: fwtypes.StringEnumType[awstypes.ApplicationLogLevel](),
													Optional:   true,
													Computed:   true,
													PlanModifiers: []planmodifier.String{
														stringplanmodifier.RequiresReplaceIfConfigured(),
														stringplanmodifier.UseNonNullStateForUnknown(),
													},
												},
												"log_group": schema.StringAttribute{
													Optional: true,
													PlanModifiers: []planmodifier.String{
														stringplanmodifier.RequiresReplace(),
													},
												},
												"system_log_level": schema.StringAttribute{
													CustomType: fwtypes.StringEnumType[awstypes.SystemLogLevel](),
													Optional:   true,
													PlanModifiers: []planmodifier.String{
														stringplanmodifier.RequiresReplace(),
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
			names.AttrTimeouts: timeouts.Block(ctx, timeouts.Opts{
				Create: true,
			}),
		},
	}
}

func (r *functionRevisionResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan functionRevisionResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Plan.Get(ctx, &plan))
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().LambdaWebClient(ctx)

	functionName := fwflex.StringValueFromFramework(ctx, plan.FunctionName)
	var input lambdaweb.CreateWebFunctionRevisionInput
	smerr.AddEnrich(ctx, &resp.Diagnostics, fwflex.Expand(ctx, plan, &input))
	if resp.Diagnostics.HasError() {
		return
	}

	output, err := conn.CreateWebFunctionRevision(ctx, &input)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, functionName)
		return
	}

	revisionID := aws.ToString(output.RevisionId)
	out, err := waitFunctionRevisionActive(ctx, conn, functionName, revisionID, r.CreateTimeout(ctx, plan.Timeouts))
	if err != nil {
		// Taint the resource.
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("function_name"), functionName)...)
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("revision_id"), revisionID)...)
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, functionName)
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, r.flatten(ctx, out, &plan))
	if resp.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &plan))
}

func (r *functionRevisionResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state functionRevisionResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().LambdaWebClient(ctx)

	functionName, revisionID := fwflex.StringValueFromFramework(ctx, state.FunctionName), fwflex.StringValueFromFramework(ctx, state.RevisionID)
	out, err := findWebFunctionRevisionByTwoPartKey(ctx, conn, functionName, revisionID)
	if retry.NotFound(err) {
		resp.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, revisionID)
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, r.flatten(ctx, out, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &state))
}

func (r *functionRevisionResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state functionRevisionResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().LambdaWebClient(ctx)

	functionName, revisionID := fwflex.StringValueFromFramework(ctx, state.FunctionName), fwflex.StringValueFromFramework(ctx, state.RevisionID)
	input := lambdaweb.DeleteWebFunctionRevisionInput{
		FunctionName: aws.String(functionName),
		RevisionId:   aws.String(revisionID),
	}
	_, err := conn.DeleteWebFunctionRevision(ctx, &input)
	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return
	}
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, revisionID)
		return
	}
}

func (r *functionRevisionResource) flatten(ctx context.Context, revision *lambdaweb.GetWebFunctionRevisionOutput, data *functionRevisionResourceModel) diag.Diagnostics {
	var diags diag.Diagnostics
	diags.Append(fwflex.Flatten(ctx, revision, data)...)
	return diags
}

/*
// environmentVariablesMaxBytes is the documented 32 KB limit the service
// enforces on a revision's environment variables.
const environmentVariablesMaxBytes = 32 * 1024

// validateEnvironmentVariablesSize reports environment variables that cannot
// fit. The service rejects them, but only up to a point: past roughly 37 KB the
// whole request is too large and the error stops mentioning environment
// variables at all ("Request must be smaller than 37888 bytes"), which sends
// people looking in the wrong place. The keys and values are summed without the
// serialization overhead the service also counts, so this only ever rejects a
// configuration the service would reject too.
func validateEnvironmentVariablesSize(ctx context.Context, cfg functionResourceModel, resp *resource.ValidateConfigResponse) {
	if cfg.RevisionConfig.IsNull() || cfg.RevisionConfig.IsUnknown() {
		return
	}

	revisionConfig, d := cfg.RevisionConfig.ToPtr(ctx)
	if d.HasError() || revisionConfig == nil || revisionConfig.ServiceConfig.IsNull() || revisionConfig.ServiceConfig.IsUnknown() {
		return
	}

	serviceConfig, d := revisionConfig.ServiceConfig.ToPtr(ctx)
	if d.HasError() || serviceConfig == nil || serviceConfig.EnvironmentVariables.IsNull() || serviceConfig.EnvironmentVariables.IsUnknown() {
		return
	}

	var size int
	for name, value := range serviceConfig.EnvironmentVariables.Elements() {
		if value.IsUnknown() {
			return
		}
		size += len(name)
		if v, ok := value.(types.String); ok {
			size += len(v.ValueString())
		}
	}

	if size > environmentVariablesMaxBytes {
		resp.Diagnostics.AddAttributeError(path.Root("revision_config"),
			"Environment variables too large",
			fmt.Sprintf("`environment_variables` must not exceed %d bytes across all names and values; the configuration holds %d.",
				environmentVariablesMaxBytes, size))
	}
}
*/

/*
// restoreTelemetryConfig keeps telemetry_config as configuration-only state.
// The service assigns default telemetry (a log group and INFO log levels) to
// every revision and reports it back, but revisions are immutable, so
// telemetry the practitioner never wrote is a server default, not drift:
// reading it into state breaks every apply with "block count changed from 0 to
// 1", and a partially configured block (say, log_group only) would surface the
// filled-in log levels as a perpetual diff. prior is the configured
// revision_config (the plan on create, prior state on refresh); its
// telemetry_config replaces the server's wholesale. On import there is no
// prior, so telemetry_config is left unpopulated rather than importing the
// server defaults.
func restoreTelemetryConfig(ctx context.Context, revModel *revisionConfigModel, prior fwtypes.ListNestedObjectValueOf[revisionConfigModel]) {
	desired := fwtypes.NewListNestedObjectValueOfNull[telemetryConfigModel](ctx)
	if !prior.IsNull() && !prior.IsUnknown() {
		if priorRev, d := prior.ToPtr(ctx); !d.HasError() && priorRev != nil && !priorRev.ServiceConfig.IsNull() && !priorRev.ServiceConfig.IsUnknown() {
			if priorSvc, d := priorRev.ServiceConfig.ToPtr(ctx); !d.HasError() && priorSvc != nil && !priorSvc.TelemetryConfig.IsUnknown() {
				desired = priorSvc.TelemetryConfig
			}
		}
	}

	if revModel.ServiceConfig.IsNull() || revModel.ServiceConfig.IsUnknown() {
		return
	}
	svc, d := revModel.ServiceConfig.ToPtr(ctx)
	if d.HasError() || svc == nil {
		return
	}
	svc.TelemetryConfig = desired
	revModel.ServiceConfig = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, svc)
}
*/

var _ inttypes.ImportIDParser = functionRevisionImportID{}

type functionRevisionImportID struct{}

func (functionRevisionImportID) Parse(id string) (string, map[string]any, error) {
	const (
		importIDParts = 2
	)
	parts, err := intflex.ExpandResourceId(id, importIDParts, true)
	if err != nil {
		return "", nil, err
	}

	return id, map[string]any{
		"function_name": parts[0],
		"revision_id":   parts[1],
	}, nil
}

func listFunctionRevisions(ctx context.Context, conn *lambdaweb.Client, input *lambdaweb.ListWebFunctionRevisionsInput, optFns ...func(*lambdaweb.Options)) iter.Seq2[awstypes.FunctionRevisionSummary, error] {
	return tfiter.ConcatValuesWithError(listFunctionRevisionPages(ctx, conn, input, optFns...))
}

func listFunctionRevisionPages(ctx context.Context, conn *lambdaweb.Client, input *lambdaweb.ListWebFunctionRevisionsInput, optFns ...func(*lambdaweb.Options)) iter.Seq2[[]awstypes.FunctionRevisionSummary, error] {
	return func(yield func([]awstypes.FunctionRevisionSummary, error) bool) {
		pages := lambdaweb.NewListWebFunctionRevisionsPaginator(conn, input)
		for pages.HasMorePages() {
			page, err := pages.NextPage(ctx, optFns...)
			if err != nil {
				yield(nil, fmt.Errorf("listing Lambda Web Function Revisions: %w", err))
				return
			}

			if !yield(page.Revisions, nil) {
				return
			}
		}
	}
}

func findWebFunctionRevisionByTwoPartKey(ctx context.Context, conn *lambdaweb.Client, functionName, revisionID string) (*lambdaweb.GetWebFunctionRevisionOutput, error) {
	input := lambdaweb.GetWebFunctionRevisionInput{
		FunctionName: aws.String(functionName),
		RevisionId:   aws.String(revisionID),
	}
	return findWebFunctionRevision(ctx, conn, &input)
}

func findWebFunctionRevision(ctx context.Context, conn *lambdaweb.Client, input *lambdaweb.GetWebFunctionRevisionInput) (*lambdaweb.GetWebFunctionRevisionOutput, error) {
	out, err := conn.GetWebFunctionRevision(ctx, input)
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

	return out, nil
}

func statusFunctionRevision(conn *lambdaweb.Client, functionName, revisionID string) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		out, err := findWebFunctionRevisionByTwoPartKey(ctx, conn, functionName, revisionID)
		if retry.NotFound(err) {
			return nil, "", nil
		}
		if err != nil {
			return nil, "", smarterr.NewError(err)
		}

		return out, string(out.State), nil
	}
}

func waitFunctionRevisionActive(ctx context.Context, conn *lambdaweb.Client, functionName, revisionID string, timeout time.Duration) (*lambdaweb.GetWebFunctionRevisionOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   enum.Slice(awstypes.RevisionStatePending),
		Target:                    enum.Slice(awstypes.RevisionStateActive),
		Refresh:                   statusFunctionRevision(conn, functionName, revisionID),
		Timeout:                   timeout,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*lambdaweb.GetWebFunctionRevisionOutput); ok {
		retry.SetLastError(err, revisionError(out))
		return out, smarterr.NewError(err)
	}

	return nil, smarterr.NewError(err)
}

func findNewestFunctionRevision(ctx context.Context, conn *lambdaweb.Client, functionName string) (*awstypes.FunctionRevisionSummary, error) {
	var newest *awstypes.FunctionRevisionSummary
	input := lambdaweb.ListWebFunctionRevisionsInput{
		FunctionName: aws.String(functionName),
	}
	for item, err := range listFunctionRevisions(ctx, conn, &input) {
		if err != nil {
			return nil, smarterr.NewError(err)
		}

		if newest == nil || item.CreatedAt.After(aws.ToTime(newest.CreatedAt)) {
			newest = &item
		}
	}

	return newest, nil
}

type functionRevisionResourceModel struct {
	framework.WithRegionModel
	BuildConfig   fwtypes.ListNestedObjectValueOf[buildConfigModel]   `tfsdk:"build_config"`
	Description   types.String                                        `tfsdk:"description"`
	FunctionName  types.String                                        `tfsdk:"function_name"`
	KMSKeyARN     fwtypes.ARN                                         `tfsdk:"kms_key_arn"`
	RevisionARN   types.String                                        `tfsdk:"arn"`
	RevisionID    types.String                                        `tfsdk:"revision_id"`
	ServiceConfig fwtypes.ListNestedObjectValueOf[serviceConfigModel] `tfsdk:"service_config"`
	Timeouts      timeouts.Value                                      `tfsdk:"timeouts"`
}

type buildConfigModel struct {
	CodeConfig    fwtypes.ListNestedObjectValueOf[codeConfigModel]    `tfsdk:"code_config"`
	RuntimeConfig fwtypes.ListNestedObjectValueOf[runtimeConfigModel] `tfsdk:"runtime_config"`
}

type codeConfigModel struct {
	S3Object fwtypes.ListNestedObjectValueOf[s3ObjectModel] `tfsdk:"s3_object"`
}

type s3ObjectModel struct {
	Bucket    types.String `tfsdk:"bucket"`
	Key       types.String `tfsdk:"key"`
	VersionID types.String `tfsdk:"version_id"`
}

type runtimeConfigModel struct {
	Runtime types.String `tfsdk:"runtime"`
}

type serviceConfigModel struct {
	EgressNetworkConnectorARN    fwtypes.ARN                                           `tfsdk:"egress_network_connector_arn"`
	EnvironmentVariables         fwtypes.MapOfString                                   `tfsdk:"environment_variables"`
	ExecutionRoleARN             fwtypes.ARN                                           `tfsdk:"execution_role_arn"`
	MaxConcurrencyPerEnvironment types.Int64                                           `tfsdk:"max_concurrency_per_environment"`
	TelemetryConfig              fwtypes.ListNestedObjectValueOf[telemetryConfigModel] `tfsdk:"telemetry_config"`
	TimeoutSeconds               types.Int64                                           `tfsdk:"timeout_seconds"`
}

type telemetryConfigModel struct {
	LoggingConfig fwtypes.ListNestedObjectValueOf[loggingConfigModel] `tfsdk:"logging_config"`
}

type loggingConfigModel struct {
	ApplicationLogLevel fwtypes.StringEnum[awstypes.ApplicationLogLevel] `tfsdk:"application_log_level"`
	LogGroup            types.String                                     `tfsdk:"log_group"`
	SystemLogLevel      fwtypes.StringEnum[awstypes.SystemLogLevel]      `tfsdk:"system_log_level"`
}
