// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package bedrockagentcore

import (
	"context"
	"fmt"
	"maps"
	"slices"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/YakDriver/smarterr"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/bedrockagentcorecontrol"
	awstypes "github.com/aws/aws-sdk-go-v2/service/bedrockagentcorecontrol/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/float64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/mapvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	intflex "github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_bedrockagentcore_gateway_rate_limit", name="Gateway Rate Limit")
// @IdentityAttribute("gateway_identifier")
// @IdentityAttribute("rate_limit_id")
// @ImportIDHandler("gatewayRateLimitImportID")
// New resource, so there is no pre-identity provider version to test against.
// @Testing(hasNoPreExistingResource=true)
// Two-part import ID, so the generated identity tests need to be told how to
// build it; without these they fall back to a plain "id" attribute this
// resource does not have.
// @Testing(importStateIdFunc="testAccGatewayRateLimitImportStateIDFunc")
// @Testing(importStateIdAttribute="rate_limit_id")
func newGatewayRateLimitResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &gatewayRateLimitResource{}

	r.SetDefaultCreateTimeout(5 * time.Minute)
	r.SetDefaultUpdateTimeout(5 * time.Minute)
	r.SetDefaultDeleteTimeout(5 * time.Minute)

	return r, nil
}

type gatewayRateLimitResource struct {
	framework.ResourceWithModel[gatewayRateLimitResourceModel]
	framework.WithImportByIdentity
	framework.WithTimeouts
}

// dimensionKeyPattern mirrors the API's published pattern for dimension key
// names. Note the JWT arm embeds an arbitrary claim name, so the set of valid
// keys is open-ended and cannot be expressed as an enum.
var dimensionKeyPattern = regexache.MustCompile(`^(targetName|toolName|qualifiedModelId|\$\.context\.iam\.principal|\$\.context\.iam\.sourceIdentity|\$\.context\.jwt\.[a-zA-Z_][a-zA-Z0-9_\-.]{0,61}[a-zA-Z0-9_])$`)

const dimensionWildcard = "*"

func dimensionKeyValidators() []validator.String {
	return []validator.String{
		stringvalidator.LengthBetween(1, 80),
		stringvalidator.RegexMatches(dimensionKeyPattern, "must be one of targetName, toolName, qualifiedModelId, $.context.iam.principal, $.context.iam.sourceIdentity, or $.context.jwt.<claim>"),
	}
}

// rateConfigBlock builds one of the three metric blocks. The API caps each at a
// single entry, and constrains which periods each metric accepts: requests
// takes both, tokens is minute-only, connections is second-only.
func rateConfigBlock(ctx context.Context, periods ...awstypes.Period) schema.ListNestedBlock {
	periodValidators := []validator.String{}
	if len(periods) > 0 {
		allowed := make([]string, 0, len(periods))
		for _, p := range periods {
			allowed = append(allowed, string(p))
		}
		periodValidators = append(periodValidators, stringvalidator.OneOf(allowed...))
	}

	return schema.ListNestedBlock{
		CustomType: fwtypes.NewListNestedObjectTypeOf[rateConfigModel](ctx),
		Validators: []validator.List{
			listvalidator.SizeAtMost(1),
			// "at least one metric per entry" is NOT enforced here: a relative
			// AtParent() path from inside a nested block does not resolve to the
			// entry object. entriesMatchDimensionKeysValidator does it instead,
			// where the whole entry is in scope.
		},
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				"period": schema.StringAttribute{
					Required:   true,
					CustomType: fwtypes.StringEnumType[awstypes.Period](),
					Validators: periodValidators,
				},
				"rate": schema.Float64Attribute{
					Required: true,
					// Zero is meaningful: it blocks all matching traffic.
					Validators: []validator.Float64{
						float64validator.Between(0, 10000000),
					},
				},
			},
		},
	}
}

