// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package logs

import (
	"context"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs"
	awstypes "github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs/types"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
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

// @FrameworkResource("aws_cloudwatch_log_delivery", name="Delivery")
// @Tags(identifierAttribute="arn")
// @Testing(tagsTest=false)
func newDeliveryResource(context.Context) (resource.ResourceWithConfigure, error) {
	r := &deliveryResource{}

	return r, nil
}

type deliveryResource struct {
	framework.ResourceWithConfigure
	framework.WithImportByID
}

func (r *deliveryResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrARN: framework.ARNAttributeComputedOnly(),
			"delivery_destination_arn": schema.StringAttribute{
				CustomType: fwtypes.ARNType,
				Required:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"delivery_source_name": schema.StringAttribute{
				Required: true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 60),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"field_delimiter": schema.StringAttribute{
				Optional: true,
				Computed: true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(0, 5),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			names.AttrID: framework.IDAttribute(),
			"record_fields": schema.ListAttribute{
				CustomType:  fwtypes.ListOfStringType,
				Optional:    true,
				Computed:    true,
				ElementType: types.StringType,
				Validators: []validator.List{
					listvalidator.SizeBetween(0, 128),
					listvalidator.ValueStringsAre(stringvalidator.LengthBetween(1, 64)),
				},
				PlanModifiers: []planmodifier.List{
					listplanmodifier.UseStateForUnknown(),
				},
			},
			"s3_delivery_configuration": framework.ResourceOptionalComputedListOfObjectsAttribute[s3DeliveryConfigurationModel](ctx, 1, s3DeliveryConfigurationListOptions, listplanmodifier.UseStateForUnknown()),
			names.AttrTags:              tftags.TagsAttribute(),
			names.AttrTagsAll:           tftags.TagsAttributeComputedOnly(),
		},
	}
}

var s3DeliveryConfigurationListOptions = []fwtypes.ListNestedObjectOfOption[s3DeliveryConfigurationModel]{
	fwtypes.WithSemanticEqualityFunc(s3DeliverySemanticEquality),
}

func s3DeliverySemanticEquality(ctx context.Context, oldValue, newValue fwtypes.ListNestedObjectValueOf[s3DeliveryConfigurationModel]) (bool, diag.Diagnostics) {
	var diags diag.Diagnostics

	oldValPtr, di := oldValue.ToPtr(ctx)
	diags = append(diags, di...)
	if diags.HasError() {
		return false, diags
	}

	newValPtr, di := newValue.ToPtr(ctx)
	diags = append(diags, di...)
	if diags.HasError() {
		return false, diags
	}

	if oldValPtr != nil && newValPtr != nil {
		if strings.HasSuffix(oldValPtr.SuffixPath.ValueString(), newValPtr.SuffixPath.ValueString()) &&
			oldValPtr.EnableHiveCompatiblePath.Equal(newValPtr.EnableHiveCompatiblePath) {
			return true, diags
		}
	}

	return false, diags
}

func (r *deliveryResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data deliveryResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().LogsClient(ctx)

	input := cloudwatchlogs.CreateDeliveryInput{}
	response.Diagnostics.Append(fwflex.Expand(ctx, data, &input)...)
	if response.Diagnostics.HasError() {
		return
	}

	// Additional fields.
	input.Tags = getTagsIn(ctx)

	output, err := conn.CreateDelivery(ctx, &input)

	if err != nil {
		response.Diagnostics.AddError("creating CloudWatch Logs Delivery", err.Error())

		return
	}

	// Set values for unknowns.
	data.ID = fwflex.StringToFramework(ctx, output.Delivery.Id)

	delivery, err := findDeliveryByID(ctx, conn, data.ID.ValueString())

	if err != nil {
		response.State.SetAttribute(ctx, path.Root(names.AttrID), data.ID) // Set 'id' so as to taint the resource.
		response.Diagnostics.AddError(fmt.Sprintf("reading CloudWatch Logs Delivery (%s)", data.ID.ValueString()), err.Error())

		return
	}

	// Normalize FieldDelimiter.
	if aws.ToString(delivery.FieldDelimiter) == "" && data.FieldDelimiter.IsNull() {
		delivery.FieldDelimiter = nil
	}

	// Normalize S3DeliveryConfiguration.EnableHiveCompatiblePath.
	if delivery.S3DeliveryConfiguration != nil && !aws.ToBool(delivery.S3DeliveryConfiguration.EnableHiveCompatiblePath) {
		if !data.S3DeliveryConfiguration.IsNull() {
			s3DeliveryConfiguration, diags := data.S3DeliveryConfiguration.ToPtr(ctx)
			response.Diagnostics.Append(diags...)
			if response.Diagnostics.HasError() {
				return
			}
			if s3DeliveryConfiguration == nil || s3DeliveryConfiguration.EnableHiveCompatiblePath.IsNull() {
				delivery.S3DeliveryConfiguration.EnableHiveCompatiblePath = nil
			}
		}
	}

	// set s3_delivery_configuration.suffix_path to what was in configuration
	if delivery.S3DeliveryConfiguration != nil && aws.ToString(delivery.S3DeliveryConfiguration.SuffixPath) != "" {
		if !data.S3DeliveryConfiguration.IsNull() {
			s3DeliveryConfiguration, diags := data.S3DeliveryConfiguration.ToPtr(ctx)
			response.Diagnostics.Append(diags...)
			if response.Diagnostics.HasError() {
				return
			}

			if s3DeliveryConfiguration != nil && !s3DeliveryConfiguration.SuffixPath.IsNull() {
				delivery.S3DeliveryConfiguration.SuffixPath = s3DeliveryConfiguration.SuffixPath.ValueStringPointer()
			}
		}
	}

	// Set values for unknowns.
	response.Diagnostics.Append(fwflex.Flatten(ctx, delivery, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, data)...)
}

func (r *deliveryResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data deliveryResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().LogsClient(ctx)

	output, err := findDeliveryByID(ctx, conn, data.ID.ValueString())

	if tfresource.NotFound(err) {
		response.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)

		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading CloudWatch Logs Delivery (%s)", data.ID.ValueString()), err.Error())

		return
	}

	// Normalize FieldDelimiter.
	if aws.ToString(output.FieldDelimiter) == "" && data.FieldDelimiter.IsNull() {
		output.FieldDelimiter = nil
	}

	// Normalize S3DeliveryConfiguration.EnableHiveCompatiblePath.
	if output.S3DeliveryConfiguration != nil && !aws.ToBool(output.S3DeliveryConfiguration.EnableHiveCompatiblePath) {
		if !data.S3DeliveryConfiguration.IsNull() {
			s3DeliveryConfiguration, diags := data.S3DeliveryConfiguration.ToPtr(ctx)
			response.Diagnostics.Append(diags...)
			if response.Diagnostics.HasError() {
				return
			}
			if s3DeliveryConfiguration == nil || s3DeliveryConfiguration.EnableHiveCompatiblePath.IsNull() {
				output.S3DeliveryConfiguration.EnableHiveCompatiblePath = nil
			}
		}
	}

	// Set attributes for import.
	response.Diagnostics.Append(fwflex.Flatten(ctx, output, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	// Additional fields.
	// Delivery tags aren't set in the Get response.
	// setTagsOut(ctx, output.Tags)

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *deliveryResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var old, new deliveryResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &new)...)
	if response.Diagnostics.HasError() {
		return
	}
	response.Diagnostics.Append(request.State.Get(ctx, &old)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().LogsClient(ctx)

	if !new.FieldDelimiter.Equal(old.FieldDelimiter) || !new.RecordFields.Equal(old.RecordFields) || !new.S3DeliveryConfiguration.Equal(old.S3DeliveryConfiguration) {
		input := cloudwatchlogs.UpdateDeliveryConfigurationInput{}
		response.Diagnostics.Append(fwflex.Expand(ctx, new, &input)...)
		if response.Diagnostics.HasError() {
			return
		}

		_, err := conn.UpdateDeliveryConfiguration(ctx, &input)

		if err != nil {
			response.Diagnostics.AddError(fmt.Sprintf("updating CloudWatch Logs Delivery (%s)", new.ID.ValueString()), err.Error())

			return
		}
	}

	response.Diagnostics.Append(response.State.Set(ctx, &new)...)
}

func (r *deliveryResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data deliveryResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().LogsClient(ctx)

	_, err := conn.DeleteDelivery(ctx, &cloudwatchlogs.DeleteDeliveryInput{
		Id: fwflex.StringFromFramework(ctx, data.ID),
	})

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("deleting CloudWatch Logs Delivery (%s)", data.ID.ValueString()), err.Error())

		return
	}
}

