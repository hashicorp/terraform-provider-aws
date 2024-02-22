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
					listvalidator.SizeAtLeast(1),
				},
				NestedObject: schema.NestedBlockObject{
					Blocks: map[string]schema.Block{
						"aws_log_source_resource": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[subscriberAwsLogSourceSourceModel](ctx),
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"source_name": schema.StringAttribute{
										Required: true,
										PlanModifiers: []planmodifier.String{
											stringplanmodifier.RequiresReplace(),
										},
									},
									"source_version": schema.StringAttribute{
										Required: true,
										PlanModifiers: []planmodifier.String{
											stringplanmodifier.RequiresReplace(),
										},
									},
								},
							},
						},
						"custom_log_source_resource": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[subscriberCustomLogSourceSourceModel](ctx),
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"source_name": schema.StringAttribute{
										Required: true,
										PlanModifiers: []planmodifier.String{
											stringplanmodifier.RequiresReplace(),
										},
									},
									"source_version": schema.StringAttribute{
										Required: true,
										PlanModifiers: []planmodifier.String{
											stringplanmodifier.RequiresReplace(),
										},
									},
								},
								Blocks: map[string]schema.Block{
									"attributes": schema.ListNestedBlock{
										CustomType: fwtypes.NewListNestedObjectTypeOf[subscriberCustomLogSourceAttributesModel](ctx),
										NestedObject: schema.NestedBlockObject{
											Attributes: map[string]schema.Attribute{
												"crawler_arn": schema.StringAttribute{
													Required: true,
													PlanModifiers: []planmodifier.String{
														stringplanmodifier.RequiresReplace(),
													},
												},
												"database_arn": schema.StringAttribute{
													Required: true,
													PlanModifiers: []planmodifier.String{
														stringplanmodifier.RequiresReplace(),
													},
												},
												"table_arn": schema.StringAttribute{
													Required: true,
													PlanModifiers: []planmodifier.String{
														stringplanmodifier.RequiresReplace(),
													},
												},
											},
										},
									},
									"provider": schema.ListNestedBlock{
										CustomType: fwtypes.NewListNestedObjectTypeOf[subscriberCustomLogSourceAttributesModel](ctx),
										NestedObject: schema.NestedBlockObject{
											Attributes: map[string]schema.Attribute{
												"location": schema.StringAttribute{
													Required: true,
													PlanModifiers: []planmodifier.String{
														stringplanmodifier.RequiresReplace(),
													},
												},
												"role_arn": schema.StringAttribute{
													Required: true,
													PlanModifiers: []planmodifier.String{
														stringplanmodifier.RequiresReplace(),
													},
												},
											},
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

	var sourcesData []subscriberSourcesModel
	response.Diagnostics.Append(data.Sources.ElementsAs(ctx, &sourcesData, false)...)
	if response.Diagnostics.HasError() {
		return
	}
	sources, _ := expandSubscriptionSources(ctx, sourcesData)

	var subscriberIdentityData []subscriberIdentiryModel
	response.Diagnostics.Append(data.SubscriberIdentity.ElementsAs(ctx, &subscriberIdentityData, false)...)

	input := &securitylake.CreateSubscriberInput{
		Sources:            sources,
		SubscriberIdentity: expandSubsciberIdentity(ctx, subscriberIdentityData),
		SubscriberName:     flex.StringFromFramework(ctx, data.SubscriberName),
	}

	out, err := conn.CreateSubscriber(ctx, input)
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

	data.SubscriberArn = flex.StringToFramework(ctx, subscriber.SubscriberArn)
	data.ID = flex.StringToFramework(ctx, subscriber.SubscriberId)
	subscriberIdentityOut,d := flattenSubsciberIdentity(ctx, subscriber.SubscriberIdentity)
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

	response.Diagnostics.Append(response.State.Set(ctx, data)...)
}

func (r *subscriberResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	conn := r.Meta().SecurityLakeClient(ctx)

	var state subscriberResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	out, err := findSubscriberByID(ctx, conn, state.ID.ValueString())

	if tfresource.NotFound(err) {
		response.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		response.Diagnostics.AddError(
			create.ProblemStandardMessage(names.SecurityLake, create.ErrActionSetting, ResNameSubscriber, state.ID.String(), err),
			err.Error(),
		)
		return
	}

	state.SubscriberArn = flex.StringToFramework(ctx, out.SubscriberArn)
	state.ID = flex.StringToFramework(ctx, out.SubscriberId)
	state.SubscriberName = flex.StringToFramework(ctx, out.SubscriberName)

	// var sources subscriberSourcesModel
	// response.Diagnostics.Append(fwflex.Flatten(ctx, out.Sources, &sources)...)
	// if response.Diagnostics.HasError() {
	// 	return
	// }

	// state.Sources = fwtypes.NewListNestedObjectValueOfPtr(ctx, &sources)

	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
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

func expandSubscriptionSources(ctx context.Context, subscriberSourcesModels []subscriberSourcesModel) ([]awstypes.LogSourceResource, diag.Diagnostics) {
	sources := []awstypes.LogSourceResource{}
	var diags diag.Diagnostics

	for _, item := range subscriberSourcesModels {
		if !item.AwsLogSourceResource.IsNull() {
			var awsLogSources []subscriberAwsLogSourceSourceModel
			diags.Append(item.AwsLogSourceResource.ElementsAs(ctx, &awsLogSources, false)...)
			logSource := expandSubscriberAwsLogSourceSource(ctx, awsLogSources)
			sources = append(sources, logSource)
		}
		if !item.CustomLogSourceResource.IsNull() {
			var customLogSource []subscriberCustomLogSourceSourceModel
			diags.Append(item.CustomLogSourceResource.ElementsAs(ctx, &customLogSource, false)...)
			logSource := expandSubscriberCustomLogSourceSource(ctx, customLogSource)
			sources = append(sources, logSource)
		}
	}

	return sources, diags
}

func expandSubscriberAwsLogSourceSource(ctx context.Context, awsLogSources []subscriberAwsLogSourceSourceModel) *awstypes.LogSourceResourceMemberAwsLogSource {
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

func expandSubscriberCustomLogSourceSource(ctx context.Context, customLogSources []subscriberCustomLogSourceSourceModel) *awstypes.LogSourceResourceMemberCustomLogSource {
	if len(customLogSources) == 0 {
		return nil
	}
	var diags diag.Diagnostics
	var customLogSourceAttributes []subscriberCustomLogSourceAttributesModel
	diags.Append(customLogSources[0].Attributes.ElementsAs(ctx, &customLogSourceAttributes, false)...)

	var customLogSourceProvider []subscriberCustomLogSourceProviderModel
	diags.Append(customLogSources[0].Provider.ElementsAs(ctx, &customLogSourceProvider, false)...)

	return &awstypes.LogSourceResourceMemberCustomLogSource{
		Value: awstypes.CustomLogSourceResource{
			SourceName:    flex.StringFromFramework(ctx, customLogSources[0].SourceName),
			SourceVersion: flex.StringFromFramework(ctx, customLogSources[0].SourceVersion),
			Attributes:    expandsubscriberCustomLogSourceAttributes(ctx, customLogSourceAttributes),
			Provider:      expandSubscriberCustomLogSourceProvider(ctx, customLogSourceProvider),
		},
	}
}

func expandsubscriberCustomLogSourceAttributes(ctx context.Context, customLogSourceAtributes []subscriberCustomLogSourceAttributesModel) *awstypes.CustomLogSourceAttributes {
	if len(customLogSourceAtributes) == 0 {
		return nil
	}
	return &awstypes.CustomLogSourceAttributes{
		CrawlerArn:  flex.StringFromFramework(ctx, customLogSourceAtributes[0].CrawlerARN),
		DatabaseArn: flex.StringFromFramework(ctx, customLogSourceAtributes[0].DatabaseARN),
		TableArn:    flex.StringFromFramework(ctx, customLogSourceAtributes[0].TableARN),
	}
}

func expandSubscriberCustomLogSourceProvider(ctx context.Context, customLogSourceProvider []subscriberCustomLogSourceProviderModel) *awstypes.CustomLogSourceProvider {
	if len(customLogSourceProvider) == 0 {
		return nil
	}
	return &awstypes.CustomLogSourceProvider{
		Location: flex.StringFromFramework(ctx, customLogSourceProvider[0].Location),
		RoleArn:  flex.StringFromFramework(ctx, customLogSourceProvider[0].RoleARN),
	}
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

func flattenSubscriberSourcesModel(ctx context.Context, apiObject []awstypes.LogSourceResource) (types.List, diag.Diagnostics) {
	var diags diag.Diagnostics
	elemType := types.ObjectType{AttrTypes: subscriberSourcesModelAttrTypes}

	elems := []attr.Value{}

	for _,item := range apiObject {
		switch v := item.(type) {
			case *awstypes.LogSourceResourceMemberAwsLogSource:
				subscriberSource, d := flattenSubscriberAwsLogSourceSourceModel(ctx, &v.Value)
				diags.Append(d...)
				elems = append(elems, subscriberSource)
			case *awstypes.LogSourceResourceMemberCustomLogSource:
				subscriberSource, d := flattenSubscriberCustomLogSourceSourceModel(ctx, &v.Value)
				diags.Append(d...)
				elems = append(elems, subscriberSource)
		}
	}

	listVal, d := types.ListValue(elemType, elems)
	diags.Append(d...)

	return listVal, diags
}

func flattenSubscriberAwsLogSourceSourceModel(ctx context.Context, apiObject *awstypes.AwsLogSourceResource) (types.List, diag.Diagnostics) {
	var diags diag.Diagnostics
	elemType := types.ObjectType{AttrTypes: subscriberAwsLogSourceSourceModelAttrTypes}

	if apiObject == nil {
		return types.ListValueMust(elemType, []attr.Value{}), diags
	}

	obj := map[string]attr.Value{
		"source_name": flex.StringValueToFramework(ctx, apiObject.SourceName),
		"source_version": flex.StringToFramework(ctx, apiObject.SourceVersion),
	}

	objVal, d := types.ObjectValue(subscriberAwsLogSourceSourceModelAttrTypes, obj)
	diags.Append(d...)

	listVal, d := types.ListValue(elemType, []attr.Value{objVal})
	diags.Append(d...)

	return listVal, diags
}

func flattenSubscriberCustomLogSourceSourceModel(ctx context.Context, apiObject *awstypes.CustomLogSourceResource) (types.List, diag.Diagnostics) {
	var diags diag.Diagnostics
	elemType := types.ObjectType{AttrTypes: subscriberCustomLogSourceSourceModelAttrTypes}

	if apiObject == nil {
		return types.ListValueMust(elemType, []attr.Value{}), diags
	}

	attributes, d := flattensubscriberCustomLogSourceAttributeModel(ctx, apiObject.Attributes)
	diags.Append(d...)
	provider, d := flattensubscriberCustomLogSourceProviderModel(ctx, apiObject.Provider)
	diags.Append(d...)

	obj := map[string]attr.Value{
		"attributes": attributes,
		"provider":   provider,
		"source_name": flex.StringToFramework(ctx, apiObject.SourceName),
		"source_version": flex.StringToFramework(ctx, apiObject.SourceVersion),
	}

	objVal, d := types.ObjectValue(subscriberCustomLogSourceSourceModelAttrTypes, obj)
	diags.Append(d...)

	listVal, d := types.ListValue(elemType, []attr.Value{objVal})
	diags.Append(d...)

	return listVal, diags
}

func flattensubscriberCustomLogSourceAttributeModel(ctx context.Context, apiObject *awstypes.CustomLogSourceAttributes) (types.List, diag.Diagnostics) {
	var diags diag.Diagnostics
	elemType := types.ObjectType{AttrTypes: subscriberCustomLogSourceAttributesModelAttrTypes }

	if apiObject == nil {
		return types.ListValueMust(elemType, []attr.Value{}), diags
	}

	obj := map[string]attr.Value{
		"crawler_arn": flex.StringToFramework(ctx, apiObject.CrawlerArn),
		"database_arn": flex.StringToFramework(ctx, apiObject.DatabaseArn),
		"table_arn": flex.StringToFramework(ctx, apiObject.TableArn),
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
		"location": flex.StringToFramework(ctx, apiObject.Location),
		"role_arn": flex.StringToFramework(ctx, apiObject.RoleArn),
	}

	objVal, d := types.ObjectValue(subscriberCustomLogSourceProviderModelAttrTypes, obj)
	diags.Append(d...)

	listVal, d := types.ListValue(elemType, []attr.Value{objVal})
	diags.Append(d...)

	return listVal, diags
}

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

	SubscriberIdentityAttrTypes = map[string]attr.Type{
		"subscriber_identity": types.ListType{ElemType: types.ObjectType{AttrTypes: subscriberIdentiryModelAttrTypes}},
	}

	subscriberCustomLogSourceProviderModelAttrTypes = map[string]attr.Type{
		"location": types.StringType,
		"role_arn": types.StringType,
	}

	subscriberCustomLogSourceAttributesModelAttrTypes = map[string]attr.Type{
		"crawler_arn":  types.StringType,
		"database_arn": types.StringType,
		"table_arn":    types.StringType,
	}

	subscriberCustomLogSourceSourceModelAttrTypes = map[string]attr.Type{
		"attributes":    types.ListType{ElemType: types.ObjectType{AttrTypes: subscriberCustomLogSourceAttributesModelAttrTypes}},
		"provider":      types.ListType{ElemType: types.ObjectType{AttrTypes: subscriberCustomLogSourceProviderModelAttrTypes}},
		"source_name":   types.StringType,
		"source_version": types.StringType,
	}

	subscriberAwsLogSourceSourceModelAttrTypes = map[string]attr.Type{
		"source_name":   types.StringType,
		"source_version": types.StringType,
	}

	subscriberSourcesModelAttrTypes = map[string]attr.Type{
		"aws_log_source_resource": types.ListType{ElemType: types.ObjectType{AttrTypes: subscriberAwsLogSourceSourceModelAttrTypes}},
		"custom_log_source_resource": types.ListType{ElemType: types.ObjectType{AttrTypes: subscriberCustomLogSourceSourceModelAttrTypes}},
	}

	sourcesAttrTypes = map[string]attr.Type{
		"sources": types.ListType{ElemType: types.ObjectType{AttrTypes: subscriberSourcesModelAttrTypes}},
	}

)

type subscriberResourceModel struct {
	AccessTypes           fwtypes.SetValueOf[types.String]                         `tfsdk:"access_types"`
	Sources               types.List  `tfsdk:"sources"`
	SubscriberDescription types.String                                             `tfsdk:"subscriber_description"`
	SubscriberIdentity    types.List `tfsdk:"subscriber_identity"`
	SubscriberName        types.String                                             `tfsdk:"subscriber_name"`
	Tags                  types.Map                                                `tfsdk:"tags"`
	TagsAll               types.Map                                                `tfsdk:"tags_all"`
	SubscriberArn         types.String                                             `tfsdk:"arn"`
	ID                    types.String                                             `tfsdk:"id"`
}

type subscriberSourcesModel struct {
	AwsLogSourceResource    fwtypes.ListNestedObjectValueOf[subscriberAwsLogSourceSourceModel]    `tfsdk:"aws_log_source_resource"`
	CustomLogSourceResource fwtypes.ListNestedObjectValueOf[subscriberCustomLogSourceSourceModel] `tfsdk:"custom_log_source_resource"`
}

type subscriberAwsLogSourceSourceModel struct {
	SourceName    types.String `tfsdk:"source_name"`
	SourceVersion types.String `tfsdk:"source_version"`
}

type subscriberCustomLogSourceSourceModel struct {
	Attributes    fwtypes.ListNestedObjectValueOf[subscriberCustomLogSourceAttributesModel] `tfsdk:"attributes"`
	Provider      fwtypes.ListNestedObjectValueOf[subscriberCustomLogSourceProviderModel]   `tfsdk:"provider"`
	SourceName    types.String                                                              `tfsdk:"source_name"`
	SourceVersion types.String                                                              `tfsdk:"source_version"`
}

type subscriberCustomLogSourceAttributesModel struct {
	CrawlerARN  types.String `tfsdk:"crawler_arn"`
	DatabaseARN types.String `tfsdk:"database_arn"`
	TableARN    types.String `tfsdk:"table_arn"`
}

type subscriberCustomLogSourceProviderModel struct {
	Location types.String `tfsdk:"location"`
	RoleARN  types.String `tfsdk:"role_arn"`
}

type subscriberIdentiryModel struct {
	ExternalID types.String `tfsdk:"external_id"`
	Principal  types.String `tfsdk:"principal"`
}
