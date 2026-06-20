// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package resiliencehubv2

import (
	"context"
	"errors"

	"github.com/YakDriver/smarterr"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/resiliencehubv2"
	awstypes "github.com/aws/aws-sdk-go-v2/service/resiliencehubv2/types"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	fwschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	intflex "github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

const assertionImportIDPartCount = 2

// @FrameworkResource("aws_resiliencehubv2_assertion", name="Assertion")
// @IdentityAttribute("service_arn")
// @IdentityAttribute("assertion_id")
// @ImportIDHandler("assertionImportID", setIDAttribute=true)
// @Testing(existsType="github.com/aws/aws-sdk-go-v2/service/resiliencehubv2/types;awstypes;awstypes.Assertion")
// @Testing(hasNoPreExistingResource=true)
func newResourceAssertion(context.Context) (resource.ResourceWithConfigure, error) {
	return &resourceAssertion{}, nil
}

type resourceAssertion struct {
	framework.ResourceWithModel[resourceAssertionModel]
	framework.WithImportByIdentity
}

func (r *resourceAssertion) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = fwschema.Schema{
		Attributes: map[string]fwschema.Attribute{
			names.AttrID: fwschema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"service_arn": fwschema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"text": fwschema.StringAttribute{
				Required: true,
			},
			"assertion_id": fwschema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *resourceAssertion) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan resourceAssertionModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Plan.Get(ctx, &plan))
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().ResilienceHubV2Client(ctx)

	var input resiliencehubv2.CreateAssertionInput
	smerr.AddEnrich(ctx, &resp.Diagnostics, flex.Expand(ctx, plan, &input))
	if resp.Diagnostics.HasError() {
		return
	}

	output, err := conn.CreateAssertion(ctx, &input)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err)
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, flex.Flatten(ctx, output.Assertion, &plan))
	if resp.Diagnostics.HasError() {
		return
	}

	plan.AssertionId = types.StringPointerValue(output.Assertion.AssertionId)
	plan.ServiceArn = types.StringPointerValue(output.Assertion.ServiceArn)
	plan.ID = types.StringValue(plan.ServiceArn.ValueString() + "," + plan.AssertionId.ValueString())

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, plan))
}

func (r *resourceAssertion) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state resourceAssertionModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().ResilienceHubV2Client(ctx)

	assertion, err := findAssertionByID(ctx, conn, state.ServiceArn.ValueString(), state.AssertionId.ValueString())
	if retry.NotFound(err) {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, state.ID.ValueString())
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, r.flatten(ctx, assertion, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, state))
}

func (r *resourceAssertion) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state resourceAssertionModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Plan.Get(ctx, &plan))
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().ResilienceHubV2Client(ctx)

	input := resiliencehubv2.UpdateAssertionInput{
		ServiceArn:  state.ServiceArn.ValueStringPointer(),
		AssertionId: state.AssertionId.ValueStringPointer(),
		Text:        plan.Text.ValueStringPointer(),
	}

	output, err := conn.UpdateAssertion(ctx, &input)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, state.ID.ValueString())
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, flex.Flatten(ctx, output.Assertion, &plan))
	if resp.Diagnostics.HasError() {
		return
	}

	plan.ServiceArn = state.ServiceArn
	plan.AssertionId = state.AssertionId
	plan.ID = state.ID

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, plan))
}

func (r *resourceAssertion) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state resourceAssertionModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().ResilienceHubV2Client(ctx)

	input := resiliencehubv2.DeleteAssertionInput{
		ServiceArn:  state.ServiceArn.ValueStringPointer(),
		AssertionId: state.AssertionId.ValueStringPointer(),
	}
	_, err := conn.DeleteAssertion(ctx, &input)
	if err != nil {
		var nfe *awstypes.ResourceNotFoundException
		if errors.As(err, &nfe) {
			return
		}
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, state.ID.ValueString())
	}
}

type assertionImportID struct{}

func (assertionImportID) Parse(id string) (string, map[string]any, error) {
	parts, err := intflex.ExpandResourceId(id, assertionImportIDPartCount, false)
	if err != nil {
		return "", nil, err
	}

	result := map[string]any{
		"service_arn":  parts[0],
		"assertion_id": parts[1],
	}

	return id, result, nil
}

func (assertionImportID) Create(ctx context.Context, state tfsdk.State) string {
	var serviceArn, assertionID types.String
	state.GetAttribute(ctx, path.Root("service_arn"), &serviceArn)
	state.GetAttribute(ctx, path.Root("assertion_id"), &assertionID)

	return serviceArn.ValueString() + "," + assertionID.ValueString()
}

func (r *resourceAssertion) flatten(ctx context.Context, assertion *awstypes.Assertion, data *resourceAssertionModel) diag.Diagnostics {
	var diags diag.Diagnostics

	diags.Append(flex.Flatten(ctx, assertion, data)...)
	if diags.HasError() {
		return diags
	}

	data.ID = types.StringValue(data.ServiceArn.ValueString() + "," + data.AssertionId.ValueString())

	return diags
}

func findAssertionByID(ctx context.Context, conn *resiliencehubv2.Client, serviceArn, assertionId string) (*awstypes.Assertion, error) {
	input := resiliencehubv2.ListAssertionsInput{
		ServiceArn: aws.String(serviceArn),
	}
	output, err := conn.ListAssertions(ctx, &input)
	if err != nil {
		var nfe *awstypes.ResourceNotFoundException
		if errors.As(err, &nfe) {
			return nil, smarterr.NewError(&retry.NotFoundError{LastError: err})
		}
		return nil, smarterr.NewError(err)
	}

	for _, a := range output.Assertions {
		if aws.ToString(a.AssertionId) == assertionId {
			return &a, nil
		}
	}

	return nil, smarterr.NewError(tfresource.NewEmptyResultError())
}

type resourceAssertionModel struct {
	framework.WithRegionModel
	AssertionId types.String `tfsdk:"assertion_id"`
	ID          types.String `tfsdk:"id"`
	ServiceArn  types.String `tfsdk:"service_arn"`
	Text        types.String `tfsdk:"text"`
}
