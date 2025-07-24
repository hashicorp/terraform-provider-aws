// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package appsync

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/YakDriver/smarterr"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/appsync"
	awstypes "github.com/aws/aws-sdk-go-v2/service/appsync/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/mapplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_appsync_api", name="API")
// @Tags(identifierAttribute="arn")
// @IdentityAttribute("id")
// @Testing(importStateIdAttribute="id")
// @Testing(existsType="github.com/aws/aws-sdk-go-v2/service/appsync/types;awstypes;awstypes.Api")
// @Testing(preIdentityVersion="v6.3.0")
func newAPIResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &resourceAPI{}

	r.SetDefaultCreateTimeout(30 * time.Minute)
	r.SetDefaultUpdateTimeout(30 * time.Minute)
	r.SetDefaultDeleteTimeout(30 * time.Minute)

	return r, nil
}

const (
	ResNameAPI = "API"
)

type resourceAPI struct {
	framework.ResourceWithModel[resourceAPIModel]
	framework.WithTimeouts
}

func (r *resourceAPI) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrARN: framework.ARNAttributeComputedOnly(),
			"id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"created": schema.StringAttribute{
				CustomType: timetypes.RFC3339Type{},
				Computed:   true,
			},
			"dns": schema.MapAttribute{
				CustomType: fwtypes.MapOfStringType,
				Computed:   true,
				PlanModifiers: []planmodifier.Map{
					mapplanmodifier.UseStateForUnknown(),
				},
			},
			names.AttrName: schema.StringAttribute{
				Required: true,
			},
			"owner_contact": schema.StringAttribute{
				Optional: true,
			},
			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
			"waf_web_acl_arn": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"xray_enabled": schema.BoolAttribute{
				Computed: true,
			},
		},
		Blocks: map[string]schema.Block{
			"event_config": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[eventConfigModel](ctx),
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
				},
				NestedObject: schema.NestedBlockObject{
					Blocks: map[string]schema.Block{
						"auth_providers": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[authProviderModel](ctx),
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"auth_type": schema.StringAttribute{
										Required:   true,
										CustomType: fwtypes.StringEnumType[awstypes.AuthenticationType](),
									},
								},
								Blocks: map[string]schema.Block{
									"cognito_config": schema.ListNestedBlock{
										CustomType: fwtypes.NewListNestedObjectTypeOf[cognitoConfigModel](ctx),
										Validators: []validator.List{
											listvalidator.SizeAtMost(1),
										},
										NestedObject: schema.NestedBlockObject{
											Attributes: map[string]schema.Attribute{
												"aws_region": schema.StringAttribute{
													Required: true,
												},
												"user_pool_id": schema.StringAttribute{
													Required: true,
												},
												"app_id_client_regex": schema.StringAttribute{
													Optional: true,
												},
											},
										},
									},
									"lambda_authorizer_config": schema.ListNestedBlock{
										CustomType: fwtypes.NewListNestedObjectTypeOf[lambdaAuthorizerConfigModel](ctx),
										Validators: []validator.List{
											listvalidator.SizeAtMost(1),
										},
										NestedObject: schema.NestedBlockObject{
											Attributes: map[string]schema.Attribute{
												"authorizer_result_ttl_in_seconds": schema.Int64Attribute{
													Optional: true,
													Computed: true,
													PlanModifiers: []planmodifier.Int64{
														int64planmodifier.UseStateForUnknown(),
													},
												},
												"authorizer_uri": schema.StringAttribute{
													Required: true,
												},
												"identity_validation_expression": schema.StringAttribute{
													Optional: true,
												},
											},
										},
									},
									"openid_connect_config": schema.ListNestedBlock{
										CustomType: fwtypes.NewListNestedObjectTypeOf[openIDConnectConfigModel](ctx),
										Validators: []validator.List{
											listvalidator.SizeAtMost(1),
										},
										NestedObject: schema.NestedBlockObject{
											Attributes: map[string]schema.Attribute{
												"auth_ttl": schema.Int64Attribute{
													Optional: true,
													Computed: true,
													PlanModifiers: []planmodifier.Int64{
														int64planmodifier.UseStateForUnknown(),
													},
												},
												names.AttrClientID: schema.StringAttribute{
													Optional: true,
												},
												"iat_ttl": schema.Int64Attribute{
													Optional: true,
													Computed: true,
													PlanModifiers: []planmodifier.Int64{
														int64planmodifier.UseStateForUnknown(),
													},
												},
												names.AttrIssuer: schema.StringAttribute{
													Required: true,
												},
											},
										},
									},
								},
							},
						},
						"connection_auth_modes": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[authModeModel](ctx),
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"auth_type": schema.StringAttribute{
										Required:   true,
										CustomType: fwtypes.StringEnumType[awstypes.AuthenticationType](),
									},
								},
							},
						},
						"default_publish_auth_modes": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[authModeModel](ctx),
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"auth_type": schema.StringAttribute{
										Required:   true,
										CustomType: fwtypes.StringEnumType[awstypes.AuthenticationType](),
									},
								},
							},
						},
						"default_subscribe_auth_modes": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[authModeModel](ctx),
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"auth_type": schema.StringAttribute{
										Required:   true,
										CustomType: fwtypes.StringEnumType[awstypes.AuthenticationType](),
									},
								},
							},
						},
						"log_config": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[eventLogConfigModel](ctx),
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"cloudwatch_logs_role_arn": schema.StringAttribute{
										Required: true,
									},
									"log_level": schema.StringAttribute{
										Required:   true,
										CustomType: fwtypes.StringEnumType[awstypes.EventLogLevel](),
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

func (r *resourceAPI) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := r.Meta().AppSyncClient(ctx)

	var plan resourceAPIModel
	smerr.EnrichAppend(ctx, &resp.Diagnostics, req.Plan.Get(ctx, &plan))
	if resp.Diagnostics.HasError() {
		return
	}

	var input appsync.CreateApiInput
	smerr.EnrichAppend(ctx, &resp.Diagnostics, flex.Expand(ctx, plan, &input))
	if resp.Diagnostics.HasError() {
		return
	}

	input.Tags = getTagsIn(ctx)

	out, err := conn.CreateApi(ctx, &input)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, plan.Name.String())
		return
	}
	if out == nil || out.Api == nil {
		smerr.AddError(ctx, &resp.Diagnostics, errors.New("empty output"), smerr.ID, plan.Name.String())
		return
	}

	smerr.EnrichAppend(ctx, &resp.Diagnostics, flex.Flatten(ctx, out.Api, &plan), smerr.ID, plan.Name.String())
	if resp.Diagnostics.HasError() {
		return
	}

	createTimeout := r.CreateTimeout(ctx, plan.Timeouts)
	_, err = waitAPICreated(ctx, conn, plan.ApiId.ValueString(), createTimeout)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, plan.Name.String())
		return
	}

	smerr.EnrichAppend(ctx, &resp.Diagnostics, resp.State.Set(ctx, plan), smerr.ID, plan.ApiId.String())
}

