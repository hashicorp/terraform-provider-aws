// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package resourceexplorer2

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/aws/aws-sdk-go-v2/service/resourceexplorer2"
	awstypes "github.com/aws/aws-sdk-go-v2/service/resourceexplorer2/types"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource(name="View")
// @Tags(identifierAttribute="id")
func newResourceView(context.Context) (resource.ResourceWithConfigure, error) {
	return &resourceView{}, nil
}

type resourceView struct {
	framework.ResourceWithConfigure
}

func (r *resourceView) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = "aws_resourceexplorer2_view"
}

func (r *resourceView) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"arn": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"default_view": schema.BoolAttribute{
				Optional: true,
				Computed: true,
				Default:  booldefault.StaticBool(false),
			},
			"id": framework.IDAttribute(),
			"name": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.LengthAtMost(64),
					stringvalidator.RegexMatches(regexp.MustCompile(`^[a-zA-Z0-9\-]+$`), `can include letters, digits, and the dash (-) character`),
				},
			},
			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
		},
		Blocks: map[string]schema.Block{
			"filters": schema.ListNestedBlock{
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"filter_string": schema.StringAttribute{
							Required: true,
							Validators: []validator.String{
								stringvalidator.LengthAtMost(2048),
							},
						},
					},
				},
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
				},
			},
			"included_property": schema.ListNestedBlock{
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							Required: true,
							Validators: []validator.String{
								enum.FrameworkValidate[propertyName](),
							},
						},
					},
				},
			},
		},
	}
}

func (r *resourceView) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data resourceViewData

	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)

	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().ResourceExplorer2Client(ctx)

	input := &resourceexplorer2.CreateViewInput{
		ClientToken:        aws.String(id.UniqueId()),
		Filters:            flex.ExpandFrameworkListNestedBlockPtr(ctx, data.Filters, r.expandSearchFilter),
		IncludedProperties: flex.ExpandFrameworkListNestedBlock(ctx, data.IncludedProperties, r.expandIncludedProperty),
		Tags:               getTagsIn(ctx),
		ViewName:           aws.String(data.ViewName.ValueString()),
	}

	output, err := conn.CreateView(ctx, input)

	if err != nil {
		response.Diagnostics.AddError("creating Resource Explorer View", err.Error())

		return
	}

	arn := aws.ToString(output.View.ViewArn)

	if data.DefaultView.ValueBool() {
		input := &resourceexplorer2.AssociateDefaultViewInput{
			ViewArn: aws.String(arn),
		}

		_, err := conn.AssociateDefaultView(ctx, input)

		if err != nil {
			response.Diagnostics.AddError(fmt.Sprintf("setting Resource Explorer View (%s) as the default", arn), err.Error())

			return
		}
	}

	// Set values for unknowns.
	data.ViewArn = types.StringValue(arn)
	data.ID = types.StringValue(arn)

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *resourceView) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data resourceViewData

	response.Diagnostics.Append(request.State.Get(ctx, &data)...)

	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().ResourceExplorer2Client(ctx)

	output, err := findViewByARN(ctx, conn, data.ID.ValueString())

	if tfresource.NotFound(err) {
		response.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)

		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading Resource Explorer View (%s)", data.ID.ValueString()), err.Error())

		return
	}

	defaultViewARN, err := findDefaultViewARN(ctx, conn)

	if err != nil {
		response.Diagnostics.AddError("reading Resource Explorer Default View", err.Error())

		return
	}

	view := output.View
	data.ViewArn = flex.StringToFramework(ctx, view.ViewArn)
	data.DefaultView = types.BoolValue(defaultViewARN == data.ViewArn.ValueString())
	data.Filters = r.flattenSearchFilter(ctx, view.Filters)
	data.IncludedProperties = flex.FlattenFrameworkListNestedBlock[viewIncludedPropertyData](ctx, view.IncludedProperties, r.flattenIncludedProperty)

	arn, err := arn.Parse(data.ViewArn.ValueString())

	if err != nil {
		response.Diagnostics.AddError("parsing Resource Explorer View ARN", err.Error())

		return
	}

	// view/${ViewName}/${ViewUuid}
	parts := strings.Split(arn.Resource, "/")

	if n := len(parts); n != 3 {
		response.Diagnostics.AddError("incorrect Resource Explorer View ARN format", fmt.Sprintf("%d parts", n))

		return
	}

	name := parts[1]
	data.ViewName = types.StringValue(name)

	setTagsOut(ctx, output.Tags)

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *resourceView) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var old, new resourceViewData

	response.Diagnostics.Append(request.State.Get(ctx, &old)...)

	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(request.Plan.Get(ctx, &new)...)

	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().ResourceExplorer2Client(ctx)

	if !new.Filters.Equal(old.Filters) || !new.IncludedProperties.Equal(old.IncludedProperties) {
		input := &resourceexplorer2.UpdateViewInput{
			Filters:            flex.ExpandFrameworkListNestedBlockPtr(ctx, new.Filters, r.expandSearchFilter),
			IncludedProperties: flex.ExpandFrameworkListNestedBlock(ctx, new.IncludedProperties, r.expandIncludedProperty),
			ViewArn:            flex.StringFromFramework(ctx, new.ID),
		}

		_, err := conn.UpdateView(ctx, input)

		if err != nil {
			response.Diagnostics.AddError(fmt.Sprintf("updating Resource Explorer View (%s)", new.ID.ValueString()), err.Error())

			return
		}
	}

	if !new.DefaultView.Equal(old.DefaultView) {
		if new.DefaultView.ValueBool() {
			input := &resourceexplorer2.AssociateDefaultViewInput{
				ViewArn: flex.StringFromFramework(ctx, new.ID),
			}

			_, err := conn.AssociateDefaultView(ctx, input)

			if err != nil {
				response.Diagnostics.AddError(fmt.Sprintf("setting Resource Explorer View (%s) as the default", new.ID.ValueString()), err.Error())

				return
			}
		} else {
			input := &resourceexplorer2.DisassociateDefaultViewInput{}

			_, err := conn.DisassociateDefaultView(ctx, input)

			if err != nil {
				response.Diagnostics.AddError(fmt.Sprintf("unsetting Resource Explorer View (%s) as the default", new.ID.ValueString()), err.Error())

				return
			}
		}
	}

	response.Diagnostics.Append(response.State.Set(ctx, &new)...)
}

