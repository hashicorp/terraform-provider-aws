// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package qbusiness

import (
	"context"
	"fmt"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/qbusiness"
	awstypes "github.com/aws/aws-sdk-go-v2/service/qbusiness/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_qbusiness_dataaccessor", name="DataAccessor")
// @Tags(identifierAttribute="arn")
func newResourceDataAccessor(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &resourceDataAccessor{}

	r.SetDefaultCreateTimeout(30 * time.Minute)
	r.SetDefaultDeleteTimeout(30 * time.Minute)
	r.SetDefaultUpdateTimeout(30 * time.Minute)

	return r, nil
}

const (
	ResNameDataaccessor            = "Dataaccessor"
	actionConfigurationSchemaLevel = 3
)

type resourceDataAccessor struct {
	framework.ResourceWithConfigure
	framework.WithImportByID
	framework.WithTimeouts
}

func (r *resourceDataAccessor) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = "aws_qbusiness_dataaccessor"
}

func documentAttributeSchema(ctx context.Context) schema.SingleNestedBlock {
	return schema.SingleNestedBlock{
		CustomType: fwtypes.NewObjectTypeOf[documentAttributeData](ctx),
		Attributes: map[string]schema.Attribute{
			names.AttrName: schema.StringAttribute{
				Description: "Identifier for the attribute",
				Optional:    true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 200),
					stringvalidator.RegexMatches(regexache.MustCompile(`^[a-zA-Z0-9_][a-zA-Z0-9_-]*$`), "must be a valid document attribute"),
				},
			},
		},
		Blocks: map[string]schema.Block{
			"value": documentAttributeValueSchema(ctx),
		},
	}
}

func attributeFilterArraySchema(ctx context.Context, level int) schema.ListNestedBlock {
	return schema.ListNestedBlock{
		NestedObject: attributeFilterBlockObject(ctx, level-1),
	}
}

func attributeFilterBlockObject(ctx context.Context, level int) schema.NestedBlockObject {
	b := schema.NestedBlockObject{
		Blocks: map[string]schema.Block{
			"contains_all":           documentAttributeSchema(ctx),
			"contains_any":           documentAttributeSchema(ctx),
			"equals_to":              documentAttributeSchema(ctx),
			"greater_than":           documentAttributeSchema(ctx),
			"greater_than_or_equals": documentAttributeSchema(ctx),
			"less_than":              documentAttributeSchema(ctx),
			"less_than_or_equals":    documentAttributeSchema(ctx),
		},
	}
	if level > 0 {
		b.Blocks["and_all_filters"] = attributeFilterArraySchema(ctx, level-1)
		b.Blocks["not_filter"] = attributeFilterSchema(ctx, level-1)
		b.Blocks["or_all_filters"] = attributeFilterArraySchema(ctx, level-1)
	}
	return b
}

func attributeFilterSchema(ctx context.Context, level int) schema.ListNestedBlock {
	return schema.ListNestedBlock{
		CustomType: fwtypes.NewListNestedObjectTypeOf[attributeFilterData](ctx),
		Validators: []validator.List{
			listvalidator.SizeAtMost(1),
		},
		NestedObject: attributeFilterBlockObject(ctx, level),
	}
}

