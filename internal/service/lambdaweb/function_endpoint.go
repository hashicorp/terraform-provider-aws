// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package lambdaweb

import (
	"context"
	"fmt"
	"maps"
	"time"

	"github.com/YakDriver/smarterr"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/lambdaweb"
	awstypes "github.com/aws/aws-sdk-go-v2/service/lambdaweb/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/mapplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setplanmodifier"
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
	tfstringvalidator "github.com/hashicorp/terraform-provider-aws/internal/framework/validators/stringvalidator"
	tfmaps "github.com/hashicorp/terraform-provider-aws/internal/maps"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_lambdaweb_function_endpoint", name="Function Endpoint")
// @IdentityAttribute("function_name")
// @IdentityAttribute("endpoint_name")
// @ImportIDHandler("functionEndpointImportID")
// @Testing(importStateIdAttribute="endpoint_name")
// @Testing(hasNoPreExistingResource=true)
// @Testing(preCheck="testAccPreCheck")
// @Testing(existsType="github.com/aws/aws-sdk-go-v2/service/lambdaweb;lambdaweb.GetWebFunctionEndpointOutput")
// @Testing(importStateIdFunc="testAccFunctionEndpointImportStateIDFunc")
func newFunctionEndpointResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &functionEndpointResource{}

	r.SetDefaultCreateTimeout(15 * time.Minute)
	r.SetDefaultUpdateTimeout(15 * time.Minute)
	r.SetDefaultDeleteTimeout(15 * time.Minute)

	return r, nil
}

type functionEndpointResource struct {
	framework.ResourceWithModel[functionEndpointResourceModel]
	framework.WithTimeouts
	framework.WithImportByIdentity
}

func (r *functionEndpointResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrARN: framework.ARNAttributeComputedOnly(),
			// Editable in place: UpdateWebFunctionEndpoint accepts authType and
			// the domain name survives the change (verified live 2026-10-05).
			"auth_type": schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.AuthType](),
				Required:   true,
			},
			"auto_deployment_mode": schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.AutoDeploymentMode](),
				Optional:   true,
				Computed:   true,
				Validators: []validator.String{
					tfstringvalidator.AlsoRequiresWhenEquals(
						awstypes.AutoDeploymentModeDisabled,
						path.MatchRoot("revision_weight"),
					),
					tfstringvalidator.ConflictsWithWhenNotEquals(
						awstypes.AutoDeploymentModeDisabled,
						path.MatchRoot("revision_weight"),
					),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			names.AttrDescription: schema.StringAttribute{
				Optional: true,
			},
			names.AttrDomainName: schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"endpoint_name": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 64),
				},
			},
			names.AttrEndpointType: schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.EndpointType](),
				Required:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"function_name": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 64),
				},
			},
			"regional_domain_names": schema.MapAttribute{
				CustomType:  fwtypes.MapOfStringType,
				Computed:    true,
				ElementType: types.StringType,
				PlanModifiers: []planmodifier.Map{
					mapplanmodifier.UseStateForUnknown(),
				},
			},
			"regions": schema.SetAttribute{
				CustomType:  fwtypes.SetOfStringType,
				Optional:    true,
				Computed:    true,
				ElementType: types.StringType,
				PlanModifiers: []planmodifier.Set{
					setplanmodifier.RequiresReplace(),
					setplanmodifier.UseStateForUnknown(),
				},
				Validators: []validator.Set{
					// Model limit. The number of regions an endpoint may
					// replicate to is a service quota, enforced by the API.
					setvalidator.SizeBetween(1, 100),
				},
			},
			"scaling_config":  framework.ResourceOptionalComputedListOfObjectsAttribute(ctx, framework.WithPlanModifiers[scalingConfigModel](listplanmodifier.UseStateForUnknown())),
			"throttle_config": framework.ResourceOptionalComputedListOfObjectsAttribute(ctx, framework.WithPlanModifiers[throttleConfigModel](listplanmodifier.UseStateForUnknown())),
		},
		Blocks: map[string]schema.Block{
			"revision_weight": schema.ListNestedBlock{
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

func (r *functionEndpointResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var data functionEndpointResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Config.Get(ctx, &data))
	if resp.Diagnostics.HasError() {
		return
	}

	if data.AutoDeploymentMode.IsUnknown() || data.EndpointType.IsUnknown() || data.RevisionWeights.IsUnknown() {
		return
	}

	weights, d := data.RevisionWeights.ToSlice(ctx)
	smerr.AddEnrich(ctx, &resp.Diagnostics, d)
	if resp.Diagnostics.HasError() {
		return
	}
	hasWeights := len(weights) > 0

	endpointType := data.EndpointType.ValueEnum()
	multiRegional := endpointType == awstypes.EndpointTypeMultiRegion || endpointType == awstypes.EndpointTypePerRegion
	if mode := data.AutoDeploymentMode.ValueEnum(); multiRegional && mode == awstypes.AutoDeploymentModeLatestRevision {
		resp.Diagnostics.AddAttributeError(path.Root("auto_deployment_mode"),
			"Invalid auto_deployment_mode",
			fmt.Sprintf("%s endpoints require `auto_deployment_mode = \"Disabled\"` with explicit `revision_weights`.", endpointType))
	}

	// A PerRegion endpoint needs at least 2 distinct regions after the service
	// adds the home region, so omitting regions is rejected by the API
	// (verified live 2026-10-05). A single region is valid when it is not the
	// home region; the provider cannot tell at validation time, so that case
	// is left to the API, whose error message names the resulting region set.
	if endpointType == awstypes.EndpointTypePerRegion && data.Regions.IsNull() {
		resp.Diagnostics.AddAttributeError(path.Root("regions"),
			"Missing regions",
			fmt.Sprintf("%s endpoints require `regions` with at least one Region other than the home Region, which is added automatically.", endpointType))
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

func (r *functionEndpointResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan functionEndpointResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Plan.Get(ctx, &plan))
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().LambdaWebClient(ctx)

	functionName, endpointName := fwflex.StringValueFromFramework(ctx, plan.FunctionName), fwflex.StringValueFromFramework(ctx, plan.EndpointName)
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

	_, err := conn.CreateWebFunctionEndpoint(ctx, &input)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, endpointName)
		return
	}

	out, err := waitFunctionEndpointActive(ctx, conn, functionName, endpointName, r.CreateTimeout(ctx, plan.Timeouts))
	if err != nil {
		// Taint the resource.
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("function_name"), functionName)...)
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("endpoint_name"), endpointName)...)
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, endpointName)
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, r.flatten(ctx, out, &plan))
	if resp.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &plan))
}

