// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package securitylake

import (
	"context"
	"errors"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/securitylake"
	awstypes "github.com/aws/aws-sdk-go-v2/service/securitylake/types"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
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
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource(name="Subscriber Notification")
func newSubscriberNotificationResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &subscriberNotificationResource{}

	return r, nil
}

const (
	ResNameSubscriberNotification = "Subscriber Notification"
)

type subscriberNotificationResource struct {
	framework.ResourceWithConfigure
}

func (r *subscriberNotificationResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "aws_securitylake_subscriber_notification"
}

func (r *subscriberNotificationResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrID: framework.IDAttribute(),
			"endpoint_id": schema.StringAttribute{
				Computed: true,
			},
			"subscriber_id": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
		Blocks: map[string]schema.Block{
			"configuration": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[subscriberNotificationResourceConfigurationModel](ctx),
				Validators: []validator.List{
					listvalidator.IsRequired(),
					listvalidator.SizeAtMost(1),
				},
				PlanModifiers: []planmodifier.List{
					listplanmodifier.UseStateForUnknown(),
				},
				NestedObject: schema.NestedBlockObject{
					Blocks: map[string]schema.Block{
						"https_notification_configuration": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[httpsNotificationConfigurationModel](ctx),
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
							},
							PlanModifiers: []planmodifier.List{
								listplanmodifier.UseStateForUnknown(),
								listplanmodifier.RequiresReplace(),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"authorization_api_key_name": schema.StringAttribute{
										Optional: true,
									},
									"authorization_api_key_value": schema.StringAttribute{
										Optional: true,
									},
									"endpoint": schema.StringAttribute{
										Optional: true,
									},
									"http_method": schema.StringAttribute{
										CustomType: fwtypes.StringEnumType[awstypes.HttpMethod](),
										Optional:   true,
									},
									"target_role_arn": schema.StringAttribute{
										CustomType: fwtypes.ARNType,
										Optional:   true,
									},
								},
							},
						},
						"sqs_notification_configuration": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[sqsNotificationConfigurationModel](ctx),
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
							},
							PlanModifiers: []planmodifier.List{
								listplanmodifier.UseStateForUnknown(),
								listplanmodifier.RequiresReplace(),
							},
						},
					},
				},
			},
		},
	}
}

func (r *subscriberNotificationResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data subscriberNotificationResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().SecurityLakeClient(ctx)

	var configurationData []subscriberNotificationResourceConfigurationModel
	response.Diagnostics.Append(data.Configuration.ElementsAs(ctx, &configurationData, false)...)
	if response.Diagnostics.HasError() {
		return
	}

	configuration, d := expandSubscriberNotificationResourceConfiguration(ctx, configurationData)
	response.Diagnostics.Append(d...)

	input := &securitylake.CreateSubscriberNotificationInput{
		SubscriberId:  fwflex.StringFromFramework(ctx, data.SubscriberID),
		Configuration: configuration,
	}

	_, err := conn.CreateSubscriberNotification(ctx, input)
	if err != nil {
		response.Diagnostics.AddError("creating Security Lake Subscriber Notification", err.Error())

		return
	}

	output, endpointID, err := findSubscriberNotificationByEndPointID(ctx, conn, data.SubscriberID.ValueString())
	if err != nil {
		response.Diagnostics.AddError("creating Security Lake Subscriber Notification", err.Error())

		return
	}
	parts, err := flex.ExpandResourceId(aws.ToString(output), subscriberNotificationIdPartCount, false)
	if err != nil {
		response.Diagnostics.AddError("creating Security Lake Subscriber Notification", err.Error())

		return
	}

	// Set values for unknowns.
	data.SubscriberID = fwflex.StringToFramework(ctx, &parts[0])
	data.EndpointID = fwflex.StringToFramework(ctx, endpointID)
	data.setID()

	response.Diagnostics.Append(response.State.Set(ctx, data)...)
}

