// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

package ecr

import (
	"context"
	"fmt"
	"slices"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ecr"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ecr/types"
	"github.com/hashicorp/terraform-plugin-framework/path"
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
)

// @FrameworkResource("aws_ecr_pull_time_update_exclusion", name="Pull Time Update Exclusion")
func newPullTimeUpdateExclusionResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &pullTimeUpdateExclusionResource{}
	return r, nil
}

type pullTimeUpdateExclusionResource struct {
	framework.ResourceWithModel[pullTimeUpdateExclusionResourceModel]
	framework.WithNoUpdate
}

func (r *pullTimeUpdateExclusionResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"principal_arn": schema.StringAttribute{
				CustomType: fwtypes.ARNType,
				Required:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
	}
}

func (r *pullTimeUpdateExclusionResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan pullTimeUpdateExclusionResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().ECRClient(ctx)

	principalARN := fwflex.StringValueFromFramework(ctx, plan.PrincipalARN)
	input := ecr.RegisterPullTimeUpdateExclusionInput{
		PrincipalArn: aws.String(principalARN),
	}
	_, err := conn.RegisterPullTimeUpdateExclusion(ctx, &input)
	if err != nil {
		resp.Diagnostics.AddError(
			fmt.Sprintf("creating ECR Pull Time Update Exclusion (%s)", principalARN),
			err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *pullTimeUpdateExclusionResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state pullTimeUpdateExclusionResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().ECRClient(ctx)

	principalARN := fwflex.StringValueFromFramework(ctx, state.PrincipalARN)
	err := findPullTimeUpdateExclusionByPrincipalARN(ctx, conn, principalARN)
	if retry.NotFound(err) {
		resp.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError(
			fmt.Sprintf("reading ECR Pull Time Update Exclusion (%s)", principalARN),
			err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *pullTimeUpdateExclusionResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state pullTimeUpdateExclusionResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().ECRClient(ctx)

	principalARN := fwflex.StringValueFromFramework(ctx, state.PrincipalARN)
	input := ecr.DeregisterPullTimeUpdateExclusionInput{
		PrincipalArn: aws.String(principalARN),
	}

	_, err := conn.DeregisterPullTimeUpdateExclusion(ctx, &input)
	if errs.IsA[*awstypes.ExclusionNotFoundException](err) {
		return
	}
	if err != nil {
		resp.Diagnostics.AddError(
			fmt.Sprintf("deleting ECR Pull Time Update Exclusion (%s)", principalARN),
			err.Error(),
		)
		return
	}
}

func (r *pullTimeUpdateExclusionResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("principal_arn"), req, resp)
}

func findPullTimeUpdateExclusionByPrincipalARN(ctx context.Context, conn *ecr.Client, arn string) error {
	var input ecr.ListPullTimeUpdateExclusionsInput
	output, err := findPullTimeUpdateExclusions(ctx, conn, &input)
	if err != nil {
		return err
	}

	if slices.Contains(output, arn) {
		return nil
	}

	return &retry.NotFoundError{}
}

func findPullTimeUpdateExclusions(ctx context.Context, conn *ecr.Client, input *ecr.ListPullTimeUpdateExclusionsInput) ([]string, error) {
	var output []string

	err := listPullTimeUpdateExclusionsPages(ctx, conn, input, func(page *ecr.ListPullTimeUpdateExclusionsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		output = append(output, page.PullTimeUpdateExclusions...)

		return !lastPage
	})

	if err != nil {
		return nil, err
	}

	return output, nil
}

type pullTimeUpdateExclusionResourceModel struct {
	framework.WithRegionModel
	PrincipalARN fwtypes.ARN `tfsdk:"principal_arn"`
}
