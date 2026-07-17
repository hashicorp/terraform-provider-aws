// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// Pre-GA: this resource targets the not-yet-public aws-sdk-go-v2
// "lambdaweb" service client, currently satisfied by the hand-written
// shim in .pre-ga-sdk/ (see the replace directive in go.mod). Swap to the
// real SDK module when it ships at GA.

package lambdaweb

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/lambdaweb"
	awstypes "github.com/aws/aws-sdk-go-v2/service/lambdaweb/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_lambdaweb_endpoint", name="Endpoint")
// @IdentityAttribute("function_name")
// @IdentityAttribute("endpoint_name")
// @ImportIDHandler("endpointImportID")
// @Testing(hasNoPreExistingResource=true)
// @Testing(existsType="github.com/aws/aws-sdk-go-v2/service/lambdaweb;lambdaweb.GetWebFunctionEndpointOutput")
// @Testing(importStateIdFunc="testAccEndpointImportStateIDFunc")
func newEndpointResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &endpointResource{}

	r.SetDefaultCreateTimeout(15 * time.Minute)
	r.SetDefaultUpdateTimeout(15 * time.Minute)
	r.SetDefaultDeleteTimeout(15 * time.Minute)

	return r, nil
}

const (
	ResNameEndpoint = "Endpoint"
)

type endpointResource struct {
	framework.ResourceWithModel[endpointResourceModel]
	framework.WithTimeouts
	framework.WithImportByIdentity
}

func (r *endpointResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
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
			"endpoint_name": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			names.AttrEndpointType: schema.StringAttribute{
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
				Computed:    true,
				ElementType: types.StringType,
				PlanModifiers: []planmodifier.List{
					listplanmodifier.RequiresReplace(),
					listplanmodifier.UseStateForUnknown(),
				},
				Validators: []validator.List{
					listvalidator.SizeAtMost(5), // service replication quota (model ceiling is 100)
				},
			},
			names.AttrDomainName: schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			names.AttrState: schema.StringAttribute{
				Computed: true,
			},
		},
		Blocks: map[string]schema.Block{
			"revision_weights": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[revisionWeightModel](ctx),
				Validators: []validator.List{
					listvalidator.SizeBetween(1, 2),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"revision_id": schema.StringAttribute{
							Required: true,
						},
						names.AttrWeight: schema.Int64Attribute{
							Required: true,
							Validators: []validator.Int64{
								int64validator.Between(1, 100),
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

func (r *endpointResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var cfg endpointResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Config.Get(ctx, &cfg))
	if resp.Diagnostics.HasError() {
		return
	}
	if cfg.AutoDeploymentMode.IsUnknown() || cfg.RevisionWeights.IsUnknown() || cfg.EndpointType.IsUnknown() {
		return
	}

	mode := awstypes.AutoDeploymentModeLatestRevision
	if !cfg.AutoDeploymentMode.IsNull() {
		mode = awstypes.AutoDeploymentMode(cfg.AutoDeploymentMode.ValueString())
	}

	weights, d := cfg.RevisionWeights.ToSlice(ctx)
	smerr.AddEnrich(ctx, &resp.Diagnostics, d)
	if resp.Diagnostics.HasError() {
		return
	}
	hasWeights := len(weights) > 0

	if mode == awstypes.AutoDeploymentModeLatestRevision && hasWeights {
		resp.Diagnostics.AddAttributeError(path.Root("revision_weights"),
			"Invalid revision_weights",
			"`revision_weights` must be omitted when `auto_deployment_mode` is `LatestRevision`.")
	}
	if mode == awstypes.AutoDeploymentModeDisabled && !hasWeights {
		resp.Diagnostics.AddAttributeError(path.Root("revision_weights"),
			"Missing revision_weights",
			"`revision_weights` is required when `auto_deployment_mode` is `Disabled`.")
	}
	if cfg.EndpointType.ValueString() == string(awstypes.EndpointTypeMultiRegion) && mode == awstypes.AutoDeploymentModeLatestRevision {
		resp.Diagnostics.AddAttributeError(path.Root("auto_deployment_mode"),
			"Invalid auto_deployment_mode",
			"MultiRegion endpoints require `auto_deployment_mode = \"Disabled\"` with explicit `revision_weights`.")
	}

	if hasWeights {
		var sum int64
		seen := make(map[string]struct{}, len(weights))
		for _, w := range weights {
			if w.Weight.IsUnknown() || w.RevisionID.IsUnknown() {
				return
			}
			sum += w.Weight.ValueInt64()
			id := w.RevisionID.ValueString()
			if _, dup := seen[id]; dup {
				resp.Diagnostics.AddAttributeError(path.Root("revision_weights"),
					"Duplicate revision_id",
					fmt.Sprintf("`revision_weights` must not contain duplicate revision ids; %q appears more than once.", id))
			}
			seen[id] = struct{}{}
		}
		if sum != 100 {
			resp.Diagnostics.AddAttributeError(path.Root("revision_weights"),
				"Invalid revision_weights sum",
				fmt.Sprintf("`revision_weights` weights must sum to 100 (got %d). Use a single entry at weight 100, or two entries summing to 100.", sum))
		}
	}
}

func (r *endpointResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := r.Meta().LambdaWebClient(ctx)

	var plan endpointResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Plan.Get(ctx, &plan))
	if resp.Diagnostics.HasError() {
		return
	}

	var input lambdaweb.CreateWebFunctionEndpointInput
	smerr.AddEnrich(ctx, &resp.Diagnostics, fwflex.Expand(ctx, plan, &input))
	if resp.Diagnostics.HasError() {
		return
	}

	// The API defaults a missing autoDeploymentMode to Disabled, which then
	// rejects the request unless revisionWeights are supplied. Default to
	// LatestRevision explicitly (mirrors ValidateConfig's assumption).
	if input.AutoDeploymentMode == "" && len(input.RevisionWeights) == 0 {
		input.AutoDeploymentMode = awstypes.AutoDeploymentModeLatestRevision
	}

	functionName := plan.FunctionName.ValueString()
	endpointName := plan.EndpointName.ValueString()

	if _, err := conn.CreateWebFunctionEndpoint(ctx, &input); err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, endpointName)
		return
	}

	out, err := waitEndpointActive(ctx, conn, functionName, endpointName, r.CreateTimeout(ctx, plan.Timeouts))
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, endpointName)
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, fwflex.Flatten(ctx, out, &plan))
	if resp.Diagnostics.HasError() {
		return
	}
	normalizeRevisionWeights(ctx, &plan)

	plan.ID = types.StringValue(endpointCreateResourceID(functionName, endpointName))
	plan.ARN = fwflex.StringToFramework(ctx, out.EndpointArn)

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &plan))
}

