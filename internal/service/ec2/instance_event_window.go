// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

package ec2

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/YakDriver/smarterr"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/hashicorp/aws-sdk-go-base/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/int32validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
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

// @FrameworkResource("aws_ec2_instance_event_window", name="Instance Event Window")
// @Tags(identifierAttribute="id")
// @IdentityAttribute("id")
// @Testing(hasNoPreExistingResource=true)
// @Testing(existsType="github.com/aws/aws-sdk-go-v2/service/ec2/types;awstypes;awstypes.InstanceEventWindow")
// @Testing(preCheck="testAccPreCheck")
func newInstanceEventWindowResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &instanceEventWindowResource{}

	r.SetDefaultCreateTimeout(30 * time.Minute)
	r.SetDefaultUpdateTimeout(30 * time.Minute)
	r.SetDefaultDeleteTimeout(30 * time.Minute)

	return r, nil
}

const (
	ResNameInstanceEventWindow = "Instance Event Window"
)

type instanceEventWindowResource struct {
	framework.ResourceWithModel[instanceEventWindowResourceModel]
	framework.WithTimeouts
	framework.WithImportByIdentity
}

func (r *instanceEventWindowResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrID: framework.IDAttribute(),
			names.AttrName: schema.StringAttribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			// cron_expression and time_ranges are mutually exclusive: an event window
			// is defined either as a cron expression or as a set of time ranges, but
			// never both. All times are UTC. Per AWS: each individual time range must
			// be at least 2 hours, and the combined time ranges must total at least 4
			// hours; see ValidateConfig, which checks this at plan time (AWS also
			// enforces it server-side). A maximum of 200 event windows can be created
			// per Region. See
			// https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/event-windows.html.
			"cron_expression": schema.StringAttribute{
				Optional: true,
				Validators: []validator.String{
					stringvalidator.ExactlyOneOf(path.MatchRoot("time_ranges")),
				},
			},
			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
		},
		Blocks: map[string]schema.Block{
			// See the cron_expression comment above: mutually exclusive with it.
			"time_ranges": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[instanceEventWindowTimeRangeModel](ctx),
				Validators: []validator.List{
					listvalidator.ExactlyOneOf(path.MatchRoot("cron_expression")),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"start_week_day": schema.StringAttribute{
							CustomType: fwtypes.StringEnumType[awstypes.WeekDay](),
							Required:   true,
						},
						"start_hour": schema.Int32Attribute{
							Required: true,
							Validators: []validator.Int32{
								int32validator.Between(0, 23),
							},
						},
						"end_week_day": schema.StringAttribute{
							CustomType: fwtypes.StringEnumType[awstypes.WeekDay](),
							Required:   true,
						},
						"end_hour": schema.Int32Attribute{
							Required: true,
							Validators: []validator.Int32{
								int32validator.Between(0, 23),
							},
						},
					},
				},
			},
			names.AttrTimeouts: timeouts.Block(ctx, timeouts.Opts{
				Create: true,
				Update: true,
				Delete: true,
			}),
		},
	}
}

