package securitylake

import (
	"context"
	"errors"
	"fmt"
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
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
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
										Computed: true,
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
								},
							},
						},
					},
				},
			},
			"subscriber_identity": schema.ListNestedBlock{
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
	conn := r.Meta().SecurityLakeClient(ctx)
	var diags diag.Diagnostics

	var data subscriberResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	var subscriberIdentityData []subscriberIdentiryModel
	response.Diagnostics.Append(data.SubscriberIdentity.ElementsAs(ctx, &subscriberIdentityData, false)...)
	if response.Diagnostics.HasError() {
		return
	}

	var sourcesData []subscriberSourcesModel
	response.Diagnostics.Append(data.Sources.ElementsAs(ctx, &sourcesData, false)...)
	if response.Diagnostics.HasError() {
		return
	}

	sources, d := expandSubscriptionValueSources(ctx, sourcesData)
	response.Diagnostics.Append(d...)

	input := &securitylake.CreateSubscriberInput{
		Sources:            sources,
		SubscriberIdentity: expandSubsciberIdentity(ctx, subscriberIdentityData),
		SubscriberName:     flex.StringFromFramework(ctx, data.SubscriberName),
	}

	out, err := conn.CreateSubscriber(ctx, input)
	fmt.Println(out)
	if err != nil {
		response.Diagnostics.AddError(
			create.ProblemStandardMessage(names.SecurityLake, create.ErrActionCreating, ResNameSubscriber, data.SubscriberName.String(), err),
			err.Error(),
		)
		return
	}
	if out == nil || out.Subscriber == nil {
		response.Diagnostics.AddError(
			create.ProblemStandardMessage(names.SecurityLake, create.ErrActionCreating, ResNameSubscriber, data.SubscriberName.String(), nil),
			errors.New("empty output").Error(),
		)
		return
	}

	subscriber := out.Subscriber

	data.ID = flex.StringToFramework(ctx, subscriber.SubscriberId)
	data.SubscriberArn = flex.StringToFramework(ctx, subscriber.SubscriberArn)
	data.SubscriberName = flex.StringToFramework(ctx, subscriber.SubscriberName)
	subscriberIdentityOut, d := flattenSubsciberIdentity(ctx, subscriber.SubscriberIdentity)
	diags.Append(d...)
	data.SubscriberIdentity = subscriberIdentityOut
	sourcesOut, d := flattenSubscriberSourcesModel(ctx, subscriber.Sources)
	diags.Append(d...)
	data.Sources = sourcesOut

	// TIP: -- 6. Use a waiter to wait for create to complete
	// createTimeout := r.CreateTimeout(ctx, plan.Timeouts)
	// _, err = waitSubscriberCreated(ctx, conn, plan.ID.ValueString(), createTimeout)
	// if err != nil {
	// 	resp.Diagnostics.AddError(
	// 		create.ProblemStandardMessage(names.SecurityLake, create.ErrActionWaitingForCreation, ResNameSubscriber, plan.Name.String(), err),
	// 		err.Error(),
	// 	)
	// 	return
	// }
	// var sources subscriberSourcesModel
	// response.Diagnostics.Append(fwflex.Flatten(ctx, out.Subscriber.Sources, &sources)...)
	// if response.Diagnostics.HasError() {
	// 	return
	// }
	fmt.Println(response)
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

	data.ID = flex.StringToFramework(ctx, output.SubscriberId)
	data.SubscriberArn = flex.StringToFramework(ctx, output.SubscriberArn)
	data.SubscriberName = flex.StringToFramework(ctx, output.SubscriberName)
	subscriberIdentityOut, d := flattenSubsciberIdentity(ctx, output.SubscriberIdentity)
	diags.Append(d...)
	data.SubscriberIdentity = subscriberIdentityOut
	sourcesOut, d := flattenSubscriberSourcesModel(ctx, output.Sources)
	diags.Append(d...)
	data.Sources = sourcesOut

	// var sources subscriberSourcesModel
	// response.Diagnostics.Append(fwflex.Flatten(ctx, out.Sources, &sources)...)
	// if response.Diagnostics.HasError() {
	// 	return
	// }

	// state.Sources = fwtypes.NewListNestedObjectValueOfPtr(ctx, &sources)

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *subscriberResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	// conn := r.Meta().SecurityLakeClient(ctx)

	// // TIP: -- 2. Fetch the plan
	// var plan, state subscriberResourceData
	// resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	// resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	// if resp.Diagnostics.HasError() {
	// 	return
	// }

	// // TIP: -- 3. Populate a modify input structure and check for changes
	// if !plan.Name.Equal(state.Name) ||
	// 	!plan.Description.Equal(state.Description) ||
	// 	!plan.ComplexArgument.Equal(state.ComplexArgument) ||
	// 	!plan.Type.Equal(state.Type) {

	// 	in := &securitylake.UpdateSubscriberInput{
	// 		// TIP: Mandatory or fields that will always be present can be set when
	// 		// you create the Input structure. (Replace these with real fields.)
	// 		SubscriberId:   aws.String(plan.ID.ValueString()),
	// 		SubscriberName: aws.String(plan.Name.ValueString()),
	// 		SubscriberType: aws.String(plan.Type.ValueString()),
	// 	}

	// 	if !plan.Description.IsNull() {
	// 		// TIP: Optional fields should be set based on whether or not they are
	// 		// used.
	// 		in.Description = aws.String(plan.Description.ValueString())
	// 	}
	// 	if !plan.ComplexArgument.IsNull() {
	// 		// TIP: Use an expander to assign a complex argument. The elements must be
	// 		// deserialized into the appropriate struct before being passed to the expander.
	// 		var tfList []complexArgumentData
	// 		resp.Diagnostics.Append(plan.ComplexArgument.ElementsAs(ctx, &tfList, false)...)
	// 		if resp.Diagnostics.HasError() {
	// 			return
	// 		}

	// 		in.ComplexArgument = expandComplexArgument(tfList)
	// 	}

	// 	// TIP: -- 4. Call the AWS modify/update function
	// 	out, err := conn.UpdateSubscriber(ctx, in)
	// 	if err != nil {
	// 		resp.Diagnostics.AddError(
	// 			create.ProblemStandardMessage(names.SecurityLake, create.ErrActionUpdating, ResNameSubscriber, plan.ID.String(), err),
	// 			err.Error(),
	// 		)
	// 		return
	// 	}
	// 	if out == nil || out.Subscriber == nil {
	// 		resp.Diagnostics.AddError(
	// 			create.ProblemStandardMessage(names.SecurityLake, create.ErrActionUpdating, ResNameSubscriber, plan.ID.String(), nil),
	// 			errors.New("empty output").Error(),
	// 		)
	// 		return
	// 	}

	// 	// TIP: Using the output from the update function, re-set any computed attributes
	// 	plan.ARN = flex.StringToFramework(ctx, out.Subscriber.Arn)
	// 	plan.ID = flex.StringToFramework(ctx, out.Subscriber.SubscriberId)
	// }

	// // TIP: -- 5. Use a waiter to wait for update to complete
	// updateTimeout := r.UpdateTimeout(ctx, plan.Timeouts)
	// _, err := waitSubscriberUpdated(ctx, conn, plan.ID.ValueString(), updateTimeout)
	// if err != nil {
	// 	resp.Diagnostics.AddError(
	// 		create.ProblemStandardMessage(names.SecurityLake, create.ErrActionWaitingForUpdate, ResNameSubscriber, plan.ID.String(), err),
	// 		err.Error(),
	// 	)
	// 	return
	// }

	// // TIP: -- 6. Save the request plan to response state
	// resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
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

	// TIP: -- 5. Use a waiter to wait for delete to complete
	// deleteTimeout := r.DeleteTimeout(ctx, state.Timeouts)
	// _, err = waitSubscriberDeleted(ctx, conn, state.ID.ValueString(), deleteTimeout)
	// if err != nil {
	// 	resp.Diagnostics.AddError(
	// 		create.ProblemStandardMessage(names.SecurityLake, create.ErrActionWaitingForDeletion, ResNameSubscriber, state.ID.String(), err),
	// 		err.Error(),
	// 	)
	// 	return
	// }
}

