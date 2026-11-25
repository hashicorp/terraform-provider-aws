// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package odb

import (
	"encoding/json"
	"io"
	"net/http"
	"slices"
	"strings"
	"testing"

	"github.com/hashicorp/aws-sdk-go-base/v2/endpoints"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/awsv2"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestSweepAutonomousDatabases(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		pages   []string
		wantIDs []string
	}{
		"all pages and test names only": {
			pages: []string{
				`{"autonomousDatabases":[{"autonomousDatabaseId":"adb-first","displayName":"tf-acc-test-first"},{"autonomousDatabaseId":"adb-production","displayName":"production"}],"nextToken":"second-page"}`,
				`{"autonomousDatabases":[{"autonomousDatabaseId":"adb-second","displayName":"tf-acc-test-second"},{"displayName":"tf-acc-test-without-id"},{"autonomousDatabaseId":"adb-without-name"},{"autonomousDatabaseId":"adb-terminated","displayName":"tf-acc-test-terminated","status":"TERMINATED"}]}`,
			},
			wantIDs: []string{"adb-first", "adb-second"},
		},
		"empty page with continuation token": {
			pages:   []string{`{"autonomousDatabases":[],"nextToken":"second-page"}`, `{"autonomousDatabases":[{"autonomousDatabaseId":"adb-second","displayName":"tf-acc-test-second"}]}`},
			wantIDs: []string{"adb-second"},
		},
		"nonstandard prefixes are not swept": {
			pages: []string{`{"autonomousDatabases":[{"autonomousDatabaseId":"adb-old-prefix","displayName":"tf-odb-adbs-example"},{"autonomousDatabaseId":"adb-prefix-only","displayName":"tf-acc-test"},{"autonomousDatabaseId":"adb-prefix-lookalike","displayName":"tf-acc-testexample"},{"autonomousDatabaseId":"adb-prefix-suffix","displayName":"production-tf-acc-test-example"}]}`},
		},
		"empty inventory": {pages: []string{`{"autonomousDatabases":[]}`}},
	}
	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			ctx := t.Context()
			var pageCount int
			var deletedIDs []string
			client := newAutonomousDatabaseSweepClient(t, autonomousDatabaseSweepTransport(func(request *http.Request) (*http.Response, error) {
				var input map[string]string
				if err := json.NewDecoder(request.Body).Decode(&input); err != nil {
					t.Fatal(err)
				}
				status := http.StatusOK
				var body string
				switch request.Header.Get("X-Amz-Target") {
				case "Odb.ListAutonomousDatabases":
					wantToken := ""
					if pageCount > 0 {
						wantToken = "second-page"
					}
					if got := input["nextToken"]; got != wantToken {
						t.Errorf("nextToken = %q, want %q", got, wantToken)
					}
					if pageCount >= len(testCase.pages) {
						t.Fatal("unexpected extra list request")
					}
					body = testCase.pages[pageCount]
					pageCount++
				case "Odb.DeleteAutonomousDatabase":
					deletedIDs = append(deletedIDs, input["autonomousDatabaseId"])
					status = http.StatusBadRequest
					body = `{"__type":"ResourceNotFoundException","message":"already deleted"}`
				default:
					t.Fatalf("unexpected AWS operation: %s", request.Header.Get("X-Amz-Target"))
				}
				return &http.Response{StatusCode: status, Header: http.Header{"Content-Type": {"application/x-amz-json-1.0"}}, Body: io.NopCloser(strings.NewReader(body)), Request: request}, nil
			}))

			resources, err := sweepAutonomousDatabases(ctx, client)
			if err != nil {
				t.Fatal(err)
			}
			if pageCount != len(testCase.pages) {
				t.Errorf("listed %d pages, want %d", pageCount, len(testCase.pages))
			}
			if len(resources) != len(testCase.wantIDs) {
				t.Fatalf("sweepable count = %d, want %d", len(resources), len(testCase.wantIDs))
			}
			for _, resource := range resources {
				if err := resource.Delete(ctx); err != nil {
					t.Fatal(err)
				}
			}
			if !slices.Equal(deletedIDs, testCase.wantIDs) {
				t.Errorf("deleted IDs = %v, want %v", deletedIDs, testCase.wantIDs)
			}
		})
	}
}

