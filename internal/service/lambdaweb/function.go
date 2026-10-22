// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package lambdaweb

import (
	"context"
	"errors"
	"fmt"
	"iter"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/YakDriver/smarterr"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/lambdaweb"
	awstypes "github.com/aws/aws-sdk-go-v2/service/lambdaweb/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	tfiter "github.com/hashicorp/terraform-provider-aws/internal/iter"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

var validFunctionName validator.String = stringvalidator.RegexMatches(
	regexache.MustCompile(`^[0-9A-Za-z][0-9A-Za-z_-]{0,62}[0-9A-Za-z]$`),
	`Function name must be 2 to 64 characters, and can't start or end with a hyphen (-) or underscore (_). Valid characters are a-z, A-Z, 0-9, hyphens (-), and underscores (_).`,
)

// @FrameworkResource("aws_lambdaweb_function", name="Function")
// @IdentityAttribute("function_name")
// @Tags(identifierAttribute="arn")
// @Testing(hasNoPreExistingResource=true)
// @Testing(preCheck="testAccPreCheck")
// @Testing(existsType="github.com/aws/aws-sdk-go-v2/service/lambdaweb;lambdaweb.GetWebFunctionOutput")
// @Testing(importStateIdAttribute="function_name")
func newFunctionResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &functionResource{}

	r.SetDefaultCreateTimeout(15 * time.Minute)
	r.SetDefaultDeleteTimeout(15 * time.Minute)

	return r, nil
}

type functionResource struct {
	framework.ResourceWithModel[functionResourceModel]
	framework.WithTimeouts
	framework.WithImportByIdentity
}

func (r *functionResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrARN: framework.ARNAttributeComputedOnly(),
			"function_name": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					validFunctionName,
				},
			},
			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
		},
		Blocks: map[string]schema.Block{
			names.AttrTimeouts: timeouts.Block(ctx, timeouts.Opts{
				Create: true,
				Delete: true,
			}),
		},
	}
}

func (r *functionResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan functionResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Plan.Get(ctx, &plan))
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().LambdaWebClient(ctx)

	name := fwflex.StringValueFromFramework(ctx, plan.FunctionName)
	var input lambdaweb.CreateWebFunctionInput
	smerr.AddEnrich(ctx, &resp.Diagnostics, fwflex.Expand(ctx, plan, &input))
	if resp.Diagnostics.HasError() {
		return
	}

	input.Tags = getTagsIn(ctx)

	// IAM/S3-policy eventual consistency: a freshly created execution role or
	// bucket policy can take tens of seconds to become visible to the service.
	_, err := tfresource.RetryWhen(ctx, iamPropagationTimeout,
		func(ctx context.Context) (any, error) {
			return conn.CreateWebFunction(ctx, &input)
		},
		func(err error) (bool, error) {
			if errs.IsAErrorMessageContains[*awstypes.ValidationException](err, "cannot be assumed") {
				return true, err
			}
			if errs.IsAErrorMessageContains[*awstypes.ValidationException](err, "does not have s3:GetObject") {
				return true, err
			}
			if errs.IsAErrorMessageContains[*awstypes.ValidationException](err, "S3 versioning enabled") {
				return true, err
			}
			return false, err
		})
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, name)
		return
	}

	function, err := waitFunctionCreated(ctx, conn, name, r.CreateTimeout(ctx, plan.Timeouts))
	if err != nil {
		// Taint the resource.
		resp.State.SetAttribute(ctx, path.Root("function_name"), name)
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, name)
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, r.flatten(ctx, function, &plan))
	if resp.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &plan))
}

func (r *functionResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state functionResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().LambdaWebClient(ctx)

	name := fwflex.StringValueFromFramework(ctx, state.FunctionName)
	function, err := findWebFunctionByName(ctx, conn, name)
	if retry.NotFound(err) {
		resp.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, name)
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, r.flatten(ctx, function, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &state))
}

func (r *functionResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state functionResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().LambdaWebClient(ctx)

	name := fwflex.StringValueFromFramework(ctx, state.FunctionName)
	input := lambdaweb.DeleteWebFunctionInput{
		FunctionName: aws.String(name),
	}
	_, err := conn.DeleteWebFunction(ctx, &input)
	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return
	}
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, name)
		return
	}

	if _, err := waitFunctionDeleted(ctx, conn, name, r.DeleteTimeout(ctx, state.Timeouts)); err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, name)
		return
	}
}

