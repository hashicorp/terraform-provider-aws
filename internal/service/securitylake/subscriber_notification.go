package securitylake

import (
	"context"
	"errors"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/securitylake"
	awstypes "github.com/aws/aws-sdk-go-v2/service/securitylake/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource(name="Subscriber Notification")
func newSubscriberNotificationResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &subscriberNotificationResource{}

	r.SetDefaultCreateTimeout(30 * time.Minute)
	r.SetDefaultUpdateTimeout(30 * time.Minute)
	r.SetDefaultDeleteTimeout(30 * time.Minute)

	return r, nil
}

const (
	ResNameSubscriberNotification = "Subscriber Notification"
)

type subscriberNotificationResource struct {
	framework.ResourceWithConfigure
	framework.WithTimeouts
}

func (r *subscriberNotificationResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "aws_securitylake_subscriber_notification"
}

func (r *subscriberNotificationResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"arn": framework.ARNAttributeComputedOnly(),
			"id":  framework.IDAttribute(),
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
				Validators: []validator.List{
					listvalidator.IsRequired(),
					listvalidator.SizeAtMost(1),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"sqs_notification_configuration": schema.BoolAttribute{
							Optional: true,
						},
					},
					Blocks: map[string]schema.Block{
						"https_notification_configuration": schema.ListNestedBlock{
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"endpoint": schema.StringAttribute{
										Required: true,
									},
									"target_role_arn": schema.StringAttribute{
										Required: true,
									},
									"authorization_api_key_name": schema.StringAttribute{
										Optional: true,
									},
									"authorization_api_key_value ": schema.StringAttribute{
										Optional: true,
									},
									"http_method": schema.StringAttribute{
										Optional: true,
									},
								},
							},
						},
					},
				},
			},
			"timeouts": timeouts.Block(ctx, timeouts.Opts{
				Create: true,
				Update: true,
				Delete: true,
			}),
		},
	}
}

func (r *subscriberNotificationResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	conn := r.Meta().SecurityLakeClient(ctx)

	var data subscriberNotificationResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	input := &securitylake.CreateSubscriberNotificationInput{}
	response.Diagnostics.Append(fwflex.Expand(ctx, data, input)...)
	if response.Diagnostics.HasError() {
		return
	}

	output, err := conn.CreateSubscriberNotification(ctx, input)
	if err != nil {
		response.Diagnostics.AddError(
			create.ProblemStandardMessage(names.SecurityLake, create.ErrActionCreating, ResNameSubscriberNotification, data.ID.String(), err),
			err.Error(),
		)
		return
	}
	if output == nil {
		response.Diagnostics.AddError(
			create.ProblemStandardMessage(names.SecurityLake, create.ErrActionCreating, ResNameSubscriberNotification, data.ID.String(), nil),
			errors.New("empty output").Error(),
		)
		return
	}

	data.ID = flex.StringToFramework(ctx, output.SubscriberEndpoint)

	response.Diagnostics.Append(response.State.Set(ctx, data)...)
}