func (r *endpointResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().LambdaWebClient(ctx)

	var state endpointResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := findEndpointByName(ctx, conn, state.FunctionName.ValueString(), state.EndpointName.ValueString())
	if retry.NotFound(err) {
		resp.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, state.EndpointName.ValueString())
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, fwflex.Flatten(ctx, out, &state))
	if resp.Diagnostics.HasError() {
		return
	}
	normalizeRevisionWeights(ctx, &state)

	state.ID = types.StringValue(endpointCreateResourceID(state.FunctionName.ValueString(), state.EndpointName.ValueString()))
	state.ARN = fwflex.StringToFramework(ctx, out.EndpointArn)

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &state))
}

func endpointCreateResourceID(functionName, endpointName string) string {
	return functionName + flex.ResourceIdSeparator + endpointName
}

func (r *endpointResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	conn := r.Meta().LambdaWebClient(ctx)

	var plan, state endpointResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Plan.Get(ctx, &plan))
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	functionName := plan.FunctionName.ValueString()
	endpointName := plan.EndpointName.ValueString()

	var input lambdaweb.UpdateWebFunctionEndpointInput
	smerr.AddEnrich(ctx, &resp.Diagnostics, fwflex.Expand(ctx, plan, &input))
	if resp.Diagnostics.HasError() {
		return
	}
	input.FunctionName = aws.String(functionName)
	input.EndpointName = aws.String(endpointName)

	if _, err := conn.UpdateWebFunctionEndpoint(ctx, &input); err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, endpointName)
		return
	}

	out, err := waitEndpointUpdated(ctx, conn, functionName, endpointName, r.UpdateTimeout(ctx, plan.Timeouts))
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, endpointName)
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, fwflex.Flatten(ctx, out, &plan))
	if resp.Diagnostics.HasError() {
		return
	}
	normalizeRevisionWeights(ctx, &plan)
	plan.ID = types.StringValue(endpointCreateResourceID(functionName, endpointName))
	plan.ARN = fwflex.StringToFramework(ctx, out.EndpointArn)

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &plan))
}

