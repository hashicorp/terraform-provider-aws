// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package elbv2

import (
	"context"
	"fmt"
	"strconv"

	"github.com/aws/aws-sdk-go-v2/aws"
	awstypes "github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2/types"
	"github.com/hashicorp/terraform-plugin-framework-validators/datasourcevalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkDataSource("aws_lb_listener_rule", name="Listener Rule")
// @Tags(identifierAttribute="arn")
func newDataSourceListenerRule(context.Context) (datasource.DataSourceWithConfigure, error) {
	return &dataSourceListenerRule{}, nil
}

const (
	dsNameListenerRule = "Listener Rule Data Source"
)

type dataSourceListenerRule struct {
	framework.DataSourceWithConfigure
}

func (d *dataSourceListenerRule) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrARN: schema.StringAttribute{
				CustomType: fwtypes.ARNType,
				Optional:   true,
				Computed:   true,
			},
			"listener_arn": schema.StringAttribute{
				CustomType: fwtypes.ARNType,
				Optional:   true,
				Computed:   true,
			},
			names.AttrPriority: schema.Int32Attribute{
				Optional: true,
				Computed: true,
			},
			names.AttrTags: tftags.TagsAttributeComputedOnly(),
		},
		Blocks: map[string]schema.Block{
			names.AttrAction: schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[actionModel](ctx),
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"order": schema.Int32Attribute{
							Computed: true,
						},
						names.AttrType: schema.StringAttribute{
							Computed: true,
						},
					},
					Blocks: map[string]schema.Block{
						"authenticate_cognito": schema.SingleNestedBlock{ // nosemgrep:ci.avoid-SingleNestedBlock pre-existing, will be converted
							CustomType: fwtypes.NewObjectTypeOf[authenticateCognitoActionConfigModel](ctx),
							Attributes: map[string]schema.Attribute{
								"authentication_request_extra_params": schema.MapAttribute{
									ElementType: types.StringType,
									Computed:    true,
								},
								"on_unauthenticated_request": schema.StringAttribute{
									Computed: true,
								},
								names.AttrScope: schema.StringAttribute{
									Computed: true,
								},
								"session_cookie_name": schema.StringAttribute{
									Computed: true,
								},
								"session_timeout": schema.Int64Attribute{
									Computed: true,
								},
								"user_pool_arn": schema.StringAttribute{
									Computed: true,
								},
								"user_pool_client_id": schema.StringAttribute{
									Computed: true,
								},
								"user_pool_domain": schema.StringAttribute{
									Computed: true,
								},
							},
						},
						"authenticate_oidc": schema.SingleNestedBlock{ // nosemgrep:ci.avoid-SingleNestedBlock pre-existing, will be converted
							Attributes: map[string]schema.Attribute{
								"authentication_request_extra_params": schema.MapAttribute{
									ElementType: types.StringType,
									Computed:    true,
								},
								"authorization_endpoint": schema.StringAttribute{
									Computed: true,
								},
								names.AttrClientID: schema.StringAttribute{
									Computed: true,
								},
								names.AttrIssuer: schema.StringAttribute{
									Computed: true,
								},
								"on_unauthenticated_request": schema.StringAttribute{
									Computed: true,
								},
								names.AttrScope: schema.StringAttribute{
									Computed: true,
								},
								"session_cookie_name": schema.StringAttribute{
									Computed: true,
								},
								"session_timeout": schema.Int64Attribute{
									Computed: true,
								},
								"token_endpoint": schema.StringAttribute{
									Computed: true,
								},
								"user_info_endpoint": schema.StringAttribute{
									Computed: true,
								},
							},
						},
						"fixed_response": schema.SingleNestedBlock{ // nosemgrep:ci.avoid-SingleNestedBlock pre-existing, will be converted
							Attributes: map[string]schema.Attribute{
								names.AttrContentType: schema.StringAttribute{
									Computed: true,
								},
								"message_body": schema.StringAttribute{
									Computed: true,
								},
								names.AttrStatusCode: schema.StringAttribute{
									Computed: true,
								},
							},
						},
						"forward": schema.SingleNestedBlock{ // nosemgrep:ci.avoid-SingleNestedBlock pre-existing, will be converted
							Blocks: map[string]schema.Block{
								"stickiness": schema.SingleNestedBlock{ // nosemgrep:ci.avoid-SingleNestedBlock pre-existing, will be converted
									Attributes: map[string]schema.Attribute{
										names.AttrDuration: schema.Int32Attribute{
											Computed: true,
										},
										names.AttrEnabled: schema.BoolAttribute{
											Computed: true,
										},
									},
								},
								"target_group": schema.SetNestedBlock{
									NestedObject: schema.NestedBlockObject{
										Attributes: map[string]schema.Attribute{
											names.AttrARN: schema.StringAttribute{
												Computed: true,
											},
											names.AttrWeight: schema.Int32Attribute{
												Computed: true,
											},
										},
									},
								},
							},
						},
						"redirect": schema.SingleNestedBlock{ // nosemgrep:ci.avoid-SingleNestedBlock pre-existing, will be converted
							Attributes: map[string]schema.Attribute{
								"host": schema.StringAttribute{
									Computed: true,
								},
								names.AttrPath: schema.StringAttribute{
									Computed: true,
								},
								names.AttrPort: schema.StringAttribute{
									Computed: true,
								},
								names.AttrProtocol: schema.StringAttribute{
									Computed: true,
								},
								"query": schema.StringAttribute{
									Computed: true,
								},
								names.AttrStatusCode: schema.StringAttribute{
									Computed: true,
								},
							},
						},
					},
				},
			},
			names.AttrCondition: schema.SetNestedBlock{
				CustomType: fwtypes.NewSetNestedObjectTypeOf[ruleConditionModel](ctx),
				NestedObject: schema.NestedBlockObject{
					Blocks: map[string]schema.Block{
						"host_header": schema.SingleNestedBlock{ // nosemgrep:ci.avoid-SingleNestedBlock pre-existing, will be converted
							CustomType: fwtypes.NewObjectTypeOf[hostHeaderConfigModel](ctx),
							Attributes: map[string]schema.Attribute{
								names.AttrValues: schema.SetAttribute{
									ElementType: types.StringType,
									Computed:    true,
								},
							},
						},
						"http_header": schema.SingleNestedBlock{ // nosemgrep:ci.avoid-SingleNestedBlock pre-existing, will be converted
							CustomType: fwtypes.NewObjectTypeOf[httpHeaderConfigModel](ctx),
							Attributes: map[string]schema.Attribute{
								"http_header_name": schema.StringAttribute{
									Computed: true,
								},
								names.AttrValues: schema.SetAttribute{
									ElementType: types.StringType,
									Computed:    true,
								},
							},
						},
						"http_request_method": schema.SingleNestedBlock{ // nosemgrep:ci.avoid-SingleNestedBlock pre-existing, will be converted
							CustomType: fwtypes.NewObjectTypeOf[httpRquestMethodConfigModel](ctx),
							Attributes: map[string]schema.Attribute{
								names.AttrValues: schema.SetAttribute{
									ElementType: types.StringType,
									Computed:    true,
								},
							},
						},
						"path_pattern": schema.SingleNestedBlock{ // nosemgrep:ci.avoid-SingleNestedBlock pre-existing, will be converted
							CustomType: fwtypes.NewObjectTypeOf[pathPatternConfigModel](ctx),
							Attributes: map[string]schema.Attribute{
								names.AttrValues: schema.SetAttribute{
									ElementType: types.StringType,
									Computed:    true,
								},
							},
						},
						"query_string": schema.SingleNestedBlock{ // nosemgrep:ci.avoid-SingleNestedBlock pre-existing, will be converted
							CustomType: fwtypes.NewObjectTypeOf[queryStringConfigModel](ctx),
							Blocks: map[string]schema.Block{
								names.AttrValues: schema.SetNestedBlock{
									CustomType: fwtypes.NewSetNestedObjectTypeOf[queryStringKeyValuePairModel](ctx),
									NestedObject: schema.NestedBlockObject{
										Attributes: map[string]schema.Attribute{
											names.AttrKey: schema.StringAttribute{
												Computed: true,
											},
											names.AttrValue: schema.StringAttribute{
												Computed: true,
											},
										},
									},
								},
							},
						},
						"source_ip": schema.SingleNestedBlock{ // nosemgrep:ci.avoid-SingleNestedBlock pre-existing, will be converted
							CustomType: fwtypes.NewObjectTypeOf[pathPatternConfigModel](ctx),
							Attributes: map[string]schema.Attribute{
								names.AttrValues: schema.SetAttribute{
									ElementType: types.StringType,
									Computed:    true,
								},
							},
						},
					},
				},
			},
		},
	}
}

