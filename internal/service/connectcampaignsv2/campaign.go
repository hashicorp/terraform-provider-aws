// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package connectcampaignsv2

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/connectcampaignsv2"
	awstypes "github.com/aws/aws-sdk-go-v2/service/connectcampaignsv2/types"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_connectcampaignsv2_campaign", name="Campaign")
// @Tags(identifierAttribute="arn")
// @Testing(tagsTest=false)
func newCampaignResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	return &campaignResource{}, nil
}

type campaignResource struct {
	framework.ResourceWithModel[campaignResourceModel]
	framework.WithImportByID
}

func (r *campaignResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrARN: framework.ARNAttributeComputedOnly(),
			"connect_campaign_flow_arn": schema.StringAttribute{
				CustomType:  fwtypes.ARNType,
				Optional:    true,
				Description: "Amazon Resource Name (ARN) of the Amazon Connect campaign flow. Updating this value calls UpdateCampaignFlowAssociation; removing it after creation is not supported by this resource.",
			},
			"connect_instance_id": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			names.AttrID: framework.IDAttribute(),
			names.AttrName: schema.StringAttribute{
				Required: true,
			},
			names.AttrTags:    tftags.TagsAttributeForceNew(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
			"type": schema.StringAttribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
		Blocks: map[string]schema.Block{
			"entry_limits_config": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[entryLimitsConfigModel](ctx),
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"max_entry_count": schema.Int64Attribute{
							Required:    true,
							Description: "Maximum number of times a participant can enter the campaign. Use 0 for unlimited entries.",
						},
						"min_entry_interval": schema.StringAttribute{
							Required:    true,
							Description: "Minimum time interval that must pass before a participant can enter the campaign again, in ISO 8601 duration format.",
						},
					},
				},
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
				},
			},
			"schedule": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[scheduleModel](ctx),
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"end_time": schema.StringAttribute{
							Required:    true,
							Description: "Campaign end time in RFC3339 format.",
						},
						"refresh_frequency": schema.StringAttribute{
							Optional:    true,
							Description: "Campaign refresh frequency in ISO 8601 duration format.",
						},
						"start_time": schema.StringAttribute{
							Required:    true,
							Description: "Campaign start time in RFC3339 format.",
						},
					},
				},
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
				},
			},
			"source": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[sourceModel](ctx),
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"customer_profiles_segment_arn": schema.StringAttribute{
							Optional:    true,
							Description: "Customer Profiles segment ARN source for the campaign. Exactly one of this attribute or event_trigger must be configured.",
						},
					},
					Blocks: map[string]schema.Block{
						"event_trigger": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[eventTriggerModel](ctx),
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"customer_profiles_domain_arn": schema.StringAttribute{
										Optional:    true,
										Description: "Customer Profiles domain ARN for an event-triggered campaign source.",
									},
								},
							},
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
							},
						},
					},
				},
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
				},
			},
		},
	}
}

func (r *campaignResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data campaignResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().ConnectCampaignsV2Client(ctx)
	input := &connectcampaignsv2.CreateCampaignInput{
		ConnectCampaignFlowArn: data.ConnectCampaignFlowARN.ValueStringPointer(),
		ConnectInstanceId:      data.ConnectInstanceID.ValueStringPointer(),
		Name:                   data.Name.ValueStringPointer(),
		Tags:                   getTagsIn(ctx),
	}
	entryLimitsConfig, err := expandEntryLimitsConfig(ctx, data.EntryLimitsConfig)
	if err != nil {
		response.Diagnostics.AddAttributeError(path.Root("entry_limits_config"), "Invalid entry limits config", err.Error())
		return
	}
	input.EntryLimitsConfig = entryLimitsConfig
	schedule, err := expandSchedule(ctx, data.Schedule)
	if err != nil {
		response.Diagnostics.AddAttributeError(path.Root("schedule"), "Invalid schedule", err.Error())
		return
	}
	input.Schedule = schedule
	source, err := expandSource(ctx, data.Source)
	if err != nil {
		response.Diagnostics.AddAttributeError(path.Root("source"), "Invalid source", err.Error())
		return
	}
	input.Source = source
	if !data.Type.IsNull() && !data.Type.IsUnknown() {
		input.Type = awstypes.ExternalCampaignType(data.Type.ValueString())
	}

	output, err := conn.CreateCampaign(ctx, input)
	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("creating Connect Campaigns V2 Campaign (%s)", data.Name.ValueString()), err.Error())
		return
	}

	data.ARN = fwtypes.ARNValue(aws.ToString(output.Arn))
	data.ID = types.StringPointerValue(output.Id)

	response.Diagnostics.Append(response.State.Set(ctx, data)...)
}