func (r *functionEndpointResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state functionEndpointResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().LambdaWebClient(ctx)

	functionName, endpointName := fwflex.StringValueFromFramework(ctx, state.FunctionName), fwflex.StringValueFromFramework(ctx, state.EndpointName)
	out, err := findWebFunctionEndpointByTwoPartKey(ctx, conn, functionName, endpointName)
	if retry.NotFound(err) {
		resp.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, endpointName)
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, r.flatten(ctx, out, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &state))
}

func (r *functionEndpointResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state functionEndpointResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Plan.Get(ctx, &plan))
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().LambdaWebClient(ctx)

	functionName, endpointName := fwflex.StringValueFromFramework(ctx, plan.FunctionName), fwflex.StringValueFromFramework(ctx, plan.EndpointName)
	var input lambdaweb.UpdateWebFunctionEndpointInput
	smerr.AddEnrich(ctx, &resp.Diagnostics, fwflex.Expand(ctx, plan, &input))
	if resp.Diagnostics.HasError() {
		return
	}

	// An omitted description means "leave unchanged" to the API, so removing one
	// from the configuration would never reach the service and the read-back
	// would not match the plan. An empty string is what clears it.
	if plan.Description.IsNull() && !state.Description.IsNull() {
		input.Description = aws.String("")
	}

	_, err := conn.UpdateWebFunctionEndpoint(ctx, &input)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, endpointName)
		return
	}

	out, err := waitFunctionEndpointUpdated(ctx, conn, functionName, endpointName, r.UpdateTimeout(ctx, plan.Timeouts))
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, endpointName)
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, r.flatten(ctx, out, &plan))
	if resp.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &plan))
}