func (d *dataSourceListenerRule) ConfigValidators(_ context.Context) []datasource.ConfigValidator {
	return []datasource.ConfigValidator{
		datasourcevalidator.ExactlyOneOf(
			path.MatchRoot(names.AttrARN),
			path.MatchRoot("listener_arn"),
		),
		datasourcevalidator.RequiredTogether(
			path.MatchRoot("listener_arn"),
			path.MatchRoot(names.AttrPriority),
		),
		datasourcevalidator.Conflicting(
			path.MatchRoot(names.AttrARN),
			path.MatchRoot(names.AttrPriority),
		),
	}
}

func (d *dataSourceListenerRule) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	conn := d.Meta().ELBV2Client(ctx)

	var data dataSourceListenerRuleModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var out *awstypes.Rule
	if !data.ARN.IsNull() {
		var err error
		out, err = findListenerRuleByARN(ctx, conn, data.ARN.ValueString())
		if err != nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.ELBV2, create.ErrActionReading, dsNameListenerRule, data.ARN.String(), err),
				err.Error(),
			)
			return
		}
	} else {
		var err error
		out, err = findListenerRuleByListenerAndPriority(ctx, conn, data.ListenerARN.ValueString(), strconv.Itoa(int(data.Priority.ValueInt32())))
		if err != nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.ELBV2, create.ErrActionReading, dsNameListenerRule, fmt.Sprintf("%s/%s", data.ListenerARN.String(), data.Priority.String()), err),
				err.Error(),
			)
			return
		}
	}

	sortListenerActions(out.Actions)

	resp.Diagnostics.Append(flex.Flatten(ctx, out, &data, flex.WithFieldNamePrefix("Rule"))...)
	if resp.Diagnostics.HasError() {
		return
	}

	// The listener arn isn't in the response but can be derived from the rule arn
	data.ListenerARN = fwtypes.ARNValue(listenerARNFromRuleARN(aws.ToString(out.RuleArn)))

	priority, err := strconv.ParseInt(aws.ToString(out.Priority), 10, 32)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.ELBV2, create.ErrActionReading, dsNameListenerRule, data.ARN.String(), err),
			err.Error(),
		)
		return
	}
	data.Priority = types.Int32Value(int32(priority))

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