func (r *campaignResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data campaignResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().ConnectCampaignsV2Client(ctx)
	campaign, err := findCampaignByID(ctx, conn, data.ID.ValueString())
	if retry.NotFound(err) {
		response.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading Connect Campaigns V2 Campaign (%s)", data.ID.ValueString()), err.Error())
		return
	}

	data.ARN = fwtypes.ARNValue(aws.ToString(campaign.Arn))
	if campaign.ConnectCampaignFlowArn != nil {
		data.ConnectCampaignFlowARN = fwtypes.ARNValue(aws.ToString(campaign.ConnectCampaignFlowArn))
	} else {
		data.ConnectCampaignFlowARN = fwtypes.ARNNull()
	}
	data.ConnectInstanceID = types.StringPointerValue(campaign.ConnectInstanceId)
	data.EntryLimitsConfig = flattenEntryLimitsConfig(ctx, campaign.EntryLimitsConfig)
	data.ID = types.StringPointerValue(campaign.Id)
	data.Name = types.StringPointerValue(campaign.Name)
	data.Schedule = flattenSchedule(ctx, campaign.Schedule)
	data.Source = flattenSource(ctx, campaign.Source)
	setTagsOut(ctx, campaign.Tags)
	if campaign.Type != "" {
		data.Type = types.StringValue(string(campaign.Type))
	}

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *campaignResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var plan, state campaignResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().ConnectCampaignsV2Client(ctx)
	id := state.ID.ValueString()

	if !plan.ConnectCampaignFlowARN.Equal(state.ConnectCampaignFlowARN) {
		if plan.ConnectCampaignFlowARN.IsNull() || plan.ConnectCampaignFlowARN.IsUnknown() {
			response.Diagnostics.AddAttributeError(path.Root("connect_campaign_flow_arn"), "Unsupported removal", "The Connect Campaigns V2 API supports updating a campaign flow association but does not document clearing it. Configure a replacement value or replace the resource.")
			return
		}
		_, err := conn.UpdateCampaignFlowAssociation(ctx, &connectcampaignsv2.UpdateCampaignFlowAssociationInput{
			ConnectCampaignFlowArn: plan.ConnectCampaignFlowARN.ValueStringPointer(),
			Id:                     aws.String(id),
		})
		if err != nil {
			response.Diagnostics.AddError(fmt.Sprintf("updating Connect Campaigns V2 Campaign (%s) flow association", id), err.Error())
			return
		}
	}

	if !plan.EntryLimitsConfig.Equal(state.EntryLimitsConfig) {
		entryLimitsConfig, err := expandEntryLimitsConfig(ctx, plan.EntryLimitsConfig)
		if err != nil {
			response.Diagnostics.AddAttributeError(path.Root("entry_limits_config"), "Invalid entry limits config", err.Error())
			return
		}
		if entryLimitsConfig == nil {
			_, err = conn.DeleteCampaignEntryLimits(ctx, &connectcampaignsv2.DeleteCampaignEntryLimitsInput{
				Id: aws.String(id),
			})
		} else {
			_, err = conn.UpdateCampaignEntryLimits(ctx, &connectcampaignsv2.UpdateCampaignEntryLimitsInput{
				EntryLimitsConfig: entryLimitsConfig,
				Id:                aws.String(id),
			})
		}
		if err != nil {
			response.Diagnostics.AddError(fmt.Sprintf("updating Connect Campaigns V2 Campaign (%s) entry limits", id), err.Error())
			return
		}
	}

	if !plan.Name.Equal(state.Name) {
		_, err := conn.UpdateCampaignName(ctx, &connectcampaignsv2.UpdateCampaignNameInput{
			Id:   aws.String(id),
			Name: plan.Name.ValueStringPointer(),
		})
		if err != nil {
			response.Diagnostics.AddError(fmt.Sprintf("updating Connect Campaigns V2 Campaign (%s) name", id), err.Error())
			return
		}
	}

	if !plan.Schedule.Equal(state.Schedule) {
		schedule, err := expandSchedule(ctx, plan.Schedule)
		if err != nil {
			response.Diagnostics.AddAttributeError(path.Root("schedule"), "Invalid schedule", err.Error())
			return
		}
		if schedule == nil {
			response.Diagnostics.AddAttributeError(path.Root("schedule"), "Unsupported removal", "The Connect Campaigns V2 API supports updating schedule but does not document clearing it. Configure a replacement value or replace the resource.")
			return
		}
		_, err = conn.UpdateCampaignSchedule(ctx, &connectcampaignsv2.UpdateCampaignScheduleInput{
			Id:       aws.String(id),
			Schedule: schedule,
		})
		if err != nil {
			response.Diagnostics.AddError(fmt.Sprintf("updating Connect Campaigns V2 Campaign (%s) schedule", id), err.Error())
			return
		}
	}

	if !plan.Source.Equal(state.Source) {
		source, err := expandSource(ctx, plan.Source)
		if err != nil {
			response.Diagnostics.AddAttributeError(path.Root("source"), "Invalid source", err.Error())
			return
		}
		if source == nil {
			response.Diagnostics.AddAttributeError(path.Root("source"), "Unsupported removal", "The Connect Campaigns V2 API supports updating campaign source but does not document clearing it. Configure a replacement value or replace the resource.")
			return
		}
		_, err = conn.UpdateCampaignSource(ctx, &connectcampaignsv2.UpdateCampaignSourceInput{
			Id:     aws.String(id),
			Source: source,
		})
		if err != nil {
			response.Diagnostics.AddError(fmt.Sprintf("updating Connect Campaigns V2 Campaign (%s) source", id), err.Error())
			return
		}
	}

	response.Diagnostics.Append(response.State.Set(ctx, plan)...)
}