func (r *subscriberResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// func waitSubscriberCreated(ctx context.Context, conn *securitylake.Client, id string, timeout time.Duration) (*securitylake.Subscriber, error) {
// 	stateConf := &retry.StateChangeConf{
// 		Pending:                   []string{},
// 		Target:                    []string{statusNormal},
// 		Refresh:                   statusSubscriber(ctx, conn, id),
// 		Timeout:                   timeout,
// 		NotFoundChecks:            20,
// 		ContinuousTargetOccurence: 2,
// 	}

// 	outputRaw, err := stateConf.WaitForStateContext(ctx)
// 	if out, ok := outputRaw.(*securitylake.Subscriber); ok {
// 		return out, err
// 	}

// 	return nil, err
// }

// func waitSubscriberUpdated(ctx context.Context, conn *securitylake.Client, id string, timeout time.Duration) (*securitylake.Subscriber, error) {
// 	stateConf := &retry.StateChangeConf{
// 		Pending:                   []string{statusChangePending},
// 		Target:                    []string{statusUpdated},
// 		Refresh:                   statusSubscriber(ctx, conn, id),
// 		Timeout:                   timeout,
// 		NotFoundChecks:            20,
// 		ContinuousTargetOccurence: 2,
// 	}

// 	outputRaw, err := stateConf.WaitForStateContext(ctx)
// 	if out, ok := outputRaw.(*securitylake.Subscriber); ok {
// 		return out, err
// 	}

// 	return nil, err
// }

// func waitSubscriberDeleted(ctx context.Context, conn *securitylake.Client, id string, timeout time.Duration) (*securitylake.Subscriber, error) {
// 	stateConf := &retry.StateChangeConf{
// 		Pending: []string{statusDeleting, statusNormal},
// 		Target:  []string{},
// 		Refresh: statusSubscriber(ctx, conn, id),
// 		Timeout: timeout,
// 	}

// 	outputRaw, err := stateConf.WaitForStateContext(ctx)
// 	if out, ok := outputRaw.(*securitylake.Subscriber); ok {
// 		return out, err
// 	}

// 	return nil, err
// }

// func statusSubscriber(ctx context.Context, conn *securitylake.Client, id string) retry.StateRefreshFunc {
// 	return func() (interface{}, string, error) {
// 		out, err := findSubscriberByID(ctx, conn, id)
// 		if tfresource.NotFound(err) {
// 			return nil, "", nil
// 		}

// 		if err != nil {
// 			return nil, "", err
// 		}

// 		return out, aws.ToString(out.Status), nil
// 	}
// }

func findSubscriberByID(ctx context.Context, conn *securitylake.Client, id string) (*awstypes.SubscriberResource, error) {
	in := &securitylake.GetSubscriberInput{
		SubscriberId: aws.String(id),
	}

	out, err := conn.GetSubscriber(ctx, in)
	if err != nil {
		return nil, err
	}

	return out.Subscriber, nil
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
			SourceName:    awstypes.AwsLogSourceName(*flex.StringFromFramework(ctx, awsLogSources[0].SourceName)),
			SourceVersion: flex.StringFromFramework(ctx, awsLogSources[0].SourceVersion),
		},
	}
}

