// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package cloudwatch

import (
	"context"
	"errors"
	"time"

	"github.com/YakDriver/smarterr"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatch"
	awstypes "github.com/aws/aws-sdk-go-v2/service/cloudwatch/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_cloudwatch_otel_enrichment", name="OTel Enrichment")
// @SingletonIdentity(identityDuplicateAttributes="id")
// @Testing(serialize=true, hasNoPreExistingResource=true)
func newOtelEnrichmentResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &otelEnrichmentResource{}

	r.SetDefaultCreateTimeout(5 * time.Minute)
	r.SetDefaultDeleteTimeout(5 * time.Minute)

	return r, nil
}

type otelEnrichmentResource struct {
	framework.ResourceWithModel[otelEnrichmentResourceModel]
	framework.WithTimeouts
	framework.WithNoUpdate
	framework.WithImportByIdentity
}

func (r *otelEnrichmentResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrID: framework.IDAttributeDeprecatedWithAlternate(path.Root(names.AttrRegion)),
		},
		Blocks: map[string]schema.Block{
			names.AttrTimeouts: timeouts.Block(ctx, timeouts.Opts{
				Create: true,
				Delete: true,
			}),
		},
	}
}

func (r *otelEnrichmentResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data otelEnrichmentResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Plan.Get(ctx, &data))
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().CloudWatchClient(ctx)

	var input cloudwatch.StartOTelEnrichmentInput
	_, err := conn.StartOTelEnrichment(ctx, &input)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, "operation", "starting CloudWatch OTel Enrichment")
		return
	}

	createTimeout := r.CreateTimeout(ctx, data.Timeouts)
	_, err = waitOtelEnrichmentReady(ctx, conn, createTimeout)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, "operation", "waiting for CloudWatch OTel Enrichment start")
		return
	}

	data.ID = fwflex.StringValueToFramework(ctx, r.Meta().Region(ctx))

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &data))
}

func (r *otelEnrichmentResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data otelEnrichmentResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &data))
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().CloudWatchClient(ctx)

	out, err := findOtelEnrichment(ctx, conn)
	if retry.NotFound(err) {
		smerr.AddOne(ctx, &resp.Diagnostics, fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, "operation", "reading CloudWatch OTel Enrichment")
		return
	}

	_ = out // Resource exists if status is Running.

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &data))
}

func (r *otelEnrichmentResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data otelEnrichmentResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &data))
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().CloudWatchClient(ctx)

	var input cloudwatch.StopOTelEnrichmentInput
	_, err := conn.StopOTelEnrichment(ctx, &input)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, "operation", "stopping CloudWatch OTel Enrichment")
		return
	}

	deleteTimeout := r.DeleteTimeout(ctx, data.Timeouts)
	_, err = waitOtelEnrichmentDeleted(ctx, conn, deleteTimeout)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, "operation", "waiting for CloudWatch OTel Enrichment stop")
		return
	}
}

func findOtelEnrichmentStatus(ctx context.Context, conn *cloudwatch.Client, input *cloudwatch.GetOTelEnrichmentInput) (*cloudwatch.GetOTelEnrichmentOutput, error) {
	out, err := conn.GetOTelEnrichment(ctx, input)
	if err != nil {
		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return nil, smarterr.NewError(&retry.NotFoundError{
				LastError: err,
			})
		}
		return nil, smarterr.NewError(err)
	}

	if out == nil {
		return nil, smarterr.NewError(tfresource.NewEmptyResultError())
	}

	return out, nil
}

func findOtelEnrichment(ctx context.Context, conn *cloudwatch.Client) (*cloudwatch.GetOTelEnrichmentOutput, error) {
	var input cloudwatch.GetOTelEnrichmentInput
	out, err := findOtelEnrichmentStatus(ctx, conn, &input)
	if err != nil {
		return nil, smarterr.NewError(err)
	}

	if out.Status != awstypes.OTelEnrichmentStatusRunning {
		return nil, smarterr.NewError(&retry.NotFoundError{
			LastError: errors.New("OTel enrichment not running"),
		})
	}

	return out, nil
}

func waitOtelEnrichmentReady(ctx context.Context, conn *cloudwatch.Client, timeout time.Duration) (*cloudwatch.GetOTelEnrichmentOutput, error) { //nolint:unparam
	// For now, just return immediately since we don't have waiter logic
	return findOtelEnrichment(ctx, conn)
}

func waitOtelEnrichmentDeleted(ctx context.Context, conn *cloudwatch.Client, timeout time.Duration) (*cloudwatch.GetOTelEnrichmentOutput, error) { //nolint:unparam
	var input cloudwatch.GetOTelEnrichmentInput
	out, err := findOtelEnrichmentStatus(ctx, conn, &input)
	if retry.NotFound(err) {
		return nil, nil //nolint:nilnil
	}
	if err != nil {
		return nil, smarterr.NewError(err)
	}

	if out.Status == awstypes.OTelEnrichmentStatusStopped {
		return nil, nil //nolint:nilnil
	}

	return nil, smarterr.NewError(errors.New("OTel enrichment still running"))
}

type otelEnrichmentResourceModel struct {
	framework.WithRegionModel
	ID       types.String   `tfsdk:"id"`
	Timeouts timeouts.Value `tfsdk:"timeouts"`
}