func (r *subscriberNotificationResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// conn := r.Meta().SecurityLakeClient(ctx)

	// var state subscriberNotificationResourceModel
	// resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	// if resp.Diagnostics.HasError() {
	// 	return
	// }

	// resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *subscriberNotificationResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// conn := r.Meta().SecurityLakeClient(ctx)

	// // TIP: -- 2. Fetch the plan
	// var plan, state subscriberNotificationResourceModel
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

	// 	in := &securitylake.UpdateSubscriberNotificationInput{
	// 		SubscriberNotificationId:   aws.String(plan.ID.ValueString()),
	// 		SubscriberNotificationName: aws.String(plan.Name.ValueString()),
	// 		SubscriberNotificationType: aws.String(plan.Type.ValueString()),
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
	// 	out, err := conn.UpdateSubscriberNotification(ctx, in)
	// 	if err != nil {
	// 		resp.Diagnostics.AddError(
	// 			create.ProblemStandardMessage(names.SecurityLake, create.ErrActionUpdating, ResNameSubscriberNotification, plan.ID.String(), err),
	// 			err.Error(),
	// 		)
	// 		return
	// 	}
	// 	if out == nil || out.SubscriberNotification == nil {
	// 		resp.Diagnostics.AddError(
	// 			create.ProblemStandardMessage(names.SecurityLake, create.ErrActionUpdating, ResNameSubscriberNotification, plan.ID.String(), nil),
	// 			errors.New("empty output").Error(),
	// 		)
	// 		return
	// 	}

	// 	// TIP: Using the output from the update function, re-set any computed attributes
	// 	plan.ARN = flex.StringToFramework(ctx, out.SubscriberNotification.Arn)
	// 	plan.ID = flex.StringToFramework(ctx, out.SubscriberNotification.SubscriberNotificationId)
	// }

	// // TIP: -- 5. Use a waiter to wait for update to complete
	// updateTimeout := r.UpdateTimeout(ctx, plan.Timeouts)
	// _, err := waitSubscriberNotificationUpdated(ctx, conn, plan.ID.ValueString(), updateTimeout)
	// if err != nil {
	// 	resp.Diagnostics.AddError(
	// 		create.ProblemStandardMessage(names.SecurityLake, create.ErrActionWaitingForUpdate, ResNameSubscriberNotification, plan.ID.String(), err),
	// 		err.Error(),
	// 	)
	// 	return
	// }

	// resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *subscriberNotificationResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := r.Meta().SecurityLakeClient(ctx)

	var state subscriberNotificationResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	in := &securitylake.DeleteSubscriberNotificationInput{}

	_, err := conn.DeleteSubscriberNotification(ctx, in)
	if err != nil {
		var nfe *awstypes.ResourceNotFoundException
		if errors.As(err, &nfe) {
			return
		}
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.SecurityLake, create.ErrActionDeleting, ResNameSubscriberNotification, state.ID.String(), err),
			err.Error(),
		)
		return
	}
}

func (r *subscriberNotificationResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// func findSubscriberNotificationByID(ctx context.Context, conn *securitylake.Client, id string) (*securitylake.SubscriberNotification, error) {
// 	in := &securitylake.GetSubscriberNotificationInput{
// 		Id: aws.String(id),
// 	}

// 	out, err := conn.GetSubscriberNotification(ctx, in)
// 	if err != nil {
// 		var nfe *awstypes.ResourceNotFoundException
// 		if errors.As(err, &nfe) {
// 			return nil, &retry.NotFoundError{
// 				LastError:   err,
// 				LastRequest: in,
// 			}
// 		}

// 		return nil, err
// 	}

// 	if out == nil || out.SubscriberNotification == nil {
// 		return nil, tfresource.NewEmptyResultError(in)
// 	}

// 	return out.SubscriberNotification, nil
// }

type subscriberNotificationResourceModel struct {
	Configuration fwtypes.ListNestedObjectValueOf[subscriberNotificationResourceConfigurationModel] `tfsdk:"configuration"`
	EndointID     types.String                                                                      `tfsdk:"endpoint_id"`
	SubscriberID  types.String                                                                      `tfsdk:"subscriber_id"`
	ID            types.String                                                                      `tfsdk:"id"`
}

type subscriberNotificationResourceConfigurationModel struct {
	SqsNotificationConfiguration   types.String                                                         `tfsdk:"sqs_notification_configuration"`
	HTTPSNotificationConfiguration fwtypes.ListNestedObjectValueOf[HTTPSNotificationConfigurationModel] `tfsdk:"https_notification_configuration"`
}

type HTTPSNotificationConfigurationModel struct {
	Endpoint                 types.String `tfsdk:"endpoint"`
	TargetRoleArn            types.String `tfsdk:"target_role_arn"`
	AuthorizationApiKeyName  types.String `tfsdk:"authorization_api_key_name"`
	AuthorizationApiKeyValue types.String `tfsdk:"authorization_api_key_value"`
	HTTPMethod               types.String `tfsdk:"http_method"`
}