func expandSubscriberCustomLogSourceSource(ctx context.Context, customLogSources []customLogSubscriberSourceModel) *awstypes.LogSourceResourceMemberCustomLogSource {
	if len(customLogSources) == 0 {
		return nil
	}

	customLogSourceResource := &awstypes.LogSourceResourceMemberCustomLogSource{
		Value: awstypes.CustomLogSourceResource{
			SourceName:    flex.StringFromFramework(ctx, customLogSources[0].SourceName),
			SourceVersion: flex.StringFromFramework(ctx, customLogSources[0].SourceVersion),
		},
	}

	return customLogSourceResource
}

func expandSubsciberIdentity(ctx context.Context, subscriberIdentities []subscriberIdentiryModel) *awstypes.AwsIdentity {
	if len(subscriberIdentities) == 0 {
		return nil
	}
	return &awstypes.AwsIdentity{
		ExternalId: flex.StringFromFramework(ctx, subscriberIdentities[0].ExternalID),
		Principal:  flex.StringFromFramework(ctx, subscriberIdentities[0].Principal),
	}
}

// func flattenSubscriberSourcesModel(ctx context.Context, apiObject []awstypes.LogSourceResource) (types.List, diag.Diagnostics) {
// 	var diags diag.Diagnostics
// 	var elemType types.ObjectType

