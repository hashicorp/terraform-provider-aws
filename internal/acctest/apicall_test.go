// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package acctest

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/conns/apicall"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
)

func TestCheckAPICallMade(t *testing.T) {
	t.Parallel()

	rec := apicall.NewRecorder()
	rec.Record("Pinpoint", "GetApp", nil)

	if err := CheckAPICallMade(rec, nil, "Pinpoint", "GetApp")(nil); err != nil {
		t.Errorf("expected pass, got: %v", err)
	}

	err := CheckAPICallMade(rec, nil, "Pinpoint", "GetApplicationSettings")(nil)
	if err == nil {
		t.Fatal("expected failure for un-recorded call")
	}
	if !strings.Contains(err.Error(), "Pinpoint.GetApplicationSettings") {
		t.Errorf("error missing service.operation: %v", err)
	}
	if !strings.Contains(err.Error(), "Pinpoint.GetApp") {
		t.Errorf("error missing recorded calls list: %v", err)
	}
}

func TestCheckAPICallNotMade(t *testing.T) {
	t.Parallel()

	rec := apicall.NewRecorder()
	rec.Record("Pinpoint", "GetApp", nil)

	if err := CheckAPICallNotMade(rec, nil, "Pinpoint", "GetApplicationSettings")(nil); err != nil {
		t.Errorf("expected pass, got: %v", err)
	}

	err := CheckAPICallNotMade(rec, nil, "Pinpoint", "GetApp")(nil)
	if err == nil {
		t.Fatal("expected failure for recorded call")
	}
	if !strings.Contains(err.Error(), "Pinpoint.GetApp") {
		t.Errorf("error missing service.operation: %v", err)
	}
}

// TestCheckAPICall_PointerCursorMidFlight exercises the PreConfig pattern:
// the test author declares the cursor variable, builds the Check slice
// referencing &cursor, and only later (in PreConfig) populates *cursor.
func TestCheckAPICall_PointerCursorMidFlight(t *testing.T) {
	t.Parallel()

	rec := apicall.NewRecorder()
	rec.Record("Pinpoint", "GetApp", nil)

	var mark apicall.Cursor
	check := CheckAPICallNotMade(rec, &mark, "Pinpoint", "GetApp")

	// Before the mark is set: cursor is 0, GetApp WAS recorded -> failure.
	if err := check(nil); err == nil {
		t.Fatal("expected failure when mark is at 0 and call was recorded")
	}

	// Simulate PreConfig assignment: mark advances past the existing call.
	mark = rec.Mark()

	if err := check(nil); err != nil {
		t.Errorf("expected pass after mark advanced, got: %v", err)
	}
}

func TestCheckAPICall_NilRecorder(t *testing.T) {
	t.Parallel()

	if err := CheckAPICallMade(nil, nil, "S3", "GetObject")(nil); err == nil {
		t.Error("CheckAPICallMade(nil) returned no error")
	}
	if err := CheckAPICallNotMade(nil, nil, "S3", "GetObject")(nil); err == nil {
		t.Error("CheckAPICallNotMade(nil) returned no error")
	}
}

func TestFormatCalls(t *testing.T) {
	t.Parallel()

	if got := formatCalls(nil); got != "(none)" {
		t.Errorf("formatCalls(nil) = %q, want (none)", got)
	}
	if got := formatCalls([]apicall.Call{}); got != "(none)" {
		t.Errorf("formatCalls([]) = %q, want (none)", got)
	}

	got := formatCalls([]apicall.Call{
		{Service: "Pinpoint", Operation: "GetApp"},
		{Service: "Pinpoint", Operation: "DeleteApp", Err: errors.New("boom")},
	})
	if !strings.Contains(got, "Pinpoint.GetApp") {
		t.Errorf("formatted output missing GetApp: %s", got)
	}
	if !strings.Contains(got, "Pinpoint.DeleteApp(err=boom)") {
		t.Errorf("formatted output missing DeleteApp with error: %s", got)
	}
}

// TestAPICallRecorderWrapper verifies the wrapper attaches the recorder
// to a non-nil *conns.AWSClient and tolerates the nil/error return path.
func TestAPICallRecorderWrapper(t *testing.T) {
	t.Parallel()

	rec := apicall.NewRecorder()

	t.Run("attaches recorder on success", func(t *testing.T) {
		t.Parallel()
		client := &conns.AWSClient{}
		original := func(_ context.Context, _ *schema.ResourceData) (any, diag.Diagnostics) {
			return client, nil
		}
		wrapped := APICallRecorderWrapper(rec)(original)

		v, diags := wrapped(context.Background(), nil)
		if diags.HasError() {
			t.Fatalf("unexpected diagnostics: %v", diags)
		}
		if v != client {
			t.Errorf("returned %p, want %p", v, client)
		}
		if client.CallRecorder() != rec {
			t.Errorf("CallRecorder() = %p, want %p", client.CallRecorder(), rec)
		}
	})

	t.Run("tolerates nil meta", func(t *testing.T) {
		t.Parallel()
		original := func(_ context.Context, _ *schema.ResourceData) (any, diag.Diagnostics) {
			return nil, sdkdiag.AppendErrorf(nil, "configure failed")
		}
		wrapped := APICallRecorderWrapper(rec)(original)

		v, diags := wrapped(context.Background(), nil)
		if !diags.HasError() {
			t.Error("expected error diagnostics")
		}
		if v != nil {
			t.Errorf("returned %v, want nil", v)
		}
	})

	t.Run("tolerates typed-nil meta", func(t *testing.T) {
		t.Parallel()
		var typedNil *conns.AWSClient
		original := func(_ context.Context, _ *schema.ResourceData) (any, diag.Diagnostics) {
			return typedNil, nil
		}
		wrapped := APICallRecorderWrapper(rec)(original)

		// Must not panic.
		_, _ = wrapped(context.Background(), nil)
	})
}
