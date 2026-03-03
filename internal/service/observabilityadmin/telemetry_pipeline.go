// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package observabilityadmin

import (
	"context"
	"errors"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/observabilityadmin"
	awstypes "github.com/aws/aws-sdk-go-v2/service/observabilityadmin/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
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

// @FrameworkResource("aws_observabilityadmin_telemetry_pipeline", name="Telemetry Pipeline")
// @Tags(identifierAttribute="arn")
func newTelemetryPipelineResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &telemetryPipelineResource{}

	r.SetDefaultCreateTimeout(30 * time.Minute)
	r.SetDefaultUpdateTimeout(30 * time.Minute)
	r.SetDefaultDeleteTimeout(30 * time.Minute)

	return r, nil
}

type telemetryPipelineResource struct {
	framework.ResourceWithModel[telemetryPipelineResourceModel]
	framework.WithTimeouts
}

func (r *telemetryPipelineResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrARN: framework.ARNAttributeComputedOnly(),
			names.AttrName: schema.StringAttribute{
				Required: true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(3, 28),
					stringvalidator.RegexMatches(regexache.MustCompile(`^[a-z][a-z0-9\-]+$`), "must start with a lowercase letter and contain only lowercase letters, digits, and hyphens"),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
		},
		Blocks: map[string]schema.Block{
			names.AttrConfiguration: schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[telemetryPipelineConfigurationModel](ctx),
				Validators: []validator.List{
					listvalidator.IsRequired(),
					listvalidator.SizeAtLeast(1),
					listvalidator.SizeAtMost(1),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"body": schema.StringAttribute{
							Required: true,
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

func (r *telemetryPipelineResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data telemetryPipelineResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().ObservabilityAdminClient(ctx)

	name := fwflex.StringValueFromFramework(ctx, data.Name)
	var input observabilityadmin.CreateTelemetryPipelineInput
	response.Diagnostics.Append(fwflex.Expand(ctx, data, &input)...)
	if response.Diagnostics.HasError() {
		return
	}

	// Additional fields.
	input.Tags = getTagsIn(ctx)

	output, err := conn.CreateTelemetryPipeline(ctx, &input)
	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, name)
		return
	}

	// Set values for unknowns.
	data.ARN = fwflex.StringToFramework(ctx, output.Arn)

	if _, err := waitTelemetryPipelineReady(ctx, conn, name, r.CreateTimeout(ctx, data.Timeouts)); err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, name)
		return
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, response.State.Set(ctx, data))
}

func (r *telemetryPipelineResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data telemetryPipelineResourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.State.Get(ctx, &data))
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().ObservabilityAdminClient(ctx)

	name := fwflex.StringValueFromFramework(ctx, data.Name)
	out, err := findTelemetryPipelineByName(ctx, conn, name)
	if retry.NotFound(err) {
		smerr.AddOne(ctx, &response.Diagnostics, fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)
		return
	}

	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, name)
		return
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, fwflex.Flatten(ctx, out, &data))
	if response.Diagnostics.HasError() {
		return
	}

	setTagsOut(ctx, out.Tags)

	smerr.AddEnrich(ctx, &response.Diagnostics, response.State.Set(ctx, &data))
}

func (r *telemetryPipelineResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var new, old telemetryPipelineResourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.Plan.Get(ctx, &new))
	smerr.AddEnrich(ctx, &response.Diagnostics, request.State.Get(ctx, &old))
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().ObservabilityAdminClient(ctx)

	diff, d := fwflex.Diff(ctx, new, old)
	smerr.AddEnrich(ctx, &response.Diagnostics, d)
	if response.Diagnostics.HasError() {
		return
	}

	if diff.HasChanges() {
		arn := fwflex.StringValueFromFramework(ctx, new.ARN)
		var input observabilityadmin.UpdateTelemetryPipelineInput
		smerr.AddEnrich(ctx, &response.Diagnostics, fwflex.Expand(ctx, new, &input))
		if response.Diagnostics.HasError() {
			return
		}

		// Additional fields.
		input.PipelineIdentifier = aws.String(arn)

		_, err := conn.UpdateTelemetryPipeline(ctx, &input)
		if err != nil {
			smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, arn)
			return
		}

		name := fwflex.StringValueFromFramework(ctx, new.Name)
		if _, err := waitTelemetryPipelineReady(ctx, conn, name, r.UpdateTimeout(ctx, new.Timeouts)); err != nil {
			smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, name)
			return
		}
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, response.State.Set(ctx, &new))
}

