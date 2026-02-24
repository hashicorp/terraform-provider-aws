// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

package bedrockagentcore

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/YakDriver/smarterr"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/bedrockagentcorecontrol"
	awstypes "github.com/aws/aws-sdk-go-v2/service/bedrockagentcorecontrol/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	sdkid "github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
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

// @FrameworkResource("aws_bedrockagentcore_gateway", name="Gateway")
// @Tags(identifierAttribute="gateway_arn")
// @Testing(tagsTest=false)
func newGatewayResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &gatewayResource{}

	r.SetDefaultCreateTimeout(30 * time.Minute)
	r.SetDefaultUpdateTimeout(30 * time.Minute)
	r.SetDefaultDeleteTimeout(30 * time.Minute)

	return r, nil
}

type gatewayResource struct {
	framework.ResourceWithModel[gatewayResourceModel]
	framework.WithTimeouts
}

func (r *gatewayResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"authorizer_type": schema.StringAttribute{
				Required:   true,
				CustomType: fwtypes.StringEnumType[awstypes.AuthorizerType](),
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			names.AttrDescription: schema.StringAttribute{
				Optional: true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 200),
				},
			},
			"exception_level": schema.StringAttribute{
				Optional:   true,
				CustomType: fwtypes.StringEnumType[awstypes.ExceptionLevel](),
			},
			"gateway_arn": framework.ARNAttributeComputedOnly(),
			"gateway_id":  framework.IDAttribute(),
			"gateway_url": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			names.AttrKMSKeyARN: schema.StringAttribute{
				CustomType: fwtypes.ARNType,
				Optional:   true,
			},
			names.AttrName: schema.StringAttribute{
				Required: true,
				Validators: []validator.String{
					stringvalidator.RegexMatches(regexache.MustCompile(`^([0-9a-zA-Z][-]?){1,100}$`), ""),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"protocol_type": schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.GatewayProtocolType](),
				Required:   true,
			},
			names.AttrRoleARN: schema.StringAttribute{
				CustomType: fwtypes.ARNType,
				Required:   true,
			},
			names.AttrTags:              tftags.TagsAttribute(),
			names.AttrTagsAll:           tftags.TagsAttributeComputedOnly(),
			"workload_identity_details": framework.ResourceComputedListOfObjectsAttribute[workloadIdentityDetailsModel](ctx, listplanmodifier.UseStateForUnknown()),
		},
		Blocks: map[string]schema.Block{
			"authorizer_configuration": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[authorizerConfigurationModel](ctx),
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
				},
				NestedObject: schema.NestedBlockObject{
					Blocks: map[string]schema.Block{
						"custom_jwt_authorizer": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[customJWTAuthorizerConfigurationModel](ctx),
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"allowed_audience": schema.SetAttribute{
										CustomType: fwtypes.SetOfStringType,
										Optional:   true,
									},
									"allowed_clients": schema.SetAttribute{
										CustomType: fwtypes.SetOfStringType,
										Optional:   true,
									},
									"discovery_url": schema.StringAttribute{
										Required: true,
									},
								},
							},
						},
					},
				},
			},
			"interceptor_configuration": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[gatewayInterceptorConfigurationModel](ctx),
				Validators: []validator.List{
					listvalidator.SizeBetween(1, 2),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"interception_points": schema.SetAttribute{
							CustomType: fwtypes.SetOfStringEnumType[awstypes.GatewayInterceptionPoint](),
							Required:   true,
						},
					},
					Blocks: map[string]schema.Block{
						"input_configuration": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[interceptorInputConfigurationModel](ctx),
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"pass_request_headers": schema.BoolAttribute{
										Required: true,
									},
								},
							},
						},
						"interceptor": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[interceptorConfigurationModel](ctx),
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
							},
							NestedObject: schema.NestedBlockObject{
								Blocks: map[string]schema.Block{
									"lambda": schema.ListNestedBlock{
										CustomType: fwtypes.NewListNestedObjectTypeOf[lambdaInterceptorConfigurationModel](ctx),
										Validators: []validator.List{
											listvalidator.SizeAtMost(1),
										},
										NestedObject: schema.NestedBlockObject{
											Attributes: map[string]schema.Attribute{
												names.AttrARN: schema.StringAttribute{
													CustomType: fwtypes.ARNType,
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
			"protocol_configuration": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[gatewayProtocolConfigurationModel](ctx),
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
				},
				NestedObject: schema.NestedBlockObject{
					Blocks: map[string]schema.Block{
						"mcp": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[mcpGatewayConfigurationModel](ctx),
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"instructions": schema.StringAttribute{
										Optional: true,
										Validators: []validator.String{
											stringvalidator.LengthBetween(1, 2048),
										},
									},
									"search_type": schema.StringAttribute{
										CustomType: fwtypes.StringEnumType[awstypes.SearchType](),
										Optional:   true,
									},
									"supported_versions": schema.SetAttribute{
										CustomType: fwtypes.SetOfStringType,
										Optional:   true,
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

func (r *gatewayResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var data gatewayResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Config.Get(ctx, &data))
	if resp.Diagnostics.HasError() {
		return
	}

	if data.AuthorizerType.ValueEnum() == awstypes.AuthorizerTypeCustomJwt {
		if data.AuthorizerConfiguration.IsNull() {
			resp.Diagnostics.AddAttributeError(
				path.Root("authorizer_configuration"),
				"Missing Required Attribute",
				"authorizer_configuration is required when authorizer_type is CUSTOM_JWT",
			)
		}
	}
}

func (r *gatewayResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data gatewayResourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.Plan.Get(ctx, &data))
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().BedrockAgentCoreClient(ctx)

	var input bedrockagentcorecontrol.CreateGatewayInput
	smerr.AddEnrich(ctx, &response.Diagnostics, fwflex.Expand(ctx, data, &input))
	if response.Diagnostics.HasError() {
		return
	}

	// Additional fields.
	input.ClientToken = aws.String(sdkid.UniqueId())
	input.Tags = getTagsIn(ctx)

	out, err := conn.CreateGateway(ctx, &input)
	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, data.Name.String())
		return
	}

	gatewayID := aws.ToString(out.GatewayId)

	if _, err := waitGatewayCreated(ctx, conn, gatewayID, r.CreateTimeout(ctx, data.Timeouts)); err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, gatewayID)
		return
	}

	gateway, err := findGatewayByID(ctx, conn, gatewayID)
	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, gatewayID)
		return
	}

	// Set values for unknowns.
	smerr.AddEnrich(ctx, &response.Diagnostics, fwflex.Flatten(ctx, gateway, &data))
	if response.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, response.State.Set(ctx, data))
}