func (r *resourceDataAccessor) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrApplicationID: schema.StringAttribute{
				Description: "The application ID that the dataaccessor belongs to.",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.RegexMatches(regexache.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9-]{35}$`), "must be a valid application ID"),
				},
			},
			names.AttrID:  framework.IDAttribute(),
			names.AttrARN: framework.ARNAttributeComputedOnly(),
			names.AttrDisplayName: schema.StringAttribute{
				Description: "Display name of the dataaccessor",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 100),
					stringvalidator.RegexMatches(regexache.MustCompile(`[a-zA-Z0-9][a-zA-Z0-9_-]*`), "must not contain control characters"),
				},
			},
			"data_accessor_id":    framework.IDAttribute(),
			"idc_application_arn": framework.ARNAttributeComputedOnly(),
			names.AttrPrincipal: schema.StringAttribute{
				CustomType:  fwtypes.ARNType,
				Description: "ARN of the IAM role for the ISV that will be accessing the data.",
				Required:    true,
			},
			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
		},
		Blocks: map[string]schema.Block{
			"action_configuration": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[actionConfigurationData](ctx),
				Validators: []validator.List{
					listvalidator.SizeAtMost(10),
					listvalidator.SizeAtLeast(1),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"action": schema.StringAttribute{
							Description: "Q Business action that is allowed.",
							Required:    true,
							Validators: []validator.String{
								stringvalidator.RegexMatches(regexache.MustCompile(`qbusiness:[a-zA-Z]+`), "must be a valid qbusiness action"),
							},
						},
					},
					Blocks: map[string]schema.Block{
						"filter_configuration": schema.SingleNestedBlock{
							CustomType: fwtypes.NewObjectTypeOf[actionFilterConfigurationData](ctx),
							Blocks: map[string]schema.Block{
								"document_attribute_filter": attributeFilterSchema(ctx, actionConfigurationSchemaLevel),
							},
						},
					},
				},
			},
			names.AttrTimeouts: timeouts.Block(ctx, timeouts.Opts{
				Create: true,
				Delete: true,
				Update: true,
			}),
		},
	}
}

func (r *resourceDataAccessor) ModifyPlan(ctx context.Context, request resource.ModifyPlanRequest, response *resource.ModifyPlanResponse) {
	r.SetTagsAll(ctx, request, response)
}

func (r *resourceDataAccessor) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data resourceDataAccessorData
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().QBusinessClient(ctx)
	input := &qbusiness.CreateDataAccessorInput{}

	resp.Diagnostics.Append(fwflex.Expand(ctx, data, input)...)
	if resp.Diagnostics.HasError() {
		return
	}

	input.Tags = getTagsIn(ctx)
	input.ClientToken = aws.String(id.UniqueId())

	output, err := conn.CreateDataAccessor(ctx, input)
	if err != nil {
		resp.Diagnostics.AddError("failed to create Q Business dataaccessor", err.Error())
		return
	}

	data.DataAccessorId = fwflex.StringToFramework(ctx, output.DataAccessorId)
	data.DataAccessorArn = fwflex.StringToFramework(ctx, output.DataAccessorArn)
	data.IdcApplicationArn = fwflex.StringToFramework(ctx, output.IdcApplicationArn)

	resp.Diagnostics.Append(data.setID()...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *resourceDataAccessor) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data resourceDataAccessorData
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().QBusinessClient(ctx)
	input := &qbusiness.DeleteDataAccessorInput{
		ApplicationId:  data.ApplicationId.ValueStringPointer(),
		DataAccessorId: data.DataAccessorId.ValueStringPointer(),
	}

	_, err := conn.DeleteDataAccessor(ctx, input)
	if err != nil {
		resp.Diagnostics.AddError("failed to delete Q Business dataaccessor", err.Error())
		return
	}
}

func (r *resourceDataAccessor) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data resourceDataAccessorData
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().QBusinessClient(ctx)
	out, err := FindDataAccessorByID(ctx, conn, data.ID.ValueString())

	if tfresource.NotFound(err) {
		resp.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		resp.State.RemoveResource(ctx)
		return
	}

	if err != nil {
		resp.Diagnostics.AddError(fmt.Sprintf("failed to read Q Business dataaccessor (%s)", data.ID.ValueString()), err.Error())
		return
	}

	resp.Diagnostics.Append(fwflex.Flatten(ctx, out, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *resourceDataAccessor) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
}

type documentAttributeData struct {
	Name  types.String                                              `tfsdk:"name"`
	Value fwtypes.ObjectValueOf[resourceDocumentAttributeValueData] `tfsdk:"value"`
}

type attributeFilterData3 struct {
	ContainsAll         fwtypes.ObjectValueOf[documentAttributeData] `tfsdk:"contains_all"`
	ContainsAny         fwtypes.ObjectValueOf[documentAttributeData] `tfsdk:"contains_any"`
	EqualsTo            fwtypes.ObjectValueOf[documentAttributeData] `tfsdk:"equals_to"`
	GreaterThan         fwtypes.ObjectValueOf[documentAttributeData] `tfsdk:"greater_than"`
	GreaterThanOrEquals fwtypes.ObjectValueOf[documentAttributeData] `tfsdk:"greater_than_or_equals"`
	LessThan            fwtypes.ObjectValueOf[documentAttributeData] `tfsdk:"less_than"`
	LessThanOrEquals    fwtypes.ObjectValueOf[documentAttributeData] `tfsdk:"less_than_or_equals"`
}

type attributeFilterData2 struct {
	AndAllFilters       fwtypes.ListNestedObjectValueOf[attributeFilterData3] `tfsdk:"and_all_filters"`
	ContainsAll         fwtypes.ObjectValueOf[documentAttributeData]          `tfsdk:"contains_all"`
	ContainsAny         fwtypes.ObjectValueOf[documentAttributeData]          `tfsdk:"contains_any"`
	EqualsTo            fwtypes.ObjectValueOf[documentAttributeData]          `tfsdk:"equals_to"`
	GreaterThan         fwtypes.ObjectValueOf[documentAttributeData]          `tfsdk:"greater_than"`
	GreaterThanOrEquals fwtypes.ObjectValueOf[documentAttributeData]          `tfsdk:"greater_than_or_equals"`
	LessThan            fwtypes.ObjectValueOf[documentAttributeData]          `tfsdk:"less_than"`
	LessThanOrEquals    fwtypes.ObjectValueOf[documentAttributeData]          `tfsdk:"less_than_or_equals"`
	NotFilter           fwtypes.ObjectValueOf[attributeFilterData3]           `tfsdk:"not_filter"`
	OrAllFilters        fwtypes.ListNestedObjectValueOf[attributeFilterData3] `tfsdk:"or_all_filters"`
}

type attributeFilterData struct {
	AndAllFilters       fwtypes.ListNestedObjectValueOf[attributeFilterData2] `tfsdk:"and_all_filters"`
	ContainsAll         fwtypes.ObjectValueOf[documentAttributeData]          `tfsdk:"contains_all"`
	ContainsAny         fwtypes.ObjectValueOf[documentAttributeData]          `tfsdk:"contains_any"`
	EqualsTo            fwtypes.ObjectValueOf[documentAttributeData]          `tfsdk:"equals_to"`
	GreaterThan         fwtypes.ObjectValueOf[documentAttributeData]          `tfsdk:"greater_than"`
	GreaterThanOrEquals fwtypes.ObjectValueOf[documentAttributeData]          `tfsdk:"greater_than_or_equals"`
	LessThan            fwtypes.ObjectValueOf[documentAttributeData]          `tfsdk:"less_than"`
	LessThanOrEquals    fwtypes.ObjectValueOf[documentAttributeData]          `tfsdk:"less_than_or_equals"`
	NotFilter           fwtypes.ObjectValueOf[attributeFilterData2]           `tfsdk:"not_filter"`
	OrAllFilters        fwtypes.ListNestedObjectValueOf[attributeFilterData2] `tfsdk:"or_all_filters"`
}

type actionFilterConfigurationData struct {
	DocumentAttributeFilter fwtypes.ListNestedObjectValueOf[attributeFilterData] `tfsdk:"document_attribute_filter"`
}

type actionConfigurationData struct {
	Action              types.String                                         `tfsdk:"action"`
	FilterConfiguration fwtypes.ObjectValueOf[actionFilterConfigurationData] `tfsdk:"filter_configuration"`
}

type resourceDataAccessorData struct {
	ActionConfigurations fwtypes.ListNestedObjectValueOf[actionConfigurationData] `tfsdk:"action_configuration"`
	ApplicationId        types.String                                             `tfsdk:"application_id"`
	DataAccessorArn      types.String                                             `tfsdk:"arn"`
	DataAccessorId       types.String                                             `tfsdk:"data_accessor_id"`
	DisplayName          types.String                                             `tfsdk:"display_name"`
	ID                   types.String                                             `tfsdk:"id"`
	IdcApplicationArn    types.String                                             `tfsdk:"idc_application_arn"`
	Principal            fwtypes.ARN                                              `tfsdk:"principal"`
	Tags                 tftags.Map                                               `tfsdk:"tags"`
	TagsAll              tftags.Map                                               `tfsdk:"tags_all"`
	Timeouts             timeouts.Value                                           `tfsdk:"timeouts"`
}

const (
	dataAccessorResourceIDPartCount = 2
)

func (data *resourceDataAccessorData) setID() diag.Diagnostics {
	var diags diag.Diagnostics

	id, err := flex.FlattenResourceId([]string{data.ApplicationId.ValueString(), data.DataAccessorId.ValueString()}, dataAccessorResourceIDPartCount, false)
	if err != nil {
		diags.AddError(
			create.ProblemStandardMessage(names.QBusiness, create.ErrActionFlatteningResourceId, ResNameDataaccessor, id, err),
			err.Error())
		return diags
	}
	data.ID = types.StringValue(id)
	return diags
}

func FindDataAccessorByID(ctx context.Context, conn *qbusiness.Client, id string) (*qbusiness.GetDataAccessorOutput, error) {
	parts, err := flex.ExpandResourceId(id, dataAccessorResourceIDPartCount, false)

	if err != nil {
		return nil, err
	}

	input := &qbusiness.GetDataAccessorInput{
		ApplicationId:  aws.String(parts[0]),
		DataAccessorId: aws.String(parts[1]),
	}

	output, err := conn.GetDataAccessor(ctx, input)

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

	return output, nil
}