type dataSourceListenerRuleModel struct {
	Action      fwtypes.ListNestedObjectValueOf[actionModel]       `tfsdk:"action"`
	ARN         fwtypes.ARN                                        `tfsdk:"arn"`
	Condition   fwtypes.SetNestedObjectValueOf[ruleConditionModel] `tfsdk:"condition"`
	ListenerARN fwtypes.ARN                                        `tfsdk:"listener_arn"`
	Priority    types.Int32                                        `tfsdk:"priority" autoflex:"-"`
	Tags        tftags.Map                                         `tfsdk:"tags"`
}

// The API includes a TargetGroupArn field at the root level of the Action. This only applies when Type == "forward"
// and there is a single target group. The read also populates the ForwardConfig with the single TargetGroup,
// so it is redundant. Using the ForwardConfig in all cases improves consistency.
type actionModel struct {
	Type                      types.String                                                `tfsdk:"type"`
	AuthenticateCognitoConfig fwtypes.ObjectValueOf[authenticateCognitoActionConfigModel] `tfsdk:"authenticate_cognito"`
	AuthenticateOidcConfig    fwtypes.ObjectValueOf[authenticateOIDCActionConfigModel]    `tfsdk:"authenticate_oidc"`
	FixedResponseConfig       fwtypes.ObjectValueOf[fixedResponseActionConfigModel]       `tfsdk:"fixed_response"`
	ForwardConfig             fwtypes.ObjectValueOf[forwardActionConfigModel]             `tfsdk:"forward"`
	Order                     types.Int32                                                 `tfsdk:"order"`
	RedirectConfig            fwtypes.ObjectValueOf[redirectActionConfigModel]            `tfsdk:"redirect"`
}