func (r *functionResource) flatten(ctx context.Context, function *lambdaweb.GetWebFunctionOutput, data *functionResourceModel) diag.Diagnostics {
	var diags diag.Diagnostics
	diags.Append(fwflex.Flatten(ctx, function, data)...)
	return diags
}

func listFunctions(ctx context.Context, conn *lambdaweb.Client, input *lambdaweb.ListWebFunctionsInput, optFns ...func(*lambdaweb.Options)) iter.Seq2[awstypes.FunctionSummary, error] {
	return tfiter.ConcatValuesWithError(listFunctionPages(ctx, conn, input, optFns...))
}

func listFunctionPages(ctx context.Context, conn *lambdaweb.Client, input *lambdaweb.ListWebFunctionsInput, optFns ...func(*lambdaweb.Options)) iter.Seq2[[]awstypes.FunctionSummary, error] {
	return func(yield func([]awstypes.FunctionSummary, error) bool) {
		pages := lambdaweb.NewListWebFunctionsPaginator(conn, input)
		for pages.HasMorePages() {
			page, err := pages.NextPage(ctx, optFns...)
			if err != nil {
				yield(nil, fmt.Errorf("listing Lambda Web Functions: %w", err))
				return
			}

			if !yield(page.Functions, nil) {
				return
			}
		}
	}
}

func findWebFunctionByName(ctx context.Context, conn *lambdaweb.Client, name string) (*lambdaweb.GetWebFunctionOutput, error) {
	input := lambdaweb.GetWebFunctionInput{
		FunctionName: aws.String(name),
	}
	return findWebFunction(ctx, conn, &input)
}

func findWebFunction(ctx context.Context, conn *lambdaweb.Client, input *lambdaweb.GetWebFunctionInput) (*lambdaweb.GetWebFunctionOutput, error) {
	out, err := conn.GetWebFunction(ctx, input)
	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, smarterr.NewError(&retry.NotFoundError{
			LastError: err,
		})
	}
	if err != nil {
		return nil, smarterr.NewError(err)
	}
	if out == nil {
		return nil, smarterr.NewError(tfresource.NewEmptyResultError())
	}

	return out, nil
}

func statusFunction(conn *lambdaweb.Client, name string) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		out, err := findWebFunctionByName(ctx, conn, name)
		if retry.NotFound(err) {
			return nil, "", nil
		}
		if err != nil {
			return nil, "", smarterr.NewError(err)
		}

		return out, string(out.State), nil
	}
}

func waitFunctionCreated(ctx context.Context, conn *lambdaweb.Client, name string, timeout time.Duration) (*lambdaweb.GetWebFunctionOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   enum.Slice(awstypes.FunctionStatePending),
		Target:                    enum.Slice(awstypes.FunctionStateActive),
		Refresh:                   statusFunction(conn, name),
		Timeout:                   timeout,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*lambdaweb.GetWebFunctionOutput); ok {
		retry.SetLastError(err, errors.New(aws.ToString(out.StateReason)))
		return out, smarterr.NewError(err)
	}

	return nil, smarterr.NewError(err)
}

func waitFunctionDeleted(ctx context.Context, conn *lambdaweb.Client, name string, timeout time.Duration) (*lambdaweb.GetWebFunctionOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.FunctionStateActive, awstypes.FunctionStateDeleting),
		Target:  []string{},
		Refresh: statusFunction(conn, name),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*lambdaweb.GetWebFunctionOutput); ok {
		retry.SetLastError(err, errors.New(aws.ToString(out.StateReason)))
		return out, smarterr.NewError(err)
	}

	return nil, smarterr.NewError(err)
}

type functionResourceModel struct {
	framework.WithRegionModel
	FunctionARN  types.String   `tfsdk:"arn"`
	FunctionName types.String   `tfsdk:"function_name"`
	Tags         tftags.Map     `tfsdk:"tags"`
	TagsAll      tftags.Map     `tfsdk:"tags_all"`
	Timeouts     timeouts.Value `tfsdk:"timeouts"`
}