func (r *instanceEventWindowResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var data instanceEventWindowResourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if data.TimeRanges.IsNull() || data.TimeRanges.IsUnknown() {
		return
	}

	timeRanges, diags := data.TimeRanges.ToSlice(ctx)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Per AWS: each individual time range must be at least 2 hours, and the
	// combined time ranges must total at least 4 hours. AWS also enforces this
	// server-side, but catching it at plan time gives a faster, clearer error.
	// See https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/event-windows.html.
	const (
		minRangeHours = 2
		minTotalHours = 4
		hoursPerWeek  = 7 * 24
	)

	var totalHours int32
	var hasUnknown bool
	for i, tr := range timeRanges {
		if tr.StartWeekDay.IsNull() || tr.StartWeekDay.IsUnknown() ||
			tr.EndWeekDay.IsNull() || tr.EndWeekDay.IsUnknown() ||
			tr.StartHour.IsNull() || tr.StartHour.IsUnknown() ||
			tr.EndHour.IsNull() || tr.EndHour.IsUnknown() {
			hasUnknown = true
			continue
		}

		startIdx, ok := weekDayIndex(awstypes.WeekDay(tr.StartWeekDay.ValueString()))
		if !ok {
			hasUnknown = true
			continue
		}
		endIdx, ok := weekDayIndex(awstypes.WeekDay(tr.EndWeekDay.ValueString()))
		if !ok {
			hasUnknown = true
			continue
		}

		startTotal := startIdx*24 + tr.StartHour.ValueInt32()
		endTotal := endIdx*24 + tr.EndHour.ValueInt32()
		duration := (endTotal - startTotal + hoursPerWeek) % hoursPerWeek

		if duration < minRangeHours {
			resp.Diagnostics.AddAttributeError(
				path.Root("time_ranges").AtListIndex(i),
				"Invalid Attribute Value",
				fmt.Sprintf("time range must be at least %d hours, got %d hour(s)", minRangeHours, duration),
			)
		}

		totalHours += duration
	}

	// Skip the aggregate check if any range couldn't be evaluated (e.g. it
	// references a value not yet known at plan time): partial totalHours would
	// otherwise produce a false "combined time_ranges" error.
	if !hasUnknown && totalHours < minTotalHours {
		resp.Diagnostics.AddAttributeError(
			path.Root("time_ranges"),
			"Invalid Attribute Value",
			fmt.Sprintf("combined time_ranges must total at least %d hours, got %d hour(s)", minTotalHours, totalHours),
		)
	}
}

// weekDayIndex returns the zero-based day-of-week index (sunday=0) for the
// awstypes.WeekDay enum, used to compute an hour-of-week for duration math.
func weekDayIndex(d awstypes.WeekDay) (int32, bool) {
	switch d {
	case awstypes.WeekDaySunday:
		return 0, true
	case awstypes.WeekDayMonday:
		return 1, true
	case awstypes.WeekDayTuesday:
		return 2, true
	case awstypes.WeekDayWednesday:
		return 3, true
	case awstypes.WeekDayThursday:
		return 4, true
	case awstypes.WeekDayFriday:
		return 5, true
	case awstypes.WeekDaySaturday:
		return 6, true
	default:
		return 0, false
	}
}

func (r *instanceEventWindowResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := r.Meta().EC2Client(ctx)

	var plan instanceEventWindowResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Plan.Get(ctx, &plan))
	if resp.Diagnostics.HasError() {
		return
	}

	var input ec2.CreateInstanceEventWindowInput
	smerr.AddEnrich(ctx, &resp.Diagnostics, fwflex.Expand(ctx, plan, &input))
	if resp.Diagnostics.HasError() {
		return
	}
	input.TagSpecifications = getTagSpecificationsIn(ctx, awstypes.ResourceTypeInstanceEventWindow)

	out, err := conn.CreateInstanceEventWindow(ctx, &input)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, plan.Name.String())
		return
	}
	if out == nil || out.InstanceEventWindow == nil {
		smerr.AddError(ctx, &resp.Diagnostics, errors.New("empty output"), smerr.ID, plan.Name.String())
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, fwflex.Flatten(ctx, out.InstanceEventWindow, &plan, fwflex.WithFieldNamePrefix("InstanceEventWindow")))
	if resp.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, plan))
}

func (r *instanceEventWindowResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().EC2Client(ctx)

	var state instanceEventWindowResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := findInstanceEventWindowByID(ctx, conn, state.ID.ValueString())
	if retry.NotFound(err) {
		resp.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, state.ID.String())
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, fwflex.Flatten(ctx, out, &state, fwflex.WithFieldNamePrefix("InstanceEventWindow")))
	if resp.Diagnostics.HasError() {
		return
	}

	setTagsOut(ctx, out.Tags)

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &state))
}

func (r *instanceEventWindowResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	conn := r.Meta().EC2Client(ctx)

	var plan, state instanceEventWindowResourceModel
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
		var input ec2.ModifyInstanceEventWindowInput
		smerr.AddEnrich(ctx, &resp.Diagnostics, fwflex.Expand(ctx, plan, &input))
		if resp.Diagnostics.HasError() {
			return
		}
		input.InstanceEventWindowId = plan.ID.ValueStringPointer()

		out, err := conn.ModifyInstanceEventWindow(ctx, &input)
		if err != nil {
			smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, plan.ID.String())
			return
		}
		if out == nil || out.InstanceEventWindow == nil {
			smerr.AddError(ctx, &resp.Diagnostics, errors.New("empty output"), smerr.ID, plan.ID.String())
			return
		}

		smerr.AddEnrich(ctx, &resp.Diagnostics, fwflex.Flatten(ctx, out.InstanceEventWindow, &plan, fwflex.WithFieldNamePrefix("InstanceEventWindow")))
		if resp.Diagnostics.HasError() {
			return
		}
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &plan))
}