type authenticateCognitoActionConfigModel struct {
	AuthenticationRequestExtraParams fwtypes.MapOfString `tfsdk:"authentication_request_extra_params"`
	OnUnauthenticatedRequest         types.String        `tfsdk:"on_unauthenticated_request"`
	Scope                            types.String        `tfsdk:"scope"`
	SessionCookieName                types.String        `tfsdk:"session_cookie_name"`
	SessionTimeout                   types.Int64         `tfsdk:"session_timeout"`
	UserPoolArn                      types.String        `tfsdk:"user_pool_arn"`
	UserPoolClientId                 types.String        `tfsdk:"user_pool_client_id"`
	UserPoolDomain                   types.String        `tfsdk:"user_pool_domain"`
}

type authenticateOIDCActionConfigModel struct {
	AuthorizationEndpoint            types.String        `tfsdk:"authorization_endpoint"`
	AuthenticationRequestExtraParams fwtypes.MapOfString `tfsdk:"authentication_request_extra_params"`
	ClientId                         types.String        `tfsdk:"client_id"`
	Issuer                           types.String        `tfsdk:"issuer"`
	OnUnauthenticatedRequest         types.String        `tfsdk:"on_unauthenticated_request"`
	Scope                            types.String        `tfsdk:"scope"`
	SessionCookieName                types.String        `tfsdk:"session_cookie_name"`
	SessionTimeout                   types.Int64         `tfsdk:"session_timeout"`
	TokenEndpoint                    types.String        `tfsdk:"token_endpoint"`
	UserInfoEndpoint                 types.String        `tfsdk:"user_info_endpoint"`
}

type fixedResponseActionConfigModel struct {
	ContentType types.String `tfsdk:"content_type"`
	MessageBody types.String `tfsdk:"message_body"`
	StatusCode  types.String `tfsdk:"status_code"`
}

type forwardActionConfigModel struct {
	TargetGroupStickinessConfig fwtypes.ObjectValueOf[targetGroupStickinessConfigModel] `tfsdk:"stickiness"`
	TargetGroups                fwtypes.SetNestedObjectValueOf[targetGroupTupleModel]   `tfsdk:"target_group"`
}

type targetGroupStickinessConfigModel struct {
	DurationSeconds types.Int32 `tfsdk:"duration"`
	Enabled         types.Bool  `tfsdk:"enabled"`
}

type targetGroupTupleModel struct {
	TargetGroupArn types.String `tfsdk:"arn"`
	Weight         types.Int32  `tfsdk:"weight"`
}

type redirectActionConfigModel struct {
	Host       types.String `tfsdk:"host"`
	Path       types.String `tfsdk:"path"`
	Port       types.String `tfsdk:"port"`
	Protocol   types.String `tfsdk:"protocol"`
	Query      types.String `tfsdk:"query"`
	StatusCode types.String `tfsdk:"status_code"`
}

type ruleConditionModel struct {
	HostHeaderConfig        fwtypes.ObjectValueOf[hostHeaderConfigModel]       `tfsdk:"host_header"`
	HttpHeaderConfig        fwtypes.ObjectValueOf[httpHeaderConfigModel]       `tfsdk:"http_header"`
	HttpRequestMethodConfig fwtypes.ObjectValueOf[httpRquestMethodConfigModel] `tfsdk:"http_request_method"`
	PathPatternConfig       fwtypes.ObjectValueOf[pathPatternConfigModel]      `tfsdk:"path_pattern"`
	QueryStringConfig       fwtypes.ObjectValueOf[queryStringConfigModel]      `tfsdk:"query_string"`
	SourceIpConfig          fwtypes.ObjectValueOf[sourceIPConfigModel]         `tfsdk:"source_ip"`
}

type hostHeaderConfigModel struct {
	Values fwtypes.SetValueOf[types.String] `tfsdk:"values"`
}

type httpHeaderConfigModel struct {
	HTTPHeaderName types.String                     `tfsdk:"http_header_name"`
	Values         fwtypes.SetValueOf[types.String] `tfsdk:"values"`
}

type httpRquestMethodConfigModel struct {
	Values fwtypes.SetValueOf[types.String] `tfsdk:"values"`
}

type pathPatternConfigModel struct {
	Values fwtypes.SetValueOf[types.String] `tfsdk:"values"`
}

type queryStringConfigModel struct {
	Values fwtypes.SetNestedObjectValueOf[queryStringKeyValuePairModel] `tfsdk:"values"`
}

type queryStringKeyValuePairModel struct {
	Key   types.String `tfsdk:"key"`
	Value types.String `tfsdk:"value"`
}

type sourceIPConfigModel struct {
	Values fwtypes.SetValueOf[types.String] `tfsdk:"values"`
}
