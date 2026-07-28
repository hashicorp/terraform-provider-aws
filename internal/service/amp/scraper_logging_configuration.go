// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package amp

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/amp"
	awstypes "github.com/aws/aws-sdk-go-v2/service/amp/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_prometheus_scraper_logging_configuration", name="ScraperLoggingConfiguration")
// @Testing(existsType="github.com/aws/aws-sdk-go-v2/service/amp;amp.DescribeScraperLoggingConfigurationOutput")
func newScraperLoggingConfigurationResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &scraperLoggingConfigurationResource{}

	r.SetDefaultCreateTimeout(5 * time.Minute)
	r.SetDefaultUpdateTimeout(5 * time.Minute)
	r.SetDefaultDeleteTimeout(5 * time.Minute)

	return r, nil
}

type scraperLoggingConfigurationResource struct {
	framework.ResourceWithModel[scraperLoggingConfigurationResourceModel]
	framework.WithTimeouts
}

func (r *scraperLoggingConfigurationResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"scraper_components": schema.SetAttribute{
				CustomType:  fwtypes.SetOfStringEnumType[awstypes.ScraperComponentType](),
				ElementType: types.StringType,
				Optional:    true,
				Computed:    true,
				Validators: []validator.Set{
					setvalidator.SizeAtLeast(1),
				},
				PlanModifiers: []planmodifier.Set{
					setplanmodifier.UseStateForUnknown(),
				},
			},
			"scraper_id": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
		Blocks: map[string]schema.Block{
			"logging_destination": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[scraperLoggingDestinationModel](ctx),
				Validators: []validator.List{
					listvalidator.SizeAtLeast(1),
					listvalidator.SizeAtMost(1),
					listvalidator.IsRequired(),
				},
				NestedObject: schema.NestedBlockObject{
					Blocks: map[string]schema.Block{
						names.AttrCloudWatchLogs: schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[cloudWatchLogDestinationModel](ctx),
							Validators: []validator.List{
								listvalidator.SizeAtLeast(1),
								listvalidator.SizeAtMost(1),
								listvalidator.IsRequired(),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"log_group_arn": schema.StringAttribute{
										CustomType: fwtypes.ARNType,
										Required:   true,
										Validators: []validator.String{
											stringvalidator.RegexMatches(regexache.MustCompile(`:\*$`), "ARN must end with `:*`"),
										},
									},
								},
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

func (r *scraperLoggingConfigurationResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data scraperLoggingConfigurationResourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.Plan.Get(ctx, &data))
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().AMPClient(ctx)

	scraperID := fwflex.StringValueFromFramework(ctx, data.ScraperID)
	var input amp.UpdateScraperLoggingConfigurationInput
	smerr.AddEnrich(ctx, &response.Diagnostics, fwflex.Expand(ctx, data, &input))
	if response.Diagnostics.HasError() {
		return
	}
	if scraperComponents := fwflex.ExpandFrameworkStringyValueSet[awstypes.ScraperComponentType](ctx, data.ScraperComponents); len(scraperComponents) > 0 {
		input.ScraperComponents = tfslices.ApplyToAll(scraperComponents, func(v awstypes.ScraperComponentType) awstypes.ScraperComponent {
			return awstypes.ScraperComponent{
				Type: v,
			}
		})
	}

	_, err := conn.UpdateScraperLoggingConfiguration(ctx, &input)

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("creating Prometheus Scraper Logging Configuration (%s)", scraperID), err.Error())

		return
	}

	output, err := waitScraperLoggingConfigurationCreated(ctx, conn, scraperID, r.CreateTimeout(ctx, data.Timeouts))

	if err != nil {
		// Taint the resource.
		response.State.SetAttribute(ctx, path.Root("scraper_id"), scraperID)
		response.Diagnostics.AddError(fmt.Sprintf("waiting for Prometheus Scraper Logging Configuration (%s) create", scraperID), err.Error())

		return
	}

	// Set values for unknowns.
	data.ScraperComponents = fwflex.FlattenFrameworkStringyValueSetOfStringEnum(ctx, tfslices.ApplyToAll(output.ScraperComponents, func(v awstypes.ScraperComponent) awstypes.ScraperComponentType {
		return v.Type
	}))

	smerr.AddEnrich(ctx, &response.Diagnostics, response.State.Set(ctx, data))
}

func (r *scraperLoggingConfigurationResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data scraperLoggingConfigurationResourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.State.Get(ctx, &data))
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().AMPClient(ctx)

	scraperID := fwflex.StringValueFromFramework(ctx, data.ScraperID)
	output, err := findScraperLoggingConfigurationByID(ctx, conn, scraperID)

	if retry.NotFound(err) {
		smerr.AddOne(ctx, &response.Diagnostics, fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)

		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading Prometheus Scraper Logging Configuration (%s)", scraperID), err.Error())

		return
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, fwflex.Flatten(ctx, output, &data))
	if response.Diagnostics.HasError() {
		return
	}

	data.ScraperComponents = fwflex.FlattenFrameworkStringyValueSetOfStringEnum(ctx, tfslices.ApplyToAll(output.ScraperComponents, func(v awstypes.ScraperComponent) awstypes.ScraperComponentType {
		return v.Type
	}))

	smerr.AddEnrich(ctx, &response.Diagnostics, response.State.Set(ctx, &data))
}

func (r *scraperLoggingConfigurationResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var new scraperLoggingConfigurationResourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.Plan.Get(ctx, &new))
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().AMPClient(ctx)

	scraperID := fwflex.StringValueFromFramework(ctx, new.ScraperID)
	var input amp.UpdateScraperLoggingConfigurationInput
	smerr.AddEnrich(ctx, &response.Diagnostics, fwflex.Expand(ctx, new, &input))
	if response.Diagnostics.HasError() {
		return
	}
	if scraperComponents := fwflex.ExpandFrameworkStringyValueSet[awstypes.ScraperComponentType](ctx, new.ScraperComponents); len(scraperComponents) > 0 {
		input.ScraperComponents = tfslices.ApplyToAll(scraperComponents, func(v awstypes.ScraperComponentType) awstypes.ScraperComponent {
			return awstypes.ScraperComponent{
				Type: v,
			}
		})
	}

	_, err := conn.UpdateScraperLoggingConfiguration(ctx, &input)

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("updating Prometheus Scraper Logging Configuration (%s)", scraperID), err.Error())

		return
	}

	if _, err := waitScraperLoggingConfigurationUpdated(ctx, conn, scraperID, r.UpdateTimeout(ctx, new.Timeouts)); err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("waiting for Prometheus Scraper Logging Configuration (%s) update", scraperID), err.Error())

		return
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, response.State.Set(ctx, new))
}

func (r *scraperLoggingConfigurationResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data scraperLoggingConfigurationResourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.State.Get(ctx, &data))
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().AMPClient(ctx)

	scraperID := fwflex.StringValueFromFramework(ctx, data.ScraperID)
	input := amp.DeleteScraperLoggingConfigurationInput{
		ClientToken: aws.String(create.UniqueId(ctx)),
		ScraperId:   aws.String(scraperID),
	}
	_, err := conn.DeleteScraperLoggingConfiguration(ctx, &input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("deleting Prometheus Scraper Logging Configuration (%s)", scraperID), err.Error())

		return
	}

	if _, err := waitScraperLoggingConfigurationDeleted(ctx, conn, scraperID, r.DeleteTimeout(ctx, data.Timeouts)); err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("waiting for Prometheus Scraper Logging Configuration (%s) delete", scraperID), err.Error())

		return
	}
}

func (r *scraperLoggingConfigurationResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("scraper_id"), request, response)
}

func findScraperLoggingConfigurationByID(ctx context.Context, conn *amp.Client, id string) (*amp.DescribeScraperLoggingConfigurationOutput, error) {
	input := amp.DescribeScraperLoggingConfigurationInput{
		ScraperId: aws.String(id),
	}

	return findScraperLoggingConfiguration(ctx, conn, &input)
}

func findScraperLoggingConfiguration(ctx context.Context, conn *amp.Client, input *amp.DescribeScraperLoggingConfigurationInput) (*amp.DescribeScraperLoggingConfigurationOutput, error) {
	output, err := conn.DescribeScraperLoggingConfiguration(ctx, input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError: err,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.Status == nil {
		return nil, tfresource.NewEmptyResultError()
	}

	return output, nil
}

func statusScraperLoggingConfiguration(conn *amp.Client, id string) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		output, err := findScraperLoggingConfigurationByID(ctx, conn, id)

		if retry.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.Status.StatusCode), nil
	}
}