func (r *gatewayResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data gatewayResourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.State.Get(ctx, &data))
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().BedrockAgentCoreClient(ctx)

	gatewayID := fwflex.StringValueFromFramework(ctx, data.GatewayID)
	out, err := findGatewayByID(ctx, conn, gatewayID)
	if retry.NotFound(err) {
		smerr.AddOne(ctx, &response.Diagnostics, fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, gatewayID)
		return
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, fwflex.Flatten(ctx, out, &data))
	if response.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, response.State.Set(ctx, &data))
}

func (r *gatewayResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var new, old gatewayResourceModel
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
		gatewayID := fwflex.StringValueFromFramework(ctx, new.GatewayID)
		var input bedrockagentcorecontrol.UpdateGatewayInput
		smerr.AddEnrich(ctx, &response.Diagnostics, fwflex.Expand(ctx, new, &input))
		if response.Diagnostics.HasError() {
			return
		}

		// Additional fields.
		input.GatewayIdentifier = aws.String(gatewayID)

		_, err := conn.UpdateGateway(ctx, &input)
		if err != nil {
			smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, gatewayID)
			return
		}

		if _, err := waitGatewayUpdated(ctx, conn, gatewayID, r.UpdateTimeout(ctx, new.Timeouts)); err != nil {
			smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, gatewayID)
			return
		}
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, response.State.Set(ctx, &new))
}

func (r *gatewayResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data gatewayResourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.State.Get(ctx, &data))
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().BedrockAgentCoreClient(ctx)

	gatewayID := fwflex.StringValueFromFramework(ctx, data.GatewayID)
	input := bedrockagentcorecontrol.DeleteGatewayInput{
		GatewayIdentifier: aws.String(gatewayID),
	}
	_, err := conn.DeleteGateway(ctx, &input)
	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return
	}
	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, gatewayID)
		return
	}

	if _, err := waitGatewayDeleted(ctx, conn, gatewayID, r.DeleteTimeout(ctx, data.Timeouts)); err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, gatewayID)
		return
	}
}

func (r *gatewayResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("gateway_id"), request, response)
}

