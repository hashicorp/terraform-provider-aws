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
		pages     []string
		errorCode string
		wantIDs   []string
		wantSkip  bool
	}{
		"all pages and test names only": {
			pages: []string{
				`{"autonomousDatabases":[{"autonomousDatabaseId":"adb-first","displayName":"tf-acc-test-first"},{"autonomousDatabaseId":"adb-production","displayName":"production"}],"nextToken":"second-page"}`,
				`{"autonomousDatabases":[{"autonomousDatabaseId":"adb-second","displayName":"tf-odb-adbs-second"},{"displayName":"tf-acc-test-without-id"},{"autonomousDatabaseId":"adb-without-name"},{"autonomousDatabaseId":"adb-terminated","displayName":"tf-acc-test-terminated","status":"TERMINATED"}]}`,
			},
			wantIDs: []string{"adb-first", "adb-second"},
		},
		"empty page with continuation token": {
			pages:   []string{`{"autonomousDatabases":[],"nextToken":"second-page"}`, `{"autonomousDatabases":[{"autonomousDatabaseId":"adb-second","displayName":"tf-acc-test-second"}]}`},
			wantIDs: []string{"adb-second"},
		},
		"empty inventory":                       {pages: []string{`{"autonomousDatabases":[]}`}},
		"failure discards partial inventory":    {pages: []string{`{"autonomousDatabases":[{"autonomousDatabaseId":"adb-first","displayName":"tf-acc-test-first"}],"nextToken":"second-page"}`}, errorCode: "ValidationException"},
		"unsupported account remains skippable": {errorCode: "AccessDeniedException", wantSkip: true},
	}
	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			ctx := t.Context()
			var pageCount int
			var deletedIDs []string
			client := new(conns.AWSClient)
			client.SetHTTPClient(ctx, &http.Client{Transport: autonomousDatabaseSweepTransport(func(request *http.Request) (*http.Response, error) {
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
					if pageCount < len(testCase.pages) {
						body = testCase.pages[pageCount]
					} else {
						if testCase.errorCode == "" {
							t.Fatal("unexpected extra list request")
						}
						status = http.StatusBadRequest
						body = `{"__type":"` + testCase.errorCode + `","message":"listing failed"}`
					}
					pageCount++
				case "Odb.DeleteAutonomousDatabase":
					deletedIDs = append(deletedIDs, input["autonomousDatabaseId"])
					status = http.StatusBadRequest
					body = `{"__type":"ResourceNotFoundException","message":"already deleted"}`
				default:
					t.Fatalf("unexpected AWS operation: %s", request.Header.Get("X-Amz-Target"))
				}
				return &http.Response{StatusCode: status, Header: http.Header{"Content-Type": {"application/x-amz-json-1.0"}}, Body: io.NopCloser(strings.NewReader(body)), Request: request}, nil
			})})
			client.SetServicePackages(ctx, map[string]conns.ServicePackage{names.ODB: &servicePackage{}})
			config := conns.Config{AccessKey: "test", SecretKey: "test", Region: endpoints.UsEast1RegionID, SkipCredsValidation: true, SkipRequestingAccountId: true, MaxRetries: 0, SharedConfigFiles: []string{}, SharedCredentialsFiles: []string{}}
			client, diags := config.ConfigureProvider(ctx, client)
			if diags.HasError() {
				t.Fatal(diags)
			}

			resources, err := sweepAutonomousDatabases(ctx, client)
			if testCase.errorCode != "" {
				if err == nil || !strings.Contains(err.Error(), testCase.errorCode) {
					t.Fatalf("error = %v, want %s", err, testCase.errorCode)
				}
				if resources != nil {
					t.Fatalf("partial inventory must be discarded on list failure: %v", resources)
				}
				if got := awsv2.SkipSweepError(err); got != testCase.wantSkip {
					t.Errorf("SkipSweepError = %t, want %t", got, testCase.wantSkip)
				}
				return
			}
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

type autonomousDatabaseSweepTransport func(*http.Request) (*http.Response, error)

func (f autonomousDatabaseSweepTransport) RoundTrip(request *http.Request) (*http.Response, error) {
	return f(request)
}
