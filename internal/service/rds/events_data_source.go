// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package rds

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/rds"
	awstypes "github.com/aws/aws-sdk-go-v2/service/rds/types"
	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkDataSource("aws_rds_events", name="Events")
func newEventsDataSource(context.Context) (datasource.DataSourceWithConfigure, error) {
	return &eventsDataSource{}, nil
}

type eventsDataSource struct {
	framework.DataSourceWithModel[eventsDataSourceModel]
}

func (d *eventsDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"source_identifier": schema.StringAttribute{
				Optional: true,
				Validators: []validator.String{
					stringvalidator.AlsoRequires(path.MatchRoot(names.AttrSourceType)),
				},
			},
			names.AttrSourceType: schema.StringAttribute{
				Optional: true,
				Validators: []validator.String{
					enum.FrameworkValidate[awstypes.SourceType](),
				},
			},
			"event_categories": schema.SetAttribute{
				CustomType:  fwtypes.SetOfStringType,
				ElementType: types.StringType,
				Optional:    true,
			},
			"duration": schema.Int64Attribute{
				Optional: true,
				Validators: []validator.Int64{
					int64validator.ConflictsWith(path.MatchRoot("start_time"), path.MatchRoot("end_time")),
				},
			},
			"start_time": schema.StringAttribute{
				CustomType: timetypes.RFC3339Type{},
				Optional:   true,
				Validators: []validator.String{
					stringvalidator.ConflictsWith(path.MatchRoot("duration")),
				},
			},
			"end_time": schema.StringAttribute{
				CustomType: timetypes.RFC3339Type{},
				Optional:   true,
				Validators: []validator.String{
					stringvalidator.ConflictsWith(path.MatchRoot("duration")),
				},
			},
			"events": framework.DataSourceComputedListOfObjectAttribute[eventModel](ctx),
		},
	}
}

func (d *eventsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	conn := d.Meta().RDSClient(ctx)

	var data eventsDataSourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Config.Get(ctx, &data))
	if resp.Diagnostics.HasError() {
		return
	}

	var input rds.DescribeEventsInput
	smerr.AddEnrich(ctx, &resp.Diagnostics, flex.Expand(ctx, data, &input))
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := findEvents(ctx, conn, &input)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err)
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, flex.Flatten(ctx, out, &data.Events))
	if resp.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &data))
}

// findEvents returns all events matching input (paginated).
func findEvents(ctx context.Context, conn *rds.Client, input *rds.DescribeEventsInput) ([]awstypes.Event, error) {
	var output []awstypes.Event

	pages := rds.NewDescribeEventsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)
		if err != nil {
			return nil, err
		}

		output = append(output, page.Events...)
	}

	return output, nil
}

// findEventMessagesAfter returns message text for one source, in the given
// categories, since a start time. Categories are a parameter (not hardcoded)
// because the relevant categories differ per scenario.
func findEventMessagesAfter(ctx context.Context, conn *rds.Client, sourceID string, sourceType awstypes.SourceType, since time.Time, categories []string) ([]string, error) {
	input := &rds.DescribeEventsInput{
		SourceIdentifier: aws.String(sourceID),
		SourceType:       sourceType,
		EventCategories:  categories,
		StartTime:        aws.Time(since),
	}

	events, err := findEvents(ctx, conn, input)
	if err != nil {
		return nil, err
	}

	return tfslices.ApplyToAll(events, func(e awstypes.Event) string {
		return aws.ToString(e.Message)
	}), nil
}

var (
	// upgradeEventCategories covers both source types' upgrade-failure events.
	// The Aurora cluster pre-check failure (RDS-EVENT-0412) and the db-instance
	// upgrade-failed/rollback event (RDS-EVENT-0270) are both "maintenance";
	// some db-instance rollbacks (RDS-EVENT-0188) are "failure". A plain
	// "failure"-only filter would miss the Aurora pre-check case.
	upgradeEventCategories = []string{"failure", "maintenance"}

	// createTimeEventCategories is deliberately narrower than
	// upgradeEventCategories: the create-time enhanced-monitoring
	// configuration failure (#41037) is confirmed "failure"-only.
	createTimeEventCategories = []string{"failure"}
)

type eventsDataSourceModel struct {
	framework.WithRegionModel
	SourceIdentifier types.String                                `tfsdk:"source_identifier"`
	SourceType       types.String                                `tfsdk:"source_type"`
	EventCategories  fwtypes.SetOfString                         `tfsdk:"event_categories"`
	Duration         types.Int64                                 `tfsdk:"duration"`
	StartTime        timetypes.RFC3339                           `tfsdk:"start_time"`
	EndTime          timetypes.RFC3339                           `tfsdk:"end_time"`
	Events           fwtypes.ListNestedObjectValueOf[eventModel] `tfsdk:"events"`
}

type eventModel struct {
	Date             timetypes.RFC3339    `tfsdk:"date"`
	EventCategories  fwtypes.ListOfString `tfsdk:"event_categories"`
	Message          types.String         `tfsdk:"message"`
	SourceARN        types.String         `tfsdk:"source_arn"`
	SourceIdentifier types.String         `tfsdk:"source_identifier"`
	SourceType       types.String         `tfsdk:"source_type"`
}
