// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package bedrockagentcore

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"time"

	"github.com/YakDriver/smarterr"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/bedrockagentcorecontrol"
	awstypes "github.com/aws/aws-sdk-go-v2/service/bedrockagentcorecontrol/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
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

// Function annotations are used for resource registration to the Provider. DO NOT EDIT.
// @FrameworkResource("aws_bedrockagentcore_gateway", name="Gateway")
// @Tags(identifierAttribute="gateway_arn")
// @Testing(tagsTest=false)
func newResourceGateway(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &resourceGateway{}

	r.SetDefaultCreateTimeout(30 * time.Minute)
	r.SetDefaultUpdateTimeout(30 * time.Minute)
	r.SetDefaultDeleteTimeout(30 * time.Minute)

	return r, nil
}

const (
	ResNameGateway = "Gateway"
)

type resourceGateway struct {
	framework.ResourceWithModel[resourceGatewayModel]
	framework.WithTimeouts
}

func (r *resourceGateway) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	authorizerType := fwtypes.StringEnumType[awstypes.AuthorizerType]()
	protocolType := fwtypes.StringEnumType[awstypes.GatewayProtocolType]()
	exceptionLevel := fwtypes.StringEnumType[awstypes.ExceptionLevel]()
	searchType := fwtypes.StringEnumType[awstypes.SearchType]()

	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"gateway_arn": schema.StringAttribute{
				CustomType: fwtypes.ARNType,
				Computed:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"authorizer_type": schema.StringAttribute{
				Computed:   true,
				Optional:   true,
				CustomType: authorizerType,
				Default:    authorizerType.AttributeDefault(awstypes.AuthorizerTypeCustomJwt),
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"client_token": schema.StringAttribute{
				Optional: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			names.AttrDescription: schema.StringAttribute{
				Optional: true,
			},
			"exception_level": schema.StringAttribute{
				Optional:   true,
				CustomType: exceptionLevel,
			},
			"gateway_url": schema.StringAttribute{
				Computed: true,
			},
			"gateway_id": framework.IDAttribute(),
			names.AttrKMSKeyARN: schema.StringAttribute{
				CustomType: fwtypes.ARNType,
				Optional:   true,
			},
			names.AttrName: schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"protocol_type": schema.StringAttribute{
				Computed:   true,
				Optional:   true,
				CustomType: protocolType,
				Default:    protocolType.AttributeDefault(awstypes.GatewayProtocolTypeMcp),
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			names.AttrRoleARN: schema.StringAttribute{
				CustomType: fwtypes.ARNType,
				Required:   true,
			},
			"workload_identity_details": schema.ListAttribute{
				CustomType: fwtypes.NewListNestedObjectTypeOf[gatewayWorkloadIdentityDetailsModel](ctx),
				Computed:   true,
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
				},
			},
			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
		},
		Blocks: map[string]schema.Block{
			"authorizer_configuration": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[gatewayAuthorizerConfigurationModel](ctx),
				Validators: []validator.List{
					listvalidator.IsRequired(),
					listvalidator.SizeAtMost(1),
				},
				NestedObject: schema.NestedBlockObject{
					Blocks: map[string]schema.Block{
						"custom_jwt_authorizer": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[gatewayCustomJWTAuthorizerConfigurationModel](ctx),
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"discovery_url": schema.StringAttribute{
										Required: true,
									},
									"allowed_audience": schema.SetAttribute{
										CustomType: fwtypes.SetOfStringType,
										Optional:   true,
									},
									"allowed_clients": schema.SetAttribute{
										CustomType: fwtypes.SetOfStringType,
										Optional:   true,
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
									},
									"search_type": schema.StringAttribute{
										Optional:   true,
										CustomType: searchType,
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

func (r *resourceGateway) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := r.Meta().BedrockAgentCoreClient(ctx)

	var plan resourceGatewayModel
	smerr.EnrichAppend(ctx, &resp.Diagnostics, req.Plan.Get(ctx, &plan))
	if resp.Diagnostics.HasError() {
		return
	}

	var input bedrockagentcorecontrol.CreateGatewayInput
	smerr.EnrichAppend(ctx, &resp.Diagnostics, flex.Expand(ctx, plan, &input))
	if resp.Diagnostics.HasError() {
		return
	}

	input.Tags = getTagsIn(ctx)

	out, err := conn.CreateGateway(ctx, &input)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, plan.Name.String())
		return
	}
	if out == nil {
		smerr.AddError(ctx, &resp.Diagnostics, errors.New("empty output"), smerr.ID, plan.Name.String())
		return
	}

	smerr.EnrichAppend(ctx, &resp.Diagnostics, flex.Flatten(ctx, out, &plan))
	if resp.Diagnostics.HasError() {
		return
	}

	createTimeout := r.CreateTimeout(ctx, plan.Timeouts)
	_, err = waitGatewayCreated(ctx, conn, plan.GatewayID.ValueString(), createTimeout)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, plan.Name.String())
		return
	}

	smerr.EnrichAppend(ctx, &resp.Diagnostics, resp.State.Set(ctx, plan))
}

func (r *resourceGateway) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().BedrockAgentCoreClient(ctx)

	var state resourceGatewayModel
	smerr.EnrichAppend(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := findGatewayByID(ctx, conn, state.GatewayID.ValueString())
	if tfresource.NotFound(err) {
		resp.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, state.GatewayID.String())
		return
	}

	smerr.EnrichAppend(ctx, &resp.Diagnostics, flex.Flatten(ctx, out, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	smerr.EnrichAppend(ctx, &resp.Diagnostics, resp.State.Set(ctx, &state))
}

func (r *resourceGateway) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	conn := r.Meta().BedrockAgentCoreClient(ctx)

	var plan, state resourceGatewayModel
	smerr.EnrichAppend(ctx, &resp.Diagnostics, req.Plan.Get(ctx, &plan))
	smerr.EnrichAppend(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	diff, d := flex.Diff(ctx, plan, state)
	smerr.EnrichAppend(ctx, &resp.Diagnostics, d)
	if resp.Diagnostics.HasError() {
		return
	}

	if diff.HasChanges() {
		var input bedrockagentcorecontrol.UpdateGatewayInput
		smerr.EnrichAppend(ctx, &resp.Diagnostics, flex.Expand(ctx, plan, &input, flex.WithFieldNameSuffix("entifier")))
		if resp.Diagnostics.HasError() {
			return
		}

		out, err := conn.UpdateGateway(ctx, &input)
		if err != nil {
			smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, plan.GatewayID.String())
			return
		}
		if out == nil {
			smerr.AddError(ctx, &resp.Diagnostics, errors.New("empty output"), smerr.ID, plan.GatewayID.String())
			return
		}

		smerr.EnrichAppend(ctx, &resp.Diagnostics, flex.Flatten(ctx, out, &plan))
		if resp.Diagnostics.HasError() {
			return
		}

		updateTimeout := r.UpdateTimeout(ctx, plan.Timeouts)
		_, err = waitGatewayUpdated(ctx, conn, plan.GatewayID.ValueString(), updateTimeout)
		if err != nil {
			smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, plan.GatewayID.String())
			return
		}
	} else {
		plan.GatewayURL = state.GatewayURL
		plan.WorkloadIdentityDetails = state.WorkloadIdentityDetails
	}

	smerr.EnrichAppend(ctx, &resp.Diagnostics, resp.State.Set(ctx, &plan))
}