func (r *subscriberNotificationResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data subscriberNotificationResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	if err := data.InitFromID(); err != nil {
		response.Diagnostics.AddError("parsing resource ID", err.Error())

		return
	}

	conn := r.Meta().SecurityLakeClient(ctx)

	output, endpointID, err := findSubscriberNotificationByEndPointID(ctx, conn, data.SubscriberID.ValueString())

	if tfresource.NotFound(err) || output == nil {
		response.State.RemoveResource(ctx)
		return
	}

	parts, err := flex.ExpandResourceId(aws.ToString(output), subscriberNotificationIdPartCount, false)
	if err != nil {
		response.Diagnostics.AddError("creating Security Lake Subscriber Notification", err.Error())

		return
	}

	data.SubscriberID = fwflex.StringToFramework(ctx, &parts[0])
	data.EndpointID = fwflex.StringToFramework(ctx, endpointID)

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *subscriberNotificationResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var old, new subscriberNotificationResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &old)...)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(request.State.Get(ctx, &new)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().SecurityLakeClient(ctx)

	if !old.Configuration.Equal(new.Configuration) {
		var configurationData []subscriberNotificationResourceConfigurationModel
		response.Diagnostics.Append(new.Configuration.ElementsAs(ctx, &configurationData, false)...)
		if response.Diagnostics.HasError() {
			return
		}

		configuration, d := expandSubscriberNotificationResourceConfiguration(ctx, configurationData)
		response.Diagnostics.Append(d...)
		if response.Diagnostics.HasError() {
			return
		}

		in := &securitylake.UpdateSubscriberNotificationInput{
			SubscriberId:  new.SubscriberID.ValueStringPointer(),
			Configuration: configuration,
		}

		_, err := conn.UpdateSubscriberNotification(ctx, in)
		if err != nil {
			response.Diagnostics.AddError(
				create.ProblemStandardMessage(names.SecurityLake, create.ErrActionUpdating, ResNameSubscriberNotification, new.ID.String(), err),
				err.Error(),
			)
			return
		}
	}

	response.Diagnostics.Append(response.State.Set(ctx, &new)...)
}

func (r *subscriberNotificationResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := r.Meta().SecurityLakeClient(ctx)

	var data subscriberNotificationResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	in := &securitylake.DeleteSubscriberNotificationInput{
		SubscriberId: fwflex.StringFromFramework(ctx, data.SubscriberID),
	}

	_, err := conn.DeleteSubscriberNotification(ctx, in)
	if err != nil {
		var nfe *awstypes.ResourceNotFoundException
		if errors.As(err, &nfe) {
			return
		}
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.SecurityLake, create.ErrActionDeleting, ResNameSubscriberNotification, data.ID.String(), err),
			err.Error(),
		)
		return
	}
}

func (r *subscriberNotificationResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root(names.AttrID), req, resp)
}

func findSubscriberNotificationByEndPointID(ctx context.Context, conn *securitylake.Client, subscriberID string) (*string, *string, error) {
	var resourceID string
	output, err := findSubscriberByID(ctx, conn, subscriberID)

	if err != nil {
		return nil, nil, err
	}

	if output == nil || output.SubscriberEndpoint == nil {
		return nil, nil, &tfresource.EmptyResultError{}
	}

	resourceID = fmt.Sprintf("%s,%s", aws.ToString(output.SubscriberId), "notification")

	return &resourceID, output.SubscriberEndpoint, nil
}

func expandSubscriberNotificationResourceConfiguration(ctx context.Context, subscriberNotificationResourceConfigurationModels []subscriberNotificationResourceConfigurationModel) (awstypes.NotificationConfiguration, diag.Diagnostics) {
	configuration := []awstypes.NotificationConfiguration{}
	var diags diag.Diagnostics

	for _, item := range subscriberNotificationResourceConfigurationModels {
		if !item.SqsNotificationConfiguration.IsNull() && (len(item.SqsNotificationConfiguration.Elements()) > 0) {
			var sqsNotificationConfiguration []sqsNotificationConfigurationModel
			diags.Append(item.SqsNotificationConfiguration.ElementsAs(ctx, &sqsNotificationConfiguration, false)...)
			notificationConfiguration := expandSQSNotificationConfigurationModel(sqsNotificationConfiguration)
			configuration = append(configuration, notificationConfiguration)
		}
		if (!item.HTTPSNotificationConfiguration.IsNull()) && (len(item.HTTPSNotificationConfiguration.Elements()) > 0) {
			var hppsNotificationConfiguration []httpsNotificationConfigurationModel
			diags.Append(item.HTTPSNotificationConfiguration.ElementsAs(ctx, &hppsNotificationConfiguration, false)...)
			notificationConfiguration := expandHTTPSNotificationConfigurationModel(ctx, hppsNotificationConfiguration)
			configuration = append(configuration, notificationConfiguration)
		}
	}

	return configuration[0], diags
}

