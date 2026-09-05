// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

package dsql

import (
	"context"
	"errors"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dsql"
	awstypes "github.com/aws/aws-sdk-go-v2/service/dsql/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

var clusterPolicyIdentifierRegex = regexache.MustCompile(`^[a-z0-9]{26}$`)

type clusterPolicyResource struct {
	framework.ResourceWithModel[clusterPolicyResourceModel]
	framework.WithTimeouts
	framework.WithImportByIdentity
}

// @FrameworkResource("aws_dsql_cluster_policy", name="Cluster Policy")
// @IdentityAttribute("identifier")
// @Testing(importIgnore="bypass_policy_lockout_safety_check;policy")
// @Testing(hasNoPreExistingResource=true)
// @Testing(generator=false)
// @Testing(importStateIdAttribute="identifier")
func newClusterPolicyResource(context.Context) (resource.ResourceWithConfigure, error) {
	r := &clusterPolicyResource{}
	r.SetDefaultCreateTimeout(1 * time.Minute)
	r.SetDefaultUpdateTimeout(1 * time.Minute)
	r.SetDefaultDeleteTimeout(1 * time.Minute)

	return r, nil
}

func (r *clusterPolicyResource) Schema(ctx context.Context, _ resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"bypass_policy_lockout_safety_check": schema.BoolAttribute{
				Optional: true,
				Computed: true,
				Default:  booldefault.StaticBool(false),
			},
			names.AttrIdentifier: schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.RegexMatches(clusterPolicyIdentifierRegex, "must be a valid Aurora DSQL cluster identifier"),
				},
			},
			names.AttrPolicy: schema.StringAttribute{
				CustomType: fwtypes.IAMPolicyType,
				Required:   true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 20480),
				},
			},
			"policy_version": schema.StringAttribute{
				Computed: true,
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

func (r *clusterPolicyResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data clusterPolicyResourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.Plan.Get(ctx, &data))
	if response.Diagnostics.HasError() {
		return
	}
	id := data.Identifier.ValueString()

	conn := r.Meta().DSQLClient(ctx)

	var input dsql.PutClusterPolicyInput
	smerr.AddEnrich(ctx, &response.Diagnostics, flex.Expand(ctx, data, &input))
	if response.Diagnostics.HasError() {
		return
	}

	// Additional fields
	input.ClientToken = aws.String(create.UniqueId(ctx))

	output, err := r.putClusterPolicyAndWait(ctx, conn, &input, r.CreateTimeout(ctx, data.Timeouts))
	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, id)
		return
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, flex.Flatten(ctx, output, &data))
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *clusterPolicyResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data clusterPolicyResourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.State.Get(ctx, &data))
	if response.Diagnostics.HasError() {
		return
	}
	id := data.Identifier.ValueString()

	conn := r.Meta().DSQLClient(ctx)
	output, err := findClusterPolicyByID(ctx, conn, id)
	if retry.NotFound(err) {
		response.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, id)
		return
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, r.flatten(ctx, output, &data))
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *clusterPolicyResource) flatten(ctx context.Context, output *dsql.GetClusterPolicyOutput, data *clusterPolicyResourceModel) diag.Diagnostics {
	return flex.Flatten(ctx, output, data)
}

func (r *clusterPolicyResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var state, plan clusterPolicyResourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.Plan.Get(ctx, &plan))
	smerr.AddEnrich(ctx, &response.Diagnostics, request.State.Get(ctx, &state))
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().DSQLClient(ctx)
	id := plan.Identifier.ValueString()

	var input dsql.PutClusterPolicyInput
	smerr.AddEnrich(ctx, &response.Diagnostics, flex.Expand(ctx, plan, &input))
	if response.Diagnostics.HasError() {
		return
	}

	// Additional fields
	input.ClientToken = aws.String(create.UniqueId(ctx))

	output, err := r.putClusterPolicyAndWait(ctx, conn, &input, r.UpdateTimeout(ctx, plan.Timeouts))
	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, id)
		return
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, flex.Flatten(ctx, output, &plan))
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &plan)...)
}

func (r *clusterPolicyResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data clusterPolicyResourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.State.Get(ctx, &data))
	if response.Diagnostics.HasError() {
		return
	}
	id := data.Identifier.ValueString()

	conn := r.Meta().DSQLClient(ctx)
	input := dsql.DeleteClusterPolicyInput{
		Identifier: data.Identifier.ValueStringPointer(),
	}

	_, err := conn.DeleteClusterPolicy(ctx, &input)
	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return
	}

	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, id)
		return
	}

	if _, err = tfresource.RetryUntilNotFound(ctx, r.DeleteTimeout(ctx, data.Timeouts), func(ctx context.Context) (any, error) {
		return findClusterPolicyByID(ctx, conn, id)
	}); err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, id)
		return
	}
}

func (r *clusterPolicyResource) putClusterPolicyAndWait(ctx context.Context, conn *dsql.Client, input *dsql.PutClusterPolicyInput, timeout time.Duration) (*dsql.GetClusterPolicyOutput, error) {
	output, err := conn.PutClusterPolicy(ctx, input)
	if err != nil {
		return nil, err
	}
	if output == nil || output.PolicyVersion == nil {
		return nil, errors.New("empty output")
	}

	id := aws.ToString(input.Identifier)
	policy := aws.ToString(input.Policy)
	policyVersion := aws.ToString(output.PolicyVersion)

	return waitClusterPolicyUpdated(ctx, conn, id, policyVersion, policy, timeout)
}

func findClusterPolicyByID(ctx context.Context, conn *dsql.Client, id string) (*dsql.GetClusterPolicyOutput, error) {
	input := dsql.GetClusterPolicyInput{
		Identifier: aws.String(id),
	}

	output, err := conn.GetClusterPolicy(ctx, &input)
	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError: err,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.Policy == nil || output.PolicyVersion == nil {
		return nil, tfresource.NewEmptyResultError()
	}

	return output, nil
}

func waitClusterPolicyUpdated(ctx context.Context, conn *dsql.Client, id, expectedPolicyVersion, expectedPolicy string, timeout time.Duration) (*dsql.GetClusterPolicyOutput, error) {
	var policyOutput *dsql.GetClusterPolicyOutput

	err := tfresource.WaitUntil(ctx, timeout, func(ctx context.Context) (bool, error) {
		var err error

		policyOutput, err = findClusterPolicyByID(ctx, conn, id)
		if err != nil {
			return false, err
		}

		return policyOutput != nil &&
			aws.ToString(policyOutput.PolicyVersion) == expectedPolicyVersion &&
			verify.PolicyStringsEquivalent(expectedPolicy, aws.ToString(policyOutput.Policy)), nil
	}, tfresource.WaitOpts{
		MinTimeout: 5 * time.Second,
	})

	if err != nil {
		return nil, err
	}

	if policyOutput == nil {
		return nil, tfresource.NewEmptyResultError()
	}

	return policyOutput, nil
}

type clusterPolicyResourceModel struct {
	framework.WithRegionModel
	BypassPolicyLockoutSafetyCheck types.Bool        `tfsdk:"bypass_policy_lockout_safety_check"`
	Identifier                     types.String      `tfsdk:"identifier"`
	Policy                         fwtypes.IAMPolicy `tfsdk:"policy"`
	PolicyVersion                  types.String      `tfsdk:"policy_version"`
	Timeouts                       timeouts.Value    `tfsdk:"timeouts"`
}