func waitGatewayCreated(ctx context.Context, conn *bedrockagentcorecontrol.Client, id string, timeout time.Duration) (*bedrockagentcorecontrol.GetGatewayOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   enum.Slice(awstypes.GatewayStatusCreating),
		Target:                    enum.Slice(awstypes.GatewayStatusReady),
		Refresh:                   statusGateway(conn, id),
		Timeout:                   timeout,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*bedrockagentcorecontrol.GetGatewayOutput); ok {
		retry.SetLastError(err, errors.New(strings.Join(out.StatusReasons, "; ")))
		return out, smarterr.NewError(err)
	}

	return nil, smarterr.NewError(err)
}

func waitGatewayUpdated(ctx context.Context, conn *bedrockagentcorecontrol.Client, id string, timeout time.Duration) (*bedrockagentcorecontrol.GetGatewayOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   enum.Slice(awstypes.GatewayStatusUpdating),
		Target:                    enum.Slice(awstypes.GatewayStatusReady),
		Refresh:                   statusGateway(conn, id),
		Timeout:                   timeout,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*bedrockagentcorecontrol.GetGatewayOutput); ok {
		retry.SetLastError(err, errors.New(strings.Join(out.StatusReasons, "; ")))
		return out, smarterr.NewError(err)
	}

	return nil, smarterr.NewError(err)
}

func waitGatewayDeleted(ctx context.Context, conn *bedrockagentcorecontrol.Client, id string, timeout time.Duration) (*bedrockagentcorecontrol.GetGatewayOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.GatewayStatusDeleting, awstypes.GatewayStatusReady),
		Target:  []string{},
		Refresh: statusGateway(conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*bedrockagentcorecontrol.GetGatewayOutput); ok {
		retry.SetLastError(err, errors.New(strings.Join(out.StatusReasons, "; ")))
		return out, smarterr.NewError(err)
	}

	return nil, smarterr.NewError(err)
}

func statusGateway(conn *bedrockagentcorecontrol.Client, id string) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		out, err := findGatewayByID(ctx, conn, id)
		if retry.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", smarterr.NewError(err)
		}

		return out, string(out.Status), nil
	}
}

func findGatewayByID(ctx context.Context, conn *bedrockagentcorecontrol.Client, id string) (*bedrockagentcorecontrol.GetGatewayOutput, error) {
	input := bedrockagentcorecontrol.GetGatewayInput{
		GatewayIdentifier: aws.String(id),
	}

	return findGateway(ctx, conn, &input)
}

func findGateway(ctx context.Context, conn *bedrockagentcorecontrol.Client, input *bedrockagentcorecontrol.GetGatewayInput) (*bedrockagentcorecontrol.GetGatewayOutput, error) {
	out, err := conn.GetGateway(ctx, input)

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

type gatewayResourceModel struct {
	framework.WithRegionModel
	AuthorizerConfiguration   fwtypes.ListNestedObjectValueOf[authorizerConfigurationModel]         `tfsdk:"authorizer_configuration"`
	AuthorizerType            fwtypes.StringEnum[awstypes.AuthorizerType]                           `tfsdk:"authorizer_type"`
	Description               types.String                                                          `tfsdk:"description"`
	ExceptionLevel            fwtypes.StringEnum[awstypes.ExceptionLevel]                           `tfsdk:"exception_level"`
	GatewayARN                types.String                                                          `tfsdk:"gateway_arn"`
	GatewayID                 types.String                                                          `tfsdk:"gateway_id"`
	GatewayURL                types.String                                                          `tfsdk:"gateway_url"`
	InterceptorConfigurations fwtypes.ListNestedObjectValueOf[gatewayInterceptorConfigurationModel] `tfsdk:"interceptor_configuration"`
	KMSKeyARN                 fwtypes.ARN                                                           `tfsdk:"kms_key_arn"`
	Name                      types.String                                                          `tfsdk:"name"`
	ProtocolConfiguration     fwtypes.ListNestedObjectValueOf[gatewayProtocolConfigurationModel]    `tfsdk:"protocol_configuration"`
	ProtocolType              fwtypes.StringEnum[awstypes.GatewayProtocolType]                      `tfsdk:"protocol_type"`
	RoleARN                   fwtypes.ARN                                                           `tfsdk:"role_arn"`
	Tags                      tftags.Map                                                            `tfsdk:"tags"`
	TagsAll                   tftags.Map                                                            `tfsdk:"tags_all"`
	Timeouts                  timeouts.Value                                                        `tfsdk:"timeouts"`
	WorkloadIdentityDetails   fwtypes.ListNestedObjectValueOf[workloadIdentityDetailsModel]         `tfsdk:"workload_identity_details"`
}

type gatewayProtocolConfigurationModel struct {
	MCP fwtypes.ListNestedObjectValueOf[mcpGatewayConfigurationModel] `tfsdk:"mcp"`
}

var (
	_ fwflex.Expander  = gatewayProtocolConfigurationModel{}
	_ fwflex.Flattener = &gatewayProtocolConfigurationModel{}
)

func (m *gatewayProtocolConfigurationModel) Flatten(ctx context.Context, v any) diag.Diagnostics {
	var diags diag.Diagnostics
	switch t := v.(type) {
	case awstypes.GatewayProtocolConfigurationMemberMcp:
		var data mcpGatewayConfigurationModel
		smerr.AddEnrich(ctx, &diags, fwflex.Flatten(ctx, t.Value, &data))
		if diags.HasError() {
			return diags
		}
		m.MCP = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &data)

	default:
		diags.AddError(
			"Unsupported Type",
			fmt.Sprintf("gateway protocol configuration flatten: %T", v),
		)
	}
	return diags
}