func TestSweepAutonomousDatabases_listErrors(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		body      string
		wantError string
		wantSkip  bool
		firstPage string
	}{
		"failure discards partial inventory": {
			firstPage: `{"autonomousDatabases":[{"autonomousDatabaseId":"adb-first","displayName":"tf-acc-test-first"}],"nextToken":"second-page"}`,
			body:      `{"__type":"ValidationException","message":"listing failed"}`,
			wantError: "ValidationException",
		},
		"unsupported account remains skippable": {
			body:      `{"__type":"AccessDeniedException","message":"listing failed"}`,
			wantError: "AccessDeniedException",
			wantSkip:  true,
		},
		"unsupported region": {
			body: `{"__type":"ValidationException","message":"Oracle Database@AWS doesn't support Autonomous Database in the current Region. Switch to a Region where Autonomous Database is available and try again."}`,
		},
		"unrelated validation error": {
			body:      `{"__type":"ValidationException","message":"listing failed"}`,
			wantError: "ValidationException",
		},
		"unrelated error with same message": {
			body:      `{"__type":"InternalServerException","message":"Oracle Database@AWS doesn't support Autonomous Database in the current Region"}`,
			wantError: "InternalServerException",
		},
	}
	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			ctx := t.Context()
			requests := 0
			client := newAutonomousDatabaseSweepClient(t, autonomousDatabaseSweepTransport(func(request *http.Request) (*http.Response, error) {
				if got := request.Header.Get("X-Amz-Target"); got != "Odb.ListAutonomousDatabases" {
					t.Fatalf("unexpected AWS operation: %s", got)
				}
				requests++
				status, body := http.StatusBadRequest, testCase.body
				if requests == 1 && testCase.firstPage != "" {
					status, body = http.StatusOK, testCase.firstPage
				}
				return &http.Response{StatusCode: status, Header: http.Header{"Content-Type": {"application/x-amz-json-1.0"}}, Body: io.NopCloser(strings.NewReader(body)), Request: request}, nil
			}))

			resources, err := sweepAutonomousDatabases(ctx, client)
			if testCase.wantError == "" {
				if err != nil {
					t.Fatalf("unexpected error: %s", err)
				}
			} else if err == nil || !strings.Contains(err.Error(), testCase.wantError) {
				t.Fatalf("error = %v, want %s", err, testCase.wantError)
			}
			if resources != nil {
				t.Errorf("resources = %v, want nil", resources)
			}
			if got := awsv2.SkipSweepError(err); got != testCase.wantSkip {
				t.Errorf("SkipSweepError = %t, want %t", got, testCase.wantSkip)
			}
			wantRequests := 1
			if testCase.firstPage != "" {
				wantRequests++
			}
			if requests != wantRequests {
				t.Errorf("requests = %d, want %d", requests, wantRequests)
			}
		})
	}
}

func newAutonomousDatabaseSweepClient(t *testing.T, transport http.RoundTripper) *conns.AWSClient {
	t.Helper()
	ctx := t.Context()
	client := new(conns.AWSClient)
	client.SetHTTPClient(ctx, &http.Client{Transport: transport})
	client.SetServicePackages(ctx, map[string]conns.ServicePackage{names.ODB: &servicePackage{}})
	config := conns.Config{AccessKey: "test", SecretKey: "test", Region: endpoints.UsEast1RegionID, SkipCredsValidation: true, SkipRequestingAccountId: true, MaxRetries: 0, SharedConfigFiles: []string{}, SharedCredentialsFiles: []string{}}
	client, diags := config.ConfigureProvider(ctx, client)
	if diags.HasError() {
		t.Fatal(diags)
	}
	return client
}

type autonomousDatabaseSweepTransport func(*http.Request) (*http.Response, error)

func (f autonomousDatabaseSweepTransport) RoundTrip(request *http.Request) (*http.Response, error) {
	return f(request)
}