func (r *resourceAPI) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().AppSyncClient(ctx)

	var state resourceAPIModel
	smerr.EnrichAppend(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	apiId := state.ApiId.ValueString()
	fmt.Println("hello", apiId)

	out, err := findAPIByID(ctx, conn, apiId)
	if tfresource.NotFound(err) {
		resp.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, state.ApiId.String())
		return
	}

	smerr.EnrichAppend(ctx, &resp.Diagnostics, flex.Flatten(ctx, out, &state), smerr.ID, state.ApiId.String())
	if resp.Diagnostics.HasError() {
		return
	}

	smerr.EnrichAppend(ctx, &resp.Diagnostics, resp.State.Set(ctx, &state), smerr.ID, state.ApiId.String())
}

func (r *resourceAPI) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	conn := r.Meta().AppSyncClient(ctx)

	var plan, state resourceAPIModel
	smerr.EnrichAppend(ctx, &resp.Diagnostics, req.Plan.Get(ctx, &plan))
	smerr.EnrichAppend(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	diff, d := flex.Diff(ctx, plan, state)
	smerr.EnrichAppend(ctx, &resp.Diagnostics, d, smerr.ID, plan.ApiId.String())
	if resp.Diagnostics.HasError() {
		return
	}

	if diff.HasChanges() {
		var input appsync.UpdateApiInput
		smerr.EnrichAppend(ctx, &resp.Diagnostics, flex.Expand(ctx, plan, &input), smerr.ID, plan.ApiId.String())
		if resp.Diagnostics.HasError() {
			return
		}

		input.ApiId = plan.ApiId.ValueStringPointer()

		out, err := conn.UpdateApi(ctx, &input)
		if err != nil {
			smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, plan.ApiId.String())
			return
		}
		if out == nil || out.Api == nil {
			smerr.AddError(ctx, &resp.Diagnostics, errors.New("empty output"), smerr.ID, plan.ApiId.String())
			return
		}

		smerr.EnrichAppend(ctx, &resp.Diagnostics, flex.Flatten(ctx, out.Api, &plan), smerr.ID, plan.ApiId.String())
		if resp.Diagnostics.HasError() {
			return
		}
	}

	smerr.EnrichAppend(ctx, &resp.Diagnostics, resp.State.Set(ctx, &plan), smerr.ID, plan.ApiId.String())
}