func (r *endpointResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := r.Meta().LambdaWebClient(ctx)

	var state endpointResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	functionName := state.FunctionName.ValueString()
	endpointName := state.EndpointName.ValueString()

	input := lambdaweb.DeleteWebFunctionEndpointInput{
		FunctionName: aws.String(functionName),
		EndpointName: aws.String(endpointName),
	}
	_, err := conn.DeleteWebFunctionEndpoint(ctx, &input)
	if err != nil {
		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return
		}
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, endpointName)
		return
	}

	if _, err := waitEndpointDeleted(ctx, conn, functionName, endpointName, r.DeleteTimeout(ctx, state.Timeouts)); err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, endpointName)
		return
	}
}

// normalizeRevisionWeights keeps revision_weights as configuration-only state:
// under LatestRevision the server reports its ephemeral routing weights, which
// must not surface as drift for a block the practitioner cannot configure.
// Under Disabled the weights are required configuration and reflect the server.
func normalizeRevisionWeights(ctx context.Context, m *endpointResourceModel) {
	if awstypes.AutoDeploymentMode(m.AutoDeploymentMode.ValueString()) == awstypes.AutoDeploymentModeLatestRevision {
		m.RevisionWeights = fwtypes.NewListNestedObjectValueOfNull[revisionWeightModel](ctx)
	}
}

var _ inttypes.ImportIDParser = endpointImportID{}

type endpointImportID struct{}

func (endpointImportID) Parse(id string) (string, map[string]any, error) {
	functionName, endpointName, found := strings.Cut(id, flex.ResourceIdSeparator)
	if !found {
		return "", nil, fmt.Errorf("id %q should be in the format <function-name>%s<endpoint-name>", id, flex.ResourceIdSeparator)
	}

	return id, map[string]any{
		"function_name": functionName,
		"endpoint_name": endpointName,
	}, nil
}

// ModifyPlan pins the volatile computed attribute (state) to prior state when no
// configurable attribute is changing, so a no-op plan stays empty; on create or a
// real update it stays unknown so the server's recomputed value avoids an
// "inconsistent result after apply" error.
func (r *endpointResource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	if req.State.Raw.IsNull() || req.Plan.Raw.IsNull() {
		return
	}

	var state, plan, config endpointResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Plan.Get(ctx, &plan))
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Config.Get(ctx, &config))
	if resp.Diagnostics.HasError() {
		return
	}

	unchanged := unchangedOrUnset(config.AuthType.StringValue, state.AuthType.StringValue) &&
		unchangedOrUnset(config.Description, state.Description) &&
		unchangedOrUnset(config.AutoDeploymentMode.StringValue, state.AutoDeploymentMode.StringValue) &&
		(config.RevisionWeights.IsNull() || config.RevisionWeights.Equal(state.RevisionWeights)) &&
		(config.Regions.IsNull() || config.Regions.Equal(state.Regions))

	if unchanged {
		plan.State = state.State
		smerr.AddEnrich(ctx, &resp.Diagnostics, resp.Plan.Set(ctx, &plan))
	}
}

func unchangedOrUnset(config, state types.String) bool {
	return config.IsNull() || config.Equal(state)
}

type endpointResourceModel struct {
	framework.WithRegionModel
	ARN                types.String                                         `tfsdk:"arn"`
	AuthType           fwtypes.StringEnum[awstypes.AuthType]                `tfsdk:"auth_type"`
	AutoDeploymentMode fwtypes.StringEnum[awstypes.AutoDeploymentMode]      `tfsdk:"auto_deployment_mode"`
	Description        types.String                                         `tfsdk:"description"`
	DomainName         types.String                                         `tfsdk:"domain_name"`
	EndpointName       types.String                                         `tfsdk:"endpoint_name"`
	EndpointType       fwtypes.StringEnum[awstypes.EndpointType]            `tfsdk:"endpoint_type"`
	FunctionName       types.String                                         `tfsdk:"function_name"`
	ID                 types.String                                         `tfsdk:"id"`
	Regions            fwtypes.ListOfString                                 `tfsdk:"regions"`
	RevisionWeights    fwtypes.ListNestedObjectValueOf[revisionWeightModel] `tfsdk:"revision_weights"`
	State              types.String                                         `tfsdk:"state"`
	Timeouts           timeouts.Value                                       `tfsdk:"timeouts"`
}

type revisionWeightModel struct {
	RevisionID types.String `tfsdk:"revision_id"`
	Weight     types.Int64  `tfsdk:"weight"`
}
