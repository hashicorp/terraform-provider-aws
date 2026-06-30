// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

package pinpointsmsvoicev2

import (
	"context"
	"fmt"
	"strings"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/pinpointsmsvoicev2"
	awstypes "github.com/aws/aws-sdk-go-v2/service/pinpointsmsvoicev2/types"
	"github.com/hashicorp/aws-sdk-go-base/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/resourcevalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	intflex "github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_pinpointsmsvoicev2_event_destination", name="Event Destination")
// @IdentityAttribute("configuration_set_name")
// @IdentityAttribute("event_destination_name")
// @ImportIDHandler("eventDestinationImportID")
// @Testing(hasNoPreExistingResource=true)
// @Testing(importStateIdAttributes="configuration_set_name;event_destination_name", importStateIdAttributesSep="flex.ResourceIdSeparator")
// @Testing(tagsTest=false)
func newEventDestinationResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	return &eventDestinationResource{}, nil
}

type eventDestinationResource struct {
	framework.ResourceWithModel[eventDestinationResourceModel]
	framework.WithImportByIdentity
}

func (r *eventDestinationResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"configuration_set_arn": schema.StringAttribute{
				Description: "ARN of the parent configuration set.",
				CustomType:  fwtypes.ARNType,
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"configuration_set_name": schema.StringAttribute{
				Description: "Name of the configuration set this event destination belongs to.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 256),
					stringvalidator.RegexMatches(regexache.MustCompile(`^[A-Za-z0-9_:/-]+$`), "must contain only alphanumeric characters, underscores, hyphens, colons, and slashes"),
				},
			},
			names.AttrEnabled: schema.BoolAttribute{
				Description: "Whether the event destination is enabled. Defaults to `true`.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"event_destination_name": schema.StringAttribute{
				Description: "Name of the event destination.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 64),
					stringvalidator.RegexMatches(regexache.MustCompile(`^[A-Za-z0-9_-]+$`), "must contain only alphanumeric characters, underscores, and hyphens"),
				},
			},
			"matching_event_types": schema.SetAttribute{
				Description: "Event types for which the destination receives records.",
				CustomType:  fwtypes.NewSetTypeOf[fwtypes.StringEnum[awstypes.EventType]](ctx),
				Required:    true,
				ElementType: fwtypes.StringEnumType[awstypes.EventType](),
				Validators: []validator.Set{
					setvalidator.SizeAtLeast(1),
				},
			},
		},
		Blocks: map[string]schema.Block{
			"cloudwatch_logs_destination": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[cloudWatchLogsDestinationModel](ctx),
				PlanModifiers: []planmodifier.List{
					listplanmodifier.RequiresReplaceIf(sinkTypeSwapRequiresReplace, sinkTypeSwapReplaceDescription, sinkTypeSwapReplaceDescription),
				},
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						names.AttrIAMRoleARN: schema.StringAttribute{
							Description: "ARN of the IAM role that End User Messaging SMS assumes to write to the log group.",
							CustomType:  fwtypes.ARNType,
							Required:    true,
						},
						"log_group_arn": schema.StringAttribute{
							Description: "ARN of the Amazon CloudWatch log group that receives the events.",
							CustomType:  fwtypes.ARNType,
							Required:    true,
						},
					},
				},
			},
			"kinesis_firehose_destination": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[kinesisFirehoseDestinationModel](ctx),
				PlanModifiers: []planmodifier.List{
					listplanmodifier.RequiresReplaceIf(sinkTypeSwapRequiresReplace, sinkTypeSwapReplaceDescription, sinkTypeSwapReplaceDescription),
				},
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"delivery_stream_arn": schema.StringAttribute{
							Description: "ARN of the Amazon Data Firehose delivery stream that receives the events.",
							CustomType:  fwtypes.ARNType,
							Required:    true,
						},
						names.AttrIAMRoleARN: schema.StringAttribute{
							Description: "ARN of the IAM role that End User Messaging SMS assumes to write to the delivery stream.",
							CustomType:  fwtypes.ARNType,
							Required:    true,
						},
					},
				},
			},
			"sns_destination": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[snsDestinationModel](ctx),
				PlanModifiers: []planmodifier.List{
					listplanmodifier.RequiresReplaceIf(sinkTypeSwapRequiresReplace, sinkTypeSwapReplaceDescription, sinkTypeSwapReplaceDescription),
				},
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						names.AttrTopicARN: schema.StringAttribute{
							Description: "ARN of the Amazon SNS topic that receives the events.",
							CustomType:  fwtypes.ARNType,
							Required:    true,
						},
					},
				},
			},
		},
	}
}

