// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package smerr_test

import (
	"context"
	"errors"
	"fmt"
	"iter"
	"testing"
	"testing/fstest"

	"github.com/YakDriver/smarterr"
	fwdiag "github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/list"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// listTestConfig is an in-memory global smarterr config that exercises the three
// things the smerr list wrappers must deliver: the "listing" happening (resolved
// from the List method/closure frame), the service and resource tokens (injected
// by smerr's injectContext), and the underlying error/diagnostic.
const listTestConfig = `
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

func setListTestFS(t *testing.T) {
	t.Helper()
	fsys := fstest.MapFS{
		"smarterr/smarterr.hcl": &fstest.MapFile{Data: []byte(listTestConfig)},
	}
	smarterr.SetFS(&smarterr.WrappedFS{FS: fsys}, ".")
}

// resourceCtx returns a context carrying the AWS service/resource info that
// smerr.injectContext reads. wantService mirrors injectContext's own logic.
func resourceCtx() context.Context {
	return conns.NewResourceContext(context.Background(), names.AMP, "Anomaly Detector", "aws_prometheus_anomaly_detector", "")
}

func wantService() string {
	if v, err := names.HumanFriendly(names.AMP); err == nil {
		return v
	}
	return names.AMP
}

func collectListResults(seq iter.Seq[list.ListResult]) []list.ListResult {
	var got []list.ListResult
	for r := range seq {
		got = append(got, r)
	}
	return got
}

// listClosure surfaces a fatal error from inside its iterator closure (surface b):
// smerr.NewListResultError runs while ...List.func1 is on the stack.
type listClosure struct{}

func (listClosure) List(ctx context.Context, raise error) iter.Seq[list.ListResult] {
	return func(yield func(list.ListResult) bool) {
		yield(smerr.NewListResultError(ctx, raise))
	}
}

// listBody surfaces a pre-stream fatal error from the List method body (surface a).
type listBody struct{}

func (listBody) List(ctx context.Context, raise error) iter.Seq[list.ListResult] {
	return smerr.ListStreamError(ctx, raise)
}

// listEnrich surfaces framework diagnostics (e.g., config decode) from the body.
type listEnrich struct{}

func (listEnrich) List(ctx context.Context, incoming fwdiag.Diagnostics) iter.Seq[list.ListResult] {
	return smerr.ListStreamEnrich(ctx, incoming)
}

func TestNewListResultError_InjectsContextAndHappening(t *testing.T) { //nolint:paralleltest // smarterr.SetFS sets process-global state; these tests must not run in parallel
	setListTestFS(t)

	// List returns the iterator; the framework consumes it after List has returned.
	seq := listClosure{}.List(resourceCtx(), errors.New("boom"))
	results := collectListResults(seq)

	if len(results) != 1 || !results[0].Diagnostics.HasError() {
		t.Fatalf("expected 1 error result, got %#v", results)
	}
	want := fmt.Sprintf("listing %s Anomaly Detector: boom", wantService())
	if got := results[0].Diagnostics[0].Summary(); got != want {
		t.Errorf("summary = %q, want %q", got, want)
	}
}

func TestListStreamError_InjectsContextAndHappening(t *testing.T) { //nolint:paralleltest // smarterr.SetFS sets process-global state; these tests must not run in parallel
	setListTestFS(t)

	seq := listBody{}.List(resourceCtx(), errors.New("kaboom"))
	results := collectListResults(seq)

	if len(results) != 1 || !results[0].Diagnostics.HasError() {
		t.Fatalf("expected 1 error result, got %#v", results)
	}
	want := fmt.Sprintf("listing %s Anomaly Detector: kaboom", wantService())
	if got := results[0].Diagnostics[0].Summary(); got != want {
		t.Errorf("summary = %q, want %q", got, want)
	}
}

func TestListStreamEnrich_InjectsContextAndHappening(t *testing.T) { //nolint:paralleltest // smarterr.SetFS sets process-global state; these tests must not run in parallel
	setListTestFS(t)

	var incoming fwdiag.Diagnostics
	incoming.AddError("bad config", "workspace_id is invalid")

	seq := listEnrich{}.List(resourceCtx(), incoming)
	results := collectListResults(seq)

	if len(results) != 1 || !results[0].Diagnostics.HasError() {
		t.Fatalf("expected 1 error result, got %#v", results)
	}
	want := fmt.Sprintf("listing %s Anomaly Detector: bad config", wantService())
	if got := results[0].Diagnostics[0].Summary(); got != want {
		t.Errorf("summary = %q, want %q", got, want)
	}
}
