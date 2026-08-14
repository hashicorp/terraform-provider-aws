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
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	intflex "github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_resiliencehubv2_service_function", name="Service Function")
// @IdentityAttribute("service_arn")
// @IdentityAttribute("service_function_id")
// @ImportIDHandler("serviceFunctionImportID")
// @Testing(existsType="github.com/aws/aws-sdk-go-v2/service/resiliencehubv2/types;awstypes;awstypes.ServiceFunction")
// @Testing(hasNoPreExistingResource=true)
// @Testing(importStateIdFunc="testAccCheckServiceFunctionImportStateIDFunc")
// @Testing(importStateIdAttribute="service_function_id")
func newServiceFunctionResource(context.Context) (resource.ResourceWithConfigure, error) {
	return &serviceFunctionResource{}, nil
}

type serviceFunctionResource struct {
	framework.ResourceWithModel[serviceFunctionResourceModel]
	framework.WithImportByIdentity
}

func (r *serviceFunctionResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = fwschema.Schema{
		Attributes: map[string]fwschema.Attribute{
			"criticality": fwschema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.ServiceFunctionCriticality](),
				Required:   true,
			},
			names.AttrDescription: fwschema.StringAttribute{
				Optional: true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(0, 500),
				},
			},
			names.AttrName: fwschema.StringAttribute{
				Required: true,
				Validators: []validator.String{
					validResourceName,
				},
			},
			"service_arn": fwschema.StringAttribute{
				CustomType: fwtypes.ARNType,
				Required:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"service_function_id": framework.IDAttribute(),
		},
	}
}

func (r *serviceFunctionResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan serviceFunctionResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Plan.Get(ctx, &plan))
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().ResilienceHubV2Client(ctx)

	name := fwflex.StringValueFromFramework(ctx, plan.Name)
	var input resiliencehubv2.CreateServiceFunctionInput
	smerr.AddEnrich(ctx, &resp.Diagnostics, fwflex.Expand(ctx, plan, &input))
	if resp.Diagnostics.HasError() {
		return
	}

	// Additional fields.
	input.ClientToken = aws.String(create.UniqueId(ctx))

	output, err := conn.CreateServiceFunction(ctx, &input)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.Name, name)
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, r.flatten(ctx, output.ServiceFunction, &plan))
	if resp.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, plan))
}

func (r *serviceFunctionResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state serviceFunctionResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().ResilienceHubV2Client(ctx)

	serviceARN, serviceFunctionID := fwflex.StringValueFromFramework(ctx, state.ServiceARN), fwflex.StringValueFromFramework(ctx, state.ServiceFunctionID)
	sf, err := findServiceFunctionByTwoPartKey(ctx, conn, serviceARN, serviceFunctionID)
	if retry.NotFound(err) {
		smerr.AddOne(ctx, &resp.Diagnostics, fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, serviceFunctionID)
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, r.flatten(ctx, sf, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, state))
}

func (r *serviceFunctionResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state serviceFunctionResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Plan.Get(ctx, &plan))
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().ResilienceHubV2Client(ctx)

	serviceFunctionID := fwflex.StringValueFromFramework(ctx, plan.ServiceFunctionID)
	var input resiliencehubv2.UpdateServiceFunctionInput
	smerr.AddEnrich(ctx, &resp.Diagnostics, fwflex.Expand(ctx, plan, &input))
	if resp.Diagnostics.HasError() {
		return
	}

	_, err := conn.UpdateServiceFunction(ctx, &input)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, serviceFunctionID)
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, plan))
}

func (r *serviceFunctionResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state serviceFunctionResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().ResilienceHubV2Client(ctx)

	serviceARN, serviceFunctionID := fwflex.StringValueFromFramework(ctx, state.ServiceARN), fwflex.StringValueFromFramework(ctx, state.ServiceFunctionID)
	input := resiliencehubv2.DeleteServiceFunctionInput{
		ServiceArn:        aws.String(serviceARN),
		ServiceFunctionId: aws.String(serviceFunctionID),
	}
	_, err := conn.DeleteServiceFunction(ctx, &input)
	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return
	}
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, serviceFunctionID)
	}
}