func waitScraperLoggingConfigurationCreated(ctx context.Context, conn *amp.Client, id string, timeout time.Duration) (*amp.DescribeScraperLoggingConfigurationOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.ScraperLoggingConfigurationStatusCodeCreating, awstypes.ScraperLoggingConfigurationStatusCodeUpdating),
		Target:  enum.Slice(awstypes.ScraperLoggingConfigurationStatusCodeActive),
		Refresh: statusScraperLoggingConfiguration(conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*amp.DescribeScraperLoggingConfigurationOutput); ok {
		retry.SetLastError(err, errors.New(aws.ToString(output.Status.StatusReason)))

		return output, err
	}

	return nil, err
}

func waitScraperLoggingConfigurationUpdated(ctx context.Context, conn *amp.Client, id string, timeout time.Duration) (*amp.DescribeScraperLoggingConfigurationOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.ScraperLoggingConfigurationStatusCodeUpdating),
		Target:  enum.Slice(awstypes.ScraperLoggingConfigurationStatusCodeActive),
		Refresh: statusScraperLoggingConfiguration(conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*amp.DescribeScraperLoggingConfigurationOutput); ok {
		retry.SetLastError(err, errors.New(aws.ToString(output.Status.StatusReason)))

		return output, err
	}

	return nil, err
}

func waitScraperLoggingConfigurationDeleted(ctx context.Context, conn *amp.Client, id string, timeout time.Duration) (*amp.DescribeScraperLoggingConfigurationOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.ScraperLoggingConfigurationStatusCodeDeleting, awstypes.ScraperLoggingConfigurationStatusCodeActive),
		Target:  []string{},
		Refresh: statusScraperLoggingConfiguration(conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*amp.DescribeScraperLoggingConfigurationOutput); ok {
		retry.SetLastError(err, errors.New(aws.ToString(output.Status.StatusReason)))

		return output, err
	}

	return nil, err
}

type scraperLoggingConfigurationResourceModel struct {
	framework.WithRegionModel
	LoggingDestination fwtypes.ListNestedObjectValueOf[scraperLoggingDestinationModel] `tfsdk:"logging_destination"`
	ScraperComponents  fwtypes.SetOfStringEnum[awstypes.ScraperComponentType]          `tfsdk:"scraper_components" autoflex:"-"`
	ScraperID          types.String                                                    `tfsdk:"scraper_id"`
	Timeouts           timeouts.Value                                                  `tfsdk:"timeouts"`
}

type scraperLoggingDestinationModel struct {
	CloudwatchLogs fwtypes.ListNestedObjectValueOf[cloudWatchLogDestinationModel] `tfsdk:"cloudwatch_logs"`
}

var (
	_ fwflex.Expander  = scraperLoggingDestinationModel{}
	_ fwflex.Flattener = &scraperLoggingDestinationModel{}
)

func (m scraperLoggingDestinationModel) Expand(ctx context.Context) (any, diag.Diagnostics) {
	var diags diag.Diagnostics
	var v awstypes.ScraperLoggingDestination

	switch {
	case !m.CloudwatchLogs.IsNull():
		data, d := m.CloudwatchLogs.ToPtr(ctx)
		diags.Append(d...)
		if diags.HasError() {
			return nil, diags
		}
		var apiObject awstypes.ScraperLoggingDestinationMemberCloudWatchLogs
		diags.Append(fwflex.Expand(ctx, data, &apiObject.Value)...)
		if diags.HasError() {
			return nil, diags
		}
		v = &apiObject
	}

	return v, diags
}

func (m *scraperLoggingDestinationModel) Flatten(ctx context.Context, v any) diag.Diagnostics {
	var diags diag.Diagnostics

	switch t := v.(type) {
	case awstypes.ScraperLoggingDestinationMemberCloudWatchLogs:
		var data cloudWatchLogDestinationModel
		diags.Append(fwflex.Flatten(ctx, t.Value, &data)...)
		if diags.HasError() {
			return diags
		}
		m.CloudwatchLogs = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &data)

	default:
		diags.AddError(
			"Unsupported Type",
			fmt.Sprintf("scraperLoggingDestinationModel.Flatten: %T", v),
		)
	}

	return diags
}
