// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package eventbridgev2

import (
	"context"
	"fmt"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/YakDriver/smarterr"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/eventbridgev2"
	awstypes "github.com/aws/aws-sdk-go-v2/service/eventbridgev2/types"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// The service rejects concurrent changes to subscribers and event sources of the same event bus.
const concurrentModificationTimeout = 2 * time.Minute

// @FrameworkResource("aws_eventbridgev2_event_source", name="Event Source")
// @Tags(identifierAttribute="arn")
// @ArnIdentity
// @Testing(existsType="github.com/aws/aws-sdk-go-v2/service/eventbridgev2;eventbridgev2.DescribeEventSourceOutput")
// @Testing(preCheck="testAccPreCheck")
// @Testing(hasNoPreExistingResource=true)
func newEventSourceResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	return &eventSourceResource{}, nil
}

type eventSourceResource struct {
	framework.ResourceWithModel[eventSourceResourceModel]
	framework.WithImportByIdentity
}

func (r *eventSourceResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	onFailureConfigurationBlock := schema.ListNestedBlock{
		CustomType: fwtypes.NewListNestedObjectTypeOf[onFailureConfigurationModel](ctx),
		Validators: []validator.List{
			listvalidator.SizeAtMost(1),
		},
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				names.AttrARN: schema.StringAttribute{
					CustomType: fwtypes.ARNType,
					Required:   true,
				},
			},
		},
	}

	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrARN: framework.ARNAttributeComputedOnly(),
			names.AttrCreationTime: schema.StringAttribute{
				CustomType: timetypes.RFC3339Type{},
				Computed:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			names.AttrDescription: schema.StringAttribute{
				Optional: true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"event_bus_arn": schema.StringAttribute{
				CustomType: fwtypes.ARNType,
				Required:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"last_modified_time": schema.StringAttribute{
				CustomType: timetypes.RFC3339Type{},
				Computed:   true,
			},
			names.AttrName: schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 256),
					stringvalidator.RegexMatches(regexache.MustCompile(`^[0-9A-Za-z][0-9A-Za-z._-]*$`), "must start with an alphanumeric character and contain only alphanumeric characters, periods, hyphens, and underscores"),
				},
			},
			"revoked": schema.BoolAttribute{
				Computed: true,
			},
			names.AttrState: schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.EventSourceState](),
				Computed:   true,
			},
			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
		},
		Blocks: map[string]schema.Block{
			names.AttrConfiguration: schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[eventSourceConfigurationModel](ctx),
				Validators: []validator.List{
					listvalidator.IsRequired(),
					listvalidator.SizeAtLeast(1),
					listvalidator.SizeAtMost(1),
				},
				NestedObject: schema.NestedBlockObject{
					Blocks: map[string]schema.Block{
						"aws_service_events_configuration": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[awsServiceEventsSourceConfigurationModel](ctx),
							PlanModifiers: []planmodifier.List{
								requiresReplaceOnPresenceChange(),
							},
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
								listvalidator.ExactlyOneOf(
									path.MatchRelative().AtParent().AtName("aws_service_events_configuration"),
									path.MatchRelative().AtParent().AtName("partner_events_configuration"),
								),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"aws_service": schema.StringAttribute{
										Required: true,
									},
									"pattern": schema.StringAttribute{
										CustomType: jsontypes.NormalizedType{},
										Optional:   true,
									},
								},
								Blocks: map[string]schema.Block{
									"on_failure_configuration": onFailureConfigurationBlock,
								},
							},
						},
						"partner_events_configuration": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[partnerEventsSourceConfigurationModel](ctx),
							PlanModifiers: []planmodifier.List{
								requiresReplaceOnPresenceChange(),
							},
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"partner_bus_kms_key_identifier": schema.StringAttribute{
										Optional: true,
									},
									"partner_event_source_arn": schema.StringAttribute{
										Required: true,
									},
									"pattern": schema.StringAttribute{
										CustomType: jsontypes.NormalizedType{},
										Optional:   true,
									},
								},
								Blocks: map[string]schema.Block{
									"on_failure_configuration": onFailureConfigurationBlock,
								},
							},
						},
					},
				},
			},
		},
	}
}

func (r *eventSourceResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := r.Meta().EventBridgeV2Client(ctx)

	var plan eventSourceResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Plan.Get(ctx, &plan))
	if resp.Diagnostics.HasError() {
		return
	}

	name := fwflex.StringValueFromFramework(ctx, plan.Name)
	var input eventbridgev2.CreateEventSourceInput
	smerr.AddEnrich(ctx, &resp.Diagnostics, fwflex.Expand(ctx, plan, &input, fwflex.WithFieldNamePrefix("EventSource")))
	if resp.Diagnostics.HasError() {
		return
	}

	input.Tags = getTagsIn(ctx)

	out, err := tfresource.RetryWhenIsA[*eventbridgev2.CreateEventSourceOutput, *awstypes.ConcurrentModificationException](ctx, concurrentModificationTimeout, func(ctx context.Context) (*eventbridgev2.CreateEventSourceOutput, error) {
		return conn.CreateEventSource(ctx, &input)
	})
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.Name, name)
		return
	}

	arn := aws.ToString(out.EventSourceArn)

	// Event sources are provisioned synchronously, so read back the full configuration immediately.
	output, err := findEventSourceByARN(ctx, conn, arn)
	if err != nil {
		// Taint the resource.
		resp.State.SetAttribute(ctx, path.Root(names.AttrARN), arn)
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, arn)
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, r.flatten(ctx, output, &plan))
	if resp.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, plan))
}

