// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package resiliencehubv2

import (
	"context"
	"fmt"
	"iter"

	"github.com/YakDriver/smarterr"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/resiliencehubv2"
	awstypes "github.com/aws/aws-sdk-go-v2/service/resiliencehubv2/types"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	fwschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	intflex "github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

// @FrameworkResource("aws_resiliencehubv2_assertion", name="Assertion")
// @IdentityAttribute("service_arn")
// @IdentityAttribute("assertion_id")
// @ImportIDHandler("assertionImportID")
// @Testing(existsType="github.com/aws/aws-sdk-go-v2/service/resiliencehubv2/types;awstypes;awstypes.Assertion")
// @Testing(hasNoPreExistingResource=true)
// @Testing(importStateIdFunc="testAccCheckAssertionImportStateIDFunc")
// @Testing(importStateIdAttribute="assertion_id")
func newAssertionResource(context.Context) (resource.ResourceWithConfigure, error) {
	return &assertionResource{}, nil
}

type assertionResource struct {
	framework.ResourceWithModel[assertionResourceModel]
	framework.WithImportByIdentity
}

func (r *assertionResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = fwschema.Schema{
		Attributes: map[string]fwschema.Attribute{
			"assertion_id": framework.IDAttribute(),
			"service_arn": fwschema.StringAttribute{
				CustomType: fwtypes.ARNType,
				Required:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"text": fwschema.StringAttribute{
				Required: true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 1000),
				},
			},
		},
	}
}

func (r *assertionResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan assertionResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Plan.Get(ctx, &plan))
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().ResilienceHubV2Client(ctx)

	var input resiliencehubv2.CreateAssertionInput
	smerr.AddEnrich(ctx, &resp.Diagnostics, fwflex.Expand(ctx, plan, &input))
	if resp.Diagnostics.HasError() {
		return
	}

	// Additional fields.
	input.ClientToken = aws.String(create.UniqueId(ctx))

	output, err := conn.CreateAssertion(ctx, &input)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err)
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, r.flatten(ctx, output.Assertion, &plan))
	if resp.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, plan))
}

func (r *assertionResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state assertionResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().ResilienceHubV2Client(ctx)

	serviceARN, assertionID := fwflex.StringValueFromFramework(ctx, state.ServiceARN), fwflex.StringValueFromFramework(ctx, state.AssertionID)
	assertion, err := findAssertionByTwoPartKey(ctx, conn, serviceARN, assertionID)
	if retry.NotFound(err) {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, assertionID)
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, r.flatten(ctx, assertion, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, state))
}

func (r *assertionResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state assertionResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Plan.Get(ctx, &plan))
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().ResilienceHubV2Client(ctx)

	assertionID := fwflex.StringValueFromFramework(ctx, plan.AssertionID)
	var input resiliencehubv2.UpdateAssertionInput
	smerr.AddEnrich(ctx, &resp.Diagnostics, fwflex.Expand(ctx, plan, &input))
	if resp.Diagnostics.HasError() {
		return
	}

	_, err := conn.UpdateAssertion(ctx, &input)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, assertionID)
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, plan))
}

func (r *assertionResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state assertionResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().ResilienceHubV2Client(ctx)

	serviceARN, assertionID := fwflex.StringValueFromFramework(ctx, state.ServiceARN), fwflex.StringValueFromFramework(ctx, state.AssertionID)
	input := resiliencehubv2.DeleteAssertionInput{
		AssertionId: aws.String(assertionID),
		ServiceArn:  aws.String(serviceARN),
	}
	_, err := conn.DeleteAssertion(ctx, &input)
	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return
	}
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, assertionID)
	}
}

func (r *assertionResource) flatten(ctx context.Context, assertion *awstypes.Assertion, data *assertionResourceModel) diag.Diagnostics {
	var diags diag.Diagnostics

	diags.Append(fwflex.Flatten(ctx, assertion, data)...)
	if diags.HasError() {
		return diags
	}

	return diags
}

type assertionImportID struct{}

func (assertionImportID) Parse(id string) (string, map[string]any, error) {
	const (
		assertionImportIDPartCount = 2
	)
	parts, err := intflex.ExpandResourceId(id, assertionImportIDPartCount, false)
	if err != nil {
		return "", nil, err
	}

	result := map[string]any{
		"assertion_id": parts[1],
		"service_arn":  parts[0],
	}

	return id, result, nil
}

func findAssertionByTwoPartKey(ctx context.Context, conn *resiliencehubv2.Client, serviceARN, assertionID string) (*awstypes.Assertion, error) {
	input := resiliencehubv2.ListAssertionsInput{
		ServiceArn: aws.String(serviceARN),
	}

	return findAssertion(ctx, conn, &input, tfslices.WithFilter(func(v awstypes.Assertion) bool {
		return aws.ToString(v.AssertionId) == assertionID
	}))
}

func findAssertion(ctx context.Context, conn *resiliencehubv2.Client, input *resiliencehubv2.ListAssertionsInput, optFns ...tfslices.FinderOptionsFunc[awstypes.Assertion]) (*awstypes.Assertion, error) {
	output, err := findAssertions(ctx, conn, input, optFns...)

	if err != nil {
		return nil, smarterr.NewError(err)
	}

	return smarterr.Assert(tfresource.AssertSingleValueResult(output))
}

func findAssertions(ctx context.Context, conn *resiliencehubv2.Client, input *resiliencehubv2.ListAssertionsInput, optFns ...tfslices.FinderOptionsFunc[awstypes.Assertion]) ([]awstypes.Assertion, error) {
	output, err := tfslices.CollectAndConcatWithError(listAssertionPages(ctx, conn, input), optFns...)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, smarterr.NewError(&retry.NotFoundError{
			LastError: err,
		})
	}

	return output, nil
}

func listAssertionPages(ctx context.Context, conn *resiliencehubv2.Client, input *resiliencehubv2.ListAssertionsInput, optFns ...func(*resiliencehubv2.Options)) iter.Seq2[[]awstypes.Assertion, error] {
	return func(yield func([]awstypes.Assertion, error) bool) {
		pages := resiliencehubv2.NewListAssertionsPaginator(conn, input)
		for pages.HasMorePages() {
			page, err := pages.NextPage(ctx, optFns...)
			if err != nil {
				yield(nil, fmt.Errorf("listing Resilience Hub V2 Assertions: %w", err))
				return
			}

			if !yield(page.Assertions, nil) {
				return
			}
		}
	}
}

type assertionResourceModel struct {
	framework.WithRegionModel
	AssertionID types.String `tfsdk:"assertion_id"`
	ServiceARN  fwtypes.ARN  `tfsdk:"service_arn"`
	Text        types.String `tfsdk:"text"`
}
