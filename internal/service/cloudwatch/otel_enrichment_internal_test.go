// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package cloudwatch

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatch"
	awstypes "github.com/aws/aws-sdk-go-v2/service/cloudwatch/types"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
)

// TestOTelEnrichmentFilterRoundTrip checks that filters survive a flatten and
// expand cycle, which is what Read and the StartOTelEnrichment no-op comparison
// depend on.
func TestOTelEnrichmentFilterRoundTrip(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	out := &cloudwatch.GetOTelEnrichmentOutput{
		Status: awstypes.OTelEnrichmentStatusRunning,
		IncludeFilters: []awstypes.OTelEnrichmentMetricSelector{
			{Namespace: aws.String("AWS/EC2")},
			{Namespace: aws.String("AWS/RDS"), MetricNames: []string{"CPUUtilization", "DatabaseConnections"}},
		},
		ExcludeFilters: []awstypes.OTelEnrichmentMetricSelector{
			{Namespace: aws.String("AWS/EC2"), MetricNames: []string{"NetworkPacketsIn"}},
		},
	}

	var data otelEnrichmentResourceModel
	if diags := fwflex.Flatten(ctx, out, &data); diags.HasError() {
		t.Fatalf("unexpected error flattening: %v", diags.Errors())
	}

	if got, want := data.IncludeFilters.Length(fwtypes.CollectionLengthUnhandledAsZero), 2; got != want {
		t.Errorf("include_filters: got %d selectors, want %d", got, want)
	}
	if got, want := data.ExcludeFilters.Length(fwtypes.CollectionLengthUnhandledAsZero), 1; got != want {
		t.Errorf("exclude_filters: got %d selectors, want %d", got, want)
	}

	var input cloudwatch.UpdateOTelEnrichmentInput
	if diags := fwflex.Expand(ctx, &data, &input); diags.HasError() {
		t.Fatalf("unexpected error expanding: %v", diags.Errors())
	}

	if got, want := len(input.IncludeFilters), 2; got != want {
		t.Errorf("IncludeFilters: got %d selectors, want %d", got, want)
	}
	if got, want := len(input.ExcludeFilters), 1; got != want {
		t.Errorf("ExcludeFilters: got %d selectors, want %d", got, want)
	}

	// Metric names must survive, since they are the part most easily lost.
	for _, selector := range input.IncludeFilters {
		if aws.ToString(selector.Namespace) != "AWS/RDS" {
			continue
		}
		if got, want := len(selector.MetricNames), 2; got != want {
			t.Errorf("AWS/RDS MetricNames: got %d, want %d", got, want)
		}
	}
}

// TestOTelEnrichmentNormalizeFilters guards against a perpetual difference.
//
// GetOTelEnrichment omits the filter lists entirely when none are stored, which
// flattens to a null set, whereas a configuration with no filter blocks decodes
// to an empty set. Without normalization the two never compare equal and every
// plan reports a change.
func TestOTelEnrichmentNormalizeFilters(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	var data otelEnrichmentResourceModel
	out := &cloudwatch.GetOTelEnrichmentOutput{Status: awstypes.OTelEnrichmentStatusRunning}

	if diags := fwflex.Flatten(ctx, out, &data); diags.HasError() {
		t.Fatalf("unexpected error flattening: %v", diags.Errors())
	}

	if !data.IncludeFilters.IsNull() || !data.ExcludeFilters.IsNull() {
		t.Fatal("expected absent filters to flatten to null sets")
	}

	normalizeOTelEnrichmentFilters(ctx, &data)

	if data.IncludeFilters.IsNull() {
		t.Error("include_filters is still null after normalization")
	}
	if data.ExcludeFilters.IsNull() {
		t.Error("exclude_filters is still null after normalization")
	}
	if got := data.IncludeFilters.Length(fwtypes.CollectionLengthUnhandledAsZero); got != 0 {
		t.Errorf("include_filters: got %d selectors, want 0", got)
	}
	if got := data.ExcludeFilters.Length(fwtypes.CollectionLengthUnhandledAsZero); got != 0 {
		t.Errorf("exclude_filters: got %d selectors, want 0", got)
	}

	// An empty model must clear both lists on the wire, which is how enrichment
	// reverts to every supported namespace.
	var input cloudwatch.UpdateOTelEnrichmentInput
	if diags := fwflex.Expand(ctx, &data, &input); diags.HasError() {
		t.Fatalf("unexpected error expanding: %v", diags.Errors())
	}

	if len(input.IncludeFilters) != 0 || len(input.ExcludeFilters) != 0 {
		t.Errorf("expected both filter lists to be cleared, got include=%d exclude=%d",
			len(input.IncludeFilters), len(input.ExcludeFilters))
	}
}
