// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package eventbridgev2

import (
	"context"
	"errors"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/YakDriver/smarterr"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/eventbridgev2"
	awstypes "github.com/aws/aws-sdk-go-v2/service/eventbridgev2/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
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

// @FrameworkResource("aws_eventbridgev2_event_bus", name="Event Bus")
// @Tags(identifierAttribute="arn")
// @ArnIdentity
// @Testing(existsType="github.com/aws/aws-sdk-go-v2/service/eventbridgev2;eventbridgev2.DescribeEventBusOutput")
// @Testing(preCheck="testAccPreCheck")
// @Testing(hasNoPreExistingResource=true)
func newEventBusResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &eventBusResource{}

	r.SetDefaultCreateTimeout(10 * time.Minute)
	r.SetDefaultUpdateTimeout(10 * time.Minute)
	r.SetDefaultDeleteTimeout(10 * time.Minute)

	return r, nil
}

type eventBusResource struct {
	framework.ResourceWithModel[eventBusResourceModel]
	framework.WithTimeouts
	framework.WithImportByIdentity
}

func (r *eventBusResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
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
			// Not Computed: the service returns no encryption configuration for an event bus that uses an AWS owned key,
			// so removing the attribute from configuration is a change that reverts the event bus to an AWS owned key.
			names.AttrEncryptionConfiguration: schema.ObjectAttribute{
				CustomType: fwtypes.NewObjectTypeOf[encryptionConfigurationModel](ctx),
				Optional:   true,
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
			names.AttrState: schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.BusState](),
				Computed:   true,
			},
			"state_reason": schema.StringAttribute{
				Computed: true,
			},
			// The service applies a default retention period when none is configured.
			"storage_configuration": schema.ObjectAttribute{
				CustomType: fwtypes.NewObjectTypeOf[storageConfigurationModel](ctx),
				Optional:   true,
				Computed:   true,
				PlanModifiers: []planmodifier.Object{
					objectplanmodifier.UseStateForUnknown(),
				},
			},
			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
		},
		Blocks: map[string]schema.Block{
			names.AttrTimeouts: timeouts.Block(ctx, timeouts.Opts{
				Create: true,
				Update: true,
				Delete: true,
			}),
		},
	}
}

func (r *eventBusResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := r.Meta().EventBridgeV2Client(ctx)

	var plan eventBusResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Plan.Get(ctx, &plan))
	if resp.Diagnostics.HasError() {
		return
	}

	name := fwflex.StringValueFromFramework(ctx, plan.Name)
	var input eventbridgev2.CreateEventBusInput
	smerr.AddEnrich(ctx, &resp.Diagnostics, fwflex.Expand(ctx, plan, &input, fwflex.WithFieldNamePrefix("EventBus")))
	if resp.Diagnostics.HasError() {
		return
	}

	input.Tags = getTagsIn(ctx)

	out, err := conn.CreateEventBus(ctx, &input)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.Name, name)
		return
	}

	arn := aws.ToString(out.EventBusArn)

	output, err := waitEventBusCreated(ctx, conn, arn, r.CreateTimeout(ctx, plan.Timeouts))
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

func (r *eventBusResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().EventBridgeV2Client(ctx)

	var state eventBusResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	arn := fwflex.StringValueFromFramework(ctx, state.ARN)
	out, err := findEventBusByARN(ctx, conn, arn)
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

func (r *eventBusResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	conn := r.Meta().EventBridgeV2Client(ctx)

	var plan, state eventBusResourceModel
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
		var input eventbridgev2.UpdateEventBusInput
		smerr.AddEnrich(ctx, &resp.Diagnostics, fwflex.Expand(ctx, plan, &input, fwflex.WithFieldNamePrefix("EventBus")))
		if resp.Diagnostics.HasError() {
			return
		}

		// An omitted encryption configuration leaves the key unchanged; an empty one reverts to an AWS owned key.
		if plan.EncryptionConfiguration.IsNull() && !state.EncryptionConfiguration.IsNull() {
			input.EncryptionConfiguration = &awstypes.EncryptionConfiguration{}
		}

		_, err := conn.UpdateEventBus(ctx, &input)
		if err != nil {
			smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, arn)
			return
		}

		output, err := waitEventBusUpdated(ctx, conn, arn, r.UpdateTimeout(ctx, plan.Timeouts))
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
		plan.State = state.State
		plan.StateReason = state.StateReason
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &plan))
}

