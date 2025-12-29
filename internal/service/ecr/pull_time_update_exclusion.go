// Copyright IBM Corp. 2014, 2025
// SPDX-License-Identifier: MPL-2.0

package ecr

import (
	"context"
	"fmt"
	"slices"

	"github.com/aws/aws-sdk-go-v2/service/ecr"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ecr/types"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwvalidators "github.com/hashicorp/terraform-provider-aws/internal/framework/validators"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
	sweepfw "github.com/hashicorp/terraform-provider-aws/internal/sweep/framework"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_ecr_pull_time_update_exclusion", name="Pull Time Update Exclusion")
func newResourcePullTimeUpdateExclusion(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &resourcePullTimeUpdateExclusion{}
	return r, nil
}

const (
	ResNamePullTimeUpdateExclusion = "Pull Time Update Exclusion"
)

type resourcePullTimeUpdateExclusion struct {
	framework.ResourceWithModel[resourcePullTimeUpdateExclusionModel]
	framework.WithNoUpdate
	framework.WithImportByID
}

func (r *resourcePullTimeUpdateExclusion) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrID: framework.IDAttribute(),
			"principal_arn": schema.StringAttribute{
				Required: true,
				Validators: []validator.String{
					fwvalidators.ARN(),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
	}
}

func (r *resourcePullTimeUpdateExclusion) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := r.Meta().ECRClient(ctx)

	var plan resourcePullTimeUpdateExclusionModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	input := &ecr.RegisterPullTimeUpdateExclusionInput{
		PrincipalArn: plan.PrincipalArn.ValueStringPointer(),
	}

	_, err := conn.RegisterPullTimeUpdateExclusion(ctx, input)
	if err != nil {
		resp.Diagnostics.AddError(
			fmt.Sprintf("creating ECR Pull Time Update Exclusion (%s)", plan.PrincipalArn.ValueString()),
			err.Error(),
		)
		return
	}

	plan.ID = plan.PrincipalArn

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *resourcePullTimeUpdateExclusion) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().ECRClient(ctx)

	var state resourcePullTimeUpdateExclusionModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	principalArn := state.ID.ValueString()
	found, err := findPullTimeUpdateExclusionByPrincipalARN(ctx, conn, principalArn)
	if retry.NotFound(err) {
		resp.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError(
			fmt.Sprintf("reading ECR Pull Time Update Exclusion (%s)", principalArn),
			err.Error(),
		)
		return
	}

	if !found {
		resp.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(
			fmt.Errorf("ECR Pull Time Update Exclusion (%s) not found", principalArn),
		))
		resp.State.RemoveResource(ctx)
		return
	}

	// Set the principal_arn from the ID
	state.PrincipalArn = state.ID

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *resourcePullTimeUpdateExclusion) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := r.Meta().ECRClient(ctx)

	var state resourcePullTimeUpdateExclusionModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	input := &ecr.DeregisterPullTimeUpdateExclusionInput{
		PrincipalArn: state.ID.ValueStringPointer(),
	}

	_, err := conn.DeregisterPullTimeUpdateExclusion(ctx, input)
	if err != nil {
		// If the exclusion doesn't exist, that's fine - it's already "deleted"
		if errs.IsA[*awstypes.ExclusionNotFoundException](err) {
			return
		}
		resp.Diagnostics.AddError(
			fmt.Sprintf("deleting ECR Pull Time Update Exclusion (%s)", state.ID.ValueString()),
			err.Error(),
		)
		return
	}
}

func findPullTimeUpdateExclusionByPrincipalARN(ctx context.Context, conn *ecr.Client, principalArn string) (bool, error) {
	input := &ecr.ListPullTimeUpdateExclusionsInput{}

	output, err := conn.ListPullTimeUpdateExclusions(ctx, input)
	if err != nil {
		return false, err
	}

	if slices.Contains(output.PullTimeUpdateExclusions, principalArn) {
		return true, nil
	}

	return false, &retry.NotFoundError{
		LastError: fmt.Errorf("ECR Pull Time Update Exclusion with principal ARN %s not found", principalArn),
	}
}

type resourcePullTimeUpdateExclusionModel struct {
	framework.WithRegionModel
	ID           types.String `tfsdk:"id"`
	PrincipalArn types.String `tfsdk:"principal_arn"`
}

func sweepPullTimeUpdateExclusions(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	conn := client.ECRClient(ctx)
	input := &ecr.ListPullTimeUpdateExclusionsInput{}
	var sweepResources []sweep.Sweepable

	output, err := conn.ListPullTimeUpdateExclusions(ctx, input)
	if err != nil {
		return nil, err
	}

	for _, exclusionArn := range output.PullTimeUpdateExclusions {
		sweepResources = append(sweepResources, sweepfw.NewSweepResource(newResourcePullTimeUpdateExclusion, client,
			sweepfw.NewAttribute(names.AttrID, exclusionArn),
		))
	}

	return sweepResources, nil
}