func (m gatewayProtocolConfigurationModel) Expand(ctx context.Context) (any, diag.Diagnostics) {
	var diags diag.Diagnostics
	switch {
	case !m.MCP.IsNull():
		data, d := m.MCP.ToPtr(ctx)
		smerr.AddEnrich(ctx, &diags, d)
		if diags.HasError() {
			return nil, diags
		}
		var r awstypes.GatewayProtocolConfigurationMemberMcp
		smerr.AddEnrich(ctx, &diags, fwflex.Expand(ctx, data, &r.Value))
		if diags.HasError() {
			return nil, diags
		}
		return &r, diags
	}
	return nil, diags
}

type mcpGatewayConfigurationModel struct {
	Instructions      types.String                            `tfsdk:"instructions"`
	SearchType        fwtypes.StringEnum[awstypes.SearchType] `tfsdk:"search_type"`
	SupportedVersions fwtypes.SetOfString                     `tfsdk:"supported_versions"`
}

type gatewayInterceptorConfigurationModel struct {
	InputConfiguration fwtypes.ListNestedObjectValueOf[interceptorInputConfigurationModel] `tfsdk:"input_configuration"`
	InterceptionPoints fwtypes.SetOfStringEnum[awstypes.GatewayInterceptionPoint]          `tfsdk:"interception_points"`
	Interceptor        fwtypes.ListNestedObjectValueOf[interceptorConfigurationModel]      `tfsdk:"interceptor"`
}

type interceptorInputConfigurationModel struct {
	PassRequestHeaders types.Bool `tfsdk:"pass_request_headers"`
}

type interceptorConfigurationModel struct {
	Lambda fwtypes.ListNestedObjectValueOf[lambdaInterceptorConfigurationModel] `tfsdk:"lambda"`
}

var (
	_ fwflex.Expander  = interceptorConfigurationModel{}
	_ fwflex.Flattener = &interceptorConfigurationModel{}
)

func (m *interceptorConfigurationModel) Flatten(ctx context.Context, v any) diag.Diagnostics {
	var diags diag.Diagnostics
	switch t := v.(type) {
	case awstypes.InterceptorConfigurationMemberLambda:
		var data lambdaInterceptorConfigurationModel
		smerr.AddEnrich(ctx, &diags, fwflex.Flatten(ctx, t.Value, &data))
		if diags.HasError() {
			return diags
		}
		m.Lambda = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &data)

	default:
		diags.AddError(
			"Unsupported Type",
			fmt.Sprintf("interceptor configuration flatten: %T", v),
		)
	}
	return diags
}

func (m interceptorConfigurationModel) Expand(ctx context.Context) (any, diag.Diagnostics) {
	var diags diag.Diagnostics
	switch {
	case !m.Lambda.IsNull():
		data, d := m.Lambda.ToPtr(ctx)
		smerr.AddEnrich(ctx, &diags, d)
		if diags.HasError() {
			return nil, diags
		}
		var r awstypes.InterceptorConfigurationMemberLambda
		smerr.AddEnrich(ctx, &diags, fwflex.Expand(ctx, data, &r.Value))
		if diags.HasError() {
			return nil, diags
		}
		return &r, diags
	}
	return nil, diags
}

type lambdaInterceptorConfigurationModel struct {
	ARN fwtypes.ARN `tfsdk:"arn"`
}
