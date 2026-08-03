// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package framework

import (
	"testing"

	"github.com/hashicorp/terraform-provider-aws/internal/conns"
)

// Test that ActionWithConfigure can be instantiated and has the expected methods
func TestActionWithConfigureCompilation(t *testing.T) {
	t.Parallel()

	// This test ensures our new types compile correctly
	var action ActionWithConfigure

	// Test that it has the Meta method from withMeta
	if action.Meta() != nil {
		t.Error("Expected nil meta before configuration")
	}

	// Test that it embeds withMeta correctly
	action.meta = &conns.AWSClient{}
	if action.Meta() == nil {
		t.Error("Expected non-nil meta after setting")
	}
}

// Test that ActionWithModel can be instantiated
func TestActionWithModelCompilation(t *testing.T) {
	t.Parallel()

	// Test model
	type testModel struct {
		Name string `tfsdk:"name"`
	}

	// This test ensures our new generic type compiles correctly
	var action ActionWithModel[testModel]

	// Test that it has the Meta method from ActionWithConfigure
	if action.Meta() != nil {
		t.Error("Expected nil meta before configuration")
	}

	// Test that it embeds ActionWithConfigure correctly
	action.meta = &conns.AWSClient{}
	if action.Meta() == nil {
		t.Error("Expected non-nil meta after setting")
	}
}