func (r *serviceFunctionResource) flatten(ctx context.Context, sf *awstypes.ServiceFunction, data *serviceFunctionResourceModel) diag.Diagnostics {
	var diags diag.Diagnostics

	diags.Append(fwflex.Flatten(ctx, sf, data)...)
	if diags.HasError() {
		return diags
	}

	return diags
}

type serviceFunctionImportID struct{}

func (serviceFunctionImportID) Parse(id string) (string, map[string]any, error) {
	const (
		serviceFunctionImportIDPartCount = 2
	)
	parts, err := intflex.ExpandResourceId(id, serviceFunctionImportIDPartCount, false)
	if err != nil {
		return "", nil, err
	}

	result := map[string]any{
		"service_arn":         parts[0],
		"service_function_id": parts[1],
	}

	return id, result, nil
}

func findServiceFunctionByTwoPartKey(ctx context.Context, conn *resiliencehubv2.Client, serviceARN, serviceFunctionID string) (*awstypes.ServiceFunction, error) {
	input := resiliencehubv2.ListServiceFunctionsInput{
		ServiceArn: aws.String(serviceARN),
	}

	return findServiceFunction(ctx, conn, &input, tfslices.WithFilter(func(v awstypes.ServiceFunction) bool {
		return aws.ToString(v.ServiceFunctionId) == serviceFunctionID
	}))
}

func findServiceFunction(ctx context.Context, conn *resiliencehubv2.Client, input *resiliencehubv2.ListServiceFunctionsInput, optFns ...tfslices.FinderOptionsFunc[awstypes.ServiceFunction]) (*awstypes.ServiceFunction, error) {
	output, err := findServiceFunctions(ctx, conn, input, optFns...)

	if err != nil {
		return nil, smarterr.NewError(err)
	}

	return smarterr.Assert(tfresource.AssertSingleValueResult(output))
}

func findServiceFunctions(ctx context.Context, conn *resiliencehubv2.Client, input *resiliencehubv2.ListServiceFunctionsInput, optFns ...tfslices.FinderOptionsFunc[awstypes.ServiceFunction]) ([]awstypes.ServiceFunction, error) {
	output, err := tfslices.CollectAndConcatWithError(listServiceFunctionPages(ctx, conn, input), optFns...)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, smarterr.NewError(&retry.NotFoundError{
			LastError: err,
		})
	}

	return output, nil
}

func listServiceFunctionPages(ctx context.Context, conn *resiliencehubv2.Client, input *resiliencehubv2.ListServiceFunctionsInput, optFns ...func(*resiliencehubv2.Options)) iter.Seq2[[]awstypes.ServiceFunction, error] {
	return func(yield func([]awstypes.ServiceFunction, error) bool) {
		pages := resiliencehubv2.NewListServiceFunctionsPaginator(conn, input)
		for pages.HasMorePages() {
			page, err := pages.NextPage(ctx, optFns...)
			if err != nil {
				yield(nil, fmt.Errorf("listing Resilience Hub V2 Service Functions: %w", err))
				return
			}

			if !yield(page.ServiceFunctions, nil) {
				return
			}
		}
	}
}

type serviceFunctionResourceModel struct {
	framework.WithRegionModel
	Criticality       fwtypes.StringEnum[awstypes.ServiceFunctionCriticality] `tfsdk:"criticality"`
	Description       types.String                                            `tfsdk:"description"`
	Name              types.String                                            `tfsdk:"name"`
	ServiceARN        fwtypes.ARN                                             `tfsdk:"service_arn"`
	ServiceFunctionID types.String                                            `tfsdk:"service_function_id"`
}