const sinkTypeSwapReplaceDescription = "Replaces the event destination if the sink type changes. AWS's UpdateEventDestination rejects sink-type changes with ConflictException (EVENT_DESTINATION_MISMATCH)."

// sinkTypeSwapRequiresReplace fires on populated to empty transitions of a sink block.
// Within-sink attribute changes flow through normal Update.
func sinkTypeSwapRequiresReplace(_ context.Context, req planmodifier.ListRequest, resp *listplanmodifier.RequiresReplaceIfFuncResponse) {
	resp.RequiresReplace = req.StateValue.IsNull() != req.PlanValue.IsNull()
}

func (r *eventDestinationResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := r.Meta().PinpointSMSVoiceV2Client(ctx)

	var plan eventDestinationResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Plan.Get(ctx, &plan))
	if resp.Diagnostics.HasError() {
		return
	}

	// Capture the user-requested Enabled before AWS overwrites plan via Flatten
	// below — CreateEventDestinationInput has no Enabled field and it is computed
	// by AWS Enabled=true on Create.
	planEnabled := plan.Enabled.ValueBool()

	input := &pinpointsmsvoicev2.CreateEventDestinationInput{
		ClientToken: aws.String(create.UniqueId(ctx)),
	}
	smerr.AddEnrich(ctx, &resp.Diagnostics, fwflex.Expand(ctx, plan, input))
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := tfresource.RetryWhen(ctx, iamPropagationTimeout,
		func(ctx context.Context) (*pinpointsmsvoicev2.CreateEventDestinationOutput, error) {
			return conn.CreateEventDestination(ctx, input)
		},
		func(err error) (bool, error) {
			if tfawserr.ErrMessageContains(err, "ValidationException", "Could not access Kinesis Firehose Stream") ||
				tfawserr.ErrMessageContains(err, "ValidationException", "Could not assume IAM role") {
				return true, err
			}
			return false, err
		},
	)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, plan.EventDestinationName.ValueString())
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, fwflex.Flatten(ctx, out.EventDestination, &plan))
	if resp.Diagnostics.HasError() {
		return
	}
	plan.ConfigurationSetARN = fwtypes.ARNValue(aws.ToString(out.ConfigurationSetArn))

	if !planEnabled {
		updateInput := pinpointsmsvoicev2.UpdateEventDestinationInput{
			ConfigurationSetName: plan.ConfigurationSetName.ValueStringPointer(),
			EventDestinationName: plan.EventDestinationName.ValueStringPointer(),
			Enabled:              aws.Bool(false),
		}
		updateOut, err := conn.UpdateEventDestination(ctx, &updateInput)
		if err != nil {
			smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, plan.EventDestinationName.ValueString())
			return
		}
		smerr.AddEnrich(ctx, &resp.Diagnostics, fwflex.Flatten(ctx, updateOut.EventDestination, &plan))
		if resp.Diagnostics.HasError() {
			return
		}
		plan.ConfigurationSetARN = fwtypes.ARNValue(aws.ToString(updateOut.ConfigurationSetArn))
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &plan))
}

func (r *eventDestinationResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().PinpointSMSVoiceV2Client(ctx)

	var state eventDestinationResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	eventDestination, configurationSetArn, err := findEventDestinationByTwoPartKey(ctx, conn, state.ConfigurationSetName.ValueString(), state.EventDestinationName.ValueString())
	if retry.NotFound(err) {
		resp.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, state.EventDestinationName.ValueString())
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, fwflex.Flatten(ctx, eventDestination, &state))
	if resp.Diagnostics.HasError() {
		return
	}
	state.ConfigurationSetARN = fwtypes.ARNValue(aws.ToString(configurationSetArn))

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &state))
}

func (r *eventDestinationResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	conn := r.Meta().PinpointSMSVoiceV2Client(ctx)

	var plan, state eventDestinationResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Plan.Get(ctx, &plan))
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	diff, d := fwflex.Diff(ctx, plan, state)
	smerr.AddEnrich(ctx, &resp.Diagnostics, d)
	if resp.Diagnostics.HasError() {
		return
	}

	if diff.HasChanges() {
		var input pinpointsmsvoicev2.UpdateEventDestinationInput
		smerr.AddEnrich(ctx, &resp.Diagnostics, fwflex.Expand(ctx, plan, &input, diff.IgnoredFieldNamesOpts()...))
		if resp.Diagnostics.HasError() {
			return
		}
		input.ConfigurationSetName = plan.ConfigurationSetName.ValueStringPointer()
		input.EventDestinationName = plan.EventDestinationName.ValueStringPointer()

		out, err := tfresource.RetryWhen(ctx, iamPropagationTimeout,
			func(ctx context.Context) (*pinpointsmsvoicev2.UpdateEventDestinationOutput, error) {
				return conn.UpdateEventDestination(ctx, &input)
			},
			func(err error) (bool, error) {
				if tfawserr.ErrMessageContains(err, "ValidationException", "Could not access Kinesis Firehose Stream") ||
					tfawserr.ErrMessageContains(err, "ValidationException", "Could not assume IAM role") {
					return true, err
				}
				return false, err
			},
		)
		if err != nil {
			smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, plan.EventDestinationName.ValueString())
			return
		}

		smerr.AddEnrich(ctx, &resp.Diagnostics, fwflex.Flatten(ctx, out.EventDestination, &plan))
		if resp.Diagnostics.HasError() {
			return
		}
		plan.ConfigurationSetARN = fwtypes.ARNValue(aws.ToString(out.ConfigurationSetArn))
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &plan))
}