func (r *resourceGateway) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := r.Meta().BedrockAgentCoreClient(ctx)

	var state resourceGatewayModel
	smerr.EnrichAppend(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	input := bedrockagentcorecontrol.DeleteGatewayInput{
		GatewayIdentifier: state.GatewayID.ValueStringPointer(),
	}

	_, err := conn.DeleteGateway(ctx, &input)
	if err != nil {
		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return
		}

		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, state.GatewayID.String())
		return
	}

	deleteTimeout := r.DeleteTimeout(ctx, state.Timeouts)
	_, err = waitGatewayDeleted(ctx, conn, state.GatewayID.ValueString(), deleteTimeout)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, state.GatewayID.String())
		return
	}
}

func (r *resourceGateway) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("gateway_id"), req, resp)
}

func waitGatewayCreated(ctx context.Context, conn *bedrockagentcorecontrol.Client, id string, timeout time.Duration) (*bedrockagentcorecontrol.GetGatewayOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   enum.Slice(awstypes.GatewayStatusCreating),
		Target:                    enum.Slice(awstypes.GatewayStatusReady),
		Refresh:                   statusGateway(ctx, conn, id),
		Timeout:                   timeout,
		NotFoundChecks:            20,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*bedrockagentcorecontrol.GetGatewayOutput); ok {
		return out, smarterr.NewError(err)
	}

	return nil, smarterr.NewError(err)
}

