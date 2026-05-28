// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

package appsync

import (
	"context"

	"github.com/YakDriver/regexache"
	"github.com/YakDriver/smarterr"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/appsync"
	awstypes "github.com/aws/aws-sdk-go-v2/service/appsync/types"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/mapplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
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

// @FrameworkResource("aws_appsync_api", name="API")
// @Tags(identifierAttribute="api_arn")
// @Testing(importStateIdAttribute="api_id")
// @Testing(existsType="github.com/aws/aws-sdk-go-v2/service/appsync/types;awstypes;awstypes.Api")
func newAPIResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &apiResource{}

	return r, nil
}

type apiResource struct {
	framework.ResourceWithModel[apiResourceModel]
}

func (r *apiResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"api_arn": framework.ARNAttributeComputedOnly(),
			"api_id":  framework.IDAttribute(),
			"dns": schema.MapAttribute{
				CustomType: fwtypes.MapOfStringType,
				Computed:   true,
				PlanModifiers: []planmodifier.Map{
					mapplanmodifier.UseStateForUnknown(),
				},
			},
			names.AttrName: schema.StringAttribute{
				Required: true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 50),
					stringvalidator.RegexMatches(regexache.MustCompile(`^[A-Za-z0-9_\-\ ]+$`), ""),
				},
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
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
		},
		Blocks: map[string]schema.Block{
			"event_config": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[eventConfigModel](ctx),
				Validators: []validator.List{
					listvalidator.IsRequired(),
					listvalidator.SizeAtLeast(1),
					listvalidator.SizeAtMost(1),
				},
				NestedObject: schema.NestedBlockObject{
					Blocks: map[string]schema.Block{
						"auth_provider": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[authProviderModel](ctx),
							Validators: []validator.List{
								listvalidator.IsRequired(),
								listvalidator.SizeAtLeast(1),
							},
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
												"app_id_client_regex": schema.StringAttribute{
													Optional: true,
												},
												"aws_region": schema.StringAttribute{
													Required: true,
													Validators: []validator.String{
														fwvalidators.AWSRegion(),
													},
												},
												names.AttrUserPoolID: schema.StringAttribute{
													Required: true,
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
						"connection_auth_mode": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[authModeModel](ctx),
							Validators: []validator.List{
								listvalidator.IsRequired(),
								listvalidator.SizeAtLeast(1),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"auth_type": schema.StringAttribute{
										Required:   true,
										CustomType: fwtypes.StringEnumType[awstypes.AuthenticationType](),
									},
								},
							},
						},
						"default_publish_auth_mode": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[authModeModel](ctx),
							Validators: []validator.List{
								listvalidator.IsRequired(),
								listvalidator.SizeAtLeast(1),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"auth_type": schema.StringAttribute{
										Required:   true,
										CustomType: fwtypes.StringEnumType[awstypes.AuthenticationType](),
									},
								},
							},
						},
						"default_subscribe_auth_mode": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[authModeModel](ctx),
							Validators: []validator.List{
								listvalidator.IsRequired(),
								listvalidator.SizeAtLeast(1),
							},
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
										CustomType: fwtypes.ARNType,
										Required:   true,
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
		},
	}
}

func (r *apiResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data apiResourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.Plan.Get(ctx, &data))
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().AppSyncClient(ctx)

	name := fwflex.StringValueFromFramework(ctx, data.Name)
	var input appsync.CreateApiInput
	smerr.AddEnrich(ctx, &response.Diagnostics, fwflex.Expand(ctx, data, &input))
	if response.Diagnostics.HasError() {
		return
	}

	// Additional fields.
	input.Tags = getTagsIn(ctx)

	output, err := conn.CreateApi(ctx, &input)

	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, name)
		return
	}

	// Set values for unknowns.
	api := output.Api
	apiID := aws.ToString(api.ApiId)
	smerr.AddEnrich(ctx, &response.Diagnostics, fwflex.Flatten(ctx, api, &data), smerr.ID, name)
	if response.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, response.State.Set(ctx, data), smerr.ID, apiID)
}

func (r *apiResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data apiResourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.State.Get(ctx, &data))
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().AppSyncClient(ctx)

	apiID := fwflex.StringValueFromFramework(ctx, data.ApiID)
	output, err := findAPIByID(ctx, conn, apiID)

	if retry.NotFound(err) {
		smerr.AddOne(ctx, &response.Diagnostics, fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)
		return
	}

	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, apiID)
		return
	}

	// Set attributes for import.
	smerr.AddEnrich(ctx, &response.Diagnostics, fwflex.Flatten(ctx, output, &data), smerr.ID, apiID)
	if response.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, response.State.Set(ctx, &data), smerr.ID, apiID)
}

