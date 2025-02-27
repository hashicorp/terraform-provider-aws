// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package logs

import (
	"context"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs"
	awstypes "github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs/types"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_cloudwatch_log_index_policy", name="Index Policy")
func newIndexPolicyResource(context.Context) (resource.ResourceWithConfigure, error) {
	r := &indexPolicyResource{}

	return r, nil
}

type indexPolicyResource struct {
	framework.ResourceWithConfigure
}

func (r *indexPolicyResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrLogGroupName: schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"policy_document": schema.StringAttribute{
				CustomType:  jsontypes.NormalizedType{},
				Required:    true,
				Description: "Field index filter policy, in JSON",
			},
		},
	}
}

func (r *indexPolicyResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data indexPolicyResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().LogsClient(ctx)

	input := cloudwatchlogs.PutIndexPolicyInput{
		LogGroupIdentifier: fwflex.StringFromFramework(ctx, data.LogGroupName),
		PolicyDocument:     fwflex.StringFromFramework(ctx, data.PolicyDocument),
	}

	_, err := conn.PutIndexPolicy(ctx, &input)

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("creating CloudWatch Logs Index Policy (%s)", data.LogGroupName.ValueString()), err.Error())

		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, data)...)
}

func (r *indexPolicyResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data indexPolicyResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().LogsClient(ctx)

	output, err := findIndexPolicyByLogGroupName(ctx, conn, data.LogGroupName.ValueString())

	if tfresource.NotFound(err) {
		response.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)

		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading CloudWatch Logs Index Policy (%s)", data.LogGroupName.ValueString()), err.Error())

		return
	}

	// Set attributes for import.
	data.LogGroupName = fwflex.StringValueToFramework(ctx, logGroupIdentifierToName(aws.ToString(output.LogGroupIdentifier)))
	data.PolicyDocument = jsontypes.NewNormalizedValue(aws.ToString(output.PolicyDocument))

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *indexPolicyResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var new indexPolicyResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &new)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().LogsClient(ctx)

	input := cloudwatchlogs.PutIndexPolicyInput{
		LogGroupIdentifier: fwflex.StringFromFramework(ctx, new.LogGroupName),
		PolicyDocument:     fwflex.StringFromFramework(ctx, new.PolicyDocument),
	}

	_, err := conn.PutIndexPolicy(ctx, &input)

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("updating CloudWatch Logs Index Policy (%s)", new.LogGroupName.ValueString()), err.Error())

		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &new)...)
}

func (r *indexPolicyResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data indexPolicyResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().LogsClient(ctx)

	_, err := conn.DeleteIndexPolicy(ctx, &cloudwatchlogs.DeleteIndexPolicyInput{
		LogGroupIdentifier: fwflex.StringFromFramework(ctx, data.LogGroupName),
	})

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("deleting CloudWatch Logs Index Policy (%s)", data.LogGroupName.ValueString()), err.Error())

		return
	}
}

func (r *indexPolicyResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root(names.AttrLogGroupName), request, response)
}

func findIndexPolicyByLogGroupName(ctx context.Context, conn *cloudwatchlogs.Client, name string) (*awstypes.IndexPolicy, error) {
	input := cloudwatchlogs.DescribeIndexPoliciesInput{
		LogGroupIdentifiers: []string{name},
	}
	output, err := findIndexPolicy(ctx, conn, &input)

	if err != nil {
		return nil, err
	}

	if output.PolicyDocument == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output, err
}

func findIndexPolicy(ctx context.Context, conn *cloudwatchlogs.Client, input *cloudwatchlogs.DescribeIndexPoliciesInput) (*awstypes.IndexPolicy, error) {
	output, err := findIndexPolicies(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findIndexPolicies(ctx context.Context, conn *cloudwatchlogs.Client, input *cloudwatchlogs.DescribeIndexPoliciesInput) ([]awstypes.IndexPolicy, error) {
	var output []awstypes.IndexPolicy

	err := describeIndexPoliciesPages(ctx, conn, input, func(page *cloudwatchlogs.DescribeIndexPoliciesOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		output = append(output, page.IndexPolicies...)

		return !lastPage
	})

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	return output, nil
}

type indexPolicyResourceModel struct {
	LogGroupName   types.String         `tfsdk:"log_group_name"`
	PolicyDocument jsontypes.Normalized `tfsdk:"policy_document"`
}

func logGroupIdentifierToName(identifier string) string {
	arn, err := arn.Parse(identifier)
	if err != nil {
		return identifier
	}

	return strings.TrimPrefix(arn.Resource, "log-group:")
}