func (r *instanceEventWindowResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := r.Meta().EC2Client(ctx)

	var state instanceEventWindowResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	input := ec2.DeleteInstanceEventWindowInput{
		InstanceEventWindowId: state.ID.ValueStringPointer(),
		ForceDelete:           aws.Bool(true),
	}

	_, err := conn.DeleteInstanceEventWindow(ctx, &input)
	if err != nil {
		if tfawserr.ErrCodeEquals(err, errCodeInvalidInstanceEventWindowIDNotFound) {
			return
		}

		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, state.ID.String())
		return
	}

	deleteTimeout := r.DeleteTimeout(ctx, state.Timeouts)
	_, err = waitInstanceEventWindowDeleted(ctx, conn, state.ID.ValueString(), deleteTimeout)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, state.ID.String())
		return
	}
}

func waitInstanceEventWindowDeleted(ctx context.Context, conn *ec2.Client, id string, timeout time.Duration) (*awstypes.InstanceEventWindow, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.InstanceEventWindowStateDeleting),
		Target:  []string{},
		Refresh: statusInstanceEventWindow(conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*awstypes.InstanceEventWindow); ok {
		return out, smarterr.NewError(err)
	}

	return nil, smarterr.NewError(err)
}

func statusInstanceEventWindow(conn *ec2.Client, id string) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		out, err := findInstanceEventWindowByID(ctx, conn, id)
		if retry.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", smarterr.NewError(err)
		}

		return out, string(out.State), nil
	}
}

func findInstanceEventWindowByID(ctx context.Context, conn *ec2.Client, id string) (*awstypes.InstanceEventWindow, error) {
	input := ec2.DescribeInstanceEventWindowsInput{
		InstanceEventWindowIds: []string{id},
	}

	return findInstanceEventWindow(ctx, conn, &input)
}

func findInstanceEventWindow(ctx context.Context, conn *ec2.Client, input *ec2.DescribeInstanceEventWindowsInput) (*awstypes.InstanceEventWindow, error) {
	output, err := findInstanceEventWindows(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findInstanceEventWindows(ctx context.Context, conn *ec2.Client, input *ec2.DescribeInstanceEventWindowsInput) ([]awstypes.InstanceEventWindow, error) {
	var output []awstypes.InstanceEventWindow

	pages := ec2.NewDescribeInstanceEventWindowsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if tfawserr.ErrCodeEquals(err, errCodeInvalidInstanceEventWindowIDNotFound) {
			return nil, &retry.NotFoundError{
				LastError: err,
			}
		}

		if err != nil {
			return nil, smarterr.NewError(err)
		}

		output = append(output, page.InstanceEventWindows...)
	}

	if len(output) == 0 {
		return nil, tfresource.NewEmptyResultError()
	}

	return output, nil
}

type instanceEventWindowResourceModel struct {
	framework.WithRegionModel
	CronExpression types.String                                                       `tfsdk:"cron_expression"`
	ID             types.String                                                       `tfsdk:"id"`
	Name           types.String                                                       `tfsdk:"name"`
	Tags           tftags.Map                                                         `tfsdk:"tags"`
	TagsAll        tftags.Map                                                         `tfsdk:"tags_all"`
	Timeouts       timeouts.Value                                                     `tfsdk:"timeouts"`
	TimeRanges     fwtypes.ListNestedObjectValueOf[instanceEventWindowTimeRangeModel] `tfsdk:"time_ranges"`
}

type instanceEventWindowTimeRangeModel struct {
	StartWeekDay fwtypes.StringEnum[awstypes.WeekDay] `tfsdk:"start_week_day"`
	StartHour    types.Int32                          `tfsdk:"start_hour"`
	EndWeekDay   fwtypes.StringEnum[awstypes.WeekDay] `tfsdk:"end_week_day"`
	EndHour      types.Int32                          `tfsdk:"end_hour"`
}
