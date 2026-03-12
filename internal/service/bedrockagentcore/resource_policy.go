// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

package bedrockagentcore

import (
	"context"
	"errors"

	"github.com/YakDriver/smarterr"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/bedrockagentcorecontrol"
	awstypes "github.com/aws/aws-sdk-go-v2/service/bedrockagentcorecontrol/types"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
	sweepfw "github.com/hashicorp/terraform-provider-aws/internal/sweep/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_bedrockagentcore_resource_policy", name="Resource Policy")
// @ArnIdentity("resource_arn")
// @Testing(preCheck="testAccPreCheck")
// @Testing(importIgnore="...;...")
// @Testing(hasNoPreExistingResource=true)
func newResourcePolicyResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &resourcePolicyResource{}
	return r, nil
}

const (
	ResNameResourcePolicy = "Resource Policy"
)

type resourcePolicyResource struct {
	framework.ResourceWithModel[resourcePolicyResourceModel]
	framework.WithImportByIdentity
}

func (r *resourcePolicyResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrID: framework.IDAttribute(),
			names.AttrResourceARN: schema.StringAttribute{
				CustomType: fwtypes.ARNType,
				Required:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			names.AttrPolicy: schema.StringAttribute{
				CustomType: fwtypes.IAMPolicyType,
				Required:   true,
			},
		},
	}
}

func (r *resourcePolicyResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := r.Meta().BedrockAgentCoreClient(ctx)

	var plan resourcePolicyResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Plan.Get(ctx, &plan))
	if resp.Diagnostics.HasError() {
		return
	}

	var input bedrockagentcorecontrol.PutResourcePolicyInput
	smerr.AddEnrich(ctx, &resp.Diagnostics, flex.Expand(ctx, plan, &input, flex.WithFieldNamePrefix("ResourcePolicy")))
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := conn.PutResourcePolicy(ctx, &input)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, plan.ResourceARN.String())
		return
	}
	if out == nil || out.Policy == nil {
		smerr.AddError(ctx, &resp.Diagnostics, errors.New("empty output"), smerr.ID, plan.ResourceARN.String())
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, r.flatten(ctx, out.Policy, &plan))
	if resp.Diagnostics.HasError() {
		return
	}

	plan.setID()

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, plan))
}

func (r *resourcePolicyResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().BedrockAgentCoreClient(ctx)

	var state resourcePolicyResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := findResourcePolicy(ctx, conn, state.ResourceARN.ValueString())
	if retry.NotFound(err) {
		smerr.AddOne(ctx, &resp.Diagnostics, fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, state.ResourceARN.String())
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, r.flatten(ctx, out, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	state.setID()

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &state))
}

func (r *resourcePolicyResource) flatten(_ context.Context, resourcePolicy *string, data *resourcePolicyResourceModel) (diags diag.Diagnostics) {
	if resourcePolicy != nil {
		data.Policy = fwtypes.IAMPolicyValue(aws.ToString(resourcePolicy))
	}
	return diags
}

func (r *resourcePolicyResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	conn := r.Meta().BedrockAgentCoreClient(ctx)

	var plan, state resourcePolicyResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Plan.Get(ctx, &plan))
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	diff, d := flex.Diff(ctx, plan, state)
	smerr.AddEnrich(ctx, &resp.Diagnostics, d)
	if resp.Diagnostics.HasError() {
		return
	}

	if diff.HasChanges() {
		var input bedrockagentcorecontrol.PutResourcePolicyInput
		smerr.AddEnrich(ctx, &resp.Diagnostics, flex.Expand(ctx, plan, &input, flex.WithFieldNamePrefix("ResourcePolicy")))
		if resp.Diagnostics.HasError() {
			return
		}

		out, err := conn.PutResourcePolicy(ctx, &input)
		if err != nil {
			smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, plan.ResourceARN.String())
			return
		}
		if out == nil || out.Policy == nil {
			smerr.AddError(ctx, &resp.Diagnostics, errors.New("empty output"), smerr.ID, plan.ResourceARN.String())
			return
		}

		smerr.AddEnrich(ctx, &resp.Diagnostics, r.flatten(ctx, out.Policy, &plan))
		if resp.Diagnostics.HasError() {
			return
		}
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &plan))
}

func (r *resourcePolicyResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := r.Meta().BedrockAgentCoreClient(ctx)

	var state resourcePolicyResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	input := bedrockagentcorecontrol.DeleteResourcePolicyInput{
		ResourceArn: state.ResourceARN.ValueStringPointer(),
	}

	_, err := conn.DeleteResourcePolicy(ctx, &input)
	if err != nil {
		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return
		}

		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, state.ResourceARN.String())
		return
	}
}

func findResourcePolicy(ctx context.Context, conn *bedrockagentcorecontrol.Client, resourceArn string) (*string, error) {
	input := bedrockagentcorecontrol.GetResourcePolicyInput{
		ResourceArn: aws.String(resourceArn),
	}

	out, err := conn.GetResourcePolicy(ctx, &input)
	if err != nil {
		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return nil, smarterr.NewError(&retry.NotFoundError{
				LastError: err,
			})
		}

		return nil, smarterr.NewError(err)
	}

	if out == nil || out.Policy == nil {
		return nil, smarterr.NewError(tfresource.NewEmptyResultError())
	}

	return out.Policy, nil
}