func (r *eventSourceResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().EventBridgeV2Client(ctx)

	var state eventSourceResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	arn := fwflex.StringValueFromFramework(ctx, state.ARN)
	out, err := findEventSourceByARN(ctx, conn, arn)
	if retry.NotFound(err) {
		resp.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, arn)
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, r.flatten(ctx, out, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &state))
}

func (r *eventSourceResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	conn := r.Meta().EventBridgeV2Client(ctx)

	var plan, state eventSourceResourceModel
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
		arn := fwflex.StringValueFromFramework(ctx, plan.ARN)
		var input eventbridgev2.UpdateEventSourceInput
		smerr.AddEnrich(ctx, &resp.Diagnostics, fwflex.Expand(ctx, plan, &input, fwflex.WithFieldNamePrefix("EventSource")))
		if resp.Diagnostics.HasError() {
			return
		}

		_, err := tfresource.RetryWhenIsA[*eventbridgev2.UpdateEventSourceOutput, *awstypes.ConcurrentModificationException](ctx, concurrentModificationTimeout, func(ctx context.Context) (*eventbridgev2.UpdateEventSourceOutput, error) {
			return conn.UpdateEventSource(ctx, &input)
		})
		if err != nil {
			smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, arn)
			return
		}

		output, err := findEventSourceByARN(ctx, conn, arn)
		if err != nil {
			smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, arn)
			return
		}

		smerr.AddEnrich(ctx, &resp.Diagnostics, r.flatten(ctx, output, &plan))
		if resp.Diagnostics.HasError() {
			return
		}
	} else {
		plan.LastModifiedTime = state.LastModifiedTime
		plan.Revoked = state.Revoked
		plan.State = state.State
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &plan))
}

func (r *eventSourceResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := r.Meta().EventBridgeV2Client(ctx)

	var state eventSourceResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	arn := fwflex.StringValueFromFramework(ctx, state.ARN)
	input := eventbridgev2.DeleteEventSourceInput{
		EventSourceArn: aws.String(arn),
	}

	_, err := tfresource.RetryWhenIsA[*eventbridgev2.DeleteEventSourceOutput, *awstypes.ConcurrentModificationException](ctx, concurrentModificationTimeout, func(ctx context.Context) (*eventbridgev2.DeleteEventSourceOutput, error) {
		return conn.DeleteEventSource(ctx, &input)
	})
	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return
	}
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, arn)
		return
	}
}

func (r *eventSourceResource) flatten(ctx context.Context, out *eventbridgev2.DescribeEventSourceOutput, data *eventSourceResourceModel) diag.Diagnostics {
	var diags diag.Diagnostics

	smerr.AddEnrich(ctx, &diags, fwflex.Flatten(ctx, out, data, fwflex.WithFieldNamePrefix("EventSource")))
	if diags.HasError() {
		return diags
	}

	// The service returns an empty description when none is set.
	if aws.ToString(out.Description) == "" {
		data.Description = types.StringNull()
	}

	return diags
}

func findEventSourceByARN(ctx context.Context, conn *eventbridgev2.Client, arn string) (*eventbridgev2.DescribeEventSourceOutput, error) {
	input := eventbridgev2.DescribeEventSourceInput{
		EventSourceArn: aws.String(arn),
	}

	out, err := conn.DescribeEventSource(ctx, &input)
	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, smarterr.NewError(&retry.NotFoundError{
			LastError: err,
		})
	}
	if err != nil {
		return nil, smarterr.NewError(err)
	}

	if out == nil {
		return nil, smarterr.NewError(tfresource.NewEmptyResultError())
	}

	return out, nil
}

// requiresReplaceOnPresenceChange forces replacement when a single-element block is added or removed.
// The event source type is part of the event source's ARN, so switching between configuration variants cannot be done in place.
func requiresReplaceOnPresenceChange() planmodifier.List {
	return listplanmodifier.RequiresReplaceIf(
		func(ctx context.Context, req planmodifier.ListRequest, resp *listplanmodifier.RequiresReplaceIfFuncResponse) {
			resp.RequiresReplace = (req.StateValue.Length(fwtypes.CollectionLengthUnhandledAsZero) == 0) != (req.PlanValue.Length(fwtypes.CollectionLengthUnhandledAsZero) == 0)
		},
		"Changing the event source configuration type requires replacement.",
		"Changing the event source configuration type requires replacement.",
	)
}

