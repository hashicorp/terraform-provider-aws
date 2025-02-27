// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package logs

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs"
	awstypes "github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs/types"
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
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

// @FrameworkResource("aws_cloudwatch_log_delivery_destination_policy", name="Delivery Destination Policy")
func newDeliveryDestinationPolicyResource(context.Context) (resource.ResourceWithConfigure, error) {
	r := &deliveryDestinationPolicyResource{}

	return r, nil
}

type deliveryDestinationPolicyResource struct {
	framework.ResourceWithConfigure
}

func (r *deliveryDestinationPolicyResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"delivery_destination_name": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"delivery_destination_policy": schema.StringAttribute{
				CustomType: fwtypes.IAMPolicyType,
				Required:   true,
			},
		},
	}
}

func (r *deliveryDestinationPolicyResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data deliveryDestinationPolicyResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().LogsClient(ctx)

	input := cloudwatchlogs.PutDeliveryDestinationPolicyInput{
		DeliveryDestinationName:   fwflex.StringFromFramework(ctx, data.DeliveryDestinationName),
		DeliveryDestinationPolicy: fwflex.StringFromFramework(ctx, data.DeliveryDestinationPolicy),
	}

	_, err := conn.PutDeliveryDestinationPolicy(ctx, &input)

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("creating CloudWatch Logs Delivery Destination Policy (%s)", data.DeliveryDestinationName.ValueString()), err.Error())

		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, data)...)
}

func (r *deliveryDestinationPolicyResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data deliveryDestinationPolicyResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().LogsClient(ctx)

	output, err := findDeliveryDestinationPolicyByDeliveryDestinationName(ctx, conn, data.DeliveryDestinationName.ValueString())

	if tfresource.NotFound(err) {
		response.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)

		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading CloudWatch Logs Delivery Destination Policy (%s)", data.DeliveryDestinationName.ValueString()), err.Error())

		return
	}

	// Set attributes for import.
	data.DeliveryDestinationPolicy = fwtypes.IAMPolicyValue(aws.ToString(output.DeliveryDestinationPolicy))

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *deliveryDestinationPolicyResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var new deliveryDestinationPolicyResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &new)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().LogsClient(ctx)

	input := cloudwatchlogs.PutDeliveryDestinationPolicyInput{
		DeliveryDestinationName:   fwflex.StringFromFramework(ctx, new.DeliveryDestinationName),
		DeliveryDestinationPolicy: fwflex.StringFromFramework(ctx, new.DeliveryDestinationPolicy),
	}

	_, err := conn.PutDeliveryDestinationPolicy(ctx, &input)

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("updating CloudWatch Logs Delivery Destination Policy (%s)", new.DeliveryDestinationName.ValueString()), err.Error())

		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &new)...)
}

func (r *deliveryDestinationPolicyResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data deliveryDestinationPolicyResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().LogsClient(ctx)

	_, err := conn.DeleteDeliveryDestinationPolicy(ctx, &cloudwatchlogs.DeleteDeliveryDestinationPolicyInput{
		DeliveryDestinationName: fwflex.StringFromFramework(ctx, data.DeliveryDestinationName),
	})

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("deleting CloudWatch Logs Delivery Destination Policy (%s)", data.DeliveryDestinationName.ValueString()), err.Error())

		return
	}
}

func (r *deliveryDestinationPolicyResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("delivery_destination_name"), request, response)
}

func findDeliveryDestinationPolicyByDeliveryDestinationName(ctx context.Context, conn *cloudwatchlogs.Client, name string) (*awstypes.Policy, error) {
	input := cloudwatchlogs.GetDeliveryDestinationPolicyInput{
		DeliveryDestinationName: aws.String(name),
	}
	output, err := findDeliveryDestinationPolicy(ctx, conn, &input)

	if err != nil {
		return nil, err
	}

	if output.DeliveryDestinationPolicy == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output, err
}

func findDeliveryDestinationPolicy(ctx context.Context, conn *cloudwatchlogs.Client, input *cloudwatchlogs.GetDeliveryDestinationPolicyInput) (*awstypes.Policy, error) {
	output, err := conn.GetDeliveryDestinationPolicy(ctx, input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.Policy, nil
}

type deliveryDestinationPolicyResourceModel struct {
	DeliveryDestinationName   types.String      `tfsdk:"delivery_destination_name"`
	DeliveryDestinationPolicy fwtypes.IAMPolicy `tfsdk:"delivery_destination_policy"`
}
