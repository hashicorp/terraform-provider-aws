// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: This is a GA-readiness skeleton. It targets the (not-yet-public)
// aws-sdk-go-v2 "webfunctions" service client and will not compile until that
// module exists. Use skaff to scaffold the real resource at GA. See
// CONTRIBUTION.md.

package webfunctions

import (
	"context"
	"errors"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/webfunctions"
	awstypes "github.com/aws/aws-sdk-go-v2/service/webfunctions/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_webfunctions_function", name="Function")
// @ArnIdentity
// @Testing(hasNoPreExistingResource=true)
// @Testing(existsType="github.com/aws/aws-sdk-go-v2/service/webfunctions;webfunctions.GetWebFunctionOutput")
// @Testing(importStateIdAttribute="function_name")
// @Testing(importIgnore="revision_config;endpoint_config")
func newFunctionResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &functionResource{}

	r.SetDefaultCreateTimeout(15 * time.Minute)
	r.SetDefaultUpdateTimeout(15 * time.Minute)
	r.SetDefaultDeleteTimeout(15 * time.Minute)

	return r, nil
}

const (
	ResNameFunction = "Function"
)

type functionResource struct {
	framework.ResourceWithModel[functionResourceModel]
	framework.WithTimeouts
}

func (r *functionResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrARN: framework.ARNAttributeComputedOnly(),
			names.AttrID:  framework.IDAttribute(),
			"function_name": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			names.AttrState: schema.StringAttribute{
				Computed: true,
			},
			"latest_revision_id": schema.StringAttribute{
				Computed: true,
			},
			"domain_name": schema.StringAttribute{
				Computed: true,
			},
		},
		Blocks: map[string]schema.Block{
			"revision_config": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[revisionConfigModel](ctx),
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						names.AttrDescription: schema.StringAttribute{
							Optional: true,
						},
						"kms_key_arn": schema.StringAttribute{
							CustomType: fwtypes.ARNType,
							Optional:   true,
						},
					},
					Blocks: map[string]schema.Block{
						"build_config": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[buildConfigModel](ctx),
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"runtime": schema.StringAttribute{
										Required: true,
									},
								},
								Blocks: map[string]schema.Block{
									"code_config": schema.ListNestedBlock{
										CustomType: fwtypes.NewListNestedObjectTypeOf[codeConfigModel](ctx),
										Validators: []validator.List{
											listvalidator.SizeAtMost(1),
										},
										NestedObject: schema.NestedBlockObject{
											Attributes: map[string]schema.Attribute{
												"zip_file": schema.StringAttribute{
													Optional:  true,
													Sensitive: true,
												},
											},
											Blocks: map[string]schema.Block{
												"s3_object": schema.ListNestedBlock{
													CustomType: fwtypes.NewListNestedObjectTypeOf[s3ObjectModel](ctx),
													Validators: []validator.List{
														listvalidator.SizeAtMost(1),
													},
													NestedObject: schema.NestedBlockObject{
														Attributes: map[string]schema.Attribute{
															names.AttrBucket: schema.StringAttribute{
																Required: true,
															},
															names.AttrKey: schema.StringAttribute{
																Required: true,
															},
															"version_id": schema.StringAttribute{
																Optional: true,
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
						"service_config": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[serviceConfigModel](ctx),
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"execution_role_arn": schema.StringAttribute{
										CustomType: fwtypes.ARNType,
										Required:   true,
									},
									"timeout_seconds": schema.Int64Attribute{
										Optional: true,
										Validators: []validator.Int64{
											int64validator.Between(3, 900),
										},
									},
									"max_concurrency_per_environment": schema.Int64Attribute{
										Optional: true,
										Validators: []validator.Int64{
											int64validator.Between(1, 128),
										},
									},
									"environment_variables": schema.MapAttribute{
										CustomType:  fwtypes.MapOfStringType,
										Optional:    true,
										Sensitive:   true,
										ElementType: types.StringType,
									},
								},
								Blocks: map[string]schema.Block{
									"telemetry_config": schema.ListNestedBlock{
										CustomType: fwtypes.NewListNestedObjectTypeOf[telemetryConfigModel](ctx),
										Validators: []validator.List{
											listvalidator.SizeAtMost(1),
										},
										NestedObject: schema.NestedBlockObject{
											Blocks: map[string]schema.Block{
												"logging_config": schema.ListNestedBlock{
													CustomType: fwtypes.NewListNestedObjectTypeOf[loggingConfigModel](ctx),
													Validators: []validator.List{
														listvalidator.SizeAtMost(1),
													},
													NestedObject: schema.NestedBlockObject{
														Attributes: map[string]schema.Attribute{
															"log_group": schema.StringAttribute{
																Optional: true,
															},
															"application_log_level": schema.StringAttribute{
																CustomType: fwtypes.StringEnumType[awstypes.ApplicationLogLevel](),
																Optional:   true,
															},
															"system_log_level": schema.StringAttribute{
																CustomType: fwtypes.StringEnumType[awstypes.SystemLogLevel](),
																Optional:   true,
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
				},
			},
			"endpoint_config": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[endpointConfigModel](ctx),
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"endpoint_name": schema.StringAttribute{
							Required: true,
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.RequiresReplace(),
							},
						},
						"endpoint_type": schema.StringAttribute{
							CustomType: fwtypes.StringEnumType[awstypes.EndpointType](),
							Required:   true,
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.RequiresReplace(),
							},
						},
						"auth_type": schema.StringAttribute{
							CustomType: fwtypes.StringEnumType[awstypes.AuthType](),
							Required:   true,
						},
						"auto_deployment_mode": schema.StringAttribute{
							CustomType: fwtypes.StringEnumType[awstypes.AutoDeploymentMode](),
							Optional:   true,
							Computed:   true,
						},
						names.AttrDescription: schema.StringAttribute{
							Optional: true,
						},
						"regions": schema.ListAttribute{
							CustomType:  fwtypes.ListOfStringType,
							Optional:    true,
							ElementType: types.StringType,
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

func (r *functionResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := r.Meta().WebFunctionsClient(ctx)

	var plan functionResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Plan.Get(ctx, &plan))
	if resp.Diagnostics.HasError() {
		return
	}

	var input webfunctions.CreateWebFunctionInput
	smerr.AddEnrich(ctx, &resp.Diagnostics, fwflex.Expand(ctx, plan, &input))
	if resp.Diagnostics.HasError() {
		return
	}

	name := plan.FunctionName.ValueString()

	_, err := conn.CreateWebFunction(ctx, &input)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, name)
		return
	}

	out, err := waitFunctionCreated(ctx, conn, name, r.CreateTimeout(ctx, plan.Timeouts))
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, name)
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, fwflex.Flatten(ctx, out, &plan))
	if resp.Diagnostics.HasError() {
		return
	}

	if !plan.EndpointConfig.IsNull() {
		endpointName, err := endpointNameFromConfig(ctx, plan.EndpointConfig)
		if err != nil {
			smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, name)
			return
		}

		ep, err := waitEndpointActive(ctx, conn, name, endpointName, r.CreateTimeout(ctx, plan.Timeouts))
		if err != nil {
			smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, name)
			return
		}
		plan.DomainName = fwflex.StringToFramework(ctx, ep.DomainName)
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &plan))
}

func (r *functionResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().WebFunctionsClient(ctx)

	var state functionResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := findFunctionByName(ctx, conn, state.FunctionName.ValueString())
	if retry.NotFound(err) {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, state.FunctionName.ValueString())
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, fwflex.Flatten(ctx, out, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &state))
}

func (r *functionResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	conn := r.Meta().WebFunctionsClient(ctx)

	var plan, state functionResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Plan.Get(ctx, &plan))
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	name := plan.FunctionName.ValueString()

	if !plan.RevisionConfig.Equal(state.RevisionConfig) {
		var input webfunctions.CreateWebFunctionRevisionInput
		smerr.AddEnrich(ctx, &resp.Diagnostics, fwflex.Expand(ctx, plan, &input))
		if resp.Diagnostics.HasError() {
			return
		}
		input.FunctionName = aws.String(name)

		if _, err := conn.CreateWebFunctionRevision(ctx, &input); err != nil {
			smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, name)
			return
		}
	}

	if !plan.EndpointConfig.Equal(state.EndpointConfig) && !plan.EndpointConfig.IsNull() {
		endpointName, err := endpointNameFromConfig(ctx, plan.EndpointConfig)
		if err != nil {
			smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, name)
			return
		}

		var input webfunctions.UpdateWebFunctionEndpointInput
		smerr.AddEnrich(ctx, &resp.Diagnostics, fwflex.Expand(ctx, plan, &input))
		if resp.Diagnostics.HasError() {
			return
		}
		input.FunctionName = aws.String(name)
		input.EndpointName = aws.String(endpointName)

		if _, err := conn.UpdateWebFunctionEndpoint(ctx, &input); err != nil {
			smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, name)
			return
		}
	}

	out, err := waitFunctionUpdated(ctx, conn, name, r.UpdateTimeout(ctx, plan.Timeouts))
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, name)
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, fwflex.Flatten(ctx, out, &plan))
	if resp.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &plan))
}