// 	elems := []attr.Value{}

// 	for _, item := range apiObject {
// 		switch v := item.(type) {
// 		case *awstypes.LogSourceResourceMemberAwsLogSource:
// 			subscriberLogSource, d := flattenSubscriberLogSourceResourceModel(ctx, &v.Value, nil, "aws")
// 			elems = append(elems, subscriberLogSource)
// 			diags.Append(d...)
// 		case *awstypes.LogSourceResourceMemberCustomLogSource:
// 			subscriberLogSource, d := flattenSubscriberLogSourceResourceModel(ctx, nil, &v.Value, "custom")
// 			elems = append(elems, subscriberLogSource)
// 			diags.Append(d...)
// 		}
// 	}

// 	listVal, d := types.ListValue(elemType, elems)
// 	diags.Append(d...)

// 	return listVal, diags
// }

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
			"source_name":    flex.StringValueToFramework(ctx, awsLogApiObject.SourceName),
			"source_version": flex.StringToFramework(ctx, awsLogApiObject.SourceVersion),
		}
		objVal, _ = types.ObjectValue(subscriberAwsLogSourceResourceModelAttrTypes, obj)
	} else if logSourceType == "custom" {
		elemType = types.ObjectType{AttrTypes: subscriberCustomLogSourceResourceModelAttrTypes}
		attributes, d := flattensubscriberCustomLogSourceAttributeModel(ctx, customLogApiObject.Attributes)
		diags.Append(d...)
		obj = map[string]attr.Value{
			"source_name":    flex.StringToFramework(ctx, customLogApiObject.SourceName),
			"source_version": flex.StringToFramework(ctx, customLogApiObject.SourceVersion),
			"attributes":     attributes,
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
		"crawler_arn":  flex.StringToFramework(ctx, apiObject.CrawlerArn),
		"database_arn": flex.StringToFramework(ctx, apiObject.DatabaseArn),
		"table_arn":    flex.StringToFramework(ctx, apiObject.TableArn),
	}

	objVal, d := types.ObjectValue(subscriberCustomLogSourceAttributesModelAttrTypes, obj)
	diags.Append(d...)

	listVal, d := types.ListValue(elemType, []attr.Value{objVal})
	diags.Append(d...)

	return listVal, diags
}

// func flattensubscriberCustomLogSourceProviderModel(ctx context.Context, apiObject *awstypes.CustomLogSourceProvider) (types.List, diag.Diagnostics) {
// 	var diags diag.Diagnostics
// 	elemType := types.ObjectType{AttrTypes: subscriberCustomLogSourceProviderModelAttrTypes}

// 	if apiObject == nil {
// 		return types.ListValueMust(elemType, []attr.Value{}), diags
// 	}

// 	obj := map[string]attr.Value{
// 		"location": flex.StringToFramework(ctx, apiObject.Location),
// 		"role_arn": flex.StringToFramework(ctx, apiObject.RoleArn),
// 	}

