// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package cloudwatch

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/YakDriver/smarterr"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatch"
	awstypes "github.com/aws/aws-sdk-go-v2/service/cloudwatch/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
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

// Maximum selectors allowed across include_filters and exclude_filters
// combined. The service enforces the combined total, not a per-list budget, so
// two lists of 60 satisfy each list's own limit but are rejected together.
const otelEnrichmentMaxFilters = 100

// @FrameworkResource("aws_cloudwatch_otel_enrichment", name="OTel Enrichment")
// @SingletonIdentity(identityDuplicateAttributes="id")
// @Testing(serialize=true, hasNoPreExistingResource=true, generator=false)
// @Testing(preCheck="testAccPreCheckOTelEnrichment")
func newOTelEnrichmentResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &otelEnrichmentResource{}

	r.SetDefaultCreateTimeout(5 * time.Minute)
	r.SetDefaultUpdateTimeout(5 * time.Minute)
	r.SetDefaultDeleteTimeout(5 * time.Minute)

	return r, nil
}

type otelEnrichmentResource struct {
	framework.ResourceWithModel[otelEnrichmentResourceModel]
	framework.WithTimeouts
	framework.WithImportByIdentity
}

func (r *otelEnrichmentResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	filterBlock := func() schema.SetNestedBlock {
		return schema.SetNestedBlock{
			CustomType: fwtypes.NewSetNestedObjectTypeOf[otelEnrichmentMetricSelectorModel](ctx),
			Validators: []validator.Set{
				setvalidator.SizeAtMost(otelEnrichmentMaxFilters),
			},
			NestedObject: schema.NestedBlockObject{
				Attributes: map[string]schema.Attribute{
					"metric_names": schema.SetAttribute{
						CustomType: fwtypes.SetOfStringType,
						Optional:   true,
						Validators: []validator.Set{
							setvalidator.SizeAtMost(100),
							setvalidator.ValueStringsAre(stringvalidator.LengthBetween(1, 255)),
						},
					},
					names.AttrNamespace: schema.StringAttribute{
						Required: true,
						Validators: []validator.String{
							stringvalidator.LengthBetween(1, 255),
							stringvalidator.RegexMatches(regexache.MustCompile(`^[^:]`), "must not begin with a colon"),
						},
					},
				},
			},
		}
	}

	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrID: framework.IDAttributeDeprecatedWithAlternate(path.Root(names.AttrRegion)),
		},
		Blocks: map[string]schema.Block{
			"exclude_filters": filterBlock(),
			"include_filters": filterBlock(),
			names.AttrTimeouts: timeouts.Block(ctx, timeouts.Opts{
				Create: true,
				Update: true,
				Delete: true,
			}),
		},
	}
}

// ValidateConfig enforces the combined selector budget, which no per-attribute
// validator can see because it spans two attributes.
func (r *otelEnrichmentResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var data otelEnrichmentResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Config.Get(ctx, &data))
	if resp.Diagnostics.HasError() {
		return
	}

	// Unknown values are only resolved at apply time, so the total cannot be
	// checked yet.
	if data.IncludeFilters.IsUnknown() || data.ExcludeFilters.IsUnknown() {
		return
	}

	total := data.IncludeFilters.Length(fwtypes.CollectionLengthUnhandledAsZero) +
		data.ExcludeFilters.Length(fwtypes.CollectionLengthUnhandledAsZero)
	if total > otelEnrichmentMaxFilters {
		smerr.AddOne(ctx, &resp.Diagnostics, diag.NewErrorDiagnostic(
			"Too many enrichment filters",
			fmt.Sprintf("CloudWatch allows at most %d selectors across include_filters and exclude_filters combined, but %d were configured.",
				otelEnrichmentMaxFilters, total),
		))
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
	smerr.AddEnrich(ctx, &resp.Diagnostics, fwflex.Expand(ctx, data, &input))
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := conn.StartOTelEnrichment(ctx, &input)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, "operation", "starting CloudWatch OTel Enrichment")
		return
	}

	// StartOTelEnrichment is a no-op against an account that is already running
	// enrichment: the filters in the request are discarded and the response
	// carries the filters that were already stored. Comparing the two is the
	// only way to detect this, so reconcile explicitly when they differ rather
	// than silently leaving the configured filters unapplied.
	var stored otelEnrichmentResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, fwflex.Flatten(ctx, out, &stored))
	if resp.Diagnostics.HasError() {
		return
	}
	normalizeOTelEnrichmentFilters(ctx, &stored)

	if !stored.IncludeFilters.Equal(data.IncludeFilters) || !stored.ExcludeFilters.Equal(data.ExcludeFilters) {
		if err := putOTelEnrichmentFilters(ctx, conn, &data); err != nil {
			smerr.AddError(ctx, &resp.Diagnostics, err, "operation", "setting CloudWatch OTel Enrichment filters")
			return
		}
	}

	if _, err := waitOTelEnrichmentReady(ctx, conn, r.CreateTimeout(ctx, data.Timeouts)); err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, "operation", "waiting for CloudWatch OTel Enrichment start")
		return
	}

	data.ID = fwflex.StringValueToFramework(ctx, r.Meta().Region(ctx))
	normalizeOTelEnrichmentFilters(ctx, &data)

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &data))
}

func (r *otelEnrichmentResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data otelEnrichmentResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &data))
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().CloudWatchClient(ctx)

	out, err := findOTelEnrichment(ctx, conn)
	if retry.NotFound(err) {
		smerr.AddOne(ctx, &resp.Diagnostics, fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, "operation", "reading CloudWatch OTel Enrichment")
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, fwflex.Flatten(ctx, out, &data))
	if resp.Diagnostics.HasError() {
		return
	}
	normalizeOTelEnrichmentFilters(ctx, &data)

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &data))
}