func waitGatewayUpdated(ctx context.Context, conn *bedrockagentcorecontrol.Client, id string, timeout time.Duration) (*bedrockagentcorecontrol.GetGatewayOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   enum.Slice(awstypes.GatewayStatusUpdating),
		Target:                    enum.Slice(awstypes.GatewayStatusReady),
		Refresh:                   statusGateway(ctx, conn, id),
		Timeout:                   timeout,
		NotFoundChecks:            20,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*bedrockagentcorecontrol.GetGatewayOutput); ok {
		return out, smarterr.NewError(err)
	}

	return nil, smarterr.NewError(err)
}

func waitGatewayDeleted(ctx context.Context, conn *bedrockagentcorecontrol.Client, id string, timeout time.Duration) (*bedrockagentcorecontrol.GetGatewayOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.GatewayStatusDeleting, awstypes.GatewayStatusReady),
		Target:  []string{},
		Refresh: statusGateway(ctx, conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*bedrockagentcorecontrol.GetGatewayOutput); ok {
		return out, smarterr.NewError(err)
	}

	return nil, smarterr.NewError(err)
}

func statusGateway(ctx context.Context, conn *bedrockagentcorecontrol.Client, id string) retry.StateRefreshFunc {
	return func() (any, string, error) {
		out, err := findGatewayByID(ctx, conn, id)
		if tfresource.NotFound(err) {
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

	out, err := conn.GetGateway(ctx, &input)
	if err != nil {
		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return nil, smarterr.NewError(&retry.NotFoundError{
				LastError:   err,
				LastRequest: &input,
			})
		}

		return nil, smarterr.NewError(err)
	}

	if out == nil {
		return nil, smarterr.NewError(tfresource.NewEmptyResultError(&input))
	}

	return out, nil
}

type resourceGatewayModel struct {
	framework.WithRegionModel

	GatewayARN     fwtypes.ARN                                      `tfsdk:"gateway_arn"`
	AuthorizerType fwtypes.StringEnum[awstypes.AuthorizerType]      `tfsdk:"authorizer_type"`
	ClientToken    types.String                                     `tfsdk:"client_token"`
	Description    types.String                                     `tfsdk:"description"`
	ExceptionLevel fwtypes.StringEnum[awstypes.ExceptionLevel]      `tfsdk:"exception_level" autoflex:",omitempty"`
	GatewayURL     types.String                                     `tfsdk:"gateway_url"`
	GatewayID      types.String                                     `tfsdk:"gateway_id"`
	KMSKeyArn      fwtypes.ARN                                      `tfsdk:"kms_key_arn"`
	Name           types.String                                     `tfsdk:"name"`
	ProtocolType   fwtypes.StringEnum[awstypes.GatewayProtocolType] `tfsdk:"protocol_type"`
	RoleArn        fwtypes.ARN                                      `tfsdk:"role_arn"`

	AuthorizerConfiguration fwtypes.ListNestedObjectValueOf[gatewayAuthorizerConfigurationModel] `tfsdk:"authorizer_configuration"`
	ProtocolConfiguration   fwtypes.ListNestedObjectValueOf[gatewayProtocolConfigurationModel]   `tfsdk:"protocol_configuration"`
	WorkloadIdentityDetails fwtypes.ListNestedObjectValueOf[gatewayWorkloadIdentityDetailsModel] `tfsdk:"workload_identity_details"`

	Tags     tftags.Map     `tfsdk:"tags"`
	TagsAll  tftags.Map     `tfsdk:"tags_all"`
	Timeouts timeouts.Value `tfsdk:"timeouts"`
}

type gatewayWorkloadIdentityDetailsModel struct {
	WorkloadIdentityArn types.String `tfsdk:"workload_identity_arn"`
}

type gatewayAuthorizerConfigurationModel struct {
	CustomJWTAuthorizer fwtypes.ListNestedObjectValueOf[gatewayCustomJWTAuthorizerConfigurationModel] `tfsdk:"custom_jwt_authorizer"`
}

var (
	_ flex.Expander  = gatewayAuthorizerConfigurationModel{}
	_ flex.Flattener = &gatewayAuthorizerConfigurationModel{}
)

func (m *gatewayAuthorizerConfigurationModel) Flatten(ctx context.Context, v any) (diags diag.Diagnostics) {
	switch t := v.(type) {
	case awstypes.AuthorizerConfigurationMemberCustomJWTAuthorizer:
		var model gatewayCustomJWTAuthorizerConfigurationModel
		smerr.EnrichAppend(ctx, &diags, flex.Flatten(ctx, t.Value, &model))
		if diags.HasError() {
			return diags
		}
		m.CustomJWTAuthorizer = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &model)

	default:
		diags.AddError(
			"Unsupported Type",
			fmt.Sprintf("authorization configuration flatten: %s", reflect.TypeOf(v).String()),
		)
	}
	return diags
}

