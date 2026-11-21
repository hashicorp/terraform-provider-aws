// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package framework

import (
	"context"
	"iter"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/YakDriver/smarterr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/list"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// smarterr resolves the "listing" happening by matching a ".List" call stack
// frame. These tests check that the shared List still matches, despite being a
// generic method outside any service package.

const listSpecSmerrConfig = `
template "error_summary" {
  format = "{{.happening}} {{.service}} {{.resource}}: {{.error}}"
}
template "diagnostic_summary" {
  format = "{{.happening}} {{.service}} {{.resource}}: {{.diag.summary}}"
}
token "happening" {
  source        = "call_stack"
  stack_matches = ["fw_list"]
}
token "service"  { arg = "service_name" }
token "resource" { arg = "resource_name" }
token "error"    { source = "error" }
token "diag"     { source = "diagnostic" }
stack_match "fw_list" {
  called_from = ".*\\.List(\\.func[0-9]+)?$"
  display     = "listing"
}
`

func setListSpecSmerrFS(t *testing.T) {
	t.Helper()
	smarterr.SetFS(&smarterr.WrappedFS{FS: fstest.MapFS{
		"smarterr/smarterr.hcl": &fstest.MapFile{Data: []byte(listSpecSmerrConfig)},
	}}, ".")
}

// listSpecSmerrCtx carries the service and resource that smerr injects as
// keyvals, standing in for the interceptor-populated context at runtime.
func listSpecSmerrCtx() context.Context {
	return conns.NewResourceContext(context.Background(), names.AMP, "Anomaly Detector", "aws_prometheus_anomaly_detector", "")
}

func wantListSpecService() string {
	if v, err := names.HumanFriendly(names.AMP); err == nil {
		return v
	}

	return names.AMP
}

// firstErrorSummary drives List with an enriched context and returns the
// summary of the first error diagnostic the stream produces.
func firstErrorSummary(t *testing.T, l *sdkListResourceFromSpec[listSpecItem, struct{}]) string {
	t.Helper()

	stream := &list.ListResultsStream{}
	l.List(listSpecSmerrCtx(), listSpecRequest(), stream)

	if stream.Results == nil {
		t.Fatal("List did not set stream.Results")
	}

	for result := range stream.Results {
		if result.Diagnostics.HasError() {
			return result.Diagnostics.Errors()[0].Summary()
		}
	}
	t.Fatal("stream produced no error diagnostic")

	return ""
}

func assertEnriched(t *testing.T, summary string, want ...string) {
	t.Helper()

	want = append(want, "listing", wantListSpecService(), "Anomaly Detector")
	for _, w := range want {
		if !strings.Contains(summary, w) {
			t.Errorf("summary %q is missing %q", summary, w)
		}
	}
}

// A pagination error surfaces from inside the streamed iterator, which is the
// deepest the call stack gets.
//
//nolint:paralleltest // smarterr.SetFS sets process-global state.
func TestSDKListSpecEnrichesIteratorErrors(t *testing.T) {
	setListSpecSmerrFS(t)

	l := listSpecResource(nil, errListSpec, nil, nil)
	assertEnriched(t, firstErrorSummary(t, l), errListSpec.Error())
}

// Setup diagnostics are reported from the List method body, before streaming,
// so they exercise a different frame than the iterator path.
//
//nolint:paralleltest // smarterr.SetFS sets process-global state.
func TestSDKListSpecEnrichesSetupDiagnostics(t *testing.T) {
	setListSpecSmerrFS(t)

	l := listSpecResource(nil, nil, nil, nil)
	l.spec.Items = func(context.Context, *conns.AWSClient, struct{}) (iter.Seq2[listSpecItem, error], diag.Diagnostics) {
		var diags diag.Diagnostics
		diags.AddError("expanding input", "bad filter")

		return nil, diags
	}

	assertEnriched(t, firstErrorSummary(t, l), "expanding input")
}