func (r *apiResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var new, old apiResourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.Plan.Get(ctx, &new))
	if response.Diagnostics.HasError() {
		return
	}
	smerr.AddEnrich(ctx, &response.Diagnostics, request.State.Get(ctx, &old))
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().AppSyncClient(ctx)

	apiID := fwflex.StringValueFromFramework(ctx, new.ApiID)
	diff, d := fwflex.Diff(ctx, new, old)
	smerr.AddEnrich(ctx, &response.Diagnostics, d, smerr.ID, apiID)
	if response.Diagnostics.HasError() {
		return
	}

	if diff.HasChanges() {
		var input appsync.UpdateApiInput
		smerr.AddEnrich(ctx, &response.Diagnostics, fwflex.Expand(ctx, new, &input), smerr.ID, apiID)
		if response.Diagnostics.HasError() {
			return
		}

		_, err := conn.UpdateApi(ctx, &input)

		if err != nil {
			smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, apiID)
			return
		}
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, response.State.Set(ctx, &new), smerr.ID, apiID)
}

func (r *apiResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data apiResourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.State.Get(ctx, &data))
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().AppSyncClient(ctx)

	apiID := fwflex.StringValueFromFramework(ctx, data.ApiID)
	input := appsync.DeleteApiInput{
		ApiId: aws.String(apiID),
	}
	_, err := conn.DeleteApi(ctx, &input)

	if errs.IsA[*awstypes.NotFoundException](err) {
		return
	}

	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, apiID)
		return
	}
}

func (r *apiResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("api_id"), request, response)
}

func findAPIByID(ctx context.Context, conn *appsync.Client, id string) (*awstypes.Api, error) {
	input := appsync.GetApiInput{
		ApiId: aws.String(id),
	}

	return findAPI(ctx, conn, &input)
}

func findAPI(ctx context.Context, conn *appsync.Client, input *appsync.GetApiInput) (*awstypes.Api, error) {
	output, err := conn.GetApi(ctx, input)

	if errs.IsA[*awstypes.NotFoundException](err) {
		return nil, smarterr.NewError(&retry.NotFoundError{
			LastError: err,
		})
	}

	if err != nil {
		return nil, smarterr.NewError(err)
	}

	if output == nil || output.Api == nil {
		return nil, smarterr.NewError(tfresource.NewEmptyResultError())
	}

	return output.Api, nil
}

type apiResourceModel struct {
	framework.WithRegionModel
	ApiARN       types.String                                      `tfsdk:"api_arn"`
	ApiID        types.String                                      `tfsdk:"api_id"`
	DNS          fwtypes.MapOfString                               `tfsdk:"dns"`
	EventConfig  fwtypes.ListNestedObjectValueOf[eventConfigModel] `tfsdk:"event_config"`
	Name         types.String                                      `tfsdk:"name"`
	OwnerContact types.String                                      `tfsdk:"owner_contact"`
	Tags         tftags.Map                                        `tfsdk:"tags"`
	TagsAll      tftags.Map                                        `tfsdk:"tags_all"`
	WAFWebAclARN types.String                                      `tfsdk:"waf_web_acl_arn"`
	XRayEnabled  types.Bool                                        `tfsdk:"xray_enabled"`
}

type eventConfigModel struct {
	AuthProviders             fwtypes.ListNestedObjectValueOf[authProviderModel]   `tfsdk:"auth_provider"`
	ConnectionAuthModes       fwtypes.ListNestedObjectValueOf[authModeModel]       `tfsdk:"connection_auth_mode"`
	DefaultPublishAuthModes   fwtypes.ListNestedObjectValueOf[authModeModel]       `tfsdk:"default_publish_auth_mode"`
	DefaultSubscribeAuthModes fwtypes.ListNestedObjectValueOf[authModeModel]       `tfsdk:"default_subscribe_auth_mode"`
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
	AppIDClientRegex types.String `tfsdk:"app_id_client_regex"`
	AWSRegion        types.String `tfsdk:"aws_region"`
	UserPoolID       types.String `tfsdk:"user_pool_id"`
}

type lambdaAuthorizerConfigModel struct {
	AuthorizerResultTTLInSeconds types.Int64  `tfsdk:"authorizer_result_ttl_in_seconds"`
	AuthorizerURI                types.String `tfsdk:"authorizer_uri"`
	IdentityValidationExpression types.String `tfsdk:"identity_validation_expression"`
}

type openIDConnectConfigModel struct {
	AuthTTL  types.Int64  `tfsdk:"auth_ttl"`
	ClientID types.String `tfsdk:"client_id"`
	IatTTL   types.Int64  `tfsdk:"iat_ttl"`
	Issuer   types.String `tfsdk:"issuer"`
}

type eventLogConfigModel struct {
	CloudWatchLogsRoleArn fwtypes.ARN                                `tfsdk:"cloudwatch_logs_role_arn"`
	LogLevel              fwtypes.StringEnum[awstypes.EventLogLevel] `tfsdk:"log_level"`
}