func expandHTTPSNotificationConfigurationModel(ctx context.Context, httpsNotifications []httpsNotificationConfigurationModel) *awstypes.NotificationConfigurationMemberHttpsNotificationConfiguration {
	if len(httpsNotifications) == 0 {
		return nil
	}

	return &awstypes.NotificationConfigurationMemberHttpsNotificationConfiguration{
		Value: awstypes.HttpsNotificationConfiguration{
			AuthorizationApiKeyName:  fwflex.StringFromFramework(ctx, httpsNotifications[0].AuthorizationAPIKeyName),
			AuthorizationApiKeyValue: fwflex.StringFromFramework(ctx, httpsNotifications[0].AuthorizationAPIKeyValue),
			Endpoint:                 fwflex.StringFromFramework(ctx, httpsNotifications[0].Endpoint),
			HttpMethod:               httpsNotifications[0].HTTPMethod.ValueEnum(),
			TargetRoleArn:            fwflex.StringFromFramework(ctx, httpsNotifications[0].TargetRoleARN),
		},
	}
}

func expandSQSNotificationConfigurationModel(SQSNotifications []sqsNotificationConfigurationModel) *awstypes.NotificationConfigurationMemberSqsNotificationConfiguration {
	if len(SQSNotifications) == 0 {
		return nil
	}

	return &awstypes.NotificationConfigurationMemberSqsNotificationConfiguration{
		Value: awstypes.SqsNotificationConfiguration{},
	}
}

type subscriberNotificationResourceModel struct {
	Configuration fwtypes.ListNestedObjectValueOf[subscriberNotificationResourceConfigurationModel] `tfsdk:"configuration"`
	EndpointID    types.String                                                                      `tfsdk:"endpoint_id"`
	ID            types.String                                                                      `tfsdk:"id"`
	SubscriberID  types.String                                                                      `tfsdk:"subscriber_id"`
}

const (
	subscriberNotificationIdPartCount = 2
)

func (data *subscriberNotificationResourceModel) InitFromID() error {
	id := data.ID.ValueString()
	parts, err := flex.ExpandResourceId(id, subscriberNotificationIdPartCount, false)

	if err != nil {
		return err
	}

	data.SubscriberID = types.StringValue(parts[0])

	return nil
}

func (data *subscriberNotificationResourceModel) setID() {
	data.ID = types.StringValue(errs.Must(flex.FlattenResourceId([]string{data.SubscriberID.ValueString(), "notification"}, subscriberNotificationIdPartCount, false)))
}

type subscriberNotificationResourceConfigurationModel struct {
	HTTPSNotificationConfiguration fwtypes.ListNestedObjectValueOf[httpsNotificationConfigurationModel] `tfsdk:"https_notification_configuration"`
	SqsNotificationConfiguration   fwtypes.ListNestedObjectValueOf[sqsNotificationConfigurationModel]   `tfsdk:"sqs_notification_configuration"`
}

type sqsNotificationConfigurationModel struct{}

type httpsNotificationConfigurationModel struct {
	AuthorizationAPIKeyName  types.String                            `tfsdk:"authorization_api_key_name"`
	AuthorizationAPIKeyValue types.String                            `tfsdk:"authorization_api_key_value"`
	Endpoint                 types.String                            `tfsdk:"endpoint"`
	HTTPMethod               fwtypes.StringEnum[awstypes.HttpMethod] `tfsdk:"http_method"`
	TargetRoleARN            fwtypes.ARN                             `tfsdk:"target_role_arn"`
}