func (r *deliveryResource) ModifyPlan(ctx context.Context, request resource.ModifyPlanRequest, response *resource.ModifyPlanResponse) {
	if !request.Plan.Raw.IsNull() && !request.State.Raw.IsNull() {
		var plan, state deliveryResourceModel

		response.Diagnostics.Append(request.State.Get(ctx, &state)...)
		response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)
		if response.Diagnostics.HasError() {
			return
		}

		// state can remain null after create/refresh.
		if !plan.FieldDelimiter.Equal(state.FieldDelimiter) {
			if state.FieldDelimiter.IsNull() && plan.FieldDelimiter.IsUnknown() {
				response.Diagnostics.Append(response.Plan.SetAttribute(ctx, path.Root("field_delimiter"), types.StringNull())...)
				if response.Diagnostics.HasError() {
					return
				}
			}
		}
	}
}

func findDeliveryByID(ctx context.Context, conn *cloudwatchlogs.Client, id string) (*awstypes.Delivery, error) {
	input := cloudwatchlogs.GetDeliveryInput{
		Id: aws.String(id),
	}

	return findDelivery(ctx, conn, &input)
}

func findDelivery(ctx context.Context, conn *cloudwatchlogs.Client, input *cloudwatchlogs.GetDeliveryInput) (*awstypes.Delivery, error) {
	output, err := conn.GetDelivery(ctx, input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.Delivery == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.Delivery, nil
}

type deliveryResourceModel struct {
	ARN                     types.String                                                  `tfsdk:"arn"`
	DeliveryDestinationARN  fwtypes.ARN                                                   `tfsdk:"delivery_destination_arn"`
	DeliverySourceName      types.String                                                  `tfsdk:"delivery_source_name"`
	FieldDelimiter          types.String                                                  `tfsdk:"field_delimiter"`
	ID                      types.String                                                  `tfsdk:"id"`
	RecordFields            fwtypes.ListOfString                                          `tfsdk:"record_fields"`
	S3DeliveryConfiguration fwtypes.ListNestedObjectValueOf[s3DeliveryConfigurationModel] `tfsdk:"s3_delivery_configuration"`
	Tags                    tftags.Map                                                    `tfsdk:"tags"`
	TagsAll                 tftags.Map                                                    `tfsdk:"tags_all"`
}

type s3DeliveryConfigurationModel struct {
	EnableHiveCompatiblePath types.Bool   `tfsdk:"enable_hive_compatible_path"`
	SuffixPath               types.String `tfsdk:"suffix_path"`
}
