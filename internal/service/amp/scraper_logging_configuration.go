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
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
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
			"scraper_id": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
		Blocks: map[string]schema.Block{
			"scraper_components": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[scraperComponentModel](ctx),
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						names.AttrType: schema.StringAttribute{
							CustomType: fwtypes.StringEnumType[awstypes.ScraperComponentType](),
							Required:   true,
						},
						"options": schema.MapAttribute{
							CustomType:  fwtypes.MapOfStringType,
							ElementType: types.StringType,
							Optional:    true,
						},
					},
				},
			},
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
							CustomType: fwtypes.NewListNestedObjectTypeOf[scraperCloudWatchLogDestinationModel](ctx),
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
	input.ScraperComponents = expandScraperComponents(ctx, data.ScraperComponents)

	_, err := conn.UpdateScraperLoggingConfiguration(ctx, &input)

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("creating Prometheus Scraper Logging Configuration (%s)", scraperID), err.Error())

		return
	}

	if _, err := waitScraperLoggingConfigurationCreated(ctx, conn, scraperID, r.CreateTimeout(ctx, data.Timeouts)); err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("waiting for Prometheus Scraper Logging Configuration (%s) create", scraperID), err.Error())

		return
	}

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

	smerr.AddEnrich(ctx, &response.Diagnostics, r.flattenIntoModel(ctx, output, &data))
	if response.Diagnostics.HasError() {
		return
	}

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
	input.ScraperComponents = expandScraperComponents(ctx, new.ScraperComponents)

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
	var input amp.DeleteScraperLoggingConfigurationInput
	input.ScraperId = aws.String(scraperID)
	input.ClientToken = aws.String(create.UniqueId(ctx))
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

// flattenIntoModel fills data from the API output, handling the autoflex-incompatible scraper_components field.
func (r *scraperLoggingConfigurationResource) flattenIntoModel(ctx context.Context, output *amp.DescribeScraperLoggingConfigurationOutput, data *scraperLoggingConfigurationResourceModel) diag.Diagnostics {
	var diags diag.Diagnostics

	diags.Append(fwflex.Flatten(ctx, output, data)...)
	if diags.HasError() {
		return diags
	}

	// Only flatten scraper_components if they were already in state (user-configured).
	// The API returns default components even when none were configured by the user.
	if !data.ScraperComponents.IsNull() && len(output.ScraperComponents) > 0 {
		componentPtrs := make([]*scraperComponentModel, len(output.ScraperComponents))
		for i, c := range output.ScraperComponents {
			component := scraperComponentModel{
				Type: fwtypes.StringEnumValue(c.Type),
			}

			if c.Config != nil && c.Config.Options != nil {
				elements := make(map[string]attr.Value, len(c.Config.Options))
				for k, v := range c.Config.Options {
					elements[k] = basetypes.NewStringValue(v)
				}
				optionsMap, d := fwtypes.NewMapValueOf[basetypes.StringValue](ctx, elements)
				diags.Append(d...)
				if !diags.HasError() {
					component.Options = optionsMap
				}
			} else {
				component.Options = fwtypes.NewMapValueOfNull[basetypes.StringValue](ctx)
			}
			componentPtrs[i] = &component
		}

		data.ScraperComponents = fwtypes.NewListNestedObjectValueOfSliceMust(ctx, componentPtrs)
	}

	return diags
}

func findScraperLoggingConfigurationByID(ctx context.Context, conn *amp.Client, id string) (*amp.DescribeScraperLoggingConfigurationOutput, error) {
	var input amp.DescribeScraperLoggingConfigurationInput
	input.ScraperId = aws.String(id)

	output, err := conn.DescribeScraperLoggingConfiguration(ctx, &input)

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
	ScraperComponents  fwtypes.ListNestedObjectValueOf[scraperComponentModel]          `tfsdk:"scraper_components" autoflex:"-"`
	ScraperID          types.String                                                    `tfsdk:"scraper_id"`
	Timeouts           timeouts.Value                                                  `tfsdk:"timeouts"`
}

type scraperLoggingDestinationModel struct {
	CloudwatchLogs fwtypes.ListNestedObjectValueOf[scraperCloudWatchLogDestinationModel] `tfsdk:"cloudwatch_logs"`
}

type scraperComponentModel struct {
	Type    fwtypes.StringEnum[awstypes.ScraperComponentType] `tfsdk:"type"`
	Options fwtypes.MapOfString                               `tfsdk:"options"`
}

type scraperCloudWatchLogDestinationModel struct {
	LogGroupARN fwtypes.ARN `tfsdk:"log_group_arn"`
}

var (
	_ fwflex.Expander  = scraperLoggingDestinationModel{}
	_ fwflex.Flattener = &scraperLoggingDestinationModel{}
)

func (m scraperLoggingDestinationModel) Expand(ctx context.Context) (any, diag.Diagnostics) {
	var diags diag.Diagnostics

	cwData, d := m.CloudwatchLogs.ToPtr(ctx)
	diags.Append(d...)
	if diags.HasError() {
		return nil, diags
	}

	if cwData == nil {
		return nil, diags
	}

	return &awstypes.ScraperLoggingDestinationMemberCloudWatchLogs{
		Value: awstypes.CloudWatchLogDestination{
			LogGroupArn: cwData.LogGroupARN.ValueStringPointer(),
		},
	}, diags
}

func (m *scraperLoggingDestinationModel) Flatten(ctx context.Context, v any) diag.Diagnostics {
	var diags diag.Diagnostics

	switch t := v.(type) {
	case awstypes.ScraperLoggingDestinationMemberCloudWatchLogs:
		var data scraperCloudWatchLogDestinationModel
		diags.Append(fwflex.Flatten(ctx, t.Value, &data)...)
		if diags.HasError() {
			return diags
		}
		m.CloudwatchLogs = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &data)
	}

	return diags
}

func expandScraperComponents(ctx context.Context, src fwtypes.ListNestedObjectValueOf[scraperComponentModel]) []awstypes.ScraperComponent {
	if src.IsNull() || src.IsUnknown() {
		return nil
	}

	components, diags := src.ToSlice(ctx)
	if diags.HasError() {
		return nil
	}

	result := make([]awstypes.ScraperComponent, len(components))
	for i, component := range components {
		result[i] = awstypes.ScraperComponent{
			Type: component.Type.ValueEnum(),
		}

		if !component.Options.IsNull() && !component.Options.IsUnknown() {
			optionsMap := make(map[string]string)
			component.Options.ElementsAs(ctx, &optionsMap, false)
			result[i].Config = &awstypes.ComponentConfig{
				Options: optionsMap,
			}
		}
	}

	return result
}
