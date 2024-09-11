// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package securitylake

import (
	"context"
	"errors"
	"fmt"
	"net/url"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/aws/aws-sdk-go-v2/service/securitylake"
	awstypes "github.com/aws/aws-sdk-go-v2/service/securitylake/types"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
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
				Computed:           true,
				DeprecationMessage: "Use subscriber_endpoint instead",
			},
			"subscriber_endpoint": schema.StringAttribute{
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
			names.AttrConfiguration: schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[subscriberNotificationResourceConfigurationModel](ctx),
				Validators: []validator.List{
					listvalidator.IsRequired(),
					listvalidator.SizeAtMost(1),
				},
				NestedObject: schema.NestedBlockObject{
					Blocks: map[string]schema.Block{
						"https_notification_configuration": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[httpsNotificationConfigurationModel](ctx),
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"authorization_api_key_name": schema.StringAttribute{
										Optional: true,
									},
									"authorization_api_key_value": schema.StringAttribute{
										Optional:  true,
										Sensitive: true,
									},
									names.AttrEndpoint: schema.StringAttribute{
										Required: true,
									},
									"http_method": schema.StringAttribute{
										CustomType: fwtypes.StringEnumType[awstypes.HttpMethod](),
										Optional:   true,
									},
									"target_role_arn": schema.StringAttribute{
										CustomType: fwtypes.ARNType,
										Required:   true,
									},
								},
							},
						},
						"sqs_notification_configuration": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[sqsNotificationConfigurationModel](ctx),
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
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

	output, err := conn.CreateSubscriberNotification(ctx, input)
	if err != nil {
		response.Diagnostics.AddError("creating Security Lake Subscriber Notification", err.Error())

		return
	}

	// Set values for unknowns.
	data.EndpointID = fwflex.StringToFramework(ctx, output.SubscriberEndpoint)
	data.SubscriberEndpoint = fwflex.StringToFramework(ctx, output.SubscriberEndpoint)
	data.setID()

	response.Diagnostics.Append(response.State.Set(ctx, data)...)
}

func (r *subscriberNotificationResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data subscriberNotificationResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	if err := data.initFromID(); err != nil {
		response.Diagnostics.AddError("parsing resource ID", err.Error())
		return
	}

	conn := r.Meta().SecurityLakeClient(ctx)

	output, err := findSubscriberNotificationBySubscriberID(ctx, conn, data.SubscriberID.ValueString())

	if tfresource.NotFound(err) {
		response.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		response.Diagnostics.AddError("creating Security Lake Subscriber Notification", err.Error())
		return
	}

	data.EndpointID = fwflex.StringToFramework(ctx, output.SubscriberEndpoint)
	data.SubscriberEndpoint = fwflex.StringToFramework(ctx, output.SubscriberEndpoint)
	data.Configuration = refreshConfiguration(ctx, data.Configuration, output, &response.Diagnostics)
	if response.Diagnostics.HasError() {
		return
	}

	// For HTTPS Configurations, only the `endpoint` value can be read back from the Security Lake API.
	// `authorization_api_key_name` is configured on the EventBridge API Destination created by Security Lake
	// `authorization_api_key_value` is configured in the Secrets Manager Secret used by the EventBridge API Destination
	// Setting `http_method` does not seem to work, and it is not a parameter when using the Console, nor does it affect the EventBridge API Destination
	// `target_role_arn` is used in an unknown location

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *subscriberNotificationResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var old, new subscriberNotificationResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &new)...)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(request.State.Get(ctx, &old)...)
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

		output, err := conn.UpdateSubscriberNotification(ctx, in)
		if err != nil {
			response.Diagnostics.AddError(
				create.ProblemStandardMessage(names.SecurityLake, create.ErrActionUpdating, ResNameSubscriberNotification, new.ID.String(), err),
				err.Error(),
			)
			return
		}

		new.EndpointID = fwflex.StringToFramework(ctx, output.SubscriberEndpoint)
		new.SubscriberEndpoint = fwflex.StringToFramework(ctx, output.SubscriberEndpoint)
		new.setID()
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

// findSubscriberNotificationBySubscriberID returns an `*awstypes.SubscriberResource` because subscriber notifications are not really a standalone concept
func findSubscriberNotificationBySubscriberID(ctx context.Context, conn *securitylake.Client, subscriberID string) (*awstypes.SubscriberResource, error) {
	output, err := findSubscriberByID(ctx, conn, subscriberID)

	if err != nil {
		return nil, err
	}

	if output == nil || output.SubscriberEndpoint == nil {
		return nil, &tfresource.EmptyResultError{}
	}

	return output, nil
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
			var httpsNotificationConfiguration []httpsNotificationConfigurationModel
			diags.Append(item.HTTPSNotificationConfiguration.ElementsAs(ctx, &httpsNotificationConfiguration, false)...)
			notificationConfiguration := expandHTTPSNotificationConfigurationModel(ctx, httpsNotificationConfiguration)
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
	Configuration      fwtypes.ListNestedObjectValueOf[subscriberNotificationResourceConfigurationModel] `tfsdk:"configuration"`
	EndpointID         types.String                                                                      `tfsdk:"endpoint_id"`
	ID                 types.String                                                                      `tfsdk:"id"`
	SubscriberEndpoint types.String                                                                      `tfsdk:"subscriber_endpoint"`
	SubscriberID       types.String                                                                      `tfsdk:"subscriber_id"`
}

const (
	subscriberNotificationIdPartCount = 2
)

