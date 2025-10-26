// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package s3

import (
	"context"
	"testing"
)

func TestEmptyBucketAction_CanBeCreated(t *testing.T) {
	ctx := context.Background()
	
	// Test that the action can be created
	actionInstance, err := newEmptyBucketAction(ctx)
	if err != nil {
		t.Fatalf("Failed to create action: %v", err)
	}
	
	if actionInstance == nil {
		t.Fatal("Action instance is nil")
	}
	
	// Test that it has the expected type
	if _, ok := actionInstance.(*emptyBucketAction); !ok {
		t.Error("Action is not of expected type *emptyBucketAction")
	}
}
