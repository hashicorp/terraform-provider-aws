package securitylake

import (
	"context"
	"errors"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/securitylake"
	awstypes "github.com/aws/aws-sdk-go-v2/service/securitylake/types"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// Function annotations are used for resource registration to the Provider. DO NOT EDIT.
// @FrameworkResource(name="Subscriber")
// @Tags(identifierAttribute="arn")
func newSubscriberResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &subscriberResource{}

	r.SetDefaultCreateTimeout(30 * time.Minute)
	r.SetDefaultUpdateTimeout(30 * time.Minute)
	r.SetDefaultDeleteTimeout(30 * time.Minute)

	return r, nil
}

const (
	ResNameSubscriber = "Subscriber"
)

type subscriberResource struct {
	framework.ResourceWithConfigure
	framework.WithTimeouts
}

func (r *subscriberResource) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = "aws_securitylake_subscriber"
}

func (r *subscriberResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"arn":             framework.ARNAttributeComputedOnly(),
			names.AttrID:      framework.IDAttribute(),
			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
			"access_types": schema.SetAttribute{
				CustomType:  fwtypes.SetOfStringType,
				Optional:    true,
				ElementType: types.StringType,
			},
			"subscriber_description": schema.StringAttribute{
				Optional: true,
			},
			"subscriber_name": schema.StringAttribute{
				Required: true,
			},
		},
		Blocks: map[string]schema.Block{
			"sources": schema.ListNestedBlock{
				PlanModifiers: []planmodifier.List{
					listplanmodifier.RequiresReplace(),
				},
				Validators: []validator.List{
					listvalidator.IsRequired(),
					listvalidator.SizeAtMost(1),
				},
				NestedObject: schema.NestedBlockObject{
					Blocks: map[string]schema.Block{
						"aws_log_source_resource": schema.ListNestedBlock{
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"source_name": schema.StringAttribute{
										Optional: true,
									},
									"source_version": schema.StringAttribute{
										Optional: true,
									},
								},
							},
						},
						"custom_log_source_resource": schema.ListNestedBlock{
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"source_name": schema.StringAttribute{
										Optional: true,
									},
									"source_version": schema.StringAttribute{
										Optional: true,
									},
									"attributes": schema.ListAttribute{
										Computed:   true,
										CustomType: fwtypes.NewListNestedObjectTypeOf[subscriberCustomLogSourceAttributesModel](ctx),
										ElementType: types.ObjectType{
											AttrTypes: map[string]attr.Type{
												"crawler_arn":  types.StringType,
												"database_arn": types.StringType,
												"table_arn":    types.StringType,
											},
										},
										PlanModifiers: []planmodifier.List{
											listplanmodifier.UseStateForUnknown(),
										},
									},
									"provider": schema.ListAttribute{
										CustomType: fwtypes.NewListNestedObjectTypeOf[subscriberCustomLogSourceProviderModel](ctx),
										Computed:   true,
										ElementType: types.ObjectType{
											AttrTypes: map[string]attr.Type{
												"role_arn": types.StringType,
												"location": types.StringType,
											},
										},
										PlanModifiers: []planmodifier.List{
											listplanmodifier.UseStateForUnknown(),
										},
									},
								},
							},
						},
					},
				},
			},
			"subscriber_identity": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[subscriberIdentityModel](ctx),
				PlanModifiers: []planmodifier.List{
					listplanmodifier.RequiresReplace(),
				},
				Validators: []validator.List{
					listvalidator.IsRequired(),
					listvalidator.SizeAtMost(1),
					listvalidator.SizeAtLeast(1),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"external_id": schema.StringAttribute{
							Required: true,
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.RequiresReplace(),
							},
						},
						"principal": schema.StringAttribute{
							Required: true,
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.RequiresReplace(),
							},
						},
					},
				},
			},
		},
	}
}