func (r *eventBusResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := r.Meta().EventBridgeV2Client(ctx)

	var state eventBusResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	arn := fwflex.StringValueFromFramework(ctx, state.ARN)
	input := eventbridgev2.DeleteEventBusInput{
		EventBusArn: aws.String(arn),
	}

	_, err := conn.DeleteEventBus(ctx, &input)
	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return
	}
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, arn)
		return
	}

	if _, err := waitEventBusDeleted(ctx, conn, arn, r.DeleteTimeout(ctx, state.Timeouts)); err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, arn)
		return
	}
}

func (r *eventBusResource) flatten(ctx context.Context, out *eventbridgev2.DescribeEventBusOutput, data *eventBusResourceModel) diag.Diagnostics {
	var diags diag.Diagnostics

	smerr.AddEnrich(ctx, &diags, fwflex.Flatten(ctx, out, data, fwflex.WithFieldNamePrefix("EventBus")))
	if diags.HasError() {
		return diags
	}

	// The service returns an empty description when none is set.
	if aws.ToString(out.Description) == "" {
		data.Description = types.StringNull()
	}

	return diags
}

func waitEventBusCreated(ctx context.Context, conn *eventbridgev2.Client, arn string, timeout time.Duration) (*eventbridgev2.DescribeEventBusOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.BusStateCreating),
		Target:  enum.Slice(awstypes.BusStateActive),
		Refresh: statusEventBus(conn, arn),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*eventbridgev2.DescribeEventBusOutput); ok {
		retry.SetLastError(err, eventBusStateReasonError(out))
		return out, smarterr.NewError(err)
	}

	return nil, smarterr.NewError(err)
}

func waitEventBusUpdated(ctx context.Context, conn *eventbridgev2.Client, arn string, timeout time.Duration) (*eventbridgev2.DescribeEventBusOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.BusStateUpdating),
		Target:  enum.Slice(awstypes.BusStateActive),
		Refresh: statusEventBus(conn, arn),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*eventbridgev2.DescribeEventBusOutput); ok {
		retry.SetLastError(err, eventBusStateReasonError(out))
		return out, smarterr.NewError(err)
	}

	return nil, smarterr.NewError(err)
}

func waitEventBusDeleted(ctx context.Context, conn *eventbridgev2.Client, arn string, timeout time.Duration) (*eventbridgev2.DescribeEventBusOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.BusStateActive, awstypes.BusStateDeleting),
		Target:  []string{},
		Refresh: statusEventBus(conn, arn),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*eventbridgev2.DescribeEventBusOutput); ok {
		retry.SetLastError(err, eventBusStateReasonError(out))
		return out, smarterr.NewError(err)
	}

	return nil, smarterr.NewError(err)
}

func eventBusStateReasonError(out *eventbridgev2.DescribeEventBusOutput) error {
	if v := aws.ToString(out.StateReason); v != "" {
		return errors.New(v)
	}

	return nil
}

func statusEventBus(conn *eventbridgev2.Client, arn string) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		out, err := findEventBusByARN(ctx, conn, arn)
		if retry.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", smarterr.NewError(err)
		}

		return out, string(out.State), nil
	}
}

func findEventBusByARN(ctx context.Context, conn *eventbridgev2.Client, arn string) (*eventbridgev2.DescribeEventBusOutput, error) {
	input := eventbridgev2.DescribeEventBusInput{
		EventBusArn: aws.String(arn),
	}

	out, err := conn.DescribeEventBus(ctx, &input)
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

type eventBusResourceModel struct {
	framework.WithRegionModel
	ARN                     types.String                                        `tfsdk:"arn"`
	CreationTime            timetypes.RFC3339                                   `tfsdk:"creation_time"`
	Description             types.String                                        `tfsdk:"description"`
	EncryptionConfiguration fwtypes.ObjectValueOf[encryptionConfigurationModel] `tfsdk:"encryption_configuration"`
	LastModifiedTime        timetypes.RFC3339                                   `tfsdk:"last_modified_time"`
	Name                    types.String                                        `tfsdk:"name"`
	State                   fwtypes.StringEnum[awstypes.BusState]               `tfsdk:"state"`
	StateReason             types.String                                        `tfsdk:"state_reason"`
	StorageConfiguration    fwtypes.ObjectValueOf[storageConfigurationModel]    `tfsdk:"storage_configuration"`
	Tags                    tftags.Map                                          `tfsdk:"tags"`
	TagsAll                 tftags.Map                                          `tfsdk:"tags_all"`
	Timeouts                timeouts.Value                                      `tfsdk:"timeouts"`
}

type encryptionConfigurationModel struct {
	KMSKeyIdentifier types.String `tfsdk:"kms_key_identifier"`
}

type storageConfigurationModel struct {
	RetentionPeriodInDays types.Int32 `tfsdk:"retention_period_in_days"`
}
