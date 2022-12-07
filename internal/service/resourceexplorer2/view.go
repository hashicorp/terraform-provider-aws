package resourceexplorer2

import (
	"context"
	"fmt"
	"regexp"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/resourceexplorer2"
	awstypes "github.com/aws/aws-sdk-go-v2/service/resourceexplorer2/types"
	"github.com/hashicorp/terraform-plugin-framework-validators/resourcevalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	sdkresource "github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func init() {
	registerFrameworkResourceFactory(newResourceView)
}

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
			"id": framework.IDAttribute(),
			"name": schema.StringAttribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
					stringplanmodifier.UseStateForUnknown(),
				},
				Validators: []validator.String{
					stringvalidator.LengthAtMost(64),
					stringvalidator.RegexMatches(regexp.MustCompile(`^[a-zA-Z0-9\-]$`), `can include letters, digits, and the dash (-) character`),
				},
			},
			"name_prefix": schema.StringAttribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
					stringplanmodifier.UseStateForUnknown(),
				},
				Validators: []validator.String{
					stringvalidator.LengthAtMost(64 - sdkresource.UniqueIDSuffixLength),
					stringvalidator.RegexMatches(regexp.MustCompile(`^[a-zA-Z0-9\-]$`), `can include letters, digits, and the dash (-) character`),
				},
			},
			"tags":     tftags.TagsAttribute(),
			"tags_all": tftags.TagsAttributeComputedOnly(),
		},
		Blocks: map[string]schema.Block{
			"filters": schema.SingleNestedBlock{
				Attributes: map[string]schema.Attribute{
					"filter_string": schema.StringAttribute{
						Required: true,
						Validators: []validator.String{
							stringvalidator.LengthAtMost(2048),
						},
					},
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

	conn := r.Meta().ResourceExplorer2Client
	defaultTagsConfig := r.Meta().DefaultTagsConfig
	ignoreTagsConfig := r.Meta().IgnoreTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(data.Tags))

	name := create.Name(data.Name.ValueString(), data.NamePrefix.ValueString())
	input := &resourceexplorer2.CreateViewInput{
		ClientToken: aws.String(sdkresource.UniqueId()),
		ViewName:    aws.String(name),
	}

	if v, diags := r.expandSearchFilter(ctx, data.Filters); diags.HasError() {
		response.Diagnostics.Append(diags...)

		return
	} else {
		input.Filters = v
	}

	if v, diags := r.expandIncludedProperties(ctx, data.IncludedProperties); diags.HasError() {
		response.Diagnostics.Append(diags...)

		return
	} else {
		input.IncludedProperties = v
	}

	if len(tags) > 0 {
		input.Tags = Tags(tags.IgnoreAWS())
	}

	output, err := conn.CreateView(ctx, input)

	if err != nil {
		response.Diagnostics.AddError("creating Resource Explorer View", err.Error())

		return
	}

	// Set values for unknowns.
	arn := aws.ToString(output.View.ViewArn)
	data.ARN = types.StringValue(arn)
	data.ID = types.StringValue(arn)
	data.Name = types.StringValue(name)
	data.NamePrefix = flex.StringToFramework(ctx, create.NamePrefixFromName(name))
	data.TagsAll = flex.FlattenFrameworkStringValueMap(ctx, tags.IgnoreAWS().IgnoreConfig(ignoreTagsConfig).Map())

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *resourceView) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data resourceViewData

	response.Diagnostics.Append(request.State.Get(ctx, &data)...)

	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().ResourceExplorer2Client
	defaultTagsConfig := r.Meta().DefaultTagsConfig
	ignoreTagsConfig := r.Meta().IgnoreTagsConfig

	output, err := findViewByARN(ctx, conn, data.ARN.ValueString())

	if tfresource.NotFound(err) {
		response.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)

		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading Resource Explorer View (%s)", data.ID.ValueString()), err.Error())

		return
	}

	data.ARN = flex.StringToFramework(ctx, output.View.ViewArn)

	tags := KeyValueTags(output.Tags).IgnoreAWS().IgnoreConfig(ignoreTagsConfig)
	// AWS APIs often return empty lists of tags when none have been configured.
	if tags := tags.RemoveDefaultConfig(defaultTagsConfig).Map(); len(tags) == 0 {
		data.Tags = tftags.Null
	} else {
		data.Tags = flex.FlattenFrameworkStringValueMap(ctx, tags)
	}
	data.TagsAll = flex.FlattenFrameworkStringValueMap(ctx, tags.Map())

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

	conn := r.Meta().ResourceExplorer2Client

	// if !new.Type.Equal(old.Type) {
	// 	input := &resourceexplorer2.UpdateIndexTypeInput{
	// 		Arn:  flex.StringFromFramework(ctx, new.ID),
	// 		Type: awstypes.IndexType(new.Type.ValueString()),
	// 	}

	// 	_, err := conn.UpdateIndexType(ctx, input)

	// 	if err != nil {
	// 		response.Diagnostics.AddError(fmt.Sprintf("updating Resource Explorer Index (%s)", new.ID.ValueString()), err.Error())

	// 		return
	// 	}
	// }

	if !new.TagsAll.Equal(old.TagsAll) {
		if err := UpdateTags(ctx, conn, new.ID.ValueString(), old.TagsAll, new.TagsAll); err != nil {
			response.Diagnostics.AddError(fmt.Sprintf("updating Resource Explorer View (%s) tags", new.ID.ValueString()), err.Error())

			return
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

	conn := r.Meta().ResourceExplorer2Client

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

func (r *resourceView) ConfigValidators(context.Context) []resource.ConfigValidator {
	return []resource.ConfigValidator{
		resourcevalidator.Conflicting(
			path.MatchRoot("name"),
			path.MatchRoot("name_prefix"),
		),
	}
}

func (r *resourceView) expandSearchFilter(ctx context.Context, tfObject types.Object) (*awstypes.SearchFilter, diag.Diagnostics) {
	if tfObject.IsNull() || tfObject.IsUnknown() {
		return nil, nil
	}

	var data viewSearchFilterData

	diags := tfObject.As(ctx, &data, types.ObjectAsOptions{})

	if diags.HasError() {
		return nil, diags
	}

	apiObject := &awstypes.SearchFilter{
		FilterString: flex.StringFromFramework(ctx, data.FilterString),
	}

	return apiObject, nil
}

func (r *resourceView) expandIncludedProperties(ctx context.Context, tfList types.List) ([]awstypes.IncludedProperty, diag.Diagnostics) {
	if tfList.IsNull() || tfList.IsUnknown() {
		return nil, nil
	}

	var data []viewIncludedPropertyData

	diags := tfList.ElementsAs(ctx, &data, false)

	if diags.HasError() {
		return nil, diags
	}

	var apiObjects []awstypes.IncludedProperty

	for _, v := range data {
		apiObjects = append(apiObjects, r.expandIncludedProperty(ctx, v))
	}

	return apiObjects, nil
}

func (r *resourceView) expandIncludedProperty(ctx context.Context, data viewIncludedPropertyData) awstypes.IncludedProperty {
	apiObject := awstypes.IncludedProperty{
		Name: flex.StringFromFramework(ctx, data.Name),
	}

	return apiObject
}

type resourceViewData struct {
	ARN                types.String `tfsdk:"arn"`
	Filters            types.Object `tfsdk:"filters"`
	ID                 types.String `tfsdk:"id"`
	IncludedProperties types.List   `tfsdk:"included_property"`
	Name               types.String `tfsdk:"name"`
	NamePrefix         types.String `tfsdk:"name_prefix"`
	Tags               types.Map    `tfsdk:"tags"`
	TagsAll            types.Map    `tfsdk:"tags_all"`
}

type viewSearchFilterData struct {
	FilterString types.String `tfsdk:"filter_string"`
}

type viewIncludedPropertyData struct {
	Name types.String `tfsdk:"name"`
}

func findViewByARN(ctx context.Context, conn *resourceexplorer2.Client, arn string) (*resourceexplorer2.GetViewOutput, error) {
	input := &resourceexplorer2.GetViewInput{
		ViewArn: aws.String(arn),
	}

	output, err := conn.GetView(ctx, input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, &sdkresource.NotFoundError{
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
	propertyNameTags propertyName = "Tags"
)

func (propertyName) Values() []propertyName {
	return []propertyName{
		propertyNameTags,
	}
}