func (r *resourceAPI) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := r.Meta().AppSyncClient(ctx)

	var state resourceAPIModel
	smerr.EnrichAppend(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	input := appsync.DeleteApiInput{
		ApiId: state.ApiId.ValueStringPointer(),
	}

	_, err := conn.DeleteApi(ctx, &input)
	if err != nil {
		if errs.IsA[*awstypes.NotFoundException](err) {
			return
		}

		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, state.ApiId.String())
		return
	}

	deleteTimeout := r.DeleteTimeout(ctx, state.Timeouts)
	_, err = waitAPIDeleted(ctx, conn, state.ApiId.ValueString(), deleteTimeout)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, state.ApiId.String())
		return
	}
}

func (r *resourceAPI) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

const (
	statusDeleting = "Deleting"
	statusNormal   = "Normal"
)

func waitAPICreated(ctx context.Context, conn *appsync.Client, id string, timeout time.Duration) (*awstypes.Api, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   []string{},
		Target:                    []string{statusNormal},
		Refresh:                   statusAPI(ctx, conn, id),
		Timeout:                   timeout,
		NotFoundChecks:            20,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if err != nil {
		fmt.Printf("[DEBUG] waitAPICreated: Wait failed with error: %v\n", err)
	} else {
		fmt.Printf("[DEBUG] waitAPICreated: Wait completed successfully for API: %s\n", id)
	}

	if out, ok := outputRaw.(*awstypes.Api); ok {
		return out, smarterr.NewError(err)
	}

	return nil, smarterr.NewError(err)
}