func (data *subscriberNotificationResourceModel) initFromID() error {
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

func refreshConfiguration(ctx context.Context, config fwtypes.ListNestedObjectValueOf[subscriberNotificationResourceConfigurationModel], subscriber *awstypes.SubscriberResource, diags *diag.Diagnostics) fwtypes.ListNestedObjectValueOf[subscriberNotificationResourceConfigurationModel] {
	var configData *subscriberNotificationResourceConfigurationModel
	if config.IsNull() {
		configData = &subscriberNotificationResourceConfigurationModel{}
	} else {
		configData, _ = config.ToPtr(ctx)
		if configData == nil {
			configData = &subscriberNotificationResourceConfigurationModel{}
		}
	}
	configData.refresh(ctx, subscriber, diags)
	configurationValue := fwtypes.NewListNestedObjectValueOfPtrMust(ctx, configData)

	return configurationValue
}

type subscriberNotificationResourceConfigurationModel struct {
	HTTPSNotificationConfiguration fwtypes.ListNestedObjectValueOf[httpsNotificationConfigurationModel] `tfsdk:"https_notification_configuration"`
	SqsNotificationConfiguration   fwtypes.ListNestedObjectValueOf[sqsNotificationConfigurationModel]   `tfsdk:"sqs_notification_configuration"`
}

func (m *subscriberNotificationResourceConfigurationModel) refresh(ctx context.Context, subscriber *awstypes.SubscriberResource, diags *diag.Diagnostics) {
	switch getNotificationType(subscriber) {
	case notificationTypeHttps:
		m.refreshHTTPSConfiguration(ctx, subscriber)

	case notificationTypeSqs:
		m.refreshSQSConfiguration(ctx, subscriber)

	default:
		diags.Append(diag.NewWarningDiagnostic(
			"Unexpected Endpoint Type",
			fmt.Sprintf("The subscriber endpoint %q references an unexpected endpoint type. ", aws.ToString(subscriber.SubscriberEndpoint))+
				"Either an SQS topic ARN or an HTTP or HTTPS URL were expected.\n"+
				"Please report this to the provider developer.",
		))
	}
}

func (m *subscriberNotificationResourceConfigurationModel) refreshHTTPSConfiguration(ctx context.Context, subscriber *awstypes.SubscriberResource) {
	var configData *httpsNotificationConfigurationModel
	if m.HTTPSNotificationConfiguration.IsNull() {
		configData = &httpsNotificationConfigurationModel{}
	} else {
		configData, _ = m.HTTPSNotificationConfiguration.ToPtr(ctx)
		if configData == nil {
			configData = &httpsNotificationConfigurationModel{}
		}
	}
	configData.refresh(ctx, subscriber)
	m.HTTPSNotificationConfiguration = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, configData)

	m.SqsNotificationConfiguration = fwtypes.NewListNestedObjectValueOfNull[sqsNotificationConfigurationModel](ctx)
}

func (m *subscriberNotificationResourceConfigurationModel) refreshSQSConfiguration(ctx context.Context, subscriber *awstypes.SubscriberResource) {
	m.HTTPSNotificationConfiguration = fwtypes.NewListNestedObjectValueOfNull[httpsNotificationConfigurationModel](ctx)

	var configData *sqsNotificationConfigurationModel
	if m.HTTPSNotificationConfiguration.IsNull() {
		configData = &sqsNotificationConfigurationModel{}
	} else {
		configData, _ = m.SqsNotificationConfiguration.ToPtr(ctx)
		if configData == nil {
			configData = &sqsNotificationConfigurationModel{}
		}
	}
	configData.refresh(ctx, subscriber)
	m.SqsNotificationConfiguration = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, configData)
}

type sqsNotificationConfigurationModel struct{}

func (m *sqsNotificationConfigurationModel) refresh(_ context.Context, _ *awstypes.SubscriberResource) {
	// no-op
}

type httpsNotificationConfigurationModel struct {
	AuthorizationAPIKeyName  types.String                            `tfsdk:"authorization_api_key_name"`
	AuthorizationAPIKeyValue types.String                            `tfsdk:"authorization_api_key_value"`
	Endpoint                 types.String                            `tfsdk:"endpoint"`
	HTTPMethod               fwtypes.StringEnum[awstypes.HttpMethod] `tfsdk:"http_method"`
	TargetRoleARN            fwtypes.ARN                             `tfsdk:"target_role_arn"`
}

func (m *httpsNotificationConfigurationModel) refresh(ctx context.Context, subscriber *awstypes.SubscriberResource) {
	m.Endpoint = fwflex.StringToFramework(ctx, subscriber.SubscriberEndpoint)
}

type notificationType int

const (
	notificationTypeInvalid notificationType = iota
	notificationTypeHttps
	notificationTypeSqs
)

// getNotificationType takes a `*awstypes.SubscriberResource` because subscriber notifications are not really a standalone concept
func getNotificationType(subscriber *awstypes.SubscriberResource) notificationType {
	endpoint := aws.ToString(subscriber.SubscriberEndpoint)

	if isSQSEndpoint(endpoint) {
		return notificationTypeSqs
	}

	if isHTTPSEndpoint(endpoint) {
		return notificationTypeHttps
	}

	return notificationTypeInvalid
}

func isSQSEndpoint(endpoint string) bool {
	if !arn.IsARN(endpoint) {
		return false
	}

	p, _ := arn.Parse(endpoint)
	return p.Service == "sqs"
}

func isHTTPSEndpoint(endpoint string) bool {
	u, err := url.Parse(endpoint)
	if err != nil {
		return false
	}

	return u.IsAbs() && (u.Scheme == "http" || u.Scheme == "https")
}
