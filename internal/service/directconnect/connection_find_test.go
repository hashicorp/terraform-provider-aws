// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package directconnect

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/directconnect"
	awstypes "github.com/aws/aws-sdk-go-v2/service/directconnect/types"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
)

// findConnections reads through the generated describeConnectionsPages pager.
// DescribeConnections silently caps a single response at 100 connections, so an
// account with more than that gets a truncated list unless every page is
// followed. These tests drive a stubbed client so that the shipped pager is
// exercised rather than a reimplementation of it.

func TestFindConnectionsPaginates(t *testing.T) {
	t.Parallel()

	ctx := t.Context()
	conn := testStubbedConnectionsClient(t,
		[]awstypes.Connection{
			{ConnectionId: aws.String("dxcon-ff000001")},
			{ConnectionId: aws.String("dxcon-ff000002")},
		},
		[]awstypes.Connection{
			{ConnectionId: aws.String("dxcon-ff000003")},
		},
	)

	output, err := findConnections(ctx, conn, &directconnect.DescribeConnectionsInput{}, tfslices.PredicateTrue[*awstypes.Connection]())
	if err != nil {
		t.Fatalf("finding connections: %s", err)
	}

	want := []string{"dxcon-ff000001", "dxcon-ff000002", "dxcon-ff000003"}

	if got, want := len(output), len(want); got != want {
		t.Fatalf("connection count = %d, want %d (connections past the first page were dropped)", got, want)
	}

	for i, want := range want {
		if got := aws.ToString(output[i].ConnectionId); got != want {
			t.Errorf("connection %d = %q, want %q", i, got, want)
		}
	}
}

func TestFindConnectionsFilterSpansPages(t *testing.T) {
	t.Parallel()

	ctx := t.Context()

	// The match is on the second page: a filter applied per-page and short-circuited
	// on the first empty result would miss it.
	conn := testStubbedConnectionsClient(t,
		[]awstypes.Connection{
			{ConnectionId: aws.String("dxcon-ff000001"), ConnectionName: aws.String("other")},
		},
		[]awstypes.Connection{
			{ConnectionId: aws.String("dxcon-ff000002"), ConnectionName: aws.String("wanted")},
			{ConnectionId: aws.String("dxcon-ff000003"), ConnectionName: aws.String("other")},
		},
	)

	output, err := findConnections(ctx, conn, &directconnect.DescribeConnectionsInput{}, func(v *awstypes.Connection) bool {
		return aws.ToString(v.ConnectionName) == "wanted"
	})
	if err != nil {
		t.Fatalf("finding connections: %s", err)
	}

	if got, want := len(output), 1; got != want {
		t.Fatalf("connection count = %d, want %d", got, want)
	}

	if got, want := aws.ToString(output[0].ConnectionId), "dxcon-ff000002"; got != want {
		t.Errorf("connection ID = %q, want %q", got, want)
	}
}

func TestFindConnectionsNoMatchIsNotAnError(t *testing.T) {
	t.Parallel()

	ctx := t.Context()
	conn := testStubbedConnectionsClient(t, []awstypes.Connection{
		{ConnectionId: aws.String("dxcon-ff000001")},
	})

	output, err := findConnections(ctx, conn, &directconnect.DescribeConnectionsInput{}, func(*awstypes.Connection) bool {
		return false
	})

	// Nothing matched the filter: that is an empty result, not a failure. The
	// plural data source reports an empty list for this, so it must not surface
	// as an error.
	if err != nil {
		t.Fatalf("finding connections: %s", err)
	}

	if got, want := len(output), 0; got != want {
		t.Errorf("connection count = %d, want %d", got, want)
	}
}

func TestFindConnectionEmptyIsNotFound(t *testing.T) {
	t.Parallel()

	ctx := t.Context()
	conn := testStubbedConnectionsClient(t, []awstypes.Connection{})

	// findConnections yields an empty slice and no error, so the NotFound that
	// the singular data source and the aws_dx_connection resource depend on comes
	// from findConnection's AssertSingleValueResult. Pin it here: those callers
	// treat a non-NotFound error as a hard failure.
	_, err := findConnection(ctx, conn, &directconnect.DescribeConnectionsInput{}, tfslices.PredicateTrue[*awstypes.Connection]())

	if err == nil {
		t.Fatal("no error; want a not-found error")
	}

	if !retry.NotFound(err) {
		t.Errorf("error = %v; want one that retry.NotFound matches", err)
	}
}

// testStubbedConnectionsClient returns a Direct Connect client whose
// DescribeConnections responses are the given pages, joined by a nextToken.
func testStubbedConnectionsClient(t *testing.T, pages ...[]awstypes.Connection) *directconnect.Client {
	t.Helper()

	var page int

	return directconnect.New(directconnect.Options{
		Region:      "us-west-2",
		Credentials: aws.AnonymousCredentials{},
		HTTPClient: testHTTPClientFunc(func(*http.Request) (*http.Response, error) {
			// Keys are the API's wire names, not the Go field names.
			body := map[string]any{}

			if page < len(pages) {
				connections := make([]map[string]any, 0, len(pages[page]))
				for _, v := range pages[page] {
					connections = append(connections, map[string]any{
						"connectionId":   aws.ToString(v.ConnectionId),
						"connectionName": aws.ToString(v.ConnectionName),
					})
				}
				body["connections"] = connections

				if page < len(pages)-1 {
					body["nextToken"] = "page-token"
				}
				page++
			}

			b, err := json.Marshal(body)
			if err != nil {
				return nil, err
			}

			return &http.Response{
				StatusCode: http.StatusOK,
				Header:     http.Header{"Content-Type": []string{"application/x-amz-json-1.1"}},
				Body:       io.NopCloser(bytes.NewReader(b)),
			}, nil
		}),
	})
}

type testHTTPClientFunc func(*http.Request) (*http.Response, error)

func (f testHTTPClientFunc) Do(req *http.Request) (*http.Response, error) {
	return f(req)
}