func (r *campaignResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data campaignResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().ConnectCampaignsV2Client(ctx)
	_, err := conn.DeleteCampaign(ctx, &connectcampaignsv2.DeleteCampaignInput{
		Id: data.ID.ValueStringPointer(),
	})
	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return
	}
	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("deleting Connect Campaigns V2 Campaign (%s)", data.ID.ValueString()), err.Error())
		return
	}
}

func findCampaignByID(ctx context.Context, conn *connectcampaignsv2.Client, id string) (*awstypes.Campaign, error) {
	output, err := conn.DescribeCampaign(ctx, &connectcampaignsv2.DescribeCampaignInput{
		Id: aws.String(id),
	})
	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{LastError: err}
	}
	if err != nil {
		return nil, err
	}
	if output == nil || output.Campaign == nil {
		return nil, tfresource.NewEmptyResultError()
	}

	return output.Campaign, nil
}

func expandEntryLimitsConfig(ctx context.Context, value fwtypes.ListNestedObjectValueOf[entryLimitsConfigModel]) (*awstypes.EntryLimitsConfig, error) {
	if value.IsNull() || value.IsUnknown() {
		return nil, nil
	}

	model, diags := value.ToPtr(ctx)
	if diags.HasError() {
		return nil, fmt.Errorf("converting value: %s", diags.Errors())
	}
	if model == nil {
		return nil, nil
	}

	maxEntryCount := int32(model.MaxEntryCount.ValueInt64())
	return &awstypes.EntryLimitsConfig{
		MaxEntryCount:    aws.Int32(maxEntryCount),
		MinEntryInterval: model.MinEntryInterval.ValueStringPointer(),
	}, nil
}

func flattenEntryLimitsConfig(ctx context.Context, value *awstypes.EntryLimitsConfig) fwtypes.ListNestedObjectValueOf[entryLimitsConfigModel] {
	if value == nil {
		return fwtypes.NewListNestedObjectValueOfNull[entryLimitsConfigModel](ctx)
	}

	return fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &entryLimitsConfigModel{
		MaxEntryCount:    types.Int64Value(int64(aws.ToInt32(value.MaxEntryCount))),
		MinEntryInterval: types.StringPointerValue(value.MinEntryInterval),
	})
}

func expandSchedule(ctx context.Context, value fwtypes.ListNestedObjectValueOf[scheduleModel]) (*awstypes.Schedule, error) {
	if value.IsNull() || value.IsUnknown() {
		return nil, nil
	}

	model, diags := value.ToPtr(ctx)
	if diags.HasError() {
		return nil, fmt.Errorf("converting value: %s", diags.Errors())
	}
	if model == nil {
		return nil, nil
	}

	startTime, err := time.Parse(time.RFC3339, model.StartTime.ValueString())
	if err != nil {
		return nil, fmt.Errorf("start_time must be RFC3339: %w", err)
	}
	endTime, err := time.Parse(time.RFC3339, model.EndTime.ValueString())
	if err != nil {
		return nil, fmt.Errorf("end_time must be RFC3339: %w", err)
	}

	return &awstypes.Schedule{
		EndTime:          aws.Time(endTime),
		RefreshFrequency: model.RefreshFrequency.ValueStringPointer(),
		StartTime:        aws.Time(startTime),
	}, nil
}

func flattenSchedule(ctx context.Context, value *awstypes.Schedule) fwtypes.ListNestedObjectValueOf[scheduleModel] {
	if value == nil {
		return fwtypes.NewListNestedObjectValueOfNull[scheduleModel](ctx)
	}

	return fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &scheduleModel{
		EndTime:          types.StringValue(aws.ToTime(value.EndTime).Format(time.RFC3339)),
		RefreshFrequency: types.StringPointerValue(value.RefreshFrequency),
		StartTime:        types.StringValue(aws.ToTime(value.StartTime).Format(time.RFC3339)),
	})
}

