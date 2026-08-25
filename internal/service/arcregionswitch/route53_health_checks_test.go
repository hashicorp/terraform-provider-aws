// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package arcregionswitch

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/arcregionswitch"
	awstypes "github.com/aws/aws-sdk-go-v2/service/arcregionswitch/types"
)

// mockListRoute53HealthChecksClient returns its configured pages in order,
// letting the SDK paginator drive multi-page traversal.
type mockListRoute53HealthChecksClient struct {
	pages []*arcregionswitch.ListRoute53HealthChecksOutput
	calls int
}

func (m *mockListRoute53HealthChecksClient) ListRoute53HealthChecks(_ context.Context, _ *arcregionswitch.ListRoute53HealthChecksInput, _ ...func(*arcregionswitch.Options)) (*arcregionswitch.ListRoute53HealthChecksOutput, error) {
	page := m.pages[m.calls]
	m.calls++
	return page, nil
}

func TestFindRoute53HealthChecksPagination(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	client := &mockListRoute53HealthChecksClient{
		pages: []*arcregionswitch.ListRoute53HealthChecksOutput{
			{
				// Page 1 carries a NextToken, so the paginator must fetch page 2.
				HealthChecks: []awstypes.Route53HealthCheck{
					{HealthCheckId: aws.String("hc-1")},
					{HealthCheckId: aws.String("hc-2")},
				},
				NextToken: aws.String("page-2"),
			},
			{
				// Final page: nil NextToken ends pagination.
				HealthChecks: []awstypes.Route53HealthCheck{
					{HealthCheckId: aws.String("hc-3")},
					{HealthCheckId: aws.String("hc-4")},
				},
				NextToken: nil,
			},
		},
	}

	got, err := findRoute53HealthChecks(ctx, client, &arcregionswitch.ListRoute53HealthChecksInput{
		Arn: aws.String("arn:aws:arc-region-switch::123456789012:plan/test:abc123"),
	})
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	if got, want := len(got), 4; got != want {
		t.Errorf("aggregated health checks across pages = %d, want %d", got, want)
	}
	if got, want := client.calls, 2; got != want {
		t.Errorf("ListRoute53HealthChecks calls = %d, want %d (paginator did not follow NextToken)", got, want)
	}
}