func (r *resourceView) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data resourceViewData

	response.Diagnostics.Append(request.State.Get(ctx, &data)...)

	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().ResourceExplorer2Client(ctx)

	tflog.Debug(ctx, "deleting Resource Explorer View", map[string]interface{}{
		"id": data.ID.ValueString(),
	})
	_, err := conn.DeleteView(ctx, &resourceexplorer2.DeleteViewInput{
		ViewArn: flex.StringFromFramework(ctx, data.ID),
	})

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("deleting Resource Explorer View (%s)", data.ID.ValueString()), err.Error())

		return
	}
}

func (r *resourceView) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), request, response)
}

func (r *resourceView) ModifyPlan(ctx context.Context, request resource.ModifyPlanRequest, response *resource.ModifyPlanResponse) {
	r.SetTagsAll(ctx, request, response)
}

func (r *resourceView) expandSearchFilter(ctx context.Context, data viewSearchFilterData) *awstypes.SearchFilter {
	return &awstypes.SearchFilter{
		FilterString: flex.StringFromFramework(ctx, data.FilterString),
	}
}

func (r *resourceView) flattenSearchFilter(ctx context.Context, apiObject *awstypes.SearchFilter) types.List {
	attributeTypes := flex.AttributeTypesMust[viewSearchFilterData](ctx)
	elementType := types.ObjectType{AttrTypes: attributeTypes}

	// The default is
	//
	//   "Filters": {
	// 	   "FilterString": ""
	//   },
	//
	// a view that performs no filtering.
	// See https://docs.aws.amazon.com/resource-explorer/latest/apireference/API_CreateView.html#API_CreateView_Example_1.
	if apiObject == nil || len(aws.ToString(apiObject.FilterString)) == 0 {
		return types.ListNull(elementType)
	}

	return types.ListValueMust(elementType, []attr.Value{
		types.ObjectValueMust(attributeTypes, map[string]attr.Value{
			"filter_string": flex.StringToFramework(ctx, apiObject.FilterString),
		}),
	})
}

func (r *resourceView) expandIncludedProperty(ctx context.Context, data viewIncludedPropertyData) awstypes.IncludedProperty {
	return awstypes.IncludedProperty{
		Name: flex.StringFromFramework(ctx, data.Name),
	}
}

func (r *resourceView) flattenIncludedProperty(ctx context.Context, apiObject awstypes.IncludedProperty) viewIncludedPropertyData {
	return viewIncludedPropertyData{
		Name: flex.StringToFramework(ctx, apiObject.Name),
	}
}

// See https://docs.aws.amazon.com/resource-explorer/latest/apireference/API_View.html.
type resourceViewData struct {
	ViewArn            types.String `tfsdk:"arn"`
	DefaultView        types.Bool   `tfsdk:"default_view"`
	Filters            types.List   `tfsdk:"filters"`
	ID                 types.String `tfsdk:"id"`
	IncludedProperties types.List   `tfsdk:"included_property"`
	ViewName           types.String `tfsdk:"name"`
	Tags               types.Map    `tfsdk:"tags"`
	TagsAll            types.Map    `tfsdk:"tags_all"`
}

type viewSearchFilterData struct {
	FilterString types.String `tfsdk:"filter_string"`
}

type viewIncludedPropertyData struct {
	Name types.String `tfsdk:"name"`
}

func findDefaultViewARN(ctx context.Context, conn *resourceexplorer2.Client) (string, error) {
	input := &resourceexplorer2.GetDefaultViewInput{}

	output, err := conn.GetDefaultView(ctx, input)

	if err != nil {
		return "", err
	}

	if output == nil {
		return "", nil
	}

	return aws.ToString(output.ViewArn), nil
}

func findViewByARN(ctx context.Context, conn *resourceexplorer2.Client, arn string) (*resourceexplorer2.GetViewOutput, error) {
	input := &resourceexplorer2.GetViewInput{
		ViewArn: aws.String(arn),
	}

	output, err := conn.GetView(ctx, input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) || errs.IsA[*awstypes.UnauthorizedException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.View == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output, nil
}

type propertyName string

// Enum values for propertyName.
const (
	propertyNameTags propertyName = "tags"
)

func (propertyName) Values() []propertyName {
	return []propertyName{
		propertyNameTags,
	}
}
