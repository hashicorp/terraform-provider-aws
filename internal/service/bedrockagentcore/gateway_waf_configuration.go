// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package bedrockagentcore

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/bedrockagentcorecontrol"
	awstypes "github.com/aws/aws-sdk-go-v2/service/bedrockagentcorecontrol/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_bedrockagentcore_gateway_waf_configuration", name="Gateway WAF Configuration")
func newGatewayWAFConfigurationResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &gatewayWAFConfigurationResource{}

	r.SetDefaultCreateTimeout(30 * time.Minute)
	r.SetDefaultUpdateTimeout(30 * time.Minute)
	r.SetDefaultDeleteTimeout(30 * time.Minute)

	return r, nil
}

type gatewayWAFConfigurationResource struct {
	framework.ResourceWithModel[gatewayWAFConfigurationResourceModel]
	framework.WithTimeouts
}

func (r *gatewayWAFConfigurationResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"failure_mode": schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.WafFailureMode](),
				Required:   true,
			},
			"gateway_identifier": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"web_acl_arn": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
		Blocks: map[string]schema.Block{
			names.AttrTimeouts: timeouts.Block(ctx, timeouts.Opts{
				Create: true,
				Update: true,
				Delete: true,
			}),
		},
	}
}

func (r *gatewayWAFConfigurationResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data gatewayWAFConfigurationResourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.Plan.Get(ctx, &data))
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().BedrockAgentCoreClient(ctx)

	gatewayID := fwflex.StringValueFromFramework(ctx, data.GatewayIdentifier)
	wafConfiguration := &awstypes.WafConfiguration{
		FailureMode: data.FailureMode.ValueEnum(),
	}

	if err := r.putWAFConfiguration(ctx, conn, gatewayID, wafConfiguration, r.CreateTimeout(ctx, data.Timeouts)); err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, gatewayID)
		return
	}

	gateway, err := findGatewayByID(ctx, conn, gatewayID)
	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, gatewayID)
		return
	}

	data.WebACLARN = fwflex.StringToFramework(ctx, gateway.WebAclArn)

	smerr.AddEnrich(ctx, &response.Diagnostics, response.State.Set(ctx, &data))
}

func (r *gatewayWAFConfigurationResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data gatewayWAFConfigurationResourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.State.Get(ctx, &data))
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().BedrockAgentCoreClient(ctx)

	gatewayID := fwflex.StringValueFromFramework(ctx, data.GatewayIdentifier)
	gateway, err := findGatewayByID(ctx, conn, gatewayID)
	if retry.NotFound(err) {
		smerr.AddOne(ctx, &response.Diagnostics, fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, gatewayID)
		return
	}

	// The WAF configuration was cleared out-of-band (e.g. the WebACL association was
	// removed). Treat the resource as gone so Terraform plans to recreate it.
	if gateway.WafConfiguration == nil {
		response.State.RemoveResource(ctx)
		return
	}

	data.FailureMode = fwtypes.StringEnumValue(gateway.WafConfiguration.FailureMode)
	data.WebACLARN = fwflex.StringToFramework(ctx, gateway.WebAclArn)

	smerr.AddEnrich(ctx, &response.Diagnostics, response.State.Set(ctx, &data))
}

func (r *gatewayWAFConfigurationResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var plan, state gatewayWAFConfigurationResourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.Plan.Get(ctx, &plan))
	if response.Diagnostics.HasError() {
		return
	}
	smerr.AddEnrich(ctx, &response.Diagnostics, request.State.Get(ctx, &state))
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().BedrockAgentCoreClient(ctx)

	gatewayID := fwflex.StringValueFromFramework(ctx, plan.GatewayIdentifier)

	// failure_mode is the only server-relevant mutable attribute (gateway_identifier is
	// RequiresReplace, web_acl_arn is Computed, and timeouts is client-side only). Skip the
	// read-modify-write UpdateGateway round trip when it is unchanged so a plan that only
	// edits the timeouts block does not needlessly drive the gateway through UPDATING.
	if !plan.FailureMode.Equal(state.FailureMode) {
		wafConfiguration := &awstypes.WafConfiguration{
			FailureMode: plan.FailureMode.ValueEnum(),
		}

		if err := r.putWAFConfiguration(ctx, conn, gatewayID, wafConfiguration, r.UpdateTimeout(ctx, plan.Timeouts)); err != nil {
			smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, gatewayID)
			return
		}

		gateway, err := findGatewayByID(ctx, conn, gatewayID)
		if err != nil {
			smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, gatewayID)
			return
		}

		plan.WebACLARN = fwflex.StringToFramework(ctx, gateway.WebAclArn)
	} else {
		plan.WebACLARN = state.WebACLARN
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, response.State.Set(ctx, &plan))
}

func (r *gatewayWAFConfigurationResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data gatewayWAFConfigurationResourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.State.Get(ctx, &data))
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().BedrockAgentCoreClient(ctx)

	gatewayID := fwflex.StringValueFromFramework(ctx, data.GatewayIdentifier)

	// Clearing the WAF configuration requires a WebACL to still be associated with the
	// gateway. During destroy the aws_wafv2_web_acl_association is often torn down at the
	// same time; once it is gone the gateway no longer has WAF behavior, so a failure to
	// clear is benign (the configuration is effectively already removed).
	err := r.putWAFConfiguration(ctx, conn, gatewayID, nil, r.DeleteTimeout(ctx, data.Timeouts))
	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return
	}
	if errs.IsAErrorMessageContains[*awstypes.ValidationException](err, "WebACL") {
		return
	}
	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, gatewayID)
		return
	}
}

func (r *gatewayWAFConfigurationResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("gateway_identifier"), request, response)
}

// putWAFConfiguration performs a read-modify-write against the gateway: it reads the
// current gateway, preserves all of its fields, sets the WAF configuration (nil clears
// it), calls UpdateGateway, and waits for the gateway to become ready. UpdateGateway is a
// full replacement, so every field must be carried over to avoid clobbering attributes
// managed by the aws_bedrockagentcore_gateway resource.
func (r *gatewayWAFConfigurationResource) putWAFConfiguration(ctx context.Context, conn *bedrockagentcorecontrol.Client, gatewayID string, wafConfiguration *awstypes.WafConfiguration, timeout time.Duration) error {
	gateway, err := findGatewayByID(ctx, conn, gatewayID)
	if err != nil {
		return err
	}

	input := expandUpdateGatewayFromGateway(gatewayID, gateway, wafConfiguration)
	if _, err := conn.UpdateGateway(ctx, input); err != nil {
		return err
	}

	if err := waitGatewayUpdated(ctx, conn, gatewayID, timeout); err != nil {
		return err
	}

	return nil
}

func expandUpdateGatewayFromGateway(gatewayID string, gateway *bedrockagentcorecontrol.GetGatewayOutput, wafConfiguration *awstypes.WafConfiguration) *bedrockagentcorecontrol.UpdateGatewayInput {
	return &bedrockagentcorecontrol.UpdateGatewayInput{
		GatewayIdentifier:            aws.String(gatewayID),
		Name:                         gateway.Name,
		RoleArn:                      gateway.RoleArn,
		AuthorizerType:               gateway.AuthorizerType,
		AuthorizerConfiguration:      gateway.AuthorizerConfiguration,
		CustomTransformConfiguration: gateway.CustomTransformConfiguration,
		Description:                  gateway.Description,
		ExceptionLevel:               gateway.ExceptionLevel,
		InterceptorConfigurations:    gateway.InterceptorConfigurations,
		KmsKeyArn:                    gateway.KmsKeyArn,
		PolicyEngineConfiguration:    gateway.PolicyEngineConfiguration,
		ProtocolConfiguration:        gateway.ProtocolConfiguration,
		ProtocolType:                 gateway.ProtocolType,
		WafConfiguration:             wafConfiguration,
	}
}

type gatewayWAFConfigurationResourceModel struct {
	framework.WithRegionModel
	FailureMode       fwtypes.StringEnum[awstypes.WafFailureMode] `tfsdk:"failure_mode"`
	GatewayIdentifier types.String                                `tfsdk:"gateway_identifier"`
	Timeouts          timeouts.Value                              `tfsdk:"timeouts"`
	WebACLARN         types.String                                `tfsdk:"web_acl_arn"`
}