func (r *functionEndpointResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state functionEndpointResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().LambdaWebClient(ctx)

	functionName, endpointName := fwflex.StringValueFromFramework(ctx, state.FunctionName), fwflex.StringValueFromFramework(ctx, state.EndpointName)
	input := lambdaweb.DeleteWebFunctionEndpointInput{
		FunctionName: aws.String(functionName),
		EndpointName: aws.String(endpointName),
	}
	_, err := conn.DeleteWebFunctionEndpoint(ctx, &input)
	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return
	}
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, endpointName)
		return
	}

	if _, err := waitFunctionEndpointDeleted(ctx, conn, functionName, endpointName, r.DeleteTimeout(ctx, state.Timeouts)); err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, endpointName)
		return
	}
}

func (r *functionEndpointResource) flatten(ctx context.Context, endpoint *lambdaweb.GetWebFunctionEndpointOutput, data *functionEndpointResourceModel) diag.Diagnostics {
	var diags diag.Diagnostics

	// normalizeDescription maps the empty string the API reports for a cleared
	// description back to null, so it matches a configuration that omits it.
	if aws.ToString(endpoint.Description) == "" {
		endpoint.Description = nil
	}

	diags.Append(fwflex.Flatten(ctx, endpoint, data)...)
	if diags.HasError() {
		return diags
	}

	data.RegionalDomainNames = regionalDomainNames(ctx, endpoint.RegionalEndpoints)

	// normalizeRevisionWeights keeps revision_weights as configuration-only state:
	// under LatestRevision the server reports its ephemeral routing weights, which
	// must not surface as drift for a block the practitioner cannot configure.
	// Under Disabled the weights are required configuration and reflect the server.
	if data.AutoDeploymentMode.ValueEnum() == awstypes.AutoDeploymentModeLatestRevision {
		data.RevisionWeights = fwtypes.NewListNestedObjectValueOfNull[revisionWeightModel](ctx)
	}

	return diags
}

// regionalDomainNames projects the API regionalEndpoints map (region ->
// RegionalEndpoint) onto the flat region -> domain name map exposed in state;
// AutoFlex cannot derive this field projection. Only PerRegion endpoints carry
// per-region domains: MultiRegion regional entries have no domain of their own
// because traffic uses the global domain.
func regionalDomainNames(ctx context.Context, regionalEndpoints map[string]awstypes.RegionalEndpoint) fwtypes.MapOfString {
	m := tfmaps.ApplyToAllValues(regionalEndpoints, func(v awstypes.RegionalEndpoint) string {
		return aws.ToString(v.DomainName)
	})
	maps.DeleteFunc(m, func(k string, v string) bool { return v == "" })
	return fwflex.FlattenFrameworkStringValueMapOfString(ctx, m)
}

var _ inttypes.ImportIDParser = functionEndpointImportID{}

type functionEndpointImportID struct{}

func (functionEndpointImportID) Parse(id string) (string, map[string]any, error) {
	const (
		importIDParts = 2
	)
	parts, err := intflex.ExpandResourceId(id, importIDParts, true)
	if err != nil {
		return "", nil, err
	}

	return id, map[string]any{
		"function_name": parts[0],
		"endpoint_name": parts[1],
	}, nil
}

func findWebFunctionEndpointByTwoPartKey(ctx context.Context, conn *lambdaweb.Client, functionName, endpointName string) (*lambdaweb.GetWebFunctionEndpointOutput, error) {
	input := lambdaweb.GetWebFunctionEndpointInput{
		FunctionName: aws.String(functionName),
		EndpointName: aws.String(endpointName),
	}
	return findWebFunctionEndpoint(ctx, conn, &input)
}

func findWebFunctionEndpoint(ctx context.Context, conn *lambdaweb.Client, input *lambdaweb.GetWebFunctionEndpointInput) (*lambdaweb.GetWebFunctionEndpointOutput, error) {
	out, err := conn.GetWebFunctionEndpoint(ctx, input)
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

func statusFunctionEndpoint(conn *lambdaweb.Client, functionName, endpointName string) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		out, err := findWebFunctionEndpointByTwoPartKey(ctx, conn, functionName, endpointName)
		if retry.NotFound(err) {
			return nil, "", nil
		}
		if err != nil {
			return nil, "", smarterr.NewError(err)
		}

		return out, string(out.State), nil
	}
}

func statusFunctionEndpointUpdate(conn *lambdaweb.Client, functionName, endpointName string) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		out, err := findWebFunctionEndpointByTwoPartKey(ctx, conn, functionName, endpointName)
		if retry.NotFound(err) {
			return nil, "", nil
		}
		if err != nil {
			return nil, "", smarterr.NewError(err)
		}

		return out, string(out.UpdateStatus), nil
	}
}

