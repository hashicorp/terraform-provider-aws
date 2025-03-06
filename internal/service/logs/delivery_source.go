// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package logs

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs"
	awstypes "github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs/types"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_cloudwatch_log_delivery_source", name="Delivery Source")
// @Tags(identifierAttribute="arn")
// @Testing(tagsTest=false)
func newDeliverySourceResource(context.Context) (resource.ResourceWithConfigure, error) {
	r := &deliverySourceResource{}

	return r, nil
}

type deliverySourceResource struct {
	framework.ResourceWithConfigure
	framework.WithNoOpUpdate[deliverySourceResourceModel]
}

func (*deliverySourceResource) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = "aws_cloudwatch_log_delivery_source"
}

func (r *deliverySourceResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrARN: framework.ARNAttributeComputedOnly(),
			"log_type": schema.StringAttribute{
				Required: true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 255),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			names.AttrName: schema.StringAttribute{
				Required: true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 60),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			names.AttrResourceARN: schema.StringAttribute{
				CustomType: fwtypes.ARNType,
				Required:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"service": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
		},
	}
}

func (r *deliverySourceResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data deliverySourceResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().LogsClient(ctx)

	input := cloudwatchlogs.PutDeliverySourceInput{}
	response.Diagnostics.Append(fwflex.Expand(ctx, data, &input)...)
	if response.Diagnostics.HasError() {
		return
	}

	// Additional fields.
	input.Tags = getTagsIn(ctx)

	output, err := conn.PutDeliverySource(ctx, &input)

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("creating CloudWatch Logs Delivery Source (%s)", data.Name.ValueString()), err.Error())

		return
	}

	// Set values for unknowns.
	data.ARN = fwflex.StringToFramework(ctx, output.DeliverySource.Arn)
	data.Service = fwflex.StringToFramework(ctx, output.DeliverySource.Service)

	response.Diagnostics.Append(response.State.Set(ctx, data)...)
}

func (r *deliverySourceResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data deliverySourceResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().LogsClient(ctx)

	output, err := findDeliverySourceByName(ctx, conn, data.Name.ValueString())

	if tfresource.NotFound(err) {
		response.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)

		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading CloudWatch Logs Delivery Source (%s)", data.Name.ValueString()), err.Error())

		return
	}

	// Set attributes for import.
	response.Diagnostics.Append(fwflex.Flatten(ctx, output, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	// Additional fields.
	if len(output.ResourceArns) > 0 {
		data.ResourceARN = fwtypes.ARNValue(output.ResourceArns[0])
	}
	// Delivery Source tags aren't set in the Get response.
	// setTagsOut(ctx, output.Tags)

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *deliverySourceResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data deliverySourceResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().LogsClient(ctx)

	_, err := conn.DeleteDeliverySource(ctx, &cloudwatchlogs.DeleteDeliverySourceInput{
		Name: fwflex.StringFromFramework(ctx, data.Name),
	})

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("deleting CloudWatch Logs Delivery Source (%s)", data.Name.ValueString()), err.Error())

		return
	}
}

func (r *deliverySourceResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root(names.AttrName), request, response)
}

func findDeliverySourceByName(ctx context.Context, conn *cloudwatchlogs.Client, name string) (*awstypes.DeliverySource, error) {
	input := cloudwatchlogs.GetDeliverySourceInput{
		Name: aws.String(name),
	}

	return findDeliverySource(ctx, conn, &input)
}

func findDeliverySource(ctx context.Context, conn *cloudwatchlogs.Client, input *cloudwatchlogs.GetDeliverySourceInput) (*awstypes.DeliverySource, error) {
	output, err := conn.GetDeliverySource(ctx, input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.DeliverySource == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.DeliverySource, nil
}

type deliverySourceResourceModel struct {
	ARN         types.String `tfsdk:"arn"`
	LogType     types.String `tfsdk:"log_type"`
	Name        types.String `tfsdk:"name"`
	ResourceARN fwtypes.ARN  `tfsdk:"resource_arn"`
	Service     types.String `tfsdk:"service"`
	Tags        tftags.Map   `tfsdk:"tags"`
	TagsAll     tftags.Map   `tfsdk:"tags_all"`
}
