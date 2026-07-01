// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

package pinpointsmsvoicev2

import (
	"context"
	"fmt"
	"time"

	"github.com/YakDriver/smarterr"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/pinpointsmsvoicev2"
	awstypes "github.com/aws/aws-sdk-go-v2/service/pinpointsmsvoicev2/types"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

const (
	propagationTimeout = 2 * time.Minute
)

// @FrameworkResource("aws_pinpointsmsvoicev2_resource_policy", name="Resource Policy")
// @ArnIdentity("resource_arn")
// @Testing(hasNoPreExistingResource=true)
// @Testing(existsType="github.com/aws/aws-sdk-go-v2/service/pinpointsmsvoicev2;pinpointsmsvoicev2.GetResourcePolicyOutput")
// @Testing(generator=false)
// @Testing(tagsTest=false)
// @Testing(importIgnore="policy")
func newResourcePolicyResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	return &resourcePolicyResource{}, nil
}

type resourcePolicyResource struct {
	framework.ResourceWithModel[resourcePolicyResourceModel]
	framework.WithImportByIdentity
}

func (r *resourcePolicyResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrPolicy: schema.StringAttribute{
				CustomType: fwtypes.IAMPolicyType,
				Required:   true,
			},
			names.AttrResourceARN: schema.StringAttribute{
				CustomType: fwtypes.ARNType,
				Required:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
	}
}

func (r *resourcePolicyResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := r.Meta().PinpointSMSVoiceV2Client(ctx)

	var plan resourcePolicyResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Plan.Get(ctx, &plan))
	if resp.Diagnostics.HasError() {
		return
	}

	var input pinpointsmsvoicev2.PutResourcePolicyInput
	smerr.AddEnrich(ctx, &resp.Diagnostics, fwflex.Expand(ctx, plan, &input))
	if resp.Diagnostics.HasError() {
		return
	}

	if _, err := conn.PutResourcePolicy(ctx, &input); err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, plan.ResourceARN.ValueString())
		return
	}

	// Post-Put readback tolerates AWS eventual-consistency propagation and
	// captures the canonical AWS form of the policy for state.
	output, err := tfresource.RetryWhenNotFound(ctx, propagationTimeout, func(ctx context.Context) (*pinpointsmsvoicev2.GetResourcePolicyOutput, error) {
		return findResourcePolicyByARN(ctx, conn, plan.ResourceARN.ValueString())
	})
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, plan.ResourceARN.ValueString())
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, fwflex.Flatten(ctx, output, &plan))
	if resp.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &plan))
}

func (r *resourcePolicyResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().PinpointSMSVoiceV2Client(ctx)

	var state resourcePolicyResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	output, err := findResourcePolicyByARN(ctx, conn, state.ResourceARN.ValueString())
	if retry.NotFound(err) {
		resp.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, state.ResourceARN.ValueString())
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, fwflex.Flatten(ctx, output, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &state))
}

func (r *resourcePolicyResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	conn := r.Meta().PinpointSMSVoiceV2Client(ctx)

	var plan resourcePolicyResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Plan.Get(ctx, &plan))
	if resp.Diagnostics.HasError() {
		return
	}

	var input pinpointsmsvoicev2.PutResourcePolicyInput
	smerr.AddEnrich(ctx, &resp.Diagnostics, fwflex.Expand(ctx, plan, &input))
	if resp.Diagnostics.HasError() {
		return
	}

	output, err := conn.PutResourcePolicy(ctx, &input)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, plan.ResourceARN.ValueString())
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, fwflex.Flatten(ctx, output, &plan))
	if resp.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &plan))
}

func (r *resourcePolicyResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := r.Meta().PinpointSMSVoiceV2Client(ctx)

	var state resourcePolicyResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	input := pinpointsmsvoicev2.DeleteResourcePolicyInput{
		ResourceArn: state.ResourceARN.ValueStringPointer(),
	}
	_, err := conn.DeleteResourcePolicy(ctx, &input)
	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return
	}
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, state.ResourceARN.ValueString())
		return
	}

	// Post-Delete readback waits for AWS eventual-consistency propagation
	// so a subsequent refresh does not observe the policy still attached.
	if _, err := tfresource.RetryUntilNotFound(ctx, propagationTimeout, func(ctx context.Context) (any, error) {
		return findResourcePolicyByARN(ctx, conn, state.ResourceARN.ValueString())
	}); err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, state.ResourceARN.ValueString())
		return
	}
}

func findResourcePolicyByARN(ctx context.Context, conn *pinpointsmsvoicev2.Client, arn string) (*pinpointsmsvoicev2.GetResourcePolicyOutput, error) {
	input := pinpointsmsvoicev2.GetResourcePolicyInput{
		ResourceArn: aws.String(arn),
	}

	output, err := conn.GetResourcePolicy(ctx, &input)
	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, smarterr.NewError(&retry.NotFoundError{LastError: err})
	}
	if err != nil {
		return nil, smarterr.NewError(err)
	}
	if output == nil || output.Policy == nil {
		return nil, smarterr.NewError(tfresource.NewEmptyResultError())
	}
	// when no user policy is attached — including after a successful
	// DeleteResourcePolicy — AWS returns HTTP 200 with the literal string "{}"
	// in the Policy field. Treat that as NotFound so Read removes the
	// resource from state and the post-Delete readback converges.
	if aws.ToString(output.Policy) == "{}" {
		return nil, smarterr.NewError(&retry.NotFoundError{LastError: fmt.Errorf("no resource policy attached to %s", arn)})
	}

	return output, nil
}

type resourcePolicyResourceModel struct {
	framework.WithRegionModel
	Policy      fwtypes.IAMPolicy `tfsdk:"policy"`
	ResourceARN fwtypes.ARN       `tfsdk:"resource_arn"`
}
