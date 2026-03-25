// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package observabilityadmin

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/observabilityadmin"
	awstypes "github.com/aws/aws-sdk-go-v2/service/observabilityadmin/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
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

// @FrameworkResource("aws_observabilityadmin_telemetry_enrichment", name="Telemetry Enrichment")
// @SingletonIdentity
// @Testing(preCheck="testAccTelemetryEnrichmentPreCheck")
// @Testing(tagsTest=false)
// @Testing(identityTest=false)
// @Testing(hasNoPreExistingResource=true)
// @Testing(generator=false)
func newTelemetryEnrichmentResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &telemetryEnrichmentResource{}

	r.SetDefaultCreateTimeout(5 * time.Minute)
	r.SetDefaultDeleteTimeout(5 * time.Minute)

	return r, nil
}

type telemetryEnrichmentResource struct {
	framework.ResourceWithModel[telemetryEnrichmentResourceModel]
	framework.WithNoUpdate
	framework.WithImportByIdentity
	framework.WithTimeouts
}

func (r *telemetryEnrichmentResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"aws_resource_explorer_managed_view_arn": framework.ARNAttributeComputedOnly(),
		},
		Blocks: map[string]schema.Block{
			names.AttrTimeouts: timeouts.Block(ctx, timeouts.Opts{
				Create: true,
				Delete: true,
			}),
		},
	}
}

func (r *telemetryEnrichmentResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data telemetryEnrichmentResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Plan.Get(ctx, &data))
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().ObservabilityAdminClient(ctx)

	var input observabilityadmin.StartTelemetryEnrichmentInput
	_, err := conn.StartTelemetryEnrichment(ctx, &input)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err)
		return
	}

	out, err := waitTelemetryEnrichmentRunning(ctx, conn, r.CreateTimeout(ctx, data.Timeouts))
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err)
		return
	}

	data.AWSResourceExplorerManagedViewARN = fwflex.StringToFramework(ctx, out.AwsResourceExplorerManagedViewArn)

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, data))
}

func (r *telemetryEnrichmentResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data telemetryEnrichmentResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &data))
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().ObservabilityAdminClient(ctx)

	out, err := findTelemetryEnrichment(ctx, conn)
	if retry.NotFound(err) {
		smerr.AddOne(ctx, &resp.Diagnostics, fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err)
		return
	}

	data.AWSResourceExplorerManagedViewARN = fwflex.StringToFramework(ctx, out.AwsResourceExplorerManagedViewArn)

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &data))
}

func (r *telemetryEnrichmentResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data telemetryEnrichmentResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &data))
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().ObservabilityAdminClient(ctx)

	var input observabilityadmin.StopTelemetryEnrichmentInput
	const (
		timeout = 1 * time.Minute
	)
	_, err := tfresource.RetryWhenIsA[any, *awstypes.ConflictException](ctx, timeout, func(ctx context.Context) (any, error) {
		return conn.StopTelemetryEnrichment(ctx, &input)
	})
	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return
	}
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err)
		return
	}

	if _, err := waitTelemetryEnrichmentStopped(ctx, conn, r.DeleteTimeout(ctx, data.Timeouts)); err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err)
		return
	}
}

func findTelemetryEnrichment(ctx context.Context, conn *observabilityadmin.Client) (*observabilityadmin.GetTelemetryEnrichmentStatusOutput, error) {
	var input observabilityadmin.GetTelemetryEnrichmentStatusInput
	out, err := findTelemetryEnrichmentStatus(ctx, conn, &input)

	if err != nil {
		return nil, err
	}

	if status := out.Status; status == awstypes.TelemetryEnrichmentStatusStopped {
		return nil, &retry.NotFoundError{
			Message: string(status),
		}
	}

	return out, nil
}

func findTelemetryEnrichmentStatus(ctx context.Context, conn *observabilityadmin.Client, input *observabilityadmin.GetTelemetryEnrichmentStatusInput) (*observabilityadmin.GetTelemetryEnrichmentStatusOutput, error) {
	out, err := conn.GetTelemetryEnrichmentStatus(ctx, input)

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

func statusTelemetryEnrichment(conn *observabilityadmin.Client) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		var input observabilityadmin.GetTelemetryEnrichmentStatusInput
		out, err := findTelemetryEnrichmentStatus(ctx, conn, &input)
		if retry.NotFound(err) {
			return nil, "", nil
		}
		if err != nil {
			return nil, "", err
		}
		return out, string(out.Status), nil
	}
}

func waitTelemetryEnrichmentRunning(ctx context.Context, conn *observabilityadmin.Client, timeout time.Duration) (*observabilityadmin.GetTelemetryEnrichmentStatusOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   []string{},
		Target:                    enum.Slice(awstypes.TelemetryEnrichmentStatusRunning),
		Refresh:                   statusTelemetryEnrichment(conn),
		Timeout:                   timeout,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*observabilityadmin.GetTelemetryEnrichmentStatusOutput); ok {
		return out, err
	}

	return nil, err
}

func waitTelemetryEnrichmentStopped(ctx context.Context, conn *observabilityadmin.Client, timeout time.Duration) (*observabilityadmin.GetTelemetryEnrichmentStatusOutput, error) { //nolint:unparam
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.TelemetryEnrichmentStatusRunning, awstypes.TelemetryEnrichmentStatusImpaired),
		Target:  enum.Slice(awstypes.TelemetryEnrichmentStatusStopped),
		Refresh: statusTelemetryEnrichment(conn),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*observabilityadmin.GetTelemetryEnrichmentStatusOutput); ok {
		return out, err
	}

	return nil, err
}

type telemetryEnrichmentResourceModel struct {
	framework.WithRegionModel
	AWSResourceExplorerManagedViewARN types.String   `tfsdk:"aws_resource_explorer_managed_view_arn"`
	Timeouts                          timeouts.Value `tfsdk:"timeouts"`
}