func (r *functionResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := r.Meta().WebFunctionsClient(ctx)

	var state functionResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	name := state.FunctionName.ValueString()

	_, err := conn.DeleteWebFunction(ctx, &webfunctions.DeleteWebFunctionInput{
		FunctionName: aws.String(name),
	})
	if err != nil {
		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return
		}
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, name)
		return
	}

	if _, err := waitFunctionDeleted(ctx, conn, name, r.DeleteTimeout(ctx, state.Timeouts)); err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, name)
		return
	}
}

func (r *functionResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("function_name"), req, resp)
}

func endpointNameFromConfig(ctx context.Context, l fwtypes.ListNestedObjectValueOf[endpointConfigModel]) (string, error) {
	models, d := l.ToSlice(ctx)
	if d.HasError() {
		return "", errors.New("reading endpoint_config")
	}
	if len(models) == 0 {
		return "", errors.New("empty endpoint_config")
	}
	return models[0].EndpointName.ValueString(), nil
}

type functionResourceModel struct {
	framework.WithRegionModel
	ARN              types.String                                         `tfsdk:"arn"`
	DomainName       types.String                                         `tfsdk:"domain_name"`
	EndpointConfig   fwtypes.ListNestedObjectValueOf[endpointConfigModel] `tfsdk:"endpoint_config"`
	FunctionName     types.String                                         `tfsdk:"function_name"`
	ID               types.String                                         `tfsdk:"id"`
	LatestRevisionID types.String                                         `tfsdk:"latest_revision_id"`
	RevisionConfig   fwtypes.ListNestedObjectValueOf[revisionConfigModel] `tfsdk:"revision_config"`
	State            types.String                                         `tfsdk:"state"`
	Timeouts         timeouts.Value                                       `tfsdk:"timeouts"`
}

