// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package observabilityadmin

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/observabilityadmin"
	awstypes "github.com/aws/aws-sdk-go-v2/service/observabilityadmin/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	fwvalidators "github.com/hashicorp/terraform-provider-aws/internal/framework/validators"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_observabilityadmin_telemetry_evaluation_for_organization", name="Telemetry Evaluation For Organization")
// @SingletonIdentity(identityDuplicateAttributes="id")
// @Testing(serialize=true)
// @Testing(hasNoPreExistingResource=true)
// @Testing(preCheck="github.com/hashicorp/terraform-provider-aws/internal/acctest;acctest.PreCheckOrganizationManagementAccount")
// @Testing(preCheck="testAccTelemetryEvaluationForOrganizationPreCheck")
// @Testing(generator=false)
func newTelemetryEvaluationForOrganizationResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &telemetryEvaluationForOrganizationResource{}

	r.SetDefaultCreateTimeout(5 * time.Minute)
	r.SetDefaultDeleteTimeout(5 * time.Minute)

	return r, nil
}

type telemetryEvaluationForOrganizationResource struct {
	framework.ResourceWithModel[telemetryEvaluationForOrganizationResourceModel]
	framework.WithNoUpdate
	framework.WithImportByIdentity
	framework.WithTimeouts
}

func (r *telemetryEvaluationForOrganizationResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"all_regions": schema.BoolAttribute{
				Optional: true,
			},
			"failure_reason": schema.StringAttribute{
				Computed: true,
			},
			"home_region": schema.StringAttribute{
				Computed: true,
			},
			names.AttrID: framework.IDAttributeDeprecatedWithAlternate(path.Root(names.AttrRegion)),
			"regions": schema.SetAttribute{
				ElementType: types.StringType,
				CustomType:  fwtypes.SetOfStringType,
				Optional:    true,
				Validators: []validator.Set{
					setvalidator.ValueStringsAre(fwvalidators.AWSRegion()),
				},
			},
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

func (r *telemetryEvaluationForOrganizationResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data telemetryEvaluationForOrganizationResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Plan.Get(ctx, &data))
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().ObservabilityAdminClient(ctx)

	var input observabilityadmin.StartTelemetryEvaluationForOrganizationInput
	smerr.AddEnrich(ctx, &resp.Diagnostics, fwflex.Expand(ctx, data, &input))
	if resp.Diagnostics.HasError() {
		return
	}

	const (
		timeout = 1 * time.Minute
	)
	_, err := tfresource.RetryWhenIsA[any, *awstypes.ConflictException](ctx, timeout, func(ctx context.Context) (any, error) {
		return conn.StartTelemetryEvaluationForOrganization(ctx, &input)
	})
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err)
		return
	}

	out, err := waitTelemetryEvaluationForOrganizationRunning(ctx, conn, r.CreateTimeout(ctx, data.Timeouts))
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err)
		return
	}

	// Set values for unknowns.
	smerr.AddEnrich(ctx, &resp.Diagnostics, fwflex.Flatten(ctx, out, &data))
	if resp.Diagnostics.HasError() {
		return
	}
	data.ID = fwflex.StringValueToFramework(ctx, r.Meta().Region(ctx))

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, data))
}

func (r *telemetryEvaluationForOrganizationResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data telemetryEvaluationForOrganizationResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &data))
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().ObservabilityAdminClient(ctx)

	out, err := findTelemetryEvaluationForOrganization(ctx, conn)
	if retry.NotFound(err) {
		smerr.AddOne(ctx, &resp.Diagnostics, fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err)
		return
	}

	// Set attributes for import.
	smerr.AddEnrich(ctx, &resp.Diagnostics, fwflex.Flatten(ctx, out, &data))
	if resp.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &data))
}

func (r *telemetryEvaluationForOrganizationResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data telemetryEvaluationForOrganizationResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &data))
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().ObservabilityAdminClient(ctx)

	var input observabilityadmin.StopTelemetryEvaluationForOrganizationInput
	const (
		timeout = 1 * time.Minute
	)
	_, err := tfresource.RetryWhenIsA[any, *awstypes.ConflictException](ctx, timeout, func(ctx context.Context) (any, error) {
		return conn.StopTelemetryEvaluationForOrganization(ctx, &input)
	})
	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return
	}
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err)
		return
	}

	if _, err := waitTelemetryEvaluationForOrganizationStopped(ctx, conn, r.DeleteTimeout(ctx, data.Timeouts)); err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err)
		return
	}
}

func findTelemetryEvaluationForOrganization(ctx context.Context, conn *observabilityadmin.Client) (*observabilityadmin.GetTelemetryEvaluationStatusForOrganizationOutput, error) {
	var input observabilityadmin.GetTelemetryEvaluationStatusForOrganizationInput
	out, err := findTelemetryEvaluationForOrganizationStatus(ctx, conn, &input)

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

func findTelemetryEvaluationForOrganizationStatus(ctx context.Context, conn *observabilityadmin.Client, input *observabilityadmin.GetTelemetryEvaluationStatusForOrganizationInput) (*observabilityadmin.GetTelemetryEvaluationStatusForOrganizationOutput, error) {
	out, err := conn.GetTelemetryEvaluationStatusForOrganization(ctx, input)

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

func statusTelemetryEvaluationForOrganization(conn *observabilityadmin.Client) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		var input observabilityadmin.GetTelemetryEvaluationStatusForOrganizationInput
		out, err := findTelemetryEvaluationForOrganizationStatus(ctx, conn, &input)
		if retry.NotFound(err) {
			return nil, "", nil
		}
		if err != nil {
			return nil, "", err
		}
		return out, string(out.Status), nil
	}
}

func waitTelemetryEvaluationForOrganizationRunning(ctx context.Context, conn *observabilityadmin.Client, timeout time.Duration) (*observabilityadmin.GetTelemetryEvaluationStatusForOrganizationOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   enum.Slice(awstypes.StatusStarting),
		Target:                    enum.Slice(awstypes.StatusRunning),
		Refresh:                   statusTelemetryEvaluationForOrganization(conn),
		Timeout:                   timeout,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*observabilityadmin.GetTelemetryEvaluationStatusForOrganizationOutput); ok {
		return out, err
	}

	return nil, err
}

func waitTelemetryEvaluationForOrganizationStopped(ctx context.Context, conn *observabilityadmin.Client, timeout time.Duration) (*observabilityadmin.GetTelemetryEvaluationStatusForOrganizationOutput, error) { //nolint:unparam
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.StatusRunning, awstypes.StatusStopping, awstypes.StatusFailedStart, awstypes.StatusFailedStop),
		Target:  enum.Slice(awstypes.StatusStopped),
		Refresh: statusTelemetryEvaluationForOrganization(conn),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*observabilityadmin.GetTelemetryEvaluationStatusForOrganizationOutput); ok {
		return out, err
	}

	return nil, err
}

type telemetryEvaluationForOrganizationResourceModel struct {
	framework.WithRegionModel
	AllRegions    types.Bool          `tfsdk:"all_regions"`
	FailureReason types.String        `tfsdk:"failure_reason"`
	HomeRegion    types.String        `tfsdk:"home_region"`
	ID            types.String        `tfsdk:"id"`
	Regions       fwtypes.SetOfString `tfsdk:"regions"`
	Status        types.String        `tfsdk:"status"`
	Timeouts      timeouts.Value      `tfsdk:"timeouts"`
}