func waitAPIDeleted(ctx context.Context, conn *appsync.Client, id string, timeout time.Duration) (*awstypes.Api, error) {
	fmt.Printf("[DEBUG] waitAPIDeleted: Starting wait for API deletion: %s\n", id)
	stateConf := &retry.StateChangeConf{
		Pending: []string{statusDeleting, statusNormal},
		Target:  []string{},
		Refresh: statusAPI(ctx, conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if err != nil {
		fmt.Printf("[DEBUG] waitAPIDeleted: Wait failed with error: %v\n", err)
	} else {
		fmt.Printf("[DEBUG] waitAPIDeleted: Wait completed successfully for API: %s\n", id)
	}

	if out, ok := outputRaw.(*awstypes.Api); ok {
		return out, smarterr.NewError(err)
	}

	return nil, smarterr.NewError(err)
}

func statusAPI(ctx context.Context, conn *appsync.Client, id string) retry.StateRefreshFunc {
	return func() (any, string, error) {
		out, err := findAPIByID(ctx, conn, id)
		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", smarterr.NewError(err)
		}

		return out, statusNormal, nil
	}
}

func findAPIByID(ctx context.Context, conn *appsync.Client, id string) (*awstypes.Api, error) {
	input := appsync.GetApiInput{
		ApiId: aws.String(id),
	}

	out, err := conn.GetApi(ctx, &input)
	if err != nil {
		if errs.IsA[*awstypes.NotFoundException](err) {
			return nil, smarterr.NewError(&retry.NotFoundError{
				LastError:   err,
				LastRequest: &input,
			})
		}
		return nil, smarterr.NewError(err)
	}

	if out == nil || out.Api == nil {
		return nil, smarterr.NewError(tfresource.NewEmptyResultError(&input))
	}

	return out.Api, nil
}

type resourceAPIModel struct {
	framework.WithRegionModel
	ApiArn       types.String                                      `tfsdk:"arn"`
	ApiId        types.String                                      `tfsdk:"id"`
	Created      timetypes.RFC3339                                 `tfsdk:"created"`
	Dns          fwtypes.MapOfString                               `tfsdk:"dns"`
	EventConfig  fwtypes.ListNestedObjectValueOf[eventConfigModel] `tfsdk:"event_config"`
	Name         types.String                                      `tfsdk:"name"`
	OwnerContact types.String                                      `tfsdk:"owner_contact"`
	Tags         tftags.Map                                        `tfsdk:"tags"`
	TagsAll      tftags.Map                                        `tfsdk:"tags_all"`
	Timeouts     timeouts.Value                                    `tfsdk:"timeouts"`
	WafWebAclArn types.String                                      `tfsdk:"waf_web_acl_arn"`
	XrayEnabled  types.Bool                                        `tfsdk:"xray_enabled"`
}

type eventConfigModel struct {
	AuthProviders             fwtypes.ListNestedObjectValueOf[authProviderModel]   `tfsdk:"auth_providers"`
	ConnectionAuthModes       fwtypes.ListNestedObjectValueOf[authModeModel]       `tfsdk:"connection_auth_modes"`
	DefaultPublishAuthModes   fwtypes.ListNestedObjectValueOf[authModeModel]       `tfsdk:"default_publish_auth_modes"`
	DefaultSubscribeAuthModes fwtypes.ListNestedObjectValueOf[authModeModel]       `tfsdk:"default_subscribe_auth_modes"`
	LogConfig                 fwtypes.ListNestedObjectValueOf[eventLogConfigModel] `tfsdk:"log_config"`
}

type authProviderModel struct {
	AuthType               fwtypes.StringEnum[awstypes.AuthenticationType]              `tfsdk:"auth_type"`
	CognitoConfig          fwtypes.ListNestedObjectValueOf[cognitoConfigModel]          `tfsdk:"cognito_config"`
	LambdaAuthorizerConfig fwtypes.ListNestedObjectValueOf[lambdaAuthorizerConfigModel] `tfsdk:"lambda_authorizer_config"`
	OpenIDConnectConfig    fwtypes.ListNestedObjectValueOf[openIDConnectConfigModel]    `tfsdk:"openid_connect_config"`
}

type authModeModel struct {
	AuthType fwtypes.StringEnum[awstypes.AuthenticationType] `tfsdk:"auth_type"`
}

type cognitoConfigModel struct {
	AwsRegion        types.String `tfsdk:"aws_region"`
	UserPoolId       types.String `tfsdk:"user_pool_id"`
	AppIdClientRegex types.String `tfsdk:"app_id_client_regex"`
}

type lambdaAuthorizerConfigModel struct {
	AuthorizerResultTtlInSeconds types.Int64  `tfsdk:"authorizer_result_ttl_in_seconds"`
	AuthorizerUri                types.String `tfsdk:"authorizer_uri"`
	IdentityValidationExpression types.String `tfsdk:"identity_validation_expression"`
}

type openIDConnectConfigModel struct {
	AuthTtl  types.Int64  `tfsdk:"auth_ttl"`
	ClientId types.String `tfsdk:"client_id"`
	IatTtl   types.Int64  `tfsdk:"iat_ttl"`
	Issuer   types.String `tfsdk:"issuer"`
}

type eventLogConfigModel struct {
	CloudWatchLogsRoleArn types.String                               `tfsdk:"cloudwatch_logs_role_arn"`
	LogLevel              fwtypes.StringEnum[awstypes.EventLogLevel] `tfsdk:"log_level"`
}