type revisionConfigModel struct {
	Description   types.String                                        `tfsdk:"description"`
	KMSKeyARN     fwtypes.ARN                                         `tfsdk:"kms_key_arn"`
	BuildConfig   fwtypes.ListNestedObjectValueOf[buildConfigModel]   `tfsdk:"build_config"`
	ServiceConfig fwtypes.ListNestedObjectValueOf[serviceConfigModel] `tfsdk:"service_config"`
}

type buildConfigModel struct {
	CodeConfig    fwtypes.ListNestedObjectValueOf[codeConfigModel]    `tfsdk:"code_config"`
	RuntimeConfig fwtypes.ListNestedObjectValueOf[runtimeConfigModel] `tfsdk:"runtime_config"`
	Runtime       types.String                                        `tfsdk:"runtime"`
}

type codeConfigModel struct {
	S3Object fwtypes.ListNestedObjectValueOf[s3ObjectModel] `tfsdk:"s3_object"`
	ZipFile  types.String                                   `tfsdk:"zip_file"`
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
	ExecutionRoleARN             fwtypes.ARN                                           `tfsdk:"execution_role_arn"`
	TimeoutSeconds               types.Int64                                           `tfsdk:"timeout_seconds"`
	MaxConcurrencyPerEnvironment types.Int64                                           `tfsdk:"max_concurrency_per_environment"`
	EnvironmentVariables         fwtypes.MapOfString                                   `tfsdk:"environment_variables"`
	TelemetryConfig              fwtypes.ListNestedObjectValueOf[telemetryConfigModel] `tfsdk:"telemetry_config"`
}

type telemetryConfigModel struct {
	LoggingConfig fwtypes.ListNestedObjectValueOf[loggingConfigModel] `tfsdk:"logging_config"`
}

type loggingConfigModel struct {
	ApplicationLogLevel fwtypes.StringEnum[awstypes.ApplicationLogLevel] `tfsdk:"application_log_level"`
	LogGroup            types.String                                     `tfsdk:"log_group"`
	SystemLogLevel      fwtypes.StringEnum[awstypes.SystemLogLevel]      `tfsdk:"system_log_level"`
}

type endpointConfigModel struct {
	EndpointName       types.String                                    `tfsdk:"endpoint_name"`
	Description        types.String                                    `tfsdk:"description"`
	EndpointType       fwtypes.StringEnum[awstypes.EndpointType]       `tfsdk:"endpoint_type"`
	AuthType           fwtypes.StringEnum[awstypes.AuthType]           `tfsdk:"auth_type"`
	AutoDeploymentMode fwtypes.StringEnum[awstypes.AutoDeploymentMode] `tfsdk:"auto_deployment_mode"`
	Regions            fwtypes.ListOfString                            `tfsdk:"regions"`
}