func (r *otelEnrichmentResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var old, new otelEnrichmentResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &old))
	if resp.Diagnostics.HasError() {
		return
	}
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Plan.Get(ctx, &new))
	if resp.Diagnostics.HasError() {
		return
	}

	if !new.IncludeFilters.Equal(old.IncludeFilters) || !new.ExcludeFilters.Equal(old.ExcludeFilters) {
		conn := r.Meta().CloudWatchClient(ctx)

		if err := putOTelEnrichmentFilters(ctx, conn, &new); err != nil {
			smerr.AddError(ctx, &resp.Diagnostics, err, "operation", "updating CloudWatch OTel Enrichment filters")
			return
		}
	}

	normalizeOTelEnrichmentFilters(ctx, &new)

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &new))
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

	if _, err := waitOTelEnrichmentDeleted(ctx, conn, r.DeleteTimeout(ctx, data.Timeouts)); err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, "operation", "waiting for CloudWatch OTel Enrichment stop")
		return
	}
}

// putOTelEnrichmentFilters replaces the stored filters with those in data.
//
// UpdateOTelEnrichment replaces rather than merges, and the two lists move as a
// pair: whatever the request omits is cleared. Expanding the full model on every
// call therefore expresses the configuration exactly, and an empty model clears
// both lists, reverting enrichment to every supported namespace.
func putOTelEnrichmentFilters(ctx context.Context, conn *cloudwatch.Client, data *otelEnrichmentResourceModel) error {
	var input cloudwatch.UpdateOTelEnrichmentInput
	if diags := fwflex.Expand(ctx, data, &input); diags.HasError() {
		return fwdiag.DiagnosticsError(diags)
	}

	if _, err := conn.UpdateOTelEnrichment(ctx, &input); err != nil {
		return smarterr.NewError(err)
	}

	return nil
}

// normalizeOTelEnrichmentFilters rewrites null filter sets as empty sets.
//
// The service treats an omitted list, a null list and an empty list as
// identical, and returns nothing at all when no filters are stored, which
// flattens to null. A configuration with no filter blocks decodes to an empty
// set rather than null, so leaving the flattened value null would report a
// permanent difference between configuration and state.
func normalizeOTelEnrichmentFilters(ctx context.Context, data *otelEnrichmentResourceModel) {
	empty := fwtypes.NewSetNestedObjectValueOfValueSliceMust(ctx, []otelEnrichmentMetricSelectorModel{})

	if data.IncludeFilters.IsNull() {
		data.IncludeFilters = empty
	}
	if data.ExcludeFilters.IsNull() {
		data.ExcludeFilters = empty
	}
}

func findOTelEnrichmentStatus(ctx context.Context, conn *cloudwatch.Client, input *cloudwatch.GetOTelEnrichmentInput) (*cloudwatch.GetOTelEnrichmentOutput, error) {
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

func findOTelEnrichment(ctx context.Context, conn *cloudwatch.Client) (*cloudwatch.GetOTelEnrichmentOutput, error) {
	var input cloudwatch.GetOTelEnrichmentInput
	out, err := findOTelEnrichmentStatus(ctx, conn, &input)
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

func statusOTelEnrichment(conn *cloudwatch.Client) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		var input cloudwatch.GetOTelEnrichmentInput
		out, err := findOTelEnrichmentStatus(ctx, conn, &input)
		if retry.NotFound(err) {
			return nil, "", nil
		}
		if err != nil {
			return nil, "", smarterr.NewError(err)
		}

		return out, string(out.Status), nil
	}
}

func waitOTelEnrichmentReady(ctx context.Context, conn *cloudwatch.Client, timeout time.Duration) (*cloudwatch.GetOTelEnrichmentOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending:    []string{string(awstypes.OTelEnrichmentStatusStopped)},
		Target:     []string{string(awstypes.OTelEnrichmentStatusRunning)},
		Refresh:    statusOTelEnrichment(conn),
		Timeout:    timeout,
		MinTimeout: 2 * time.Second,
	}

	out, err := stateConf.WaitForStateContext(ctx)
	if v, ok := out.(*cloudwatch.GetOTelEnrichmentOutput); ok {
		return v, smarterr.NewError(err)
	}

	return nil, smarterr.NewError(err)
}

func waitOTelEnrichmentDeleted(ctx context.Context, conn *cloudwatch.Client, timeout time.Duration) (*cloudwatch.GetOTelEnrichmentOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending:    []string{string(awstypes.OTelEnrichmentStatusRunning)},
		Target:     []string{string(awstypes.OTelEnrichmentStatusStopped), ""},
		Refresh:    statusOTelEnrichment(conn),
		Timeout:    timeout,
		MinTimeout: 2 * time.Second,
	}

	out, err := stateConf.WaitForStateContext(ctx)
	if v, ok := out.(*cloudwatch.GetOTelEnrichmentOutput); ok {
		return v, smarterr.NewError(err)
	}

	return nil, smarterr.NewError(err)
}

type otelEnrichmentResourceModel struct {
	framework.WithRegionModel
	ExcludeFilters fwtypes.SetNestedObjectValueOf[otelEnrichmentMetricSelectorModel] `tfsdk:"exclude_filters"`
	ID             types.String                                                      `tfsdk:"id"`
	IncludeFilters fwtypes.SetNestedObjectValueOf[otelEnrichmentMetricSelectorModel] `tfsdk:"include_filters"`
	Timeouts       timeouts.Value                                                    `tfsdk:"timeouts"`
}

type otelEnrichmentMetricSelectorModel struct {
	MetricNames fwtypes.SetOfString `tfsdk:"metric_names"`
	Namespace   types.String        `tfsdk:"namespace"`
}
