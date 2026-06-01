// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

package bedrock

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/bedrock"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_bedrock_use_case_for_model_access", name="Use Case For Model Access")
// @SingletonIdentity
// @Testing(hasNoPreExistingResource=true)
// @Testing(identityTest=false)
// @NoImport
func newUseCaseForModelAccessResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &useCaseForModelAccessResource{}

	r.SetDefaultCreateTimeout(30 * time.Minute)
	r.SetDefaultUpdateTimeout(30 * time.Minute)
	r.SetDefaultDeleteTimeout(30 * time.Minute)

	return r, nil
}

const (
	ResNameUseCaseForModelAccess = "Use Case For Model Access"
)

type useCaseForModelAccessResource struct {
	framework.ResourceWithModel[useCaseForModelAccessResourceModel]
	framework.WithTimeouts
	framework.WithNoOpDelete
}

func (r *useCaseForModelAccessResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"form_data": schema.StringAttribute{
				Required:   true,
				CustomType: jsontypes.NormalizedType{},
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

func (r *useCaseForModelAccessResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := r.Meta().BedrockClient(ctx)

	var plan useCaseForModelAccessResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Plan.Get(ctx, &plan))
	if resp.Diagnostics.HasError() {
		return
	}

	var input bedrock.PutUseCaseForModelAccessInput
	smerr.AddEnrich(ctx, &resp.Diagnostics, flex.Expand(ctx, plan, &input, flex.WithFieldNamePrefix("UseCaseForModelAccess")))
	if resp.Diagnostics.HasError() {
		return
	}

	input.FormData = []byte(plan.FormData.ValueString())
	_, err := conn.PutUseCaseForModelAccess(ctx, &input)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, plan.Region.String())
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, plan))
}

func (r *useCaseForModelAccessResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().BedrockClient(ctx)

	var state useCaseForModelAccessResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	input := bedrock.GetUseCaseForModelAccessInput{}
	out, err := conn.GetUseCaseForModelAccess(ctx, &input)
	if retry.NotFound(err) || out == nil {
		smerr.AddOne(ctx, &resp.Diagnostics, fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, state.Region.String())
		return
	}

	v, diags := flattenFormData(ctx, aws.String(string(out.FormData)))
	smerr.AddEnrich(ctx, &resp.Diagnostics, diags)
	if resp.Diagnostics.HasError() {
		return
	}

	state.FormData = jsontypes.NewNormalizedValue(v.ValueString())
	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &state))
}

func flattenFormData(ctx context.Context, configs *string) (types.String, diag.Diagnostics) { // nosemgrep:ci.semgrep.framework.manual-flattener-functions
	var diags diag.Diagnostics

	// FormData is base64encoded in the AWS API
	if configs != nil {
		v, err := inttypes.Base64Decode(aws.ToString(configs))
		if err != nil {
			diags.AddError("base64 decoding form data", err.Error())
			return types.StringNull(), diags
		}

		return flex.StringValueToFramework(ctx, v), diags
	}

	return types.StringNull(), diags
}

func (r *useCaseForModelAccessResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	conn := r.Meta().BedrockClient(ctx)

	var plan, state useCaseForModelAccessResourceModel
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
		var input bedrock.PutUseCaseForModelAccessInput
		smerr.AddEnrich(ctx, &resp.Diagnostics, flex.Expand(ctx, plan, &input, flex.WithFieldNamePrefix("UseCaseForModelAccess")))
		if resp.Diagnostics.HasError() {
			return
		}

		input.FormData = []byte(plan.FormData.ValueString())

		_, err := conn.PutUseCaseForModelAccess(ctx, &input)
		if err != nil {
			smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, plan.Region.String())
			return
		}
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &plan))
}

type useCaseForModelAccessResourceModel struct {
	framework.WithRegionModel
	FormData jsontypes.Normalized `tfsdk:"form_data" autoflex:"-"`
	Timeouts timeouts.Value       `tfsdk:"timeouts"`
}
