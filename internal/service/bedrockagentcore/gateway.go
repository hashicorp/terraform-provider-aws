// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

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
	sdkretry "github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
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
					listvalidator.IsRequired(),
					listvalidator.SizeAtLeast(1),
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

func (r *gatewayResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data gatewayResourceModel
	smerr.EnrichAppend(ctx, &response.Diagnostics, request.Plan.Get(ctx, &data))
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().BedrockAgentCoreClient(ctx)

	var input bedrockagentcorecontrol.CreateGatewayInput
	smerr.EnrichAppend(ctx, &response.Diagnostics, fwflex.Expand(ctx, data, &input))
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
	smerr.EnrichAppend(ctx, &response.Diagnostics, fwflex.Flatten(ctx, gateway, &data))
	if response.Diagnostics.HasError() {
		return
	}

	smerr.EnrichAppend(ctx, &response.Diagnostics, response.State.Set(ctx, data))
}

func (r *gatewayResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data gatewayResourceModel
	smerr.EnrichAppend(ctx, &response.Diagnostics, request.State.Get(ctx, &data))
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().BedrockAgentCoreClient(ctx)

	gatewayID := fwflex.StringValueFromFramework(ctx, data.GatewayID)
	out, err := findGatewayByID(ctx, conn, gatewayID)
	if tfresource.NotFound(err) {
		response.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, gatewayID)
		return
	}

	smerr.EnrichAppend(ctx, &response.Diagnostics, fwflex.Flatten(ctx, out, &data))
	if response.Diagnostics.HasError() {
		return
	}

	smerr.EnrichAppend(ctx, &response.Diagnostics, response.State.Set(ctx, &data))
}

func (r *gatewayResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var new, old gatewayResourceModel
	smerr.EnrichAppend(ctx, &response.Diagnostics, request.Plan.Get(ctx, &new))
	smerr.EnrichAppend(ctx, &response.Diagnostics, request.State.Get(ctx, &old))
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().BedrockAgentCoreClient(ctx)

	diff, d := fwflex.Diff(ctx, new, old)
	smerr.EnrichAppend(ctx, &response.Diagnostics, d)
	if response.Diagnostics.HasError() {
		return
	}

	if diff.HasChanges() {
		gatewayID := fwflex.StringValueFromFramework(ctx, new.GatewayID)
		var input bedrockagentcorecontrol.UpdateGatewayInput
		smerr.EnrichAppend(ctx, &response.Diagnostics, fwflex.Expand(ctx, new, &input))
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

	smerr.EnrichAppend(ctx, &response.Diagnostics, response.State.Set(ctx, &new))
}

func (r *gatewayResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data gatewayResourceModel
	smerr.EnrichAppend(ctx, &response.Diagnostics, request.State.Get(ctx, &data))
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
		tfresource.SetLastError(err, errors.New(strings.Join(out.StatusReasons, "; ")))
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
		tfresource.SetLastError(err, errors.New(strings.Join(out.StatusReasons, "; ")))
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
		tfresource.SetLastError(err, errors.New(strings.Join(out.StatusReasons, "; ")))
		return out, smarterr.NewError(err)
	}

	return nil, smarterr.NewError(err)
}

func statusGateway(conn *bedrockagentcorecontrol.Client, id string) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
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

	return findGateway(ctx, conn, &input)
}

func findGateway(ctx context.Context, conn *bedrockagentcorecontrol.Client, input *bedrockagentcorecontrol.GetGatewayInput) (*bedrockagentcorecontrol.GetGatewayOutput, error) {
	out, err := conn.GetGateway(ctx, input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, smarterr.NewError(&sdkretry.NotFoundError{
			LastError:   err,
			LastRequest: &input,
		})
	}

	if err != nil {
		return nil, smarterr.NewError(err)
	}

	if out == nil {
		return nil, smarterr.NewError(tfresource.NewEmptyResultError(&input))
	}

	return out, nil
}

type gatewayResourceModel struct {
	framework.WithRegionModel
	AuthorizerConfiguration fwtypes.ListNestedObjectValueOf[authorizerConfigurationModel]      `tfsdk:"authorizer_configuration"`
	AuthorizerType          fwtypes.StringEnum[awstypes.AuthorizerType]                        `tfsdk:"authorizer_type"`
	Description             types.String                                                       `tfsdk:"description"`
	ExceptionLevel          fwtypes.StringEnum[awstypes.ExceptionLevel]                        `tfsdk:"exception_level"`
	GatewayARN              types.String                                                       `tfsdk:"gateway_arn"`
	GatewayID               types.String                                                       `tfsdk:"gateway_id"`
	GatewayURL              types.String                                                       `tfsdk:"gateway_url"`
	KMSKeyARN               fwtypes.ARN                                                        `tfsdk:"kms_key_arn"`
	Name                    types.String                                                       `tfsdk:"name"`
	ProtocolConfiguration   fwtypes.ListNestedObjectValueOf[gatewayProtocolConfigurationModel] `tfsdk:"protocol_configuration"`
	ProtocolType            fwtypes.StringEnum[awstypes.GatewayProtocolType]                   `tfsdk:"protocol_type"`
	RoleARN                 fwtypes.ARN                                                        `tfsdk:"role_arn"`
	Tags                    tftags.Map                                                         `tfsdk:"tags"`
	TagsAll                 tftags.Map                                                         `tfsdk:"tags_all"`
	Timeouts                timeouts.Value                                                     `tfsdk:"timeouts"`
	WorkloadIdentityDetails fwtypes.ListNestedObjectValueOf[workloadIdentityDetailsModel]      `tfsdk:"workload_identity_details"`
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
		smerr.EnrichAppend(ctx, &diags, fwflex.Flatten(ctx, t.Value, &data))
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
		smerr.EnrichAppend(ctx, &diags, d)
		if diags.HasError() {
			return nil, diags
		}
		var r awstypes.GatewayProtocolConfigurationMemberMcp
		smerr.EnrichAppend(ctx, &diags, fwflex.Expand(ctx, data, &r.Value))
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