func (r *eventDestinationResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := r.Meta().PinpointSMSVoiceV2Client(ctx)

	var state eventDestinationResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	input := &pinpointsmsvoicev2.DeleteEventDestinationInput{
		ConfigurationSetName: state.ConfigurationSetName.ValueStringPointer(),
		EventDestinationName: state.EventDestinationName.ValueStringPointer(),
	}

	_, err := conn.DeleteEventDestination(ctx, input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return
	}

	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, state.EventDestinationName.ValueString())
		return
	}
}

func (r *eventDestinationResource) ConfigValidators(_ context.Context) []resource.ConfigValidator {
	return []resource.ConfigValidator{
		resourcevalidator.ExactlyOneOf(
			path.MatchRoot("cloudwatch_logs_destination"),
			path.MatchRoot("kinesis_firehose_destination"),
			path.MatchRoot("sns_destination"),
		),
	}
}

func findEventDestinationByTwoPartKey(ctx context.Context, conn *pinpointsmsvoicev2.Client, configurationSetName, eventDestinationName string) (*awstypes.EventDestination, *string, error) {
	parent, err := findConfigurationSetByID(ctx, conn, configurationSetName)
	if err != nil {
		return nil, nil, err
	}

	for i, ed := range parent.EventDestinations {
		if aws.ToString(ed.EventDestinationName) == eventDestinationName {
			return &parent.EventDestinations[i], parent.ConfigurationSetArn, nil
		}
	}

	return nil, nil, &retry.NotFoundError{
		Message: fmt.Sprintf("event destination %q not found in configuration set %q", eventDestinationName, configurationSetName),
	}
}

type eventDestinationResourceModel struct {
	framework.WithRegionModel
	CloudWatchLogsDestination  fwtypes.ListNestedObjectValueOf[cloudWatchLogsDestinationModel]  `tfsdk:"cloudwatch_logs_destination"`
	ConfigurationSetARN        fwtypes.ARN                                                      `tfsdk:"configuration_set_arn"`
	ConfigurationSetName       types.String                                                     `tfsdk:"configuration_set_name"`
	Enabled                    types.Bool                                                       `tfsdk:"enabled"`
	EventDestinationName       types.String                                                     `tfsdk:"event_destination_name"`
	KinesisFirehoseDestination fwtypes.ListNestedObjectValueOf[kinesisFirehoseDestinationModel] `tfsdk:"kinesis_firehose_destination"`
	MatchingEventTypes         fwtypes.SetOfStringEnum[awstypes.EventType]                      `tfsdk:"matching_event_types"`
	SnsDestination             fwtypes.ListNestedObjectValueOf[snsDestinationModel]             `tfsdk:"sns_destination"`
}

type cloudWatchLogsDestinationModel struct {
	IAMRoleARN  fwtypes.ARN `tfsdk:"iam_role_arn"`
	LogGroupARN fwtypes.ARN `tfsdk:"log_group_arn"`
}

type kinesisFirehoseDestinationModel struct {
	DeliveryStreamARN fwtypes.ARN `tfsdk:"delivery_stream_arn"`
	IAMRoleARN        fwtypes.ARN `tfsdk:"iam_role_arn"`
}

type snsDestinationModel struct {
	TopicARN fwtypes.ARN `tfsdk:"topic_arn"`
}

var (
	_ inttypes.ImportIDParser = eventDestinationImportID{}
)

type eventDestinationImportID struct{}

func (eventDestinationImportID) Parse(id string) (string, map[string]any, error) {
	configurationSetName, eventDestinationName, found := strings.Cut(id, intflex.ResourceIdSeparator)
	if !found {
		return "", nil, fmt.Errorf("id %q should be in the format <configuration-set-name>%s<event-destination-name>", id, intflex.ResourceIdSeparator)
	}

	return id, map[string]any{
		"configuration_set_name": configurationSetName,
		"event_destination_name": eventDestinationName,
	}, nil
}
