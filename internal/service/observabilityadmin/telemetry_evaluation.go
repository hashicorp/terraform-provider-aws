// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package observabilityadmin

import (
	"context"
	"errors"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/observabilityadmin"
	awstypes "github.com/aws/aws-sdk-go-v2/service/observabilityadmin/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_observabilityadmin_telemetry_evaluation", name="Telemetry Evaluation")
// @SingletonIdentity(identityDuplicateAttributes="id")
// @Testing(serialize=true)
// @Testing(hasNoPreExistingResource=true)
// @Testing(preCheck="testAccTelemetryEvaluationPreCheck")
// @Testing(generator=false)
func newTelemetryEvaluationResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &telemetryEvaluationResource{}

	r.SetDefaultCreateTimeout(5 * time.Minute)
	r.SetDefaultDeleteTimeout(5 * time.Minute)

	return r, nil
}

type telemetryEvaluationResource struct {
	framework.ResourceWithModel[telemetryEvaluationResourceModel]
	framework.WithNoUpdate
	framework.WithImportByIdentity
	framework.WithTimeouts
}

func (r *telemetryEvaluationResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"failure_reason": schema.StringAttribute{
				Computed: true,
			},
			names.AttrID: framework.IDAttributeDeprecatedWithAlternate(path.Root(names.AttrRegion)),
			names.AttrStatus: schema.StringAttribute{
				Computed: true,
			},
		},
		Blocks: map[string]schema.Block{
			names.AttrTimeouts: timeouts.Block(ctx, timeouts.Opts{
				Create: true,
				Delete: true,
			}),
		},
	}
}

func (r *telemetryEvaluationResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data telemetryEvaluationResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Plan.Get(ctx, &data))
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().ObservabilityAdminClient(ctx)

	var input observabilityadmin.StartTelemetryEvaluationInput
	const (
		timeout = 1 * time.Minute
	)
	_, err := tfresource.RetryWhenIsA[any, *awstypes.ConflictException](ctx, timeout, func(ctx context.Context) (any, error) {
		return conn.StartTelemetryEvaluation(ctx, &input)
	})
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err)
		return
	}

	out, err := waitTelemetryEvaluationRunning(ctx, conn, r.CreateTimeout(ctx, data.Timeouts))
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err)
		return
	}

	data.FailureReason = fwflex.StringToFramework(ctx, out.FailureReason)
	data.ID = fwflex.StringValueToFramework(ctx, r.Meta().Region(ctx))
	data.Status = fwflex.StringValueToFramework(ctx, out.Status)

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, data))
}

func (r *telemetryEvaluationResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data telemetryEvaluationResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &data))
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().ObservabilityAdminClient(ctx)

	out, err := findTelemetryEvaluation(ctx, conn)
	if retry.NotFound(err) {
		smerr.AddOne(ctx, &resp.Diagnostics, fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err)
		return
	}

	data.FailureReason = fwflex.StringToFramework(ctx, out.FailureReason)
	data.Status = fwflex.StringValueToFramework(ctx, out.Status)

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &data))
}

func (r *telemetryEvaluationResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data telemetryEvaluationResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &data))
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().ObservabilityAdminClient(ctx)

	var input observabilityadmin.StopTelemetryEvaluationInput
	const (
		timeout = 1 * time.Minute
	)
	_, err := tfresource.RetryWhenIsA[any, *awstypes.ConflictException](ctx, timeout, func(ctx context.Context) (any, error) {
		return conn.StopTelemetryEvaluation(ctx, &input)
	})
	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return
	}
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err)
		return
	}

	if _, err := waitTelemetryEvaluationStopped(ctx, conn, r.DeleteTimeout(ctx, data.Timeouts)); err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err)
		return
	}
}

func findTelemetryEvaluation(ctx context.Context, conn *observabilityadmin.Client) (*observabilityadmin.GetTelemetryEvaluationStatusOutput, error) {
	var input observabilityadmin.GetTelemetryEvaluationStatusInput
	out, err := findTelemetryEvaluationStatus(ctx, conn, &input)

	if err != nil {
		return nil, err
	}

	if status := out.Status; status == awstypes.StatusNotStarted || status == awstypes.StatusStopped {
		return nil, &retry.NotFoundError{
			Message: string(status),
		}
	}

	return out, nil
}

func findTelemetryEvaluationStatus(ctx context.Context, conn *observabilityadmin.Client, input *observabilityadmin.GetTelemetryEvaluationStatusInput) (*observabilityadmin.GetTelemetryEvaluationStatusOutput, error) {
	out, err := conn.GetTelemetryEvaluationStatus(ctx, input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError: err,
		}
	}

	if err != nil {
		return nil, err
	}

	if out == nil {
		return nil, tfresource.NewEmptyResultError()
	}

	return out, nil
}

func statusTelemetryEvaluation(conn *observabilityadmin.Client) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		var input observabilityadmin.GetTelemetryEvaluationStatusInput
		out, err := findTelemetryEvaluationStatus(ctx, conn, &input)
		if retry.NotFound(err) {
			return nil, "", nil
		}
		if err != nil {
			return nil, "", err
		}
		return out, string(out.Status), nil
	}
}

func waitTelemetryEvaluationRunning(ctx context.Context, conn *observabilityadmin.Client, timeout time.Duration) (*observabilityadmin.GetTelemetryEvaluationStatusOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   enum.Slice(awstypes.StatusStarting),
		Target:                    enum.Slice(awstypes.StatusRunning),
		Refresh:                   statusTelemetryEvaluation(conn),
		Timeout:                   timeout,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*observabilityadmin.GetTelemetryEvaluationStatusOutput); ok {
		retry.SetLastError(err, errors.New(aws.ToString(out.FailureReason)))
		return out, err
	}

	return nil, err
}

func waitTelemetryEvaluationStopped(ctx context.Context, conn *observabilityadmin.Client, timeout time.Duration) (*observabilityadmin.GetTelemetryEvaluationStatusOutput, error) { //nolint:unparam
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.StatusRunning, awstypes.StatusStopping, awstypes.StatusFailedStart, awstypes.StatusFailedStop),
		Target:  enum.Slice(awstypes.StatusStopped),
		Refresh: statusTelemetryEvaluation(conn),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*observabilityadmin.GetTelemetryEvaluationStatusOutput); ok {
		retry.SetLastError(err, errors.New(aws.ToString(out.FailureReason)))
		return out, err
	}

	return nil, err
}

type telemetryEvaluationResourceModel struct {
	framework.WithRegionModel
	FailureReason types.String   `tfsdk:"failure_reason"`
	ID            types.String   `tfsdk:"id"`
	Status        types.String   `tfsdk:"status"`
	Timeouts      timeouts.Value `tfsdk:"timeouts"`
}