// 	objVal, d := types.ObjectValue(subscriberCustomLogSourceProviderModelAttrTypes, obj)
// 	diags.Append(d...)

// 	listVal, d := types.ListValue(elemType, []attr.Value{objVal})
// 	diags.Append(d...)

// 	return listVal, diags
// }

func flattenSubsciberIdentity(ctx context.Context, apiObject *awstypes.AwsIdentity) (types.List, diag.Diagnostics) {
	var diags diag.Diagnostics
	elemType := types.ObjectType{AttrTypes: subscriberIdentiryModelAttrTypes}

	if apiObject == nil {
		return types.ListValueMust(elemType, []attr.Value{}), diags
	}

	obj := map[string]attr.Value{
		"external_id": flex.StringToFramework(ctx, apiObject.ExternalId),
		"principal":   flex.StringToFramework(ctx, apiObject.Principal),
	}

	objVal, d := types.ObjectValue(subscriberIdentiryModelAttrTypes, obj)
	diags.Append(d...)

	listVal, d := types.ListValue(elemType, []attr.Value{objVal})
	diags.Append(d...)

	return listVal, diags
}

var (
	subscriberIdentiryModelAttrTypes = map[string]attr.Type{
		"external_id": types.StringType,
		"principal":   types.StringType,
	}

	awsLogSubscriberSourcesModelAttrTypes   = types.ObjectType{AttrTypes: subscriberAwsLogSourceResourceModelAttrTypes}
	customLogSubscriberSourceModelAttrTypes = types.ObjectType{AttrTypes: subscriberCustomLogSourceResourceModelAttrTypes}

	subscriberAwsLogSourceResourceModelAttrTypes = map[string]attr.Type{
		"source_name":    types.StringType,
		"source_version": types.StringType,
	}

	subscriberCustomLogSourceResourceModelAttrTypes = map[string]attr.Type{
		"source_name":    types.StringType,
		"source_version": types.StringType,
		"attributes":     types.ListType{ElemType: types.ObjectType{AttrTypes: subscriberCustomLogSourceAttributesModelAttrTypes}},
	}

	subscriberCustomLogSourceAttributesModelAttrTypes = map[string]attr.Type{
		"crawler_arn":  types.StringType,
		"database_arn": types.StringType,
		"table_arn":    types.StringType,
	}

	subscriberSourcesModelAttrTypes = map[string]attr.Type{
		"aws_log_source_resource":    types.ListType{ElemType: awsLogSubscriberSourcesModelAttrTypes},
		"custom_log_source_resource": types.ListType{ElemType: customLogSubscriberSourceModelAttrTypes},
	}
)

type subscriberResourceModel struct {
	AccessTypes           fwtypes.SetValueOf[types.String] `tfsdk:"access_types"`
	Sources               types.List                       `tfsdk:"sources"`
	SubscriberDescription types.String                     `tfsdk:"subscriber_description"`
	SubscriberIdentity    types.List                       `tfsdk:"subscriber_identity"`
	SubscriberName        types.String                     `tfsdk:"subscriber_name"`
	Tags                  types.Map                        `tfsdk:"tags"`
	TagsAll               types.Map                        `tfsdk:"tags_all"`
	SubscriberArn         types.String                     `tfsdk:"arn"`
	ID                    types.String                     `tfsdk:"id"`
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
	SourceName    types.String `tfsdk:"source_name"`
	SourceVersion types.String `tfsdk:"source_version"`
	Attributes    types.List   `tfsdk:"attributes"`
}

type subscriberCustomLogSourceAttributesModel struct {
	CrawlerARN  types.String `tfsdk:"crawler_arn"`
	DatabaseARN types.String `tfsdk:"database_arn"`
	TableARN    types.String `tfsdk:"table_arn"`
}

type subscriberIdentiryModel struct {
	ExternalID types.String `tfsdk:"external_id"`
	Principal  types.String `tfsdk:"principal"`
}
