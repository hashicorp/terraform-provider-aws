// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

package bedrock

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/bedrock"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
)

// @FrameworkResource("aws_bedrock_use_case_for_model_access", name="Use Case For Model Access")
// @Region(global=true)
// @SingletonIdentity
// @Testing(hasNoPreExistingResource=true)
// @Testing(generator=false)
// @Testing(importStateIdFunc=importStateIDAccountID, importStateIdAttribute="form_data")
// @Testing(preCheck="testAccPreCheckFoundationModelUseCase")
// @Testing(checkDestroyNoop=true)
// @Testing(serialize=true)
func newUseCaseForModelAccessResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &useCaseForModelAccessResource{}

	return r, nil
}

const (
	ResNameUseCaseForModelAccess = "Use Case For Model Access"
)

type useCaseForModelAccessResource struct {
	framework.ResourceWithModel[useCaseForModelAccessResourceModel]
	framework.WithNoOpDelete
	framework.WithNoUpdate
	framework.WithImportByIdentity
}

func (r *useCaseForModelAccessResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"form_data": schema.StringAttribute{
				Required:   true,
				CustomType: jsontypes.NormalizedType{},
			},
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

	// Validate if exist, as the resource cannot be updated or deleted.
	// If exists, import if the form data is the same, otherwise return an error as the form data cannot be updated.
	in := bedrock.GetUseCaseForModelAccessInput{}
	out, err := conn.GetUseCaseForModelAccess(ctx, &in)
	if err == nil {
		v, diags := flattenFormData(ctx, out.FormData)
		smerr.AddEnrich(ctx, &resp.Diagnostics, diags)
		if resp.Diagnostics.HasError() {
			return
		}

		equal, diags := plan.FormData.StringSemanticEquals(ctx, jsontypes.NewNormalizedValue(v.ValueString()))
		smerr.AddEnrich(ctx, &resp.Diagnostics, diags)
		if resp.Diagnostics.HasError() {
			return
		}

		if !equal {
			smerr.AddError(ctx, &resp.Diagnostics, fmt.Errorf("resource already exists with different form data. Form data cannot be updated, please update the form data from %s to %s and create/import the resource", plan.FormData.ValueString(), v.ValueString()))
			return
		}

		plan.FormData = jsontypes.NewNormalizedValue(plan.FormData.ValueString())

		smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, plan))

		return
	}

	var input bedrock.PutUseCaseForModelAccessInput
	smerr.AddEnrich(ctx, &resp.Diagnostics, flex.Expand(ctx, plan, &input, flex.WithFieldNamePrefix("UseCaseForModelAccess")))
	if resp.Diagnostics.HasError() {
		return
	}

	input.FormData = []byte(plan.FormData.ValueString())
	_, err = conn.PutUseCaseForModelAccess(ctx, &input)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err)
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
		smerr.AddError(ctx, &resp.Diagnostics, err)
		return
	}

	v, diags := flattenFormData(ctx, out.FormData)
	smerr.AddEnrich(ctx, &resp.Diagnostics, diags)
	if resp.Diagnostics.HasError() {
		return
	}

	state.FormData = jsontypes.NewNormalizedValue(v.ValueString())

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &state))
}

func flattenFormData(ctx context.Context, formData []byte) (types.String, diag.Diagnostics) { // nosemgrep:ci.semgrep.framework.manual-flattener-functions
	var diags diag.Diagnostics

	// FormData is base64encoded in the AWS API
	v, err := inttypes.Base64Decode(string(formData))
	if err != nil {
		diags.AddError("base64 decoding form data", err.Error())
		return types.StringNull(), diags
	}

	return flex.StringValueToFramework(ctx, v), diags
}

func (r *useCaseForModelAccessResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	r.WithImportByIdentity.ImportState(ctx, req, resp)

	// Touch a value to bypass a Framework check
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("form_data"), types.StringUnknown())...)
}

type useCaseForModelAccessResourceModel struct {
	FormData jsontypes.Normalized `tfsdk:"form_data" autoflex:"-"`
}
