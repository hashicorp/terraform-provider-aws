// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package logs

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs"
	awstypes "github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs/types"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
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

// @FrameworkResource("aws_cloudwatch_log_delivery_destination", name="Delivery Destination")
// @Tags(identifierAttribute="arn")
// @Testing(tagsTest=false)
func newDeliveryDestinationResource(context.Context) (resource.ResourceWithConfigure, error) {
	r := &deliveryDestinationResource{}

	return r, nil
}

type deliveryDestinationResource struct {
	framework.ResourceWithConfigure
}

func (r *deliveryDestinationResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrARN: framework.ARNAttributeComputedOnly(),
			"delivery_destination_type": schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.DeliveryDestinationType](),
				Computed:   true,
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
			"output_format": schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.OutputFormat](),
				Optional:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
		},
		Blocks: map[string]schema.Block{
			"delivery_destination_configuration": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[deliveryDestinationConfigurationModel](ctx),
				Validators: []validator.List{
					listvalidator.IsRequired(),
					listvalidator.SizeAtLeast(1),
					listvalidator.SizeAtMost(1),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"destination_resource_arn": schema.StringAttribute{
							CustomType: fwtypes.ARNType,
							Required:   true,
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.RequiresReplaceIf(requiresReplaceIfARNServiceChanges, "", ""),
							},
						},
					},
				},
			},
		},
	}
}

func (r *deliveryDestinationResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data deliveryDestinationResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().LogsClient(ctx)

	input := cloudwatchlogs.PutDeliveryDestinationInput{}
	response.Diagnostics.Append(fwflex.Expand(ctx, data, &input)...)
	if response.Diagnostics.HasError() {
		return
	}

	// Additional fields.
	input.Tags = getTagsIn(ctx)

	output, err := conn.PutDeliveryDestination(ctx, &input)

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("creating CloudWatch Logs Delivery Destination (%s)", data.Name.ValueString()), err.Error())

		return
	}

	// Set values for unknowns.
	data.ARN = fwflex.StringToFramework(ctx, output.DeliveryDestination.Arn)
	data.DeliveryDestinationType = fwtypes.StringEnumValue(output.DeliveryDestination.DeliveryDestinationType)

	response.Diagnostics.Append(response.State.Set(ctx, data)...)
}

func (r *deliveryDestinationResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data deliveryDestinationResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().LogsClient(ctx)

	output, err := findDeliveryDestinationByName(ctx, conn, data.Name.ValueString())

	if tfresource.NotFound(err) {
		response.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)

		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading CloudWatch Logs Delivery Destination (%s)", data.Name.ValueString()), err.Error())

		return
	}

	// Set attributes for import.
	response.Diagnostics.Append(fwflex.Flatten(ctx, output, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	// Additional fields.
	// Delivery Destination tags aren't set in the Get response.
	// setTagsOut(ctx, output.Tags)

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *deliveryDestinationResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var old, new deliveryDestinationResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &new)...)
	if response.Diagnostics.HasError() {
		return
	}
	response.Diagnostics.Append(request.State.Get(ctx, &old)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().LogsClient(ctx)

	if !new.DeliveryDestinationConfiguration.Equal(old.DeliveryDestinationConfiguration) || !new.OutputFormat.Equal(old.OutputFormat) {
		input := cloudwatchlogs.PutDeliveryDestinationInput{}
		response.Diagnostics.Append(fwflex.Expand(ctx, new, &input)...)
		if response.Diagnostics.HasError() {
			return
		}

		// Additional fields.
		input.Tags = getTagsIn(ctx)

		output, err := conn.PutDeliveryDestination(ctx, &input)

		if err != nil {
			response.Diagnostics.AddError(fmt.Sprintf("updating CloudWatch Logs Delivery Destination (%s)", new.Name.ValueString()), err.Error())

			return
		}

		// Set values for unknowns.
		new.DeliveryDestinationType = fwtypes.StringEnumValue(output.DeliveryDestination.DeliveryDestinationType)
	} else {
		new.DeliveryDestinationType = old.DeliveryDestinationType
	}

	response.Diagnostics.Append(response.State.Set(ctx, &new)...)
}

func (r *deliveryDestinationResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data deliveryDestinationResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().LogsClient(ctx)

	_, err := conn.DeleteDeliveryDestination(ctx, &cloudwatchlogs.DeleteDeliveryDestinationInput{
		Name: fwflex.StringFromFramework(ctx, data.Name),
	})

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("deleting CloudWatch Logs Delivery Destination (%s)", data.Name.ValueString()), err.Error())

		return
	}
}

func (r *deliveryDestinationResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root(names.AttrName), request, response)
}

func findDeliveryDestinationByName(ctx context.Context, conn *cloudwatchlogs.Client, name string) (*awstypes.DeliveryDestination, error) {
	input := cloudwatchlogs.GetDeliveryDestinationInput{
		Name: aws.String(name),
	}

	return findDeliveryDestination(ctx, conn, &input)
}

func findDeliveryDestination(ctx context.Context, conn *cloudwatchlogs.Client, input *cloudwatchlogs.GetDeliveryDestinationInput) (*awstypes.DeliveryDestination, error) {
	output, err := conn.GetDeliveryDestination(ctx, input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.DeliveryDestination == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.DeliveryDestination, nil
}

// requiresReplaceIfARNServiceChanges forces a new resource if and ARN attribute's service changes.
// If the new value is unknown, force a new resource.
func requiresReplaceIfARNServiceChanges(ctx context.Context, request planmodifier.StringRequest, response *stringplanmodifier.RequiresReplaceIfFuncResponse) {
	requiresReplace := false

	if request.PlanValue.IsUnknown() {
		requiresReplace = true
	} else {
		new, _ := arn.Parse(fwflex.StringValueFromFramework(ctx, request.PlanValue))
		old, _ := arn.Parse(fwflex.StringValueFromFramework(ctx, request.StateValue))
		if new.Service != old.Service {
			requiresReplace = true
		}
	}

	response.RequiresReplace = requiresReplace
}

type deliveryDestinationResourceModel struct {
	ARN                              types.String                                                           `tfsdk:"arn"`
	DeliveryDestinationConfiguration fwtypes.ListNestedObjectValueOf[deliveryDestinationConfigurationModel] `tfsdk:"delivery_destination_configuration"`
	DeliveryDestinationType          fwtypes.StringEnum[awstypes.DeliveryDestinationType]                   `tfsdk:"delivery_destination_type"`
	Name                             types.String                                                           `tfsdk:"name"`
	OutputFormat                     fwtypes.StringEnum[awstypes.OutputFormat]                              `tfsdk:"output_format"`
	Tags                             tftags.Map                                                             `tfsdk:"tags"`
	TagsAll                          tftags.Map                                                             `tfsdk:"tags_all"`
}

type deliveryDestinationConfigurationModel struct {
	DestinationResourceARN fwtypes.ARN `tfsdk:"destination_resource_arn"`
}