func (r *gatewayRateLimitResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrDescription: schema.StringAttribute{
				// Optional but NOT Computed: unlike UpdateGatewayRule, this API
				// treats an omitted description as "clear it", so removing the
				// argument genuinely removes the value. Sending "" is a
				// different thing - it sets an empty string.
				Optional: true,
				Validators: []validator.String{
					stringvalidator.LengthAtMost(512),
				},
			},
			"dimension_keys": schema.ListAttribute{
				// A list, not a set: ordering is load-bearing. Wildcards are
				// legal only in trailing positions, judged against this order.
				CustomType: fwtypes.ListOfStringType,
				Required:   true,
				Validators: []validator.List{
					listvalidator.SizeBetween(1, 10),
					listvalidator.ValueStringsAre(dimensionKeyValidators()...),
				},
				PlanModifiers: []planmodifier.List{
					// Immutable: UpdateGatewayRateLimit has no dimensionKeys field.
					listplanmodifier.RequiresReplace(),
				},
			},
			"gateway_identifier": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"rate_limit_id": schema.StringAttribute{
				// Optional+Computed: the API accepts a customer-defined id and
				// generates one otherwise. There is no name attribute on this
				// object, so this is its only human-readable handle.
				Optional: true,
				Computed: true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(2, 64),
					stringvalidator.RegexMatches(
						regexache.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9\-_.]{0,62}[a-zA-Z0-9]$`),
						"must be 2-64 characters, starting and ending with an alphanumeric character",
					),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
		Blocks: map[string]schema.Block{
			"entries": schema.SetNestedBlock{
				// A set, not a list: the gateway matches entries by computed
				// specificity, so their order carries no meaning.
				CustomType: fwtypes.NewSetNestedObjectTypeOf[limitEntryModel](ctx),
				Validators: []validator.Set{
					setvalidator.IsRequired(),
					setvalidator.SizeBetween(1, 1000),
					entriesMatchDimensionKeysValidator{},
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"dimensions": schema.MapAttribute{
							CustomType: fwtypes.MapOfStringType,
							Required:   true,
							Validators: []validator.Map{
								mapvalidator.SizeBetween(1, 10),
								mapvalidator.KeysAre(dimensionKeyValidators()...),
								mapvalidator.ValueStringsAre(stringvalidator.LengthBetween(1, 256)),
							},
						},
					},
					Blocks: map[string]schema.Block{
						"connections": rateConfigBlock(ctx, awstypes.PeriodSecond),
						"requests":    rateConfigBlock(ctx),
						"tokens":      rateConfigBlock(ctx, awstypes.PeriodMinute),
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

func (r *gatewayRateLimitResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data gatewayRateLimitResourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.Plan.Get(ctx, &data))
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().BedrockAgentCoreClient(ctx)

	gatewayIdentifier := fwflex.StringValueFromFramework(ctx, data.GatewayIdentifier)
	var input bedrockagentcorecontrol.CreateGatewayRateLimitInput
	smerr.AddEnrich(ctx, &response.Diagnostics, fwflex.Expand(ctx, data, &input))
	if response.Diagnostics.HasError() {
		return
	}

	// Serialize modifications of a single gateway. The service holds its lock on
	// the gateway rather than the child, so this shares gateway_rule's key
	// deliberately - rules and rate limits contend with each other.
	mutexKey := fmt.Sprintf("bedrockagentcore-gateway-%s", gatewayIdentifier)
	conns.GlobalMutexKV.Lock(mutexKey)
	defer conns.GlobalMutexKV.Unlock(mutexKey)

	// Only ConflictException is retried: it marks transient gateway-level lock
	// contention. A duplicate dimension_keys set arrives as a terminal
	// ValidationException (verified against the live API), so it must fail
	// fast rather than retry.
	out, err := tfresource.RetryWhenIsA[*bedrockagentcorecontrol.CreateGatewayRateLimitOutput, *awstypes.ConflictException](ctx, r.CreateTimeout(ctx, data.Timeouts), func(ctx context.Context) (*bedrockagentcorecontrol.CreateGatewayRateLimitOutput, error) {
		return conn.CreateGatewayRateLimit(ctx, &input)
	})
	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, gatewayIdentifier)
		return
	}

	rateLimitID := aws.ToString(out.RateLimitId)

	rateLimit, err := waitGatewayRateLimitCreated(ctx, conn, gatewayIdentifier, rateLimitID, r.CreateTimeout(ctx, data.Timeouts))
	if err != nil {
		// Taint the resource so a follow-up plan can reconcile it.
		response.State.SetAttribute(ctx, path.Root("gateway_identifier"), gatewayIdentifier)
		response.State.SetAttribute(ctx, path.Root("rate_limit_id"), rateLimitID)
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, rateLimitID)
		return
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, r.flatten(ctx, rateLimit, &data))
	if response.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, response.State.Set(ctx, &data))
}

func (r *gatewayRateLimitResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data gatewayRateLimitResourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.State.Get(ctx, &data))
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().BedrockAgentCoreClient(ctx)

	gatewayIdentifier, rateLimitID := fwflex.StringValueFromFramework(ctx, data.GatewayIdentifier), fwflex.StringValueFromFramework(ctx, data.RateLimitID)
	// Deleting a gateway cascade-deletes its rate limits, and the API then
	// reports the *gateway* as missing. Both cases arrive here as NotFound, and
	// both mean the same thing: the resource is gone.
	out, err := findGatewayRateLimitByTwoPartKey(ctx, conn, gatewayIdentifier, rateLimitID)
	if retry.NotFound(err) {
		smerr.AddOne(ctx, &response.Diagnostics, fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, rateLimitID)
		return
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, r.flatten(ctx, out, &data))
	if response.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, response.State.Set(ctx, &data))
}

func (r *gatewayRateLimitResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var plan, state gatewayRateLimitResourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.Plan.Get(ctx, &plan))
	smerr.AddEnrich(ctx, &response.Diagnostics, request.State.Get(ctx, &state))
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().BedrockAgentCoreClient(ctx)

	diff, d := fwflex.Diff(ctx, plan, state)
	smerr.AddEnrich(ctx, &response.Diagnostics, d)
	if response.Diagnostics.HasError() {
		return
	}

	if diff.HasChanges() {
		gatewayIdentifier, rateLimitID := fwflex.StringValueFromFramework(ctx, plan.GatewayIdentifier), fwflex.StringValueFromFramework(ctx, plan.RateLimitID)
		var input bedrockagentcorecontrol.UpdateGatewayRateLimitInput
		smerr.AddEnrich(ctx, &response.Diagnostics, fwflex.Expand(ctx, plan, &input))
		if response.Diagnostics.HasError() {
			return
		}
		// A nil description is omitted from the request, which this API treats
		// as "clear it" - exactly the behaviour we want when the practitioner
		// removes the argument. No empty-value coercion needed here.

		mutexKey := fmt.Sprintf("bedrockagentcore-gateway-%s", gatewayIdentifier)
		conns.GlobalMutexKV.Lock(mutexKey)
		defer conns.GlobalMutexKV.Unlock(mutexKey)

		_, err := tfresource.RetryWhenIsA[*bedrockagentcorecontrol.UpdateGatewayRateLimitOutput, *awstypes.ConflictException](ctx, r.UpdateTimeout(ctx, plan.Timeouts), func(ctx context.Context) (*bedrockagentcorecontrol.UpdateGatewayRateLimitOutput, error) {
			return conn.UpdateGatewayRateLimit(ctx, &input)
		})
		if err != nil {
			smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, rateLimitID)
			return
		}

		rateLimit, err := waitGatewayRateLimitUpdated(ctx, conn, gatewayIdentifier, rateLimitID, r.UpdateTimeout(ctx, plan.Timeouts))
		if err != nil {
			smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, rateLimitID)
			return
		}

		smerr.AddEnrich(ctx, &response.Diagnostics, r.flatten(ctx, rateLimit, &plan))
		if response.Diagnostics.HasError() {
			return
		}
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, response.State.Set(ctx, &plan))
}

func (r *gatewayRateLimitResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data gatewayRateLimitResourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.State.Get(ctx, &data))
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().BedrockAgentCoreClient(ctx)

	gatewayIdentifier, rateLimitID := fwflex.StringValueFromFramework(ctx, data.GatewayIdentifier), fwflex.StringValueFromFramework(ctx, data.RateLimitID)
	input := bedrockagentcorecontrol.DeleteGatewayRateLimitInput{
		GatewayIdentifier: aws.String(gatewayIdentifier),
		RateLimitId:       aws.String(rateLimitID),
	}

	mutexKey := fmt.Sprintf("bedrockagentcore-gateway-%s", gatewayIdentifier)
	conns.GlobalMutexKV.Lock(mutexKey)
	defer conns.GlobalMutexKV.Unlock(mutexKey)

	_, err := tfresource.RetryWhenIsA[*bedrockagentcorecontrol.DeleteGatewayRateLimitOutput, *awstypes.ConflictException](ctx, r.DeleteTimeout(ctx, data.Timeouts), func(ctx context.Context) (*bedrockagentcorecontrol.DeleteGatewayRateLimitOutput, error) {
		return conn.DeleteGatewayRateLimit(ctx, &input)
	})
	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return
	}
	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, rateLimitID)
		return
	}

	if _, err := waitGatewayRateLimitDeleted(ctx, conn, gatewayIdentifier, rateLimitID, r.DeleteTimeout(ctx, data.Timeouts)); err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, rateLimitID)
		return
	}
}

func (r *gatewayRateLimitResource) flatten(ctx context.Context, rateLimit any, data *gatewayRateLimitResourceModel) diag.Diagnostics {
	var diags diag.Diagnostics
	diags.Append(fwflex.Flatten(ctx, rateLimit, data)...)
	return diags
}

type gatewayRateLimitImportID struct{}

func (gatewayRateLimitImportID) Parse(id string) (string, map[string]any, error) {
	const gatewayRateLimitIDParts = 2

	// allowEmptyPart=false: a blank gateway identifier or rate limit id is never
	// valid, and letting one through produces a confusing not-found later.
	parts, err := intflex.ExpandResourceId(id, gatewayRateLimitIDParts, false)
	if err != nil {
		return "", nil, err
	}

	result := map[string]any{
		"gateway_identifier": parts[0],
		"rate_limit_id":      parts[1],
	}

	return id, result, nil
}

var (
	_ inttypes.ImportIDParser = gatewayRateLimitImportID{}
)

// Status has no *_FAILED members, so no waiter gets a failure target: a stuck
// operation can only time out.

func waitGatewayRateLimitCreated(ctx context.Context, conn *bedrockagentcorecontrol.Client, gatewayIdentifier, rateLimitID string, timeout time.Duration) (*bedrockagentcorecontrol.GetGatewayRateLimitOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   enum.Slice(awstypes.GatewayRateLimitStatusCreating),
		Target:                    enum.Slice(awstypes.GatewayRateLimitStatusActive),
		Refresh:                   statusGatewayRateLimit(conn, gatewayIdentifier, rateLimitID),
		Timeout:                   timeout,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*bedrockagentcorecontrol.GetGatewayRateLimitOutput); ok {
		return out, smarterr.NewError(err)
	}
	return nil, smarterr.NewError(err)
}

func waitGatewayRateLimitUpdated(ctx context.Context, conn *bedrockagentcorecontrol.Client, gatewayIdentifier, rateLimitID string, timeout time.Duration) (*bedrockagentcorecontrol.GetGatewayRateLimitOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   enum.Slice(awstypes.GatewayRateLimitStatusUpdating),
		Target:                    enum.Slice(awstypes.GatewayRateLimitStatusActive),
		Refresh:                   statusGatewayRateLimit(conn, gatewayIdentifier, rateLimitID),
		Timeout:                   timeout,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*bedrockagentcorecontrol.GetGatewayRateLimitOutput); ok {
		return out, smarterr.NewError(err)
	}
	return nil, smarterr.NewError(err)
}

func waitGatewayRateLimitDeleted(ctx context.Context, conn *bedrockagentcorecontrol.Client, gatewayIdentifier, rateLimitID string, timeout time.Duration) (*bedrockagentcorecontrol.GetGatewayRateLimitOutput, error) { //nolint:unparam
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.GatewayRateLimitStatusDeleting, awstypes.GatewayRateLimitStatusActive),
		Target:  []string{},
		Refresh: statusGatewayRateLimit(conn, gatewayIdentifier, rateLimitID),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*bedrockagentcorecontrol.GetGatewayRateLimitOutput); ok {
		return out, smarterr.NewError(err)
	}
	return nil, smarterr.NewError(err)
}

func statusGatewayRateLimit(conn *bedrockagentcorecontrol.Client, gatewayIdentifier, rateLimitID string) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		out, err := findGatewayRateLimitByTwoPartKey(ctx, conn, gatewayIdentifier, rateLimitID)
		if retry.NotFound(err) {
			return nil, "", nil
		}
		if err != nil {
			return nil, "", smarterr.NewError(err)
		}
		return out, string(out.Status), nil
	}
}

func findGatewayRateLimitByTwoPartKey(ctx context.Context, conn *bedrockagentcorecontrol.Client, gatewayIdentifier, rateLimitID string) (*bedrockagentcorecontrol.GetGatewayRateLimitOutput, error) {
	input := bedrockagentcorecontrol.GetGatewayRateLimitInput{
		GatewayIdentifier: aws.String(gatewayIdentifier),
		RateLimitId:       aws.String(rateLimitID),
	}

	out, err := conn.GetGatewayRateLimit(ctx, &input)
	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, smarterr.NewError(&retry.NotFoundError{LastError: err})
	}
	if err != nil {
		return nil, smarterr.NewError(err)
	}
	if out == nil {
		return nil, smarterr.NewError(tfresource.NewEmptyResultError())
	}
	return out, nil
}

// entriesMatchDimensionKeysValidator enforces the two rules that the type system
// cannot carry: every entry's dimensions map must be keyed by exactly the
// parent's dimension_keys, and a "*" wildcard may only appear in trailing
// positions judged against dimension_keys ordering.
//
// The server rejects violations with a ValidationException at apply; both
// operands are usually known at plan time, so checking here is cheap and
// catches the mistake earlier.
type entriesMatchDimensionKeysValidator struct{}

func (v entriesMatchDimensionKeysValidator) Description(ctx context.Context) string {
	return "entry dimensions must match dimension_keys, with wildcards only in trailing positions"
}

func (v entriesMatchDimensionKeysValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v entriesMatchDimensionKeysValidator) ValidateSet(ctx context.Context, request validator.SetRequest, response *validator.SetResponse) {
	if request.ConfigValue.IsNull() || request.ConfigValue.IsUnknown() {
		return
	}

	var dimensionKeys []string
	diags := request.Config.GetAttribute(ctx, path.Root("dimension_keys"), &dimensionKeys)
	// An unknown or unresolvable dimension_keys leaves nothing to check
	// against; let the server have the final word in that case.
	if diags.HasError() || len(dimensionKeys) == 0 {
		return
	}

	for _, element := range request.ConfigValue.Elements() {
		object, ok := element.(types.Object)
		if !ok || object.IsNull() || object.IsUnknown() {
			continue
		}

		// Every entry needs at least one rate config. The API rejects an entry
		// without one; mirror its wording.
		if !hasAnyRateConfig(object) {
			response.Diagnostics.AddAttributeError(request.Path.AtSetValue(element),
				"Missing Rate Configuration",
				"Each entry must have at least one rate config: \"requests\", \"tokens\", or \"connections\".")
		}

		dimensions, ok := object.Attributes()["dimensions"].(types.Map)
		if !ok || dimensions.IsNull() || dimensions.IsUnknown() {
			continue
		}

		values, ok := dimensionValues(dimensions.Elements())
		if !ok {
			continue // an unknown value; defer to the server
		}

		entryPath := request.Path.AtSetValue(element).AtName("dimensions")
		for _, violation := range validateEntryDimensions(dimensionKeys, values) {
			response.Diagnostics.AddAttributeError(entryPath, violation.summary, violation.detail)
		}
	}
}

// hasAnyRateConfig reports whether an entry sets at least one of the three
// metric blocks. An unknown value counts as set: the server gets the last word.
func hasAnyRateConfig(entry types.Object) bool {
	for _, name := range []string{"requests", "tokens", "connections"} {
		value, ok := entry.Attributes()[name].(types.List)
		if !ok {
			continue
		}
		if value.IsUnknown() {
			return true
		}
		if !value.IsNull() && len(value.Elements()) > 0 {
			return true
		}
	}
	return false
}

type dimensionViolation struct {
	summary string
	detail  string
}

// validateEntryDimensions holds the whole rule, free of framework types so it
// can be unit tested directly. Returns one violation per problem found.
func validateEntryDimensions(dimensionKeys []string, values map[string]string) []dimensionViolation {
	var violations []dimensionViolation

	// Rule 1: the key sets must match exactly. The server reports the missing
	// key first, so mirror that ordering.
	for _, key := range dimensionKeys {
		if _, found := values[key]; !found {
			violations = append(violations, dimensionViolation{
				summary: "Missing Dimension Key",
				detail:  fmt.Sprintf("dimensions is missing required key %q from dimension_keys.", key),
			})
		}
	}
	for _, key := range slices.Sorted(maps.Keys(values)) {
		if !slices.Contains(dimensionKeys, key) {
			violations = append(violations, dimensionViolation{
				summary: "Unexpected Dimension Key",
				detail:  fmt.Sprintf("dimensions contains key %q, which is not in dimension_keys %v.", key, dimensionKeys),
			})
		}
	}
	if len(violations) > 0 {
		// The wildcard rule is positional, so it is meaningless until the keys
		// line up.
		return violations
	}

	// Rule 2: once a position holds "*", every later position must too. The
	// gateway falls back by dropping one position at a time from the right, so
	// a non-trailing wildcard would be unreachable.
	wildcardAt := -1
	for position, key := range dimensionKeys {
		if values[key] == dimensionWildcard {
			if wildcardAt < 0 {
				wildcardAt = position
			}
			continue
		}
		if wildcardAt >= 0 {
			violations = append(violations, dimensionViolation{
				summary: "Non-Trailing Wildcard",
				detail: fmt.Sprintf("Wildcard %q may only appear at trailing positions. Found non-wildcard value for %q at position %d after a wildcard at position %d.",
					dimensionWildcard, key, position+1, wildcardAt+1),
			})
			break
		}
	}

	return violations
}

// dimensionValues converts a map's elements to plain strings, reporting false
// if any value is unknown.
func dimensionValues(elements map[string]attr.Value) (map[string]string, bool) {
	values := make(map[string]string, len(elements))
	for key, element := range elements {
		value, ok := element.(types.String)
		if !ok || value.IsUnknown() {
			return nil, false
		}
		values[key] = value.ValueString()
	}
	return values, true
}

var (
	_ validator.Set = entriesMatchDimensionKeysValidator{}
)

// Models.

type gatewayRateLimitResourceModel struct {
	framework.WithRegionModel
	Description       types.String                                    `tfsdk:"description"`
	DimensionKeys     fwtypes.ListOfString                            `tfsdk:"dimension_keys"`
	Entries           fwtypes.SetNestedObjectValueOf[limitEntryModel] `tfsdk:"entries"`
	GatewayIdentifier types.String                                    `tfsdk:"gateway_identifier"`
	RateLimitID       types.String                                    `tfsdk:"rate_limit_id"`
	Timeouts          timeouts.Value                                  `tfsdk:"timeouts"`
}

type limitEntryModel struct {
	Connections fwtypes.ListNestedObjectValueOf[rateConfigModel] `tfsdk:"connections"`
	Dimensions  fwtypes.MapOfString                              `tfsdk:"dimensions"`
	Requests    fwtypes.ListNestedObjectValueOf[rateConfigModel] `tfsdk:"requests"`
	Tokens      fwtypes.ListNestedObjectValueOf[rateConfigModel] `tfsdk:"tokens"`
}

type rateConfigModel struct {
	Period fwtypes.StringEnum[awstypes.Period] `tfsdk:"period"`
	Rate   types.Float64                       `tfsdk:"rate"`
}