type eventSourceResourceModel struct {
	framework.WithRegionModel
	ARN              types.String                                                   `tfsdk:"arn"`
	Configuration    fwtypes.ListNestedObjectValueOf[eventSourceConfigurationModel] `tfsdk:"configuration"`
	CreationTime     timetypes.RFC3339                                              `tfsdk:"creation_time"`
	Description      types.String                                                   `tfsdk:"description"`
	EventBusARN      fwtypes.ARN                                                    `tfsdk:"event_bus_arn"`
	LastModifiedTime timetypes.RFC3339                                              `tfsdk:"last_modified_time"`
	Name             types.String                                                   `tfsdk:"name"`
	Revoked          types.Bool                                                     `tfsdk:"revoked"`
	State            fwtypes.StringEnum[awstypes.EventSourceState]                  `tfsdk:"state"`
	Tags             tftags.Map                                                     `tfsdk:"tags"`
	TagsAll          tftags.Map                                                     `tfsdk:"tags_all"`
}

type eventSourceConfigurationModel struct {
	AWSServiceEventsConfiguration fwtypes.ListNestedObjectValueOf[awsServiceEventsSourceConfigurationModel] `tfsdk:"aws_service_events_configuration"`
	PartnerEventsConfiguration    fwtypes.ListNestedObjectValueOf[partnerEventsSourceConfigurationModel]    `tfsdk:"partner_events_configuration"`
}

var (
	_ fwflex.Expander  = eventSourceConfigurationModel{}
	_ fwflex.Flattener = &eventSourceConfigurationModel{}
)

func (m eventSourceConfigurationModel) Expand(ctx context.Context) (any, diag.Diagnostics) {
	var diags diag.Diagnostics

	switch {
	case !m.AWSServiceEventsConfiguration.IsNull():
		data, d := m.AWSServiceEventsConfiguration.ToPtr(ctx)
		smerr.AddEnrich(ctx, &diags, d)
		if diags.HasError() {
			return nil, diags
		}

		var r awstypes.EventSourceConfigurationMemberAwsServiceEventsConfiguration
		smerr.AddEnrich(ctx, &diags, fwflex.Expand(ctx, data, &r.Value))
		if diags.HasError() {
			return nil, diags
		}

		return &r, diags

	case !m.PartnerEventsConfiguration.IsNull():
		data, d := m.PartnerEventsConfiguration.ToPtr(ctx)
		smerr.AddEnrich(ctx, &diags, d)
		if diags.HasError() {
			return nil, diags
		}

		var r awstypes.EventSourceConfigurationMemberPartnerEventsConfiguration
		smerr.AddEnrich(ctx, &diags, fwflex.Expand(ctx, data, &r.Value))
		if diags.HasError() {
			return nil, diags
		}

		return &r, diags
	}

	return nil, diags
}

func (m *eventSourceConfigurationModel) Flatten(ctx context.Context, v any) diag.Diagnostics {
	var diags diag.Diagnostics

	switch t := v.(type) {
	case awstypes.EventSourceConfigurationMemberAwsServiceEventsConfiguration:
		var data awsServiceEventsSourceConfigurationModel
		smerr.AddEnrich(ctx, &diags, fwflex.Flatten(ctx, t.Value, &data))
		if diags.HasError() {
			return diags
		}

		m.AWSServiceEventsConfiguration = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &data)
		m.PartnerEventsConfiguration = fwtypes.NewListNestedObjectValueOfNull[partnerEventsSourceConfigurationModel](ctx)

	case awstypes.EventSourceConfigurationMemberPartnerEventsConfiguration:
		var data partnerEventsSourceConfigurationModel
		smerr.AddEnrich(ctx, &diags, fwflex.Flatten(ctx, t.Value, &data))
		if diags.HasError() {
			return diags
		}

		m.AWSServiceEventsConfiguration = fwtypes.NewListNestedObjectValueOfNull[awsServiceEventsSourceConfigurationModel](ctx)
		m.PartnerEventsConfiguration = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &data)

	default:
		diags.AddError(
			"Unsupported Type",
			fmt.Sprintf("event source configuration flatten: %T", v),
		)
	}

	return diags
}

type awsServiceEventsSourceConfigurationModel struct {
	AWSService             types.String                                                 `tfsdk:"aws_service"`
	OnFailureConfiguration fwtypes.ListNestedObjectValueOf[onFailureConfigurationModel] `tfsdk:"on_failure_configuration"`
	Pattern                jsontypes.Normalized                                         `tfsdk:"pattern"`
}

type partnerEventsSourceConfigurationModel struct {
	OnFailureConfiguration     fwtypes.ListNestedObjectValueOf[onFailureConfigurationModel] `tfsdk:"on_failure_configuration"`
	PartnerBusKMSKeyIdentifier types.String                                                 `tfsdk:"partner_bus_kms_key_identifier"`
	PartnerEventSourceARN      types.String                                                 `tfsdk:"partner_event_source_arn"`
	Pattern                    jsontypes.Normalized                                         `tfsdk:"pattern"`
}

type onFailureConfigurationModel struct {
	ARN fwtypes.ARN `tfsdk:"arn"`
}