func waitFunctionEndpointActive(ctx context.Context, conn *lambdaweb.Client, functionName, endpointName string, timeout time.Duration) (*lambdaweb.GetWebFunctionEndpointOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   enum.Slice(awstypes.EndpointStatePending),
		Target:                    enum.Slice(awstypes.EndpointStateActive),
		Refresh:                   statusFunctionEndpoint(conn, functionName, endpointName),
		Timeout:                   timeout,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*lambdaweb.GetWebFunctionEndpointOutput); ok {
		retry.SetLastError(err, endpointError(out))
		return out, smarterr.NewError(err)
	}

	return nil, smarterr.NewError(err)
}

func waitFunctionEndpointUpdated(ctx context.Context, conn *lambdaweb.Client, functionName, endpointName string, timeout time.Duration) (*lambdaweb.GetWebFunctionEndpointOutput, error) {
	stateConf := &retry.StateChangeConf{
		// An empty updateStatus can be observed if the endpoint has never been
		// updated before and the PATCH has not propagated to the Get yet.
		Pending:                   append(enum.Slice(awstypes.EndpointUpdateStatusInProgress), ""),
		Target:                    enum.Slice(awstypes.EndpointUpdateStatusSuccessful),
		Refresh:                   statusFunctionEndpointUpdate(conn, functionName, endpointName),
		Timeout:                   timeout,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*lambdaweb.GetWebFunctionEndpointOutput); ok {
		retry.SetLastError(err, endpointUpdateError(out))
		return out, smarterr.NewError(err)
	}

	return nil, smarterr.NewError(err)
}

func waitFunctionEndpointDeleted(ctx context.Context, conn *lambdaweb.Client, functionName, endpointName string, timeout time.Duration) (*lambdaweb.GetWebFunctionEndpointOutput, error) {
	stateConf := &retry.StateChangeConf{
		// Every state an endpoint can hold on the way out, because the target is
		// "gone" and anything unlisted aborts the destroy. A MultiRegion endpoint
		// sits Pending for minutes while it replicates and settles in Failed when
		// a region cannot take the revision, and both accept a delete: observing
		// either one is a matter of when the first refresh lands.
		Pending: enum.Values[awstypes.EndpointState](),
		Target:  []string{},
		Refresh: statusFunctionEndpoint(conn, functionName, endpointName),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*lambdaweb.GetWebFunctionEndpointOutput); ok {
		retry.SetLastError(err, endpointError(out))
		return out, smarterr.NewError(err)
	}

	return nil, smarterr.NewError(err)
}

type functionEndpointResourceModel struct {
	framework.WithRegionModel
	AuthType            fwtypes.StringEnum[awstypes.AuthType]                `tfsdk:"auth_type"`
	AutoDeploymentMode  fwtypes.StringEnum[awstypes.AutoDeploymentMode]      `tfsdk:"auto_deployment_mode"`
	Description         types.String                                         `tfsdk:"description"`
	DomainName          types.String                                         `tfsdk:"domain_name"`
	EndpointARN         types.String                                         `tfsdk:"arn"`
	EndpointName        types.String                                         `tfsdk:"endpoint_name"`
	EndpointType        fwtypes.StringEnum[awstypes.EndpointType]            `tfsdk:"endpoint_type"`
	FunctionName        types.String                                         `tfsdk:"function_name"`
	RegionalDomainNames fwtypes.MapOfString                                  `tfsdk:"regional_domain_names"`
	Regions             fwtypes.SetOfString                                  `tfsdk:"regions"`
	RevisionWeights     fwtypes.ListNestedObjectValueOf[revisionWeightModel] `tfsdk:"revision_weight"`
	ScalingConfig       fwtypes.ListNestedObjectValueOf[scalingConfigModel]  `tfsdk:"scaling_config"`
	ThrottleConfig      fwtypes.ListNestedObjectValueOf[throttleConfigModel] `tfsdk:"throttle_config"`
	Timeouts            timeouts.Value                                       `tfsdk:"timeouts"`
}

type revisionWeightModel struct {
	RevisionID types.String `tfsdk:"revision_id"`
	Weight     types.Int64  `tfsdk:"weight"`
}