func expandSource(ctx context.Context, value fwtypes.ListNestedObjectValueOf[sourceModel]) (awstypes.Source, error) {
	if value.IsNull() || value.IsUnknown() {
		return nil, nil
	}

	model, diags := value.ToPtr(ctx)
	if diags.HasError() {
		return nil, fmt.Errorf("converting value: %s", diags.Errors())
	}
	if model == nil {
		return nil, nil
	}

	trigger, err := expandEventTrigger(ctx, model.EventTrigger)
	if err != nil {
		return nil, err
	}
	segmentSet := !model.CustomerProfilesSegmentARN.IsNull() && !model.CustomerProfilesSegmentARN.IsUnknown() && model.CustomerProfilesSegmentARN.ValueString() != ""

	switch {
	case segmentSet && trigger != nil:
		return nil, fmt.Errorf("must set only one of customer_profiles_segment_arn or event_trigger")
	case segmentSet:
		return &awstypes.SourceMemberCustomerProfilesSegmentArn{Value: model.CustomerProfilesSegmentARN.ValueString()}, nil
	case trigger != nil:
		return &awstypes.SourceMemberEventTrigger{Value: *trigger}, nil
	default:
		return nil, fmt.Errorf("must set one of customer_profiles_segment_arn or event_trigger")
	}
}

func expandEventTrigger(ctx context.Context, value fwtypes.ListNestedObjectValueOf[eventTriggerModel]) (*awstypes.EventTrigger, error) {
	if value.IsNull() || value.IsUnknown() {
		return nil, nil
	}

	model, diags := value.ToPtr(ctx)
	if diags.HasError() {
		return nil, fmt.Errorf("converting event_trigger: %s", diags.Errors())
	}
	if model == nil {
		return nil, nil
	}

	return &awstypes.EventTrigger{
		CustomerProfilesDomainArn: model.CustomerProfilesDomainARN.ValueStringPointer(),
	}, nil
}

func flattenSource(ctx context.Context, value awstypes.Source) fwtypes.ListNestedObjectValueOf[sourceModel] {
	switch v := value.(type) {
	case nil:
		return fwtypes.NewListNestedObjectValueOfNull[sourceModel](ctx)
	case *awstypes.SourceMemberCustomerProfilesSegmentArn:
		return fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &sourceModel{
			CustomerProfilesSegmentARN: types.StringValue(v.Value),
			EventTrigger:               fwtypes.NewListNestedObjectValueOfNull[eventTriggerModel](ctx),
		})
	case *awstypes.SourceMemberEventTrigger:
		return fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &sourceModel{
			CustomerProfilesSegmentARN: types.StringNull(),
			EventTrigger: fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &eventTriggerModel{
				CustomerProfilesDomainARN: types.StringPointerValue(v.Value.CustomerProfilesDomainArn),
			}),
		})
	default:
		return fwtypes.NewListNestedObjectValueOfNull[sourceModel](ctx)
	}
}

type campaignResourceModel struct {
	framework.WithRegionModel
	ARN                    fwtypes.ARN                                             `tfsdk:"arn"`
	ConnectCampaignFlowARN fwtypes.ARN                                             `tfsdk:"connect_campaign_flow_arn"`
	ConnectInstanceID      types.String                                            `tfsdk:"connect_instance_id"`
	EntryLimitsConfig      fwtypes.ListNestedObjectValueOf[entryLimitsConfigModel] `tfsdk:"entry_limits_config"`
	ID                     types.String                                            `tfsdk:"id"`
	Name                   types.String                                            `tfsdk:"name"`
	Schedule               fwtypes.ListNestedObjectValueOf[scheduleModel]          `tfsdk:"schedule"`
	Source                 fwtypes.ListNestedObjectValueOf[sourceModel]            `tfsdk:"source"`
	Tags                   tftags.Map                                              `tfsdk:"tags"`
	TagsAll                tftags.Map                                              `tfsdk:"tags_all"`
	Type                   types.String                                            `tfsdk:"type"`
}

type entryLimitsConfigModel struct {
	MaxEntryCount    types.Int64  `tfsdk:"max_entry_count"`
	MinEntryInterval types.String `tfsdk:"min_entry_interval"`
}

type scheduleModel struct {
	EndTime          types.String `tfsdk:"end_time"`
	RefreshFrequency types.String `tfsdk:"refresh_frequency"`
	StartTime        types.String `tfsdk:"start_time"`
}

type sourceModel struct {
	CustomerProfilesSegmentARN types.String                                       `tfsdk:"customer_profiles_segment_arn"`
	EventTrigger               fwtypes.ListNestedObjectValueOf[eventTriggerModel] `tfsdk:"event_trigger"`
}

type eventTriggerModel struct {
	CustomerProfilesDomainARN types.String `tfsdk:"customer_profiles_domain_arn"`
}
