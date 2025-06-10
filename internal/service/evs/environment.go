// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package evs

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/evs"
	awstypes "github.com/aws/aws-sdk-go-v2/service/evs/types"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

// @FrameworkResource("aws_evs_environment", name="Environment")
// @Tags(identifierAttribute="environment_arn")
func newEnvironmentResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	return &environmentResource{}, nil
}

type environmentResource struct {
	framework.ResourceWithConfigure
	framework.WithNoOpUpdate[environmentResourceModel]
}

func (r *environmentResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"environment_arn": framework.ARNAttributeComputedOnly(),
			"environment_id":  framework.IDAttribute(),
		},
	}
}

func (r *environmentResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data environmentResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	var input evs.CreateEnvironmentInput
	response.Diagnostics.Append(fwflex.Expand(ctx, data, &input)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().EVSClient(ctx)

	response.Diagnostics.Append(response.State.Set(ctx, data)...)
}

func (r *environmentResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data environmentResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().EVSClient(ctx)

	id := fwflex.StringValueFromFramework(ctx, data.EnvironmentID)
	output, err := findEnvironmentByID(ctx, conn, id)

	if tfresource.NotFound(err) {
		response.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)

		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading EVS Environment (%s)", id), err.Error())

		return
	}

	response.Diagnostics.Append(fwflex.Flatten(ctx, output, &data)...)
	if response.Diagnostics.HasError() {
		return
	}
}

func (r *environmentResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data environmentResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().EVSClient(ctx)
}

func (r *environmentResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("environment_id"), request, response)
}

func findEnvironmentByID(ctx context.Context, conn *evs.Client, id string) (*awstypes.Environment, error) {
	input := evs.GetEnvironmentInput{
		EnvironmentId: aws.String(id),
	}

	output, err := conn.GetEnvironment(ctx, &input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.Environment == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.Environment, nil
}

type environmentResourceModel struct {
	EnvironmentARN types.String `tfsdk:"environment_arn"`
	EnvironmentID  types.String `tfsdk:"environment_id"`
}