func (r *telemetryPipelineResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data telemetryPipelineResourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.State.Get(ctx, &data))
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().ObservabilityAdminClient(ctx)

	arn := fwflex.StringValueFromFramework(ctx, data.ARN)
	input := observabilityadmin.DeleteTelemetryPipelineInput{
		PipelineIdentifier: aws.String(arn),
	}
	_, err := conn.DeleteTelemetryPipeline(ctx, &input)
	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return
	}
	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, arn)
		return
	}

	name := fwflex.StringValueFromFramework(ctx, data.Name)
	if _, err := waitTelemetryPipelineDeleted(ctx, conn, name, r.DeleteTimeout(ctx, data.Timeouts)); err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, name)
		return
	}
}

func (r *telemetryPipelineResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root(names.AttrName), request, response)
}

type telemetryPipelineResourceModel struct {
	framework.WithRegionModel
	ARN           types.String                                                         `tfsdk:"arn"`
	Configuration fwtypes.ListNestedObjectValueOf[telemetryPipelineConfigurationModel] `tfsdk:"configuration"`
	Name          types.String                                                         `tfsdk:"name"`
	Tags          tftags.Map                                                           `tfsdk:"tags"`
	TagsAll       tftags.Map                                                           `tfsdk:"tags_all"`
	Timeouts      timeouts.Value                                                       `tfsdk:"timeouts"`
}

type telemetryPipelineConfigurationModel struct {
	Body types.String `tfsdk:"body"`
}

func findTelemetryPipelineByName(ctx context.Context, conn *observabilityadmin.Client, name string) (*awstypes.TelemetryPipeline, error) {
	input := observabilityadmin.GetTelemetryPipelineInput{
		PipelineIdentifier: aws.String(name),
	}

	return findTelemetryPipeline(ctx, conn, &input)
}

func findTelemetryPipeline(ctx context.Context, conn *observabilityadmin.Client, input *observabilityadmin.GetTelemetryPipelineInput) (*awstypes.TelemetryPipeline, error) {
	output, err := conn.GetTelemetryPipeline(ctx, input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError: err,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.Pipeline == nil {
		return nil, tfresource.NewEmptyResultError()
	}

	return output.Pipeline, nil
}

func statusTelemetryPipeline(conn *observabilityadmin.Client, name string) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		output, err := findTelemetryPipelineByName(ctx, conn, name)

		if retry.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.Status), nil
	}
}

func waitTelemetryPipelineReady(ctx context.Context, conn *observabilityadmin.Client, name string, timeout time.Duration) (*awstypes.TelemetryPipeline, error) { //nolint:unparam
	stateConf := &retry.StateChangeConf{
		Pending:                   enum.Slice(awstypes.TelemetryPipelineStatusCreating, awstypes.TelemetryPipelineStatusUpdating),
		Target:                    enum.Slice(awstypes.TelemetryPipelineStatusActive),
		Refresh:                   statusTelemetryPipeline(conn, name),
		Timeout:                   timeout,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if output, ok := outputRaw.(*awstypes.TelemetryPipeline); ok {
		if v := output.StatusReason; v != nil && v.Description != nil {
			retry.SetLastError(err, errors.New(aws.ToString(v.Description)))
		}

		return output, err
	}

	return nil, err
}

func waitTelemetryPipelineDeleted(ctx context.Context, conn *observabilityadmin.Client, name string, timeout time.Duration) (*awstypes.TelemetryPipeline, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.TelemetryPipelineStatusDeleting),
		Target:  []string{},
		Refresh: statusTelemetryPipeline(conn, name),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if output, ok := outputRaw.(*awstypes.TelemetryPipeline); ok {
		if v := output.StatusReason; v != nil && v.Description != nil {
			retry.SetLastError(err, errors.New(aws.ToString(v.Description)))
		}

		return output, err
	}

	return nil, err
}