func (m gatewayAuthorizerConfigurationModel) Expand(ctx context.Context) (result any, diags diag.Diagnostics) {
	switch {
	case !m.CustomJWTAuthorizer.IsNull():
		model, d := m.CustomJWTAuthorizer.ToPtr(ctx)
		smerr.EnrichAppend(ctx, &diags, d)
		if diags.HasError() {
			return nil, diags
		}
		var r awstypes.AuthorizerConfigurationMemberCustomJWTAuthorizer
		smerr.EnrichAppend(ctx, &diags, flex.Expand(ctx, model, &r.Value))
		if diags.HasError() {
			return nil, diags
		}
		return &r, diags
	}
	return nil, diags
}

type gatewayCustomJWTAuthorizerConfigurationModel struct {
	DiscoveryUrl    types.String        `tfsdk:"discovery_url"`
	AllowedAudience fwtypes.SetOfString `tfsdk:"allowed_audience"`
	AllowedClients  fwtypes.SetOfString `tfsdk:"allowed_clients"`
}

type gatewayProtocolConfigurationModel struct {
	MCP fwtypes.ListNestedObjectValueOf[mcpGatewayConfigurationModel] `tfsdk:"mcp"`
}

var (
	_ flex.Expander  = gatewayProtocolConfigurationModel{}
	_ flex.Flattener = &gatewayProtocolConfigurationModel{}
)

func (m *gatewayProtocolConfigurationModel) Flatten(ctx context.Context, v any) (diags diag.Diagnostics) {
	switch t := v.(type) {
	case awstypes.GatewayProtocolConfigurationMemberMcp:
		var model mcpGatewayConfigurationModel
		smerr.EnrichAppend(ctx, &diags, flex.Flatten(ctx, t.Value, &model))
		if diags.HasError() {
			return diags
		}
		m.MCP = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &model)

	default:
		diags.AddError(
			"Unsupported Type",
			fmt.Sprintf("gateway protocol configuration flatten: %s", reflect.TypeOf(v).String()),
		)
	}
	return diags
}

func (m gatewayProtocolConfigurationModel) Expand(ctx context.Context) (result any, diags diag.Diagnostics) {
	switch {
	case !m.MCP.IsNull():
		model, d := m.MCP.ToPtr(ctx)
		smerr.EnrichAppend(ctx, &diags, d)
		if diags.HasError() {
			return nil, diags
		}
		var r awstypes.GatewayProtocolConfigurationMemberMcp
		smerr.EnrichAppend(ctx, &diags, flex.Expand(ctx, model, &r.Value))
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
