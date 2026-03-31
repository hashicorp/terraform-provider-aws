// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package odb

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/odb"
	odbtypes "github.com/aws/aws-sdk-go-v2/service/odb/types"
)

// newTestClient creates an ODB client that points at a local httptest server.
func newTestClient(t *testing.T, handler http.HandlerFunc) *odb.Client {
	t.Helper()
	ts := httptest.NewServer(handler)
	t.Cleanup(ts.Close)

	return odb.New(odb.Options{
		Region:      "us-east-1",
		BaseEndpoint: aws.String(ts.URL),
		Credentials: credentials.NewStaticCredentialsProvider("AKID", "SECRET", "SESSION"),
	})
}

// jsonHandler returns an HTTP handler that writes JSON with the given status code.
func jsonHandler(statusCode int, body any) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/x-amz-json-1.0")
		w.WriteHeader(statusCode)
		json.NewEncoder(w).Encode(body) //nolint:errcheck
	}
}

// TestStatusNetwork_NilOutput verifies statusNetwork does not panic when
// FindOracleDBNetworkResourceByID returns an empty result (nil OdbNetwork).
func TestStatusNetwork_NilOutput(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// API returns 200 with no odbNetwork field → Find returns nil + EmptyResultError.
	conn := newTestClient(t, jsonHandler(http.StatusOK, map[string]any{}))
	refreshFunc := statusNetwork(ctx, conn, "odn-doesnotexist")

	result, state, err := refreshFunc()

	// The Find function returns a NotFoundError‐style empty result error,
	// and statusNetwork swallows NotFound errors.  The key assertion is that
	// we don't panic on a nil OdbNetwork dereference.
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if result != nil {
		t.Fatalf("expected nil result, got: %v", result)
	}
	if state != "" {
		t.Fatalf("expected empty state, got: %q", state)
	}
}

// TestStatusNetwork_ValidOutput verifies statusNetwork returns the correct
// status string from a fully populated OdbNetwork response.
func TestStatusNetwork_ValidOutput(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	conn := newTestClient(t, jsonHandler(http.StatusOK, map[string]any{
		"odbNetwork": map[string]any{
			"odbNetworkId": "odn-12345",
			"status":       "AVAILABLE",
		},
	}))
	refreshFunc := statusNetwork(ctx, conn, "odn-12345")

	result, state, err := refreshFunc()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if state != string(odbtypes.ResourceStatusAvailable) {
		t.Fatalf("expected state %q, got %q", odbtypes.ResourceStatusAvailable, state)
	}
}

// TestStatusNetwork_NotFound verifies statusNetwork handles a
// ResourceNotFoundException without panicking.
func TestStatusNetwork_NotFound(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	conn := newTestClient(t, jsonHandler(http.StatusBadRequest, map[string]any{
		"__type":  "ResourceNotFoundException",
		"message": "ODB network not found",
	}))
	refreshFunc := statusNetwork(ctx, conn, "odn-doesnotexist")

	result, state, err := refreshFunc()

	if err != nil {
		t.Fatalf("expected no error for NotFound, got: %v", err)
	}
	if result != nil {
		t.Fatalf("expected nil result, got: %v", result)
	}
	if state != "" {
		t.Fatalf("expected empty state, got: %q", state)
	}
}

// s3ManagedResourceStatus is a test callback that extracts the S3 managed
// service status, matching the pattern used in production.
func s3ManagedResourceStatus(ms *odbtypes.ManagedServices) odbtypes.ManagedResourceStatus {
	if ms.S3Access == nil {
		return ""
	}
	return ms.S3Access.Status
}

// TestStatusManagedService_NilManagedServices verifies that statusManagedService
// handles a response where ManagedServices is nil without panicking.
// This is a regression test for the nil pointer dereference crash.
func TestStatusManagedService_NilManagedServices(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// Return an OdbNetwork with no managedServices field.
	conn := newTestClient(t, jsonHandler(http.StatusOK, map[string]any{
		"odbNetwork": map[string]any{
			"odbNetworkId": "odn-12345",
			"status":       "AVAILABLE",
		},
	}))
	refreshFunc := statusManagedService(ctx, conn, "odn-12345", s3ManagedResourceStatus)

	result, state, err := refreshFunc()

	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	// When ManagedServices is nil, the function should return nil without panicking.
	if result != nil {
		t.Fatalf("expected nil result when ManagedServices is nil, got: %v", result)
	}
	if state != "" {
		t.Fatalf("expected empty state, got: %q", state)
	}
}

// TestStatusManagedService_NilOutput verifies that statusManagedService
// handles a nil OdbNetwork result (empty API response) without panicking.
// This is the primary regression test for the nil pointer dereference at
// network.go:682 that caused the Atlantis crash.
func TestStatusManagedService_NilOutput(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// API returns 200 with no odbNetwork field → Find returns nil + EmptyResultError.
	conn := newTestClient(t, jsonHandler(http.StatusOK, map[string]any{}))
	refreshFunc := statusManagedService(ctx, conn, "odn-doesnotexist", s3ManagedResourceStatus)

	result, state, err := refreshFunc()

	// Find returns EmptyResultError (which is a NotFoundError), but
	// statusManagedService treats any error as propagated.  The critical
	// assertion is NO PANIC.
	if result != nil {
		t.Fatalf("expected nil result, got: %v", result)
	}
	_ = state
	_ = err
}

// TestStatusManagedService_WithManagedServices verifies the happy path where
// ManagedServices is fully populated.
func TestStatusManagedService_WithManagedServices(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	conn := newTestClient(t, jsonHandler(http.StatusOK, map[string]any{
		"odbNetwork": map[string]any{
			"odbNetworkId": "odn-12345",
			"status":       "AVAILABLE",
			"managedServices": map[string]any{
				"s3Access": map[string]any{
					"status": "ENABLED",
				},
			},
		},
	}))
	refreshFunc := statusManagedService(ctx, conn, "odn-12345", s3ManagedResourceStatus)

	result, state, err := refreshFunc()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if state != string(odbtypes.ManagedResourceStatusEnabled) {
		t.Fatalf("expected state %q, got %q", odbtypes.ManagedResourceStatusEnabled, state)
	}
}

// TestStatusManagedService_NotFound verifies that statusManagedService
// propagates the error when the network is not found.
func TestStatusManagedService_NotFound(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	conn := newTestClient(t, jsonHandler(http.StatusBadRequest, map[string]any{
		"__type":  "ResourceNotFoundException",
		"message": "ODB network not found",
	}))
	refreshFunc := statusManagedService(ctx, conn, "odn-doesnotexist", s3ManagedResourceStatus)

	_, _, err := refreshFunc()

	// statusManagedService does NOT have a NotFound guard (unlike statusNetwork),
	// so it propagates the error.
	if err == nil {
		t.Fatal("expected error for NotFound, got nil")
	}
}