func (r *subscriberResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var diags diag.Diagnostics
	var data subscriberResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().SecurityLakeClient(ctx)

	var sourcesData []subscriberSourcesModel
	response.Diagnostics.Append(data.Sources.ElementsAs(ctx, &sourcesData, false)...)
	if response.Diagnostics.HasError() {
		return
	}

	sources, d := expandSubscriptionValueSources(ctx, sourcesData)
	response.Diagnostics.Append(d...)

	input := &securitylake.CreateSubscriberInput{}
	response.Diagnostics.Append(fwflex.Expand(ctx, data, input)...)
	if response.Diagnostics.HasError() {
		return
	}

	// Additional fields.
	input.Sources = sources
	input.Tags = getTagsIn(ctx)

	output, err := conn.CreateSubscriber(ctx, input)

	if err != nil {
		response.Diagnostics.AddError(
			create.ProblemStandardMessage(names.SecurityLake, create.ErrActionCreating, ResNameSubscriber, data.SubscriberName.String(), err),
			err.Error(),
		)
		return
	}

	subscriber := output.Subscriber
	data.ID = fwflex.StringToFramework(ctx, subscriber.SubscriberId)
	data.SubscriberArn = fwflex.StringToFramework(ctx, subscriber.SubscriberArn)

	var subscriberIdentity subscriberIdentityModel
	response.Diagnostics.Append(fwflex.Flatten(ctx, subscriber.SubscriberIdentity, &subscriberIdentity)...)
	if response.Diagnostics.HasError() {
		return
	}

	data.SubscriberIdentity = fwtypes.NewListNestedObjectValueOfPtr(ctx, &subscriberIdentity)

	sourcesOutput, d := flattenSubscriberSourcesModel(ctx, subscriber.Sources)
	diags.Append(d...)
	data.Sources = sourcesOutput
	data.SubscriberName = fwflex.StringToFramework(ctx, subscriber.SubscriberName)
	data.SubscriberDescription = fwflex.StringToFramework(ctx, subscriber.SubscriberDescription)

	response.Diagnostics.Append(response.State.Set(ctx, data)...)
}

func (r *subscriberResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	conn := r.Meta().SecurityLakeClient(ctx)
	var diags diag.Diagnostics
	var data subscriberResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	output, err := findSubscriberByID(ctx, conn, data.ID.ValueString())

	if tfresource.NotFound(err) {
		response.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		response.Diagnostics.AddError(
			create.ProblemStandardMessage(names.SecurityLake, create.ErrActionSetting, ResNameSubscriber, data.ID.String(), err),
			err.Error(),
		)
		return
	}

	subscriber := output
	data.ID = fwflex.StringToFramework(ctx, subscriber.SubscriberId)
	data.SubscriberArn = fwflex.StringToFramework(ctx, subscriber.SubscriberArn)

	var subscriberIdentity subscriberIdentityModel
	response.Diagnostics.Append(fwflex.Flatten(ctx, subscriber.SubscriberIdentity, &subscriberIdentity)...)
	if response.Diagnostics.HasError() {
		return
	}

	data.SubscriberIdentity = fwtypes.NewListNestedObjectValueOfPtr(ctx, &subscriberIdentity)

	sourcesOutput, d := flattenSubscriberSourcesModel(ctx, subscriber.Sources)
	diags.Append(d...)
	data.Sources = sourcesOutput

	data.SubscriberName = fwflex.StringToFramework(ctx, subscriber.SubscriberName)
	data.SubscriberDescription = fwflex.StringToFramework(ctx, subscriber.SubscriberDescription)

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *subscriberResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	conn := r.Meta().SecurityLakeClient(ctx)
	var diags diag.Diagnostics
	var plan, state subscriberResourceModel

	response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	if !plan.AccessTypes.Equal(state.AccessTypes) ||
		!plan.SubscriberDescription.Equal(state.SubscriberDescription) ||
		!plan.SubscriberName.Equal(state.SubscriberName) ||
		!plan.SubscriberIdentity.Equal(state.SubscriberIdentity) ||
		!plan.Sources.Equal(state.Sources) ||
		!plan.Tags.Equal(state.Tags) {

		var sourcesData []subscriberSourcesModel
		response.Diagnostics.Append(plan.Sources.ElementsAs(ctx, &sourcesData, false)...)
		if response.Diagnostics.HasError() {
			return
		}

		sources, d := expandSubscriptionValueSources(ctx, sourcesData)
		response.Diagnostics.Append(d...)
		if response.Diagnostics.HasError() {
			return
		}

		input := &securitylake.UpdateSubscriberInput{}
		response.Diagnostics.Append(fwflex.Expand(ctx, plan, input)...)
		if response.Diagnostics.HasError() {
			return
		}

		// Additional fields.
		input.Sources = sources

		output, err := conn.UpdateSubscriber(ctx, input)
		if err != nil {
			response.Diagnostics.AddError(
				create.ProblemStandardMessage(names.SecurityLake, create.ErrActionUpdating, ResNameSubscriber, plan.ID.String(), err),
				err.Error(),
			)
			return
		}
		if output == nil || output.Subscriber == nil {
			response.Diagnostics.AddError(
				create.ProblemStandardMessage(names.SecurityLake, create.ErrActionUpdating, ResNameSubscriber, plan.ID.String(), nil),
				errors.New("empty output").Error(),
			)
			return
		}

		subscriber := output.Subscriber
		plan.ID = fwflex.StringToFramework(ctx, subscriber.SubscriberId)
		plan.SubscriberArn = fwflex.StringToFramework(ctx, subscriber.SubscriberArn)

		var subscriberIdentity subscriberIdentityModel
		response.Diagnostics.Append(fwflex.Flatten(ctx, subscriber.SubscriberIdentity, &subscriberIdentity)...)
		if response.Diagnostics.HasError() {
			return
		}

		plan.SubscriberIdentity = fwtypes.NewListNestedObjectValueOfPtr(ctx, &subscriberIdentity)

		sourcesOutput, d := flattenSubscriberSourcesModel(ctx, subscriber.Sources)
		diags.Append(d...)
		plan.Sources = sourcesOutput

		plan.SubscriberName = fwflex.StringToFramework(ctx, subscriber.SubscriberName)
		plan.SubscriberDescription = fwflex.StringToFramework(ctx, subscriber.SubscriberDescription)
	}

	response.Diagnostics.Append(response.State.Set(ctx, &plan)...)
}