type resourcePolicyResourceModel struct {
	framework.WithRegionModel
	ID          types.String      `tfsdk:"id"`
	ResourceARN fwtypes.ARN       `tfsdk:"resource_arn"`
	Policy      fwtypes.IAMPolicy `tfsdk:"policy"`
}

func (data *resourcePolicyResourceModel) setID() {
	data.ID = data.ResourceARN.StringValue
}

func sweepResourcePolicies(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	runtimeResourcePolicies, err := sweepResourcePoliciesForAgentRuntimes(ctx, client)
	if err != nil {
		return nil, smarterr.NewError(err)
	}
	gatewayResourcePolicies, err := sweepResourcePoliciesForGateways(ctx, client)
	if err != nil {
		return nil, smarterr.NewError(err)
	}
	runtimeEndpointResourcePolicies, err := sweepResourcePoliciesForAgentRuntimeEndpoints(ctx, client)
	if err != nil {
		return nil, smarterr.NewError(err)
	}
	return append(append(runtimeResourcePolicies, gatewayResourcePolicies...), runtimeEndpointResourcePolicies...), nil
}

func sweepResourcePoliciesForAgentRuntimes(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	var sweepResources []sweep.Sweepable

	input := bedrockagentcorecontrol.ListAgentRuntimesInput{}
	conn := client.BedrockAgentCoreClient(ctx)

	pages := bedrockagentcorecontrol.NewListAgentRuntimesPaginator(conn, &input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)
		if err != nil {
			return nil, smarterr.NewError(err)
		}

		for _, v := range page.AgentRuntimes {
			policy, err := findResourcePolicy(ctx, conn, *v.AgentRuntimeArn)
			if err != nil {
				if retry.NotFound(err) {
					continue
				}
				return nil, smarterr.NewError(err)
			}

			sweepResources = append(sweepResources, sweepfw.NewSweepResource(newResourcePolicyResource, client,
				sweepfw.NewAttribute(names.AttrResourceARN, *v.AgentRuntimeArn),
				sweepfw.NewAttribute(names.AttrPolicy, *policy),
			),
			)
		}
	}

	return sweepResources, nil
}

func sweepResourcePoliciesForAgentRuntimeEndpoints(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	input := bedrockagentcorecontrol.ListAgentRuntimesInput{}
	conn := client.BedrockAgentCoreClient(ctx)
	var sweepResources []sweep.Sweepable

	pages := bedrockagentcorecontrol.NewListAgentRuntimesPaginator(conn, &input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)
		if err != nil {
			return nil, smarterr.NewError(err)
		}

		for _, v := range page.AgentRuntimes {
			agentRuntimeID := aws.ToString(v.AgentRuntimeId)
			input := bedrockagentcorecontrol.ListAgentRuntimeEndpointsInput{
				AgentRuntimeId: aws.String(agentRuntimeID),
			}

			pages := bedrockagentcorecontrol.NewListAgentRuntimeEndpointsPaginator(conn, &input)
			for pages.HasMorePages() {
				page, err := pages.NextPage(ctx)
				if err != nil {
					return nil, smarterr.NewError(err)
				}

				for _, v := range page.RuntimeEndpoints {
					policy, err := findResourcePolicy(ctx, conn, aws.ToString(v.AgentRuntimeEndpointArn))
					if err != nil {
						if retry.NotFound(err) {
							continue
						}
						return nil, smarterr.NewError(err)
					}
					sweepResources = append(sweepResources, sweepfw.NewSweepResource(newResourcePolicyResource, client,
						sweepfw.NewAttribute(names.AttrResourceARN, aws.ToString(v.AgentRuntimeEndpointArn)),
						sweepfw.NewAttribute(names.AttrPolicy, *policy),
					),
					)
				}
			}
		}
	}

	return sweepResources, nil
}

func sweepResourcePoliciesForGateways(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	input := bedrockagentcorecontrol.ListGatewaysInput{}
	conn := client.BedrockAgentCoreClient(ctx)
	var sweepResources []sweep.Sweepable

	pages := bedrockagentcorecontrol.NewListGatewaysPaginator(conn, &input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)
		if err != nil {
			return nil, smarterr.NewError(err)
		}

		for _, v := range page.Items {
			gateway, err := findGatewayByID(ctx, conn, *v.GatewayId)
			if err != nil {
				return nil, smarterr.NewError(err)
			}
			policy, err := findResourcePolicy(ctx, conn, *gateway.GatewayArn)
			if err != nil {
				if retry.NotFound(err) {
					continue
				}
				return nil, smarterr.NewError(err)
			}

			sweepResources = append(sweepResources, sweepfw.NewSweepResource(newResourcePolicyResource, client,
				sweepfw.NewAttribute(names.AttrResourceARN, *gateway.GatewayArn),
				sweepfw.NewAttribute(names.AttrPolicy, *policy),
			),
			)
		}
	}

	return sweepResources, nil
}
