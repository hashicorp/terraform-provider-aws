// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

package pinpointsmsvoicev2

import (
	"context"
	"fmt"
	"strings"

	"github.com/YakDriver/regexache"
	awstypes "github.com/aws/aws-sdk-go-v2/service/pinpointsmsvoicev2/types"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/resourcevalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	intflex "github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_pinpointsmsvoicev2_configuration_set_event_destination", name="Configuration Set Event Destination")
// @IdentityAttribute("configuration_set_name")
// @IdentityAttribute("event_destination_name")
// @ImportIDHandler("configurationSetEventDestinationImportID")
// @Testing(hasNoPreExistingResource=true)
// @Testing(importStateIdAttributes="configuration_set_name;event_destination_name", importStateIdAttributesSep="flex.ResourceIdSeparator")
func newConfigurationSetEventDestinationResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	return &configurationSetEventDestinationResource{}, nil
}

type configurationSetEventDestinationResource struct {
	framework.ResourceWithModel[configurationSetEventDestinationResourceModel]
	framework.WithImportByIdentity
}

func (r *configurationSetEventDestinationResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
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
				Description: "Name of the configuration set this event destination belongs to. Pass the bare name; AWS's CreateEventDestination rejects the ARN form with ResourceNotFoundException.",
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
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"event_destination_name": schema.StringAttribute{
				Description: "Name of the event destination. Unique within the configuration set.",
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
				Description: "Event types for which the destination receives records. `TEXT_SENT` is not supported and is rejected by the AWS API at apply time.",
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
							Description: "ARN of the CloudWatch log group that receives the events.",
							CustomType:  fwtypes.ARNType,
							Required:    true,
						},
					},
				},
			},
			"kinesis_firehose_destination": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[kinesisFirehoseDestinationModel](ctx),
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"delivery_stream_arn": schema.StringAttribute{
							Description: "ARN of the Firehose delivery stream that receives the events.",
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
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						names.AttrTopicARN: schema.StringAttribute{
							Description: "ARN of the SNS topic that receives the events.",
							CustomType:  fwtypes.ARNType,
							Required:    true,
						},
					},
				},
			},
		},
	}
}

func (r *configurationSetEventDestinationResource) ConfigValidators(_ context.Context) []resource.ConfigValidator {
	return []resource.ConfigValidator{
		resourcevalidator.ExactlyOneOf(
			path.MatchRoot("cloudwatch_logs_destination"),
			path.MatchRoot("kinesis_firehose_destination"),
			path.MatchRoot("sns_destination"),
		),
	}
}

func (r *configurationSetEventDestinationResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// TODO: populated in the CRUD-wiring commit.
}

func (r *configurationSetEventDestinationResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// TODO: populated in the CRUD-wiring commit.
}

func (r *configurationSetEventDestinationResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// TODO: populated in the CRUD-wiring commit.
}

func (r *configurationSetEventDestinationResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// TODO: populated in the CRUD-wiring commit.
}

type configurationSetEventDestinationResourceModel struct {
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
	_ inttypes.ImportIDParser = configurationSetEventDestinationImportID{}
)

type configurationSetEventDestinationImportID struct{}

func (configurationSetEventDestinationImportID) Parse(id string) (string, map[string]any, error) {
	configurationSetName, eventDestinationName, found := strings.Cut(id, intflex.ResourceIdSeparator)
	if !found {
		return "", nil, fmt.Errorf("id %q should be in the format <configuration-set-name>%s<event-destination-name>", id, intflex.ResourceIdSeparator)
	}

	return id, map[string]any{
		"configuration_set_name": configurationSetName,
		"event_destination_name": eventDestinationName,
	}, nil
}