func (r *subscriberResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	conn := r.Meta().SecurityLakeClient(ctx)

	var data subscriberResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	in := &securitylake.DeleteSubscriberInput{
		SubscriberId: aws.String(data.ID.ValueString()),
	}

	_, err := conn.DeleteSubscriber(ctx, in)

	if err != nil {
		var nfe *awstypes.ResourceNotFoundException
		if errors.As(err, &nfe) {
			return
		}
		response.Diagnostics.AddError(
			create.ProblemStandardMessage(names.SecurityLake, create.ErrActionDeleting, ResNameSubscriber, data.ID.String(), err),
			err.Error(),
		)
		return
	}
}

func (r *subscriberResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func findSubscriberByID(ctx context.Context, conn *securitylake.Client, id string) (*awstypes.SubscriberResource, error) {
	input := &securitylake.GetSubscriberInput{
		SubscriberId: aws.String(id),
	}

	output, err := conn.GetSubscriber(ctx, input)
	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.Subscriber, nil
}

func expandSubscriptionValueSources(ctx context.Context, subscriberSourcesModels []subscriberSourcesModel) ([]awstypes.LogSourceResource, diag.Diagnostics) {
	sources := []awstypes.LogSourceResource{}
	var diags diag.Diagnostics

	for _, item := range subscriberSourcesModels {
		if !item.AwsLogSourceResource.IsNull() && (len(item.AwsLogSourceResource.Elements()) > 0) {
			var awsLogSources []awsLogSubscriberSourceModel
			diags.Append(item.AwsLogSourceResource.ElementsAs(ctx, &awsLogSources, false)...)
			subscriberLogSource := expandSubscriberAwsLogSourceSource(ctx, awsLogSources)
			sources = append(sources, subscriberLogSource)
		}
		if (!item.CustomLogSourceResource.IsNull()) && (len(item.CustomLogSourceResource.Elements()) > 0) {
			var customLogSources []customLogSubscriberSourceModel
			diags.Append(item.CustomLogSourceResource.ElementsAs(ctx, &customLogSources, false)...)
			subscriberLogSource := expandSubscriberCustomLogSourceSource(ctx, customLogSources)
			sources = append(sources, subscriberLogSource)
		}
	}

	return sources, diags
}

func expandSubscriberAwsLogSourceSource(ctx context.Context, awsLogSources []awsLogSubscriberSourceModel) *awstypes.LogSourceResourceMemberAwsLogSource {
	if len(awsLogSources) == 0 {
		return nil
	}
	return &awstypes.LogSourceResourceMemberAwsLogSource{
		Value: awstypes.AwsLogSourceResource{
			SourceName:    awstypes.AwsLogSourceName(*fwflex.StringFromFramework(ctx, awsLogSources[0].SourceName)),
			SourceVersion: fwflex.StringFromFramework(ctx, awsLogSources[0].SourceVersion),
		},
	}
}

func expandSubscriberCustomLogSourceSource(ctx context.Context, customLogSources []customLogSubscriberSourceModel) *awstypes.LogSourceResourceMemberCustomLogSource {
	if len(customLogSources) == 0 {
		return nil
	}

	customLogSourceResource := &awstypes.LogSourceResourceMemberCustomLogSource{
		Value: awstypes.CustomLogSourceResource{
			SourceName:    fwflex.StringFromFramework(ctx, customLogSources[0].SourceName),
			SourceVersion: fwflex.StringFromFramework(ctx, customLogSources[0].SourceVersion),
		},
	}

	return customLogSourceResource
}

func flattenSubscriberSourcesModel(ctx context.Context, apiObject []awstypes.LogSourceResource) (types.List, diag.Diagnostics) {
	var diags diag.Diagnostics
	elemType := types.ObjectType{AttrTypes: subscriberSourcesModelAttrTypes}

	obj := map[string]attr.Value{}

	for _, item := range apiObject {
		switch v := item.(type) {
		case *awstypes.LogSourceResourceMemberAwsLogSource:
			subscriberLogSource, d := flattenSubscriberLogSourceResourceModel(ctx, &v.Value, nil, "aws")
			diags.Append(d...)
			obj = map[string]attr.Value{
				"aws_log_source_resource":    subscriberLogSource,
				"custom_log_source_resource": types.ListNull(customLogSubscriberSourceModelAttrTypes),
			}
		case *awstypes.LogSourceResourceMemberCustomLogSource:
			subscriberLogSource, d := flattenSubscriberLogSourceResourceModel(ctx, nil, &v.Value, "custom")
			diags.Append(d...)
			obj = map[string]attr.Value{
				"aws_log_source_resource":    types.ListNull(awsLogSubscriberSourcesModelAttrTypes),
				"custom_log_source_resource": subscriberLogSource,
			}
		}
	}

	objVal, d := types.ObjectValue(subscriberSourcesModelAttrTypes, obj)
	diags.Append(d...)

	listVal, d := types.ListValue(elemType, []attr.Value{objVal})
	diags.Append(d...)

	return listVal, diags
}

func flattenSubscriberLogSourceResourceModel(ctx context.Context, awsLogApiObject *awstypes.AwsLogSourceResource, customLogApiObject *awstypes.CustomLogSourceResource, logSourceType string) (types.List, diag.Diagnostics) {
	var diags diag.Diagnostics
	elemType := types.ObjectType{}
	obj := map[string]attr.Value{}
	var objVal basetypes.ObjectValue

	if logSourceType == "aws" {
		elemType = types.ObjectType{AttrTypes: subscriberAwsLogSourceResourceModelAttrTypes}
		obj = map[string]attr.Value{
			"source_name":    fwflex.StringValueToFramework(ctx, awsLogApiObject.SourceName),
			"source_version": fwflex.StringToFramework(ctx, awsLogApiObject.SourceVersion),
		}
		objVal, _ = types.ObjectValue(subscriberAwsLogSourceResourceModelAttrTypes, obj)
	} else if logSourceType == "custom" {
		elemType = types.ObjectType{AttrTypes: subscriberCustomLogSourceResourceModelAttrTypes}
		attributes, d := flattensubscriberCustomLogSourceAttributeModel(ctx, customLogApiObject.Attributes)
		diags.Append(d...)
		provider, d := flattensubscriberCustomLogSourceProviderModel(ctx, customLogApiObject.Provider)
		diags.Append(d...)
		obj = map[string]attr.Value{
			"source_name":    fwflex.StringToFramework(ctx, customLogApiObject.SourceName),
			"source_version": fwflex.StringToFramework(ctx, customLogApiObject.SourceVersion),
			"attributes":     attributes,
			"provider":       provider,
		}
		objVal, d = types.ObjectValue(subscriberCustomLogSourceResourceModelAttrTypes, obj)
		diags.Append(d...)
	}

	listVal, d := types.ListValue(elemType, []attr.Value{objVal})
	diags.Append(d...)

	return listVal, diags
}

func flattensubscriberCustomLogSourceAttributeModel(ctx context.Context, apiObject *awstypes.CustomLogSourceAttributes) (types.List, diag.Diagnostics) {
	var diags diag.Diagnostics
	elemType := types.ObjectType{AttrTypes: subscriberCustomLogSourceAttributesModelAttrTypes}

	if apiObject == nil {
		return types.ListValueMust(elemType, []attr.Value{}), diags
	}

	obj := map[string]attr.Value{
		"crawler_arn":  fwflex.StringToFramework(ctx, apiObject.CrawlerArn),
		"database_arn": fwflex.StringToFramework(ctx, apiObject.DatabaseArn),
		"table_arn":    fwflex.StringToFramework(ctx, apiObject.TableArn),
	}

	objVal, d := types.ObjectValue(subscriberCustomLogSourceAttributesModelAttrTypes, obj)
	diags.Append(d...)

	listVal, d := types.ListValue(elemType, []attr.Value{objVal})
	diags.Append(d...)

	return listVal, diags
}

func flattensubscriberCustomLogSourceProviderModel(ctx context.Context, apiObject *awstypes.CustomLogSourceProvider) (types.List, diag.Diagnostics) {
	var diags diag.Diagnostics
	elemType := types.ObjectType{AttrTypes: subscriberCustomLogSourceProviderModelAttrTypes}

	if apiObject == nil {
		return types.ListValueMust(elemType, []attr.Value{}), diags
	}

	obj := map[string]attr.Value{
		"location": fwflex.StringToFramework(ctx, apiObject.Location),
		"role_arn": fwflex.StringToFramework(ctx, apiObject.RoleArn),
	}

	objVal, d := types.ObjectValue(subscriberCustomLogSourceProviderModelAttrTypes, obj)
	diags.Append(d...)

	listVal, d := types.ListValue(elemType, []attr.Value{objVal})
	diags.Append(d...)

	return listVal, diags
}

var (
	subscriberAwsLogSourceResourceModelAttrTypes = map[string]attr.Type{
		"source_name":    types.StringType,
		"source_version": types.StringType,
	}

	subscriberCustomLogSourceResourceModelAttrTypes = map[string]attr.Type{
		"source_name":    types.StringType,
		"source_version": types.StringType,
		"attributes":     types.ListType{ElemType: types.ObjectType{AttrTypes: subscriberCustomLogSourceAttributesModelAttrTypes}},
		"provider":       types.ListType{ElemType: types.ObjectType{AttrTypes: subscriberCustomLogSourceProviderModelAttrTypes}},
	}

	subscriberCustomLogSourceAttributesModelAttrTypes = map[string]attr.Type{
		"crawler_arn":  types.StringType,
		"database_arn": types.StringType,
		"table_arn":    types.StringType,
	}

	subscriberCustomLogSourceProviderModelAttrTypes = map[string]attr.Type{
		"location": types.StringType,
		"role_arn": types.StringType,
	}

	awsLogSubscriberSourcesModelAttrTypes   = types.ObjectType{AttrTypes: subscriberAwsLogSourceResourceModelAttrTypes}
	customLogSubscriberSourceModelAttrTypes = types.ObjectType{AttrTypes: subscriberCustomLogSourceResourceModelAttrTypes}

	subscriberSourcesModelAttrTypes = map[string]attr.Type{
		"aws_log_source_resource":    types.ListType{ElemType: awsLogSubscriberSourcesModelAttrTypes},
		"custom_log_source_resource": types.ListType{ElemType: customLogSubscriberSourceModelAttrTypes},
	}
)

type subscriberResourceModel struct {
	AccessTypes           fwtypes.SetValueOf[types.String]                         `tfsdk:"access_types"`
	Sources               types.List                                               `tfsdk:"sources"`
	SubscriberDescription types.String                                             `tfsdk:"subscriber_description"`
	SubscriberIdentity    fwtypes.ListNestedObjectValueOf[subscriberIdentityModel] `tfsdk:"subscriber_identity"`
	SubscriberName        types.String                                             `tfsdk:"subscriber_name"`
	Tags                  types.Map                                                `tfsdk:"tags"`
	TagsAll               types.Map                                                `tfsdk:"tags_all"`
	SubscriberArn         types.String                                             `tfsdk:"arn"`
	ID                    types.String                                             `tfsdk:"id"`
}

type subscriberSourcesModel struct {
	AwsLogSourceResource    types.List `tfsdk:"aws_log_source_resource"`
	CustomLogSourceResource types.List `tfsdk:"custom_log_source_resource"`
}

type awsLogSubscriberSourceModel struct {
	SourceName    types.String `tfsdk:"source_name"`
	SourceVersion types.String `tfsdk:"source_version"`
}

type customLogSubscriberSourceModel struct {
	SourceName    types.String                                                              `tfsdk:"source_name"`
	SourceVersion types.String                                                              `tfsdk:"source_version"`
	Attributes    fwtypes.ListNestedObjectValueOf[subscriberCustomLogSourceAttributesModel] `tfsdk:"attributes"`
	Provider      fwtypes.ListNestedObjectValueOf[subscriberCustomLogSourceProviderModel]   `tfsdk:"provider"`
}

type subscriberCustomLogSourceAttributesModel struct {
	CrawlerARN  types.String `tfsdk:"crawler_arn"`
	DatabaseARN types.String `tfsdk:"database_arn"`
	TableARN    types.String `tfsdk:"table_arn"`
}

type subscriberCustomLogSourceProviderModel struct {
	RoleArn  types.String `tfsdk:"role_arn"`
	Location types.String `tfsdk:"location"`
}

type subscriberIdentityModel struct {
	ExternalID types.String `tfsdk:"external_id"`
	Principal  types.String `tfsdk:"principal"`
}
